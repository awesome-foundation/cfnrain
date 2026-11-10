package pkg_test

import (
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
