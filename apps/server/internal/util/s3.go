package util

import "regexp"

var unsafeS3KeyChars = regexp.MustCompile(`[^a-zA-Z0-9\-.]`)

// SanitizeS3Key replaces characters that are awkward in S3 keys with underscores.
func SanitizeS3Key(s string) string {
	return unsafeS3KeyChars.ReplaceAllString(s, "_")
}
