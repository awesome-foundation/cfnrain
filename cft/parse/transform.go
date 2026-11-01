package parse

import (
	"errors"
	"maps"
	"strings"

	"github.com/awesome-foundation/cfnrain/cft"
	"github.com/awesome-foundation/cfnrain/internal/node"
	"gopkg.in/yaml.v3"
)

// Convert string GetAtt into array format so that it it's easier to compare
func parseGetAtt(n *yaml.Node) error {
	parts := strings.SplitN(n.Value, ".", 2)

	if len(parts) != 2 {
		return errors.New("GetAtt requires two parameters")
	}

	*n = yaml.Node{
		Kind: yaml.SequenceNode,

		HeadComment: n.HeadComment,
		LineComment: n.LineComment,
		FootComment: n.FootComment,

		Content: []*yaml.Node{
			{
				Kind:  yaml.ScalarNode,
				Style: 0,
				Tag:   "!!str",
				Value: parts[0],
			},
			{
				Kind:  yaml.ScalarNode,
				Style: 0,
				Tag:   "!!str",
				Value: parts[1],
			},
		},
	}
	return nil
}

// allTags holds every short tag the parser expands
var allTags = func() map[string]string {
	m := make(map[string]string, len(cft.Tags)+len(cft.ParseOnlyTags))
	maps.Copy(m, cft.Tags)
	maps.Copy(m, cft.ParseOnlyTags)
	return m
}()

// NormalizeNode takes a *yaml.Node and convert tag-style names into map-style,
// and converts other scalars into a canonical format
func NormalizeNode(n *yaml.Node) error {
	// Fix badly-parsed numbers
	if n.ShortTag() == "!!float" && n.Value[0] == '0' {
		n.Tag = "!!str"
	}

	// Fix badly-parsed timestamps which are often used for versions in cloudformation
	if n.ShortTag() == "!!timestamp" {
		n.Tag = "!!str"
	}

	// Convert tag-style intrinsics into map-style
	for tag, funcName := range allTags {
		if n.ShortTag() == tag {
			body := node.Clone(n)

			// Fix empty Fn values (should never be null)
			if body.Tag == "!!null" {
				body.Tag = "!!str"
			} else {
				body.Tag = ""
			}

			// Wrap in a map. The anchor labels the whole function, so an alias
			// still means the function when fmt writes the long form.
			*n = yaml.Node{
				Kind:   yaml.MappingNode,
				Style:  cft.ShortForm,
				Tag:    "!!map",
				Anchor: body.Anchor,
				Content: []*yaml.Node{
					{
						Kind:  yaml.ScalarNode,
						Style: 0,
						Tag:   "!!str",
						Value: funcName,
					},
					body,
				},
			}
			body.Anchor = ""

			break
		}
	}

	// Convert GetAtts
	if n.Kind == yaml.MappingNode && len(n.Content) == 2 {
		if n.Content[0].Value == "Fn::GetAtt" && n.Content[1].Kind == yaml.ScalarNode {
			err := parseGetAtt(n.Content[1])
			if err != nil {
				return err
			}
		}
	}

	for _, child := range n.Content {
		err := NormalizeNode(child)
		if err != nil {
			return err
		}
	}

	return nil
}
