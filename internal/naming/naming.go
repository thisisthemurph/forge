package naming

import (
	"strconv"
	"strings"
	"unicode"
)

const maxSlugLen = 40

// FeatureBranch returns the deterministic **Feature branch** name for a Feature issue.
// titleSlug must be the output of SlugFromTitle when non-empty, or "" to omit a slug suffix.
func FeatureBranch(feature int, titleSlug string) string {
	base := "forge/feature/" + itoa(feature) + "/base"
	if titleSlug == "" {
		return base
	}
	return base + "/" + titleSlug
}

// StackedBranch returns the deterministic **Stacked branch** name for a Sub-issue under a Feature.
// titleSlug must be the output of SlugFromTitle when non-empty, or "" to omit a slug suffix.
//
// The **Feature branch** lives under forge/feature/<N>/base so git can also hold
// forge/feature/<N>/issue/<M> without ref-name collisions (refs/heads/foo cannot coexist with refs/heads/foo/...).
func StackedBranch(feature, subIssue int, titleSlug string) string {
	base := "forge/feature/" + itoa(feature) + "/issue/" + itoa(subIssue)
	if titleSlug == "" {
		return base
	}
	return base + "/" + titleSlug
}

// SlugFromTitle derives an optional stable slug from issue text for branch name suffixes.
// Empty or non-alphanumeric titles yield an empty slug.
func SlugFromTitle(title string) string {
	var b strings.Builder
	var prevDash bool
	for _, r := range strings.TrimSpace(title) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r - 'A' + 'a')
			prevDash = false
		default:
			if unicode.IsLetter(r) && r < 128 {
				// Skip non-ASCII letters without transliteration.
				continue
			}
			if b.Len() == 0 || prevDash {
				continue
			}
			b.WriteByte('-')
			prevDash = true
		}
	}
	s := strings.Trim(b.String(), "-")
	if len(s) > maxSlugLen {
		s = s[:maxSlugLen]
	}
	return strings.Trim(s, "-")
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
