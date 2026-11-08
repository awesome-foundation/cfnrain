package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImplicitUploadFailure(t *testing.T) {
	for _, value := range []string{"does-not-exist", "s3://example/artifact.zip", "http://example.com/artifact.zip", "https://example.com/artifact.zip"} {
		t.Run(value, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "template.yaml")
			if err := os.WriteFile(path, []byte("Resources:\n  Function:\n    Type: AWS::Lambda::Function\n    Properties:\n      Code: "+value+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := File(path)
			if value == "does-not-exist" {
				if err == nil || !strings.Contains(err.Error(), "Resources/Function/Properties/Code") || !strings.Contains(err.Error(), value) || !strings.Contains(err.Error(), "no such file") {
					t.Fatalf("expected missing-path error, got %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}
