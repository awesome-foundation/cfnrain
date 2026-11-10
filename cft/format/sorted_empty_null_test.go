package format_test

import (
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
)

func TestSortedEmptyNullKeepsBareForm(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"original", "quotescalars", "flow"} {
		t.Run(style, func(t *testing.T) {
			format.NodeStyle = style
			source, err := parse.String("Metadata:\n  X:\nResources: {}\n")
			if err != nil {
				t.Fatal(err)
			}
			want := "Metadata:\n  X:\n\nResources: {}\n"
			if style == "quotescalars" {
				want = "\"Metadata\":\n  \"X\":\n\n\"Resources\": {}\n"
			}
			if style == "flow" {
				want = "{Metadata: {X: !!null \"\"}, Resources: {}}\n"
			}
			first := format.String(source, format.Options{})
			if first != want {
				t.Fatalf("output:\n%s\nwant:\n%s", first, want)
			}
			again, err := parse.String(first)
			if err != nil {
				t.Fatal(err)
			}
			if second := format.String(again, format.Options{}); second != first {
				t.Fatalf("second run changed:\n%s", second)
			}
		})
	}
}
