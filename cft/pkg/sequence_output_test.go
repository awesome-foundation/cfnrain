package pkg_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft"
	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
	"github.com/awesome-foundation/cfnrain/cft/pkg"
	"gopkg.in/yaml.v3"
)

func TestSequenceOutputDeterministic(t *testing.T) {
	var first string
	for i := 0; i < 50; i++ {
		packaged, err := pkg.File("./tmpl/sequence-output-template.yaml")
		if err != nil {
			t.Fatal(err)
		}
		got := format.String(packaged, format.Options{Unsorted: true})
		if i == 0 {
			first = got
		} else if got != first {
			t.Fatalf("packaging output changed on run %d:\n%s", i, got)
		}
	}
}

func TestModuleOutputResolutionErrors(t *testing.T) {
	for _, tc := range []struct {
		name, value, properties, want string
	}{
		{"missing parameter", "!Ref Unset", "", "did not find Unset"},
		{"join parameter", "!Sub '${P}'", "      P: !Join ['-', [a, b]]\n", "expected Sub reference"},
		{"sequence Sub parameter", "!Sub '${P}-suffix'", "      P: !Sub ['${X}', {X: value}]\n", "module M parameter P uses sequence-form Sub"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			module := "Parameters:\n  Unset:\n    Type: String\n  P:\n    Type: String\nResources:\n  Bucket:\n    Type: AWS::S3::Bucket\nOutputs:\n  Bad:\n    Value: " + tc.value + "\n  Good:\n    Value: !Ref Bucket\n"
			if err := os.WriteFile(filepath.Join(dir, "module.yaml"), []byte(module), 0600); err != nil {
				t.Fatal(err)
			}
			base := "Modules:\n  M:\n    Source: ./module.yaml\n"
			if tc.properties != "" {
				base += "    Properties:\n" + tc.properties
			}
			for _, use := range []string{"!GetAtt M.Good", "!GetAtt M.Bad", "!Sub '${M.Bad}'"} {
				t.Run(use, func(t *testing.T) {
					parent, err := parse.String(base + "Outputs:\n  Result:\n    Value: " + use + "\n")
					if err != nil {
						t.Fatal(err)
					}
					_, err = pkg.Template(parent, dir, nil)
					if use == "!GetAtt M.Good" {
						if err != nil {
							t.Fatalf("unused output must not fail packaging: %v", err)
						}
					} else if err == nil || !strings.Contains(err.Error(), tc.want) || !strings.Contains(err.Error(), "M.Bad") {
						t.Fatalf("expected error naming M.Bad and %q, got %v", tc.want, err)
					}
				})
			}
		})
	}
}

func TestSequenceOutputPreservesParentNode(t *testing.T) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("Fn::Sub: ['${M.Out}-${Local}', {Local: suffix}] # keep this\n"), &doc); err != nil {
		t.Fatal(err)
	}
	n := doc.Content[0]
	seq := n.Content[1]
	style, comment := seq.Style, seq.LineComment
	module := &pkg.Module{Config: &cft.ModuleConfig{Name: "M"}}
	output := map[string]any{"Value": map[string]any{"Fn::Sub": []any{"${X}", map[string]any{"X": "value"}}}}
	if err := module.OutputSub("Out", output, n); err != nil {
		t.Fatal(err)
	}
	if n.Content[1] != seq || seq.Style != style || seq.LineComment != comment {
		t.Fatal("parent sequence node, style, or comment changed")
	}
}

func TestMalformedParentSubLocation(t *testing.T) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("Fn::Sub: [unrelated]\n"), &doc); err != nil {
		t.Fatal(err)
	}
	module := &pkg.Module{Config: &cft.ModuleConfig{Name: "M"}}
	err := module.OutputSub("Out", nil, doc.Content[0])
	if err == nil || !strings.Contains(err.Error(), "line 1, column 10") || strings.Contains(err.Error(), "M.Out") {
		t.Fatalf("expected location of malformed Sub without unrelated output name, got %v", err)
	}
}
