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
  Strings:
    - '0'
    - '1.0'
    - 'true'
    - 'null'
    - text
Outputs:
  Selected:
    Value: !Select
      - 0
      - [a, b]
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

func TestStylesPreserveFunctionScalarTypes(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"", "tagged", "flow", "strict-booleans", "original", "singlequoted", "doublequoted", "quotescalars"} {
		for _, argument := range []string{"!GetAZs ~", "!GetAZs null", "!Base64 true", "!Ref 1", "!Ref 1.5", "!Ref '1'", "!Ref '1.5'", "!Ref '~'", "!Ref 'false'", "!Ref '.inf'", "!Ref '2026-01-01'", "!Base64 'true'", "!GetAZs 'null'", "!GetAZs ''", "!GetAtt Bucket.Arn"} {
			t.Run(style+"/"+argument, func(t *testing.T) {
				format.NodeStyle = style
				var first, second result
				formatString("Outputs:\n  Test:\n    Value: "+argument+"\n", &first)
				if first.err != nil {
					t.Fatal(first.err)
				}
				formatString(first.output, &second)
				if second.err != nil {
					t.Fatal(second.err)
				}
				if !second.ok {
					t.Fatalf("output would reformat:\n%s", first.output)
				}
			})
		}
	}
}

func TestQuoteScalarsPreservesFlowCollections(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	format.NodeStyle = "quotescalars"
	const input = `Metadata:
  Joined: !Join ["", [a, !Ref AWS::Region]]
  Tags: [{Key: k, Value: v}]
`
	var first, second result
	formatString(input, &first)
	if first.err != nil {
		t.Fatal(first.err)
	}
	for _, want := range []string{`!Join ["", ["a", !Ref "AWS::Region"]]`, `[{"Key": "k", "Value": "v"}]`} {
		if !strings.Contains(first.output, want) {
			t.Errorf("missing flow collection %s in:\n%s", want, first.output)
		}
	}
	formatString(first.output, &second)
	if second.err != nil {
		t.Fatal(second.err)
	}
	if !second.ok {
		t.Fatalf("output would reformat:\n%s", first.output)
	}
}
