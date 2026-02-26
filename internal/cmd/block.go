package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jodok/nocli/internal/notionclient"
)

type BlockCmd struct {
	Get      BlockGetCmd      `cmd:"" help:"Fetch a block record by ID"`
	Children BlockChildrenCmd `cmd:"" help:"Fetch one-level child blocks"`
}

type BlockGetCmd struct {
	ID              string `arg:"" name:"id" help:"Block ID (UUID or 32-char)"`
	NotionBlockLike bool   `name:"notion-block-like" help:"Emit Notion-like block object shape"`
	Output          string `name:"output" short:"o" help:"Write JSON output to this file instead of stdout"`
}

type BlockChildrenCmd struct {
	ID              string `arg:"" name:"id" help:"Parent block ID (UUID or 32-char)"`
	NotionBlockLike bool   `name:"notion-block-like" help:"Emit Notion-like block object shape"`
	Output          string `name:"output" short:"o" help:"Write JSON output to this file instead of stdout"`
}

func (c *BlockGetCmd) Run(ctx context.Context) error {
	client := ClientFromContext(ctx)
	if client == nil {
		return fmt.Errorf("internal error: notion client missing from context")
	}

	id, err := notionclient.ParsePageID(c.ID)
	if err != nil {
		return fmt.Errorf("parse block id: %w", err)
	}

	resp, err := client.SyncBlockRecords(ctx, []string{id})
	if err != nil {
		return err
	}

	flat := notionclient.FlattenRecordMap(resp)
	blocks := flat["block"]
	row := blocks[id]
	if len(row) == 0 {
		for _, v := range blocks {
			row = v
			break
		}
	}
	if len(row) == 0 {
		return fmt.Errorf("block not found in response")
	}

	var out any
	if c.NotionBlockLike {
		out = notionclient.NormalizeBlockObject(row)
	} else {
		out = map[string]any{"id": id, "object": row}
	}

	return writeJSON(c.Output, out)
}

func (c *BlockChildrenCmd) Run(ctx context.Context) error {
	client := ClientFromContext(ctx)
	if client == nil {
		return fmt.Errorf("internal error: notion client missing from context")
	}

	id, err := notionclient.ParsePageID(c.ID)
	if err != nil {
		return fmt.Errorf("parse block id: %w", err)
	}

	resp, err := client.SyncBlockRecords(ctx, []string{id})
	if err != nil {
		return err
	}
	flat := notionclient.FlattenRecordMap(resp)
	parent := flat["block"][id]
	if len(parent) == 0 {
		return fmt.Errorf("parent block not found")
	}
	childIDs := extractChildIDs(parent)
	if len(childIDs) == 0 {
		return writeJSON(c.Output, map[string]any{"parent_id": id, "children": []any{}})
	}

	childResp, err := client.SyncBlockRecords(ctx, childIDs)
	if err != nil {
		return err
	}
	childrenFlat := notionclient.FlattenRecordMap(childResp)
	blocks := childrenFlat["block"]

	children := make([]any, 0, len(childIDs))
	for _, childID := range childIDs {
		row := blocks[childID]
		if len(row) == 0 {
			continue
		}
		if c.NotionBlockLike {
			children = append(children, notionclient.NormalizeBlockObject(row))
		} else {
			children = append(children, map[string]any{"id": childID, "object": row})
		}
	}

	return writeJSON(c.Output, map[string]any{
		"parent_id": id,
		"children":  children,
	})
}

func extractChildIDs(block map[string]any) []string {
	arr, _ := block["content"].([]any)
	ids := make([]string, 0, len(arr))
	for _, x := range arr {
		s, _ := x.(string)
		s = strings.TrimSpace(s)
		if s != "" {
			ids = append(ids, s)
		}
	}
	return ids
}

func writeJSON(path string, v any) error {
	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	body = append(body, '\n')

	if path != "" {
		if err := os.WriteFile(path, body, 0o600); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}
		return nil
	}

	_, err = os.Stdout.Write(body)
	return err
}
