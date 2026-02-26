package notionclient

import "context"

type syncRequest struct {
	Table   string `json:"table"`
	ID      string `json:"id"`
	Version int    `json:"version"`
}

type syncRecordValuesRequest struct {
	Requests []syncRequest `json:"requests"`
}

type getRecordValuesRequest struct {
	Requests []syncRequest `json:"requests"`
}

type queryCollectionRequest struct {
	Collection     pageRef        `json:"collection"`
	CollectionView pageRef        `json:"collectionView"`
	Source         map[string]any `json:"source"`
	Loader         map[string]any `json:"loader"`
}

func (c *Client) SyncBlockRecords(ctx context.Context, blockIDs []string) (map[string]any, error) {
	reqs := make([]syncRequest, 0, len(blockIDs))
	for _, id := range blockIDs {
		reqs = append(reqs, syncRequest{Table: "block", ID: id, Version: -1})
	}
	return c.postJSON(ctx, "/api/v3/syncRecordValuesMain", syncRecordValuesRequest{Requests: reqs})
}

func (c *Client) GetUsers(ctx context.Context, userIDs []string) (map[string]any, error) {
	reqs := make([]syncRequest, 0, len(userIDs))
	for _, id := range userIDs {
		reqs = append(reqs, syncRequest{Table: "notion_user", ID: id, Version: -1})
	}
	return c.postJSON(ctx, "/api/v3/getRecordValues", getRecordValuesRequest{Requests: reqs})
}

func (c *Client) QueryCollection(ctx context.Context, collectionID string, viewID string, limit int) (map[string]any, error) {
	if limit <= 0 {
		limit = 100
	}
	payload := queryCollectionRequest{
		Collection:     pageRef{ID: collectionID},
		CollectionView: pageRef{ID: viewID},
		Source: map[string]any{
			"type": "collection",
			"id":   collectionID,
		},
		Loader: map[string]any{
			"type": "reducer",
			"reducers": map[string]any{
				"collection_group_results": map[string]any{
					"type":             "results",
					"limit":            limit,
					"loadContentCover": true,
				},
			},
			"sort":         []any{},
			"filter":       map[string]any{"filters": []any{}, "operator": "and"},
			"searchQuery":  "",
			"userTimeZone": "America/Los_Angeles",
		},
	}
	return c.postJSON(ctx, "/api/v3/queryCollection?src=initial_load", payload)
}
