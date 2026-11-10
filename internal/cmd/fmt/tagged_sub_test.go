package fmt

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
)

func TestTaggedNestedSub(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, argument := range []string{"!Sub |\n      hello ${AWS::Region}\n", "!Sub hello\n", "!Sub ''\n", "!Sub 'true'\n", "!Sub [hello, {}]\n", "|\n      hello ${AWS::Region}\n"} {
		for _, style := range []string{"tagged", "", "flow"} {
			t.Run(style+"/"+argument, func(t *testing.T) {
				format.NodeStyle = style
				var first, second result
				formatString("Metadata:\n  Value:\n    Fn::Base64: "+argument, &first)
				if first.err != nil {
					t.Fatal(first.err)
				}
				formatString(first.output, &second)
				if second.err != nil {
					t.Fatal(second.err)
				}
				if first.output != second.output || !second.ok {
					t.Fatalf("output changed:\nfirst:\n%s\nsecond:\n%s", first.output, second.output)
				}
				filename := filepath.Join(t.TempDir(), "first.yaml")
				if err := os.WriteFile(filename, []byte(first.output), 0600); err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(os.Args[0], "-test.run=^TestNullVerifyHelper$", "--", "--node-style", style, "--verify", filename)
				cmd.Env = append(os.Environ(), "CFNRAIN_NULL_VERIFY_HELPER=1")
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("fmt --verify: %v\n%s", err, output)
				}
			})
		}
	}
}
