package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/jodok/nocli/internal/config"
	"github.com/jodok/nocli/internal/notionclient"
)

type RootFlags struct {
	ConfigPath   string `name:"config" default:".notion.json" help:"Path to config file" env:"NOTION_CONFIG"`
	BaseURL      string `name:"base-url" default:"https://www.notion.so" help:"Notion base URL" env:"NOTION_BASE_URL"`
	TokenV2      string `name:"token-v2" help:"Notion token_v2 cookie value" env:"NOTION_TOKEN_V2"`
	NotionUserID string `name:"notion-user-id" help:"notion_user_id cookie value" env:"NOTION_USER_ID"`
	ActiveUserID string `name:"active-user-id" help:"x-notion-active-user-header value" env:"NOTION_ACTIVE_USER_ID"`
	Cookie       string `name:"cookie" help:"Raw Cookie header (overrides token_v2/notion_user_id)" env:"NOTION_COOKIE"`
}

type CLI struct {
	RootFlags  `embed:""`
	Page       PageCmd       `cmd:"" help:"Page operations"`
	Block      BlockCmd      `cmd:"" help:"Block operations"`
	Collection CollectionCmd `cmd:"" help:"Collection operations"`
	Auth       AuthCmd       `cmd:"" help:"Authentication helpers"`
}

func Execute(args []string) error {
	cli := &CLI{}
	parser, err := kong.New(cli,
		kong.Name("nocli"),
		kong.Description("CLI for Notion browser/private endpoints"),
		kong.UsageOnError(),
	)
	if err != nil {
		return fmt.Errorf("create parser: %w", err)
	}
	if len(args) == 0 {
		printTopLevelHelp()
		if !hasAnyAuthMaterial() {
			printAuthBootstrapHint()
		}
		return nil
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		return err
	}

	cfg, err := config.Read(config.ResolvePath(cli.ConfigPath))
	if err != nil {
		return err
	}

	baseURL := firstNonEmpty(cli.BaseURL, cfg.BaseURL)
	tokenV2 := firstNonEmpty(strings.TrimSpace(cli.TokenV2), cfg.TokenV2)
	notionUserID := firstNonEmpty(strings.TrimSpace(cli.NotionUserID), cfg.NotionUserID)
	activeUserID := firstNonEmpty(strings.TrimSpace(cli.ActiveUserID), cfg.ActiveUserID)
	cookie := firstNonEmpty(strings.TrimSpace(cli.Cookie), cfg.Cookie)

	client, err := notionclient.New(notionclient.Options{
		BaseURL:      strings.TrimSpace(baseURL),
		TokenV2:      strings.TrimSpace(tokenV2),
		NotionUserID: strings.TrimSpace(notionUserID),
		ActiveUserID: strings.TrimSpace(activeUserID),
		Cookie:       strings.TrimSpace(cookie),
	})
	if err != nil {
		return err
	}

	runCtx := context.WithValue(context.Background(), clientContextKey{}, client)
	runCtx = context.WithValue(runCtx, configPathContextKey{}, config.ResolvePath(cli.ConfigPath))
	ctx.BindTo(runCtx, (*context.Context)(nil))

	if err := ctx.Run(); err != nil {
		if msg := strings.TrimSpace(err.Error()); msg != "" {
			_, _ = fmt.Fprintln(os.Stderr, msg)
		}
		return err
	}
	return nil
}

type clientContextKey struct{}
type configPathContextKey struct{}

func ClientFromContext(ctx context.Context) *notionclient.Client {
	v := ctx.Value(clientContextKey{})
	client, _ := v.(*notionclient.Client)
	return client
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func ConfigPathFromContext(ctx context.Context) string {
	if ctx == nil {
		return config.ResolvePath("")
	}
	if v := ctx.Value(configPathContextKey{}); v != nil {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	return config.ResolvePath("")
}

func hasAnyAuthMaterial() bool {
	envToken := strings.TrimSpace(os.Getenv("NOTION_TOKEN_V2"))
	envUser := strings.TrimSpace(os.Getenv("NOTION_USER_ID"))
	envCookie := strings.TrimSpace(os.Getenv("NOTION_COOKIE"))
	if envCookie != "" || (envToken != "" && envUser != "") {
		return true
	}

	cfgPath := config.ResolvePath(os.Getenv("NOTION_CONFIG"))
	cfg, err := config.Read(cfgPath)
	if err != nil {
		return false
	}
	if strings.TrimSpace(cfg.Cookie) != "" {
		return true
	}
	return strings.TrimSpace(cfg.TokenV2) != "" && strings.TrimSpace(cfg.NotionUserID) != ""
}

func printAuthBootstrapHint() {
	_, _ = fmt.Fprintln(os.Stdout, "")
	_, _ = fmt.Fprintln(os.Stdout, "No auth config found.")
	_, _ = fmt.Fprintln(os.Stdout, "Quick setup:")
	_, _ = fmt.Fprintln(os.Stdout, "  1) Open Notion in browser and sign in.")
	_, _ = fmt.Fprintln(os.Stdout, "  2) DevTools -> Network -> pick a /api/v3/... request.")
	_, _ = fmt.Fprintln(os.Stdout, "  3) Right click -> Copy -> Copy as cURL.")
	_, _ = fmt.Fprintln(os.Stdout, "  4) Run: pbpaste | nocli auth import-curl")
}

func printTopLevelHelp() {
	_, _ = fmt.Fprintln(os.Stdout, "Usage: nocli <command> [flags]")
	_, _ = fmt.Fprintln(os.Stdout, "")
	_, _ = fmt.Fprintln(os.Stdout, "CLI for Notion browser/private endpoints")
	_, _ = fmt.Fprintln(os.Stdout, "")
	_, _ = fmt.Fprintln(os.Stdout, "Common commands:")
	_, _ = fmt.Fprintln(os.Stdout, "  nocli page fetch <url-or-id>")
	_, _ = fmt.Fprintln(os.Stdout, "  nocli page objects <url-or-id>")
	_, _ = fmt.Fprintln(os.Stdout, "  nocli block get <block-id>")
	_, _ = fmt.Fprintln(os.Stdout, "  nocli collection query <collection-id> <view-id>")
	_, _ = fmt.Fprintln(os.Stdout, "  nocli auth import-curl")
	_, _ = fmt.Fprintln(os.Stdout, "")
	_, _ = fmt.Fprintln(os.Stdout, "Run 'nocli --help' for full help.")
}
