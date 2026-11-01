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

	// The YAML output keeps the short form
	output = format.String(source, format.Options{Unsorted: true})
	if d := cmp.Diff(strings.TrimSpace(input), strings.TrimSpace(output)); d != "" {
		t.Fatalf("YAML output mismatch (-want +got):\n%s", d)
	}

	if err = parse.Verify(source, output); err != nil {
		t.Fatal(err)
	}
}
