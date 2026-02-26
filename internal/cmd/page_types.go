package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/tashi/notion/internal/notionclient"
)

type PageTypesCmd struct {
	URLOrID string `arg:"" help:"Notion page URL or page ID" name:"url_or_id"`
	Output  string `name:"output" short:"o" help:"Write JSON output to this file instead of stdout"`
}

func (c *PageTypesCmd) Run(ctx context.Context) error {
	client := ClientFromContext(ctx)
	if client == nil {
		return fmt.Errorf("internal error: notion client missing from context")
	}

	pageID, err := notionclient.ParsePageID(c.URLOrID)
	if err != nil {
		return err
	}

	resp, err := client.LoadPageChunk(ctx, pageID)
	if err != nil {
		resp, err = client.LoadCachedPageChunkV2(ctx, pageID)
		if err != nil {
			return fmt.Errorf("fetch page %s for types: %w", pageID, err)
		}
	}

	flat := notionclient.FlattenRecordMap(resp)
	seen := map[string]int{}
	for _, row := range flat["block"] {
		t, _ := row["type"].(string)
		t = strings.TrimSpace(strings.ToLower(t))
		if t != "" {
			seen[t]++
		}
	}

	officialSet := map[string]bool{}
	for _, t := range notionclient.PublicAPISupportedBlockTypes {
		officialSet[t] = true
	}

	unsupportedByPublic := make([]string, 0)
	for t := range seen {
		if !officialSet[t] {
			unsupportedByPublic = append(unsupportedByPublic, t)
		}
	}

	return writeJSON(c.Output, map[string]any{
		"page_id":                     pageID,
		"seen_block_types":            seen,
		"public_api_documented_types": notionclient.PublicAPISupportedBlockTypes,
		"not_in_public_api_type_list": notionclient.SortedKeys(mapFromSlice(unsupportedByPublic)),
	})
}

func mapFromSlice(in []string) map[string]struct{} {
	out := make(map[string]struct{}, len(in))
	for _, s := range in {
		out[s] = struct{}{}
	}
	return out
}
