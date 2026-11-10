package node_test

import (
	"testing"

	"github.com/awesome-foundation/cfnrain/internal/node"
	"gopkg.in/yaml.v3"
)

func TestCheckAliasCyclesDoesNotMutate(t *testing.T) {
	var n yaml.Node
	if err := yaml.Unmarshal([]byte(`Value: &c !Join ["", [*c]]`), &n); err != nil {
		t.Fatal(err)
	}
	value := n.Content[0].Content[1]
	if err := node.CheckAliasCycles(&n); err == nil || err.Error() != `anchor "c" refers to itself` {
		t.Fatalf("error = %v, want anchor cycle error", err)
	}
	if value.Kind != yaml.SequenceNode || value.Tag != "!Join" || value.Anchor != "c" {
		t.Fatalf("cycle detection changed the input: %+v", value)
	}
}
