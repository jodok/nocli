package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/jodok/nocli/internal/config"
)

var (
	reCookieHeader = regexp.MustCompile(`(?im)cookie\s*:\s*([^\r\n"']+)`)
	reTokenV2      = regexp.MustCompile(`(?i)token_v2=([^;\s"']+)`)
	reUserID       = regexp.MustCompile(`(?i)notion_user_id=([0-9a-f\-]{32,36})`)
	reActiveUser   = regexp.MustCompile(`(?im)x-notion-active-user-header\s*:\s*([0-9a-f\-]{32,36})`)
	reNotionURL    = regexp.MustCompile(`https://www\.notion\.so`)
)

type AuthImportCurlCmd struct {
	InputPath   string `name:"input" short:"i" help:"Path to a text file containing copied cURL"`
	StoreCookie bool   `name:"store-cookie" help:"Also store full Cookie header in config"`
}

func (c *AuthImportCurlCmd) Run(ctx context.Context) error {
	raw, err := c.readInput()
	if err != nil {
		return err
	}

	token, userID, activeUserID, cookieHeader := parseCurlAuth(raw)
	if token == "" && userID == "" && activeUserID == "" && cookieHeader == "" {
		return fmt.Errorf("no notion auth values found; paste a full 'Copy as cURL' request including headers")
	}

	path := ConfigPathFromContext(ctx)
	cfg, err := config.Read(path)
	if err != nil {
		return err
	}

	if token != "" {
		cfg.TokenV2 = token
	}
	if userID != "" {
		cfg.NotionUserID = userID
	}
	if activeUserID != "" {
		cfg.ActiveUserID = activeUserID
	}
	if c.StoreCookie && cookieHeader != "" {
		cfg.Cookie = cookieHeader
	}
	if cfg.BaseURL == "" && reNotionURL.MatchString(raw) {
		cfg.BaseURL = "https://www.notion.so"
	}

	if err := config.Write(path, cfg); err != nil {
		return err
	}

	printImportSummary(path, token, userID, activeUserID, c.StoreCookie && cookieHeader != "")
	if token == "" || userID == "" {
		fmt.Printf("warning: token_v2 or notion_user_id missing from pasted data\n")
	}
	return nil
}

func (c *AuthImportCurlCmd) readInput() (string, error) {
	if strings.TrimSpace(c.InputPath) != "" {
		b, err := os.ReadFile(c.InputPath)
		if err != nil {
			return "", fmt.Errorf("read input file: %w", err)
		}
		return string(b), nil
	}

	st, err := os.Stdin.Stat()
	if err != nil {
		return "", fmt.Errorf("stat stdin: %w", err)
	}
	if (st.Mode() & os.ModeCharDevice) != 0 {
		fmt.Fprintln(os.Stderr, "Paste Notion 'Copy as cURL' output, then press Ctrl-D:")
	}

	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}
	if strings.TrimSpace(string(b)) == "" {
		return "", fmt.Errorf("empty input")
	}
	return string(b), nil
}

func parseCurlAuth(raw string) (token string, userID string, activeUserID string, cookieHeader string) {
	if m := reCookieHeader.FindStringSubmatch(raw); len(m) > 1 {
		cookieHeader = trimSpace(m[1])
	}

	if cookieHeader != "" {
		token = parseCookieValue(cookieHeader, "token_v2")
		userID = parseCookieValue(cookieHeader, "notion_user_id")
	}

	if token == "" {
		if m := reTokenV2.FindStringSubmatch(raw); len(m) > 1 {
			token = trimSpace(m[1])
		}
	}
	if userID == "" {
		if m := reUserID.FindStringSubmatch(raw); len(m) > 1 {
			userID = trimSpace(m[1])
		}
	}
	if m := reActiveUser.FindStringSubmatch(raw); len(m) > 1 {
		activeUserID = trimSpace(m[1])
	}

	return token, userID, activeUserID, cookieHeader
}
