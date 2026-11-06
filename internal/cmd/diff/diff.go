package diff

import (
	"fmt"
	"os"

	"github.com/awesome-foundation/cfnrain/internal/cmd"
	"github.com/awesome-foundation/cfnrain/internal/ui"

	"github.com/awesome-foundation/cfnrain/cft/diff"
	"github.com/awesome-foundation/cfnrain/cft/parse"
	"github.com/spf13/cobra"
)

var longDiff = false

// Cmd is the diff command's entrypoint
var Cmd = &cobra.Command{
	Use:                   "diff <from> <to>",
	Short:                 "Compare CloudFormation templates",
	Long:                  "Outputs a summary of the changes necessary to transform the CloudFormation template named <from> into the template named <to>. Give \"-\" as <from> or <to> to read that template from stdin.",
	Args:                  cobra.MatchAll(cobra.ExactArgs(2), cmd.StdinOnce),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		leftFn, rightFn := args[0], args[1]

		left, err := parse.FileOrReader(leftFn, os.Stdin)
		if err != nil {
			panic(ui.Errorf(err, "unable to parse template '%s'", leftFn))
		}

		right, err := parse.FileOrReader(rightFn, os.Stdin)
		if err != nil {
			panic(ui.Errorf(err, "unable to parse template '%s'", leftFn))
		}

		fmt.Print(ui.ColouriseDiff(diff.New(left, right), longDiff))
	},
}

func init() {
	Cmd.Flags().BoolVarP(&longDiff, "long", "l", false, "Include unchanged elements in diff output")
}
