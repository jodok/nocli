package notionclient

import (
	"context"
)

type loadCachedPageChunkRequest struct {
	Page            pageRef `json:"page"`
	Limit           int     `json:"limit"`
	ChunkNumber     int     `json:"chunkNumber"`
	Cursor          cursor  `json:"cursor"`
	VerticalColumns bool    `json:"verticalColumns"`
}

type pageRef struct {
	ID string `json:"id"`
}

type cursor struct {
	Stack []any `json:"stack"`
}

func (c *Client) LoadCachedPageChunkV2(ctx context.Context, pageID string) (map[string]any, error) {
	payload := loadCachedPageChunkRequest{
		Page:            pageRef{ID: pageID},
		Limit:           100,
		ChunkNumber:     0,
		Cursor:          cursor{Stack: []any{}},
		VerticalColumns: false,
	}
	return c.postJSON(ctx, "/api/v3/loadCachedPageChunkV2", payload)
}

type loadPageChunkRequest struct {
	PageID          string `json:"pageId"`
	Limit           int    `json:"limit"`
	ChunkNumber     int    `json:"chunkNumber"`
	Cursor          cursor `json:"cursor"`
	VerticalColumns bool   `json:"verticalColumns"`
}

func (c *Client) LoadPageChunk(ctx context.Context, pageID string) (map[string]any, error) {
	payload := loadPageChunkRequest{
		PageID:          pageID,
		Limit:           100,
		ChunkNumber:     0,
		Cursor:          cursor{Stack: []any{}},
		VerticalColumns: false,
	}
	return c.postJSON(ctx, "/api/v3/loadPageChunk", payload)
}
