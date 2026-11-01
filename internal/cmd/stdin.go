package cmd

import (
	"errors"
	"slices"

	"github.com/spf13/cobra"
)

// StdinOnce is a cobra.PositionalArgs that rejects more than one "-" argument,
// because a command can read stdin only once
func StdinOnce(_ *cobra.Command, args []string) error {
	if i := slices.Index(args, "-"); i >= 0 && slices.Contains(args[i+1:], "-") {
		return errors.New(`"-" (stdin) can be given only once`)
	}

	return nil
}
