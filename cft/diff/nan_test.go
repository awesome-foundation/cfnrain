package diff

import (
	"math"
	"testing"
)

func TestCompareNaN(t *testing.T) {
	for _, tc := range []struct {
		name     string
		old, new interface{}
		want     Mode
	}{
		{"same", math.NaN(), math.NaN(), Unchanged},
		{"finite", math.NaN(), 1.0, Changed},
		{"from finite", 1.0, math.NaN(), Changed},
		{"infinity", math.NaN(), math.Inf(1), Changed},
		{"string", math.NaN(), ".nan", Changed},
		{"null", math.NaN(), nil, Changed},
		{"slice", []interface{}{math.NaN()}, []interface{}{math.NaN()}, Unchanged},
		{"map", map[string]interface{}{"v": math.NaN()}, map[string]interface{}{"v": math.NaN()}, Unchanged},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := compareValues(tc.old, tc.new).Mode(); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
