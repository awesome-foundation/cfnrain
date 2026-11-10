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
		{"numeric empty to null", map[interface{}]interface{}{80: ""}, map[interface{}]interface{}{80: nil}, Changed},
		{"numeric null to empty", map[interface{}]interface{}{80: nil}, map[interface{}]interface{}{80: ""}, Changed},
		{"nested empty to null", map[interface{}]interface{}{80: map[string]interface{}{"v": []interface{}{""}}}, map[interface{}]interface{}{80: map[string]interface{}{"v": []interface{}{nil}}}, Changed},
		{"nil to empty list", map[interface{}]interface{}{80: []interface{}(nil)}, map[interface{}]interface{}{80: []interface{}{}}, Changed},
		{"nil to empty mapping", map[interface{}]interface{}{80: map[string]interface{}(nil)}, map[interface{}]interface{}{80: map[string]interface{}{}}, Changed},
		{"numeric key", map[interface{}]interface{}{80: math.NaN()}, map[interface{}]interface{}{80: math.NaN()}, Unchanged},
		{"nested list", map[interface{}]interface{}{80: []interface{}{math.NaN()}}, map[interface{}]interface{}{80: []interface{}{math.NaN()}}, Unchanged},
		{"numeric key change", map[interface{}]interface{}{80: math.NaN()}, map[interface{}]interface{}{81: math.NaN()}, Changed},
		{"key type change", map[interface{}]interface{}{80: math.NaN()}, map[interface{}]interface{}{"80": math.NaN()}, Changed},
		{"value change", map[interface{}]interface{}{80: math.NaN()}, map[interface{}]interface{}{80: 1.0}, Changed},
		{"NaN key", map[interface{}]interface{}{math.NaN(): 1}, map[interface{}]interface{}{math.NaN(): 1}, Unchanged},
		{"NaN key value change", map[interface{}]interface{}{math.NaN(): 1}, map[interface{}]interface{}{math.NaN(): 2}, Changed},
		{"NaN key removed", map[interface{}]interface{}{math.NaN(): 1}, map[interface{}]interface{}{}, Changed},
		{"NaN key added", map[interface{}]interface{}{}, map[interface{}]interface{}{math.NaN(): 1}, Changed},
		{"multiple NaN keys", map[interface{}]interface{}{math.NaN(): 1, math.NaN(): 2}, map[interface{}]interface{}{math.NaN(): 2, math.NaN(): 1}, Unchanged},
		{"NaN keys cannot reuse a match", map[interface{}]interface{}{math.NaN(): 1, math.NaN(): 1}, map[interface{}]interface{}{math.NaN(): 1, math.NaN(): 2}, Changed},
		{"map", map[string]interface{}{"v": math.NaN()}, map[string]interface{}{"v": math.NaN()}, Unchanged},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := compareValues(tc.old, tc.new).Mode(); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNaNKeyMatchingIsIndependentOfMapOrder(t *testing.T) {
	for i := 0; i < 1000; i++ {
		a := map[interface{}]interface{}{math.NaN(): "", math.NaN(): nil}
		b := map[interface{}]interface{}{math.NaN(): nil, math.NaN(): ""}
		if compareValues(a, b).Mode() != Unchanged || compareValues(b, a).Mode() != Unchanged {
			t.Fatal("NaN key matching changed with map iteration order")
		}
	}
}
