package fmt

import (
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
)

// Issue #84: a merge key must retain its YAML spelling, and JSON must not
// silently treat it as an ordinary property.
func TestFormatMergeKey(t *testing.T) {
	input := "Metadata:\n  Base: &b\n    x: 1\n  Copy:\n    <<: *b\n    y: 2\n"
	oldJSON, oldStyle := jsonFlag, format.NodeStyle
	t.Cleanup(func() { jsonFlag, format.NodeStyle = oldJSON, oldStyle })
	for _, tc := range []struct {
		name      string
		style     string
		input     string
		json      bool
		want      string
		wantError string
	}{
		{name: "yaml merge", input: input, want: "    <<: *b\n"},
		{name: "yaml flow merge", input: input, style: "flow", want: "<<: *b"},
		{name: "yaml flow quoted key", input: "Metadata:\n  Copy:\n    \"<<\": 1\n", style: "flow", want: `"<<": 1`},
		{name: "json flow quoted key", input: "Metadata:\n  Copy:\n    \"<<\": 1\n", style: "flow", json: true, want: `"<<": 1`},
		{name: "json merge", input: input, json: true, wantError: "YAML merge key <<"},
		{name: "yaml ordinary alias", input: "Metadata:\n  Base: &b\n    x: 1\n  Copy: *b\n", want: "  Copy: *b\n"},
		{name: "json ordinary alias", input: "Metadata:\n  Base: &b\n    x: 1\n  Copy: *b\n", json: true, want: `"Copy": {`},
		{name: "yaml quoted key", input: "Metadata:\n  Copy:\n    \"<<\": 1\n", want: `"<<": 1`},
		{name: "json quoted key", input: "Metadata:\n  Copy:\n    \"<<\": 1\n", json: true, want: `"<<": 1`},
		{name: "json nested merge", input: "Metadata:\n  Items:\n    - <<: {x: 1}\n", json: true, wantError: "YAML merge key <<"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			jsonFlag = tc.json
			format.NodeStyle = tc.style
			var res result
			formatString(tc.input, &res)
			if tc.wantError != "" {
				if res.err == nil || !strings.Contains(res.err.Error(), tc.wantError) {
					t.Fatalf("error = %v, want %q", res.err, tc.wantError)
				}
				if res.output != "" {
					t.Fatalf("unexpected output: %s", res.output)
				}
				return
			}
			if res.err != nil {
				t.Fatal(res.err)
			}
			if !strings.Contains(res.output, tc.want) {
				t.Fatalf("output = %s, want %q", res.output, tc.want)
			}
			if strings.Contains(res.output, "!!merge") {
				t.Fatalf("explicit merge tag in output: %s", res.output)
			}
		})
	}
}
