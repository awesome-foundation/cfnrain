package pkg_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/parse"
	"github.com/awesome-foundation/cfnrain/cft/pkg"
	"github.com/google/go-cmp/cmp"
)

func TestAnchoredFunctionAlias(t *testing.T) {
	want, err := parse.String(`Outputs:
  Zone:
    Value:
      Fn::Select: [0, {Fn::GetAZs: ""}]
  Zones:
    Value:
      Fn::GetAZs: ""
`)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name     string
		function string
	}{
		{"short form", `&a !GetAZs ""`},
		{"long form", `&a {Fn::GetAZs: ""}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := "Outputs:\n  Zone:\n    Value: !Select [0, " + tc.function + "]\n  Zones:\n    Value: *a\n"
			path := filepath.Join(t.TempDir(), "template.yaml")
			if err := os.WriteFile(path, []byte(input), 0600); err != nil {
				t.Fatal(err)
			}
			got, err := pkg.File(path)
			if err != nil {
				t.Fatal(err)
			}
			if d := cmp.Diff(want.Map(), got.Map()); d != "" {
				t.Fatalf("function alias changed (-want +got):\n%s", d)
			}
		})
	}
}
