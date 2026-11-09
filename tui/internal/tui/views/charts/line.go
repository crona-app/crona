package charts

import "math"

// dotBits are a braille cell's dots by column (0 left, 1 right) and row (0 top … 3 bottom).
var dotBits = [2][4]rune{{0x01, 0x02, 0x04, 0x40}, {0x08, 0x10, 0x20, 0x80}}

// LineCanvas draws series as braille lines on width×rows cells, two dots across and four
// down per cell. Owner records which series last drew in a cell, for its colour.
type LineCanvas struct {
	Width, Rows int
	bits        [][]rune
	Owner       [][]int
}

func NewLineCanvas(width, rows int) *LineCanvas {
	c := &LineCanvas{Width: max(1, width), Rows: max(1, rows)}
	c.bits = make([][]rune, c.Rows)
	c.Owner = make([][]int, c.Rows)
	for r := range c.bits {
		c.bits[r] = make([]rune, c.Width)
		c.Owner[r] = make([]int, c.Width)
		for x := range c.Owner[r] {
			c.Owner[r][x] = -1
		}
	}
	return c
}

// PointX is the dot column of point i of n spread across the canvas, ends included.
func (c *LineCanvas) PointX(i, n int) int { return pointX(i, n, c.Width) }

// Cell is the cell column holding point i of n.
func (c *LineCanvas) Cell(i, n int) int { return PointCell(i, n, c.Width) }

// PointCell is the cell column holding point i of n across width cells, as a LineCanvas of
// that width places it; labels under a chart line up with it this way.
func PointCell(i, n, width int) int { return pointX(i, n, max(1, width)) / 2 }

func pointX(i, n, width int) int {
	dots := width * 2
	if n <= 1 {
		return dots / 2
	}
	return int(math.Round(float64(i) * float64(dots-1) / float64(n-1)))
}

func (c *LineCanvas) set(x, y, owner int) {
	if x < 0 || y < 0 || x >= c.Width*2 || y >= c.Rows*4 {
		return
	}
	row, cell := y/4, x/2
	c.bits[row][cell] |= dotBits[x%2][y%4]
	c.Owner[row][cell] = owner
}

// dotY is value's dot row from the top, lo at the bottom and hi at the top.
func (c *LineCanvas) dotY(value, lo, hi float64) int {
	span := hi - lo
	if span <= 0 {
		span = 1
	}
	frac := min(1, max(0, (value-lo)/span))
	return c.Rows*4 - 1 - int(math.Round(frac*float64(c.Rows*4-1)))
}

// Plot draws values (NaN = no value, a gap in the line) as a line between lo and hi, its
// points spread across the width, and marks owner in the cells it passes.
func (c *LineCanvas) Plot(values []float64, lo, hi float64, owner int) {
	px, py, have := 0, 0, false
	for i, v := range values {
		if math.IsNaN(v) {
			have = false
			continue
		}
		x, y := c.PointX(i, len(values)), c.dotY(v, lo, hi)
		if have {
			c.segment(px, py, x, y, owner)
		} else {
			c.set(x, y, owner)
		}
		px, py, have = x, y, true
	}
}

// segment joins two dots (Bresenham).
func (c *LineCanvas) segment(x0, y0, x1, y1, owner int) {
	dx, dy := abs(x1-x0), -abs(y1-y0)
	sx, sy := sign(x1-x0), sign(y1-y0)
	err := dx + dy
	for {
		c.set(x0, y0, owner)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

// Rune is the braille character of a cell, or 0 when it is empty.
func (c *LineCanvas) Rune(row, cell int) rune {
	if c.bits[row][cell] == 0 {
		return 0
	}
	return 0x2800 + c.bits[row][cell]
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func sign(v int) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}

// Dots marks each value (NaN skipped) as a single dot, unjoined.
func (c *LineCanvas) Dots(values []float64, lo, hi float64, owner int) {
	for i, v := range values {
		if !math.IsNaN(v) {
			c.set(c.PointX(i, len(values)), c.dotY(v, lo, hi), owner)
		}
	}
}

// Row is the cell row holding value.
func (c *LineCanvas) Row(value, lo, hi float64) int { return c.dotY(value, lo, hi) / 4 }
