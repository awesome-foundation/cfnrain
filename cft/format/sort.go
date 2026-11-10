package format

import (
	"gopkg.in/yaml.v3"
)

type ordering struct {
	props    []string
	children map[string]ordering
}

var orders = ordering{
	props: []string{
		"AWSTemplateFormatVersion",
		"Description",
		"Metadata",
		"Parameters",
		"Rules",
		"Mappings",
		"Conditions",
		"Transform",
		"Globals",
		"Resources",
		"Outputs",
	},
	children: map[string]ordering{
		"Transform": {
			props: []string{},
			children: map[string]ordering{
				"*": {
					props: []string{"Name", "Parameters"},
				},
			},
		},
		"Metadata": {
			props: []string{},
			children: map[string]ordering{
				"*": {
					props: []string{"Description"},
				},
			},
		},
		"Parameters": {
			props: []string{},
			children: map[string]ordering{
				"*": {
					props: []string{"Description", "Type", "AllowedValues", "Default"},
				},
			},
		},
		"Resources": {
			props: []string{},
			children: map[string]ordering{
				"*": {
					props: []string{"CreationPolicy", "DeletionPolicy", "UpdatePolicy", "UpdateReplacePolicy", "Type", "DependsOn", "Metadata", "Properties"},
				},
			},
		},
		"Outputs": {
			props: []string{},
			children: map[string]ordering{
				"*": {
					props: []string{"Description", "Value", "Export"},
				},
			},
		},
	},
}

func orderTemplate(node *yaml.Node) *yaml.Node {
	if node.Kind == yaml.DocumentNode {
		node = node.Content[0]
	}

	out := orderNode(node, orders)
	if out.Style&yaml.FlowStyle == 0 && out.LineComment != "" && len(out.Content) > 0 {
		// The root has no parent key to carry its trailing comment.
		first := out.Content[0]
		first.HeadComment = joinComments([]string{out.LineComment, first.HeadComment})
		out.LineComment = ""
	}
	return out
}

type nodeMap struct {
	keys   []*yaml.Node
	keyMap map[string]*yaml.Node
	values map[string]*yaml.Node
}

func toNodeMap(node *yaml.Node) nodeMap {
	out := nodeMap{
		keys:   make([]*yaml.Node, 0),
		keyMap: make(map[string]*yaml.Node),
		values: make(map[string]*yaml.Node),
	}

	if node.Kind != yaml.MappingNode {
		return out
	}

	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		value := node.Content[i+1]

		out.keys = append(out.keys, key)
		out.keyMap[key.Value] = key
		out.values[key.Value] = value
	}

	return out
}

func orderNode(node *yaml.Node, order ordering) *yaml.Node {
	if node == nil {
		return nil
	}

	if node.Kind != yaml.MappingNode {
		return node
	}

	// Discover the keys and values
	nm := toNodeMap(node)

	// Apply any sub-orderings
	for key, value := range nm.values {
		if subOrder, ok := order.children[key]; ok {
			nm.values[key] = orderNode(value, subOrder)
		} else if subOrder, ok := order.children["*"]; ok {
			nm.values[key] = orderNode(value, subOrder)
		}
		child := nm.values[key]
		if child.Kind == yaml.MappingNode && child.Style&yaml.FlowStyle == 0 && child.LineComment != "" && len(child.Content) > 0 {
			// Keep the trailing comment above the property when its mapping becomes block style.
			property := nm.keyMap[key]
			property.HeadComment = joinComments([]string{property.HeadComment, child.LineComment})
			child.LineComment = ""
		}
	}

	// Build the output node
	// Keep the comments when reordering content.
	// Keep the tag and the anchor, or a short tag or an alias to it breaks.
	out := *node
	out.Content = nil
	if NodeStyle != "original" && NodeStyle != "flow" && NodeStyle != "quotescalars" {
		out.Style = 0
	}

	if (NodeStyle == "flow" || NodeStyle == "quotescalars") && out.Style&yaml.FlowStyle != 0 {
		for _, key := range nm.keys {
			key.HeadComment = joinComments([]string{key.HeadComment, key.LineComment})
			key.LineComment = ""
		}
	}

	done := make(map[string]bool)

	// Place priority keys first
	for _, prop := range order.props {
		if child, ok := nm.values[prop]; ok {
			out.Content = append(out.Content, nm.keyMap[prop])
			out.Content = append(out.Content, child)
			done[prop] = true
		}
	}

	// Place remaining keys
	for _, key := range nm.keys {
		if !done[key.Value] {
			out.Content = append(out.Content, key)
			out.Content = append(out.Content, nm.values[key.Value])
		}
	}

	return &out
}
