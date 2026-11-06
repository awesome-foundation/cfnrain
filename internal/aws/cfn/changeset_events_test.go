package cfn

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go/ptr"
)

func TestValidationErrorFromEvents(t *testing.T) {
	events := []types.OperationEvent{
		{EventType: types.EventTypeStackEvent, LogicalResourceId: ptr.String("Stack")},
		{
			EventType:         types.EventTypeValidationError,
			LogicalResourceId: ptr.String("Skipped"),
			ValidationStatus:  types.ValidationStatusSkipped,
		},
		{
			EventType:             types.EventTypeValidationError,
			LogicalResourceId:     ptr.String("Warned"),
			ValidationStatus:      types.ValidationStatusFailed,
			ValidationFailureMode: types.HookFailureModeWarn,
		},
		{
			EventType:              types.EventTypeValidationError,
			LogicalResourceId:      ptr.String("TestBucket2"),
			ValidationFailureMode:  types.HookFailureModeFail,
			ResourceType:           ptr.String("AWS::S3::Bucket"),
			ValidationName:         ptr.String("BUCKET_EMPTINESS_VALIDATION"),
			ValidationStatus:       types.ValidationStatusFailed,
			ValidationStatusReason: ptr.String("The bucket is not empty."),
		},
		{
			EventType:              types.EventTypeValidationError,
			LogicalResourceId:      ptr.String("TestBucket1"),
			ResourceType:           ptr.String("AWS::S3::Bucket"),
			ValidationName:         ptr.String("NAME_CONFLICT_VALIDATION"),
			ValidationStatusReason: ptr.String("Resource already exists."), // status and mode left empty on purpose
		},
	}

	err := validationErrorFromEvents(events)
	if err == nil {
		t.Fatal("expected an error for validation events, got nil")
	}

	want := "changeset creation failed with validation errors:\n" +
		"  - Resource: TestBucket2 (AWS::S3::Bucket), Validation: BUCKET_EMPTINESS_VALIDATION, Reason: The bucket is not empty.\n" +
		"  - Resource: TestBucket1 (AWS::S3::Bucket), Validation: NAME_CONFLICT_VALIDATION, Reason: Resource already exists."
	if err.Error() != want {
		t.Errorf("got:\n%s\nwant:\n%s", err.Error(), want)
	}
}

// Failures that are not Early Validation failures must keep their original message.
func TestValidationErrorFromEventsNone(t *testing.T) {
	cases := map[string][]types.OperationEvent{
		"nil":    nil,
		"empty":  {},
		"others": {{EventType: types.EventTypeStackEvent}, {EventType: types.EventTypeProvisioningError}},
		"nonblocking": {
			{EventType: types.EventTypeValidationError, ValidationStatus: types.ValidationStatusSkipped},
			{EventType: types.EventTypeValidationError, ValidationStatus: types.ValidationStatusFailed, ValidationFailureMode: types.HookFailureModeWarn},
		},
	}
	for name, events := range cases {
		if err := validationErrorFromEvents(events); err != nil {
			t.Errorf("%s: expected nil, got %v", name, err)
		}
	}
}
