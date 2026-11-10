package format_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft"
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
				if d := cmp.Diff(want, []string{key.HeadComment, value.LineComment, key.FootComment}); d != "" {
					t.Errorf("formatted property comments: %s\n%s", d, actual)
				}
				if key.LineComment != "" {
					t.Errorf("unexpected property key line comment: %q", key.LineComment)
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

func TestSequenceFunctionCommentsInLists(t *testing.T) {
	for _, tt := range []struct{ name, input, expected string }{
		{
			name: "join with line and foot comments",
			input: `Resources:
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      Tags:
        - Fn::Join: # c
            - '-'
            - [a, b]
          # foot of join item
        - Key: k
          Value: v
`,
			expected: `Resources:
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      Tags:
        - !Join

          # c
          - '-'
          - - a
            - b

            # foot of join item
        - Key: k
          Value: v
`,
		},
		{
			name: "split inside select",
			input: `Resources:
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName:
        Fn::Select:
          - 0
          - Fn::Split: # beside split
              - ','
              - a,b
      Other: z
`,
			expected: `Resources:
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: !Select
        - 0
        - !Split

          # beside split
          - ','
          - a,b
      Other: z
`,
		},
		{
			name: "without comments",
			input: `Resources:
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      Tags:
        - Fn::Join:
            - '-'
            - [a, b]
        - Key: k
          Value: v
`,
			expected: `Resources:
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      Tags:
        - !Join
          - '-'
          - - a
            - b
        - Key: k
          Value: v
`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			template, err := parse.String(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			actual := format.String(template, format.Options{})
			if d := cmp.Diff(tt.expected, actual); d != "" {
				t.Errorf("first format: %s", d)
			}
			checkCommentRoundTrip(t, template, actual)
		})
	}
}

func TestShortFunctionCombinedLineComments(t *testing.T) {
	for _, function := range []struct{ name, value string }{
		{"Fn::Sub", "x-${AWS::Region}"},
		{"Ref", "AWS::Region"},
		{"Fn::Join", "['-', [a, b]]"},
		{"Fn::If", "[Condition, a, b]"},
	} {
		for _, comments := range []struct{ parent, argument string }{
			{"", " # argument"},
			{" # property", ""},
			{" # property", " # argument"},
		} {
			t.Run(function.name+comments.parent+comments.argument, func(t *testing.T) {
				input := fmt.Sprintf("Outputs:\n  O:\n    Value:%s\n      %s: # function\n        %s%s\n    Export:\n      Name: n\n", comments.parent, function.name, function.value, comments.argument)
				template, err := parse.String(input)
				if err != nil {
					t.Fatal(err)
				}
				actual := format.String(template, format.Options{})
				var raw yaml.Node
				if err := yaml.Unmarshal([]byte(actual), &raw); err != nil {
					t.Fatal(err)
				}
				property := raw.Content[0].Content[1].Content[1]
				want := "# function"
				if comments.argument != "" {
					want += "\n# argument"
				}
				if comments.parent != "" {
					want = "# property\n" + want
				}
				if d := cmp.Diff(want, property.Content[0].HeadComment); d != "" {
					t.Errorf("property comments: %s\n%s", d, actual)
				}
				for _, n := range property.Content[1:] {
					if n.HeadComment != "" || n.LineComment != "" || n.FootComment != "" {
						t.Errorf("comment escaped property head: %#v", n)
					}
				}
				checkCommentRoundTrip(t, template, actual)
			})
		}
	}
}

func checkCommentRoundTrip(t *testing.T, template *cft.Template, actual string) {
	t.Helper()
	if err := parse.Verify(template, actual); err != nil {
		t.Fatalf("%v\n%s", err, actual)
	}
	again, err := parse.String(actual)
	if err != nil {
		t.Fatal(err)
	}
	if d := cmp.Diff(actual, format.String(again, format.Options{})); d != "" {
		t.Errorf("second format: %s", d)
	}
}

func TestScalarFunctionComments(t *testing.T) {
	for _, function := range []struct{ name, input, expected string }{
		{"empty GetAZs", `Fn::GetAZs: ""`, `!GetAZs ""`},
		{"single quoted GetAZs", "Fn::GetAZs: ''", `!GetAZs ""`},
		{"short GetAZs", `!GetAZs ""`, `!GetAZs ""`},
		{"empty Base64", `Fn::Base64: ""`, `!Base64 ""`},
		{"Base64", "Fn::Base64: x", "!Base64 x"},
		{"Ref", "!Ref AWS::Region", "!Ref AWS::Region"},
		{"ImportValue", "Fn::ImportValue: x", "!ImportValue x"},
	} {
		for _, context := range []struct{ name, input, expected string }{
			{"list line comment", "    Value:\n      - %s # c\n      - next\n", "    Value:\n      - %s # c\n      - next\n"},
			{"last item", "    Value:\n      - previous\n      - %s # c\n", "    Value:\n      - previous\n      - %s # c\n"},
			{"list head comment", "    Value:\n      # head\n      - %s\n      - next\n", "    Value:\n\n      # head\n      - %s\n      - next\n"},
			{"mapping head comment", "    Value:\n      # head\n      %s\n    Export:\n      Name: n\n", "\n    # head\n    Value: %s\n    Export:\n      Name: n\n"},
			{"mapping value", "    Value:\n      %s # c\n    Export:\n      Name: n\n", "    Value: %s # c\n    Export:\n      Name: n\n"},
		} {
			t.Run(function.name+"/"+context.name, func(t *testing.T) {
				input := "Outputs:\n  Example:\n" + fmt.Sprintf(context.input, function.input)
				argument := function.expected
				if context.name == "mapping head comment" {
					argument = strings.TrimSuffix(argument, ` ""`)
				}
				expected := "Outputs:\n  Example:\n" + fmt.Sprintf(context.expected, argument)
				template, err := parse.String(input)
				if err != nil {
					t.Fatal(err)
				}
				actual := format.String(template, format.Options{})
				if d := cmp.Diff(expected, actual); d != "" {
					t.Errorf("first format: %s", d)
				}
				checkCommentRoundTrip(t, template, actual)
			})
		}
	}
}
