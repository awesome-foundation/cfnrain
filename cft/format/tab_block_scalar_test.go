package format_test

import (
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
	"gopkg.in/yaml.v3"
)

// A string that starts with a tab must not be written as a block scalar,
// because yaml.v3 cannot read a tab in the indentation of the first line.
func TestTabLeadingBlockScalarRoundTrips(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	const value = "\tfirst\nsecond\n"
	input := "Description: \"\\tfirst\\nsecond\\n\"\n"
	for _, style := range []string{"", "literal", "folded"} {
		t.Run(style, func(t *testing.T) {
			template, err := parse.String(input)
			if err != nil {
				t.Fatal(err)
			}
			format.NodeStyle = style
			output := format.String(template, format.Options{})
			var decoded map[string]string
			if err := yaml.Unmarshal([]byte(output), &decoded); err != nil {
				t.Fatalf("output is not valid YAML: %v\n%s", err, output)
			}
			if decoded["Description"] != value {
				t.Fatalf("Description = %q, want %q\n%s", decoded["Description"], value, strings.TrimSpace(output))
			}
		})
	}
}
