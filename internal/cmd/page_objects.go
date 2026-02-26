package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jodok/nocli/internal/notionclient"
)

type PageObjectsCmd struct {
	URLOrID         string `arg:"" help:"Notion page URL or page ID" name:"url_or_id"`
	Table           string `name:"table" help:"Filter by table name (block, collection, collection_view, notion_user, ...)"`
	BlockType       string `name:"block-type" help:"Filter blocks by block type"`
	NotionBlockLike bool   `name:"notion-block-like" help:"For block table, emit Notion-like block objects with private value attached"`
	Output          string `name:"output" short:"o" help:"Write JSON output to this file instead of stdout"`
}

type pageObjectsOutput struct {
	PageID   string                    `json:"page_id"`
	Counts   map[string]int            `json:"counts"`
	Objects  []map[string]any          `json:"objects"`
	Warnings []string                  `json:"warnings,omitempty"`
	Meta     map[string]map[string]any `json:"meta,omitempty"`
}

func (c *PageObjectsCmd) Run(ctx context.Context) error {
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
			return fmt.Errorf("fetch page %s for objects: %w", pageID, err)
		}
	}

	flat := notionclient.FlattenRecordMap(resp)
	counts := notionclient.TableCounts(flat)
	filterTable := strings.TrimSpace(c.Table)
	filterBlockType := strings.TrimSpace(strings.ToLower(c.BlockType))

	objects := make([]map[string]any, 0, 128)
	for _, table := range notionclient.SortedKeys(flat) {
		if filterTable != "" && table != filterTable {
			continue
		}
		rows := flat[table]
		for _, id := range notionclient.SortedKeys(rows) {
			row := rows[id]
			if table == "block" {
				if filterBlockType != "" {
					typ, _ := row["type"].(string)
					if strings.ToLower(typ) != filterBlockType {
						continue
					}
				}
				if c.NotionBlockLike {
					n := notionclient.NormalizeBlockObject(row)
					n["table"] = table
					objects = append(objects, n)
					continue
				}
			}

			objects = append(objects, map[string]any{
				"table":  table,
				"id":     id,
				"object": row,
			})
		}
	}

	out := pageObjectsOutput{
		PageID:  pageID,
		Counts:  counts,
		Objects: objects,
		Meta: map[string]map[string]any{
			"filters": {
				"table":             filterTable,
				"block_type":        filterBlockType,
				"notion_block_like": c.NotionBlockLike,
			},
		},
	}

	body, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal objects json: %w", err)
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
