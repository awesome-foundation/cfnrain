package format

import (
	"strings"

	"github.com/awesome-foundation/cfnrain/cft"
	"github.com/awesome-foundation/cfnrain/internal/node"
	"gopkg.in/yaml.v3"
)

// Global, set by command arg --node-style
var NodeStyle string

var NodeStyleDocs = "Set the node output style to tagged, doublequoted, singlequoted, literal, folded, strict-boolean, quotescalars, original, or flow"

func joinComments(comments []string) string {
	var out []string
	for _, c := range comments {
		if c != "" {
			out = append(out, c)
		}
	}
	return strings.Join(out, "\n")
}

func mergeComments(comments []string) string {
	out := strings.Builder{}
	for _, c := range comments {
		c := strings.TrimSpace(strings.TrimLeft(c, "# "))
		if c != "" {
			out.WriteString(c)
			out.WriteString(" ")
		}
	}
	return strings.TrimSpace(out.String())
}

// shortTag returns the short tag for an intrinsic function name
func shortTag(funcName string) (tag string, parseOnly bool) {
	for tag, f := range cft.Tags {
		if f == funcName {
			return tag, false
		}
	}
	for tag, f := range cft.ParseOnlyTags {
		if f == funcName {
			return tag, true
		}
	}
	return "", false
}

// Fix up yaml.Nodes on the way out of a template.
// keepForm is true inside a short !GetStackOutput: AWS forbids short functions
// there, so each function keeps the form the source used.
func formatNode(n *yaml.Node, keepForm bool) *yaml.Node {
	n = node.Clone(n)

	// Is it a map?
	if n.Kind == yaml.MappingNode {
		// Does it have just one key/value pair?
		if len(n.Content) == 2 {

			// Is the key relevant?
			funcName := n.Content[0].Value
			tag, parseOnly := shortTag(funcName)
			short := n.Style&cft.ShortForm != 0 || !(parseOnly || keepForm)
			// Keep an argument anchor on the argument, not on the whole function.
			if tag != "" && short && n.Content[1].Anchor == "" {
				anchor := n.Anchor
				// Prepare comments
				headComments := []string{n.HeadComment, n.Content[0].HeadComment, n.Content[1].HeadComment}
				lineComments := []string{n.LineComment, n.Content[0].LineComment, n.Content[1].LineComment}
				footComments := []string{n.FootComment, n.Content[0].FootComment, n.Content[1].FootComment}

				n = n.Content[1]
				n.Tag = tag
				if anchor != "" {
					n.Anchor = anchor
				}

				if funcName == "Fn::GetStackOutput" {
					keepForm = true
				}

				// Is it a GetAtt and is currently a sequence?
				getAttSequence := funcName == "Fn::GetAtt" && n.Kind == yaml.SequenceNode
				if getAttSequence {
					// Are both parts scalars?
					allScalar := true
					parts := make([]string, len(n.Content))
					for i, child := range n.Content {
						if child.Kind != yaml.ScalarNode {
							allScalar = false
							break
						}

						parts[i] = child.Value

						headComments = append(headComments, child.HeadComment)
						lineComments = append(lineComments, child.LineComment)
						footComments = append(footComments, child.FootComment)
					}

					if allScalar {
						n.Content = []*yaml.Node{}
						n.Kind = yaml.ScalarNode
						n.Value = strings.Join(parts, ".")
					}
				}

				n.HeadComment = joinComments(headComments)
				n.LineComment = joinComments(lineComments)
				n.FootComment = joinComments(footComments)
				if getAttSequence {
					// Keep the existing treatment of comments on collapsed GetAtt parts.
					n.LineComment = mergeComments(lineComments)
				}
			}
		}
	}

	// Is it a scalar?
	if n.Kind == yaml.ScalarNode {
		// An empty null list item loses its line comment when parsed again.
		if n.Tag == "!!null" && n.Value == "" && n.LineComment != "" {
			n.Value = "null"
		}
		// Is it a string
		if n.Tag == "!!str" {
			// Reformat how yaml thinks is best
			if b, err := yaml.Marshal(n.Value); err == nil {
				var newNode yaml.Node
				if err = yaml.Unmarshal(b, &newNode); err == nil {
					n.Style = newNode.Content[0].Style
				}
			}

		}
	}

	for i, child := range n.Content {
		formatted := formatNode(child, keepForm)
		if n.Kind == yaml.MappingNode && i%2 == 1 && child.Kind == yaml.MappingNode && formatted.Kind != yaml.MappingNode {
			// Comments on a shortened function describe the containing property,
			// not the next key or the first item of a sequence argument.
			key := n.Content[i-1]
			moveLineComments := formatted.Kind == yaml.SequenceNode || strings.Contains(formatted.LineComment, "\n") ||
				(key.LineComment != "" && formatted.LineComment != "")
			if moveLineComments {
				// Sequences and multiple line comments cannot share the property line.
				// Keep their text above this key instead of on the next node.
				key.HeadComment = joinComments([]string{key.HeadComment, key.LineComment, formatted.HeadComment, formatted.LineComment})
				key.LineComment = ""
				formatted.LineComment = ""
			} else {
				key.HeadComment = joinComments([]string{key.HeadComment, formatted.HeadComment})
			}
			key.FootComment = joinComments([]string{formatted.FootComment, key.FootComment})
			formatted.HeadComment = ""
			formatted.FootComment = ""
		} else if n.Kind == yaml.SequenceNode && child.Kind == yaml.MappingNode && formatted.Kind == yaml.SequenceNode {
			// yaml.v3 reads a head comment on a tagged sequence as a comment
			// on its first argument. Use that position on the first format too.
			if len(formatted.Content) > 0 {
				first := formatted.Content[0]
				first.HeadComment = joinComments([]string{formatted.HeadComment, formatted.LineComment, first.HeadComment})
				formatted.HeadComment = ""
				formatted.LineComment = ""
			}
		}
		if formatted.Style&yaml.FlowStyle == 0 && formatted.LineComment != "" && len(formatted.Content) > 0 &&
			(formatted.Kind == yaml.SequenceNode || formatted.Kind == yaml.MappingNode) {
			// Block collections cannot carry a line comment after their closing delimiter.
			// Keep it above the property or inside the nested collection instead.
			if n.Kind == yaml.MappingNode && i%2 == 1 {
				key := n.Content[i-1]
				key.HeadComment = joinComments([]string{key.HeadComment, formatted.LineComment})
			} else {
				first := formatted.Content[0]
				first.HeadComment = joinComments([]string{formatted.LineComment, first.HeadComment})
			}
			formatted.LineComment = ""
		}
		if n.Kind == yaml.SequenceNode && formatted.Kind == yaml.SequenceNode && formatted.FootComment != "" && len(formatted.Content) > 0 {
			// yaml.v3 delays a nested sequence's foot comment until the next item.
			// Put it after the final argument so it stays inside this sequence.
			last := formatted.Content[len(formatted.Content)-1]
			for last.Kind == yaml.SequenceNode && len(last.Content) > 0 {
				last = last.Content[len(last.Content)-1]
			}
			last.FootComment = joinComments([]string{last.FootComment, formatted.FootComment})
			formatted.FootComment = ""
		}
		// yaml.v3 can attach comments on a bare empty short-form argument to the
		// next node. Quote the argument so head and line comments stay with it.
		if formatted.Kind == yaml.ScalarNode && formatted.Value == "" && strings.HasPrefix(formatted.Tag, "!") && !strings.HasPrefix(formatted.Tag, "!!") &&
			(formatted.HeadComment != "" || formatted.LineComment != "") {
			formatted.Style = yaml.DoubleQuotedStyle
			if NodeStyle == "singlequoted" {
				formatted.Style = yaml.SingleQuotedStyle
			}
		}
		n.Content[i] = formatted
	}

	// Allow global user overrides
	switch NodeStyle {
	case "tagged":
		n.Style = yaml.TaggedStyle
	case "doublequoted":
		n.Style = yaml.DoubleQuotedStyle
	case "singlequoted":
		n.Style = yaml.SingleQuotedStyle
	case "literal":
		n.Style = yaml.LiteralStyle
	case "folded":
		n.Style = yaml.FoldedStyle
	case "flow":
		n.Style = yaml.FlowStyle
	case "original":
		// Do nothing, leave it alone
	case "quotescalars":
		if n.Kind == yaml.ScalarNode {
			n.Style = yaml.DoubleQuotedStyle
		}
	case "strict-booleans":
		// Enforce YAML 1.2 behavior: Only `true` and `false` are booleans, all others are strings
		if n.Kind == yaml.ScalarNode && n.Tag == "!!str" && isAmbiguousScalar(n.Value) {
			n.Style = yaml.DoubleQuotedStyle
		} else {
			n.Style = 0
		}
	case "":
		// Default style for consistent formatting
		n.Style = 0
	default:
		panic("invalid --node-style: " + NodeStyle)
	}

	// yaml.v3 emits an explicit !!merge tag even for a plain merge key.
	// Let the plain << spelling carry the tag implicitly instead.
	if n.Kind == yaml.ScalarNode && n.Tag == "!!merge" && n.Value == "<<" && (n.Style == 0 || n.Style == yaml.FlowStyle) {
		n.Tag = ""
	}
	// A literal << string must stay quoted so it cannot become a merge key.
	if n.Kind == yaml.ScalarNode && n.Tag == "!!str" && n.Value == "<<" && (n.Style == 0 || n.Style == yaml.FlowStyle) {
		n.Style = yaml.DoubleQuotedStyle
	}

	return n
}

func isAmbiguousScalar(val string) bool {
	switch strings.ToLower(val) {
	case "y", "yes", "n", "no", "true", "false", "on", "off":
		return true
	default:
		return false
	}
}
