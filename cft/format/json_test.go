package format_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/awesome-foundation/cfnrain/cft"
	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

// https://github.com/awesome-foundation/cfnrain/issues/63
func TestContainsJson(t *testing.T) {
	input := `
Rules:
  CheckBootstrapVersion:
    Assertions:
      - Assert: !Not
          - !Contains
            - - "1"
              - "2"
            - !Ref BootstrapVersion
`

	source, err := parse.String(input)
	if err != nil {
		t.Fatal(err)
	}

	output := format.String(source, format.Options{
		JSON:     true,
		Unsorted: true,
	})

	expected := `
{
    "Rules": {
        "CheckBootstrapVersion": {
            "Assertions": [
                {
                    "Assert": {
                        "Fn::Not": [
                            {
                                "Fn::Contains": [
                                    [
                                        "1",
                                        "2"
                                    ],
                                    {
                                        "Ref": "BootstrapVersion"
                                    }
                                ]
                            }
                        ]
                    }
                }
            ]
        }
    }
}
`

	if d := cmp.Diff(strings.TrimSpace(expected), strings.TrimSpace(output)); d != "" {
		t.Fatalf("JSON output mismatch (-want +got):\n%s", d)
	}

	// fmt runs this check, so the output must also pass it
	if err = parse.Verify(source, output); err != nil {
		t.Fatal(err)
	}

	// The YAML output keeps the short form that the source used
	output = format.String(source, format.Options{Unsorted: true})
	if d := cmp.Diff(strings.TrimSpace(input), strings.TrimSpace(output)); d != "" {
		t.Fatalf("YAML output mismatch (-want +got):\n%s", d)
	}

	if err = parse.Verify(source, output); err != nil {
		t.Fatal(err)
	}
}

// roundTrip formats input as YAML and as JSON, checks that the YAML output is
// want and that both outputs pass parse.Verify, and returns the template and
// the JSON output
func roundTrip(t *testing.T, input, want string) (*cft.Template, string) {
	t.Helper()

	source, err := parse.String(input)
	if err != nil {
		t.Fatal(err)
	}

	output := format.String(source, format.Options{})
	if d := cmp.Diff(want, output); d != "" {
		t.Errorf("YAML output mismatch (-want +got):\n%s", d)
	}
	if err = parse.Verify(source, output); err != nil {
		t.Errorf("YAML: %s", err)
	}

	// Unsorted output may order keys differently, but must have the same lines
	output = format.String(source, format.Options{Unsorted: true})
	if d := cmp.Diff(sortedLines(want), sortedLines(output)); d != "" {
		t.Errorf("unsorted output has different lines (-want +got):\n%s", d)
	}
	if err = parse.Verify(source, output); err != nil {
		t.Errorf("unsorted YAML: %s", err)
	}

	output = format.String(source, format.Options{JSON: true})
	if err = parse.Verify(source, output); err != nil {
		t.Errorf("JSON: %s", err)
	}

	return source, output
}

func sortedLines(s string) []string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	slices.Sort(lines)
	return lines
}

// https://github.com/awesome-foundation/cfnrain/issues/81
// fmt writes these tags in the form that the source used.
func TestParseOnlyTagsKeepForm(t *testing.T) {
	cases := []struct {
		tag   string
		short string
		long  string
	}{
		{"Contains",
			" !Contains\n      - - a\n      - b",
			"\n      Fn::Contains:\n        - - a\n        - b"},
		{"EachMemberEquals",
			" !EachMemberEquals\n      - - a\n      - b",
			"\n      Fn::EachMemberEquals:\n        - - a\n        - b"},
		{"EachMemberIn",
			" !EachMemberIn\n      - - a\n      - - b",
			"\n      Fn::EachMemberIn:\n        - - a\n        - - b"},
		{"GetStackOutput",
			" !GetStackOutput\n      StackName: s\n      OutputName: o",
			"\n      Fn::GetStackOutput:\n        StackName: s\n        OutputName: o"},
		{"Length",
			" !Length\n      - a\n      - b",
			"\n      Fn::Length:\n        - a\n        - b"},
		{"RefAll",
			" !RefAll AWS::EC2::VPC::Id",
			"\n      Fn::RefAll: AWS::EC2::VPC::Id"},
		{"ToJsonString",
			" !ToJsonString\n      a: b",
			"\n      Fn::ToJsonString:\n        a: b"},
		{"ValueOf",
			" !ValueOf\n      - P\n      - Tags",
			"\n      Fn::ValueOf:\n        - P\n        - Tags"},
		{"ValueOfAll",
			" !ValueOfAll\n      - AWS::EC2::Subnet::Id\n      - VpcId",
			"\n      Fn::ValueOfAll:\n        - AWS::EC2::Subnet::Id\n        - VpcId"},
		{"Transform",
			" !Transform\n      Name: AWS::Include\n      Parameters:\n        Location: s3://b/k",
			"\n      Fn::Transform:\n        Name: AWS::Include\n        Parameters:\n          Location: s3://b/k"},
	}

	for _, c := range cases {
		for form, value := range map[string]string{"short": c.short, "long": c.long} {
			t.Run(c.tag+"/"+form, func(t *testing.T) {
				input := "Outputs:\n  O:\n    Value:" + value + "\n"
				source, output := roundTrip(t, input, input)

				if !strings.Contains(output, `"Fn::`+c.tag+`"`) {
					t.Errorf("JSON output has no Fn::%s:\n%s", c.tag, output)
				}

				value := source.Map()["Outputs"].(map[string]interface{})["O"].(map[string]interface{})["Value"]
				if m, ok := value.(map[string]interface{}); !ok || m["Fn::"+c.tag] == nil {
					t.Errorf("Template.Map() has no Fn::%s: %v", c.tag, value)
				}
			})
		}
	}
}

func TestFunctionForm(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			// Unchanged: fmt writes documented functions short
			name:  "long Fn::Sub",
			input: "Outputs:\n  O:\n    Value:\n      Fn::Sub: ${AWS::StackName}-x\n",
			want:  "Outputs:\n  O:\n    Value: !Sub ${AWS::StackName}-x\n",
		},
		{
			// AWS forbids short functions inside a short !GetStackOutput
			name:  "short GetStackOutput, long Fn::Sub",
			input: "Outputs:\n  O:\n    Value: !GetStackOutput\n      StackName:\n        Fn::Sub: ${AWS::StackName}-x\n      OutputName: o\n",
		},
		{
			name:  "short GetStackOutput, long Ref",
			input: "Outputs:\n  O:\n    Value: !GetStackOutput\n      StackName:\n        Ref: Name\n      OutputName: o\n",
		},
		{
			// AWS forbids this, but fmt keeps the source as it is
			name:  "short GetStackOutput, short Sub",
			input: "Outputs:\n  O:\n    Value: !GetStackOutput\n      StackName: !Sub ${AWS::StackName}-x\n      OutputName: o\n",
		},
		{
			name:  "long GetStackOutput, long Fn::Sub",
			input: "Outputs:\n  O:\n    Value:\n      Fn::GetStackOutput:\n        StackName:\n          Fn::Sub: ${AWS::StackName}-x\n        OutputName: o\n",
			want:  "Outputs:\n  O:\n    Value:\n      Fn::GetStackOutput:\n        StackName: !Sub ${AWS::StackName}-x\n        OutputName: o\n",
		},
		{
			name:  "anchored short Transform",
			input: "Metadata:\n  First: &macro !Transform\n    Name: MyMacro\n    Parameters: {}\n\n  Second: *macro\n",
		},
		{
			name:  "anchored long Transform",
			input: "Metadata:\n  First: &macro\n    Fn::Transform:\n      Name: MyMacro\n\n  Second: *macro\n",
		},
		{
			name:  "anchored short Sub",
			input: "Metadata:\n  First: &a !Sub x\n\n  Second: *a\n",
		},
		{
			name:  "anchored long Fn::Sub",
			input: "Metadata:\n  First: &a\n    Fn::Sub: x\n\n  Second: *a\n",
			want:  "Metadata:\n  First: &a !Sub x\n\n  Second: *a\n",
		},
		{
			name:  "anchored scalar argument",
			input: "Metadata:\n  A:\n    Fn::Sub: &arg x-${AWS::Region}\n\n  B: *arg\n",
		},
		{
			name:  "anchored sequence argument",
			input: "Metadata:\n  A:\n    Fn::GetAtt: &arg\n      - Bucket\n      - Arn\n\n  B: *arg\n",
		},
		{
			name:  "anchored function and argument",
			input: "Metadata:\n  A: &function\n    Fn::Sub: &arg x-${AWS::Region}\n\n  B: *arg\n\n  C: *function\n",
		},
		{
			name:  "anchored mapping",
			input: "Metadata:\n  First: &a\n    x: 1\n\n  Second: *a\n",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.want == "" {
				c.want = c.input
			}
			roundTrip(t, c.input, c.want)
		})
	}
}

// An alias to an anchored short tag means the whole function in JSON too
func TestAnchoredShortTagJson(t *testing.T) {
	_, output := roundTrip(t,
		"Metadata:\n  First: &macro !Transform\n    Name: MyMacro\n\n  Second: *macro\n",
		"Metadata:\n  First: &macro !Transform\n    Name: MyMacro\n\n  Second: *macro\n")

	want := `{
    "Metadata": {
        "First": {
            "Fn::Transform": {
                "Name": "MyMacro"
            }
        },
        "Second": {
            "Fn::Transform": {
                "Name": "MyMacro"
            }
        }
    }
}
`
	if d := cmp.Diff(want, output); d != "" {
		t.Errorf("JSON output mismatch (-want +got):\n%s", d)
	}
}

// A cyclic alias must not make Jsonise recurse forever
func TestJsoniseCyclicAlias(t *testing.T) {
	var n yaml.Node
	if err := yaml.Unmarshal([]byte("Loop: &loop\n  Again: *loop\n"), &n); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() { _ = recover() }()
		format.Jsonise(&n)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Jsonise did not return on a cyclic alias")
	}
}

// Aliases that nest each other must not expand without limit
func TestJsoniseAliasExpansionCap(t *testing.T) {
	input := "a0: &a0 [x]\n"
	for i := 1; i <= 30; i++ {
		input += fmt.Sprintf("a%d: &a%d [*a%d, *a%d]\n", i, i, i-1, i-1)
	}
	var n yaml.Node
	if err := yaml.Unmarshal([]byte(input), &n); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() { _ = recover() }()
		format.Jsonise(&n)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Jsonise did not stop expanding nested aliases")
	}
}
