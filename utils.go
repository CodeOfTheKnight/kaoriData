package kaoriData

import (
	"fmt"
	"strings"
)

func NormalizeEpNumber(eps []float64) (name string) {

	for i, ep := range eps {
		if i != 0 {
			name += "-"
		}
		name += fmt.Sprintf("%.1f", ep)

		tmp := strings.Split(name, ".")
		if tmp[1] == "0" {
			name = tmp[0]
		}
	}

	return name
}
