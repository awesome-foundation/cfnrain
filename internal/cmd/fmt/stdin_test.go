package fmt_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/awesome-foundation/cfnrain/internal/cmd"
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

func TestFmtStdinOutput(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		mixed       bool
		code        int
		want        string
	}{
		{name: "empty", input: "", code: 1, want: ""},
		{name: "mixed", input: "Description: stdin\n", mixed: true, want: "--- # <stdin>\nDescription: stdin\n--- # other.yaml\nDescription: file\n...\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			input := filepath.Join(dir, "stdin.yaml")
			if err := os.WriteFile(input, []byte(tc.input), 0600); err != nil {
				t.Fatal(err)
			}
			defer pipeStdin(input)()
			t.Chdir(dir)
			args := []string{"-"}
			if tc.mixed {
				if err := os.WriteFile("other.yaml", []byte("Description: file\n"), 0600); err != nil {
					t.Fatal(err)
				}
				args = append(args, "other.yaml")
			}
			output, err := os.Create(filepath.Join(dir, "stdout"))
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			oldStdout := os.Stdout
			os.Stdout = output
			defer func() { os.Stdout = oldStdout }()
			os.Args = append([]string{os.Args[0]}, args...)
			if code := cmd.Test(rainfmt.Cmd); code != tc.code {
				t.Fatalf("exit status = %d, want %d", code, tc.code)
			}
			got, err := os.ReadFile(output.Name())
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Fatalf("output = %q, want %q", got, tc.want)
			}
		})
	}
}
