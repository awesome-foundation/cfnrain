package diff

import (
	"math"
	"reflect"

	"github.com/awesome-foundation/cfnrain/cft"
)

// New returns a Diff that represents the difference between two templates
func New(a, b *cft.Template) Diff {
	return CompareMaps(a.Map(), b.Map())
}

func compareValues(old, new interface{}) Diff {
	if reflect.TypeOf(old) != reflect.TypeOf(new) {

		// In YAML there is no difference between "" and null
		if old == "" && new == nil {
			return value{new, Unchanged}
		}

		return value{new, Changed}
	}

	// YAML has one NaN value. IEEE NaN inequality is not a template change.
	if v, ok := old.(float64); ok && math.IsNaN(v) && math.IsNaN(new.(float64)) {
		return value{old, Unchanged}
	}

	switch v := old.(type) {
	case []interface{}:
		return compareSlices(v, new.([]interface{}))
	case map[string]interface{}:
		return CompareMaps(v, new.(map[string]interface{}))
	case map[interface{}]interface{}:
		if !equalInterfaceMaps(v, new.(map[interface{}]interface{})) {
			return value{new, Changed}
		}
	default:
		if !reflect.DeepEqual(old, new) {
			return value{new, Changed}
		}
	}

	return value{old, Unchanged}
}

func compareSlices(old, new []interface{}) Diff {
	max := int(math.Max(float64(len(old)), float64(len(new))))
	d := make(slice, max)

	for i := 0; i < max; i++ {
		if i >= len(old) {
			d[i] = value{new[i], Added}
		} else if i >= len(new) {
			d[i] = value{old[i], Removed}
		} else {
			d[i] = compareValues(old[i], new[i])
		}
	}

	return d
}

func CompareMaps(old, new map[string]interface{}) Diff {
	d := make(dmap)

	// New and updated keys
	for key, val := range new {
		if _, ok := old[key]; !ok {
			d[key] = value{val, Added}
		} else {
			d[key] = compareValues(old[key], val)
		}
	}

	// Removed keys
	for key, val := range old {
		if _, ok := new[key]; !ok {
			d[key] = value{val, Removed}
		}
	}

	return d
}

// equalInterfaceMaps compares YAML mappings whose keys are not all strings.
func equalInterfaceMaps(old, new map[interface{}]interface{}) bool {
	if len(old) != len(new) || (old == nil) != (new == nil) {
		return false
	}
	// NaN keys cannot be looked up, even with the same key variable.
	// Keep their values separately and match each entry only once.
	var nanValues []interface{}
	for key, val := range new {
		if number, ok := key.(float64); ok && math.IsNaN(number) {
			nanValues = append(nanValues, val)
		}
	}
	for key, val := range old {
		if number, ok := key.(float64); ok && math.IsNaN(number) {
			matched := false
			for i, candidate := range nanValues {
				if compareValues(val, candidate).Mode() == Unchanged {
					nanValues = append(nanValues[:i], nanValues[i+1:]...)
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		} else if candidate, ok := new[key]; !ok || compareValues(val, candidate).Mode() != Unchanged {
			return false
		}
	}
	return true
}
