# Epic + stories — Competitor analytics on the public API

**Priority: P2.** 3 stories. Nothing is pushed to any tracker.

---

## Epic: Competitor analytics on the public API

Competitor analytics is a paid, differentiating part of ContentStudio: a customer picks competitor pages on Facebook and Instagram, and gets a report comparing their own performance against them. It is exactly the kind of comparative data an agency puts in front of a client, and exactly the kind of repetitive reporting an agency would want to automate.

None of it is on the public API. A customer can read their own analytics programmatically across twelve surfaces, and cannot read a single competitor number. Searching for a competitor page, adding one to a report, reading the comparison, and removing one are all browser-only.

That has a second cost beyond the missing capability. Competitor data is one of the strongest reasons a customer would build a reporting integration at all, and today discovering it is unavailable happens after they have already built the rest.

This epic gives the API the competitor feature: finding competitors, managing competitor reports, and reading the comparison.

### Out of scope

- Adding competitor support for networks the product does not support. Competitor analytics covers Facebook and Instagram, and this epic exposes what exists rather than extending it.
- Changing how competitor data is collected or how often.

### Stories

1. `[BE] Add competitor search and competitor report management to the public API`
2. `[BE] Add competitor comparison reads to the public API`
3. `[BE] Expose competitor analytics on every developer surface`

---

## [BE] Add competitor search and competitor report management to the public API

### Description

As a developer building reporting automation for an agency, I want to find competitor pages and manage competitor reports through the API, so that setting up a client's competitive benchmarking is part of my provisioning flow rather than a manual browser task.

### Workflow

1. The developer searches for a competitor page by name or handle, for a given network.
2. The developer creates a competitor report, choosing their own account and the competitors to compare against.
3. The developer lists the workspace's competitor reports, and reads one.
4. The developer updates a report, adding or removing competitors.
5. The developer deletes a report.

### Acceptance criteria

- [ ] A caller can search for competitor pages on each supported network and receives enough detail to identify the right one, matching what the product's search returns.
- [ ] A caller can create a competitor report, specifying their own account and the competitors to compare against.
- [ ] A caller can list a workspace's competitor reports, read one by id, update it, and delete it.
- [ ] Updating a report can add and remove competitors.
- [ ] Validation matches the product, including the maximum number of competitors per report, with clear errors.
- [ ] Adding a competitor that cannot be tracked, for example a private or ineligible page, returns a clear reason rather than a generic failure or a silently empty report.
- [ ] Authorization uses the existing workspace permission model, and a caller can only build reports over accounts they have access to.
- [ ] Plan gating is respected. A workspace whose plan does not include competitor analytics receives a clear refusal, not an empty result.
- [ ] The story states plainly whether adding a competitor consumes any quota or credit, and if it does, that is surfaced to the caller before the call succeeds.
- [ ] Every endpoint appears in the OpenAPI document with examples.

### Mock-ups

None. Backend-only story.

### Impact on existing data

Competitor reports created through the API are stored the same way as product-created ones. Adding a competitor starts data collection for that page exactly as it does in the product, so the story must state what the initial data availability looks like.

### Impact on other products

Competitor reports created through the API appear in the product's analytics views. The story must verify they are indistinguishable there.

### Dependencies

- None for the endpoints.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (error messages must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Add competitor comparison reads to the public API

### Description

As a developer building a client dashboard or report, I want to read the competitor comparison data through the API, so that competitive benchmarking can appear in whatever my customer already looks at.

### Workflow

1. The developer reads a competitor report's comparison for a chosen date range.
2. The developer reads the per-competitor breakdown: audience, engagement, posting activity and the rest of what the product shows.
3. The developer reads engagement broken down by competitor.
4. The developer handles a competitor for which data is not yet available.

### Acceptance criteria

- [ ] A caller can read a competitor report's comparison data for a chosen date range.
- [ ] The response includes every metric the product's competitor view shows, for the caller's own account and for each competitor.
- [ ] A caller can read engagement broken down by competitor.
- [ ] Endpoint naming, date range parameters and response conventions match the existing analytics endpoints, so a caller integrating both learns one pattern.
- [ ] A competitor with no data yet, or a period before tracking began, returns an explicit "no data" state with the reason, never zeroes that read as a real measurement of nothing.
- [ ] Metric values match what the product shows for the same report and period, verified by comparison.
- [ ] Authorization uses the existing workspace permission model, and plan gating is respected.
- [ ] Every endpoint appears in the OpenAPI document with examples, and the documentation states what each metric means and where it comes from.

### Mock-ups

None. Backend-only story.

### Impact on existing data

None. Read-only.

### Impact on other products

None. Read-only, and the product's own views are unchanged.

### Dependencies

- Depends on `[BE] Add competitor search and competitor report management to the public API` for the reports it reads.
- Should follow the conventions established by the analytics endpoints already on the API, including the explicit no-data pattern the Bluesky capabilities endpoint set.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (error and no-data messages must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Expose competitor analytics on every developer surface

### Description

As a developer or agency using ContentStudio through the CLI, an AI assistant, or a no-code tool, I want competitor analytics available there too, so that competitive benchmarking can be part of the reporting I already automate.

### Workflow

1. The developer manages competitor reports and reads comparisons from the CLI.
2. An AI assistant, through the MCP server, can answer a question like how the customer compares with their competitors this month.
3. The developer builds a no-code automation that pulls competitor numbers into a customer's own dashboard.
4. The developer finds all of it in the public documentation.

### Acceptance criteria

- [ ] Competitor search, report management and comparison reads are available in the CLI.
- [ ] The agent skill covers the same capabilities.
- [ ] The MCP server exposes them as tools, including reading a comparison so an assistant can answer competitive questions directly.
- [ ] The ChatGPT and Claude connectors expose the same capabilities, inherited through the MCP server rather than defined separately.
- [ ] Zapier, Make and n8n expose at minimum reading a competitor comparison, so competitor numbers can be pushed into another system.
- [ ] Where a surface deliberately omits a capability, that omission is recorded with its reason.
- [ ] Every surface presents the no-data case honestly, so no surface turns an absence of competitor data into an apparent zero.
- [ ] The public documentation covers the capability on every surface, with a worked example of a competitive report.
- [ ] The parity check passes for competitor analytics across every surface.
- [ ] Every surface enforces the same authorization and plan gating as the REST endpoints.

### Mock-ups

None. No graphical UI in this story.

### Impact on existing data

None.

### Impact on other products

Every developer surface is touched. Existing capabilities on those surfaces must be unaffected.

### Dependencies

- Depends on both endpoint stories in this epic.
- Depends on the developer surface parity contract epic.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, no graphical UI in this story
- [ ] Multilingual support (surface-facing text and errors translated where supported)
- [ ] UI theming support — N/A, no graphical UI in this story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
