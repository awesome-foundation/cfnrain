package cfn_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/internal/aws/cfn"
)

func TestGetStackEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Error(err)
			return
		}
		if r.Form.Get("Action") != "DescribeStacks" || r.Form.Get("StackName") != "test-stack" {
			t.Errorf("unexpected request: %v", r.Form)
		}
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprint(w, `<DescribeStacksResponse><DescribeStacksResult><Stacks/></DescribeStacksResult></DescribeStacksResponse>`)
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
	_, err := cfn.GetStack("test-stack")
	if err == nil || !strings.Contains(err.Error(), "test-stack") || !strings.Contains(err.Error(), "no stacks") {
		t.Fatalf("GetStack error = %v, want an empty-response error naming test-stack", err)
	}
}

func TestGetTypeSchemaOnlyCacheMissingType(t *testing.T) {
	schema, err := cfn.GetTypeSchema("AWS::Fake::Thing", cfn.OnlyUseCache)
	if schema != "" {
		t.Errorf("GetTypeSchema returned schema %q for a missing type", schema)
	}
	want := "resource type AWS::Fake::Thing is not in the schemas built into cfnrain"
	if err == nil || err.Error() != want {
		t.Fatalf("GetTypeSchema error = %v, want %q", err, want)
	}
}
