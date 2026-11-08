package deploy

import (
	"fmt"
	"strings"

	"github.com/awesome-foundation/cfnrain/cft"
	"github.com/awesome-foundation/cfnrain/cft/pkg"
	"github.com/awesome-foundation/cfnrain/internal/aws/cfn"
	"github.com/awesome-foundation/cfnrain/internal/aws/s3"
	"github.com/awesome-foundation/cfnrain/internal/config"
	"github.com/awesome-foundation/cfnrain/internal/console"
	"github.com/awesome-foundation/cfnrain/internal/console/spinner"
	"github.com/awesome-foundation/cfnrain/internal/ui"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go/ptr"
)

func formatChangeSet(stackName, changeSetName string) string {
	status, err := cfn.GetChangeSet(stackName, changeSetName)
	if err != nil {
		panic(ui.Errorf(err, "error getting changeset '%s' for stack '%s'", changeSetName, stackName))
	}

	out := strings.Builder{}

	out.WriteString(fmt.Sprintf("%s:\n", console.Yellow(fmt.Sprintf("Stack %s", ptr.ToString(status.StackName)))))

	// Non-stack resources
	for _, change := range status.Changes {
		if change.ResourceChange.ChangeSetId != nil {
			// Bunch up nested stacks to the end
			continue
		}

		line := fmt.Sprintf("%s %s",
			*change.ResourceChange.ResourceType,
			*change.ResourceChange.LogicalResourceId,
		)

		switch change.ResourceChange.Action {
		case types.ChangeAction("Add"):
			out.WriteString(console.Green("  + " + line))
		case types.ChangeAction("Modify"):
			out.WriteString(console.Blue("  > " + line))
		case types.ChangeAction("Remove"):
			out.WriteString(console.Red("  - " + line))
		}

		out.WriteString("\n")
	}

	// Nested stacks
	for _, change := range status.Changes {
		if change.ResourceChange.ChangeSetId == nil {
			continue
		}

		child := formatChangeSet("", ptr.ToString(change.ResourceChange.ChangeSetId))
		parts := strings.SplitN(child, "\n", 2)
		header := parts[0]
		body := console.Grey("    (no changes in resources)\n")
		if len(parts) == 2 {
			body = parts[1]
		}

		switch change.ResourceChange.Action {
		case types.ChangeAction("Add"):
			out.WriteString(console.Green("  + " + header))
		case types.ChangeAction("Modify"):
			out.WriteString(console.Blue("  > " + header))
		case types.ChangeAction("Remove"):
			out.WriteString(console.Red("  - " + header))
		}
		out.WriteString("\n")

		out.WriteString(ui.Indent("  ", body))
		out.WriteString("\n")
	}

	return strings.TrimSpace(out.String())
}

func PackageTemplate(fn string, yes bool) *cft.Template {
	// Call RainBucket for side-effects in case we want to force bucket creation
	s3.RainBucket(yes)

	t, err := pkg.File(fn)
	if err != nil {
		panic(ui.Errorf(err, "error packaging template '%s'", fn))
	}

	return t
}

func CheckStack(stackName string, yes bool) (types.Stack, bool, error) {
	return checkStack(stackName, yes, cfn.GetStack, cfn.DeleteStack, cfn.WaitForStackToSettle, console.Confirm)
}

func checkStack(stackName string, yes bool,
	getStack func(string) (types.Stack, error),
	deleteStack func(string, string) error,
	waitForStack func(string) (string, []string),
	confirm func(bool, string) bool,
) (types.Stack, bool, error) {
	// Find out if stack exists already
	// If it does and it's not in a good state, offer to wait/delete
	stack, err := getStack(stackName)

	stackExists := false
	if err == nil {
		config.Debugf("Stack exists")
		stackExists = true
	}

	spinner.Pause()
	defer spinner.Resume()

	if stackExists {
		switch {
		case stack.StackStatus == types.StackStatusRollbackComplete,
			stack.StackStatus == types.StackStatusReviewInProgress,
			stack.StackStatus == types.StackStatusCreateFailed:

			if !yes && !confirm(false, fmt.Sprintf("Stack '%s' is in %s. Delete it and its change sets before deploying?", stackName, stack.StackStatus)) {
				return stack, true, fmt.Errorf("user cancelled deployment; stack '%s' was not deleted", stackName)
			}

			message := fmt.Sprintf("Deleting existing stack '%s' (%s).", stackName, stack.StackStatus)
			fmt.Println(message)

			stackID := ptr.ToString(stack.StackId)
			err := deleteStack(stackName, "")
			if err != nil {
				return stack, true, ui.Errorf(err, "unable to delete stack '%s'", stackName)
			}

			status, _ := waitForStack(stackID)

			if status != "DELETE_COMPLETE" {
				return stack, true, fmt.Errorf("failed to delete stack '%s'", stackName)
			}

			console.ClearLines(console.CountLines(message) + 1)
			fmt.Println("Deleted existing stack.")

			stackExists = false
		case !strings.HasSuffix(string(stack.StackStatus), "_COMPLETE"):
			// Can't update
			return stack, true, fmt.Errorf("stack '%s' could not be updated: %s", stackName, ui.ColouriseStatus(string(stack.StackStatus)))
		}
	}

	return stack, stackExists, nil
}
