// Package svg provides an svg renderer for a graphic calendar for
// timeaway/trips as set out in the internal weekGrid documentation.
package svg

import (
	"fmt"
	"io"
	"maps"
	"math"
	"time"

	svg "github.com/ajstarks/svgo"
	"github.com/rorycl/timeaway/trips"
)

const (
	// canvas
	// https://www.w3.org/TR/SVG11/coords.html#ViewBoxAttribute
	// viewBox="0 0 1500 1000" preserveAspectRatio="xMidYMid"

	// styles
	rectStyle       string = "fill:%s;stroke:%s;stroke-width:%d"
	lineStyle       string = "stroke:%s;stroke-width:%d"
	fontStyle       string = "font-family:sans-serif;font-size:9pt;fill:black;text-anchor:left"
	weekLinesColour string = "black"

	// placement based on design
	keyWidth          int = 20 // px
	keySpacing        int = 10 // px
	lineBottomPadding int = 4  // px
	weeksPerRow       int = 8
	rightPadding      int = 21  // px
	topPadding        int = 21  // px
	bottomPadding     int = 34  // px
	legendOwnHeight   int = 7   // px
	weekBlockHeight   int = 62  // px
	weekBlockWidth    int = 124 // px
	stripePadding     int = 10  // px
	leftPadding       int = 34  // px
	weekNotchHeight   int = 9
	weekNotchSpacing  int = 14
	weekLinesLen      int = 98 // px
	weekLinesPadding  int = 16 // px
	weekLinesStroke   int = 2

	// target width, which requires the design to be scale
	targetWidth int = 860 // px
)

// container is the rectangle describing the content
type container struct {
	borderColour     string
	backgroundColour string
	borderWidth      int
}

func newContainer(color, bgColour string, width int) *container {
	return &container{color, bgColour, width}
}

func (c *container) render(width, height int, svg *svg.SVG) {
	svg.Rect(0, 0, width, height, fmt.Sprintf(
		rectStyle,
		c.backgroundColour,
		c.borderColour,
		c.borderWidth),
	)
}

// label is an item in the legend
type label struct {
	text        string
	colour      string
	strokeWidth int
}

// legend describes a diagram "key" by position with labels
type legend struct {
	x, y   int // absolute coordinates
	labels []label
}

func newLegend(x, y int, labels []label) *legend {
	return &legend{x, y, labels}
}

func (le *legend) render(svg *svg.SVG) {
	offsetX, offsetY := 0, 0
	for _, l := range le.labels {
		svg.Line(
			le.x+offsetX,
			le.y+offsetY-lineBottomPadding,
			le.x+offsetX+keyWidth,
			le.y+offsetY-lineBottomPadding,
			fmt.Sprintf(lineStyle, l.colour, l.strokeWidth))
		offsetX += keyWidth + keySpacing
		textLen := len(l.text) * 6
		svg.Text(le.x+offsetX, le.y+offsetY, l.text, fontStyle)
		offsetX += textLen + keySpacing
	}
}

// xyColRow is a struct describing the x, y pixel position and the
// column and row of the week in question.
type xyColRow struct {
	x, y     int
	col, row int
}

// weekGrid describes the heart of the layout system starting at
// startDate and ending at endDate setting out the weeks over rows and
// columns under the legend (set out at legendHeight). Each week,
// defined by the date of its Monday, is placed according to the
// xyColRow set out in dateMatrix.
//
//       +--+-----------+-----------+----------/ -+-----------+---+
//       |  |           |           |          /  |           |   |
//  r1   |  +------     |           |          /  |           |   |
//       |  | legend    |           |          /  |           |   |
//       |  |           |           |          /  |           |   |
//       |  |           |           |          /  |           |   |
//  r2   |  +-----------+-----------+-------   / -+---------  |-  |
//       |  |w1         |w2         |w3        /  |w8         |   |
//       |  |           |           |          /  |           |   |
//       +--+-----------+-----------+----------/ -+-----------+---+
//
//       lp c1          c2          c3            c8          R rightGutter
//       lp = left padding
//       c1, c2 are the week column positions
//
//       coordinate system runs from top left
//       legend is at x0, y0
//       week3  is at x3, y1

type weekGrid struct {
	startDate     time.Time
	endDate       time.Time
	rightGutter   int // most right hand right gutter
	weekNum       int // number of weeks
	rows          int // number of rows at 8 weeks/row
	columns       int // number of columns
	legendHeight  int // position of legend
	width, height int // overall width and height

	// report the column & row pos and coordinates
	// for each monday in the matrix
	dateMatrix map[time.Time]xyColRow
}

// newGrid makes a new weekGrid with the appropriate dimensions and
// coordinates.
func newGrid(trips *trips.Trips) (*weekGrid, error) {
	grid := weekGrid{
		columns: weeksPerRow,
	}

	// Use the start and end date of the trips by default for the
	// reporting period. However if the window extends past these dates
	// and the Trips are in Breach, use the window dates instead in
	// order to render the breach strip correctly (otherwise the breach
	// strip cannot resolve to a system coordinate).
	var err error
	minStartDate := trips.Start
	if minStartDate.After(trips.Window.Start) && trips.Breach {
		minStartDate = trips.Window.Start
	}
	grid.startDate, err = changeDate(trips.Start, 1, time.Hour*24*-1)
	if err != nil {
		return nil, fmt.Errorf("grid startDate error %w", err)
	}

	maxEndDate := trips.End
	if maxEndDate.Before(trips.Window.End) && trips.Breach {
		maxEndDate = trips.Window.End
	}
	grid.endDate, err = changeDate(maxEndDate, 0, time.Hour*24*+1)
	if err != nil {
		return nil, fmt.Errorf("grid endDate error %w", err)
	}

	// Determine widths, heights and numbers of items.
	grid.weekNum = int(math.Round(grid.endDate.Sub(grid.startDate).Hours() / (7 * 24)))
	grid.rows = int(math.Ceil(float64(grid.weekNum) / float64(weeksPerRow)))

	grid.width = leftPadding + (weekBlockWidth * grid.columns) + rightPadding
	grid.rightGutter = (weekBlockWidth * grid.columns) + weekNotchSpacing

	grid.legendHeight = topPadding + legendOwnHeight // r1
	grid.height = grid.legendHeight + (weekBlockHeight * grid.rows) + bottomPadding

	// Set out the dates/coordinates for dateMatrix.
	grid.dateMatrix = map[time.Time]xyColRow{}
	col, row := 0, 0
	for d := grid.startDate; d.Before(grid.endDate); d = d.Add(time.Hour * 24 * 7) {
		cx := leftPadding + (weekBlockWidth * col)
		cy := grid.legendHeight + weekBlockHeight + (weekBlockHeight * row) // make space for first row
		grid.dateMatrix[d] = xyColRow{
			x: cx, y: cy, col: col, row: row,
		}
		col += 1
		if col == grid.columns {
			col = 0
			row += 1
		}
	}
	return &grid, nil
}

// coordinates returns the coordinates, if any, of each date
func (wg *weekGrid) coordinates(date time.Time) (xyColRow, bool) {
	coord, ok := wg.dateMatrix[date]
	return coord, ok
}

// rowYvalue returns the y coordinate of the named row or -1 if there is
// no match.
func (wg *weekGrid) rowYvalue(row int) int {
	for v := range maps.Values(wg.dateMatrix) {
		if v.row == row {
			return v.y
		}
	}
	return -1
}

// segment describes the positioning of a horizontal line used to
// describe a "stripe". y1 and y2 will always be the same.
type segment struct {
	x1, y1, x2, y2 int
}

// getSegments returns a list of line segments describing where a stripe
// should be written. This function calculates if a stripe needs to be
// split across margins so that, for example, a week spanning 12 weeks,
// can be split over 2 or 3 rows of the output, returning either 2 or 3
// items in the return segment slice.
func (wg *weekGrid) getSegments(start, end time.Time, level int) ([]segment, error) {

	// advanceDays advances from Monday to the day in the week "notches"
	// width pixels. addDay adds a day for the end notch, because each
	// day spans one notch; there are 8 gaps and 7 notches.
	advanceDays := func(dow int, addDay bool) int {
		if addDay {
			dow += 1
		}
		return dow * weekNotchSpacing
	}

	stripeYoffset := weekLinesPadding + weekNotchHeight + (stripePadding * (level + 1))

	startMonday, err := changeDate(start, 1, time.Hour*24*-1)
	if err != nil {
		return nil, fmt.Errorf("getSegments: couldn't find startMonday %v", startMonday)
	}
	startDay := isoDOW(start)
	startCoords, ok := wg.coordinates(startMonday)
	if !ok {
		return nil, fmt.Errorf("getSegments: couldn't resolve startCoords %v", startMonday)
	}

	endMonday, err := changeDate(end, 1, time.Hour*24*-1)
	if err != nil {
		return nil, fmt.Errorf("getSegments: couldn't find endMonday %v", endMonday)
	}
	if err != nil {
		return nil, err
	}
	endDay := isoDOW(end)
	endCoords, ok := wg.coordinates(endMonday)
	if !ok {
		return nil, fmt.Errorf("getSegments: couldn't resolve endCoords %v", endMonday)
	}

	rowCount := endCoords.row - startCoords.row + 1
	if rowCount == 1 {
		return []segment{
			segment{
				x1: startCoords.x + advanceDays(startDay, false),
				y1: startCoords.y - stripeYoffset,
				x2: endCoords.x + advanceDays(endDay, true),
				y2: endCoords.y - stripeYoffset,
			},
		}, nil
	}

	started := false
	segments := []segment{}
	for row := startCoords.row; row <= endCoords.row; row++ {
		if !started {
			segments = append(segments, segment{
				x1: startCoords.x + advanceDays(startDay, false),
				y1: startCoords.y - stripeYoffset,
				x2: wg.rightGutter,
				y2: startCoords.y - stripeYoffset,
			})
			started = true
			continue
		}
		if row < endCoords.row {
			rowY := wg.rowYvalue(row) // retrieve Y coordinates for this row
			segments = append(segments, segment{
				x1: leftPadding,
				y1: rowY - stripeYoffset,
				x2: wg.rightGutter,
				y2: rowY - stripeYoffset,
			})
			continue
		}
		segments = append(segments, segment{
			x1: leftPadding,
			y1: endCoords.y - stripeYoffset,
			x2: endCoords.x + advanceDays(endDay, true),
			y2: endCoords.y - stripeYoffset,
		})
	}
	return segments, nil
}

// week describes a week "skeleton" diagram of a line with 8 notches
// describing the days of the week, with a date below.
type week struct {
	x, y   int // absolute top left coordinate
	monday time.Time
}

func newWeek(x, y int, monday time.Time) *week {
	return &week{x, y, monday}
}

func (w *week) render(svg *svg.SVG) {
	var text string
	if w.monday.Day() <= 7 {
		text = w.monday.Format("2 Jan 2006")
	} else {
		text = w.monday.Format("2")
	}

	// week label
	svg.Text(w.x, w.y, text, fontStyle)
	// horizontal week line
	svg.Line(
		w.x,
		w.y-weekLinesPadding-weekNotchHeight,
		w.x+weekLinesLen,
		w.y-weekLinesPadding-weekNotchHeight,
		fmt.Sprintf(lineStyle, weekLinesColour, weekLinesStroke),
	)
	for notch := range 8 {
		spacing := weekNotchSpacing * notch
		svg.Line(
			w.x+spacing,
			w.y-weekLinesPadding,
			w.x+spacing,
			w.y-weekLinesPadding-weekNotchHeight,
			fmt.Sprintf(lineStyle, weekLinesColour, weekLinesStroke),
		)
	}
}

// stripe represents a "stripe" of information above the calendar weeks,
// for example showing holidays just above the week skeletons (on level
// 0) or breach information (on level 1, above level 0). See the
// template for an example.
type stripe struct {
	typer              string // ok, breach or holiday
	title              string
	startDate, endDate time.Time
	colour             string
	strokeWidth        int
	level              int // 0 is first level above week notches, 1 the second
}

func newStripe(typer, info, colour string, start, end time.Time, width, level int) *stripe {
	if info != "" {
		info = " (" + info + ") "
	}
	title := fmt.Sprintf(
		"%s: %s to %s",
		typer+info,
		start.Format("2006-01-02"),
		end.Format("2006-01-02"),
	)
	return &stripe{typer, title, start, end, colour, width, level}
}

// Rendering a stripe requires some to split across visual lines. The
// grid function getSegments returns the x1, y1, x2, y2 coordinates to
// render the slice.
func (s *stripe) render(g *weekGrid, svg *svg.SVG) error {
	segments, err := g.getSegments(s.startDate, s.endDate, s.level)
	if err != nil {
		return fmt.Errorf("stripe (%s) segments error: %s", s.typer, err)
	}
	svg.Group(s.typer)
	for _, seg := range segments {
		svg.Line(
			seg.x1, seg.y1, seg.x2, seg.y2,
			fmt.Sprintf(lineStyle, s.colour, s.strokeWidth),
		)
	}
	svg.Title(s.title)
	svg.Gend()
	return nil
}

// TripsAsSVG renders a set of trips as an SVG graphic calendar marking
// the holidays, longest window or breach window according to the
// results of the Trip calculations.
func TripsAsSVG(trips *trips.Trips, w io.Writer) error {

	grid, err := newGrid(trips)
	if err != nil {
		return err
	}

	canvas := svg.New(w)
	canvas.Start(grid.width, grid.height)
	canvas.Scale(float64(targetWidth) / float64(grid.width)) // needs GEnd() -- see bottom

	background := newContainer("#c4c8b7ff", "#ecececff", 2)
	background.render(grid.width, grid.height, canvas)

	legend := newLegend(leftPadding, grid.legendHeight, []label{
		label{"holidays", "green", 5},
		label{"breach", "red", 5},
		label{"longest window without breach", "blue", 5},
	})
	legend.render(canvas)

	// render the weeks by progressing a week at a time from the start
	// date to the end date (generating grid.weekNum entries).
	for i := range grid.weekNum {
		date := grid.startDate.Add(time.Hour * 24 * 7 * time.Duration(i))
		coordinates, ok := grid.coordinates(date)
		if !ok {
			return fmt.Errorf("date %s no coordinates\n", date)
		}
		week := newWeek(coordinates.x, coordinates.y, date)
		week.render(canvas)
	}

	// stripe in the holidays
	for _, tr := range trips.OriginalHolidays {
		thisStripe := newStripe("holiday", "", "green", tr.Start, tr.End, 5, 0)
		err := thisStripe.render(grid, canvas)
		if err != nil {
			return fmt.Errorf("stripe render error: %w", err)
		}
	}

	// stripe in either the breach or no-breach longest window line
	// segments showing the first holiday start date
	// (trips.Window.OverlapStart) and last holiday end date
	// (trips.Window.OverlapEnd) overlapping with the assessment window.
	// Note that trips.Window.OverlapStart and trips.Window.OverlapEnd
	// will the same as trips.Window.Start and trips.Window.End if there
	// is a breach.
	if trips.Breach {
		info := fmt.Sprintf("%d days", trips.Window.DaysAway)
		thisStripe := newStripe("breach", info, "red", trips.Window.Start, trips.Window.End, 5, 1)
		err := thisStripe.render(grid, canvas)
		if err != nil {
			return fmt.Errorf("stripe render error: %w", err)
		}
	} else {
		info := fmt.Sprintf("%d days", trips.Window.DaysAway)
		thisStripe := newStripe("longest window", info, "blue", trips.Window.OverlapStart, trips.Window.OverlapEnd, 5, 1)
		err := thisStripe.render(grid, canvas)
		if err != nil {
			return fmt.Errorf("stripe render error: %w", err)
		}
	}

	canvas.Gend() // end Scale
	canvas.End()
	return nil
}
