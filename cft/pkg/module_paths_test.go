package pkg

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
)

func writeModulePathFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// Stop at a missing upload path after a successful script, before any AWS call.
func TestModuleResourceRunPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}
	for _, name := range []string{"module", "nested module", "absolute paths", "top level"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			root := filepath.Join(dir, "modules")
			source := "modules/worker.yaml"
			if name == "nested module" {
				root = filepath.Join(root, "nested")
				writeModulePathFile(t, filepath.Join(dir, source), "Modules:\n  Child:\n    Source: nested/worker.yaml\n")
			}
			if name == "top level" {
				root = dir
			}
			run, path := "build.sh", "dist"
			script := filepath.Join(root, run)
			wantPath := filepath.Join(root, path)
			if name == "absolute paths" {
				script = filepath.Join(dir, "external", "build.sh")
				run = script
				wantPath = filepath.Join(dir, "external", "dist")
				path = wantPath
			}
			writeModulePathFile(t, wantPath, "build-input")
			writeScript(t, script, fmt.Sprintf("test -f %q || exit 98\nrm %q\nprintf 'module-build-ran' > marker\n", wantPath, wantPath))
			body := fmt.Sprintf("Parameters:\n  Artifact:\n    Type: String\n    Default: %q\nResources:\n  Function:\n    Type: AWS::Lambda::Function\n    Properties:\n      Code: !Rain::S3\n        Path: !Ref Artifact\n        Run: %q\n", path, run)
			if name == "top level" {
				body = strings.Replace(body, "Path: !Ref Artifact", fmt.Sprintf("Path: %q", path), 1)
			}
			template := filepath.Join(dir, "template.yaml")
			if name == "top level" {
				writeModulePathFile(t, template, body)
			} else {
				writeModulePathFile(t, filepath.Join(root, "worker.yaml"), body)
				writeModulePathFile(t, template, "Modules:\n  Worker:\n    Source: "+source+"\n")
				// A parent script must never run as a fallback.
				writeScript(t, filepath.Join(dir, "build.sh"), "exit 99\n")
			}
			_, err := File(template)
			if err == nil || !strings.Contains(err.Error(), wantPath) || !strings.Contains(err.Error(), "no such file or directory") {
				t.Fatalf("want missing upload path %s after Run, got %v", wantPath, err)
			}
			marker, err := os.ReadFile(filepath.Join(root, "marker"))
			if err != nil || string(marker) != "module-build-ran" {
				t.Fatalf("Run used the wrong working directory: %q, %v", marker, err)
			}
		})
	}
}

func TestModuleResourceFilePaths(t *testing.T) {
	for _, directive := range []string{"!Rain::Embed", "!Rain::Include", "!Rain::S3", "!Rain::S3Http", "!InsertFile"} {
		for _, absolute := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/absolute=%t", directive, absolute), func(t *testing.T) {
				dir := t.TempDir()
				root := filepath.Join(dir, "modules", "nested")
				path := filepath.Join(root, "content.txt")
				writeModulePathFile(t, path, "module-content")
				writeModulePathFile(t, filepath.Join(dir, "content.txt"), "parent-content")
				for _, cachedPath := range []string{path, filepath.Join(dir, "content.txt"), filepath.Join(dir, "modules", "content.txt")} {
					key := "parent-content"
					if cachedPath == path {
						key = "module-content"
					}
					uploads[cachedPath] = &s3Path{bucket: "example-bucket", key: key, region: "eu-west-1"}
					t.Cleanup(func() { delete(uploads, cachedPath) })
				}
				arg := "content.txt"
				if absolute {
					arg = path
				}
				body := fmt.Sprintf("Resources:\n  Item:\n    Type: Test::Resource\n    Properties:\n      Content: %s %q\n", directive, arg)
				writeModulePathFile(t, filepath.Join(root, "worker.yaml"), body)
				writeModulePathFile(t, filepath.Join(dir, "modules", "outer.yaml"), "Modules:\n  Child:\n    Source: nested/worker.yaml\n")
				template := filepath.Join(dir, "template.yaml")
				writeModulePathFile(t, template, "Modules:\n  Worker:\n    Source: modules/outer.yaml\nResources:\n  Parent:\n    Type: Test::Resource\n    Properties:\n      Content: !Rain::Embed content.txt\n")
				out, err := File(template)
				if err != nil {
					t.Fatal(err)
				}
				text := format.String(out, format.Options{})
				if !strings.Contains(text, "module-content") || !strings.Contains(text, "parent-content") {
					t.Fatalf("wrong file content:\n%s", text)
				}
			})
		}
	}
}

func TestModuleIncludedDirectives(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "modules")
	writeModulePathFile(t, filepath.Join(root, "content.txt"), "module-content")
	writeModulePathFile(t, filepath.Join(root, "parts", "content.txt"), "included-content")
	writeModulePathFile(t, filepath.Join(root, "parts", "part.yaml"), "Content: !Rain::Embed content.txt\n")
	writeModulePathFile(t, filepath.Join(root, "worker.yaml"), `Constants:
  Content: !Rain::Embed content.txt
Resources:
  Item:
    Type: Test::Resource
    Properties:
      Constant: !Rain::Constant Content
      Included: !Rain::Include parts/part.yaml
`)
	template := filepath.Join(dir, "template.yaml")
	writeModulePathFile(t, template, "Modules:\n  Worker:\n    Source: modules/worker.yaml\n")
	out, err := File(template)
	if err != nil {
		t.Fatal(err)
	}
	text := format.String(out, format.Options{})
	if !strings.Contains(text, "Constant: module-content") || !strings.Contains(text, "Content: included-content") {
		t.Fatalf("wrong included content:\n%s", text)
	}
}

// This is the BuildScript/CodePath pattern used by test/webapp/webapp.yaml.
func TestModuleParentSuppliedPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}
	for _, nested := range []bool{false, true} {
		for _, legacy := range []bool{false, true} {
			t.Run(fmt.Sprintf("nested=%t/legacy=%t", nested, legacy), func(t *testing.T) {
				old := Experimental
				Experimental = true
				t.Cleanup(func() { Experimental = old })
				dir := t.TempDir()
				writeScript(t, filepath.Join(dir, "buildapi.sh"), "printf parent > marker\n")
				writeModulePathFile(t, filepath.Join(dir, "p.txt"), "parent-content")
				writeModulePathFile(t, filepath.Join(dir, "modules", "p.txt"), "wrong-content")
				module := `Parameters:
  BuildScript:
    Type: String
  CodePath:
    Type: String
  Text:
    Type: String
Resources:
  Handler:
    Type: Test::Resource
    Properties:
      Text: !Ref Text
      Code: !Rain::S3
        Path: !Ref CodePath
        Run: !Ref BuildScript
Outputs:
  Text:
    Value: !Ref Text
`
				writeModulePathFile(t, filepath.Join(dir, "modules", "api.yaml"), module)
				source := "modules/api.yaml"
				if nested {
					writeModulePathFile(t, filepath.Join(dir, "modules", "outer.yaml"), `Parameters:
  BuildScript:
    Type: String
  CodePath:
    Type: String
  Text:
    Type: String
Modules:
  Inner:
    Source: api.yaml
    Properties:
      BuildScript: !Ref BuildScript
      CodePath: !Ref CodePath
      Text: !Ref Text
`)
					source = "modules/outer.yaml"
				}
				prefix := "Modules:\n  Api:\n    Source: " + source + "\n"
				if legacy {
					prefix = "Resources:\n  Api:\n    Type: !Rain::Module " + source + "\n"
				}
				template := filepath.Join(dir, "template.yaml")
				writeModulePathFile(t, template, prefix+"    Properties:\n      BuildScript: buildapi.sh\n      CodePath: missing.zip\n      Text: !Rain::Embed p.txt\n")
				_, err := File(template)
				if err == nil || !strings.Contains(err.Error(), filepath.Join(dir, "missing.zip")) {
					t.Fatalf("wrong path or unresolved parameter: %v", err)
				}
				marker, err := os.ReadFile(filepath.Join(dir, "marker"))
				if err != nil || string(marker) != "parent" {
					t.Fatalf("wrong Run directory: %s %v", marker, err)
				}
			})
		}
	}
}

func TestModuleOverrideAndOutputOrigins(t *testing.T) {
	dir := t.TempDir()
	writeModulePathFile(t, filepath.Join(dir, "p.txt"), "parent-content")
	writeModulePathFile(t, filepath.Join(dir, "modules", "p.txt"), "module-content")
	writeModulePathFile(t, filepath.Join(dir, "modules", "worker.yaml"), `Parameters:
  Text:
    Type: String
Resources:
  Item:
    Type: Test::Resource
    Properties:
      Content: !Rain::Embed p.txt
      Supplied: !Ref Text
Outputs:
  Text:
    Value: !Rain::Embed p.txt
  Inserted:
    Value: !InsertFile p.txt
`)
	template := filepath.Join(dir, "template.yaml")
	writeModulePathFile(t, template, `Modules:
  Worker:
    Source: modules/worker.yaml
    Properties:
      Text: !Rain::Embed p.txt
    Overrides:
      Item:
        Properties:
          Content: !Rain::Embed p.txt
Outputs:
  Text:
    Value: !GetAtt Worker.Text
  Inserted:
    Value: !GetAtt Worker.Inserted
`)
	out, err := File(template)
	if err != nil {
		t.Fatal(err)
	}
	text := format.String(out, format.Options{})
	if !strings.Contains(text, "Supplied: parent-content") || !strings.Contains(text, "Content: parent-content") || strings.Count(text, "Value: module-content") != 2 {
		t.Fatalf("wrong origin:\n%s", text)
	}
}

func TestZipModulePathsIgnoreWorkingDirectory(t *testing.T) {
	dir := t.TempDir()
	writeModulePathFile(t, filepath.Join(dir, "p.txt"), "parent-content")
	f, err := os.Create(filepath.Join(dir, "modules.zip"))
	if err != nil {
		t.Fatal(err)
	}
	z := zip.NewWriter(f)
	entry, err := z.Create("worker.yaml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = entry.Write([]byte("Resources:\n  Item:\n    Type: Test::Resource\n    Properties:\n      Content: !Rain::Embed p.txt\n"))
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	template := filepath.Join(dir, "template.yaml")
	writeModulePathFile(t, template, "Modules:\n  Worker:\n    Source: modules.zip/worker.yaml\n")
	for _, cwd := range []string{dir, t.TempDir()} {
		t.Chdir(cwd)
		out, err := File(template)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(format.String(out, format.Options{}), "Content: parent-content") {
			t.Fatal("zip module used working directory")
		}
	}
}

func TestLegacyNestedSourcePaths(t *testing.T) {
	old := Experimental
	Experimental = true
	t.Cleanup(func() { Experimental = old })
	dir := t.TempDir()
	writeModulePathFile(t, filepath.Join(dir, "modules", "nested", "p.txt"), "nested-content")
	writeModulePathFile(t, filepath.Join(dir, "modules", "nested", "worker.yaml"), "Resources:\n  Item:\n    Type: Test::Resource\n    Properties:\n      Content: !Rain::Embed p.txt\n")
	writeModulePathFile(t, filepath.Join(dir, "modules", "outer.yaml"), "Resources:\n  Inner:\n    Type: !Rain::Module nested/worker.yaml\n")
	template := filepath.Join(dir, "template.yaml")
	writeModulePathFile(t, template, "Resources:\n  Outer:\n    Type: !Rain::Module modules/outer.yaml\n")
	out, err := File(template)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(format.String(out, format.Options{}), "Content: nested-content") {
		t.Fatal("wrong nested source directory")
	}
}
