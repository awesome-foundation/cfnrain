package cmd_test

import (
	"testing"

	"github.com/awesome-foundation/cfnrain/internal/cmd"
)

func TestStdinOnce(t *testing.T) {
	for _, args := range [][]string{{}, {"a.yaml"}, {"-"}, {"a.yaml", "-"}, {"./-", "-"}} {
		if err := cmd.StdinOnce(nil, args); err != nil {
			t.Errorf("%v: unexpected error: %s", args, err)
		}
	}

	for _, args := range [][]string{{"-", "-"}, {"-", "a.yaml", "-"}} {
		if err := cmd.StdinOnce(nil, args); err == nil {
			t.Errorf("%v: expected an error", args)
		}
	}
}
