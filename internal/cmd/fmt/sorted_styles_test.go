package fmt

import (
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"gopkg.in/yaml.v3"
)

func TestSortedMappingStyles(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"flow", "quotescalars"} {
		for _, section := range []string{"Metadata", "Parameters", "Resources", "Outputs", "Transform"} {
			t.Run(style+"/"+section, func(t *testing.T) {
				output := checkSortedMappingFormat(t, section+":\n  M: {Z: last, Description: first}\n", style)
				var doc yaml.Node
				if err := yaml.Unmarshal([]byte(output), &doc); err != nil {
					t.Fatal(err)
				}
				mapping := doc.Content[0].Content[1].Content[1]
				if mapping.Style&yaml.FlowStyle == 0 {
					t.Fatalf("mapping lost flow style:\n%s", output)
				}
				if section != "Transform" && section != "Resources" && mapping.Content[0].Value != "Description" {
					t.Fatalf("priority key was not sorted:\n%s", output)
				}
				if style == "quotescalars" && mapping.Content[0].Style&yaml.DoubleQuotedStyle == 0 {
					t.Fatalf("key is not quoted:\n%s", output)
				}
			})
		}
	}
}
