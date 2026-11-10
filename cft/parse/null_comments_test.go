package parse_test

import (
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/parse"
	"gopkg.in/yaml.v3"
)

func TestEmptyNullCommentsPreserveValue(t *testing.T) {
	for _, input := range []string{
		"V: # c\nZ: null\n",
		"V: !!null # c\nZ: null\n",
		"V: !!null \"\" # c\nZ: null\n",
		"V:\n  - # c\n  - null\n",
		"V:\n  - !!null # c\n  - null\n",
		"V:\n  - !!null \"\" # c\n  - null\n",
		"Metadata:\n  Rain:\n    Content: # TODO\n",
	} {
		t.Run(input, func(t *testing.T) {
			tmpl, err := parse.String(input)
			if err != nil {
				t.Fatal(err)
			}
			empty := 0
			var check func(*yaml.Node)
			check = func(n *yaml.Node) {
				if n.Kind == yaml.ScalarNode && n.Tag == "!!null" && n.LineComment != "" {
					empty++
					if n.Value != "" {
						t.Errorf("commented null value = %q, want empty", n.Value)
					}
				}
				for _, child := range n.Content {
					check(child)
				}
			}
			check(tmpl.Node)
			if empty != 1 {
				t.Fatalf("found %d commented nulls, want 1", empty)
			}
		})
	}
}
