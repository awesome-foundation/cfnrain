package pkg

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Upstream aws-cloudformation/rain#772: Run failed when the template path was absolute.
func TestRunS3Command(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}

	for _, absolute := range []bool{true, false} {
		dir := t.TempDir()
		appDir := filepath.Join(dir, "myapp")
		if err := os.MkdirAll(filepath.Join(appDir, "scripts"), 0755); err != nil {
			t.Fatal(err)
		}
		script := "#!/bin/sh\ntouch marker\n"
		if err := os.WriteFile(filepath.Join(appDir, "scripts", "build.sh"), []byte(script), 0755); err != nil {
			t.Fatal(err)
		}

		root := appDir
		if !absolute {
			t.Chdir(dir)
			root = "myapp"
		}

		if err := runS3Command(root, "scripts/build.sh"); err != nil {
			t.Errorf("absolute=%v: %v", absolute, err)
			continue
		}
		if _, err := os.Stat(filepath.Join(appDir, "marker")); err != nil {
			t.Errorf("absolute=%v: script did not run in the template directory: %v", absolute, err)
		}
	}
}
