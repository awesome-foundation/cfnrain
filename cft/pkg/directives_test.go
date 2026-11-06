package pkg

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// writeScript writes an executable shell script to path
func writeScript(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0755); err != nil {
		t.Fatal(err)
	}
}

// Upstream aws-cloudformation/rain#772: Run failed when the template path was absolute.
func TestRunS3Command(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}

	script := func(appDir string) string { return filepath.Join(appDir, "scripts", "build.sh") }
	cases := []struct {
		name string
		cwd  func(dir, appDir string) string // the working directory, or "" to keep it
		root func(appDir string) string      // the template directory as pkg gets it
		run  func(appDir string) string      // the Run value
	}{
		{"absolute root", nil, func(appDir string) string { return appDir }, func(string) string { return "scripts/build.sh" }},
		{"relative root", func(dir, _ string) string { return dir }, func(string) string { return "myapp" }, func(string) string { return "scripts/build.sh" }},
		{"dot root", func(_, appDir string) string { return appDir }, func(string) string { return "." }, func(string) string { return "scripts/build.sh" }},
		// An absolute Run path is used as it is, as Path already does
		{"absolute run", nil, func(appDir string) string { return appDir }, script},
		{"absolute run, relative root", func(dir, _ string) string { return dir }, func(string) string { return "myapp" }, script},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			appDir := filepath.Join(dir, "myapp")
			writeScript(t, script(appDir), "touch marker\n")
			if c.cwd != nil {
				t.Chdir(c.cwd(dir, appDir))
			}

			if err := runS3Command(c.root(appDir), c.run(appDir)); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(appDir, "marker")); err != nil {
				t.Errorf("script did not run in the template directory: %v", err)
			}
		})
	}
}

// A failed script must name the script and show what it wrote to stderr
func TestRunS3CommandFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}

	cases := []struct {
		name, body string
		want       []string
		notWant    []string
	}{
		{"stderr", "echo 'compiler output'\necho '  missing dependency  ' >&2\nexit 42\n",
			[]string{"run build.sh: exit status 42: missing dependency"}, []string{"compiler output"}},
		{"no stderr", "exit 3\n", []string{"run build.sh: exit status 3"}, []string{"exit status 3:"}},
		{"long stderr", "i=0; while [ $i -lt 2000 ]; do echo \"line $i\" >&2; i=$((i+1)); done; exit 1\n",
			[]string{"[...] ", "line 1999"}, []string{"line 0\n"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			writeScript(t, filepath.Join(dir, "build.sh"), c.body)

			err := runS3Command(dir, "build.sh")
			if err == nil {
				t.Fatal("expected an error")
			}
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Errorf("error %q does not wrap *exec.ExitError", err)
			}
			for _, w := range c.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q does not contain %q", err, w)
				}
			}
			for _, w := range c.notWant {
				if strings.Contains(err.Error(), w) {
					t.Errorf("error %q contains %q", err, w)
				}
			}
			if len(err.Error()) > maxRunStderr+200 {
				t.Errorf("error is %d bytes long", len(err.Error()))
			}
		})
	}
}

// After a Run script, the upload must read the files again, because the script
// may have changed them. Without Run, the cached upload is used.
func TestHandleS3RunDropsCachedUpload(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}

	for _, zip := range []bool{false, true} {
		dir := t.TempDir()
		dist := filepath.Join(dir, "dist")
		if err := os.MkdirAll(dist, 0755); err != nil {
			t.Fatal(err)
		}
		// The script removes dist, so a real upload fails with a file error
		writeScript(t, filepath.Join(dir, "build.sh"), "rm -rf dist\n")

		key := dist
		if zip {
			key = "zip:" + dist
		}
		cached := &s3Path{bucket: "cached-bucket", key: "cached-key"}
		uploads[key] = cached
		t.Cleanup(func() { delete(uploads, key) })

		// Without Run, the cached upload is used
		n, err := handleS3(dir, s3Options{Path: "dist", Zip: zip})
		if err != nil || !strings.Contains(n.Value, "cached-key") {
			t.Fatalf("zip=%v, no Run: want the cached upload, got %v, %v", zip, n, err)
		}

		// With Run, the cache is dropped, and the upload finds no dist
		_, err = handleS3(dir, s3Options{Path: "dist", Zip: zip, Run: "build.sh"})
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("zip=%v, Run: want a file-not-found error from a fresh upload, got %v", zip, err)
		}
	}
}

// A failed Run must keep the cache as it was
func TestForgetUploadsOnlyAfterSuccess(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}

	dir := t.TempDir()
	dist := filepath.Join(dir, "dist")
	writeScript(t, filepath.Join(dir, "build.sh"), "exit 1\n")
	uploads[dist] = &s3Path{bucket: "b", key: "k"}
	t.Cleanup(func() { delete(uploads, dist) })

	if _, err := handleS3(dir, s3Options{Path: "dist", Run: "build.sh"}); err == nil {
		t.Fatal("expected the Run error")
	}
	if _, ok := uploads[dist]; !ok {
		t.Error("a failed Run dropped the cached upload")
	}
}
