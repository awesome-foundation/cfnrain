package fmt

import (
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
)

func TestStylesPreserveNaN(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"", "tagged", "flow", "strict-booleans", "original", "singlequoted", "doublequoted", "quotescalars"} {
		for _, value := range []string{".nan", ".NaN", ".NAN", "!!float .nan"} {
			t.Run(style+"/"+value, func(t *testing.T) {
				format.NodeStyle = style
				input := "Metadata:\n  Number: " + value + "\n  List: [" + value + "]\n  Function: {Ref: " + value + "}\n  Ports: {80: " + value + ", 81: [" + value + "]}\n  Keys: {" + value + ": 1}\n"
				var first, second result
				formatString(input, &first)
				if first.err != nil {
					t.Fatal(first.err)
				}
				spelling := strings.TrimPrefix(value, "!!float ")
				if strings.Count(first.output, spelling) != 6 {
					t.Fatalf("NaN spelling %q changed:\n%s", spelling, first.output)
				}
				if style == "" && value == ".nan" {
					var simple result
					formatString("Metadata:\n  Number: .nan\n", &simple)
					if simple.err != nil || simple.output != "Metadata:\n  Number: .nan\n" {
						t.Fatalf("default output changed: %q (%v)", simple.output, simple.err)
					}
				}
				formatString(first.output, &second)
				if second.err != nil {
					t.Fatal(second.err)
				}
				if !second.ok || second.output != first.output {
					t.Fatalf("output would reformat:\n%s", second.output)
				}
			})
		}
	}
}
