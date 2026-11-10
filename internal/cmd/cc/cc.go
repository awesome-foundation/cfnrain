package cc

import (
	"github.com/awesome-foundation/cfnrain/cft"
	"github.com/awesome-foundation/cfnrain/internal/aws/s3"
	"github.com/awesome-foundation/cfnrain/internal/cmd"
	"github.com/awesome-foundation/cfnrain/internal/config"
	"github.com/awesome-foundation/cfnrain/plugins/deployconfig"
	"github.com/spf13/cobra"
)

// Args
var params []string
var tags []string
var configFilePath string
var Experimental bool
var yes bool
var ignoreUnknownParams bool
var unlock string

// Globals (seems bad..? but cumbersome to pass them around)
var deployedTemplate *cft.Template
var resMap map[string]*Resource
var templateConfig *deployconfig.DeployConfig

var Cmd = &cobra.Command{
	Use:   "cc <command>",
	Short: "Interact with templates using Cloud Control API instead of CloudFormation (deprecated)",
	Long: `You must pass the --experimental (-x) flag to use this command, to acknowledge that it is experimental and likely to be unstable!
This command is deprecated and will be removed in cfnrain 2.0.
`,
	PersistentPreRun: cmd.WarnDeprecated("cfnrain cc", true),
}

func addCommonParams(c *cobra.Command) {
	c.Short += " (deprecated)"
	c.Long += "\nThis command is deprecated and will be removed in cfnrain 2.0.\n"
	c.Flags().StringVarP(&config.Profile, "profile", "p", "", "AWS profile name; read from the AWS CLI configuration file")
	c.Flags().StringVarP(&config.Region, "region", "r", "", "AWS region to use")

	c.Flags().StringVar(&s3.BucketName, "s3-bucket", "", "Name of the S3 bucket that is used to upload assets")
	c.Flags().StringVar(&s3.BucketKeyPrefix, "s3-prefix", "", "Prefix to add to objects uploaded to S3 bucket")
	c.Flags().BoolVar(&config.Debug, "debug", false, "Output debugging information")
	c.Flags().BoolVarP(&Experimental, "experimental", "x", false, "Acknowledge that this is an experimental feature")
}

func init() {
	Cmd.AddCommand(CCDeployCmd)
	Cmd.AddCommand(CCRmCmd)
	Cmd.AddCommand(CCStateCmd)
	Cmd.AddCommand(CCDriftCmd)
}
