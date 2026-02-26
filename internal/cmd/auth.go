package cmd

import (
	"fmt"
)

type AuthCmd struct {
	ImportCurl AuthImportCurlCmd `cmd:"" name:"import-curl" help:"Import auth from a pasted Notion DevTools 'Copy as cURL' request"`
}

func redact(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return "********"
	}
	return s[:4] + "..." + s[len(s)-4:]
}

func parseCookieValue(cookieHeader string, key string) string {
	for _, p := range splitCookieParts(cookieHeader) {
		k, v, ok := splitKV(p)
		if !ok {
			continue
		}
		if k == key {
			return v
		}
	}
	return ""
}

func splitCookieParts(cookie string) []string {
	parts := make([]string, 0)
	cur := ""
	for _, r := range cookie {
		if r == ';' {
			parts = append(parts, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	parts = append(parts, cur)
	return parts
}

func splitKV(s string) (string, string, bool) {
	for i, r := range s {
		if r == '=' {
			k := trimSpace(s[:i])
			v := trimSpace(s[i+1:])
			if k == "" {
				return "", "", false
			}
			return k, v, true
		}
	}
	return "", "", false
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func printImportSummary(path string, token string, userID string, activeUserID string, cookieStored bool) {
	fmt.Printf("updated config: %s\n", path)
	if token != "" {
		fmt.Printf("- token_v2: %s\n", redact(token))
	}
	if userID != "" {
		fmt.Printf("- notion_user_id: %s\n", userID)
	}
	if activeUserID != "" {
		fmt.Printf("- active_user_id: %s\n", activeUserID)
	}
	if cookieStored {
		fmt.Printf("- cookie: stored\n")
	}
}
