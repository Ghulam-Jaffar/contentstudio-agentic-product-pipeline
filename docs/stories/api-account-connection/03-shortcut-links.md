# Shortcut Links: Social Account Connection via API + Facebook Background Text Posts

## Epics

| Epic | Link |
|---|---|
| Publishing API v1.13 — Social Account Connection | https://app.shortcut.com/contentstudio-team/epic/116484 |
| Publishing API v1.14 — Facebook Background Text Posts | https://app.shortcut.com/contentstudio-team/epic/116546 |

## v1.13 Stories — Account Connection

| Story | Epic | Link |
|---|---|---|
| [BE] Add social account OAuth connection endpoints to Publishing API v1 | v1.13 | https://app.shortcut.com/contentstudio-team/story/116489 |
| [BE] Add connect_social_account tool to ContentStudio MCP server | v1.13 | https://app.shortcut.com/contentstudio-team/story/116490 |
| [BE] Add accounts connect command to ContentStudio CLI | CLI & Agent Skills | https://app.shortcut.com/contentstudio-team/story/116491 |

## v1.14 Stories — Facebook Background Text Posts

| Story | Group | Link |
|---|---|---|
| [BE] Add Facebook background text post support to post creation endpoint in Publishing API v1 | Backend | https://app.shortcut.com/contentstudio-team/story/116547 |
| [BE] Update Zapier app to support Facebook background text posts | Backend | https://app.shortcut.com/contentstudio-team/story/116549 |
| [BE] Update Make.com app to support Facebook background text posts | Backend | https://app.shortcut.com/contentstudio-team/story/116550 |
| [BE] Add Facebook background text post support to MCP server create_post tool | Backend | https://app.shortcut.com/contentstudio-team/story/116551 |
| [BE] Support Facebook background text posts in content category scheduling | Backend | https://app.shortcut.com/contentstudio-team/story/116552 |
| [Technical Writing] Document Facebook background preset codes with visual reference for API users | Technical Writing | https://app.shortcut.com/contentstudio-team/story/116553 |

## Dependency Graph

```
v1.13:
  API endpoints (sc-116489)
    ├── blocks → MCP tool (sc-116490)
    └── blocks → CLI command (sc-116491)

v1.14:
  API endpoint (sc-116547)
    ├── blocks → Zapier (sc-116549)
    ├── blocks → Make.com (sc-116550)
    ├── blocks → MCP tool (sc-116551)
    └── blocks → Tech Writing docs (sc-116553)
  Smart scheduling (sc-116552) — independent
```

## Details

- **Iteration:** 06 April - 17 April - 2026
- **Priority:** High (API), Medium (integrations, docs)
- **Product Area:** Publishing
