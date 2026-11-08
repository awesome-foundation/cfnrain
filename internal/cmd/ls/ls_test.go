package ls_test

import (
	"os"

	"github.com/awesome-foundation/cfnrain/internal/cmd/ls"
)

func Example_ls_help() {
	os.Args = []string{
		os.Args[0],
		"--help",
	}

	ls.Cmd.Execute()
	// Output:
	// Displays a list of all running stacks or the contents of <stack> if provided. If the -c arg is supplied, operates on changesets instead of stacks
	//
	// Usage:
	//   ls <stack> [changeset]
	//
	// Aliases:
	//   ls, list
	//
	// Flags:
	//   -a, --all[=true|false]         list stacks in all regions; if you specify a stack, show more details
	//   -c, --changeset[=true|false]   List changesets instead of stacks
	//   -h, --help[=true|false]        help for ls
}
