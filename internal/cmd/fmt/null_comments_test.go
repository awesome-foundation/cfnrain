package fmt

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNullComments(t *testing.T) {
	for _, value := range []string{`!!null ""`, "!!null", "~", "null", "", `""`} {
		for _, mapping := range []bool{false, true} {
			name := "list/" + value
			body := "  Values:\n    - " + value + " # c\n    - end\n"
			wantValue := value
			if value == "" || strings.HasPrefix(value, "!!null") {
				wantValue = "null"
			}
			wantBody := "  Values:\n    - " + wantValue + " # c\n    - end\n"
			if mapping {
				name = "mapping/" + value
				body = "  Value: " + value + " # c\n  Z: end\n"
				if value == "" || strings.HasPrefix(value, "!!null") {
					wantBody = "  Value: # c\n\n  Z: end\n"
				} else {
					wantBody = "  Value: " + wantValue + " # c\n\n  Z: end\n"
				}
			}
			t.Run(name, func(t *testing.T) {
				var first, second result
				formatString("Metadata:\n"+body+"Resources: {}\n", &first)
				if first.err != nil {
					t.Fatal(first.err)
				}
				want := "Metadata:\n" + wantBody + "\nResources: {}\n"
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
				var decoded struct {
					Metadata map[string]interface{} `yaml:"Metadata"`
				}
				if err := yaml.Unmarshal([]byte(first.output), &decoded); err != nil {
					t.Fatal(err)
				}
				got := decoded.Metadata["Value"]
				if !mapping {
					got = decoded.Metadata["Values"].([]interface{})[0]
				}
				if value == `""` {
					if got != "" {
						t.Fatalf("empty string became %#v", got)
					}
				} else if got != nil {
					t.Fatalf("null became %#v", got)
				}
				filename := filepath.Join(t.TempDir(), "first.yaml")
				if err := os.WriteFile(filename, []byte(first.output), 0600); err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(os.Args[0], "-test.run=^TestNullVerifyHelper$", "--", "--verify", filename)
				cmd.Env = append(os.Environ(), "CFNRAIN_NULL_VERIFY_HELPER=1")
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("fmt --verify: %v\n%s", err, output)
				}
			})
		}
	}
}

func TestNullVerifyHelper(t *testing.T) {
	if os.Getenv("CFNRAIN_NULL_VERIFY_HELPER") != "1" {
		return
	}
	if err := Cmd.ParseFlags(os.Args[3:]); err != nil {
		t.Fatal(err)
	}
	Cmd.Run(Cmd, Cmd.Flags().Args())
}
