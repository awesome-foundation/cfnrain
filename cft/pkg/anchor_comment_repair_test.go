package pkg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
	"github.com/google/go-cmp/cmp"
)

func TestAnchorCommentRepairPreservesValues(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"", "original"} {
		for _, tc := range []struct{ name, input, want, packaged string }{
			{"join", "Resources:\n  T:\n    Type: AWS::SNS::Topic\n    Properties:\n      TopicName: !Join [\"-\", [a, &q, # c\n        ]]\n",
				`{"Resources":{"T":{"Type":"AWS::SNS::Topic","Properties":{"TopicName":{"Fn::Join":["-",["a",null]]}}}}}`,
				`{"Resources":{"T":{"Type":"AWS::SNS::Topic","Properties":{"TopicName":"a-"}}}}`},
			{"sequence and alias", "Metadata: {L: [x, &q, # c\n ], W: *q}\n",
				`{"Metadata":{"L":["x",null],"W":null}}`, `{"Metadata":{"L":["x",""],"W":""}}`},
			{"mapping", "M: {a: &q, # c\n }\n", `{"M":{"a":null}}`, `{"M":{"a":""}}`},
			{"alias item", "L: [&q, # c\n *q]\n", `{"L":[null,null]}`, `{"L":["",""]}`},
			{"separate properties", "L:\n - !Ref # c\n   &x\n - *x\n", `{"L":[{"Ref":""},{"Ref":""}]}`, ""},
		} {
			t.Run(style+"/"+tc.name, func(t *testing.T) {
				format.NodeStyle = style
				checkJSON := func(output, expected string) {
					t.Helper()
					var got, want any
					if err := json.Unmarshal([]byte(output), &got); err != nil {
						t.Fatal(err)
					}
					if err := json.Unmarshal([]byte(expected), &want); err != nil {
						t.Fatal(err)
					}
					if diff := cmp.Diff(want, got); diff != "" {
						t.Fatal(diff)
					}
				}
				source, err := parse.String(tc.input)
				if err != nil {
					t.Fatal(err)
				}
				checkJSON(format.String(source, format.Options{JSON: true}), tc.want)
				first := format.String(source, format.Options{})
				again, err := parse.String(first)
				if err != nil {
					t.Fatal(err)
				}
				checkJSON(format.String(again, format.Options{JSON: true}), tc.want)
				if second := format.String(again, format.Options{}); second != first {
					t.Fatalf("second format changed:\n%s", second)
				}
				wantPackaged := tc.packaged
				if wantPackaged == "" {
					wantPackaged = tc.want
				}
				inputs := []string{tc.input}
				if tc.name == "join" {
					inputs = append(inputs, first)
				}
				for _, input := range inputs {
					path := filepath.Join(t.TempDir(), "anchor.yaml")
					if err := os.WriteFile(path, []byte(input), 0600); err != nil {
						t.Fatal(err)
					}
					packaged, err := File(path)
					if err != nil {
						t.Fatal(err)
					}
					checkJSON(format.String(packaged, format.Options{JSON: true}), wantPackaged)
				}
			})
		}
	}
}
