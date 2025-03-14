// package svg provides a renderer for a sort of calendar for trips.
package main

import (
	"fmt"
	"math"
	"os"
	"time"

	svg "github.com/ajstarks/svgo"
	"github.com/rorycl/timeaway/trips"
)

func makeTrips() *trips.Trips {

	// Calculation results
	//
	// The planned trips breached the 90 days in 180 day rule with 91 days away.
	//
	// The maximum days away were for a 180 day window from Tuesday 01/07/2025 to Saturday 27/12/2025.
	//
	// The trips in this calculation are:
	//
	//  1. Tuesday 17/12/2024 to Saturday 04/01/2025 (19 days)
	//     not covered by the window.
	//  2. Friday 14/02/2025 to Thursday 27/02/2025 (14 days)
	//     not covered by the window.
	//  3. Thursday 03/04/2025 to Wednesday 23/04/2025 (21 days)
	//     not covered by the window.
	//  4. Tuesday 01/07/2025 to Wednesday 03/09/2025 (65 days)
	//     fully covered by the window.
	//  5. Sunday 09/11/2025 to Sunday 16/11/2025 (8 days)
	//     fully covered by the window.
	//  6. Wednesday 10/12/2025 to Tuesday 06/01/2026 (28 days)
	//     parially covered by the window from Wednesday 10/12/2025 for 18 days.

	return &trips.Trips{
		WindowSize: 180,
		MaxStay:    90,
		Start:      time.Date(2024, 12, 17, 0, 0, 0, 0, time.UTC),
		End:        time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
		OriginalHolidays: []trips.Holiday{
			trips.Holiday{
				Start:          time.Date(2024, 12, 17, 0, 0, 0, 0, time.UTC),
				End:            time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC),
				Duration:       19,
				PartialHoliday: nil,
			},
			trips.Holiday{
				Start:          time.Date(2025, 2, 14, 0, 0, 0, 0, time.UTC),
				End:            time.Date(2025, 2, 27, 0, 0, 0, 0, time.UTC),
				Duration:       14,
				PartialHoliday: nil,
			},
			trips.Holiday{
				Start:          time.Date(2025, 4, 3, 0, 0, 0, 0, time.UTC),
				End:            time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC),
				Duration:       21,
				PartialHoliday: nil,
			},
			trips.Holiday{
				Start:          time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
				End:            time.Date(2025, 9, 3, 0, 0, 0, 0, time.UTC),
				Duration:       65,
				PartialHoliday: nil,
			},
			trips.Holiday{
				Start:          time.Date(2025, 11, 9, 0, 0, 0, 0, time.UTC),
				End:            time.Date(2025, 11, 16, 0, 0, 0, 0, time.UTC),
				Duration:       8,
				PartialHoliday: nil,
			},
			trips.Holiday{
				Start:          time.Date(2025, 12, 10, 0, 0, 0, 0, time.UTC),
				End:            time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
				Duration:       28,
				PartialHoliday: nil,
			},
		},
		Window: trips.Window{
			Start:    time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
			End:      time.Date(2025, 12, 27, 0, 0, 0, 0, time.UTC),
			DaysAway: 91,
			Overlaps: 3,
			Holidays: []trips.Holiday{
				trips.Holiday{
					Start:          time.Date(2024, 12, 17, 0, 0, 0, 0, time.UTC),
					End:            time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC),
					Duration:       19,
					PartialHoliday: nil,
				},
				trips.Holiday{
					Start:          time.Date(2025, 2, 14, 0, 0, 0, 0, time.UTC),
					End:            time.Date(2025, 2, 27, 0, 0, 0, 0, time.UTC),
					Duration:       14,
					PartialHoliday: nil,
				},
				trips.Holiday{
					Start:          time.Date(2025, 4, 3, 0, 0, 0, 0, time.UTC),
					End:            time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC),
					Duration:       21,
					PartialHoliday: nil,
				},
				trips.Holiday{
					Start:    time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
					End:      time.Date(2025, 9, 3, 0, 0, 0, 0, time.UTC),
					Duration: 65,
					PartialHoliday: &trips.Holiday{
						Start:          time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
						End:            time.Date(2025, 9, 3, 0, 0, 0, 0, time.UTC),
						Duration:       65,
						PartialHoliday: nil,
					},
				},
				trips.Holiday{
					Start:    time.Date(2025, 11, 9, 0, 0, 0, 0, time.UTC),
					End:      time.Date(2025, 11, 16, 0, 0, 0, 0, time.UTC),
					Duration: 8,
					PartialHoliday: &trips.Holiday{
						Start:          time.Date(2025, 11, 9, 0, 0, 0, 0, time.UTC),
						End:            time.Date(2025, 11, 16, 0, 0, 0, 0, time.UTC),
						Duration:       8,
						PartialHoliday: nil,
					},
				},
				trips.Holiday{
					Start:    time.Date(2025, 12, 10, 0, 0, 0, 0, time.UTC),
					End:      time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
					Duration: 28,
					PartialHoliday: &trips.Holiday{
						Start:          time.Date(2025, 12, 10, 0, 0, 0, 0, time.UTC),
						End:            time.Date(2025, 12, 27, 0, 0, 0, 0, time.UTC),
						Duration:       18,
						PartialHoliday: nil,
					},
				},
			},
		},
		LongestDaysAway: 91,
		Error:           nil,
		Breach:          true,
	}
}

// svg types
type container struct {
	borderColour     string
	backgroundColour string
	borderWidth      int
}

func (c *container) render(x, y int, svg *svg.SVG) {
	const rectStyle string = "fill:%s;stroke:%s;stroke-width:%d"
	svg.Rect(0, 0, x, y, fmt.Sprintf(
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
		keySpacing        int    = 16 // px
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
	col, row int
	x, y     int
}

type weekGrid struct {
	startDate     time.Time
	endDate       time.Time
	weekNum       int   // number of weeks
	rows          int   // number of rows at 8 weeks/row
	columns       int   // number of columns
	colPositions  []int // left coordinate of columns, also most right hand column
	rowPositions  []int // top right coordinate of rows
	width, height int   // overall width and height

	// report the column & row pos and coordinates
	// for each monday in the matrix
	dateMatrix map[time.Time]xyColRow
}

// newGrid makes a new weekGrid
func newGrid(trips *trips.Trips) *weekGrid {
	const (
		weeksPerRow     int = 8
		leftPadding     int = 21  // px
		rightPadding    int = 21  // px
		topPadding      int = 21  // px
		bottomPadding   int = 21  // px
		legendHeight    int = 7   // px
		legendPadding   int = 21  // px between legend and week
		weekBlockHeight int = 62  // px
		weekBlockWidth  int = 124 // px
	)
	grid := weekGrid{}

	changeDate := func(date time.Time, targetDay int, d time.Duration) time.Time {
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

	grid.startDate = changeDate(trips.Start, 1, time.Hour*24*-1)
	grid.endDate = changeDate(trips.End, 0, time.Hour*24*+1)
	grid.weekNum = int(math.Round(grid.endDate.Sub(grid.startDate).Hours() / (7 * 24)))
	grid.rows = int(math.Floor(float64(grid.weekNum)/float64(weeksPerRow))) + 1 // add the legend to the number of rows

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

	// column positions from left margin
	grid.colPositions = []int{
		leftPadding,                        // c1
		leftPadding + (weekBlockWidth * 1), // c2
		leftPadding + (weekBlockWidth * 2), // c3
		leftPadding + (weekBlockWidth * 3), // c4
		leftPadding + (weekBlockWidth * 4), // c5
		leftPadding + (weekBlockWidth * 5), // c6
		leftPadding + (weekBlockWidth * 6), // c7
		leftPadding + (weekBlockWidth * 7), // c8
		leftPadding + (weekBlockWidth * 8), // c9 (R/right extent)
		// right margin is c9 + rightPadding
	}
	grid.width = grid.colPositions[len(grid.colPositions)-1] + rightPadding
	grid.columns = len(grid.colPositions) - 1 // c9 is a virtual column

	grid.height = topPadding + legendHeight + (weekBlockHeight * grid.rows) + bottomPadding

	// row positions from top margin
	grid.rowPositions = []int{
		topPadding + legendHeight, // r1 is at bottom of legend
	}
	for i := range grid.rows {
		grid.rowPositions = append(grid.rowPositions, (topPadding+legendHeight)+(weekBlockHeight*(i+1))) // bottom of each row
	}

	// build the dateMatrix
	grid.dateMatrix = map[time.Time]xyColRow{}
	col, row := 0, 0
	for d := grid.startDate; d.Before(grid.end.Add(time.Hour * 24 * 7)); d.Add(time.Hour * 24 * 7) {
		cx, cy, err := grid.coordinates(col, row+1)
		if err != nil {
			panic(err)
		}
		grid.dateMatrix[d] = xyColRow{cx, cy, col, row}
		col += 1
		if col == 8 {
			col = 0
			row += 1
		}
	}

	return &grid
}

// coordinates returns the left bottom coordinate of an item in the grid
// using a rather eccentric x/y coordinate system with x values showing
// column positions numbered from the left and y values used for row
// positions starting at the top, with the first row reserved for the
// legend.
func (w *weekGrid) coordinates(x, y int) (int, int, error) {
	if x > len(w.colPositions)-1 {
		return 0, 0, fmt.Errorf("x %d y %d : x out of range", x, y)
	}
	if y > len(w.rowPositions)-1 {
		return 0, 0, fmt.Errorf("x %d y %d : y out of range", x, y)
	}
	return w.colPositions[x], w.rowPositions[y], nil
}

// weekCoordinates returns a closure to return each week's coordinates
// in turn.
func (w *weekGrid) weekCoordinates() func() (int, int) {
	x, y := -1, 1
	return func() (int, int) {
		x += 1
		if x == w.columns {
			x = 0
			y++
		}
		wx, wy, err := w.coordinates(x, y)
		if err != nil {
			panic(err)
		}
		return wx, wy
	}
}

type week struct {
	x, y   int // absolute top left coordinate
	monday time.Time
}

func newWeek(x, y int, monday time.Time) *week {
	return &week{x, y, monday}
}

// generally needed const values
const (
	weekNotchHeight  int = 9
	weekNotchSpacing int = 14
	weekLinesLen     int = 98 // px
)

func (w *week) render(svg *svg.SVG) {
	const (
		fontStyle        string = "font-family:sans-serif;font-size:9pt;fill:black;text-anchor:left"
		lineStyle        string = "stroke:%s;stroke-width:%d"
		weekLinesColour  string = "black"
		weekLinesStroke  int    = 2
		weekLinesPadding int    = 16 // px
	)

	var text string
	if w.monday.Day() < 7 {
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
	startDate, endDate time.Time
	colour             string
	strokeWidth        int
	level              int // 0 is first level above week notches, 1 the second
}

func newStripe(typer, colour string, start, end time.Time, width, level int) *stripe {
	return &stripe{typer, start, end, colour, width, level}
}

// Rendering a stripe requires some to split across visual lines. The
// grid function dateSegments returns the slice of general coordinate
// segments for each date say cx and cy. These coordinates (cx,cy)
// should have the week offsets added for the text label and notches
// (see weekNotchHeight, weekNotchSpacing and weekLinesLen). The cx
// value also represents a monday, so if the day of the week is
// not a monday, the stripe should start or end at weekNotchSpacing *
// dayofweek (where the day of the week is an iso week, not golangs).
func (s *stripe) render(svg *svg.SVG) {

}

func main() {

	trips := makeTrips()
	grid := newGrid(trips)

	output, err := os.Create("output.svg")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer output.Close()

	canvas := svg.New(output)
	canvas.Start(grid.width, grid.height)
	background := newContainer("#c4c8b7ff", "#ecececff", 2)
	background.render(grid.width, grid.height, canvas)

	legendX, legendY, err := grid.coordinates(0, 0)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	legend := newLegend(legendX, legendY, []label{
		label{"holidays", "#00d455ff", 5},
		label{"breach", "#ff0000ff", 5},
		label{"longest window without breach", "#0000ffff", 5},
	})
	legend.render(canvas)

	gridGenerator := grid.weekCoordinates()
	for i := range grid.weekNum {
		date := grid.startDate.Add(time.Hour * 24 * 7 * time.Duration(i)) // add a week to the start date
		x, y := gridGenerator()
		week := newWeek(x, y, date)
		week.render(canvas)
	}
	canvas.End()

	fmt.Printf("%#v\n", grid)
	fmt.Printf("trips start : %s end %s\n", trips.Start.Format("Mon 02 Jan 2006"), trips.End.Format("Mon 02 Jan 2006"))
	fmt.Printf("grid  start : %s end %s\n", grid.startDate.Format("Mon 02 Jan 2006"), grid.endDate.Format("Mon 02 Jan 2006"))

}
