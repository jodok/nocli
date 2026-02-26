package cmd

import (
	"context"
	"fmt"

	"github.com/tashi/notion/internal/notionclient"
)

type CollectionCmd struct {
	Query CollectionQueryCmd `cmd:"" help:"Query a collection view"`
}

type CollectionQueryCmd struct {
	CollectionID string `arg:"" name:"collection_id" help:"Collection/database ID"`
	ViewID       string `arg:"" name:"view_id" help:"Collection view ID"`
	Limit        int    `name:"limit" default:"500" help:"Result limit"`
	Output       string `name:"output" short:"o" help:"Write JSON output to this file instead of stdout"`
	Flatten      bool   `name:"flatten" help:"Emit flattened objects instead of raw response"`
}

func (c *CollectionQueryCmd) Run(ctx context.Context) error {
	client := ClientFromContext(ctx)
	if client == nil {
		return fmt.Errorf("internal error: notion client missing from context")
	}

	collectionID, err := notionclient.ParsePageID(c.CollectionID)
	if err != nil {
		return fmt.Errorf("parse collection id: %w", err)
	}
	viewID, err := notionclient.ParsePageID(c.ViewID)
	if err != nil {
		return fmt.Errorf("parse view id: %w", err)
	}

	resp, err := client.QueryCollection(ctx, collectionID, viewID, c.Limit)
	if err != nil {
		return err
	}

	if !c.Flatten {
		return writeJSON(c.Output, resp)
	}

	flat := notionclient.FlattenRecordMap(resp)
	objects := make([]map[string]any, 0, 256)
	for _, table := range notionclient.SortedKeys(flat) {
		for _, id := range notionclient.SortedKeys(flat[table]) {
			objects = append(objects, map[string]any{
				"table":  table,
				"id":     id,
				"object": flat[table][id],
			})
		}
	}

	return writeJSON(c.Output, map[string]any{
		"collection_id": collectionID,
		"view_id":       viewID,
		"counts":        notionclient.TableCounts(flat),
		"objects":       objects,
	})
}
