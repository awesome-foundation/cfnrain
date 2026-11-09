package pkg

import (
	"fmt"
	"os"

	"github.com/awesome-foundation/cfnrain/cft/format"
	cftpkg "github.com/awesome-foundation/cfnrain/cft/pkg"
	"github.com/awesome-foundation/cfnrain/internal/config"
	"github.com/awesome-foundation/cfnrain/internal/console/spinner"
	"github.com/awesome-foundation/cfnrain/internal/node"
	"github.com/awesome-foundation/cfnrain/internal/ui"
	"github.com/spf13/cobra"
)

var outFn = ""
var dataModel bool

// Experimental is an optional argument that enables experimental features
var Experimental bool

// Cmd is the pkg command's entrypoint
var Cmd = &cobra.Command{
	Use:   "pkg <template>",
	Short: "Package local artifacts into a template",
	Long: `Performs the same functions as "aws cloudformation package" but with added functionality.

You may use the following, rain-specific directives in templates packaged with "cfnrain pkg":

  !Rain::Embed <path>          Embeds the contents of the file at <path> into the template as a string

  !Rain::Include <path>        Reads the file at <path> as YAML/JSON and inserts the resulting object into the template

  !Rain::Env <name>            Reads the <name> environmental variable and inserts value into the template as a string

  !Rain::S3Http <path>         Uploads <path> (zipping first if it is a directory) to S3
                               and embeds the S3 HTTP URL into the template as a string

  !Rain::S3 <path>             Uploads <path> (zipping first if it is a directory) to S3
                               and embeds the S3 URI into the template as a string

  !Rain::S3 <object>           supply an object with the following properties: 
    Path: <path>               a file or directory to be uploaded to S3
    Zip: true|false            If "true", cfnrain will zip <path> even if it is a file
    BucketProperty: <bucket>   If you supply "BucketProperty" and "KeyProperty", cfnrain pkg will
    KeyProperty: <key>         include the uploaded file/directory's details as an object in the template
                               with the property names you specify.
    Format: URI|Http|Object    Specify which format cfnrain pkg should return the S3 location as.
                               The default is "Object" when both BucketProperty and KeyProperty are set,
                               and "URI" otherwise. "Object" requires both property names.
    Run: <script>              Run an executable script in the template's directory before the upload.
                               The script path is relative to the template file or absolute.
    Extension: <extension>     Append a file extension to the S3 object key.

  !Rain::Module <url>          Supply a URL to a rain module. The module's Parameters define its properties,
                               and its resources are inserted into the template.
                               This is an experimental directive that must be enabled by adding the 
                               --experimental arg on the command line.
`,
	Args:                  cobra.ExactArgs(1),
	Aliases:               []string{"package"},
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		fn := args[0]

		cftpkg.Experimental = Experimental

		spinner.Push(fmt.Sprintf("Packaging template '%s'", fn))
		packaged, err := cftpkg.File(fn)
		if err != nil {
			panic(ui.Errorf(err, "unable to package template '%s'", fn))
		}
		spinner.Pop()

		var out string
		if dataModel {
			out = node.ToJson(packaged.Node)
		} else {
			out = format.String(packaged, format.Options{})
		}

		if outFn != "" {
			os.WriteFile(outFn, []byte(out), 0644)
		} else {
			fmt.Println(out)
		}
	},
}

func init() {
	Cmd.Flags().StringVarP(&outFn, "output", "o", "", "Output packaged template to a file")
	Cmd.Flags().BoolVarP(&Experimental, "experimental", "x", false, "Enable experimental features")
	Cmd.Flags().BoolVar(&config.Debug, "debug", false, "Output debugging information")
	Cmd.Flags().BoolVar(&dataModel, "datamodel", false, "Output the go yaml data model")
	Cmd.Flags().StringVar(&format.NodeStyle, "node-style", "", format.NodeStyleDocs)
	Cmd.Flags().BoolVar(&cftpkg.NoAnalytics, "no-analytics", false, "No effect, kept for compatibility. cfnrain does not write analytics to Metadata")
}
