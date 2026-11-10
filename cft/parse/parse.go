// Package parse provides functions for parsing
// CloudFormation templates from JSON and YAML inputs.
package parse

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/awesome-foundation/cfnrain/cft"
	"github.com/awesome-foundation/cfnrain/cft/diff"

	"gopkg.in/yaml.v3"
)

// Reader returns a cft.Template parsed from an io.Reader
func Reader(r io.Reader) (*cft.Template, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return &cft.Template{}, fmt.Errorf("unable to read input: %s", err)
	}

	return String(string(data))
}

// File returns a cft.Template parsed from a file specified by fileName
func File(fileName string) (*cft.Template, error) {
	source, err := os.ReadFile(fileName)
	if err != nil {
		return &cft.Template{}, fmt.Errorf("unable to read file: %s", err)
	}

	t, err := String(string(source))
	if err != nil {
		return t, err
	}
	t.FileName = fileName
	tokens := strings.Split(fileName, "/")
	if len(tokens) > 1 {
		t.Name = tokens[len(tokens)-1]
	} else {
		t.Name = fileName
	}
	return t, nil
}

// Map returns a cft.Template parsed from a map[string]interface{}
func Map(input map[string]interface{}) (*cft.Template, error) {
	var node yaml.Node
	err := node.Encode(input)
	if err != nil {
		return &cft.Template{}, err
	}

	return Node(&node)
}

// String returns a cft.Template parsed from a string
func String(input string) (*cft.Template, error) {
	// yaml.v3 keeps \r line endings in comments, which breaks fmt.
	// YAML reads CRLF and a lone CR as one line break each, so this keeps values.
	// UTF-16 input is not valid UTF-8 and is left alone.
	if utf8.ValidString(input) {
		input = strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(input)
	}
	var n yaml.Node
	err := yaml.Unmarshal([]byte(input), &n)
	if err != nil {
		return &cft.Template{}, fmt.Errorf("invalid YAML: %s", err)
	}
	// yaml.v3 can attach a comment after an empty null to the next node.
	// Give these nulls an explicit value before the parser assigns comments.
	lines := strings.Split(input, "\n")
	changed := false
	var explicitNulls func(*yaml.Node)
	explicitNulls = func(n *yaml.Node) {
		if n.Kind == yaml.ScalarNode && n.Tag == "!!null" && n.Value == "" && n.Line > 0 && n.Line <= len(lines) {
			line := []rune(lines[n.Line-1])
			if n.Column > 0 && n.Column <= len(line) {
				tail := string(line[n.Column-1:])
				rest := strings.TrimSpace(strings.TrimPrefix(tail, "!!null"))
				if strings.HasPrefix(rest, "#") {
					// Use the node's suffix so a # in an earlier key stays intact.
					pos := len(string(line[:n.Column-1])) + strings.Index(tail, "#")
					lines[n.Line-1] = lines[n.Line-1][:pos] + "null " + lines[n.Line-1][pos:]
					changed = true
				}
			}
		}
		for _, child := range n.Content {
			explicitNulls(child)
		}
	}
	explicitNulls(&n)
	if changed {
		if err := yaml.Unmarshal([]byte(strings.Join(lines, "\n")), &n); err != nil {
			return &cft.Template{}, fmt.Errorf("invalid YAML: %s", err)
		}
	}

	return Node(&n)
}

// Node returns a cft.Template parse from a *yaml.Node
func Node(n *yaml.Node) (*cft.Template, error) {
	err := NormalizeNode(n)
	return &cft.Template{Node: n}, err
}

// Verify confirms that there is no semantic difference between
// the source cft.Template and the string representation in output.
// This can be used to ensure that the parse package hasn't done
// anything unexpected to your template.
func Verify(source *cft.Template, output string) error {
	// Check it matches the original
	validate, err := String(output)
	if err != nil {
		return err
	}

	d := diff.New(source, validate)
	if d.Mode() != diff.Unchanged {
		return fmt.Errorf("semantic difference after formatting:\n%s", d.Format(false))
	}

	return nil
}
