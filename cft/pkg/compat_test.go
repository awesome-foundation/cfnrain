package pkg

import (
	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
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

func TestPackagingErrorContext(t *testing.T) {
	for _, tc := range []struct {
		name, resource, property, value string
		want                            []string
	}{
		{"list", "AWS::Lambda::Function", "Layers", "[!Rain::S3 nope.zip]", []string{"Resources/Function/Properties/Layers/0", "nope.zip"}},
		{"nested", "AWS::CloudFormation::Stack", "TemplateURL", "child.yaml", []string{"child.yaml", "Resources/Child/Properties/Code", "gone.zip"}},
		{"CodeUri", "AWS::Serverless::Function", "CodeUri", "s3://example/artifact.zip", nil},
		{"Ref", "AWS::Lambda::Function", "Code", "!Ref Artifact", nil},
		{"Sub", "AWS::Serverless::Function", "CodeUri", "!Sub 's3://${Bucket}/artifact.zip'", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "parent.yaml")
			source := "Resources:\n  Function:\n    Type: " + tc.resource + "\n    Properties:\n      " + tc.property + ": " + tc.value + "\n"
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "child.yaml"), []byte("Resources:\n  Child:\n    Type: AWS::Lambda::Function\n    Properties:\n      Code: gone.zip\n"), 0600); err != nil {
				t.Fatal(err)
			}
			result, err := File(path)
			if tc.want != nil {
				if err == nil {
					t.Fatal("expected packaging error")
				}
				for _, want := range tc.want {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error %q does not contain %q", err, want)
					}
				}
				if strings.Contains(err.Error(), ": :") {
					t.Errorf("empty error path: %v", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				expected, err := parse.String(source)
				if err != nil {
					t.Fatal(err)
				}
				if got, want := format.String(result, format.Options{}), format.String(expected, format.Options{}); got != want {
					t.Errorf("pass-through changed template:\ngot %s\nwant %s", got, want)
				}
			}
		})
	}
}
