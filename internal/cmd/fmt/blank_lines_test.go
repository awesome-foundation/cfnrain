package fmt

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
	"gopkg.in/yaml.v3"
)

func TestMultilineBlankLines(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	cases := []struct{ name, input, value string }{
		{"trailing", `"first\nsecond\n\n\n"` + "\n", "first\nsecond\n\n\n"},
		{"leading", `"\n\nfirst\nsecond"` + "\n", "\n\nfirst\nsecond"},
		{"tabs", `"a\tb\nc\t\n"` + "\n", "a\tb\nc\t\n"},
		{"join", `!Join ["\n", [a, b]]` + "\n", "\n"},
		{"keep", "|+\n    first\n    second\n\n\n\n", "first\nsecond\n\n\n\n"},
		{"literal", "|\n    first\n    second\n", "first\nsecond\n"},
		{"folded", ">\n    first\n    second\n", "first second\n"},
		{"single line", "text\n", "text"},
	}
	for _, style := range []string{"", "tagged", "flow", "strict-booleans", "original"} {
		for _, tc := range cases {
			for _, suffix := range []string{"", "  Z: end\n"} {
				t.Run(style+"/"+tc.name+"/"+suffix, func(t *testing.T) {
					format.NodeStyle = style
					input := "Metadata:\n  Text: " + tc.input + suffix
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
						t.Fatalf("second format changed output:\n%q\n%q", first.output, second.output)
					}
					for _, output := range []string{first.output, second.output} {
						parsed, err := parse.String(output)
						if err != nil {
							t.Fatal(err)
						}
						var findText func(*yaml.Node) *yaml.Node
						findText = func(n *yaml.Node) *yaml.Node {
							if n.Kind == yaml.MappingNode {
								for i := 0; i < len(n.Content); i += 2 {
									if n.Content[i].Value == "Text" {
										return n.Content[i+1]
									}
								}
							}
							for _, child := range n.Content {
								if found := findText(child); found != nil {
									return found
								}
							}
							return nil
						}
						value := findText(parsed.Node)
						if value == nil {
							t.Fatal("missing Text")
						}
						if tc.name == "join" {
							value = value.Content[1].Content[0]
						}
						if value.Value != tc.value {
							t.Fatalf("value = %q, want %q", value.Value, tc.value)
						}
					}
					filename := filepath.Join(t.TempDir(), "formatted.yaml")
					if err := os.WriteFile(filename, []byte(first.output), 0600); err != nil {
						t.Fatal(err)
					}
					cmd := exec.Command(os.Args[0], "-test.run=^TestMultilineBlankLinesVerifyHelper$", "--", "--verify", "--node-style="+style, filename)
					cmd.Env = append(os.Environ(), "CFNRAIN_MULTILINE_VERIFY_HELPER=1")
					if output, err := cmd.CombinedOutput(); err != nil {
						t.Fatalf("fmt --verify: %v\n%s", err, output)
					}
				})
			}
		}
	}
}

func TestMultilineBlankLinesVerifyHelper(t *testing.T) {
	if os.Getenv("CFNRAIN_MULTILINE_VERIFY_HELPER") != "1" {
		return
	}
	if err := Cmd.ParseFlags(os.Args[3:]); err != nil {
		t.Fatal(err)
	}
	Cmd.Run(Cmd, Cmd.Flags().Args())
}
