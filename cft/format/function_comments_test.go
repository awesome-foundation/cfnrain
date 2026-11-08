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
	for _, function := range []struct{ name, tag, value string }{
		{"Fn::Sub", "!Sub", "x-${AWS::Region}"},
		{"Ref", "!Ref", "AWS::Region"},
		{"Fn::GetAtt", "!GetAtt", "[Bucket, Arn]"},
	} {
		for _, comment := range []struct{ name, input, output string }{
			{"head", "      # why this name\n      %s: %s\n", "    Value: %s %s\n\n    # why this name\n"},
			{"line", "      %s: # why this name\n        %s\n", "    Value: %s %s # why this name\n"},
			{"foot", "      %s: %s\n      # why this name\n", "    Value: %s %s\n\n    # why this name\n"},
		} {
			t.Run(function.name+"/"+comment.name, func(t *testing.T) {
				input := "Outputs:\n  O:\n    Value:\n" + fmt.Sprintf(comment.input, function.name, function.value)
				// Check that the fixture attaches the comment to the function key.
				var raw yaml.Node
				if err := yaml.Unmarshal([]byte(input), &raw); err != nil {
					t.Fatal(err)
				}
				key := raw.Content[0].Content[1].Content[1].Content[1].Content[0]
				comments := map[string]string{"head": key.HeadComment, "line": key.LineComment, "foot": key.FootComment}
				if comments[comment.name] != "# why this name" {
					t.Fatalf("%s comment is not on the function key: %#v", comment.name, key)
				}
				template, err := parse.String(input)
				if err != nil {
					t.Fatal(err)
				}
				value := function.value
				if function.name == "Fn::GetAtt" {
					value = "Bucket.Arn"
				}
				expected := "Outputs:\n  O:\n" + fmt.Sprintf(comment.output, function.tag, value)
				actual := format.String(template, format.Options{})
				if d := cmp.Diff(expected, actual); d != "" {
					t.Error(d)
				}
				if err := parse.Verify(template, actual); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}
