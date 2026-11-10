package fmt

import (
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
)

func TestEmptyFunctionCommentsVerify(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"", "singlequoted"} {
		for _, tc := range []struct{ name, input string }{
			{"mapping head", "Outputs:\n  Example:\n    Value:\n      # head\n      Fn::GetAZs: \"\"\n    Export:\n      Name: n\n"},
			{"list head", "Outputs:\n  Example:\n    Value:\n      # head\n      - Fn::GetAZs: \"\"\n      - next\n"},
			{"mapping line", "Outputs:\n  Example:\n    Value:\n      Fn::GetAZs: \"\" # c\n    Export:\n      Name: n\n"},
		} {
			t.Run(style+"/"+tc.name, func(t *testing.T) {
				format.NodeStyle = style
				var first, second result
				formatString(tc.input, &first)
				if first.err != nil {
					t.Fatal(first.err)
				}
				formatString(first.output, &second)
				if second.err != nil {
					t.Fatal(second.err)
				}
				if first.output != second.output {
					t.Fatalf("format changed on second run:\nfirst:\n%ssecond:\n%s", first.output, second.output)
				}
				// The --verify command uses ok to select its message and exit status.
				if !second.ok {
					t.Fatalf("--verify would reformat first output:\n%s", first.output)
				}
				if style == "singlequoted" && !strings.Contains(first.output, "!GetAZs ''") {
					t.Fatalf("empty argument did not retain single-quoted style:\n%s", first.output)
				}
			})
		}
	}
}
