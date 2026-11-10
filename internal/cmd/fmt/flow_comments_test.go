package fmt

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFlowCollectionComments(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"mapping value", "  Values: [a, b] # c\n  Z: end\n", "\n  # c\n  Values:\n    - a\n    - b\n\n  Z: end\n"},
		{"list item", "  Values:\n    - [a, b] # c\n    - end\n", "  Values:\n    - # c\n      - a\n      - b\n    - end\n"},
		{"function argument", "  Value: !Select [0, [a, b]] # c\n  Z: end\n", "\n  # c\n  Value: !Select\n    - 0\n    - - a\n      - b\n\n  Z: end\n"},
		{"last key", "  Values: [a, b] # c\n", "\n  # c\n  Values:\n    - a\n    - b\n"},
		{"flow mapping", "  Values: {a: 1} # c\n", "\n  # c\n  Values:\n    a: 1\n"},
		{"block list", "  Values: # c\n    - a\n    - b\n", "  Values: # c\n    - a\n    - b\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
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
			cmd := exec.Command(os.Args[0], "-test.run=^TestFlowCollectionVerifyHelper$", "--", "--verify", filename)
			cmd.Env = append(os.Environ(), "CFNRAIN_FLOW_VERIFY_HELPER=1")
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("fmt --verify: %v\n%s", err, output)
			}
		})
	}
}

func TestFlowCollectionVerifyHelper(t *testing.T) {
	if os.Getenv("CFNRAIN_FLOW_VERIFY_HELPER") != "1" {
		return
	}
	if err := Cmd.ParseFlags(os.Args[3:]); err != nil {
		t.Fatal(err)
	}
	Cmd.Run(Cmd, Cmd.Flags().Args())
}
