package cfn_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/parse"
	"github.com/awesome-foundation/cfnrain/internal/aws/cfn"
)

func TestCreateChangeSetValidationDetails(t *testing.T) {
	const original = "The following hook(s)/validation failed: [AWS::EarlyValidation::ResourceExistenceCheck]. To troubleshoot Early Validation errors, use the DescribeEvents API for detailed failure information."
	const noChanges = "The submitted information didn't contain changes. Submit different information to create a change set."
	const event = `<member><EventType>VALIDATION_ERROR</EventType><LogicalResourceId>TestBucket</LogicalResourceId><ResourceType>AWS::S3::Bucket</ResourceType><ValidationName>NAME_CONFLICT_VALIDATION</ValidationName><ValidationStatus>FAILED</ValidationStatus><ValidationFailureMode>FAIL</ValidationFailureMode><ValidationStatusReason>Resource already exists.</ValidationStatusReason></member>`
	const detail = "changeset creation failed with validation errors:\n  - Resource: TestBucket (AWS::S3::Bucket), Validation: NAME_CONFLICT_VALIDATION, Reason: Resource already exists."
	for _, tc := range []struct {
		name, reason, first, second, want string
		denied, missingID                 bool
		calls                             int
	}{
		{name: "validation", reason: original, first: event, want: detail, calls: 1},
		{name: "pagination", reason: original, second: event, want: detail, calls: 2},
		{name: "permission denied", reason: original, denied: true, want: original, calls: 1},
		{name: "second page denied", reason: original, first: event, second: "denied", denied: true, want: original, calls: 2},
		{name: "no changes", reason: noChanges, want: noChanges, calls: 1},
		{name: "no validation events", reason: original, want: original, calls: 1},
		{name: "missing ID", reason: original, missingID: true, want: original},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseForm(); err != nil {
					t.Error(err)
					return
				}
				w.Header().Set("Content-Type", "text/xml")
				switch action := r.Form.Get("Action"); action {
				case "ListStacks":
					fmt.Fprint(w, `<ListStacksResponse><ListStacksResult><StackSummaries/></ListStacksResult></ListStacksResponse>`)
				case "CreateChangeSet":
					fmt.Fprint(w, `<CreateChangeSetResponse><CreateChangeSetResult><Id>test-change-set-id</Id></CreateChangeSetResult></CreateChangeSetResponse>`)
				case "DescribeChangeSet":
					id := "<ChangeSetId>test-change-set-id</ChangeSetId>"
					if tc.missingID {
						id = ""
					}
					fmt.Fprintf(w, `<DescribeChangeSetResponse><DescribeChangeSetResult>%s<Status>FAILED</Status><StatusReason>%s</StatusReason></DescribeChangeSetResult></DescribeChangeSetResponse>`, id, tc.reason)
				case "DescribeEvents":
					calls++
					if got := r.Form.Get("ChangeSetName"); got != "test-change-set-id" {
						t.Errorf("ChangeSetName = %q", got)
					}
					token := ""
					events := tc.first
					if calls == 1 && tc.second != "" {
						token = "<NextToken>page-two</NextToken>"
					}
					if calls == 2 {
						if got := r.Form.Get("NextToken"); got != "page-two" {
							t.Errorf("NextToken = %q", got)
						}
						events = tc.second
					}
					if calls > 2 {
						t.Error("unexpected extra page")
						w.WriteHeader(400)
						return
					}
					if tc.denied && (tc.second == "" || calls == 2) {
						w.WriteHeader(http.StatusForbidden)
						fmt.Fprint(w, `<ErrorResponse><Error><Type>Sender</Type><Code>AccessDenied</Code><Message>Denied</Message></Error></ErrorResponse>`)
						return
					}
					fmt.Fprintf(w, `<DescribeEventsResponse><DescribeEventsResult><OperationEvents>%s</OperationEvents>%s</DescribeEventsResult></DescribeEventsResponse>`, events, token)
				default:
					t.Errorf("unexpected action %q", action)
					w.WriteHeader(400)
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
			template, err := parse.String("Resources:\n  TestBucket:\n    Type: AWS::S3::Bucket\n")
			if err != nil {
				t.Fatal(err)
			}
			name, err := cfn.CreateChangeSet(&cfn.ChangeSetContext{Template: template, StackName: "test-stack", ChangeSetName: "test-change-set"})
			t.Logf("CreateChangeSet error:\n%v", err)
			if name != "test-change-set" || err == nil || err.Error() != tc.want {
				t.Errorf("CreateChangeSet = (%q, %v), want (%q, %q)", name, err, "test-change-set", tc.want)
			}
			if calls != tc.calls {
				t.Errorf("DescribeEvents calls = %d, want %d", calls, tc.calls)
			}
		})
	}
}
