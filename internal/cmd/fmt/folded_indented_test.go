package fmt

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
)

func TestFoldedIndentedStrings(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	format.NodeStyle = "folded"
	for _, value := range []string{
		"first\n  indented\nlast",
		"first\n\n  indented\nlast\n",
		"first\n\tindented\nlast\n\n",
		"  first\nsecond\nthird",
		"  first\n  second\n",
		"first\nsecond\n",
		"\n\nfirst\n  indented\nlast\n\n",
	} {
		for _, prefix := range []string{"Description: ", "Metadata:\n  Text: ", "Metadata:\n  List:\n    - ", "Outputs:\n  Script:\n    Value: !Sub "} {
			t.Run(prefix+strconv.Quote(value), func(t *testing.T) {
				var first, second result
				formatString(prefix+strconv.Quote(value)+"\n", &first)
				if first.err != nil {
					t.Fatal(first.err)
				}
				formatString(first.output, &second)
				if second.err != nil || !second.ok || first.output != second.output {
					t.Fatalf("second format: %v\n%s", second.err, second.output)
				}
				filename := filepath.Join(t.TempDir(), "formatted.yaml")
				if err := os.WriteFile(filename, []byte(first.output), 0600); err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(os.Args[0], "-test.run=^TestFoldedIndentedVerifyHelper$", "--", "--node-style", "folded", "--verify", filename)
				cmd.Env = append(os.Environ(), "CFNRAIN_FOLDED_INDENTED_VERIFY_HELPER=1")
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("fmt --verify: %v\n%s", err, output)
				}
			})
		}
	}
}

func TestFoldedIndentedVerifyHelper(t *testing.T) {
	if os.Getenv("CFNRAIN_FOLDED_INDENTED_VERIFY_HELPER") != "1" {
		return
	}
	if err := Cmd.ParseFlags(os.Args[3:]); err != nil {
		t.Fatal(err)
	}
	Cmd.Run(Cmd, Cmd.Flags().Args())
}
