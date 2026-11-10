package fmt_test

import (
	"os"
	"testing"

	rainfmt "github.com/awesome-foundation/cfnrain/internal/cmd/fmt"
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

func Example_fmtStdin() {
	defer pipeStdin("../../../test/templates/success.template")()

	os.Args = []string{os.Args[0], "-"}

	rainfmt.Cmd.Execute()
	// Output:
	// Description: This template succeeds
	//
	// Parameters:
	//   BucketName:
	//     Type: String
	//
	// Resources:
	//   Bucket1:
	//     Type: AWS::S3::Bucket
	//     Properties:
	//       BucketName: !Ref BucketName
}

func TestFmtStdinTwice(t *testing.T) {
	if err := rainfmt.Cmd.Args(rainfmt.Cmd, []string{"-", "-"}); err == nil {
		t.Error("expected an error when stdin is given twice")
	}
}

func TestFmtWriteStdin(t *testing.T) {
	defer pipeStdin("../../../test/templates/success.template")()

	os.Args = []string{os.Args[0], "-w", "-"}
	defer func() {
		write := rainfmt.Cmd.Flags().Lookup("write")
		write.Value.Set(write.DefValue)
		write.Changed = false
	}()

	if err := rainfmt.Cmd.Execute(); err == nil {
		t.Error("expected an error for --write with stdin")
	}
}
