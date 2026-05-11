package scheduling

import (
	"regexp"
	"strconv"
	"strings"
)

var issueRefRE = regexp.MustCompile(`#(\d+)`)

// ParseBlockedBySection extracts Blocker issue numbers from the ## Blocked by
// section of a Sub-issue body. Missing or empty section yields nil.
// The section runs until the next markdown heading (line starting with "## ").
func ParseBlockedBySection(body string) []int {
	lines := strings.Split(body, "\n")
	var inSection bool
	var collected []int
	seen := map[int]struct{}{}
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if !inSection {
			if trim == "## Blocked by" {
				inSection = true
			}
			continue
		}
		if strings.HasPrefix(trim, "## ") && trim != "## Blocked by" {
			break
		}
		for _, m := range issueRefRE.FindAllStringSubmatch(trim, -1) {
			n, err := strconv.Atoi(m[1])
			if err != nil {
				continue
			}
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			collected = append(collected, n)
		}
	}
	if len(collected) == 0 {
		return nil
	}
	return collected
}
