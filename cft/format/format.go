// Package format contains functionality to render a cft.Template
// into YAML or JSON
package format

import (
	"regexp"
	"strings"

	"github.com/awesome-foundation/cfnrain/cft"
	"gopkg.in/yaml.v3"
)

// Options contains options for formatting cfn templates
type Options struct {
	// JSON determines whether the outputs will be JSON (true) or YAML (false)
	JSON bool

	// Unsorted will cause the formatter to leave the ordering of template elements
	// as in the original template if true.
	// If false, the formatter will rearrange the template elements into
	// canonical order.
	Unsorted bool
}

func CheckMultilineBegin(s string) bool {
	// Match block header
	// https://yaml.org/spec/1.2.2/#8112-block-chomping-indicator
	r, _ := regexp.Compile("[|>](([0-9]*[+-])|([+-][0-9]*))?$")
	return r.MatchString(s)
}

// String returns a string representation of the supplied cft.Template
func String(t *cft.Template, opt Options) string {
	if NodeStyle == "strict-boolean" {
		NodeStyle = "strict-booleans"
		defer func() { NodeStyle = "strict-boolean" }()
	}

	node := t.Node

	node = formatNode(node, false, false)

	if !opt.Unsorted {
		node = orderTemplate(node)
	} else {
		moveUnsortedKeyComments(node)
	}

	quoteFlowNulls(node, false)

	out, err := encodeYAML(node)
	if err != nil {
		panic(err)
	}
	if NodeStyle == "folded" {
		// Each scalar advances only from folded to literal to quoted style.
		// The loop ends when no scalar needs another style. The fmt command
		// reports parse errors and structural mismatches.
		for {
			var decoded yaml.Node
			if yaml.Unmarshal([]byte(out), &decoded) != nil || len(decoded.Content) == 0 {
				break
			}
			original := node
			if original.Kind == yaml.DocumentNode {
				original = original.Content[0]
			}
			if !restoreBlockStrings(original, decoded.Content[0]) {
				break
			}
			out, err = encodeYAML(node)
			if err != nil {
				panic(err)
			}
		}
	}
	if opt.JSON {
		out = convertToJSON(out) + "\n"
	}
	return out
}

// restoreBlockStrings selects a safer style only when the actual output changes a value.
func restoreBlockStrings(original, decoded *yaml.Node) bool {
	if original.Kind != decoded.Kind || len(original.Content) != len(decoded.Content) {
		return false
	}
	changed := false
	if original.Kind == yaml.ScalarNode && original.Value != decoded.Value {
		switch original.Style {
		case yaml.FoldedStyle:
			original.Style = yaml.LiteralStyle
			changed = true
		case yaml.LiteralStyle:
			original.Style = yaml.DoubleQuotedStyle
			changed = true
		}
	}
	for i, child := range original.Content {
		if restoreBlockStrings(child, decoded.Content[i]) {
			changed = true
		}
	}
	return changed
}

// encodeYAML applies the same encoder settings and layout to every style attempt.
func encodeYAML(node *yaml.Node) (string, error) {
	buf := strings.Builder{}
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		return "", err
	}

	// Remove only the encoder's final line terminator. Blank lines and tabs
	// at the end can belong to the last scalar.
	parts := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	result := strings.Builder{}

	lastIndent := 0
	indent := 0
	lastPartWasComment := false
	lastLineWasEmpty := false
	startMultilineIndent := -1

	for i, part := range parts {

		trimmedPart := strings.TrimLeft(part, " ")
		indent = len(part) - len(trimmedPart)

		// Leave lines alone if they are in a multiline block
		// Note: CloudFormation does not comply with the YAML spec. It treats > just like |
		// https://yaml-multiline.info/
		// https://stackoverflow.com/questions/3790454/how-do-i-break-a-string-in-yaml-over-multiple-lines
		// https://yaml.org/spec/1.2-old/spec.html#id2760844
		isMultiline := false
		if startMultilineIndent > -1 {
			// Note: len(part) == 0 means empty line without indentation
			// https://github.com/aws-cloudformation/rain/issues/126
			if indent <= startMultilineIndent && len(part) != 0 {
				startMultilineIndent = -1
			} else {
				isMultiline = true
			}
		}
		trimmedRight := strings.TrimRight(part, " ")
		if !isMultiline && CheckMultilineBegin(trimmedRight) {
			startMultilineIndent = indent
		}

		isComment := false
		if len(part) > 0 && strings.HasPrefix(trimmedPart, "#") {
			isComment = true
		}

		isEmpty := len(part) == 0 // This should never be true

		// Add lines between 1st and 2nd level properties, except for comments
		if indent <= lastIndent && (indent == 0 || indent == 2) {
			if !lastPartWasComment && !isMultiline {
				// If the last line was a comment, don't newline here,
				// since we want the comment to stick to the thing it was above
				result.WriteString("\n")
				lastLineWasEmpty = true
			}
		}

		// Put a line break above first/only comment lines
		if !lastPartWasComment && isComment && !lastLineWasEmpty && !isMultiline {
			// The encoder separates foot comments from the next key with a
			// blank line. Adding one before them would turn them into head
			// comments on that next key when the YAML is parsed again.
			end := i + 1
			for end < len(parts) && strings.HasPrefix(strings.TrimSpace(parts[end]), "#") {
				end++
			}
			if end < len(parts) && strings.TrimSpace(parts[end]) != "" {
				result.WriteString("\n")
			}
		}

		result.WriteString(part)
		result.WriteString("\n")

		lastIndent = indent
		lastPartWasComment = isComment
		lastLineWasEmpty = isEmpty
	}

	// The layout pass can add a leading blank line, but the ending is data.
	out := strings.TrimLeft(result.String(), "\n")

	return out, nil
}

// CftToYaml converts a template to a YAML string
func CftToYaml(t *cft.Template) string {
	return String(t, Options{JSON: false, Unsorted: false})
}

// moveUnsortedKeyComments keeps key comments away from explicit collection tags
// and flow delimiters, where yaml.v3 can emit them between a key and its colon.
func moveUnsortedKeyComments(n *yaml.Node) {
	if n.Kind == yaml.MappingNode {
		for i := 0; i < len(n.Content); i += 2 {
			key, value := n.Content[i], n.Content[i+1]
			collection := value.Kind == yaml.MappingNode || value.Kind == yaml.SequenceNode
			if key.LineComment != "" && collection && (n.Style&yaml.FlowStyle != 0 || value.Style&yaml.TaggedStyle != 0) {
				key.HeadComment = joinComments([]string{key.HeadComment, key.LineComment})
				key.LineComment = ""
			}
		}
	}
	for _, child := range n.Content {
		moveUnsortedKeyComments(child)
	}
}
