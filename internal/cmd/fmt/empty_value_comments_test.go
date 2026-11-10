package fmt

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
)

func TestEmptyValueComments(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, style := range []string{"", "original"} {
		for _, tc := range []struct{ name, input, want string }{
			{"anchor", "  V: &a # c\n  W: *a\n", "  V: &a null # c\n\n  W: *a\n"},
			{"anchor control", "  V: &a 1 # c\n  W: *a\n", "  V: &a 1 # c\n\n  W: *a\n"},
			{"GetAZs", "  Q: !GetAZs # c\n  R: next\n", "  Q: !GetAZs \"\" # c\n\n  R: next\n"},
			{"GetAZs control", "  Q: !GetAZs '' # c\n  R: next\n", "  Q: !GetAZs \"\" # c\n\n  R: next\n"},
			{"flow", "  Values: {a: , # c\n    b: 1}\n", "  Values:\n    a: # c\n    b: 1\n"},
			{"flow control", "  Values: {a: 1, # c\n    b: 2}\n", "  Values:\n    a: 1 # c\n    b: 2\n"},
			{"Join", "  Q: !Join\n    - '-'\n    - - # c\n      - x\n", "  Q: !Join\n    - '-'\n    - - !!null \"\" # c\n      - x\n"},
			{"Join explicit null", "  Q: !Join\n    - '-'\n    - - null # c\n      - x\n", "  Q: !Join\n    - '-'\n    - - null # c\n      - x\n"},
		} {
			t.Run(style+"/"+tc.name, func(t *testing.T) {
				format.NodeStyle = style
				var first, second result
				formatString("Metadata:\n"+tc.input+"Resources: {}\n", &first)
				if first.err != nil {
					t.Fatal(first.err)
				}
				want := "Metadata:\n" + tc.want + "\nResources: {}\n"
				if first.output != want {
					t.Fatalf("first output:\n%s\nwant:\n%s", first.output, want)
				}
				formatString(first.output, &second)
				if second.err != nil {
					t.Fatal(second.err)
				}
				if second.output != first.output {
					t.Fatalf("second output changed:\n%s", second.output)
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
