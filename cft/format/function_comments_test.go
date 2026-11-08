package format_test

import (
	"fmt"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

func TestShortFunctionKeyComments(t *testing.T) {
	for _, function := range []struct{ name, value string }{
		{"Fn::Sub", "x-${AWS::Region}"},
		{"Ref", "AWS::Region"},
		{"Fn::GetAtt", "[Bucket, Arn]"},
		{"Fn::Join", "['-', [a, b]]"},
		{"Fn::If", "[Condition, a, b]"},
	} {
		for _, comment := range []struct{ name, input, head, line, foot string }{
			{name: "head", input: "      ### banner\n      # second line\n      %s: %s\n", head: "### banner\n# second line"},
			{name: "line", input: "      %s: # beside\n        %s\n", line: "# beside"},
			{name: "foot", input: "      %s: %s\n      # after\n\n", foot: "# after"},
			{name: "all", input: "      ### banner\n      # second line\n      %s: # beside\n        %s\n      # after\n\n", head: "### banner\n# second line", line: "# beside", foot: "# after"},
		} {
			t.Run(function.name+"/"+comment.name, func(t *testing.T) {
				input := "Outputs:\n  O:\n    Value:\n" + fmt.Sprintf(comment.input, function.name, function.value) + "    Export:\n      Name: n\n"
				var raw yaml.Node
				if err := yaml.Unmarshal([]byte(input), &raw); err != nil {
					t.Fatal(err)
				}
				property := raw.Content[0].Content[1].Content[1]
				key := property.Content[1].Content[0]
				want := []string{comment.head, comment.line, comment.foot}
				if d := cmp.Diff(want, []string{key.HeadComment, key.LineComment, key.FootComment}); d != "" {
					t.Fatalf("fixture comments: %s", d)
				}
				template, err := parse.String(input)
				if err != nil {
					t.Fatal(err)
				}
				// Sequence line comments move above the containing property.
				if (function.name == "Fn::Join" || function.name == "Fn::If") && comment.line != "" {
					if want[0] != "" {
						want[0] += "\n"
					}
					want[0] += comment.line
					want[1] = ""
				}
				actual := format.String(template, format.Options{})
				if err := parse.Verify(template, actual); err != nil {
					t.Fatalf("%v\n%s", err, actual)
				}
				if err := yaml.Unmarshal([]byte(actual), &raw); err != nil {
					t.Fatal(err)
				}
				property = raw.Content[0].Content[1].Content[1]
				key, value, next := property.Content[0], property.Content[1], property.Content[2]
				if d := cmp.Diff(want, []string{key.HeadComment, key.LineComment + value.LineComment, key.FootComment}); d != "" {
					t.Errorf("formatted property comments: %s\n%s", d, actual)
				}
				if next.HeadComment != "" || next.FootComment != "" || next.LineComment != "" {
					t.Errorf("comment moved to Export: %#v", next)
				}
				again, err := parse.String(actual)
				if err != nil {
					t.Fatal(err)
				}
				if d := cmp.Diff(actual, format.String(again, format.Options{})); d != "" {
					t.Errorf("second format: %s", d)
				}
			})
		}
	}
}
