package fmt

import (
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
)

func TestQuotedStylesPreserveScalarTypes(t *testing.T) {
	const input = `Metadata:
  Integer: 1
  Float: 1.5
  Boolean: true
  NullValue: null
  Strings: ['0', '1.0', 'true', 'null', text]
Outputs:
  Selected:
    Value: !Select [0, [a, b]]
  Reference:
    Value: !Ref Name
`
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"singlequoted", "doublequoted", "quotescalars"} {
		t.Run(style, func(t *testing.T) {
			format.NodeStyle = style
			var first result
			formatString(input, &first)
			if first.err != nil {
				t.Fatal(first.err)
			}
			quote := `"`
			if style == "singlequoted" {
				quote = "'"
			}
			for _, scalar := range []string{"0", "1.0", "true", "null", "text"} {
				if !strings.Contains(first.output, "- "+quote+scalar+quote+"\n") {
					t.Errorf("missing quoted string %q in:\n%s", scalar, first.output)
				}
			}
			for _, field := range []string{"Integer" + quote + ": 1\n", "Float" + quote + ": 1.5\n", "Boolean" + quote + ": true\n", "NullValue" + quote + ": null\n", "- 0\n"} {
				if !strings.Contains(first.output, field) {
					t.Errorf("missing plain scalar %q in:\n%s", field, first.output)
				}
			}
			var second result
			formatString(first.output, &second)
			if second.err != nil {
				t.Fatal(second.err)
			}
			// --verify uses ok to decide whether the formatted file passes.
			if !second.ok || second.output != first.output {
				t.Fatalf("second format changed the output:\n%s", second.output)
			}
		})
	}
}
