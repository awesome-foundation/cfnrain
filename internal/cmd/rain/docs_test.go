package rain

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestEveryCommandIsDocumented fails when a command is not named on any
// page of the documentation site, so a new command cannot ship undocumented
func TestEveryCommandIsDocumented(t *testing.T) {
	var docs strings.Builder
	err := filepath.WalkDir("../../../docs", func(path string, d fs.DirEntry, err error) error {
		if err != nil || filepath.Ext(path) != ".md" {
			return err
		}
		b, err := os.ReadFile(path)
		docs.Write(b)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	text := docs.String()

	var check func(c *cobra.Command)
	check = func(c *cobra.Command) {
		for _, sub := range c.Commands() {
			if !sub.IsAvailableCommand() {
				continue
			}
			if !regexp.MustCompile(regexp.QuoteMeta(sub.CommandPath()) + `\b`).MatchString(text) {
				t.Errorf("%q is not on any page in docs/", sub.CommandPath())
			}
			check(sub)
		}
	}
	check(Cmd)
}
