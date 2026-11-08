package deploy_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/internal/aws"
	"github.com/awesome-foundation/cfnrain/internal/cmd/deploy"
	"github.com/awesome-foundation/cfnrain/internal/console/spinner"
)

const testStackID = "arn:aws:cloudformation:us-east-1:000000000000:stack/test-stack/test-id"

func configureLookupAWS(t *testing.T, endpoint string) {
	t.Helper()
	t.Setenv("AWS_ENDPOINT_URL_CLOUDFORMATION", endpoint)
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	t.Setenv("AWS_CONFIG_FILE", t.TempDir()+"/config")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", t.TempDir()+"/credentials")
	t.Setenv("AWS_PROFILE", "")
	aws.NamedConfig(t.Name())
}

func stackResponse(w http.ResponseWriter, status string) {
	fmt.Fprintf(w, "<DescribeStacksResponse><DescribeStacksResult><Stacks><member><StackName>test-stack</StackName><StackId>%s</StackId><StackStatus>%s</StackStatus></member></Stacks></DescribeStacksResult></DescribeStacksResponse>", testStackID, status)
}

func lookupFailure(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, "<ErrorResponse><Error><Code>ValidationError</Code><Message>Stack with id test-stack does not exist</Message></Error></ErrorResponse>")
}

func TestDeployDeleteCompletesBeforeFirstPoll(t *testing.T) {
	for _, status := range []string{"ROLLBACK_COMPLETE", "REVIEW_IN_PROGRESS", "CREATE_FAILED"} {
		for _, nameResolves := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/nameResolves=%v", status, nameResolves), func(t *testing.T) {
				deleted, polls := false, 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if err := r.ParseForm(); err != nil {
						t.Error(err)
						return
					}
					w.Header().Set("Content-Type", "text/xml")
					switch action := r.Form.Get("Action"); action {
					case "DescribeStacks":
						if !deleted {
							stackResponse(w, status)
							return
						}
						polls++
						if !nameResolves && r.Form.Get("StackName") != testStackID {
							lookupFailure(w)
							return
						}
						stackResponse(w, "DELETE_COMPLETE")
					case "DeleteStack":
						deleted = true
						fmt.Fprint(w, "<DeleteStackResponse/>")
					case "DescribeChangeSet", "DescribeStackResources":
						lookupFailure(w)
					default:
						t.Errorf("unexpected action %q", action)
						w.WriteHeader(http.StatusBadRequest)
					}
				}))
				defer server.Close()
				configureLookupAWS(t, server.URL)
				defer spinner.StopTimer()
				defer func() {
					if failure := recover(); failure != nil {
						t.Errorf("delete wait failed: %v", failure)
					}
				}()
				_, exists, err := deploy.CheckStack("test-stack", true)
				if err != nil || exists || !deleted || polls != 1 {
					t.Fatalf("exists=%v, err=%v, deleted=%v, polls=%d", exists, err, deleted, polls)
				}
			})
		}
	}
}

func TestDeployFinalStackLookup(t *testing.T) {
	for _, result := range []string{"success", "empty", "error"} {
		t.Run(result, func(t *testing.T) {
			lookups := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseForm(); err != nil {
					t.Error(err)
					return
				}
				w.Header().Set("Content-Type", "text/xml")
				switch action := r.Form.Get("Action"); action {
				case "ExecuteChangeSet":
					fmt.Fprint(w, "<ExecuteChangeSetResponse><ExecuteChangeSetResult/></ExecuteChangeSetResponse>")
				case "DescribeStacks":
					lookups++
					if lookups == 1 || result == "success" {
						stackResponse(w, "CREATE_COMPLETE")
					} else if result == "empty" {
						fmt.Fprint(w, "<DescribeStacksResponse><DescribeStacksResult><Stacks/></DescribeStacksResult></DescribeStacksResponse>")
					} else {
						lookupFailure(w)
					}
				case "DescribeChangeSet":
					lookupFailure(w)
				case "DescribeStackResources":
					fmt.Fprint(w, "<DescribeStackResourcesResponse><DescribeStackResourcesResult><StackResources/></DescribeStackResourcesResult></DescribeStackResourcesResponse>")
				default:
					t.Errorf("unexpected action %q", action)
					w.WriteHeader(http.StatusBadRequest)
				}
			}))
			defer server.Close()
			configureLookupAWS(t, server.URL)
			if err := deploy.Cmd.Flags().Set("changeset", "true"); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = deploy.Cmd.Flags().Set("changeset", "false") })

			out, err := os.CreateTemp(t.TempDir(), "stdout")
			if err != nil {
				t.Fatal(err)
			}
			defer out.Close()
			previous := os.Stdout
			var failure error
			func() {
				os.Stdout = out
				defer func() {
					os.Stdout = previous
					if recovered := recover(); recovered != nil {
						t.Fatalf("unexpected panic: %v", recovered)
					}
				}()
				failure = deploy.Cmd.RunE(deploy.Cmd, []string{"test-stack", "test-change-set"})
			}()
			if _, err := out.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			output, err := io.ReadAll(out)
			if err != nil {
				t.Fatal(err)
			}
			if lookups != 2 {
				t.Fatalf("got %d lookups, want wait then final lookup; failure=%v output=%s", lookups, failure, output)
			}
			if result == "success" {
				if failure != nil || !strings.Contains(string(output), "Successfully deployed test-stack") {
					t.Fatalf("failure=%v, output=%s", failure, output)
				}
				return
			}
			wantCause := "no stacks"
			if result == "error" {
				wantCause = "does not exist"
			}
			if failure == nil || !strings.Contains(fmt.Sprint(failure), "test-stack") || !strings.Contains(fmt.Sprint(failure), wantCause) {
				t.Errorf("expected contextual lookup failure (%s), got %v", wantCause, failure)
			}
			if strings.Contains(string(output), "Successfully") || strings.Contains(string(output), "Stack ") {
				t.Errorf("printed summary or success after failed lookup: %s", output)
			}
		})
	}
}
