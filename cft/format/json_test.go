package format_test

import (
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
	"github.com/google/go-cmp/cmp"
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

	// The YAML output uses the long form, because AWS documents no !Contains short form
	expected = `
Rules:
  CheckBootstrapVersion:
    Assertions:
      - Assert: !Not
          - Fn::Contains:
              - - "1"
                - "2"
              - !Ref BootstrapVersion
`
	output = format.String(source, format.Options{Unsorted: true})
	if d := cmp.Diff(strings.TrimSpace(expected), strings.TrimSpace(output)); d != "" {
		t.Fatalf("YAML output mismatch (-want +got):\n%s", d)
	}

	if err = parse.Verify(source, output); err != nil {
		t.Fatal(err)
	}
}

// https://github.com/awesome-foundation/cfnrain/issues/81
func TestShortTagsExpand(t *testing.T) {
	cases := []struct {
		tag string
		arg string
	}{
		{"Contains", "[[a], b]"},
		{"EachMemberEquals", "[[a], b]"},
		{"EachMemberIn", "[[a], [b]]"},
		{"GetStackOutput", "{StackName: s, OutputName: o}"},
		{"Length", "[a, b]"},
		{"RefAll", "AWS::EC2::VPC::Id"},
		{"ToJsonString", "{a: b}"},
		{"ValueOf", "[P, Tags]"},
		{"ValueOfAll", "[AWS::EC2::Subnet::Id, VpcId]"},
		{"Transform", "{Name: AWS::Include, Parameters: {Location: s3://b/k}}"},
		// AWS forbids the !GetStackOutput short form around !Sub
		{"GetStackOutput", "{StackName: !Sub \"${AWS::StackName}-x\", OutputName: o}"},
	}

	for _, c := range cases {
		t.Run(c.tag, func(t *testing.T) {
			source, err := parse.String("Outputs:\n  O:\n    Value: !" + c.tag + " " + c.arg + "\n")
			if err != nil {
				t.Fatal(err)
			}

			value := source.Map()["Outputs"].(map[string]interface{})["O"].(map[string]interface{})["Value"]
			if m, ok := value.(map[string]interface{}); !ok || m["Fn::"+c.tag] == nil {
				t.Errorf("Template.Map() has no Fn::%s: %v", c.tag, value)
			}

			output := format.String(source, format.Options{JSON: true})
			if !strings.Contains(output, `"Fn::`+c.tag+`"`) {
				t.Errorf("JSON output has no Fn::%s:\n%s", c.tag, output)
			}
			if err = parse.Verify(source, output); err != nil {
				t.Errorf("JSON: %s", err)
			}

			// fmt writes the long form, because none of these short forms is safe to write
			output = format.String(source, format.Options{})
			if strings.Contains(output, "!"+c.tag) {
				t.Errorf("YAML output has the short form !%s:\n%s", c.tag, output)
			}
			if err = parse.Verify(source, output); err != nil {
				t.Errorf("YAML: %s", err)
			}
		})
	}
}

// The output for an anchored short tag must not parse back to a different
// meaning. fmt drops anchors on mappings (also on main), so the output may fail
// to parse, but the anchor must not move to the argument, where an alias that
// meant the whole function would silently mean only its argument.
func TestAnchoredShortTag(t *testing.T) {
	source, err := parse.String(`
Metadata:
  First: &macro !Transform
    Name: MyMacro
    Parameters: {}
  Second: *macro
`)
	if err != nil {
		t.Fatal(err)
	}

	output := format.String(source, format.Options{})
	reparsed, err := parse.String(output)
	if err != nil {
		return
	}
	if d := cmp.Diff(source.Map(), reparsed.Map()); d != "" {
		t.Errorf("output changes the meaning (-source +output):\n%s\n%s", d, output)
	}
}
