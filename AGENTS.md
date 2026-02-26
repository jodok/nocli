# AGENTS.md

## Project Purpose
- Build a CLI that accesses Notion data via private/browser endpoints, not the public Notion API.
- Current priority: fetch private page content similarly to what the browser can load.

## Current State
- Language: Go (`go1.26.0` verified locally).
- CLI entrypoint: `cmd/notion/main.go`.
- Root command wiring: `internal/cmd/root.go`.
- Page fetch command: `internal/cmd/page.go`.
- HTTP client: `internal/notionclient/client.go`.
- Page endpoints:
  - `loadPageChunk` (primary)
  - `loadCachedPageChunkV2` (fallback)
- Object exposure commands:
  - `auth import-curl` (extract creds from pasted DevTools cURL)
  - `page objects` (flattened `recordMap`, optional Notion-like block objects)
  - `page types` (seen block types vs public API list)
  - `block get`, `block children`
  - `collection query`

## Auth + Config
- Default config file path: `.notion.json` (repo root).
- Config loader: `internal/config/config.go`.
- Supported fields:
  - `base_url`
  - `token_v2`
  - `notion_user_id`
  - `active_user_id`
  - `cookie`
- `.notion.json` is gitignored and expected to be `0600`.

## Command to Test
```bash
cd /Users/tashi/sandbox/notion
go run ./cmd/notion page fetch 'https://www.notion.so/pinateam/Private-Jodok-246bbe48fc7e804e92c6d77450bb136f'
```

## Practical Notes
- Public API block type reference is tracked in `internal/notionclient/block_types.go` and used by `page types`.
- `loadPageChunk` expects UUID-formatted page IDs (`8-4-4-4-12`), not plain 32-char hex.
- Verified working fetch on:
  - `https://www.notion.so/pinateam/Private-Jodok-246bbe48fc7e804e92c6d77450bb136f`
  - using only `token_v2` + `notion_user_id` from `.notion.json` (no extra cookie required for this page).
- If fetch fails with auth/permission issues, capture browser request headers for:
  - `x-notion-active-user-header`
  - full cookie string (if token-only auth is insufficient)
- If payload mismatch occurs, compare browser Network request body for page load and update request structs.
- Keep endpoint strategy `auto` unless debugging a specific endpoint.

## Security Notes
- Never commit raw session credentials.
- Do not print credentials in logs or error messages.
- Treat `.notion.json` as local secret material.

## Next Engineering Steps
1. Add a `config set` command to update `.notion.json` safely from CLI.
2. Add typed conversion for more public-API-like objects (rich_text, files, mentions, emoji).
3. Add integration-style command tests with mocked HTTP transport.
4. Add retries/backoff for transient 5xx and rate-limit responses.
