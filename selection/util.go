package selection

import (
	"math"
)

// appendMap adds/increments all m2 elements to m1
func appendMap(m1, m2 map[string]int) {
	for k, v := range m2 {
		n := m1[k]
		m1[k] = v + n
	}
}

func copyMap(m map[string]int) map[string]int {
	res := make(map[string]int)
	for k, v := range m {
		res[k] = v
	}
	return res
}

// fmt.Println(Round(0.363636, 0.1)) // 0.4
func Round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}
