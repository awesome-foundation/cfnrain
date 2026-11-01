package merge_test

import (
	"os"
	"testing"

	"github.com/awesome-foundation/cfnrain/internal/cmd/merge"
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

func Example_mergeStdin() {
	defer pipeStdin("../../../test/templates/success.template")()

	os.Args = []string{
		os.Args[0],
		"-",
		"../../../test/templates/bucket.yaml",
	}

	merge.Cmd.Execute()
	// Output:
	// Description: This template succeeds
	//
	// Parameters:
	//   AppName:
	//     Description: This name be used as a prefix for resource names
	//     Type: String
	//
	//   BucketName:
	//     Type: String
	//
	// Resources:
	//   Bucket1:
	//     Type: AWS::S3::Bucket
	//     Properties:
	//       BucketName: !Ref BucketName
	//
	//   ObjectStorageBucket:
	//     Type: AWS::S3::Bucket
	//     Properties:
	//       BucketEncryption:
	//         ServerSideEncryptionConfiguration:
	//           - ServerSideEncryptionByDefault:
	//               SSEAlgorithm: AES256
	//       BucketName: !Sub ${AppName}-${AWS::Region}-${AWS::AccountId}
}

func TestMergeStdinTwice(t *testing.T) {
	if err := merge.Cmd.Args(merge.Cmd, []string{"-", "a.yaml", "-"}); err == nil {
		t.Error("expected an error when stdin is given twice")
	}

	if err := merge.Cmd.Args(merge.Cmd, []string{"a.yaml", "b.yaml"}); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
