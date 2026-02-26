package notionclient

import (
	"fmt"
	"regexp"
	"strings"
)

var pageIDPattern = regexp.MustCompile(`(?i)[0-9a-f]{32}`)

func ParsePageID(input string) (string, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", fmt.Errorf("page URL/ID is empty")
	}

	compact := strings.NewReplacer("-", "", "_", "").Replace(s)
	if m := pageIDPattern.FindString(compact); m != "" {
		return formatUUID(strings.ToLower(m)), nil
	}

	return "", fmt.Errorf("could not extract 32-char page id from input: %q", input)
}

func formatUUID(compact string) string {
	if len(compact) != 32 {
		return compact
	}
	return compact[0:8] + "-" +
		compact[8:12] + "-" +
		compact[12:16] + "-" +
		compact[16:20] + "-" +
		compact[20:32]
}
