package node

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// CheckAliasCycles rejects recursive aliases before a caller clones or expands nodes.
func CheckAliasCycles(n *yaml.Node) error {
	return rejectAliasCycles(n, make(map[*yaml.Node]bool))
}

// The map marks nodes on the current path as true and completed nodes as false.
// Completed nodes can appear again through aliases without forming a cycle.
func rejectAliasCycles(n *yaml.Node, visiting map[*yaml.Node]bool) error {
	if n == nil {
		return nil
	}
	if active, seen := visiting[n]; seen {
		if active {
			return fmt.Errorf("anchor %q refers to itself", n.Anchor)
		}
		return nil
	}
	visiting[n] = true
	if err := rejectAliasCycles(n.Alias, visiting); err != nil {
		return err
	}
	for _, child := range n.Content {
		if err := rejectAliasCycles(child, visiting); err != nil {
			return err
		}
	}
	visiting[n] = false
	return nil
}
