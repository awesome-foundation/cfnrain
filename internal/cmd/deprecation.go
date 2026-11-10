package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// WarnDeprecated warns at the CLI boundary so internal calls stay silent.
func WarnDeprecated(name string, requireExperimental bool) func(*cobra.Command, []string) {
	return func(c *cobra.Command, _ []string) {
		experimental, _ := c.Flags().GetBool("experimental")
		if !requireExperimental || experimental {
			fmt.Fprintf(c.ErrOrStderr(), "WARNING: %q is deprecated and will be removed in cfnrain 2.0. See https://github.com/awesome-foundation/cfnrain/discussions/20\n", name)
		}
	}
}
