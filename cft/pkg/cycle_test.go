package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFileAliasCycles(t *testing.T) {
	for _, tc := range []struct {
		name, value string
		cyclic      bool
	}{
		{"join", `&c !Join ["", [*c]]`, true},
		{"mapping", `&c {Again: *c}`, true},
		{"list", `&c [*c]`, true},
		{"shared function", `[&c !Join ["", [a, b]], *c]`, false},
		{"shared mapping", `[&c {x: 1}, *c, *c]`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "cycle.yaml")
			if err := os.WriteFile(path, []byte("Metadata:\n  Value: "+tc.value+"\nResources: {}\n"), 0600); err != nil {
				t.Fatal(err)
			}
			output, err := File(path)
			if tc.cyclic {
				if err == nil || !strings.Contains(err.Error(), `YAML alias "c" after packaging: its anchor is recursive`) {
					t.Fatalf("error = %v, want anchor cycle error", err)
				}
			} else if err != nil || output == nil {
				t.Fatalf("non-cyclic alias: error = %v, output = %v", err, output)
			}
		})
	}
}

// Packaging can rebind aliases after templates and modules have been parsed.
func TestResolveAliasesRejectsCycles(t *testing.T) {
	for _, tc := range []struct {
		name, value string
		cyclic      bool
	}{
		{"join", `&c !Join ["", [*c]]`, true},
		{"mapping", `&c {Again: *c}`, true},
		{"list", `&c [*c]`, true},
		{"indirect", `&c {Child: &child [*c]}`, true},
		{"shared mapping", `[&c {x: 1}, *c, *c]`, false},
		{"shared function", `[&c !Join ["", [a, b]], *c]`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var n yaml.Node
			if err := yaml.Unmarshal([]byte("Value: "+tc.value), &n); err != nil {
				t.Fatal(err)
			}
			err := resolveAliases(&n)
			if tc.cyclic {
				if err == nil || !strings.Contains(err.Error(), `YAML alias "c"`) || !strings.Contains(err.Error(), "recursive") {
					t.Fatalf("error = %v, want named alias cycle error", err)
				}
			} else if err != nil {
				t.Fatalf("non-cyclic alias: %v", err)
			}
		})
	}
}
