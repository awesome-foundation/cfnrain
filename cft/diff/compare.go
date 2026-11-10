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
				if equalYAMLValues(val, candidate) {
					nanValues = append(nanValues[:i], nanValues[i+1:]...)
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		} else if candidate, ok := new[key]; !ok || !equalYAMLValues(val, candidate) {
			return false
		}
	}
	return true
}

// equalYAMLValues keeps DeepEqual semantics inside non-string-keyed mappings,
// except that YAML NaN values compare as equal. This equality is symmetric, so
// the one-to-one NaN-key match does not depend on map iteration order.
func equalYAMLValues(old, new interface{}) bool {
	if reflect.TypeOf(old) != reflect.TypeOf(new) {
		return false
	}
	switch v := old.(type) {
	case float64:
		n := new.(float64)
		return v == n || (math.IsNaN(v) && math.IsNaN(n))
	case []interface{}:
		n := new.([]interface{})
		if len(v) != len(n) || (v == nil) != (n == nil) {
			return false
		}
		for i, val := range v {
			if !equalYAMLValues(val, n[i]) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		n := new.(map[string]interface{})
		if len(v) != len(n) || (v == nil) != (n == nil) {
			return false
		}
		for key, val := range v {
			if candidate, ok := n[key]; !ok || !equalYAMLValues(val, candidate) {
				return false
			}
		}
		return true
	case map[interface{}]interface{}:
		return equalInterfaceMaps(v, new.(map[interface{}]interface{}))
	default:
		return reflect.DeepEqual(old, new)
	}
}
