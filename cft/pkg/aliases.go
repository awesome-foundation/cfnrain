package pkg

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/awesome-foundation/cfnrain/cft/visitor"
	"gopkg.in/yaml.v3"
)

var moduleAnchorScope atomic.Uint64

// scopeModuleAnchors keeps each parsed module instance's local names distinct.
// Parameter values from the parent retain their parent's anchor names.
func scopeModuleAnchors(root *yaml.Node) {
	prefix := fmt.Sprintf("rain_module_%d_", moduleAnchorScope.Add(1))
	visitor.NewVisitor(root).Visit(func(v *visitor.Visitor) {
		n := v.GetYamlNode()
		if n.Anchor != "" {
			n.Anchor = prefix + n.Anchor
		}
		if n.Kind == yaml.AliasNode {
			n.Value = prefix + n.Value
		}
	})
}

// bindRepeatedAnchors preserves source bindings before module expansion moves nodes.
// Recursive names retain the alias resolver's existing fallback behavior.
func bindRepeatedAnchors(root *yaml.Node) {
	counts := make(map[string]int)
	recursive := make(map[string]bool)
	active := make(map[*yaml.Node]bool)
	latest := make(map[string]*yaml.Node)
	targets := make(map[*yaml.Node]*yaml.Node)
	var collect func(*yaml.Node)
	collect = func(n *yaml.Node) {
		if n.Anchor != "" {
			counts[n.Anchor]++
			active[n] = true
			// A completed nested definition still shadows an outer definition.
			latest[n.Anchor] = n
		}
		if n.Kind == yaml.AliasNode {
			targets[n] = latest[n.Value]
			if active[targets[n]] {
				recursive[n.Value] = true
			}
		}
		for _, child := range n.Content {
			collect(child)
		}
		if n.Anchor != "" {
			delete(active, n)
		}
	}
	collect(root)

	preceding := make(map[string]*yaml.Node)
	bound := make(map[*yaml.Node]*yaml.Node)
	visitor.NewVisitor(root).Visit(func(v *visitor.Visitor) {
		n := v.GetYamlNode()
		if name := n.Anchor; counts[name] > 1 && !recursive[name] {
			unique := fmt.Sprintf("rain_binding_%d_%s", moduleAnchorScope.Add(1), name)
			for counts[unique] != 0 {
				unique = fmt.Sprintf("rain_binding_%d_%s", moduleAnchorScope.Add(1), name)
			}
			n.Anchor = unique
			preceding[name] = n
		}
		if n.Kind == yaml.AliasNode {
			if target := preceding[n.Value]; target != nil {
				n.Value = target.Anchor
				bound[n] = target
			}
		}
	})

	// Copy bound values before expansion can discard their definitions.
	// Resolve nested aliases from their source bindings, including unique names.
	// Leave recursive names for resolveAliases and do not follow Alias pointers.
	var copyBound func(*yaml.Node) *yaml.Node
	copyBound = func(n *yaml.Node) *yaml.Node {
		if target := targets[n]; target != nil && !recursive[n.Value] {
			n = target
		}
		result := *n
		result.Content = make([]*yaml.Node, len(n.Content))
		for i, child := range n.Content {
			result.Content[i] = copyBound(child)
		}
		return &result
	}
	for alias, target := range bound {
		*alias = *copyBound(target)
	}
}

type aliasCycleError struct{ name string }

func (e *aliasCycleError) Error() string {
	return fmt.Sprintf("cannot resolve YAML alias %q after packaging: its anchor is recursive", e.name)
}

// resolveAliases first binds aliases in document order, then expands their
// values. A definition is available only after its children have been visited.
// The global definitions handle anchors moved after aliases by module expansion.
func resolveAliases(root *yaml.Node) error {
	definitions := make(map[string][]*yaml.Node)
	preceding := make(map[string]*yaml.Node)
	preferred := make(map[*yaml.Node]*yaml.Node)
	var collect func(*yaml.Node)
	collect = func(n *yaml.Node) {
		if n.Anchor != "" {
			definitions[n.Anchor] = append(definitions[n.Anchor], n)
		}
		if n.Kind == yaml.AliasNode {
			preferred[n] = preceding[n.Value]
		}
		for _, child := range n.Content {
			collect(child)
		}
		if n.Anchor != "" {
			preceding[n.Anchor] = n
		}
	}
	collect(root)

	active := make(map[*yaml.Node]bool)
	resolved := make(map[*yaml.Node]*yaml.Node)
	var expand func(*yaml.Node) (*yaml.Node, error)
	expand = func(n *yaml.Node) (*yaml.Node, error) {
		if result, ok := resolved[n]; ok {
			return result, nil
		}
		if active[n] {
			return nil, &aliasCycleError{n.Anchor}
		}
		active[n] = true
		defer delete(active, n)

		if n.Kind == yaml.AliasNode {
			candidates := []*yaml.Node{}
			if p := preferred[n]; p != nil {
				candidates = append(candidates, p)
			}
			defs := definitions[n.Value]
			for i := len(defs) - 1; i >= 0; i-- {
				if defs[i] != preferred[n] {
					candidates = append(candidates, defs[i])
				}
			}
			if len(candidates) == 0 {
				return nil, fmt.Errorf("cannot resolve YAML alias %q after packaging: its anchor is missing", n.Value)
			}
			for _, candidate := range candidates {
				result, err := expand(candidate)
				if err != nil {
					var cycle *aliasCycleError
					if errors.As(err, &cycle) {
						continue
					}
					return nil, err
				}
				resolved[n] = result
				return result, nil
			}
			return nil, &aliasCycleError{n.Value}
		}

		// Do not follow yaml.Node.Alias pointers: they can contain cycles or stale
		// copies of expressions that packaging has already replaced.
		result := *n
		result.Anchor = ""
		result.Alias = nil
		result.Content = make([]*yaml.Node, len(n.Content))
		for i, child := range n.Content {
			value, err := expand(child)
			if err != nil {
				return nil, err
			}
			result.Content[i] = value
		}
		resolved[n] = &result
		return &result, nil
	}
	result, err := expand(root)
	if err != nil {
		return err
	}
	*root = *result
	return nil
}
