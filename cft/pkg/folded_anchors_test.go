package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/parse"
	"github.com/awesome-foundation/cfnrain/internal/node"
	"github.com/awesome-foundation/cfnrain/internal/s11n"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

func TestPackagedFunctionAliases(t *testing.T) {
	for _, tc := range []struct {
		name, expression, want string
	}{
		{"join long", `{Fn::Join: [",", [x, y]]}`, `x,y`},
		{"join short", `!Join [",", [x, y]]`, `x,y`},
		{"select long", `{Fn::Select: [1, [x, y]]}`, `y`},
		{"select short", `!Select [1, [x, y]]`, `y`},
		{"sub long", `{Fn::Sub: '${Const::Text}'}`, `hello`},
		{"sub short", `!Sub '${Const::Text}'`, `hello`},
		{"constant ref", `!Ref Const::Text`, `hello`},
		{"constant directive", `!Rain::Constant Text`, `hello`},
		{"env", `!Rain::Env CFNRAIN_ANCHOR_TEST`, `hello`},
		{"include", `!Rain::Include text.txt`, `hello`},
		{"embed", `!Rain::Embed text.txt`, `hello`},
		{"insert file", `{Fn::InsertFile: text.txt}`, `hello`},
		{"unresolved join", `!Join [",", [!Ref Input, y]]`, `{Fn::Join: [",", [{Ref: Input}, y]]}`},
		{"unfolded split", `!Split [",", "x,y"]`, `{Fn::Split: [",", "x,y"]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CFNRAIN_ANCHOR_TEST", "hello")
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "text.txt"), []byte("hello"), 0600); err != nil {
				t.Fatal(err)
			}
			input := "Constants:\n  Text: hello\nOutputs:\n  Original:\n    Value: &s " + tc.expression + "\n  Alias:\n    Value: *s\n"
			path := filepath.Join(dir, "template.yaml")
			if err := os.WriteFile(path, []byte(input), 0600); err != nil {
				t.Fatal(err)
			}
			got, err := File(path)
			if err != nil {
				t.Fatal(err)
			}
			serialized, err := yaml.Marshal(got.Node)
			if err != nil {
				t.Fatal(err)
			}
			var actual, expected any
			if err := yaml.Unmarshal(serialized, &actual); err != nil {
				t.Fatal(err)
			}
			want := "Outputs:\n  Original:\n    Value: " + tc.want + "\n  Alias:\n    Value: " + tc.want + "\n"
			if err := yaml.Unmarshal([]byte(want), &expected); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestModuleFoldedAnchors(t *testing.T) {
	for _, tc := range []struct {
		name, expression, want string
		resolve                func(*yaml.Node) error
	}{
		{"merge", `{Fn::Merge: [[x], [y]]}`, `[x, y]`, FnMerge},
		{"if true", `!If [Enabled, x, y]`, `x`, func(n *yaml.Node) error {
			m := &Module{ConditionValues: map[string]bool{"Enabled": true}}
			_, err := m.ProcessFnIf(n, nil)
			return err
		}},
		{"if false", `!If [Enabled, x, y]`, `y`, func(n *yaml.Node) error {
			m := &Module{ConditionValues: map[string]bool{"Enabled": false}}
			_, err := m.ProcessFnIf(n, nil)
			return err
		}},
		{"sub", `!Sub hello`, `hello`, func(n *yaml.Node) error {
			return (&Module{}).ResolveSub(n)
		}},
		{"parameter", `!Ref Input`, `hello`, func(n *yaml.Node) error {
			params := node.MakeMapping()
			node.SetMapValue(node.AddMap(params, "Input"), "Default", node.MakeScalar("hello"))
			_, err := (&Module{}).resolveParam(params, n, node.MakeMapping())
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := parse.String("Original: &s " + tc.expression + "\nAlias: *s\n")
			if err != nil {
				t.Fatal(err)
			}
			_, original, _ := s11n.GetMapValue(parsed.Node.Content[0], "Original")
			if err := tc.resolve(original); err != nil {
				t.Fatal(err)
			}
			if original.Anchor != "s" {
				t.Fatalf("anchor = %q, want s", original.Anchor)
			}
			serialized, err := yaml.Marshal(parsed.Node)
			if err != nil {
				t.Fatal(err)
			}
			var got, want any
			if err := yaml.Unmarshal(serialized, &got); err != nil {
				t.Fatal(err)
			}
			if err := yaml.Unmarshal([]byte("Original: "+tc.want+"\nAlias: "+tc.want+"\n"), &want); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestFoldedInnerAnchorError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "template.yaml")
	input := "Outputs:\n  Original:\n    Value: &s !Select [0, [&t x, y]]\n  Outer:\n    Value: *s\n  Inner:\n    Value: *t\n"
	if err := os.WriteFile(path, []byte(input), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := File(path)
	if err == nil || !strings.Contains(err.Error(), `cannot resolve YAML alias "t" after packaging: its anchor is missing`) {
		t.Fatalf("expected an error naming the removed inner anchor t, got %v", err)
	}
}
