package rm_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/awesome-foundation/cfnrain/internal/aws"
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
	configureAWS(t, server.URL)
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

func configureAWS(t *testing.T, endpoint string) {
	t.Helper()
	t.Setenv("AWS_ENDPOINT_URL_CLOUDFORMATION", endpoint)
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	t.Setenv("AWS_CONFIG_FILE", t.TempDir()+"/config")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", t.TempDir()+"/credentials")
	t.Setenv("AWS_PROFILE", "")

	// Force a reload because the AWS configuration is cached between tests.
	aws.NamedConfig(t.Name())
}

func TestDeleteCompletesBeforeFirstPoll(t *testing.T) {
	const stackID = "arn:aws:cloudformation:us-east-1:000000000000:stack/test-stack/test-id"
	deleted := false
	polls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Error(err)
			return
		}
		w.Header().Set("Content-Type", "text/xml")
		switch action := r.Form.Get("Action"); action {
		case "DescribeStacks":
			status := "REVIEW_IN_PROGRESS"
			if deleted {
				polls++
				if r.Form.Get("StackName") != stackID {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprint(w, `<ErrorResponse><Error><Code>ValidationError</Code><Message>Stack with id test-stack does not exist</Message></Error></ErrorResponse>`)
					return
				}
				status = "DELETE_COMPLETE"
			}
			fmt.Fprintf(w, `<DescribeStacksResponse><DescribeStacksResult><Stacks><member><StackName>test-stack</StackName><StackId>%s</StackId><StackStatus>%s</StackStatus></member></Stacks></DescribeStacksResult></DescribeStacksResponse>`, stackID, status)
		case "GetTemplate":
			fmt.Fprint(w, `<GetTemplateResponse><GetTemplateResult><TemplateBody>{"Resources":{}}</TemplateBody></GetTemplateResult></GetTemplateResponse>`)
		case "DeleteStack":
			deleted = true
			fmt.Fprint(w, `<DeleteStackResponse/>`)
		case "DescribeChangeSet", "DescribeStackResources":
			// Status rendering may query resources by name after the stack is gone.
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `<ErrorResponse><Error><Code>ValidationError</Code><Message>Stack with id test-stack does not exist</Message></Error></ErrorResponse>`)
		default:
			t.Errorf("unexpected CloudFormation action %q", action)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()
	configureAWS(t, server.URL)
	if err := rm.Cmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rm.Cmd.Flags().Set("yes", "false") })
	rm.Cmd.Run(rm.Cmd, []string{"test-stack"})
	if !deleted || polls != 1 {
		t.Fatalf("deleted = %v, polls = %d; want a completed delete with one poll", deleted, polls)
	}
}
