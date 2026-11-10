package fmt

import (
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
	"gopkg.in/yaml.v3"
)

func TestEmptyFlowNull(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"original", "quotescalars", "flow"} {
		for _, comment := range []string{"", " # c\n       "} {
			for _, location := range []string{"Metadata:\n  M: ", "Resources:\n  B:\n    Type: AWS::S3::Bucket\n    Metadata:\n      M: "} {
				t.Run(style+"/"+comment+"/"+location, func(t *testing.T) {
					format.NodeStyle = style
					input := location + "{a: ," + comment + " b: 1}\n"
					var first, second result
					formatString(input, &first)
					if first.err != nil {
						t.Fatal(first.err)
					}
					formatString(first.output, &second)
					if second.err != nil {
						t.Fatal(second.err)
					}
					if first.output != second.output || !second.ok {
						t.Fatalf("--verify would reformat:\n%s", second.output)
					}
					parsed, err := parse.String(first.output)
					if err != nil {
						t.Fatal(err)
					}
					var check func(*yaml.Node)
					check = func(n *yaml.Node) {
						if n.Kind == yaml.MappingNode {
							for i := 0; i < len(n.Content); i += 2 {
								if n.Content[i].Value == "a" && n.Content[i+1].Tag != "!!null" {
									t.Fatal("empty value lost null type")
								}
							}
						}
						for _, child := range n.Content {
							check(child)
						}
					}
					check(parsed.Node)
					if comment != "" && strings.Count(first.output, "# c") != 1 {
						t.Fatalf("lost comment:\n%s", first.output)
					}
				})
			}
		}
	}
}
