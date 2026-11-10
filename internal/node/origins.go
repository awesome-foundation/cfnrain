package node

import "gopkg.in/yaml.v3"

// Origins records the source directory of YAML values while packaging moves them.
// It belongs to one template operation, and is not serialized into its output.
type Origins map[*yaml.Node]Source

// Source identifies the local directory and optional remote base of a value.
type Source struct {
	Directory string
	BaseURI   string
}

// Mark records a directory without replacing the origin of an existing value.
func (o Origins) Mark(n *yaml.Node, directory string) {
	o.MarkModule(n, directory, "")
}

// Copy preserves a value's source when a node is copied or replaced.
func (o Origins) Copy(dst, src *yaml.Node) {
	if o == nil {
		return
	}
	if directory, ok := o[src]; ok {
		o[dst] = directory
	}
}

// Directory returns the value's source directory, or the supplied default.
func (o Origins) Directory(n *yaml.Node, fallback string) string {
	if directory, ok := o[n]; ok {
		return directory.Directory
	}
	return fallback
}

// MarkModule also preserves the base URL used by relative remote module sources.
func (o Origins) MarkModule(n *yaml.Node, directory, baseURI string) {
	if n == nil || o == nil {
		return
	}
	if _, ok := o[n]; !ok {
		o[n] = Source{directory, baseURI}
	}
	for _, child := range n.Content {
		o.MarkModule(child, directory, baseURI)
	}
}

// BaseURI returns the remote base associated with a source value.
func (o Origins) BaseURI(n *yaml.Node, fallback string) string {
	if source, ok := o[n]; ok {
		return source.BaseURI
	}
	return fallback
}
