package fmt

import (
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"gopkg.in/yaml.v3"
)

func TestSortedMappingStyles(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"flow", "quotescalars"} {
		for _, section := range []string{"Metadata", "Parameters", "Resources", "Outputs", "Transform"} {
			t.Run(style+"/"+section, func(t *testing.T) {
				output := checkSortedMappingFormat(t, section+":\n  M: {Z: last, Description: first}\n", style)
				var doc yaml.Node
				if err := yaml.Unmarshal([]byte(output), &doc); err != nil {
					t.Fatal(err)
				}
				mapping := doc.Content[0].Content[1].Content[1]
				if mapping.Style&yaml.FlowStyle == 0 {
					t.Fatalf("mapping lost flow style:\n%s", output)
				}
				if section != "Transform" && section != "Resources" && mapping.Content[0].Value != "Description" {
					t.Fatalf("priority key was not sorted:\n%s", output)
				}
				if style == "quotescalars" && mapping.Content[0].Style&yaml.DoubleQuotedStyle == 0 {
					t.Fatalf("key is not quoted:\n%s", output)
				}
			})
		}
	}
}

func TestSortedMappingExactOutput(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, tc := range []struct{ name, style, input, want string }{
		{"flow root", "flow", "Resources: {}\nMetadata:\n  M: {a: 1}\n", "{Metadata: {M: {a: 1}}, Resources: {}}\n"},
		{"quoted block root", "quotescalars", "Resources: {}\nMetadata:\n  M: {a: 1}\n", "\"Metadata\":\n  \"M\": {\"a\": 1}\n\n\"Resources\": {}\n"},
		{"quoted flow root", "quotescalars", "{Resources: {}, Metadata: {M: {a: 1}}}\n", "{\"Metadata\": {\"M\": {\"a\": 1}}, \"Resources\": {}}\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := checkSortedMappingFormat(t, tc.input, tc.style); got != tc.want {
				t.Fatalf("output:\n%s\nwant:\n%s", got, tc.want)
			}
		})
	}
}

func TestSortedFlowEmptyValues(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, tc := range []struct{ name, style, input, want string }{
		{"properties", "flow", "Resources:\n  R:\n    Properties:\n", "{Resources: {R: {Properties: !!null \"\"}}}\n"},
		{"default", "flow", "Parameters:\n  P:\n    Default:\n", "{Parameters: {P: {Default: !!null \"\"}}}\n"},
		{"description", "flow", "Description:\nResources: {}\n", "{Description: !!null \"\", Resources: {}}\n"},
		{"quoted empty value", "quotescalars", "Metadata:\n  M: {a: }\n", "\"Metadata\":\n  \"M\": {\"a\": !!null \"\"}\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := checkSortedMappingFormat(t, tc.input, tc.style); got != tc.want {
				t.Fatalf("output:\n%s\nwant:\n%s", got, tc.want)
			}
		})
	}
}

func TestSortedQuotedFlowKeyComments(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, input := range []string{
		"{\n? Metadata # c\n: {x: y}\n}\n",
		"Metadata: {\n? M # c\n: {x: y}\n}\n",
	} {
		t.Run(input, func(t *testing.T) { checkSortedMappingFormat(t, input, "quotescalars") })
	}
}
