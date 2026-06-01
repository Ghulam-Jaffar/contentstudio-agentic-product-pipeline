# ContentStudio — Looker Studio Community Connector

## Overview

A Google Apps Script (GAS) community connector that lets ContentStudio users connect their social analytics data directly to Google Looker Studio for custom reporting.

Users access it via the **Looker Studio** button in the analytics tab bar. A modal lets them choose between a pre-built platform template (Facebook, Instagram, LinkedIn) or a fresh data source. The Go API generates the appropriate deep link and the user is taken directly to Looker Studio with all credentials pre-filled — no manual entry required.

---

## File Structure

```
src/looker-studio/
├── main.gs        — Auth, config flow, schema/data dispatch, shared utilities
├── facebook.gs    — Facebook fields + data fetcher
├── instagram.gs   — Instagram fields + data fetcher
├── linkedin.gs    — LinkedIn fields + data fetcher
├── tiktok.gs      — TikTok fields + data fetcher
├── youtube.gs     — YouTube fields + data fetcher
├── pinterest.gs   — Pinterest fields + data fetcher
├── twitter.gs     — Twitter/X fields + data fetcher
└── gmb.gs         — Google My Business fields + data fetcher
```

All `.gs` files share a single global scope (GAS project). Platform functions (`getFields_*`, `getData_*`) defined in platform files are called directly from `main.gs`.

---

## Backend Endpoint

```
GET /analytics/looker-studio/connect
Authorization: Bearer <jwt>
```

### Query Parameters

| Param | Required | Description |
|---|---|---|
| `platform` | ✅ | Platform key (`facebook`, `instagram`, `linkedin`, etc.) |
| `workspace_id` | ✅ | Active workspace `_id` |
| `account_id` | ✅ | Platform account ID |
| `template_id` | ❌ | Looker Studio template report ID — triggers template copy flow |

### Response

```json
{
  "status": true,
  "url": "https://lookerstudio.google.com/..."
}
```

### Go Config

| Env Var | Description |
|---|---|
| `APP_LOOKER_STUDIO_CONNECTOR_ID` | GAS deployment ID (Apps Script → Deploy → Manage deployments) |
| `APP_LOOKER_STUDIO_ENV` | Environment passed to the connector (`qa`, `production`) |

---

## Deep Link Formats

### Template Copy (when `template_id` is provided)

Opens a copy of the pre-built report with the connector pre-configured. Bypasses the config wizard entirely.

```
https://lookerstudio.google.com/reporting/create
  ?c.reportId={templateId}
  &ds.ds0.connector=community
  &ds.ds0.connectorId={connectorDeploymentId}
  &ds.ds0.access_token={apiKey}
  &ds.ds0.workspace_id={workspaceId}
  &ds.ds0.platform={platform}
  &ds.ds0.account_id={accountId}
```

> **Important:** `ds.ds0.connector` must be the literal string `community`. The actual deployment ID goes in `ds.ds0.connectorId`. Using the deployment ID in `connector` directly causes a "not a valid value" error.

### Fresh Data Source (no `template_id`)

Opens the connector setup page to create a new data source. All config params are pre-filled via `connectorConfig`.

```
https://lookerstudio.google.com/datasources/create
  ?connectorId={connectorDeploymentId}
  &connectorConfig={urlEncodedJSON}
```

Where `connectorConfig` is a URL-encoded JSON object:
```json
{
  "access_token": "<api_key>",
  "workspace_id": "<workspace_id>",
  "platform": "<platform>",
  "account_id": "<account_id>"
}
```

---

## Available Templates

| Platform | Template Name | Report ID |
|---|---|---|
| Facebook | Facebook Analytics Dashboard | `ff026271-696a-4bf2-8140-29115808d46e` |
| Instagram | Instagram Analytics Dashboard | `556f5ac6-f00d-40ad-8ff9-8bc5785fbab4` |
| LinkedIn | LinkedIn Analytics Dashboard | `5936a476-decc-4d3a-a933-bcebe31b4932` |

Templates are defined in the frontend at:
`src/modules/analytics/views/common/LookerStudioModal.vue` → `PLATFORM_TEMPLATES`

Platforms without a template (TikTok, YouTube, Pinterest, Twitter, GMB) skip the template/fresh choice and go directly to the fresh data source flow.

---

## Connector Config Flow (GAS)

### Fast Path (deep link from ContentStudio)

When `workspace_id`, `platform`, and `account_id` all arrive in `request.configParams`:

1. Token is persisted to `UserProperties` (`cs_token`)
2. Workspace, platform, and account selectors are rendered (pre-selected)
3. `config.setIsSteppedConfig(false)` — wizard is marked complete immediately
4. User clicks **Connect** once and is done

### Stepped Path (manual connector setup in Looker Studio)

Fallback when params are missing:

```
Step 1 — API Key       paste ContentStudio API key
Step 2 — Workspace     dropdown from GET /api/v1/workspaces
Step 3 — Platform      static list (8 platforms)
Step 4 — Account       dropdown from GET /api/v1/workspaces/{id}/accounts
```

---

## Data Source Alias

All templates use `ds0` as the data source alias. This must match the alias in the template report's data source configuration. If a template is rebuilt with a different alias, update `defaultDatasourceAlias` in `handler.go`.

---

## Deployment

1. Open the GAS project at [script.google.com](https://script.google.com)
2. Copy all `.gs` file contents into the corresponding GAS files
3. Deploy → Manage deployments → New deployment (type: Add-on)
4. Set access to **Anyone with a Google Account** (required for template copy URLs)
5. Copy the Deployment ID into `APP_LOOKER_STUDIO_CONNECTOR_ID`

### Current Deployment ID
```
AKfycbwEZw5klORFkztSMJj8xi8Th0pu4I7bMMkdUd2c8b3ahBOXOozhSb3HuEVUY6jYy3F9
```

> **Note:** The deployment access must be set to **"Anyone with a Google Account"** or broader. A deployment set to "Only myself" will cause template copy URLs to fail with "not a valid value for ds.ds0.connector".

---

## Frontend Integration

### Modal UI (`LookerStudioModal.vue`)

The modal is triggered from `TabsComponent.vue` via the **Looker Studio** button in the analytics tab bar.

**Modal sections:**
1. **Auto-detected context** — shows Workspace, Platform, Account with green checkmarks (pre-filled from current analytics view)
2. **How would you like to proceed?** — shown only for platforms with a template (Facebook, Instagram, LinkedIn):
   - **Use Template** (pre-selected, Recommended) — copies the pre-built dashboard with data wired in
   - **Start Fresh** — creates a new data source from scratch
3. **CTA button** — "Open Template in Looker Studio" or "Open in Looker Studio" depending on selection

### Button Placement

The Looker Studio button sits in the tabs row alongside the Sync data button:
- File: `src/modules/analytics/views/common/TabsComponent.vue`
- Shown for all platforms except `overview` and `group` types
- Tooltip: "View analytics in Looker Studio" (appears to the left)

---

## Environment Resolution (GAS)

All API URLs are baked into the GAS script via `getEnv()`:

```js
function getEnv() {
  return {
    ANALYTICS:         'https://features-analytics-pipeline.contentstudio.io/analytics/overview/',
    ANALYTICS_GO:      'https://features-analytics-pipeline.contentstudio.io/analytics/overview/',
    ANALYTICS_BACKEND: 'https://qa-api.contentstudio.io'
  };
}
```

---

## Troubleshooting

| Error | Cause | Fix |
|---|---|---|
| "not a valid value for ds.ds0.connector" | Deployment access is restricted, or `connector` param has deployment ID instead of `community` | Set deployment access to "Anyone"; ensure `ds.ds0.connector=community` and ID is in `ds.ds0.connectorId` |
| Redirected to manual config wizard | Template copy URL params not received by connector | Verify all 4 params (`access_token`, `workspace_id`, `platform`, `account_id`) are present in the URL |
| "No active API key found" | User has no API key in MongoDB | User must generate an API key from ContentStudio Settings |
| Template opens but shows no data | Data source alias mismatch (`ds0` vs template's actual alias) | Check template's data source alias in Looker Studio → Resource → Manage added data sources |
