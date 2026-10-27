package pkg

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestUnzipRejectsPathTraversal(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "evil.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	fw, err := w.Create("../escaped.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()

	f, err = os.Open(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	dest := filepath.Join(tmp, "dest")
	if err := Unzip(f, dest); err == nil {
		t.Fatal("expected an error for a zip entry outside the destination")
	}
	if _, err := os.Stat(filepath.Join(tmp, "escaped.txt")); !os.IsNotExist(err) {
		t.Fatal("zip entry was written outside the destination")
	}
}

func TestSafeJoin(t *testing.T) {
	cases := map[string]bool{
		"a.txt":        true,
		"dir/a.txt":    true,
		"../a.txt":     false,
		"dir/../../a":  false,
		"/etc/passwd":  true, // Join keeps it under dir
		"..":           false,
		"dir/../a.txt": true,
	}
	for name, ok := range cases {
		_, err := safeJoin("/tmp/dest", name)
		if (err == nil) != ok {
			t.Errorf("safeJoin(%q): got err=%v, want ok=%v", name, err, ok)
		}
	}
}
