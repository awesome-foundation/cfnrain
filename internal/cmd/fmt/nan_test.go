package fmt

import (
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
)

func TestStylesPreserveNaN(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"", "tagged", "flow", "strict-booleans", "original", "singlequoted", "doublequoted", "quotescalars"} {
		for _, value := range []string{".nan", ".NaN", ".NAN"} {
			t.Run(style+"/"+value, func(t *testing.T) {
				format.NodeStyle = style
				input := "Metadata:\n  Number: " + value + "\n  List: [" + value + "]\n  Function: !Ref " + value + "\n"
				var first, second result
				formatString(input, &first)
				if first.err != nil {
					t.Fatal(first.err)
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
