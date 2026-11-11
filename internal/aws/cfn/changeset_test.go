package cfn

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go/ptr"
)

func TestNewChangeSetInputExpress(t *testing.T) {
	ctx := &ChangeSetContext{
		StackName:     "app",
		IncludeNested: true,
		RoleArn:       "arn:aws:iam::123456789012:role/deploy",
	}

	want := &cloudformation.CreateChangeSetInput{
		ChangeSetType:       types.ChangeSetTypeUpdate,
		ChangeSetName:       ptr.String("app-1"),
		StackName:           ptr.String("app"),
		Tags:                []types.Tag{},
		IncludeNestedStacks: ptr.Bool(true),
		Capabilities:        []types.Capability{"CAPABILITY_NAMED_IAM", "CAPABILITY_AUTO_EXPAND"},
		RoleARN:             ptr.String("arn:aws:iam::123456789012:role/deploy"),
		TemplateBody:        ptr.String("Resources: {}"),
	}

	// Without --express the request has no DeploymentConfig, as before, with or without --keep
	for _, disableRollback := range []bool{false, true} {
		ctx.DisableRollback = disableRollback
		got := newChangeSetInput(ctx, "UPDATE", "app-1", "Resources: {}")
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("without Express, DisableRollback=%v:\n got %+v\nwant %+v", disableRollback, got, want)
		}
	}

	var got *cloudformation.CreateChangeSetInput

	for _, disableRollback := range []bool{false, true} {
		ctx.Express = true
		ctx.DisableRollback = disableRollback
		got = newChangeSetInput(ctx, "UPDATE", "app-1", "Resources: {}")
		if got.DeploymentConfig == nil {
			t.Fatalf("with Express: DeploymentConfig is nil")
		}
		if got.DeploymentConfig.Mode != types.DeploymentConfigModeExpress {
			t.Errorf("with Express: Mode = %q, want EXPRESS", got.DeploymentConfig.Mode)
		}
		// Express mode disables rollback unless told otherwise, so the value is always sent
		if got.DeploymentConfig.DisableRollback == nil || *got.DeploymentConfig.DisableRollback != disableRollback {
			t.Errorf("with Express: DisableRollback = %v, want %v", got.DeploymentConfig.DisableRollback, disableRollback)
		}
		got.DeploymentConfig = nil
		if !reflect.DeepEqual(got, want) {
			t.Errorf("with Express: other fields changed:\n got %+v\nwant %+v", got, want)
		}
	}
}
