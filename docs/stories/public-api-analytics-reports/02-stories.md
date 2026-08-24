# Epic + stories — Analytics reports and shareable links on the public API

**Priority: P1.** 4 stories. Nothing is pushed to any tracker.

---

## Epic: Analytics reports and shareable links on the public API

The public API's analytics read coverage is genuinely strong: twelve surfaces including Meta Ads and Google Ads, with real depth on each. What it cannot do is produce a report.

Everything on the output side is browser-only. An integration can read every metric that goes into a report and cannot ask for the report itself. It cannot generate one, list what has been generated, download one, retry a failed one, schedule a recurring one, or create the shareable link that lets a client view a report without a ContentStudio login.

That is the wrong way round for the customers most likely to use the API. An agency's monthly reporting is the most repetitive, most schedulable, most obviously automatable thing they do, and it is the one part of analytics the API cannot touch. Today they either rebuild the report themselves from the raw metric endpoints, losing the branding and layout the product gives them for free, or they log in every month and click.

Shareable links matter for the same reason. The analytics service already honours a shareable link as an authentication path, so the mechanism exists. What is missing is any way to create or manage one programmatically, which is exactly what an agency delivering reports to twenty clients would want.

This epic gives the API the output side of analytics: generate, schedule, share, and the account preferences that shape what a report contains.

### Out of scope

- Cross-workspace and organization-level reports. Explicitly not wanted.
- Changing what a report contains or how it looks.
- Analytics job triggers, which force a data refresh before reading. Not triaged either way. If wanted, they are one more story here.

### Stories

1. `[BE] Add analytics report generation and download to the public API`
2. `[BE] Add scheduled analytics reports to the public API`
3. `[BE] Add shareable analytics report links and account preferences to the public API`
4. `[BE] Expose analytics reports on every developer surface`

---

## [BE] Add analytics report generation and download to the public API

### Description

As a developer building reporting automation, I want to generate, list and download ContentStudio analytics reports through the API, so that my customer's monthly reporting runs without anyone logging in, and keeps the branding and layout the product already produces.

### Workflow

1. The developer asks what report options are available: the report types, the sections each can contain, and the accounts and date ranges it accepts.
2. The developer requests a report, choosing its type, its accounts, its date range and its sections.
3. Because report generation takes time, the request returns immediately with something the developer can poll or be notified about.
4. The developer checks the report's state and, once ready, downloads it.
5. If a report failed, the developer retries it without rebuilding the request.
6. The developer lists previously generated reports, and deletes ones no longer needed.

### Acceptance criteria

- [ ] A caller can read the available report options: report types, the sections each supports, and the accounts and date ranges accepted.
- [ ] A caller can request a report by type, accounts, date range and sections.
- [ ] Report generation follows the asynchronous job contract the public API already uses for AI video generation, rather than inventing a second async pattern. A caller receives a job reference immediately.
- [ ] A caller can read a report job's state, distinguishing queued, generating, ready and failed, with a reason on failure.
- [ ] A caller can download a completed report.
- [ ] A caller can retry a failed report without resubmitting the full request.
- [ ] A caller can list previously generated reports with their state and creation time, and can delete one.
- [ ] A report generated through the API is identical to the same report generated in the browser, including branding, sections and layout, verified by comparing both.
- [ ] Authorization uses the existing workspace permission model, and a caller can only reference accounts they have access to.
- [ ] Plan gating is respected. A workspace whose plan does not include reporting receives a clear refusal.
- [ ] Requesting a report for a period with no data returns a report that states there is no data, matching product behavior, rather than an error or an empty file.
- [ ] Report generation through the API is rate-limited, and the limit is documented, so a loop cannot exhaust the report generator.
- [ ] Every endpoint appears in the OpenAPI document with examples, including the full async flow.

### Mock-ups

None. Backend-only story. The report itself is existing product output and is not redesigned.

### Impact on existing data

Reports generated through the API are stored the same way as reports generated in the product and count against the same retention. No existing report data is migrated or changed.

### Impact on other products

Reports generated through the API appear in the product's report list. The story must verify they are indistinguishable there.

### Dependencies

- Reuses the existing asynchronous job contract from the AI video generation work rather than defining a new one.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (report content and error messages must respect the workspace locale as they do in the product)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review (reports carry white-label branding, and an API-generated report must carry the same branding as a browser-generated one)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Add scheduled analytics reports to the public API

### Description

As a developer setting up a customer's recurring reporting, I want to create and manage scheduled reports through the API, so that a monthly client report can be provisioned as part of onboarding and never touched again.

### Workflow

1. The developer creates a schedule: which report, for which accounts, on what cadence, sent to which recipients.
2. The developer lists the schedules on a workspace, and reads one.
3. The developer updates a schedule, pauses it, or deletes it.
4. The developer triggers a scheduled report to send once, immediately, without waiting for its next run.
5. The developer sees when a schedule last ran and when it runs next.

### Acceptance criteria

- [ ] A caller can create a scheduled report, specifying the report type, accounts, cadence, and recipients.
- [ ] A caller can list a workspace's scheduled reports, read one, update it, and delete it.
- [ ] A caller can pause and resume a schedule without deleting it.
- [ ] A caller can trigger a scheduled report to send immediately, and the send is the same one the schedule would perform.
- [ ] Reading a schedule returns when it last ran, whether that run succeeded, and when it runs next.
- [ ] Validation matches the product on cadence options, recipient limits and account selection, with clear errors.
- [ ] Recipients that are not ContentStudio users are handled exactly as the product handles them.
- [ ] Authorization uses the existing permission model, and a caller can only schedule reports over accounts they have access to.
- [ ] Plan gating is respected.
- [ ] Manual triggering is rate-limited so it cannot be used to send repeatedly to a recipient, and the limit is documented.
- [ ] A schedule created through the API is indistinguishable from one created in the product, verified in the product UI.
- [ ] Every endpoint appears in the OpenAPI document with examples.

### Mock-ups

None. Backend-only story.

### Impact on existing data

Schedules created through the API are stored the same way as product-created ones. No migration.

### Impact on other products

Scheduled reports created through the API appear in the product's schedule list and send the same email. The story must verify both.

### Dependencies

- Depends on `[BE] Add analytics report generation and download to the public API` for the report types and options it schedules.
- The scheduled send uses the existing email path. Coordinates with the notifications architecture epic so this does not become a second send path.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (the scheduled email and its report must respect the recipient locale as they do today)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review (scheduled report emails carry white-label branding and sender identity)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Add shareable analytics report links and account preferences to the public API

### Description

As a developer delivering reports to a customer's clients, I want to create and manage shareable analytics links through the API, and set the analytics account preferences that shape what a report shows, so that a client can view their report without a ContentStudio login and the report reflects the settings the customer expects.

### Workflow

1. The developer creates a shareable analytics link for a set of accounts and a report configuration.
2. The developer optionally protects the link with a password and gives it an expiry.
3. The developer lists the workspace's shareable links, reads one, updates it, disables it, or deletes it.
4. A client opens the link and sees the analytics, with no ContentStudio account, exactly as they would from a link created in the browser.
5. Separately, the developer reads and sets the analytics preferences on an account, so reports and dashboards reflect them.

### Acceptance criteria

- [ ] A caller can create a shareable analytics link for a chosen set of accounts and configuration.
- [ ] A caller can optionally set a password and an expiry on a link.
- [ ] A caller can list a workspace's shareable links, read one, update it, disable it, and delete it.
- [ ] A link created through the API works exactly as one created in the browser, verified by opening it as an anonymous visitor.
- [ ] Disabling or deleting a link stops it working immediately, verified.
- [ ] A link grants read access only, and only to the accounts and configuration it was created for. It cannot be used to reach any other workspace, account or capability, and this is verified explicitly with a negative test.
- [ ] A caller can read and set the analytics preferences on an account, matching what the product allows.
- [ ] Authorization uses the existing permission model. A caller who cannot create share links in the product cannot create them through the API.
- [ ] Plan gating is respected.
- [ ] Every endpoint appears in the OpenAPI document with examples, and the documentation states plainly what a shareable link does and does not grant.

### Mock-ups

None. Backend-only story. The shared view is existing product output.

### Impact on existing data

Links created through the API are stored the same way as product-created ones. Existing links are unaffected.

### Impact on other products

Shareable links are served under white-label domains and are viewed outside the app entirely. The story must verify an API-created link renders correctly under a white-label domain.

### Dependencies

- Depends on `[BE] Add analytics report generation and download to the public API` for the report configuration a link points at.
- Overlaps the service authorization epic, which is separately assessing how shareable links authenticate against the analytics service. This story must not weaken that path, and should reuse whatever it concludes.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (the shared view is opened by clients on phones and must work there, unchanged from today)
- [ ] Multilingual support (the shared view and any password prompt must respect locale as they do today)
- [ ] UI theming support (default + white-label)
- [ ] White-label domains impact review (shared links are the most white-label-visible surface in the product)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Expose analytics reports on every developer surface

### Description

As a developer or agency using ContentStudio through the CLI, an AI assistant, or a no-code tool, I want report generation, scheduling and sharing available there too, so that reporting is part of the automation I already run rather than a separate manual step every month.

### Workflow

1. The developer generates, lists and downloads reports from the CLI.
2. An AI assistant, through the MCP server, can generate a report and hand the user a shareable link.
3. The developer builds a no-code automation that generates a report on a schedule and delivers it to a customer's own system.
4. The developer finds all of it in the public documentation.

### Acceptance criteria

- [ ] Report generation, listing, download, retry, scheduling and shareable link management are available in the CLI.
- [ ] The agent skill covers the same capabilities.
- [ ] The MCP server exposes them as tools, including generating a report and returning a shareable link.
- [ ] The ChatGPT and Claude connectors expose the same capabilities, inherited through the MCP server rather than defined separately.
- [ ] The asynchronous nature of report generation is handled correctly on every surface: no surface blocks indefinitely, and each has a documented way to learn a report is ready.
- [ ] Zapier, Make and n8n expose at minimum generating a report, being triggered when a report is ready, and creating a shareable link.
- [ ] Report delivery on each surface makes clear that a shareable link grants access to whoever holds it, so a developer does not distribute one unknowingly.
- [ ] Where a surface deliberately omits a capability, that omission is recorded with its reason.
- [ ] The public documentation covers the capability on every surface, with a worked example of monthly client reporting end to end.
- [ ] The parity check passes for analytics reports across every surface.
- [ ] Every surface enforces the same authorization and plan gating as the REST endpoints.

### Mock-ups

None. No graphical UI in this story.

### Impact on existing data

None.

### Impact on other products

Every developer surface is touched. Existing capabilities on those surfaces must be unaffected.

### Dependencies

- Depends on the three endpoint stories in this epic.
- Depends on the developer surface parity contract epic.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, no graphical UI in this story
- [ ] Multilingual support (surface-facing text and errors translated where supported)
- [ ] UI theming support — N/A, no graphical UI in this story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
