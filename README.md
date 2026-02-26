# notion

CLI prototype for calling Notion's browser/private endpoints (not the public Notion API).

## Commands

- `notion page fetch <url-or-page-id>`: Calls Notion private page endpoints and prints JSON.
- `notion auth import-curl`: Import credentials from a pasted Notion DevTools "Copy as cURL" request.
- `notion page objects <url-or-page-id>`: Exposes flattened `recordMap` objects across all tables.
- `notion page types <url-or-page-id>`: Shows block types seen in the page vs documented public API block types.
- `notion block get <block-id>`: Fetches a single block.
- `notion block children <block-id>`: Fetches direct child blocks.
- `notion collection query <collection-id> <view-id>`: Queries a collection view.

## Auth inputs

The tool reads auth material from flags or environment:

- `NOTION_TOKEN_V2`: value of the `token_v2` cookie
- `NOTION_USER_ID`: value of the `notion_user_id` cookie (optional)
- `NOTION_ACTIVE_USER_ID`: value for `x-notion-active-user-header` (optional)
- `NOTION_COOKIE`: full `Cookie` header string (overrides `NOTION_TOKEN_V2`/`NOTION_USER_ID`)

## Endpoint strategy

`page fetch` supports:

- `--endpoint auto` (default): try `loadPageChunk`, then fallback to `loadCachedPageChunkV2`
- `--endpoint loadPageChunk`
- `--endpoint loadCachedPageChunkV2`

## Example

```bash
NOTION_TOKEN_V2='...' \
NOTION_USER_ID='...' \
NOTION_ACTIVE_USER_ID='...' \
go run ./cmd/notion page fetch \
  'https://www.notion.so/pinateam/Private-Jodok-246bbe48fc7e804e92c6d77450bb136f'
```

### Easiest auth import flow

1. In browser DevTools on a Notion page, open Network and select a Notion API request.
2. Right-click request -> Copy -> Copy as cURL.
3. Paste into:

```bash
pbpaste | go run ./cmd/notion auth import-curl
```

This stores extracted values into `.notion.json` (`token_v2`, `notion_user_id`, and optionally `active_user_id`).

You can force the second endpoint with:

```bash
go run ./cmd/notion page fetch --endpoint loadCachedPageChunkV2 '<url>'
```

Get Notion-like block objects:

```bash
go run ./cmd/notion page objects '<url>' --table block --notion-block-like
```

Query a board/database directly:

```bash
go run ./cmd/notion collection query '<collection-id>' '<view-id>' --flatten
```
