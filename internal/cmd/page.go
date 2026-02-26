package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jodok/nocli/internal/notionclient"
)

type PageCmd struct {
	Fetch   PageFetchCmd   `cmd:"" help:"Fetch a page via Notion private endpoints"`
	Objects PageObjectsCmd `cmd:"" help:"Expose flattened objects from a page recordMap"`
	Types   PageTypesCmd   `cmd:"" help:"List block types seen in page vs official Notion API block types"`
}

type PageFetchCmd struct {
	URLOrID  string `arg:"" help:"Notion page URL or page ID" name:"url_or_id"`
	Output   string `name:"output" short:"o" help:"Write JSON output to this file instead of stdout"`
	Endpoint string `name:"endpoint" enum:"auto,loadPageChunk,loadCachedPageChunkV2" default:"auto" help:"Endpoint strategy to use"`
}

func (c *PageFetchCmd) Run(ctx context.Context) error {
	client := ClientFromContext(ctx)
	if client == nil {
		return fmt.Errorf("internal error: notion client missing from context")
	}

	pageID, err := notionclient.ParsePageID(c.URLOrID)
	if err != nil {
		return err
	}

	var resp map[string]any
	switch c.Endpoint {
	case "loadPageChunk":
		resp, err = client.LoadPageChunk(ctx, pageID)
	case "loadCachedPageChunkV2":
		resp, err = client.LoadCachedPageChunkV2(ctx, pageID)
	default:
		resp, err = client.LoadPageChunk(ctx, pageID)
		if err != nil {
			resp, err = client.LoadCachedPageChunkV2(ctx, pageID)
		}
	}
	if err != nil {
		return fmt.Errorf("fetch page %s via %s: %w", pageID, c.Endpoint, err)
	}

	body, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal response json: %w", err)
	}
	body = append(body, '\n')

	if c.Output != "" {
		if err := os.WriteFile(c.Output, body, 0o600); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}
		return nil
	}

	_, err = os.Stdout.Write(body)
	return err
}
