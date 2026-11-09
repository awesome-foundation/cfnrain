package deploy

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/internal/console"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go/ptr"
)

func TestCheckStackDeletionConfirmation(t *testing.T) {
	for _, status := range []types.StackStatus{
		types.StackStatusRollbackComplete,
		types.StackStatusCreateFailed,
		types.StackStatusReviewInProgress,
	} {
		for _, tc := range []struct {
			name        string
			yes, accept bool
			want        []string
		}{
			{"accepted", false, true, []string{"get", "confirm", "delete", "wait"}},
			{"declined", false, false, []string{"get", "confirm"}},
			{"yes", true, false, []string{"get", "delete", "wait"}},
		} {
			t.Run(string(status)+"/"+tc.name, func(t *testing.T) {
				var calls []string
				get := func(name string) (types.Stack, error) {
					calls = append(calls, "get")
					if name != "test-stack" {
						t.Fatalf("unexpected stack %q", name)
					}
					return types.Stack{StackStatus: status, StackId: ptr.String("test-stack-id")}, nil
				}
				confirm := func(defaultYes bool, prompt string) bool {
					calls = append(calls, "confirm")
					if defaultYes || !strings.Contains(prompt, "test-stack") || !strings.Contains(prompt, string(status)) || !strings.Contains(prompt, "change sets") {
						t.Fatalf("unsafe confirmation: defaultYes=%v, prompt=%q", defaultYes, prompt)
					}
					return tc.accept
				}
				delete := func(name, role string) error {
					calls = append(calls, "delete")
					if name != "test-stack" || role != "" {
						t.Fatalf("unexpected delete: %q, %q", name, role)
					}
					return nil
				}
				wait := func(name string) (string, []string) {
					calls = append(calls, "wait")
					if name != "test-stack-id" {
						t.Fatalf("unexpected wait: %q", name)
					}
					return "DELETE_COMPLETE", nil
				}
				_, exists, err := checkStack("test-stack", tc.yes, get, delete, wait, confirm)
				declined := !tc.yes && !tc.accept
				if exists != declined || (err != nil) != declined {
					t.Fatalf("exists=%v, err=%v; declined=%v", exists, err, declined)
				}
				if declined && !strings.Contains(err.Error(), "user cancelled deployment") {
					t.Fatalf("unclear error: %v", err)
				}
				if !reflect.DeepEqual(calls, tc.want) {
					t.Fatalf("calls=%v, want %v", calls, tc.want)
				}
			})
		}
	}
}

func TestCheckStackWithoutTerminalDoesNotDelete(t *testing.T) {
	oldTTY := console.IsTTY
	console.IsTTY = false
	t.Cleanup(func() { console.IsTTY = oldTTY })
	defer func() {
		if failure := recover(); failure == nil || !strings.Contains(fmt.Sprint(failure), "no interactive terminal") {
			t.Errorf("expected existing terminal error, got %v", failure)
		}
	}()
	checkStack("test-stack", false,
		func(string) (types.Stack, error) {
			return types.Stack{StackStatus: types.StackStatusReviewInProgress}, nil
		},
		func(string, string) error { t.Fatal("deleted without confirmation"); return nil },
		func(string) (string, []string) { t.Fatal("waited without deletion"); return "", nil },
		console.Confirm)
}

func TestCheckStackMissingID(t *testing.T) {
	_, exists, err := checkStack("test-stack", true,
		func(string) (types.Stack, error) {
			return types.Stack{StackStatus: types.StackStatusCreateFailed}, nil
		},
		func(string, string) error { return nil },
		func(name string) (string, []string) {
			if name != "test-stack" {
				t.Fatalf("wait stack = %q, want test-stack", name)
			}
			return "DELETE_COMPLETE", nil
		},
		func(bool, string) bool { t.Fatal("unexpected confirmation"); return false })
	if err != nil || exists {
		t.Fatalf("exists=%v, err=%v", exists, err)
	}
}
