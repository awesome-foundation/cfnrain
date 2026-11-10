package fmt

import (
	"strings"
	"testing"
)

func TestFormatAliasCycles(t *testing.T) {
	oldJSON := jsonFlag
	t.Cleanup(func() { jsonFlag = oldJSON })
	for _, json := range []bool{false, true} {
		jsonFlag = json
		mode := "yaml"
		if json {
			mode = "json"
		}
		for _, tc := range []struct {
			name, value string
			cyclic      bool
		}{
			{"join", `&c !Join ["", [*c]]`, true},
			{"mapping", `&c {Again: *c}`, true},
			{"list", `&c [*c]`, true},
			{"indirect", `&c {Child: &child [*c]}`, true},
			{"shared mapping", `[&c {x: 1}, *c, *c]`, false},
			{"shared function", `[&c !Join ["", [a, b]], *c]`, false},
		} {
			t.Run(mode+"/"+tc.name, func(t *testing.T) {
				var res result
				formatString("Metadata:\n  Value: "+tc.value+"\n", &res)
				if tc.cyclic {
					if res.err == nil || !strings.Contains(res.err.Error(), `anchor "c" refers to itself`) {
						t.Fatalf("error = %v, want anchor cycle error", res.err)
					}
					if res.output != "" {
						t.Fatalf("unexpected output: %s", res.output)
					}
				} else if res.err != nil || res.output == "" {
					t.Fatalf("non-cyclic alias: error = %v, output = %q", res.err, res.output)
				}
			})
		}
	}
}
