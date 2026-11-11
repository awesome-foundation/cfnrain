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

// FileOrReader returns a cft.Template parsed from r if fileName is "-",
// and from the file named fileName otherwise
func FileOrReader(fileName string, r io.Reader) (*cft.Template, error) {
	if fileName == "-" {
		return Reader(r)
	}

	return File(fileName)
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
	// yaml.v3 can attach a comment after an empty scalar to the next node.
	// Parse explicit values to attach comments, then restore the empty values.
	lines := strings.Split(input, "\n")
	rewritten := make(map[int]bool)
	index := 0
	var explicitNulls func(*yaml.Node)
	explicitNulls = func(n *yaml.Node) {
		current := index
		index++
		// Quoted and block values can follow their tag and comment on the next line.
		if n.Kind == yaml.ScalarNode && (n.Tag == "!!null" || n.Tag == "!!str" || strings.HasPrefix(n.Tag, "!") && !strings.HasPrefix(n.Tag, "!!")) && n.Value == "" &&
			n.Style&(yaml.DoubleQuotedStyle|yaml.SingleQuotedStyle|yaml.LiteralStyle|yaml.FoldedStyle) == 0 && n.Line > 0 && n.Line <= len(lines) {
			line := []rune(lines[n.Line-1])
			if n.Column > 0 && n.Column <= len(line) {
				tail := string(line[n.Column-1:])
				rest := strings.TrimLeft(tail, " \t")
				// A node can start with a tag, an anchor, or both.
				for strings.HasPrefix(rest, "!") || strings.HasPrefix(rest, "&") {
					delimiters := " \t"
					if strings.HasPrefix(rest, "&") {
						delimiters += ",[]{}"
					}
					end := strings.IndexAny(rest, delimiters)
					if end < 0 {
						break
					}
					rest = strings.TrimLeft(rest[end:], " \t")
				}
				comment := strings.TrimSpace(strings.TrimPrefix(rest, ","))
				if strings.HasPrefix(comment, "#") {
					// Insert before a flow separator, but after any tag and anchor.
					pos := len(string(line[:n.Column-1])) + len(tail) - len(rest)
					value := "null "
					if n.Tag != "!!null" {
						value = "\"\" "
					}
					if pos > 0 && !strings.ContainsAny(lines[n.Line-1][pos-1:pos], " \t") {
						value = " " + value
					}
					lines[n.Line-1] = lines[n.Line-1][:pos] + value + lines[n.Line-1][pos:]
					rewritten[current] = true
				}
			}
		}
		for _, child := range n.Content {
			explicitNulls(child)
		}
	}
	explicitNulls(&n)
	if len(rewritten) > 0 {
		var repaired yaml.Node
		if err := yaml.Unmarshal([]byte(strings.Join(lines, "\n")), &repaired); err != nil {
			return Node(&n)
		}
		// Restore empty values, then check that the repair changed only comments.
		index = 0
		var restoreNulls func(*yaml.Node)
		restoreNulls = func(n *yaml.Node) {
			if rewritten[index] {
				n.Value = ""
			}
			index++
			for _, child := range n.Content {
				restoreNulls(child)
			}
		}
		restoreNulls(&repaired)
		if sameParsedValues(&n, &repaired) {
			n = repaired
		}
	}

	return Node(&n)
}

// sameParsedValues checks structure and scalar text before accepting a comment repair.
// Alias names identify their targets without following potentially recursive pointers.
func sameParsedValues(a, b *yaml.Node) bool {
	if a.Kind != b.Kind || a.Tag != b.Tag || a.Value != b.Value ||
		a.Anchor != b.Anchor || len(a.Content) != len(b.Content) {
		return false
	}
	for i := range a.Content {
		if !sameParsedValues(a.Content[i], b.Content[i]) {
			return false
		}
	}
	return true
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
