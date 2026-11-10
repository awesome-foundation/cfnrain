package fmt

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
)

func TestUnsortedKeyComments(t *testing.T) {
	oldStyle, oldUnsorted := format.NodeStyle, unsortedFlag
	t.Cleanup(func() { format.NodeStyle, unsortedFlag = oldStyle, oldUnsorted })
	unsortedFlag = true
	for _, style := range []string{"flow", "tagged"} {
		for _, input := range []string{
			"Resources:\n  R: # rline\n    Type: AWS::SNS::Topic\n",
			"Metadata: # section\n  Z: # collection\n    - a\n    - b\n  A: last\n",
			"Metadata:\n  Z: # scalar\n    text\n  A: {} # empty\n",
			"Metadata:\n  # head\n  Z: # key\n    K: v\n  A: last\n",
		} {
			t.Run(style+"/"+input, func(t *testing.T) {
				format.NodeStyle = style
				var first, second result
				formatString(input, &first)
				if first.err != nil {
					t.Fatal(first.err)
				}
				for _, comment := range []string{"# rline", "# section", "# collection", "# scalar", "# empty", "# head", "# key"} {
					if strings.Count(first.output, comment) != strings.Count(input, comment) {
						t.Fatalf("comment %q lost or duplicated:\n%s", comment, first.output)
					}
				}
				formatString(first.output, &second)
				if second.err != nil || !second.ok || first.output != second.output {
					t.Fatalf("second format: %v\n%s", second.err, second.output)
				}
				filename := filepath.Join(t.TempDir(), "formatted.yaml")
				if err := os.WriteFile(filename, []byte(first.output), 0600); err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(os.Args[0], "-test.run=^TestUnsortedKeyCommentVerifyHelper$", "--", "--unsorted", "--node-style", style, "--verify", filename)
				cmd.Env = append(os.Environ(), "CFNRAIN_UNSORTED_KEY_VERIFY_HELPER=1")
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("fmt --verify: %v\n%s", err, output)
				}
			})
		}
	}
}

func TestUnsortedKeyCommentVerifyHelper(t *testing.T) {
	if os.Getenv("CFNRAIN_UNSORTED_KEY_VERIFY_HELPER") != "1" {
		return
	}
	if err := Cmd.ParseFlags(os.Args[3:]); err != nil {
		t.Fatal(err)
	}
	Cmd.Run(Cmd, Cmd.Flags().Args())
}
