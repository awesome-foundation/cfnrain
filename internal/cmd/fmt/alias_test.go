package fmt

import (
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/parse"
	"github.com/google/go-cmp/cmp"
)

func TestFormatAnchoredFunctionAlias(t *testing.T) {
	oldJSON := jsonFlag
	t.Cleanup(func() { jsonFlag = oldJSON })
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
			for _, mode := range []string{"yaml", "json"} {
				t.Run(mode, func(t *testing.T) {
					jsonFlag = mode == "json"
					var res result
					formatString(input, &res)
					if res.err != nil {
						t.Fatal(res.err)
					}
					got, err := parse.String(res.output)
					if err != nil {
						t.Fatal(err)
					}
					if d := cmp.Diff(want.Map(), got.Map()); d != "" {
						t.Fatalf("function alias changed (-want +got):\n%s", d)
					}
				})
			}
		})
	}
}
