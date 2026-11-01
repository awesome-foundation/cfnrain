package format

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/awesome-foundation/cfnrain/cft/parse"
	"gopkg.in/yaml.v3"
)

func handleScalar(node *yaml.Node) interface{} {
	if node.Tag != "" && !strings.HasPrefix(node.Tag, "!!") {
		// Convert back to map style
		tag := strings.TrimPrefix(node.Tag, "!")

		if tag != "Ref" {
			tag = "Fn::" + tag
		}

		return MapSlice{
			MapItem{
				Key:   tag,
				Value: node.Value,
			},
		}
	}

	// Marshal and unmarshal the node and return the result
	intermediate, err := yaml.Marshal(node)
	if err != nil {
		panic(err)
	}

	var out interface{}
	err = yaml.Unmarshal(intermediate, &out)
	if err != nil {
		panic(err)
	}

	return out
}

func Jsonise(node *yaml.Node) interface{} {
	return (&jsoniser{expanding: map[*yaml.Node]bool{}}).jsonise(node)
}

// maxAliasExpansions caps how many aliases one Jsonise call expands, so a
// small template whose aliases nest each other cannot expand without limit
const maxAliasExpansions = 10000

// jsoniser holds the alias targets that are being expanded, so a cyclic alias
// stops, and counts the expansions against maxAliasExpansions
type jsoniser struct {
	expanding  map[*yaml.Node]bool
	expansions int
}

func (j *jsoniser) jsonise(node *yaml.Node) interface{} {
	switch node.Kind {
	case yaml.DocumentNode:
		return j.jsonise(node.Content[0])
	case yaml.MappingNode:
		out := make(MapSlice, len(node.Content)/2)
		for i := 0; i < len(node.Content); i += 2 {
			key, value := node.Content[i], node.Content[i+1]
			out[i/2] = MapItem{
				Key:   j.jsonise(key),
				Value: j.jsonise(value),
			}
		}
		return out
	case yaml.SequenceNode:
		out := make([]interface{}, len(node.Content))
		for i, n := range node.Content {
			out[i] = j.jsonise(n)
		}
		return out
	case yaml.AliasNode:
		// JSON has no aliases, so write the anchored value again. A cyclic
		// alias, or one past the cap, goes to handleScalar as before aliases
		// were expanded, which fails with an unknown anchor error.
		if j.expanding[node.Alias] || j.expansions >= maxAliasExpansions {
			return handleScalar(node)
		}
		j.expansions++
		j.expanding[node.Alias] = true
		defer delete(j.expanding, node.Alias)
		return j.jsonise(node.Alias)
	default:
		return handleScalar(node)
	}
}

func convertToJSON(in string) string {

	var d yaml.Node

	err := yaml.Unmarshal([]byte(in), &d)
	if err != nil {
		panic(err)
	}

	err = parse.NormalizeNode(&d)
	if err != nil {
		panic(err)
	}

	intermediate := Jsonise(&d)

	out, err := ToJson(intermediate, "    ")
	if err != nil {
		panic(err)
	}
	s := string(out)
	return s
}

// converts struct to a JSON formatted string
func PrettyPrint(i interface{}) string {
	s, _ := ToJson(i, "\t")
	return string(s)
}

// ToJson overrides the default behavior of json.Marshal to leave < > alone
func ToJson(i interface{}, indent string) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", indent)
	err := enc.Encode(i)
	retval := bytes.TrimRight(buf.Bytes(), "\n")
	return retval, err
}
