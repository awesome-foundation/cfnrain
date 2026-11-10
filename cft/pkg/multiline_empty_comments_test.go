package pkg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
	"github.com/google/go-cmp/cmp"
)

func TestMultilineEmptyComments(t *testing.T) {
	for _, body := range []string{
		"V: !Sub # c\n  \"\"\n",
		"V: !Sub # c\n  ''\n",
		"V: !Sub # c\n  |\n",
		"V: !Sub # c\n  >\n",
		"V:\n  - !Ref # c\n    \"\"\n",
		"V: &a !GetAZs # c\n  \"\"\nW: *a\n",
		"V: !!null # c\n  \"\"\n",
	} {
		t.Run(body, func(t *testing.T) {
			source, err := parse.String(body)
			if err != nil {
				t.Fatal(err)
			}
			first := format.String(source, format.Options{})
			if err := parse.Verify(source, first); err != nil {
				t.Fatal(err)
			}
			again, err := parse.String(first)
			if err != nil {
				t.Fatal(err)
			}
			if second := format.String(again, format.Options{}); second != first {
				t.Fatalf("second run changed:\n%s", second)
			}
			json := format.String(source, format.Options{JSON: true})
			if err := parse.Verify(source, json); err != nil {
				t.Fatal(err)
			}
			var want any
			for i, text := range []string{body, first} {
				path := filepath.Join(t.TempDir(), "empty.yaml")
				if err := os.WriteFile(path, []byte(text), 0600); err != nil {
					t.Fatal(err)
				}
				packaged, err := File(path)
				if err != nil {
					t.Fatal(err)
				}
				var got any
				if err := packaged.Node.Decode(&got); err != nil {
					t.Fatal(err)
				}
				if i == 0 {
					want = got
				} else if diff := cmp.Diff(want, got); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}
