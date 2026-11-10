package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
	"gopkg.in/yaml.v3"
)

func TestFormatPreservesEmptyJoinItems(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"", "original"} {
		for _, value := range []string{"", "null", "~", `!!null ""`} {
			t.Run(style+"/"+value, func(t *testing.T) {
				format.NodeStyle = style
				input := "Outputs:\n  Joined:\n    Value: !Join\n      - '-'\n      - - " + value + " # c\n        - x\n"
				source, err := parse.String(input)
				if err != nil {
					t.Fatal(err)
				}
				first := format.String(source, format.Options{})
				next, err := parse.String(first)
				if err != nil {
					t.Fatal(err)
				}
				second := format.String(next, format.Options{})
				if first != second {
					t.Fatalf("second format changed:\n%s", second)
				}
				want := value + "-x"
				if strings.HasPrefix(value, "!!null") {
					want = "-x"
				}
				for _, text := range []string{input, first, second} {
					path := filepath.Join(t.TempDir(), "join.yaml")
					if err := os.WriteFile(path, []byte(text), 0600); err != nil {
						t.Fatal(err)
					}
					packaged, err := File(path)
					if err != nil {
						t.Fatal(err)
					}
					var got struct {
						Outputs map[string]struct {
							Value string `yaml:"Value"`
						} `yaml:"Outputs"`
					}
					data, err := yaml.Marshal(packaged.Node)
					if err != nil {
						t.Fatal(err)
					}
					if err := yaml.Unmarshal(data, &got); err != nil {
						t.Fatal(err)
					}
					if got.Outputs["Joined"].Value != want {
						t.Fatalf("joined value = %q, want %q", got.Outputs["Joined"].Value, want)
					}
				}
			})
		}
	}
}
