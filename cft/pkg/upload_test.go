package pkg

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/internal/aws/s3"
)

func TestUploadCacheExtension(t *testing.T) {
	// Isolate the cached AWS configuration from other package tests.
	if path := os.Getenv("CFNRAIN_UPLOAD_TEST_TEMPLATE"); path != "" {
		s3.BucketName = "test-bucket"
		result, err := File(path)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Print(format.String(result, format.Options{}))
		return
	}

	for _, tc := range []struct {
		name, first, second string
		wantUploads         int
	}{
		{"different", "json", "txt", 2},
		{"same", "json", "json", 1},
		{"omitted", "", "", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var mu sync.Mutex
			var keys []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				_, _ = io.Copy(io.Discard, r.Body)
				switch r.Method {
				case http.MethodPost:
					fmt.Fprint(w, "<GetCallerIdentityResponse><GetCallerIdentityResult><Account>000000000000</Account></GetCallerIdentityResult></GetCallerIdentityResponse>")
				case http.MethodHead:
					w.WriteHeader(http.StatusOK)
				case http.MethodPut:
					mu.Lock()
					keys = append(keys, strings.TrimPrefix(r.URL.Path, "/test-bucket/"))
					mu.Unlock()
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
					w.WriteHeader(http.StatusBadRequest)
				}
			}))
			defer server.Close()

			dir := t.TempDir()
			content := []byte("{}\n")
			if err := os.WriteFile(filepath.Join(dir, "config.json"), content, 0600); err != nil {
				t.Fatal(err)
			}
			source := "Resources:\n"
			var wantKeys []string
			for i, extension := range []string{tc.first, tc.second} {
				source += fmt.Sprintf("  Template%d: !Rain::S3\n    Path: config.json\n", i)
				key := fmt.Sprintf("%x", sha256.Sum256(content))
				if extension != "" {
					source += "    Extension: " + extension + "\n"
					key += "." + extension
				}
				wantKeys = append(wantKeys, key)
			}
			path := filepath.Join(dir, "template.yaml")
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(os.Args[0], "-test.run=^TestUploadCacheExtension$")
			for _, entry := range os.Environ() {
				if !strings.HasPrefix(entry, "AWS_") && !strings.HasPrefix(entry, "CFNRAIN_UPLOAD_TEST_TEMPLATE=") {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			cmd.Env = append(cmd.Env,
				"CFNRAIN_UPLOAD_TEST_TEMPLATE="+path,
				"AWS_ACCESS_KEY_ID=test", "AWS_SECRET_ACCESS_KEY=test",
				"AWS_REGION=us-east-1", "AWS_EC2_METADATA_DISABLED=true",
				"AWS_CONFIG_FILE="+filepath.Join(dir, "no-config"),
				"AWS_SHARED_CREDENTIALS_FILE="+filepath.Join(dir, "no-credentials"),
				"AWS_ENDPOINT_URL="+server.URL)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("package template: %v\n%s", err, output)
			}
			for i, key := range wantKeys {
				want := fmt.Sprintf("Template%d: s3://test-bucket/%s", i, key)
				if !strings.Contains(string(output), want) {
					t.Errorf("want %s, got:\n%s", want, output)
				}
			}
			mu.Lock()
			defer mu.Unlock()
			if len(keys) != tc.wantUploads {
				t.Errorf("want %d uploads, got %d: %v", tc.wantUploads, len(keys), keys)
			}
		})
	}
}
