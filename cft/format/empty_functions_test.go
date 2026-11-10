package format_test

import (
	"encoding/json"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
	"github.com/google/go-cmp/cmp"
)

func TestGetAZsJSON(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		want  any
		yaml  string
	}{
		{"single quotes", "!GetAZs ''", "", "!GetAZs"},
		{"double quotes", "!GetAZs \"\"", "", "!GetAZs"},
		{"long form", "{Fn::GetAZs: \"\"}", "", "!GetAZs"},
		{"bare", "!GetAZs", "", "!GetAZs"},
		{"region ref", "!GetAZs\n        Ref: AWS::Region", map[string]any{"Ref": "AWS::Region"}, "!GetAZs\n        Ref: AWS::Region"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := "Outputs:\n  Zone:\n    Value: !Select\n      - 0\n      - " + tc.input + "\n"
			template, err := parse.String(input)
			if err != nil {
				t.Fatal(err)
			}
			yamlOutput := format.String(template, format.Options{})
			wantYAML := "Outputs:\n  Zone:\n    Value: !Select\n      - 0\n      - " + tc.yaml + "\n"
			if d := cmp.Diff(wantYAML, yamlOutput); d != "" {
				t.Fatalf("YAML changed (-want +got):\n%s", d)
			}
			for _, source := range []string{input, yamlOutput} {
				template, err := parse.String(source)
				if err != nil {
					t.Fatal(err)
				}
				output := format.String(template, format.Options{JSON: true})
				var got any
				if err := json.Unmarshal([]byte(output), &got); err != nil {
					t.Fatal(err)
				}
				want := map[string]any{"Outputs": map[string]any{"Zone": map[string]any{
					"Value": map[string]any{"Fn::Select": []any{float64(0), map[string]any{"Fn::GetAZs": tc.want}}},
				}}}
				if d := cmp.Diff(want, got); d != "" {
					t.Errorf("JSON mismatch (-want +got):\n%s", d)
				}
			}
		})
	}
}

func TestEmptyFunctionArgumentsJSON(t *testing.T) {
	for _, tc := range []struct {
		tag string
		key string
	}{
		{"!Ref", "Ref"},
		{"!Sub", "Fn::Sub"},
		{"!Base64", "Fn::Base64"},
	} {
		for _, suffix := range []string{" ''", "", " ~", " null"} {
			t.Run(tc.tag+suffix, func(t *testing.T) {
				template, err := parse.String("Outputs:\n  Result:\n    Value: " + tc.tag + suffix + "\n")
				if err != nil {
					t.Fatal(err)
				}
				output := format.String(template, format.Options{JSON: true})
				var got any
				if err := json.Unmarshal([]byte(output), &got); err != nil {
					t.Fatal(err)
				}
				var argument any = ""
				if suffix == " ~" || suffix == " null" {
					argument = nil
				}
				want := map[string]any{"Outputs": map[string]any{"Result": map[string]any{
					"Value": map[string]any{tc.key: argument},
				}}}
				if d := cmp.Diff(want, got); d != "" {
					t.Errorf("JSON mismatch (-want +got):\n%s", d)
				}
			})
		}
	}
}
