// package svg provides a renderer for a sort of calendar for trips.
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

// svg types
type container struct {
	borderColour     string
	backgroundColour string
	borderWidth      int
}

func (c *container) render(width, height int, svg *svg.SVG) {
	const rectStyle string = "fill:%s;stroke:%s;stroke-width:%d"
	svg.Rect(0, 0, width, height, fmt.Sprintf(
		rectStyle,
		c.backgroundColour,
		c.borderColour,
		c.borderWidth),
	)
}

func newContainer(color, bgColour string, width int) *container {
	return &container{color, bgColour, width}
}

type label struct {
	text        string
	colour      string
	strokeWidth int
}

func newLabel(text, colour string, strokeWidth int) *label {
	return &label{text, colour, strokeWidth}
}

type legend struct {
	x, y   int // absolute coordinates
	labels []label
}

func newLegend(x, y int, labels []label) *legend {
	return &legend{x, y, labels}
}

func (le *legend) render(svg *svg.SVG) {
	const (
		keyWidth          int    = 20 // px
		keySpacing        int    = 10 // px
		fontStyle         string = "font-family:sans-serif;font-size:9pt;fill:black;text-anchor:left"
		lineStyle         string = "stroke:%s;stroke-width:%d"
		lineBottomPadding int    = 4 // px
	)

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

type xyColRow struct {
	x, y     int
	col, row int
}

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

// changeDate is a function to either advance or retreat towards a day
// of the week
func changeDate(date time.Time, targetDay int, d time.Duration) time.Time {
	if int(date.Weekday()) == targetDay {
		return date
	}
	for i := 0; i < 6; i++ {
		date = date.Add(d)
		if int(date.Weekday()) == targetDay {
			return date
		}
	}
	panic("date could not be changed")
}

// newGrid makes a new weekGrid
func newGrid(trips *trips.Trips) *weekGrid {
	const (
		weeksPerRow     int = 8
		rightPadding    int = 21  // px
		topPadding      int = 21  // px
		bottomPadding   int = 34  // px
		legendOwnHeight int = 7   // px
		weekBlockHeight int = 62  // px
		weekBlockWidth  int = 124 // px
	)
	grid := weekGrid{
		columns: weeksPerRow,
	}

	grid.startDate = changeDate(trips.Start, 1, time.Hour*24*-1)
	grid.endDate = changeDate(trips.End, 0, time.Hour*24*+1)
	grid.weekNum = int(math.Round(grid.endDate.Sub(grid.startDate).Hours() / (7 * 24)))
	grid.rows = int(math.Ceil(float64(grid.weekNum) / float64(weeksPerRow)))

	//       +--+-----------+-----------+----------/ -+-----------+---+
	//       |  |           |           |          /  |           |   |
	//  r1   |  +------     |           |          /  |           |   |
	//       |  | legend    |           |          /  |           |   |
	//       |  |           |           |          /  |           |   |
	//       |  |           |           |          /  |           |   |
	//  r2   |  +-----------+-----------+-------   / -+---------  |-  |
	//       |  |w1         |w2         |w3        /  |w8         |R  | R = right extent
	//       |  |           |           |          /  |           |   |
	//       +--+-----------+-----------+----------/ -+-----------+---+
	//
	//       lp c1          c2          c3            c8          c9
	//       lp = left padding
	//       c1, c2 are the week column positions
	//
	//       coordinate system runs from top left
	//       legend is at x0, y0
	//       week3  is at x3, y1

	grid.width = leftPadding + (weekBlockWidth * grid.columns) + rightPadding
	grid.rightGutter = (weekBlockWidth * grid.columns) + weekNotchSpacing

	grid.legendHeight = topPadding + legendOwnHeight // r1
	grid.height = grid.legendHeight + (weekBlockHeight * grid.rows) + bottomPadding

	// build the dateMatrix
	grid.dateMatrix = map[time.Time]xyColRow{}
	col, row := 0, 0
	for d := grid.startDate; d.Before(grid.endDate); d = d.Add(time.Hour * 24 * 7) {
		cx := leftPadding + (weekBlockWidth * col)
		cy := grid.legendHeight + weekBlockHeight + (weekBlockHeight * row) // make space for first row
		grid.dateMatrix[d] = xyColRow{
			x: cx, y: cy, col: col, row: row,
		}
		col += 1
		if col == 8 {
			col = 0
			row += 1
		}
	}

	return &grid
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

type segment struct {
	x1, y1, x2, y2 int
}

// getSegments returns a list of line segments describing where a stripe
// should be written. The function panics if a date can't be found
func (wg *weekGrid) getSegments(start, end time.Time, level int) []segment {

	const (
		stripePadding int = 10 // px
	)

	isoDOW := func(date time.Time) int {
		dow := (int(date.Weekday()) - 1) % 7
		if dow == -1 {
			return 6
		}
		return dow
	}

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

	startMonday := changeDate(start, 1, time.Hour*24*-1)
	startDay := isoDOW(start)
	startCoords, ok := wg.coordinates(startMonday)
	if !ok {
		panic(fmt.Sprintf("couldn't resolve startCoords %v", startMonday))
	}

	endMonday := changeDate(end, 1, time.Hour*24*-1)
	endDay := isoDOW(end)
	endCoords, ok := wg.coordinates(endMonday)
	if !ok {
		panic(fmt.Sprintf("couldn't resolve endCoords %v", endMonday))
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
		}
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
			// to do this properly need to find the y height of the
			// monday in this week
			rowY := wg.rowYvalue(row)
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
	return segments
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

// generally needed const values
const (
	leftPadding      int = 34 // px
	weekNotchHeight  int = 9
	weekNotchSpacing int = 14
	weekLinesLen     int = 98 // px
	weekLinesPadding int = 16 // px
)

func (w *week) render(svg *svg.SVG) {
	const (
		fontStyle       string = "font-family:sans-serif;font-size:9pt;fill:black;text-anchor:left"
		lineStyle       string = "stroke:%s;stroke-width:%d"
		weekLinesColour string = "black"
		weekLinesStroke int    = 2
	)

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
func (s *stripe) render(g *weekGrid, svg *svg.SVG) {
	const (
		lineStyle string = "stroke:%s;stroke-width:%d"
	)
	segments := g.getSegments(s.startDate, s.endDate, s.level)
	svg.Group(s.typer)
	for _, seg := range segments {
		svg.Line(
			seg.x1, seg.y1, seg.x2, seg.y2,
			fmt.Sprintf(lineStyle, s.colour, s.strokeWidth),
		)
	}
	svg.Title(s.title)
	svg.Gend()
}

// TripsAsSVG renders a set of trips as an SVG graphic calendar marking
// the holidays, longest window or breach window according to the
// results of the Trip calculations.
func TripsAsSVG(trips *trips.Trips, w io.Writer) error {

	grid := newGrid(trips)

	canvas := svg.New(w)
	canvas.Start(grid.width, grid.height)
	background := newContainer("#c4c8b7ff", "#ecececff", 2)
	background.render(grid.width, grid.height, canvas)

	legend := newLegend(leftPadding, grid.legendHeight, []label{
		label{"holidays", "green", 5},
		label{"breach", "red", 5},
		label{"longest window without breach", "blue", 5},
	})
	legend.render(canvas)

	for i := range grid.weekNum {
		date := grid.startDate.Add(time.Hour * 24 * 7 * time.Duration(i)) // add a week to the start date
		coordinates, ok := grid.coordinates(date)
		if !ok {
			return fmt.Errorf("date %s no coordinates\n", date)
		}
		week := newWeek(coordinates.x, coordinates.y, date)
		week.render(canvas)
	}

	for _, tr := range trips.OriginalHolidays {
		thisStripe := newStripe("holiday", "", "green", tr.Start, tr.End, 5, 0)
		thisStripe.render(grid, canvas)
	}

	if trips.Breach {
		info := fmt.Sprintf("%d days", trips.Window.DaysAway)
		thisStripe := newStripe("breach", info, "red", trips.Window.Start, trips.Window.End, 5, 1)
		thisStripe.render(grid, canvas)
	} else {
		info := fmt.Sprintf("%d days", trips.Window.DaysAway)
		thisStripe := newStripe("longest window", info, "blue", trips.Window.Start, trips.Window.End, 5, 1)
		thisStripe.render(grid, canvas)
	}

	canvas.End()
	return nil
}
