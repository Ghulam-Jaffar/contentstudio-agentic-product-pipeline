# Epic + stories — API foundations: webhooks, limits and documentation

**Priority: P0.** 4 stories. Small, independent, and every one of them removes a reason an integration has to fall back to the browser or to guesswork. Nothing is pushed to any tracker.

---

## Epic: API foundations: webhooks, limits and documentation

Three problems that are not missing features so much as missing foundations. Each is cheap. Each is currently costing every integration something.

**Webhooks cannot be managed through the API.** The public webhooks feature shipped and works: create an endpoint, rotate its secret, send a test event, read the delivery log, list the event types. All of it is behind the session login, on the browser's own authentication. None of it is on the public API. So the flagship feature for event-driven integrations can only be configured by a human logging in. An integration cannot provision its own webhook, cannot rotate its own secret, and cannot read its own delivery failures to find out why it stopped receiving events. For a customer on the API-centric plan that is a workflow break, not a missing convenience.

**No caller can read its own limits or consumption.** There is no endpoint answering how many posts a workspace can still publish, how many AI credits are left, what the rate limit is, or how much of the API allowance has been used. The product computes all of it already. An integration discovers its limits by hitting them, which means the first time a developer learns about a limit is a failed request in production.

**The published documentation has drifted from the code.** The OpenAPI document is missing capabilities that exist. The ten Bluesky analytics endpoints are entirely absent from it; the only Bluesky path documented is connecting an account. So a capability we built, including a thoughtful endpoint that reports exactly which metrics the network cannot provide and why, is invisible to every developer reading the docs. And because nothing checks, we do not know what else is missing.

### Relationship to the developer surface parity contract epic

That epic will eventually **generate** the OpenAPI document from a single capability definition source, which is the permanent fix. This epic does the tactical one: resync what is there now and guard it, so the drift stops today rather than when the larger work lands. The story here is deliberately written so that the parity contract epic supersedes it rather than duplicating it, and it says so.

### Out of scope

- API key creation. Bootstrapping the first credential has to happen somewhere trusted, and the browser is a reasonable answer.
- Changing rate limits, plan limits, or how webhooks are delivered.
- Building the usage record itself. That is the usage visibility epic. This epic exposes what already exists.

### Stories

1. `[BE] Put webhook management on the public API`
2. `[BE] Add a limits and consumption endpoint to the public API`
3. `[BE] Resync the OpenAPI document with the code and guard it`
4. `[BE] Expose webhooks and limits on every developer surface`

---

## [BE] Put webhook management on the public API

### Description

As a developer building an event-driven integration, I want to manage my webhooks through the API, so that I can provision an endpoint, rotate a secret and diagnose a delivery failure from my own system instead of asking someone to log into ContentStudio.

### Workflow

1. The developer lists the event types available to subscribe to.
2. The developer creates a webhook endpoint, choosing the events it receives, and receives its signing secret once.
3. The developer sends a test event to confirm the endpoint is reachable before relying on it.
4. The developer lists their webhooks, reads one, updates the events it subscribes to, disables it, or deletes it.
5. The developer rotates a secret without downtime.
6. When events stop arriving, the developer reads the delivery log for that webhook, sees which deliveries failed and why, and reads a single delivery in full.
7. The developer reads how much of their webhook delivery allowance has been used.

### Acceptance criteria

- [ ] A caller can list the available webhook event types, with a description of each and the shape of its payload.
- [ ] A caller can create a webhook endpoint, choosing which events it subscribes to, and the signing secret is returned exactly once at creation.
- [ ] A caller can list webhooks, read one, update it, disable and re-enable it, and delete it.
- [ ] A caller can rotate a webhook's signing secret, and the response makes clear how to roll over without missing deliveries.
- [ ] A caller can send a test event to a webhook and see the result.
- [ ] A caller can read a webhook's delivery log with each delivery's status, and read a single delivery in full including the response the endpoint returned.
- [ ] A caller can read how much of their webhook delivery allowance has been consumed and when it resets.
- [ ] Every capability available in the browser today is available through the API, and the browser surface continues to work unchanged.
- [ ] Authorization scopes webhooks to their owner exactly as the browser surface does. A caller cannot read, modify or delete a webhook they do not own, verified with a negative test.
- [ ] A secret is never returned again after creation or rotation, on any endpoint, including the read and list endpoints.
- [ ] Webhook creation and test sends are rate-limited, and the limits are documented, so the test facility cannot be used to send traffic at an arbitrary third party.
- [ ] A webhook endpoint URL is validated before it is accepted, and a URL that would target an internal address is refused.
- [ ] Every endpoint appears in the OpenAPI document with examples, including the signature verification a receiver must perform.

### Mock-ups

None. Backend-only story. The existing browser surface is unchanged.

### Impact on existing data

None. Webhooks created through the API are stored the same way as browser-created ones and appear in the browser surface. Existing webhooks are untouched and keep working.

### Impact on other products

The browser webhook management screens continue to work unchanged and must show webhooks created through the API. The story must verify both directions.

### Dependencies

- None. This is exposure of a shipped feature.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (error messages must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Add a limits and consumption endpoint to the public API

### Description

As a developer building on ContentStudio, I want to read my own limits and how much of each I have used, so that my integration can back off, warn, or queue gracefully instead of discovering a limit as a failed request in production.

### Workflow

1. The developer reads the limits and current consumption for the account or a workspace.
2. The response covers everything metered: posts, social accounts, workspaces, team members, media storage, AI credits of each kind, API calls, webhook deliveries, X posting, listening topics and mentions, automations.
3. Each figure comes with what is used, what the limit is, and when it resets.
4. The developer reads the API rate limit that applies to them and how much of the current window remains.
5. The developer's integration checks before a large batch and warns the customer rather than failing halfway.

### Acceptance criteria

- [ ] A caller can read their limits and current consumption in one call.
- [ ] Every metered thing is covered: posts, social accounts, workspaces, team members, media storage, AI text credits, AI image credits, AI video credits, video clip credits, AI auto-reply credits, API call allowance, webhook delivery allowance, X posting, social listening topics and mentions, and automations.
- [ ] Each entry states what has been used, what the limit is, and when it resets, so a caller never has to infer a reset date.
- [ ] The response includes any addon that raises a limit, so the figure reflects the customer's actual entitlement rather than their base plan.
- [ ] A caller can read the API rate limit applying to them and how much of the current window remains.
- [ ] Figures match exactly what the product shows for the same account and workspace, verified by comparison.
- [ ] The response distinguishes an unlimited entry from a very large one, so a caller does not have to treat a large number as unlimited.
- [ ] Where a limit does not apply to the caller's plan, that is stated explicitly rather than returned as zero, which would read as an exhausted allowance.
- [ ] Authorization uses the existing model. Account-level figures are only readable by a caller entitled to see them, and workspace-level figures only for workspaces the caller can access.
- [ ] No provider cost, margin or other internal figure appears in the response.
- [ ] The endpoint appears in the OpenAPI document with a fully worked example.
- [ ] The endpoint is cheap enough to be called before a batch without itself becoming a rate-limit problem, and any caching applied is documented so a caller knows how fresh the figures are.

### Mock-ups

None. Backend-only story.

### Impact on existing data

None. Read-only over figures the product already computes.

### Impact on other products

The product's own usage indicators must continue to show the same figures. If this endpoint and the product ever disagree, one of them is wrong, so the story must verify they agree rather than assume.

### Dependencies

- Overlaps the usage visibility epic, which is building a per-event usage record with a fuller picture. This endpoint must not become a second, divergent source of truth. If that epic lands first, this endpoint reads from what it establishes; if this ships first, that epic must reconcile against it. Whichever way round, one source.
- The media storage figure should agree with the storage read in the media library epic.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (any human-readable label or error must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review (a white-label customer's limits must reflect their own entitlement)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Resync the OpenAPI document with the code and guard it

### Description

As a developer reading the ContentStudio API documentation, I want it to describe everything the API actually does, so that I do not miss a capability that exists or build against one that does not.

### Workflow

1. The team compares the published API document against the routes that actually exist.
2. Every capability present in the code and absent from the document is documented.
3. Every capability present in the document and absent from the code is removed or corrected.
4. A check is added that fails when the two diverge again.
5. The check runs automatically, so the document cannot drift silently.

### Acceptance criteria

- [ ] A documented comparison lists every discrepancy found between the API's actual routes and the published document, in both directions.
- [ ] Every capability that exists in the code is present in the document, with request and response examples.
- [ ] The Bluesky analytics endpoints are documented, including the endpoint that reports which metrics the network cannot provide, and the reason each is unavailable.
- [ ] Anything documented but not implemented is removed, or corrected if it was a naming error.
- [ ] A check compares the published document against the actual routes and fails on any discrepancy, naming the route and the direction of the mismatch.
- [ ] The check runs automatically on every change rather than on request.
- [ ] The check is verified by deliberately adding an undocumented route and confirming it fails, then documenting it and confirming it passes.
- [ ] Deliberate exclusions, such as internal routes that should never be published, are recorded with a reason and pass the check.
- [ ] The check does not go live with a suppressed list of pre-existing failures. Every discrepancy is fixed or explicitly excluded first.
- [ ] The story records that the developer surface parity contract epic will later generate this document from the capability definition source, and this check is expected to be replaced by that rather than maintained alongside it.

### Mock-ups

None. Backend-only story.

### Impact on existing data

None.

### Impact on other products

The published documentation is what developers read. Correcting it may reveal capabilities that existing integrations were unaware of, which is the intent.

### Dependencies

- Superseded later by the developer surface parity contract epic, which generates the document rather than checking it. Deliberately built so that supersession is clean.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support — N/A, API reference documentation is in English, consistent with today
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Expose webhooks and limits on every developer surface

### Description

As a developer using ContentStudio through the CLI, an AI assistant, or a no-code tool, I want webhook management and my usage figures available there too, so that I can set up event delivery and check my headroom without switching tools.

### Workflow

1. The developer creates and inspects webhooks from the CLI, including reading a delivery log to diagnose a failure.
2. The developer checks their limits and consumption from the CLI before running a large batch.
3. An AI assistant, through the MCP server, can answer how much of an allowance is left and why a webhook stopped delivering.
4. The developer finds all of it in the public documentation.

### Acceptance criteria

- [ ] Webhook management, delivery log reads, and the limits and consumption read are available in the CLI.
- [ ] The agent skill covers the same capabilities.
- [ ] The MCP server exposes them as tools, including reading limits so an assistant can answer allowance questions directly.
- [ ] The ChatGPT and Claude connectors expose the same capabilities, inherited through the MCP server rather than defined separately.
- [ ] No surface ever displays or logs a webhook signing secret after creation.
- [ ] Deleting a webhook or rotating a secret through an AI assistant requires the explicit confirmation the MCP epic defines for state-changing actions, since either can silently stop a customer's event delivery.
- [ ] Zapier, Make and n8n expose the limits read at minimum. Whether they expose webhook management, given they have their own trigger mechanisms, is decided explicitly and recorded with its reason.
- [ ] The public documentation covers the capability on every surface, including a worked example of diagnosing a failed delivery.
- [ ] The parity check passes for webhooks and limits across every surface.
- [ ] Every surface enforces the same authorization as the REST endpoints, including that a caller can only reach their own webhooks.

### Mock-ups

None. No graphical UI in this story.

### Impact on existing data

None.

### Impact on other products

Every developer surface is touched. Existing capabilities on those surfaces must be unaffected.

### Dependencies

- Depends on `[BE] Put webhook management on the public API` and `[BE] Add a limits and consumption endpoint to the public API`.
- Depends on the developer surface parity contract epic.
- Depends on the MCP epic for how an assistant confirms a state-changing action.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, no graphical UI in this story
- [ ] Multilingual support (surface-facing text and errors translated where supported)
- [ ] UI theming support — N/A, no graphical UI in this story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
