package diff_test

import (
	"os"
	"testing"

	"github.com/awesome-foundation/cfnrain/internal/cmd/diff"
	"github.com/awesome-foundation/cfnrain/internal/console"
)

// pipeStdin replaces os.Stdin with a pipe that holds the contents of fn.
// The returned function restores os.Stdin and os.Args.
func pipeStdin(fn string) func() {
	data, err := os.ReadFile(fn)
	if err != nil {
		panic(err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	w.Write(data)
	w.Close()

	oldStdin, oldArgs := os.Stdin, os.Args
	os.Stdin = r

	return func() {
		r.Close()
		os.Stdin, os.Args = oldStdin, oldArgs
	}
}

func Example_diffStdin() {
	defer pipeStdin("../../../test/templates/success.template")()

	os.Args = []string{
		os.Args[0],
		"-",
		"../../../test/templates/failure.template",
	}

	oldNoColour := console.NoColour
	defer func() { console.NoColour = oldNoColour }()
	console.NoColour = true

	diff.Cmd.Execute()
	// Output:
	// (>) Description: This template fails
	// (-) Parameters: {...}
	// (|) Resources:
	// (|)   Bucket1:
	// (-)     Properties: {...}
	// (+)   Bucket2:
	// (+)     Properties:
	// (+)       BucketName:
	// (+)         Ref: Bucket1
	// (+)     Type: AWS::S3::Bucket
}

func TestDiffStdinTwice(t *testing.T) {
	if err := diff.Cmd.Args(diff.Cmd, []string{"-", "-"}); err == nil {
		t.Error("expected an error when stdin is given twice")
	}

	if err := diff.Cmd.Args(diff.Cmd, []string{"a.yaml", "b.yaml"}); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
