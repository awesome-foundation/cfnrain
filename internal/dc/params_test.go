package dc

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft"
	"github.com/aws/smithy-go/ptr"
	"gopkg.in/yaml.v3"
)

func unknownParamsTemplate(t *testing.T) *cft.Template {
	t.Helper()
	var node yaml.Node
	if err := yaml.Unmarshal([]byte("Parameters:\n  Known:\n    Type: String\n"), &node); err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}
	return &cft.Template{Node: &node}
}

func TestGetParametersReportsAllUnknownParams(t *testing.T) {
	params := map[string]string{"Known": "a", "Zulu": "b", "Alpha": "c", "Mike": "d"}

	var got string
	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected a panic for unknown parameters")
			}
			got = fmt.Sprint(r)
		}()
		GetParameters(unknownParamsTemplate(t), params, nil, false, true, false)
	}()

	want := "unknown parameter: Alpha, Mike, Zulu\nif you want to proceed as is, add the --ignore-unknown-params flag"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGetParametersIgnoreUnknownParamsIsSilent(t *testing.T) {
	params := map[string]string{"Known": "a", "Zulu": "b", "Alpha": "c"}
	template := unknownParamsTemplate(t)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	stdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = stdout }()
	got := GetParameters(template, params, nil, false, true, true)
	os.Stdout = stdout
	w.Close()
	out, _ := io.ReadAll(r)

	if len(out) != 0 {
		t.Errorf("expected no output, got %q", out)
	}
	if len(got) != 1 || ptr.ToString(got[0].ParameterKey) != "Known" || ptr.ToString(got[0].ParameterValue) != "a" {
		t.Errorf("unexpected parameters: %v", got)
	}
}
