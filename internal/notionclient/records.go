package notionclient

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// FlattenRecordMap converts Notion's varying recordMap envelopes into
// table -> id -> record value.
func FlattenRecordMap(payload map[string]any) map[string]map[string]map[string]any {
	out := map[string]map[string]map[string]any{}
	rm, ok := payload["recordMap"].(map[string]any)
	if !ok {
		return out
	}

	for table, tableRaw := range rm {
		tableMap, ok := tableRaw.(map[string]any)
		if !ok || strings.HasPrefix(table, "__") {
			continue
		}
		rows := map[string]map[string]any{}
		for id, rowRaw := range tableMap {
			if row, ok := unwrapRecordValue(rowRaw); ok {
				rows[id] = row
			}
		}
		if len(rows) > 0 {
			out[table] = rows
		}
	}

	return out
}

func unwrapRecordValue(v any) (map[string]any, bool) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, false
	}

	cur := m
	for {
		nextRaw, has := cur["value"]
		if !has {
			break
		}
		next, ok := nextRaw.(map[string]any)
		if !ok {
			break
		}
		cur = next
	}

	if len(cur) == 0 {
		return nil, false
	}

	return cur, true
}

func TableCounts(flat map[string]map[string]map[string]any) map[string]int {
	counts := map[string]int{}
	for table, rows := range flat {
		counts[table] = len(rows)
	}
	return counts
}

func SortedKeys[M ~map[string]V, V any](m M) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func ToPartialUserObject(userID string) map[string]any {
	if strings.TrimSpace(userID) == "" {
		return nil
	}
	return map[string]any{
		"object": "user",
		"id":     userID,
	}
}

func MillisToISO8601(v any) string {
	ms, err := parseInt64(v)
	if err != nil || ms <= 0 {
		return ""
	}
	return time.UnixMilli(ms).UTC().Format(time.RFC3339Nano)
}

func parseInt64(v any) (int64, error) {
	switch x := v.(type) {
	case int:
		return int64(x), nil
	case int64:
		return x, nil
	case float64:
		return int64(x), nil
	case float32:
		return int64(x), nil
	default:
		return 0, fmt.Errorf("unsupported numeric type %T", v)
	}
}

func NormalizeBlockObject(block map[string]any) map[string]any {
	obj := map[string]any{
		"object": "block",
	}

	if id, _ := block["id"].(string); id != "" {
		obj["id"] = id
	}
	if typ, _ := block["type"].(string); typ != "" {
		obj["type"] = typ
	}

	if parentID, _ := block["parent_id"].(string); parentID != "" {
		parentType := "block_id"
		if parentTable, _ := block["parent_table"].(string); parentTable == "collection" {
			parentType = "database_id"
		} else if parentTable == "space" {
			parentType = "workspace"
		}
		parent := map[string]any{"type": parentType}
		switch parentType {
		case "database_id":
			parent["database_id"] = parentID
		case "workspace":
			parent["workspace"] = true
		default:
			parent["block_id"] = parentID
		}
		obj["parent"] = parent
	}

	if ts := MillisToISO8601(block["created_time"]); ts != "" {
		obj["created_time"] = ts
	}
	if ts := MillisToISO8601(block["last_edited_time"]); ts != "" {
		obj["last_edited_time"] = ts
	}
	if createdBy, _ := block["created_by_id"].(string); createdBy != "" {
		obj["created_by"] = ToPartialUserObject(createdBy)
	}
	if editedBy, _ := block["last_edited_by_id"].(string); editedBy != "" {
		obj["last_edited_by"] = ToPartialUserObject(editedBy)
	}

	alive, aliveSet := block["alive"].(bool)
	if aliveSet {
		obj["archived"] = !alive
		obj["in_trash"] = !alive
	}

	content, _ := block["content"].([]any)
	obj["has_children"] = len(content) > 0

	if typ, _ := block["type"].(string); typ != "" {
		typePayload := map[string]any{}
		if props, ok := block["properties"].(map[string]any); ok {
			typePayload["properties"] = props
		}
		if format, ok := block["format"].(map[string]any); ok {
			typePayload["format"] = format
		}
		if len(content) > 0 {
			typePayload["children"] = content
		}
		if len(typePayload) > 0 {
			obj[typ] = typePayload
		}
	}

	obj["private_value"] = block
	return obj
}
