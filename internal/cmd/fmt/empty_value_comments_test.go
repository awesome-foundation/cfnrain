package fmt

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
	"gopkg.in/yaml.v3"
)

func TestEmptyValueComments(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"", "original"} {
		for _, tc := range []struct{ name, input, want string }{
			{"flow anchor", "  L: [&q, # c\n    *q]\n", ""},
			{"combined anchors", "  L: [&q, # bad\n    *q]\n  V: &a # c\n  W: *a\n", ""},
			{"anchor", "  V: &a # c\n  W: *a\n", "  V: &a null # c\n\n  W: *a\n"},
			{"string anchor", "  V: &a !!str # c\n  W: *a\n", "  V: &a \"\" # c\n\n  W: *a\n"},
			{"string", "  V: !!str # c\n  W: x\n", "  V: \"\" # c\n\n  W: x\n"},
			{"anchor control", "  V: &a 1 # c\n  W: *a\n", "  V: &a 1 # c\n\n  W: *a\n"},
			{"GetAZs", "  Q: !GetAZs # c\n  R: next\n", "  Q: !GetAZs \"\" # c\n\n  R: next\n"},
			{"GetAZs control", "  Q: !GetAZs '' # c\n  R: next\n", "  Q: !GetAZs \"\" # c\n\n  R: next\n"},
			{"flow", "  Values: {a: , # c\n    b: 1}\n", "  Values:\n    a: # c\n    b: 1\n"},
			{"flow control", "  Values: {a: 1, # c\n    b: 2}\n", "  Values:\n    a: 1 # c\n    b: 2\n"},
			{"Join", "  Q: !Join\n    - '-'\n    - - # c\n      - x\n", "  Q: !Join\n    - '-'\n    - - !!null \"\" # c\n      - x\n"},
			{"Join explicit null", "  Q: !Join\n    - '-'\n    - - null # c\n      - x\n", "  Q: !Join\n    - '-'\n    - - null # c\n      - x\n"},
		} {
			t.Run(style+"/"+tc.name, func(t *testing.T) {
				format.NodeStyle = style
				var first, second result
				formatString("Metadata:\n"+tc.input+"Resources: {}\n", &first)
				if first.err != nil {
					t.Fatal(first.err)
				}
				want := "Metadata:\n" + tc.want + "\nResources: {}\n"
				if tc.name == "flow anchor" || tc.name == "combined anchors" {
					parsed, err := parse.String(first.output)
					if err != nil {
						t.Fatal(err)
					}
					wantComments := map[string]string{"q": "# c"}
					if tc.name == "combined anchors" {
						wantComments = map[string]string{"q": "# bad", "a": "# c"}
					}
					var check func(*yaml.Node)
					check = func(n *yaml.Node) {
						if want, ok := wantComments[n.Anchor]; ok {
							if n.LineComment != want {
								t.Fatalf("anchor %s comment = %q, want %q", n.Anchor, n.LineComment, want)
							}
							delete(wantComments, n.Anchor)
						}
						for _, child := range n.Content {
							check(child)
						}
					}
					check(parsed.Node)
					if len(wantComments) != 0 {
						t.Fatalf("missing anchors: %v", wantComments)
					}
				} else if strings.HasPrefix(tc.name, "flow") {
					parsed, err := parse.String(first.output)
					if err != nil {
						t.Fatal(err)
					}
					var found bool
					var check func(*yaml.Node)
					check = func(n *yaml.Node) {
						if n.Kind == yaml.MappingNode {
							for i := 0; i < len(n.Content); i += 2 {
								if n.Content[i].Value == "a" {
									found = true
									comments := n.Content[i].LineComment + n.Content[i+1].LineComment
									if comments != "# c" {
										t.Fatalf("comment on a = %q", comments)
									}
								}
								if n.Content[i].Value == "b" && (n.Content[i].LineComment != "" || n.Content[i+1].LineComment != "") {
									t.Fatal("comment moved to b")
								}
							}
						}
						for _, child := range n.Content {
							check(child)
						}
					}
					check(parsed.Node)
					if !found {
						t.Fatal("missing a")
					}
				} else if first.output != want {
					t.Fatalf("first output:\n%s\nwant:\n%s", first.output, want)
				}
				formatString(first.output, &second)
				if second.err != nil {
					t.Fatal(second.err)
				}
				if second.output != first.output {
					t.Fatalf("second output changed:\n%s", second.output)
				}
				filename := filepath.Join(t.TempDir(), "first.yaml")
				if err := os.WriteFile(filename, []byte(first.output), 0600); err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestNullVerifyHelper$", "--", "--node-style", style, "--verify", filename)
				cmd.Env = append(os.Environ(), "CFNRAIN_NULL_VERIFY_HELPER=1")
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("fmt --verify: %v\n%s", err, output)
				}
			})
		}
	}
}
