package pkg_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/pkg"
)

func TestBareParseOnlyFunctionOutput(t *testing.T) {
	for _, name := range []string{"Contains", "Length"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "template.yaml")
			input := "Outputs:\n  Result:\n    Value: !" + name + "\n"
			if err := os.WriteFile(path, []byte(input), 0600); err != nil {
				t.Fatal(err)
			}
			template, err := pkg.File(path)
			if err != nil {
				t.Fatal(err)
			}
			want := "Outputs:\n  Result:\n    Value:\n      Fn::" + name + ": \"\"\n"
			if got := format.String(template, format.Options{}); got != want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

func TestBareFunctionErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		message string
	}{
		{"Join", "should be a Sequence: Fn::Join:"},
		{"Select", "expected Fn::Select to have 2 elements: Fn::Select:"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "template.yaml")
			input := "Outputs:\n  Result:\n    Value: !" + tc.name + "\n"
			if err := os.WriteFile(path, []byte(input), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := pkg.File(path)
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.HasSuffix(err.Error(), tc.message) {
				t.Fatalf("got %q, want error ending in %q", err, tc.message)
			}
		})
	}
}

func TestEmptySubOutput(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{"single quoted", "!Sub ''", "!Sub"},
		{"double quoted", "!Sub \"\"", "!Sub"},
		{"bare", "!Sub", "!Sub"},
		{"long form", "{Fn::Sub: \"\"}", "!Sub"},
		{"empty constant", "!Sub ${Rain::X}", "!Sub"},
		{"non-empty control", "!Sub hello", "hello"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "template.yaml")
			input := "Rain:\n  Constants:\n    X: ''\nOutputs:\n  Result:\n    Value: " + tc.value + "\n"
			if err := os.WriteFile(path, []byte(input), 0600); err != nil {
				t.Fatal(err)
			}
			template, err := pkg.File(path)
			if err != nil {
				t.Fatal(err)
			}
			want := "Outputs:\n  Result:\n    Value: " + tc.want + "\n"
			if got := format.String(template, format.Options{}); got != want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, want)
			}
			var output struct {
				Outputs map[string]struct {
					Value interface{}
				}
			}
			if err := json.Unmarshal([]byte(format.String(template, format.Options{JSON: true})), &output); err != nil {
				t.Fatal(err)
			}
			value := output.Outputs["Result"].Value
			if tc.want == "hello" {
				if value != "hello" {
					t.Fatalf("got JSON value %#v, want hello", value)
				}
			} else {
				function, ok := value.(map[string]interface{})
				if !ok || len(function) != 1 || function["Fn::Sub"] != "" {
					t.Fatalf("got JSON value %#v, want Fn::Sub with an empty string", value)
				}
			}
		})
	}
}
