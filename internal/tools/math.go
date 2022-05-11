package tools

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"math"
)

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Humanize(v float64) string {
	switch {
	case v > math.Pow10(6):
		return fmt.Sprintf("%.1fM", v/math.Pow10(6))
	case v > math.Pow10(3):
		return fmt.Sprintf("%.1fK", v/math.Pow10(3))
	case v > math.Pow10(2):
		return fmt.Sprintf("%.0f", v/math.Pow10(3))
	case v == math.Round(v):
		return fmt.Sprintf("%.0f", v)
	case v < 1:
		return fmt.Sprintf("%.3f", v)
	default:
		return fmt.Sprintf("%.1f", v)
	}
}
