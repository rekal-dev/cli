package session

import "regexp"

var nonAlphanumeric = regexp.MustCompile(`[^a-zA-Z0-9]`)

// SanitizeRepoPath replicates Claude Code's path sanitization:
// non-alphanumeric characters are replaced with dashes.
// e.g. /Users/frank/projects/rekal → -Users-frank-projects-rekal
func SanitizeRepoPath(repoPath string) string {
	return nonAlphanumeric.ReplaceAllString(repoPath, "-")
}
