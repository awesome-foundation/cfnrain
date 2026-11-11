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
		{"join parameter", "!Sub '${P}'", "      P: !Join ['-', [a, b]]\n", "expected Sub reference"},
		{"sequence Sub parameter", "!Sub '${P}-suffix'", "      P: !Sub ['${X}', {X: value}]\n", "module M parameter P uses sequence-form Sub"},
		{"nested sequence Sub parameter", "!Join ['', [!Select [0, [!Sub '${P}-suffix']], !Ref Bucket]]", "      P: !Sub ['${X}', {X: value}]\n", "module M parameter P uses sequence-form Sub"},
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
			for _, use := range []string{"!GetAtt M.Good", "!GetAtt M.Bad", "!Sub '${M.Bad}'", "!Join ['', [!Select [0, [!GetAtt M.Bad]], !Sub '${AWS::Region}']]"} {
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

func TestMissingModuleParameterParentScope(t *testing.T) {
	for _, value := range []string{"!Ref Unset", "!Sub '${Unset}-suffix'", "!Sub ['${Unset}-suffix', {}]"} {
		for _, parent := range []struct {
			name, declarations string
			fallback           bool
		}{
			{"parameter", "Parameters:\n  Unset:\n    Type: String\n", true},
			{"absent", "", false},
			{"resource", "Resources:\n  Unset:\n    Type: AWS::SNS::Topic\n", false},
		} {
			for _, use := range []string{"resource", "GetAtt", "Sub", "unused"} {
				t.Run(value+"/"+parent.name+"/"+use, func(t *testing.T) {
					dir := t.TempDir()
					module := "Parameters:\n  Unset:\n    Type: String\n"
					if use == "resource" {
						module += "Resources:\n  Example:\n    Type: Custom::Example\n    Properties:\n      Value: " + value + "\n"
					} else {
						module += "Resources:\n  Example:\n    Type: Custom::Example\nOutputs:\n  Out:\n    Value: " + value + "\n"
					}
					if err := os.WriteFile(filepath.Join(dir, "module.yaml"), []byte(module), 0600); err != nil {
						t.Fatal(err)
					}
					input := parent.declarations + "Modules:\n  M:\n    Source: ./module.yaml\n"
					if use == "GetAtt" {
						input += "Outputs:\n  Result:\n    Value: !GetAtt M.Out\n"
					} else if use == "Sub" {
						input += "Outputs:\n  Result:\n    Value: !Sub '${M.Out}-outer'\n"
					}
					tmpl, err := parse.String(input)
					if err != nil {
						t.Fatal(err)
					}
					got, err := pkg.Template(tmpl, dir, nil)
					if !parent.fallback && use != "unused" {
						if err == nil || !strings.Contains(err.Error(), "did not find Unset in parent template Properties") {
							t.Fatalf("expected missing parameter error, got %v", err)
						}
						if use != "resource" && !strings.Contains(err.Error(), "module output M.Out") {
							t.Fatalf("expected output name in error, got %v", err)
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					if use != "unused" {
						output := format.String(got, format.Options{})
						want := "${Unset}"
						if value == "!Ref Unset" && use != "Sub" {
							want = "!Ref Unset"
						}
						if !strings.Contains(output, want) || strings.Contains(output, "MUnset") {
							t.Fatalf("expected reference to parent parameter %q, got:\n%s", want, output)
						}
					}
				})
			}
		}
	}
}

func TestSequenceSubFunctionOutput(t *testing.T) {
	for _, tc := range []struct {
		name, value, resolved string
	}{
		{"Join", "!Join ['-', [a, !Ref Bucket]]", "!Join ['-', [a, !Ref MBucket]]"},
		{"Select", "!Select [0, !GetAZs '']", "!Select [0, !GetAZs '']"},
		{"If", "!Ref P", "!If [C, a, b]"},
		{"List", "[a, b]", ""},
	} {
		for _, collision := range []bool{false, true} {
			name := tc.name
			if collision {
				name += "/collision"
			}
			t.Run(name, func(t *testing.T) {
				dir := t.TempDir()
				module := "Parameters: {P: {Type: String}}\nResources:\n  Bucket: {Type: 'AWS::S3::Bucket'}\nOutputs:\n  " + tc.name + ": {Value: " + tc.value + "}\n"
				if err := os.WriteFile(filepath.Join(dir, "module.yaml"), []byte(module), 0600); err != nil {
					t.Fatal(err)
				}
				binding := "RainOutputM_" + tc.name
				variables := "Y: y"
				if collision {
					variables += ", " + binding + ": existing"
					binding += "_1"
				}
				conditions := "Conditions: {C: !Equals [a, b]}\n"
				input := conditions + "Modules:\n  M:\n    Source: ./module.yaml\n    Properties: {P: !If [C, a, b]}\nOutputs:\n  Result:\n    Value: !Sub ['${M." + tc.name + "}-${Y}', {" + variables + "}]\n  ScalarControl:\n    Value: !Sub '${M." + tc.name + "}-x'\n"
				tmpl, err := parse.String(input)
				if err != nil {
					t.Fatal(err)
				}
				got, err := pkg.Template(tmpl, dir, nil)
				if tc.name == "List" {
					if err == nil || !strings.Contains(err.Error(), "module output M.List is a list and cannot be interpolated into a Sub string") {
						t.Fatalf("expected error naming list output M.List, got %v", err)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				expected, err := parse.String(conditions + "Resources:\n  MBucket: {Type: 'AWS::S3::Bucket'}\nOutputs:\n  Result:\n    Value: !Sub ['${" + binding + "}-${Y}', {" + variables + ", " + binding + ": " + tc.resolved + "}]\n  ScalarControl: {Value: '-x'}\n")
				if err != nil {
					t.Fatal(err)
				}
				want := format.String(expected, format.Options{})
				if actual := format.String(got, format.Options{}); actual != want {
					t.Fatalf("expected:\n%s\ngot:\n%s", want, actual)
				}
			})
		}
	}
}
