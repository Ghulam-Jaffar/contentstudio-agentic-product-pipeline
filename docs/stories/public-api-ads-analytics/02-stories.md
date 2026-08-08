# Epic: Ads Analytics across the public developer surfaces

## Problem

ContentStudio supports Meta Ads and Google Ads analytics in the product. Both are fully built: connected ad accounts, ClickHouse-backed ingestion, 15 Meta sections and 21 Google sections served by the analytics service, dedicated dashboard tabs, and PDF reports. None of it is reachable through the public API, which today exposes organic analytics only.

Customers with paid campaigns can therefore see their spend and campaign performance in our UI but cannot pull it. Agencies who want paid and organic side by side in their own client dashboards go back to Meta and Google directly, which is exactly the consolidation they came to ContentStudio for.

## Goal

Expose Meta Ads and Google Ads analytics through the public API, with a deliberate public resource model rather than a mirror of internal chart endpoints, discoverable ad accounts, explicit currency handling, and the same auth, error and rate-limit behaviour as the existing public analytics endpoints. Then propagate to every developer surface we ship.

## Scope (surfaces)

Backend public API first, then: **API docs and OpenAPI spec, CLI (`contentstudio-cli`) plus the agent skill, MCP (`contentstudio-mcp`), the Claude extension, the GPT app, Zapier, Make.com, n8n, and the published ClawHub skill.**

## Rules

- **Design the public resource model, don't mirror internal sections.** Several internal endpoints are chart-shaped (`impressionsVsSpend`, `clicksVsCtr`) rather than resource-shaped. The public API exposes accounts, campaigns, ad sets or ad groups, ads, keywords, conversions, demographics and a summary, with metrics and breakdowns as parameters.
- **Ad accounts must be discoverable.** A caller cannot query what they cannot list. There is no public ad-account listing today.
- **Currency is explicit.** Every monetary figure states its currency. Ad accounts have their own currency and we must never return a bare number for spend.
- **Access control is preserved.** Per-member ad-account access already exists internally and must apply to API-key requests too.
- **AI insights are out of scope for v1.** They are generated narrative, not metrics.
- **Same conventions as the organic public analytics endpoints:** API key auth, workspace-scoped paths, shared error schemas, documented pagination.
- Keep the OpenAPI spec, the API guide, and the agent `SKILL.md` in sync.

## Sequencing

Stories 1 and 2 define the two platform contracts and can run in parallel once the shared resource model is agreed; agreeing that model is the first task inside story 1. Story 3 closes out the documented contract. Stories 4 through 11 depend on both platform contracts being frozen. Within the surfaces: CLI and MCP first since we own them and they feed the agent skill and the ClawHub listing, then the Claude extension and GPT app, then Zapier, Make.com and n8n, and ClawHub last.

## Out of scope

- **Ad management.** This epic is read-only analytics. Creating, pausing or editing campaigns through the API is separate and much larger.
- **New metrics.** Only what the analytics service already ingests. Adding metrics to ingestion is separate work.
- **AI insights endpoints.** Deliberately excluded from v1.
- **New ad platforms.** Meta and Google only, matching what the product supports.
- **Mobile.** Ads analytics is web only.

## Stories

1. `[BE] Add public Ads Analytics API for Meta Ads`
2. `[BE] Add public Ads Analytics API for Google Ads`
3. `[BE] Document the Ads Analytics endpoints in the OpenAPI spec and API guide`
4. `[CLI] Add ads analytics commands to the CLI and the agent skill`
5. `[MCP] Add Meta Ads and Google Ads analytics tools to the MCP server`
6. `[Claude Extension] Surface ads analytics in the Claude extension`
7. `[GPT App] Add ads analytics actions to the GPT app schema`
8. `[Zapier] Add ads analytics search steps`
9. `[Make] Add ads analytics modules`
10. `[n8n] Add ads analytics operations`
11. `[ClawHub] Refresh the published ContentStudio skill with the ads analytics commands`

---
---

# 1. [BE] Add public Ads Analytics API for Meta Ads

### Description

As a ContentStudio customer with Meta ad accounts connected, I want to pull my Meta Ads performance through the public API with my API key, so I can put paid and organic results side by side in my own dashboard or client report instead of exporting PDFs or going back to Meta.

### Workflow

*(Public API change. The "user" here is a developer calling the API with their API key.)*

1. A developer authenticates to the public API with their API key.
2. They list the Meta ad accounts connected to a workspace and see, for each, its name, identifier, status and currency.
3. They request a performance summary for an ad account over a date range and receive the headline metrics with the account's currency stated.
4. They request campaigns for that account, filtered and paginated, and receive per-campaign metrics.
5. They drill into ad sets for a campaign, and ads within an ad set, using the same parameter conventions.
6. They request demographic breakdowns for an account or campaign.
7. Every figure they receive matches what the Meta Ads dashboard shows in ContentStudio for the same account, range and filters.
8. The endpoints are discoverable in the public API reference.

### Acceptance criteria

- [ ] A public endpoint lists the Meta ad accounts connected to a workspace, returning at least name, identifier, status and the account's currency.
- [ ] A public endpoint returns a performance summary for a Meta ad account over a date range, including spend, impressions, clicks, click-through rate, cost per click and results by objective.
- [ ] A public endpoint returns campaigns for an ad account with per-campaign metrics, supporting pagination and the documented filters.
- [ ] A public endpoint returns ad sets, scoped to an ad account or a campaign, with per-ad-set metrics and pagination.
- [ ] A public endpoint returns ads, scoped to an ad account, campaign or ad set, with per-ad metrics and pagination.
- [ ] A public endpoint returns demographic breakdowns, covering both the age-and-gender breakdown and the region-and-country breakdown available internally.
- [ ] A public endpoint returns performance over time for an ad account, at a documented granularity, so callers can chart trends without pulling every row.
- [ ] Breakdown by placement or platform is available, matching the breakdown the product already computes.
- [ ] Every monetary value in every response is accompanied by an explicit currency, and the currency is the ad account's own.
- [ ] Every metric returned is documented with its definition, so a caller knows whether a click figure means link clicks or all clicks.
- [ ] Endpoints authenticate with the customer's API key using the same mechanism and header as the existing public analytics endpoints.
- [ ] Date range, timezone and pagination parameters follow the same conventions as the existing public analytics endpoints.
- [ ] Per-member ad-account access control is enforced: a key belonging to a member without access to an ad account cannot read that account's data.
- [ ] Requesting an ad account outside the API key's workspace returns a permission error and reveals nothing about the account.
- [ ] Requesting an ad account that exists but has no data for the range returns an empty result with a success status, distinguishable from an error.
- [ ] Standard error responses match the existing analytics endpoints: 401 unauthorized, 403 workspace or account permission denied, 422 validation error, 429 rate limited, 502 upstream analytics error.
- [ ] Rate limiting and per-plan API credit accounting are applied consistently with the other public API analytics endpoints, and the credit cost of the paginated list endpoints is defined rather than incidental.
- [ ] Values returned by the public API match what ContentStudio's own Meta Ads analytics shows for the same account, range and filters. No discrepancies.
- [ ] AI insights sections are not exposed by this endpoint set.

### Mock-ups

N/A. Public API endpoints, documented via OpenAPI.

### Impact on existing data

None. Read-only over ads analytics data already ingested into the analytics store. No schema changes and no new ingestion.

### Impact on other products

- Public API: new endpoints for external consumers.
- Analytics service: new callers on existing internal endpoints. Worth confirming the added read load is acceptable, since public API callers pull larger ranges than a dashboard user does.
- Web app, mobile apps and Chrome extension: unaffected, no contract of theirs changes.
- Every downstream surface in this epic builds on this contract.

### Dependencies

- The public resource model shared with **[BE] Add public Ads Analytics API for Google Ads** must be agreed first. That decision belongs to whichever of the two starts.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A, backend only
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A, no UI
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- The data is already served, by the Go service rather than Laravel. `contentstudio-social-analytics-go/src/api/analytics/router.go` registers 15 Meta sections under `/analytics/overview/meta-ads/`: `summary`, `resultsByObjective`, `impressionsVsSpend`, `clicksVsCtr`, `topCampaigns`, `performanceTrend`, `performanceByLevel`, `performanceByPlatform`, `campaignsList`, `adSetsList`, `adsList`, `demographicsAgeGender`, `demographicsRegionCountry`, plus the two AI insights ones. The public endpoints proxy to these the way the Facebook public analytics controller proxies to its organic equivalents.
- Note that `impressionsVsSpend` and `clicksVsCtr` are chart-shaped rather than resource-shaped. Folding them into a single performance-over-time resource with a metrics parameter is likely the better public contract than exposing them one-to-one.
- `contentstudio-social-analytics-go/src/api/analytics/meta_ads/handler.go` centralises parameter parsing in `parseRequest` and routes every section through one `handle` wrapper, so the internal parameter set is uniform and worth reading before defining the public one.
- The public API pattern to mirror is `contentstudio-backend/app/Http/Controllers/Api/V1/Analytics/FacebookAnalyticsController.php`, with the shared error schemas in `AnalyticsApiSchemas.php` and `BaseApiController.php` as the API-key base. Route group goes in `routes/api/v1.php` beside the existing per-network analytics groups.
- Ad accounts are `ExternalLinkIntegration` records, connected via `app/Strategy/Integrations/MetaAdsConnector.php`. Per-member access is already modelled in `app/Data/Workspace/WorkspaceMemberPermissionsData.php` and `app/Http/Requests/Settings/Team/SocialAccountAccessRequest.php`, so the enforcement path exists.
- `app/Services/Analytics/GoReportsClient.php` already treats `meta_ads` as a first-class report type and contains the account-grouping and platform-detection helpers (`groupAccountsByPlatform`, `accountPlatform`, `accountId`) that a public controller will need for validation.
- `app/Libraries/SampleWorkspace/SampleMetaAdsClickhouseSeeder.php` gives a seeded dataset, useful for documentation examples and for testing the endpoints without a live ad spend.

---
---

# 2. [BE] Add public Ads Analytics API for Google Ads

### Description

As a ContentStudio customer with Google Ads accounts connected, I want to pull my Google Ads performance through the public API with my API key, so I can report on paid search alongside my social results in my own tooling.

### Workflow

*(Public API change. The "user" here is a developer calling the API with their API key.)*

1. A developer authenticates to the public API with their API key.
2. They list the Google Ads accounts connected to a workspace, with name, identifier, status and currency.
3. They request a performance summary for an account over a date range.
4. They request campaigns, then ad groups within a campaign, then ads within an ad group, with pagination and the documented filters.
5. They request keywords and search terms for an account or campaign.
6. They request conversion data: conversions over time, conversions by action, and the conversion actions available.
7. They request demographic breakdowns.
8. Every figure matches what the Google Ads dashboard shows in ContentStudio for the same account, range and filters.

### Acceptance criteria

- [ ] A public endpoint lists the Google Ads accounts connected to a workspace, returning at least name, identifier, status and the account's currency.
- [ ] A public endpoint returns a performance summary for a Google Ads account over a date range, including spend, impressions, clicks, click-through rate, cost per click and conversions.
- [ ] A public endpoint returns campaigns with per-campaign metrics, supporting pagination and the documented filters, including breakdown by campaign type.
- [ ] A public endpoint returns ad groups, scoped to an account or campaign, with per-ad-group metrics and pagination.
- [ ] A public endpoint returns ads, scoped to an account, campaign or ad group, with per-ad metrics and pagination.
- [ ] A public endpoint returns keywords with per-keyword metrics, supporting pagination.
- [ ] A public endpoint returns search terms with per-term metrics, supporting pagination.
- [ ] Public endpoints return conversion data: conversions over time, conversions grouped by action, and the list of conversion actions configured on the account.
- [ ] A public endpoint returns the conversion funnel data the product already computes.
- [ ] A public endpoint returns demographic breakdowns.
- [ ] A public endpoint returns performance over time at a documented granularity.
- [ ] Shopping campaign data is either exposed or explicitly documented as not available through the public API, not silently omitted.
- [ ] Every monetary value is accompanied by an explicit currency, and the currency is the ad account's own.
- [ ] Every metric returned is documented with its definition, including how conversions are counted and attributed.
- [ ] Endpoints authenticate with the customer's API key using the same mechanism and header as the existing public analytics endpoints.
- [ ] Date range, timezone and pagination parameters follow the same conventions as the existing public analytics endpoints, and match the Meta Ads endpoints so a caller integrating both learns one pattern.
- [ ] Per-member ad-account access control is enforced.
- [ ] Requesting an ad account outside the API key's workspace returns a permission error and reveals nothing about the account.
- [ ] Requesting an account with no data for the range returns an empty result with a success status.
- [ ] Standard error responses match the existing analytics endpoints: 401, 403, 422, 429, 502.
- [ ] Rate limiting and per-plan API credit accounting are applied consistently with the other public API analytics endpoints.
- [ ] Values returned match what ContentStudio's own Google Ads analytics shows for the same account, range and filters. No discrepancies.
- [ ] AI insights sections are not exposed by this endpoint set.

### Mock-ups

N/A. Public API endpoints, documented via OpenAPI.

### Impact on existing data

None. Read-only over data already ingested. No schema changes.

### Impact on other products

- Public API: new endpoints for external consumers.
- Analytics service: new callers on existing internal endpoints, with the same read-load consideration as the Meta story.
- Web app, mobile apps and Chrome extension: unaffected.

### Dependencies

- Shares the public resource model with **[BE] Add public Ads Analytics API for Meta Ads**. The two must agree on parameter names, pagination and currency handling so a caller integrating both does not learn two conventions.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A, backend only
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A, no UI
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- `contentstudio-social-analytics-go/src/api/analytics/router.go` registers 21 Google sections under `/analytics/overview/google-ads/`: `summary`, `conversionsByAction`, `impressionsVsSpend`, `clicksVsCtr`, `topCampaigns`, `performanceTrend`, `performanceByLevel`, `performanceByType`, `campaignsList`, `adGroupsList`, `adsList`, `keywordsList`, `searchTermsList`, `shoppingList`, `conversionFunnel`, `conversionsOverTime`, `conversionActionsList`, `campaignConversionActions`, `demographics`, plus the two AI insights ones. The public endpoints proxy to these.
- `contentstudio-social-analytics-go/src/api/analytics/google_ads/handler.go` follows the same `parseRequest` plus shared `handle` wrapper structure as the Meta handler, so the two public controllers can share a shape.
- Google Ads has a richer surface than Meta (keywords, search terms, shopping, conversion actions). Where a concept has no Meta equivalent, the public API should still keep shared concepts named identically across the two platforms.
- Accounts are connected via `app/Strategy/Integrations/GoogleAdsConnector.php` and stored as `ExternalLinkIntegration` records. `app/Repository/Integrations/Platforms/ExternalLinkIntegrationRepo.php` is the read path.
- `useAnalyticsRoutes.ts` in the frontend notes that Google Ads is not currently feature-flag gated the way Meta Ads is. Worth confirming whether the public endpoints should be gated at all, since a plan gate on the public API is a billing decision.
- `app/Libraries/SampleWorkspace/SampleGoogleAdsClickhouseSeeder.php` provides a seeded dataset for documentation examples and testing.

---
---

# 3. [BE] Document the Ads Analytics endpoints in the OpenAPI spec and API guide

### Description

As a developer integrating ContentStudio's ads analytics, I want the endpoints fully documented, so I know what every metric means, what currency a figure is in, how to paginate, and which filters exist, without reverse-engineering it from responses.

### Workflow

1. A developer opens the public API reference and finds the ads analytics endpoints for both platforms.
2. Each endpoint lists its parameters, filters, pagination behaviour and response schema.
3. A metric glossary defines every metric returned, including how conversions are counted and what a click means on each platform.
4. The currency contract is documented: where currency appears and that it is the ad account's own.
5. Worked examples show listing ad accounts, then pulling a summary, then paginating campaigns.
6. The differences between the two platforms are stated, including anything Google-only and anything documented as unavailable.
7. The agent skill definition reflects the new capability.

### Acceptance criteria

- [ ] All ads analytics endpoints for both platforms appear in the generated OpenAPI spec with full parameter and response schemas.
- [ ] A metric glossary documents every metric returned by either platform, with its definition and, where the platforms differ, the per-platform meaning.
- [ ] The currency contract is documented explicitly.
- [ ] Pagination is documented, including page size limits and how to page through a large campaign or keyword set.
- [ ] Every supported filter is documented with its accepted values.
- [ ] Worked request and response examples are included for the account listing, the summary, and a paginated list endpoint.
- [ ] Platform differences are documented, including anything available on Google Ads but not Meta Ads, and anything explicitly not exposed.
- [ ] The exclusion of AI insights from the public API is stated, so developers do not go looking for it.
- [ ] Rate limits and credit consumption for the ads endpoints are documented alongside the other analytics endpoints, including the cost of paginated calls.
- [ ] Access-control behaviour is documented: what a key without access to a given ad account receives.
- [ ] The agent `SKILL.md` is updated so an agent using the skill knows ads analytics is available and how to request it.
- [ ] The public API changelog notes the addition.

### Mock-ups

N/A.

### Impact on existing data

None.

### Impact on other products

- Public API documentation and the agent skill definition. No runtime behaviour changes.

### Dependencies

- Depends on **[BE] Add public Ads Analytics API for Meta Ads** and **[BE] Add public Ads Analytics API for Google Ads** for frozen contracts.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A, docs only
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A, API docs are English
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- `contentstudio-backend/app/Http/Controllers/Api/V1/SwaggerController.php` drives docs generation, so the new endpoints must be reachable from there.
- The metric glossary overlaps with the metric-definition audit in the Analytics widget and metric consistency epic. If that audit produces a canonical metric dictionary, this documentation should cite the same definitions rather than writing a second set.

---
---

# 4. [CLI] Add ads analytics commands to the CLI and the agent skill

### Description

As a developer or agent using the ContentStudio CLI, I want commands that pull Meta Ads and Google Ads performance, so I can script paid reporting or let an agent answer questions about ad spend.

### Workflow

1. A developer runs the CLI with their API key configured.
2. They list their connected ad accounts and see which platform each belongs to.
3. They pull a performance summary for one account over a date range.
4. They pull campaigns for that account, paginated, in structured output, and pipe it into their own tooling.
5. They pull keywords for a Google Ads account.
6. An agent using the ContentStudio skill can call the same commands to answer a question about ad performance.

### Acceptance criteria

- [ ] Commands exist to list connected ad accounts, following the CLI's existing command naming style, showing the platform for each.
- [ ] Commands exist to pull a performance summary, campaigns, ad sets or ad groups, and ads, for both platforms.
- [ ] Commands exist to pull keywords, search terms and conversion data for Google Ads.
- [ ] Commands exist to pull demographic breakdowns for both platforms.
- [ ] Date range and filter arguments match the API's parameter names.
- [ ] Paginated commands either page automatically or expose paging arguments, and the behaviour is documented rather than surprising.
- [ ] Default output is human-readable and shows monetary values with their currency, never a bare number.
- [ ] Structured output mode returns the API payload inside the CLI's standard result envelope.
- [ ] A workspace with no connected ad accounts produces a clear message, not an empty table.
- [ ] Requesting an ad account the key cannot access produces a clear permission message and a non-zero exit code.
- [ ] Authentication and rate-limit errors surface in the same wording style the CLI uses elsewhere.
- [ ] The commands appear in the CLI's help output and README.
- [ ] The agent skill definition lists the commands with an example of answering an ad-performance question.

### Mock-ups

N/A. Command-line output.

### Impact on existing data

None.

### Impact on other products

- CLI package and the agent skill repo. Feeds the ClawHub listing refresh.

### Dependencies

- Depends on both backend platform stories.
- Feeds **[ClawHub] Refresh the published ContentStudio skill with the ads analytics commands**.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A, CLI
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A, CLI is English
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- The CLI is the `contentstudio-cli` npm package, binary `contentstudio`, colon-style commands, a structured `--json` result envelope and a `--dry-run` flag. An `ads:` command family fits the existing shape, with the platform as an argument or as separate command families.
- Currency formatting in human-readable output is worth getting right once in a shared helper: several commands return money and a bare number is actively misleading when a customer runs multiple accounts in different currencies.

---
---

# 5. [MCP] Add Meta Ads and Google Ads analytics tools to the MCP server

### Description

As someone using ContentStudio through an MCP client, I want tools that read my ads performance, so I can ask the assistant how my campaigns are doing and get real numbers instead of a suggestion to go check the dashboard.

### Workflow

1. A user connects the ContentStudio MCP server with their API key.
2. They ask how their ad campaigns performed last month.
3. The assistant lists the connected ad accounts, then pulls the summary and campaigns for the relevant one.
4. The assistant reports the numbers with their currency.
5. The user asks a follow-up about which keywords cost the most, and the assistant pulls the keyword data for the Google Ads account.

### Acceptance criteria

- [ ] Tools are exposed for listing connected ad accounts, and for reading summary, campaigns, ad sets or ad groups, ads and demographics for both platforms.
- [ ] Tools are exposed for Google Ads keywords, search terms and conversions.
- [ ] Tool descriptions state that monetary values carry a currency and that the currency is per ad account, so the model does not compare or sum across currencies.
- [ ] Tool descriptions state each tool's date-range parameter clearly enough that the model does not default to an unbounded range.
- [ ] Paginated tools expose paging in a way the model can follow, and the description says so.
- [ ] Metric names in tool output are the documented public names, so the model's explanations match our docs.
- [ ] A workspace with no connected ad accounts returns an explicit empty result the model can relay.
- [ ] Permission, authentication and rate-limit failures return actionable messages rather than raw HTTP errors.
- [ ] Tools are documented in the MCP server's README and tool listing.
- [ ] Tools work in the packaged desktop bundle as well as the npx invocation.

### Mock-ups

N/A.

### Impact on existing data

None.

### Impact on other products

- MCP server package and its packaged desktop bundle. Reachable through the automation platforms that proxy MCP.

### Dependencies

- Depends on both backend platform stories.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- The MCP server is the `contentstudio-mcp` npm package, run via npx with the API key in the environment, also shipped as a desktop bundle. The existing account-fetching tool is the closest shape for workspace scoping and parameter handling.
- Tool count matters for MCP clients. Consider a smaller set of tools with a platform parameter rather than one tool per platform per section, which would roughly double the server's tool surface.

---
---

# 6. [Claude Extension] Surface ads analytics in the Claude extension

### Description

As a ContentStudio user working in Claude, I want to ask about my ad performance and get real numbers from my connected accounts, so I can pull a quick read on spend and campaign results without opening the dashboard.

### Workflow

1. A user has the ContentStudio Claude extension installed and connected with their API key.
2. They ask how their ads performed over a period.
3. The extension lists the connected ad accounts and pulls the relevant performance data.
4. The extension reports figures with their currency and names the metrics as our docs name them.
5. The user follows up about a specific campaign or keyword and the extension drills in.

### Acceptance criteria

- [ ] The extension can list connected ad accounts and read ads performance for both platforms.
- [ ] Monetary values are always reported with their currency, and figures from accounts in different currencies are never summed together.
- [ ] Date ranges are handled explicitly, so an unqualified question produces a stated range rather than an unbounded pull.
- [ ] A workspace with no connected ad accounts is reported plainly.
- [ ] Requests for an account the user cannot access are reported as a permission issue, not as missing data.
- [ ] Metric names match the documented public names.
- [ ] The extension's documented capability list mentions ads analytics.
- [ ] Authentication and rate-limit failures produce actionable messages.

### Mock-ups

N/A. Conversational surface.

### Impact on existing data

None.

### Impact on other products

- The Claude extension bundle only.

### Dependencies

- Depends on both backend platform stories.
- Best done after **[MCP] Add Meta Ads and Google Ads analytics tools to the MCP server**, since the extension is the packaged path to the same tools.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

# 7. [GPT App] Add ads analytics actions to the GPT app schema

### Description

As a ContentStudio user using our GPT app, I want to ask about my ad performance and get real figures, so I can review spend and results in the conversation.

### Workflow

1. A user opens the ContentStudio GPT app and authenticates.
2. They ask about their ad performance for a period.
3. The app lists ad accounts and calls the relevant analytics actions.
4. The app reports figures with their currency.
5. The user drills into a campaign or keyword and the app calls the corresponding action.

### Acceptance criteria

- [ ] Actions are added for listing ad accounts and reading summary, campaigns, ad sets or ad groups, ads and demographics for both platforms.
- [ ] Actions are added for Google Ads keywords and conversions.
- [ ] Response schemas are defined so the model reads metrics and currency reliably.
- [ ] Action descriptions state the currency contract and the date-range requirement.
- [ ] Paginated actions expose paging in a form the model can use.
- [ ] Empty results and permission failures are representable and distinguishable by the model.
- [ ] Actions respect the app's existing authentication flow.
- [ ] Rate-limit responses are handled so the app reports a retry rather than failing opaquely.
- [ ] The app's documented capabilities mention ads analytics.

### Mock-ups

N/A.

### Impact on existing data

None.

### Impact on other products

- The GPT app action schema only.

### Dependencies

- Depends on both backend platform stories.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- GPT app action schemas have practical size limits. If the full ads surface does not fit, prefer a smaller set of actions with platform and section parameters over dropping capabilities silently, and document what was consolidated.

---
---

# 8. [Zapier] Add ads analytics search steps

### Description

As a Zapier user, I want steps that pull ContentStudio ads analytics, so I can push spend and campaign performance into a spreadsheet, a Slack digest or a client report on a schedule.

### Workflow

1. A user builds a Zap and adds a ContentStudio ads step.
2. They pick a workspace and an ad account from dropdowns.
3. They set a date range and any filters.
4. They test the step and see the metrics in the sample output.
5. They map the fields into a spreadsheet row or a message.

### Acceptance criteria

- [ ] Search steps are available for reading an ads performance summary and for reading campaigns, for both platforms.
- [ ] A step is available for listing connected ad accounts, so account dropdowns elsewhere can be populated.
- [ ] Workspace and ad account are selected from dropdowns populated by the listing endpoint, not typed by hand.
- [ ] Date range and filters are inputs with sensible defaults.
- [ ] Output fields are flat enough to map directly into a spreadsheet row or message template.
- [ ] Monetary output fields include the currency as its own field, so a mapped row is never an ambiguous number.
- [ ] Paginated results are handled so a user is not silently given only the first page. If only the first page is returned, the step says so.
- [ ] A workspace with no connected ad accounts produces a clear step error.
- [ ] Permission, authentication and rate-limit errors surface in Zapier-friendly wording.
- [ ] Steps have sample data so they can be tested and mapped before real data exists.

### Mock-ups

N/A. Zapier's own UI.

### Impact on existing data

None.

### Impact on other products

- The ContentStudio Zapier app only.

### Dependencies

- Depends on both backend platform stories.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

# 9. [Make] Add ads analytics modules

### Description

As a Make.com user, I want modules that read ContentStudio ads analytics, so my scenario can route spend and campaign performance into reporting or alerting.

### Workflow

1. A user adds a ContentStudio ads module to a scenario.
2. They pick a workspace and an ad account from dropdowns.
3. They set a date range and filters.
4. They run the module once and inspect the output.
5. They map the fields downstream.

### Acceptance criteria

- [ ] Modules are available for listing ad accounts, reading an ads performance summary, and reading campaigns, for both platforms.
- [ ] Workspace and ad account are selected from dropdowns populated by the listing endpoint.
- [ ] Date range and filters are inputs with sensible defaults.
- [ ] Module output interfaces are defined so downstream mapping shows named fields rather than raw JSON.
- [ ] Currency is exposed as its own output field alongside every monetary value.
- [ ] Paginated list modules iterate properly, or clearly document that they return a single page.
- [ ] Empty results and permission failures produce clear, distinguishable module errors.
- [ ] Authentication and rate-limit errors surface in a form a Make user can act on.
- [ ] Modules are documented in the app's module list.

### Mock-ups

N/A. Make.com's own UI.

### Impact on existing data

None.

### Impact on other products

- The ContentStudio Make.com app only.

### Dependencies

- Depends on both backend platform stories.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

# 10. [n8n] Add ads analytics operations

### Description

As an n8n user, I want operations that read ContentStudio ads analytics, so my workflow can pull paid performance into whatever it reports into.

### Workflow

1. A user adds the ContentStudio node and picks an ads operation.
2. They select a workspace and an ad account.
3. They set a date range and filters.
4. They execute the node and inspect the output items.
5. They reference the fields in later nodes.

### Acceptance criteria

- [ ] Operations are available for listing ad accounts, reading an ads performance summary, and reading campaigns, for both platforms.
- [ ] Workspace and ad account are selectable from loaded options rather than typed identifiers.
- [ ] Date range and filters are parameters with sensible defaults.
- [ ] Output items are shaped so later nodes can reference metrics by name.
- [ ] Currency is present as its own field alongside every monetary value.
- [ ] Paginated operations either return all pages or expose paging explicitly, documented either way.
- [ ] Empty results and permission failures surface as clear, distinguishable node errors.
- [ ] Authentication and rate-limit errors surface in the node output.
- [ ] Operations are documented in the node's documentation.

### Mock-ups

N/A. n8n's own UI.

### Impact on existing data

None.

### Impact on other products

- The ContentStudio n8n node only.

### Dependencies

- Depends on both backend platform stories.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

# 11. [ClawHub] Refresh the published ContentStudio skill with the ads analytics commands

### Description

As someone installing ContentStudio from ClawHub, I want the published skill to include the ads analytics capability, so the agent I install it into can answer questions about my ad performance without me updating anything by hand.

### Workflow

1. A user installs the ContentStudio skill from ClawHub into their agent runtime.
2. They ask the agent about their ad performance.
3. The agent uses the ads commands from the installed skill and answers with real figures.
4. The published listing shows the current capability set, including ads analytics.

### Acceptance criteria

- [ ] The published ClawHub listing is updated to the skill version that includes the ads analytics commands.
- [ ] The listing's description and capability summary mention Meta Ads and Google Ads analytics.
- [ ] Installing from ClawHub into a supported agent runtime yields working ads commands end to end against a real workspace with a connected ad account.
- [ ] The listing's version pin matches the CLI and skill version that shipped the commands.
- [ ] The refresh follows the documented release process.

### Mock-ups

N/A.

### Impact on existing data

None. Distribution only.

### Impact on other products

- The ClawHub listing only. No code changes in the product repos.

### Dependencies

- Depends on **[CLI] Add ads analytics commands to the CLI and the agent skill**.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- If the best-time epic ships first, its ClawHub refresh will have established the release process. Reuse it rather than re-deriving it, and batch the two refreshes if the timing allows so ClawHub is republished once.
