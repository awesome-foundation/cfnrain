package fmt

import (
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
)

func TestBlockStylesPreserveScalarTypes(t *testing.T) {
	const input = `Metadata:
  Integer: 1
  Float: 1.5
  Boolean: true
  NullValue: null
  Infinity: .inf
  Strings: ['1', 'true', 'null', text]
  List: [1, 1.5, true, null]
Outputs:
  Number:
    Value: !Ref 1
  Boolean:
    Value: !Base64 true
  NullValue:
    Value: !GetAZs null
  Empty:
    Value: !GetAZs ''
`
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"", "tagged", "flow", "strict-booleans", "original", "singlequoted", "doublequoted", "quotescalars", "literal", "folded"} {
		t.Run(style, func(t *testing.T) {
			format.NodeStyle = style
			var first, second result
			formatString(input, &first)
			if first.err != nil {
				t.Fatal(first.err)
			}
			formatString(first.output, &second)
			if second.err != nil {
				t.Fatal(second.err)
			}
			if !second.ok || first.output != second.output {
				t.Fatalf("output would reformat:\n%s", second.output)
			}
		})
	}
}

func TestBlockStylesPreserveFunctionStrings(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"literal", "folded"} {
		for _, argument := range []string{"!Ref '1'", "!Base64 'true'", "!GetAZs 'null'", `!Sub "first\nsecond\n"`} {
			t.Run(style+"/"+argument, func(t *testing.T) {
				format.NodeStyle = style
				var first, second result
				formatString("Outputs:\n  Value:\n    Value: "+argument+"\n", &first)
				if first.err != nil {
					t.Fatal(first.err)
				}
				formatString(first.output, &second)
				if second.err != nil {
					t.Fatal(second.err)
				}
				if !second.ok || first.output != second.output {
					t.Fatalf("output would reformat:\n%s", second.output)
				}
			})
		}
	}
}
