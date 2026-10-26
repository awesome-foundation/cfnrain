package cft

import "testing"

// Deliberately failing test: proves that main's ruleset blocks a red PR. Never merge.
func TestRulesetProbe(t *testing.T) {
	t.Fatal("deliberate failure to test branch protection")
}
