package pkg

import (
	"strings"

	"github.com/awesome-foundation/cfnrain/cft/parse"
)

// isSubNeeded keeps escaped variables for CloudFormation to resolve as literals.
func isSubNeeded(s string) bool {
	return parse.IsSubNeeded(s) || strings.Contains(s, "${!")
}
