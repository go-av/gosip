package utils

import "strings"

const RFC3261BranchMagicCookie = "z9hG4bK"

func GenerateBranchID() string {
	return strings.Join([]string{
		RFC3261BranchMagicCookie,
		RandString(32),
	}, "-")
}
