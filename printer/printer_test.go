package printer

import (
	"math"
	"testing"
)

func TestLinearDistance(t *testing.T) {
	p := Printer{}

	if p.linearDistance(1, 1) != math.Sqrt(2) {
		t.Error("wrong distance")
	}
	if p.linearDistance(0, 1) != 1 {
		t.Error("wrong distance")
	}
}
