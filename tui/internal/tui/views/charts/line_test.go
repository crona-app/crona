package charts

import (
	"math"
	"testing"
)

func TestLineCanvasSpansWidthAndGapsOnNaN(t *testing.T) {
	c := NewLineCanvas(10, 2)
	c.Plot([]float64{1, 5, math.NaN(), 1}, 1, 5, 0)
	if c.Cell(0, 4) != 0 || c.Cell(3, 4) != 9 {
		t.Fatalf("expected the ends on the first and last cells, got %d and %d", c.Cell(0, 4), c.Cell(3, 4))
	}
	if c.Rune(1, 0) == 0 || c.Rune(0, c.Cell(1, 4)) == 0 {
		t.Fatalf("expected the first point low and the second high")
	}
	// No segment joins the second point to the fourth across the gap.
	for cell := c.Cell(1, 4) + 1; cell < c.Cell(3, 4); cell++ {
		if c.Rune(0, cell) != 0 || c.Rune(1, cell) != 0 {
			t.Fatalf("expected a gap at cell %d", cell)
		}
	}
}
