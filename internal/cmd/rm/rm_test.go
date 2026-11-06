package rm_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/awesome-foundation/cfnrain/internal/cmd/rm"
)

func TestMissingTerminationProtection(t *testing.T) {
	var actions []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Error(err)
			return
		}
		action := r.Form.Get("Action")
		actions = append(actions, action)
		w.Header().Set("Content-Type", "text/xml")
		switch action {
		case "DescribeStacks":
			fmt.Fprint(w, `<DescribeStacksResponse><DescribeStacksResult><Stacks><member><StackName>test-stack</StackName><StackStatus>CREATE_COMPLETE</StackStatus></member></Stacks></DescribeStacksResult></DescribeStacksResponse>`)
		case "GetTemplate":
			fmt.Fprint(w, `<GetTemplateResponse><GetTemplateResult><TemplateBody>{"Resources":{}}</TemplateBody></GetTemplateResult></GetTemplateResponse>`)
		case "DeleteStack":
			fmt.Fprint(w, `<DeleteStackResponse/>`)
		default:
			t.Errorf("unexpected CloudFormation action %q", action)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()
	t.Setenv("AWS_ENDPOINT_URL_CLOUDFORMATION", server.URL)
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	t.Setenv("AWS_CONFIG_FILE", t.TempDir()+"/config")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", t.TempDir()+"/credentials")
	t.Setenv("AWS_PROFILE", "")
	for _, flag := range []string{"yes", "detach"} {
		if err := rm.Cmd.Flags().Set(flag, "true"); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = rm.Cmd.Flags().Set(flag, "false") })
	}
	rm.Cmd.Run(rm.Cmd, []string{"test-stack"})
	if want := []string{"DescribeStacks", "GetTemplate", "DeleteStack"}; !reflect.DeepEqual(actions, want) {
		t.Fatalf("actions = %v, want %v", actions, want)
	}
}
