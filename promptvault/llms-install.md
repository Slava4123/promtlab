# LLM install guide for ПромтЛаб MCP server

**Note to the installing agent (Cline, Claude, etc.):** ПромтЛаб is a
**remote** MCP server over Streamable HTTP. Do **not** clone this repo
or run anything locally — there is no `npm install` or container to
launch. The only setup is creating an API key and pointing the client
at our public endpoint.

## One-time setup

1. Open <https://promtlabs.ru> and sign up (email / GitHub / Google / Яндекс).
2. Go to **Settings → API-ключи → Создать**. Copy the key — it's shown once and starts with `pvlt_`.

## Client configuration

Add the following MCP server definition to the client config:

```json
{
  "mcpServers": {
    "promptvault": {
      "url": "https://promtlabs.ru/mcp",
      "headers": {
        "Authorization": "Bearer pvlt_YOUR_KEY_HERE"
      }
    }
  }
}
```

For Cline, paste the block above into the `cline_mcp_settings.json`
file (usually `~/.config/Cline/cline_mcp_settings.json` on macOS/Linux,
`%APPDATA%\Cline\cline_mcp_settings.json` on Windows).

For Claude Code, run instead:

```bash
claude mcp add promptvault --transport http https://promtlabs.ru/mcp \
  --header "Authorization: Bearer pvlt_YOUR_KEY_HERE"
```

## Verify

After restarting the client, ask it:

> "Покажи мои ПромтЛаб-инструменты"

It should list around 30 tools (`search_prompts`, `list_prompts`,
`get_prompt`, `create_prompt`, `list_teams`, `whoami`, `list_trash`,
…). If the list is empty or the client reports "unauthorized", re-check
the API key value in the config.

## Troubleshooting

- **401 / unauthorized** — API key is invalid or revoked. Create a new
  key in the web UI and update the config.
- **403 / read-only access** — you're a viewer in a team; writes are
  disabled. Ask an owner/editor to upgrade your role.
- **429 / too many requests** — rate limit (60 req/min/user). Wait a
  minute.
- **402 / MCP quota exceeded** — you hit the daily MCP-write quota
  (Free: 5, Pro: 30, Max: unlimited). Upgrade or wait until midnight UTC+3.

## Full documentation

- User-facing: <https://promtlabs.ru/help/mcp>
- Developer: <https://github.com/Slava4123/promtlab/blob/main/promptvault/docs/MCP.md>
- Server manifest (for MCP Registry): <https://github.com/Slava4123/promtlab/blob/main/promptvault/server.json>
