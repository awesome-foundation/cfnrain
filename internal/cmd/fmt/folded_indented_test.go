package fmt

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
				checkFoldedStringFormat(t, prefix+strconv.Quote(value)+"\n")
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

func checkFoldedStringFormat(t *testing.T, input string) string {
	t.Helper()
	var first, second result
	formatString(input, &first)
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
	return first.output
}

func TestFoldedBlockStyleControls(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	format.NodeStyle = "folded"
	for _, tc := range []struct{ name, input, want string }{
		{"indented folded", "Description: " + strconv.Quote("\n\n    deep\n x\n") + "\n", "\"Description\": >2\n\n\n      deep\n   x\n"},
		{"list folded", "Metadata:\n  Values: [" + strconv.Quote("\n\nfirst\n") + "]\n", "\"Metadata\":\n  \"Values\":\n    - >2\n\n\n      first\n\n\n"},
		{"plain folded", "Description: " + strconv.Quote("first\nsecond\n") + "\n", "\"Description\": >\n  first\n\n  second\n\n"},
		{"literal fallback", "Description: " + strconv.Quote("first\n  indented\nlast") + "\n", "\"Description\": |-\n  first\n    indented\n  last\n"},
		{"quoted fallback", "Description: " + strconv.Quote("\nfirst\n") + "\n", "\"Description\": \"\\nfirst\\n\"\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := checkFoldedStringFormat(t, tc.input); got != tc.want {
				t.Fatalf("output:\n%s\nwant:\n%s", got, tc.want)
			}
		})
	}
	input, err := os.ReadFile("../../../test/templates/fmtmulti.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if got := checkFoldedStringFormat(t, string(input)); !strings.Contains(got, "\"ZipFile\": |\n") {
		t.Fatalf("ZipFile did not use literal style:\n%s", got)
	}
}
