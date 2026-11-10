package fmt

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
)

func TestStrictBooleanStyleNames(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"strict-boolean", "strict-booleans"} {
		t.Run(style, func(t *testing.T) {
			format.NodeStyle = style
			for _, tc := range []struct{ input, want string }{
				{"Metadata:\n  Count: 1\n", "Metadata:\n  Count: 1\n"},
				{"Metadata:\n  Values: [yes, no, on, off, true, false, 1]\n", "Metadata:\n  Values:\n    - \"yes\"\n    - \"no\"\n    - \"on\"\n    - \"off\"\n    - true\n    - false\n    - 1\n"},
			} {
				var first, second result
				formatString(tc.input, &first)
				if first.err != nil {
					t.Fatal(first.err)
				}
				if first.output != tc.want {
					t.Fatalf("first output:\n%s\nwant:\n%s", first.output, tc.want)
				}
				formatString(first.output, &second)
				if second.err != nil {
					t.Fatal(second.err)
				}
				if !second.ok || second.output != first.output {
					t.Fatalf("second output changed:\n%s", second.output)
				}
				filename := filepath.Join(t.TempDir(), "first.yaml")
				if err := os.WriteFile(filename, []byte(first.output), 0600); err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(os.Args[0], "-test.run=^TestStrictBooleanVerifyHelper$", "--", "--node-style", style, "--verify", filename)
				cmd.Env = append(os.Environ(), "CFNRAIN_STRICT_BOOLEAN_VERIFY_HELPER=1")
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("fmt --verify: %v\n%s", err, output)
				}
			}
		})
	}
}

func TestStrictBooleanVerifyHelper(t *testing.T) {
	if os.Getenv("CFNRAIN_STRICT_BOOLEAN_VERIFY_HELPER") != "1" {
		return
	}
	if err := Cmd.ParseFlags(os.Args[3:]); err != nil {
		t.Fatal(err)
	}
	Cmd.Run(Cmd, Cmd.Flags().Args())
}
