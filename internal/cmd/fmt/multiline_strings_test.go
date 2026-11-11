package fmt

import (
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"gopkg.in/yaml.v3"
)

func TestStylesPreserveMultilineStrings(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"", "tagged", "flow", "strict-booleans", "original", "singlequoted", "doublequoted", "quotescalars"} {
		values := []string{"first\nsecond", "first\nsecond\n", "first\n  indented\nlast", "it's here\n# text\nend", "first\r\nsecond", "first\u0085second", "first\u2028second", "first\u2029second", "\tfirst\nsecond\n", "\tsingle"}
		if style == "singlequoted" || style == "doublequoted" || style == "quotescalars" {
			values = append(values, "\nfirst\n\nsecond\n\n")
		}
		for _, value := range values {
			t.Run(style+"/"+value, func(t *testing.T) {
				format.NodeStyle = style
				// A quoted scalar works at each depth and as a short function argument.
				scalar := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value, Style: yaml.DoubleQuotedStyle}
				encoded, err := yaml.Marshal(scalar)
				if err != nil {
					t.Fatal(err)
				}
				for _, input := range []string{
					"Description: " + string(encoded),
					"Metadata:\n  Text: " + string(encoded),
					"Outputs:\n  Script:\n    Value: !Sub " + string(encoded),
				} {
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
				}
			})
		}
	}
}

// TestJSONPreservesTabLeadingStrings checks that --json keeps strings that start with a tab.
func TestJSONPreservesTabLeadingStrings(t *testing.T) {
	oldJSON := jsonFlag
	t.Cleanup(func() { jsonFlag = oldJSON })
	jsonFlag = true
	for _, value := range []string{"\tfirst\nsecond\n", "\tsingle"} {
		scalar := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value, Style: yaml.DoubleQuotedStyle}
		encoded, err := yaml.Marshal(scalar)
		if err != nil {
			t.Fatal(err)
		}
		var first, second result
		formatString("Description: "+string(encoded), &first)
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
	}
}
