package rain

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/parse"
	"github.com/awesome-foundation/cfnrain/internal/cmd/build"
	"github.com/awesome-foundation/cfnrain/internal/cmd/cc"
	"github.com/awesome-foundation/cfnrain/internal/cmd/forecast"
	"github.com/spf13/cobra"
)

func TestDeprecationWarnings(t *testing.T) {
	oldCompletion := Cmd.CompletionOptions.DisableDefaultCmd
	Cmd.CompletionOptions.DisableDefaultCmd = true
	defer func() { Cmd.CompletionOptions.DisableDefaultCmd = oldCompletion }()
	for _, tc := range []struct {
		name        string
		command     *cobra.Command
		args        []string
		warningName string
	}{
		{"build", build.Cmd, []string{"build"}, "cfnrain build"},
		{"forecast", forecast.Cmd, []string{"forecast", "-x", "template.yaml"}, "cfnrain forecast"},
		{"cc state", cc.CCStateCmd, []string{"cc", "state", "--experimental", "example"}, "cfnrain cc"},
		{"cc rm", cc.CCRmCmd, []string{"cc", "rm", "-x", "example"}, "cfnrain cc"},
		{"cc drift", cc.CCDriftCmd, []string{"cc", "drift", "-x", "example"}, "cfnrain cc"},
		{"cc deploy", cc.CCDeployCmd, []string{"cc", "deploy", "-x", "template.yaml", "example"}, "cfnrain cc"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			oldRun := tc.command.Run
			defer func() { tc.command.Run = oldRun }()
			defer func() {
				forecast.Experimental = false
				cc.Experimental = false
				Cmd.SetOut(nil)
				Cmd.SetErr(nil)
				Cmd.SetArgs(nil)
			}()
			var stdout, stderr bytes.Buffer
			Cmd.SetOut(&stdout)
			Cmd.SetErr(&stderr)
			want := fmt.Sprintf("WARNING: %q is deprecated and will be removed in cfnrain 2.0. See https://github.com/awesome-foundation/cfnrain/discussions/20\n", tc.warningName)
			calls := 0
			tc.command.Run = func(c *cobra.Command, _ []string) {
				calls++
				if stderr.String() != want {
					t.Fatalf("stderr before command work = %q, want %q", stderr.String(), want)
				}
				if tc.command == cc.CCDeployCmd {
					template, err := parse.String("Resources: {}")
					if err != nil {
						t.Fatal(err)
					}
					forecast.PredictTotalEstimate(template, false)
				}
				fmt.Fprintln(c.OutOrStdout(), "command output")
			}
			// Each execution must warn once, even in the same process.
			for range 2 {
				stdout.Reset()
				stderr.Reset()
				Cmd.SetArgs(tc.args)
				if err := Cmd.Execute(); err != nil {
					t.Fatal(err)
				}
				if stderr.String() != want {
					t.Fatalf("stderr = %q, want %q", stderr.String(), want)
				}
				if stdout.String() != "command output\n" {
					t.Fatalf("stdout = %q", stdout.String())
				}
			}
			if calls != 2 {
				t.Fatalf("command ran %d times", calls)
			}
		})
	}
}

func TestDeprecationWarningRequiresExperimental(t *testing.T) {
	for _, c := range []*cobra.Command{forecast.Cmd, cc.CCStateCmd} {
		var stderr bytes.Buffer
		c.SetErr(&stderr)
		defer c.SetErr(nil)
		if err := c.Flags().Set("experimental", "false"); err != nil {
			t.Fatal(err)
		}
		if c == forecast.Cmd {
			c.PreRun(c, nil)
		} else {
			cc.Cmd.PersistentPreRun(c, nil)
		}
		if stderr.Len() != 0 {
			t.Fatalf("stderr without --experimental = %q", stderr.String())
		}
	}
}
