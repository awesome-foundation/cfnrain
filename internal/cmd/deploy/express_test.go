package deploy_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/internal/aws/s3"
	"github.com/awesome-foundation/cfnrain/internal/cmd"
	"github.com/awesome-foundation/cfnrain/internal/cmd/deploy"
	"github.com/awesome-foundation/cfnrain/internal/console/spinner"
	"github.com/spf13/cobra"
)

func TestDeployExpressRequest(t *testing.T) {
	for _, express := range []bool{false, true} {
		for _, keep := range []bool{false, true} {
			t.Run(fmt.Sprintf("express=%v/keep=%v", express, keep), func(t *testing.T) {
				requests := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method == http.MethodHead {
						return // The artifact bucket exists.
					}
					if err := r.ParseForm(); err != nil {
						t.Error(err)
						return
					}
					w.Header().Set("Content-Type", "text/xml")
					switch action := r.Form.Get("Action"); action {
					case "GetCallerIdentity":
						fmt.Fprint(w, "<GetCallerIdentityResponse><GetCallerIdentityResult><Account>000000000000</Account></GetCallerIdentityResult></GetCallerIdentityResponse>")
					case "DescribeStacks":
						stackResponse(w, "CREATE_COMPLETE")
					case "ListStacks":
						fmt.Fprint(w, "<ListStacksResponse><ListStacksResult><StackSummaries/></ListStacksResult></ListStacksResponse>")
					case "CreateChangeSet":
						requests++
						wantMode, wantRollback := "", ""
						if express {
							wantMode, wantRollback = "EXPRESS", strconv.FormatBool(keep)
						}
						if got := r.Form.Get("DeploymentConfig.Mode"); got != wantMode {
							t.Errorf("DeploymentConfig.Mode = %q, want %q", got, wantMode)
						}
						if got := r.Form.Get("DeploymentConfig.DisableRollback"); got != wantRollback {
							t.Errorf("DeploymentConfig.DisableRollback = %q, want %q", got, wantRollback)
						}
						// Stop after the request under test, before execution or polling.
						w.WriteHeader(http.StatusBadRequest)
						fmt.Fprint(w, "<ErrorResponse><Error><Code>ValidationError</Code><Message>test stopped after CreateChangeSet</Message></Error></ErrorResponse>")
					default:
						t.Errorf("unexpected action %q", action)
						w.WriteHeader(http.StatusBadRequest)
					}
				}))
				defer server.Close()
				configureLookupAWS(t, server.URL)
				t.Setenv("AWS_ENDPOINT_URL_STS", server.URL)
				t.Setenv("AWS_ENDPOINT_URL_S3", server.URL)
				previousBucket := s3.BucketName
				s3.BucketName = "test-artifacts"
				t.Cleanup(func() { s3.BucketName = previousBucket })
				defer spinner.StopTimer()

				for _, name := range []string{"express", "keep", "yes", "changeset"} {
					flag := deploy.Cmd.Flags().Lookup(name)
					value, changed := flag.Value.String(), flag.Changed
					t.Cleanup(func() {
						_ = flag.Value.Set(value)
						flag.Changed = changed
					})
					_ = flag.Value.Set("false")
					flag.Changed = false
				}
				template := filepath.Join(t.TempDir(), "template.yaml")
				if err := os.WriteFile(template, []byte("Resources:\n  Bucket:\n    Type: AWS::S3::Bucket\n"), 0600); err != nil {
					t.Fatal(err)
				}
				deploy.Cmd.SetArgs([]string{template, "test-stack", "--yes", "--express=" + strconv.FormatBool(express), "--keep=" + strconv.FormatBool(keep)})
				defer deploy.Cmd.SetArgs(nil)
				if exitCode := cmd.Test(deploy.Cmd); exitCode != 1 || requests != 1 {
					t.Fatalf("exit code = %d, CreateChangeSet requests = %d; want 1 each", exitCode, requests)
				}
			})
		}
	}
}

// TestExpressChangeSetConflict must fail before a deploy handler can call AWS.
func TestExpressChangeSetConflict(t *testing.T) {
	oldRun := deploy.Cmd.Run
	defer func() {
		deploy.Cmd.Run = oldRun
		deploy.Cmd.SetArgs(nil)
		deploy.Cmd.SetErr(nil)
		deploy.Cmd.SetOut(nil)
	}()
	for _, name := range []string{"express", "changeset"} {
		flag := deploy.Cmd.Flags().Lookup(name)
		value, changed := flag.Value.String(), flag.Changed
		t.Cleanup(func() {
			_ = flag.Value.Set(value)
			flag.Changed = changed
		})
	}
	called := false
	deploy.Cmd.Run = func(*cobra.Command, []string) { called = true }
	var stderr, stdout bytes.Buffer
	deploy.Cmd.SetErr(&stderr)
	deploy.Cmd.SetOut(&stdout)
	deploy.Cmd.SetArgs([]string{"--express", "--changeset", "unused", "unused"})
	if code := cmd.Test(deploy.Cmd); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if called {
		t.Fatal("deploy handler ran despite conflicting flags")
	}
	if !strings.Contains(stderr.String(), "flags in the group [express changeset]") {
		t.Fatalf("missing flag-conflict error: %s", stderr.String())
	}
}
