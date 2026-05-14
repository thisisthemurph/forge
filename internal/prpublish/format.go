package prpublish

import (
	"fmt"
	"strings"
)

// ForgeManagedPRTitle returns the **PR title (v1)** for a **Forge-managed PR**:
// `[#<sub-issue>] <issue title>` per CONTEXT.md.
func ForgeManagedPRTitle(subIssue int, issueTitle string) string {
	return fmt.Sprintf("[#%d] %s", subIssue, strings.TrimSpace(issueTitle))
}

// ForgeManagedPRBody returns **PR body linking (v1)** with Fixes #<sub-issue>
// so GitHub can close the Sub-issue when the PR merges.
func ForgeManagedPRBody(subIssue int) string {
	return fmt.Sprintf("Fixes #%d\n", subIssue)
}
