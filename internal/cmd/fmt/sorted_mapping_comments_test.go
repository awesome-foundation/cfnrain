package fmt

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"gopkg.in/yaml.v3"
)

func TestSortedMappingComments(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, section := range []string{"Metadata", "Parameters", "Resources", "Outputs", "Transform"} {
		for _, style := range []string{"original", "flow", "quotescalars"} {
			t.Run(section+"/"+style, func(t *testing.T) {
				input := section + ":\n  M: {a: 1} # c\n  E: {} # d\n"
				if section != "Resources" {
					input += "Resources: {}\n"
				}
				output := checkSortedMappingFormat(t, input, style)
				var doc yaml.Node
				if err := yaml.Unmarshal([]byte(output), &doc); err != nil {
					t.Fatal(err)
				}
				root := doc.Content[0]
				for i := 0; i < len(root.Content); i += 2 {
					if root.Content[i].Value != section {
						continue
					}
					mapping := root.Content[i+1]
					for j := 1; j < len(mapping.Content); j += 2 {
						value := mapping.Content[j]
						if value.Style&yaml.FlowStyle == 0 {
							t.Errorf("mapping lost flow style:\n%s", output)
						}
					}
				}
			})
		}
	}
}

func TestEmptySortedMappingComments(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"", "original", "flow", "tagged", "doublequoted", "singlequoted", "quotescalars", "literal", "folded", "strict-booleans"} {
		t.Run(style, func(t *testing.T) {
			checkSortedMappingFormat(t, "Metadata:\n  E: {} # d\nResources: {}\n", style)
		})
	}
}

func TestSortedMappingDefaultStyle(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	output := checkSortedMappingFormat(t, "Metadata:\n  M: {a: 1} # c\n  E: {} # d\nResources: {}\n", "")
	want := "Metadata:\n\n  # c\n  M:\n    a: 1\n\n  E: {} # d\n\nResources: {}\n"
	if output != want {
		t.Fatalf("output:\n%swant:\n%s", output, want)
	}
}

func TestSortedRootMappingComment(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"flow", "quotescalars"} {
		t.Run(style, func(t *testing.T) {
			checkSortedMappingFormat(t, "{Metadata: {M: {a: 1}}, Resources: {}} # c\n", style)
		})
	}
}

func TestSortedMappingStyleCompatibility(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"tagged"} {
		for _, comments := range []bool{false, true} {
			name := style + "/without comments"
			line, rline := "", ""
			if comments {
				name = style + "/key comments"
				line, rline = " # line", " # rline"
			}
			t.Run(name, func(t *testing.T) {
				input := "Metadata:" + line + "\n  M:\n    a: 1\nResources:\n  R:" + rline + "\n    Type: AWS::SNS::Topic\n"
				want := "Metadata:" + line + "\n  M:\n    a: 1\n\nResources:\n  R:" + rline + "\n    Type: AWS::SNS::Topic\n"
				if style == "tagged" {
					want = "!!str Metadata:" + line + "\n  !!str M:\n    !!str a: !!int 1\n\n!!str Resources:\n  !!str R:" + rline + "\n    !!str Type: !!str AWS::SNS::Topic\n"
				}
				output := checkSortedMappingFormat(t, input, style)
				if output != want {
					t.Fatalf("output differs from main:\n%swant:\n%s", output, want)
				}
				// Files formatted by main must also pass --verify unchanged.
				checkSortedMappingFormat(t, want, style)
			})
		}
	}
}

func checkSortedMappingFormat(t *testing.T, input, style string) string {
	t.Helper()
	format.NodeStyle = style
	var first, second result
	formatString(input, &first)
	if first.err != nil {
		t.Fatal(first.err)
	}
	for _, comment := range []string{"# c", "# d"} {
		if strings.Count(first.output, comment) != strings.Count(input, comment) {
			t.Fatalf("comment %q lost or duplicated:\n%s", comment, first.output)
		}
	}
	formatString(first.output, &second)
	if second.err != nil {
		t.Fatal(second.err)
	}
	if first.output != second.output {
		t.Fatalf("second output changed:\n%s\nfirst:\n%s", second.output, first.output)
	}
	filename := filepath.Join(t.TempDir(), "formatted.yaml")
	if err := os.WriteFile(filename, []byte(first.output), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestFlowCollectionVerifyHelper$", "--", "--verify", "--node-style", style, filename)
	cmd.Env = append(os.Environ(), "CFNRAIN_FLOW_VERIFY_HELPER=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fmt --verify: %v\n%s", err, output)
	}
	return first.output
}

func TestSortedFlowKeyComments(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	input := "Metadata: # line\n  M:\n    a: 1\nResources:\n  R: # rline\n    Type: AWS::SNS::Topic\n"
	output := checkSortedMappingFormat(t, input, "flow")
	for _, comment := range []string{"# line", "# rline"} {
		if strings.Count(output, comment) != 1 {
			t.Fatalf("comment lost or duplicated: %s", output)
		}
	}
}
