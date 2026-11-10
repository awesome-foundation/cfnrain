package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClonedAliasCycles(t *testing.T) {
	oldExperimental := Experimental
	Experimental = true
	t.Cleanup(func() { Experimental = oldExperimental })
	for _, tc := range []struct{ name, input, module string }{
		{"constant", `Constants:
  C: &c {a: *c}
Resources:
  B:
    Type: AWS::S3::Bucket
    Properties:
      Tags: !Ref Const::C
`, ""},
		{"module body", `Resources:
  Example:
    Type: !Rain::Module module.yaml
`, `Resources:
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      Tags: &c [*c]
`},
		{"module overrides", `Resources:
  Example:
    Type: !Rain::Module module.yaml
    Overrides:
      Bucket:
        Properties:
          Tags: &c [*c]
`, `Resources:
  Bucket:
    Type: AWS::S3::Bucket
`},
	} {
		for _, fallback := range []bool{false, true} {
			name := tc.name + "/cycle"
			if fallback {
				name = tc.name + "/fallback"
			}
			t.Run(name, func(t *testing.T) {
				dir := t.TempDir()
				for name, content := range map[string]string{"template.yaml": tc.input, "module.yaml": tc.module} {
					if fallback {
						content = strings.ReplaceAll(content, "&c {a: *c}", "[&c {a: *c}, &c {a: safe}]")
						content = strings.ReplaceAll(content, "&c [*c]", "[&c [*c], &c [safe]]")
					}
					if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
						t.Fatal(err)
					}
				}
				_, err := File(filepath.Join(dir, "template.yaml"))
				if fallback {
					if err != nil {
						t.Fatal(err)
					}
					return
				}
				if err == nil || !strings.Contains(err.Error(), "recursive") || !strings.Contains(err.Error(), `c"`) {
					t.Fatalf("error = %v, want a named recursive alias error", err)
				}
			})
		}
	}
}
