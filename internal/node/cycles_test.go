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

func TestCloneAliasCycle(t *testing.T) {
	var n yaml.Node
	if err := yaml.Unmarshal([]byte(`Value: &c [*c]`), &n); err != nil {
		t.Fatal(err)
	}
	copy := node.Clone(&n)
	original := n.Content[0].Content[1]
	cloned := copy.Content[0].Content[1]
	if cloned == original || cloned.Content[0] == original.Content[0] {
		t.Fatal("clone shares nodes with the original")
	}
	if cloned.Content[0].Alias != cloned {
		t.Fatal("cloned alias does not refer to the cloned anchor")
	}
	if err := node.CheckAliasCycles(copy); err == nil || err.Error() != `anchor "c" refers to itself` {
		t.Fatalf("cloned cycle error = %v", err)
	}
}
