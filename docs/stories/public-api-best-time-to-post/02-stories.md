# Epic: Best Time to Post across the public developer surfaces

## Problem

ContentStudio computes Best Time to Post for connected accounts and uses it in the Composer to recommend and auto-apply posting slots. None of it is available through the public API. A customer automating their publishing with our API can schedule posts but cannot ask us when to schedule them, so they either post at arbitrary times or rebuild the recommendation themselves from raw analytics and get a different answer than the one our own UI shows.

There is no best-time route anywhere in the public API. The recommendation currently exists only as internal analytics queries feeding the Composer scheduling UI.

## Goal

Expose Best Time to Post per connected social account through the public API, with an explicit and documented response contract, then propagate it to every developer surface we ship so it is reachable however the customer is integrating.

## Scope (surfaces)

Backend public API first, then: **API docs and OpenAPI spec, CLI (`contentstudio-cli`) plus the agent skill, MCP (`contentstudio-mcp`), the Claude extension, the GPT app, Zapier, Make.com, n8n, and the published ClawHub skill.**

## Rules

- **One recommendation source.** The public API must return the same recommendation the Composer shows for the same account and period. If they can differ, that is a bug, not a variation.
- **Be explicit about network support.** Best-time queries exist for Facebook, Instagram and LinkedIn today. Any network without one must return a clear unsupported response, never a silently empty result that reads like "no good times to post".
- **Be explicit about insufficient data.** The internal path already tracks whether there was enough data to answer. The public contract must surface that rather than returning an empty list.
- **Don't leak internal scoring.** Expose ranked slots and a documented confidence, not a raw internal metric that we cannot change later without breaking customers.
- **Reuse the established public analytics pattern:** API key auth, the shared error schemas, and the route and naming conventions the per-network analytics endpoints already use.
- Keep the OpenAPI spec, the API guide, and the agent `SKILL.md` in sync.

## Sequencing

Story 1 defines the contract and must land first. Story 2 closes out the documented contract. Stories 3 through 11 depend on the contract being frozen. Order within the surfaces: CLI and MCP first, since we own them and they feed the agent skill and the ClawHub listing, then the Claude extension and GPT app, then Zapier, Make.com and n8n, and ClawHub last since it republishes what the CLI and skill expose.

## Out of scope

- Changing how the recommendation is computed, or adding networks to the underlying analytics. If a network has no best-time query today, adding one is separate work.
- Writing scheduling decisions back automatically. This epic exposes the recommendation; acting on it stays the caller's choice.
- Bringing the recommendation into the mobile app.

## Stories

1. `[BE] Expose Best Time to Post per connected account in the public API`
2. `[BE] Document Best Time to Post in the OpenAPI spec and API guide`
3. `[CLI] Add best posting times to the CLI and the agent skill`
4. `[MCP] Add a best-time-to-post tool to the MCP server`
5. `[Claude Extension] Surface best posting times in the Claude extension`
6. `[GPT App] Add a best-time-to-post action to the GPT app schema`
7. `[Zapier] Add a best-time-to-post search step`
8. `[Make] Add a best-time-to-post module`
9. `[n8n] Add a best-time-to-post operation`
10. `[ClawHub] Refresh the published ContentStudio skill with the best-time commands`

---
---

# 1. [BE] Expose Best Time to Post per connected account in the public API

### Description

As a ContentStudio customer automating publishing through the API, I want to ask when a given connected account should post, so my automation can schedule at the times our own analytics say perform best instead of guessing or rebuilding the calculation myself.

### Workflow

*(Public API change. The "user" here is a developer calling the API with their API key.)*

1. A developer authenticates to the public API with their API key.
2. They call the best-times endpoint for a workspace, naming a connected account.
3. They optionally pass a period or date range and a timezone.
4. They receive the recommended posting slots for that account, ranked, with a confidence indicator and the timezone the slots are expressed in.
5. If the account's network has no best-time support, they receive an explicit unsupported response naming the network.
6. If there is not enough history to answer, they receive an explicit insufficient-data response rather than an empty list.
7. The endpoint is discoverable in the public API reference.

### Acceptance criteria

- [ ] A public API endpoint returns Best Time to Post for a single connected social account within a workspace.
- [ ] The endpoint authenticates with the customer's API key, using the same mechanism and header as the existing public analytics endpoints.
- [ ] The response returns ranked recommended posting slots, each with a day and a time, plus a documented confidence value whose meaning is stated in the docs.
- [ ] The response states the timezone the slots are expressed in, and honours a caller-supplied timezone when one is passed.
- [ ] The response includes the full hour-by-day distribution when the caller asks for it, so a developer can build their own visualisation, and omits it by default to keep the response small.
- [ ] The values returned match what the Composer's Best Time to Post shows for the same account, period and timezone. No discrepancy.
- [ ] Calling for an account on a network that has no best-time support returns a documented unsupported-network response naming the network, with a success status rather than an error.
- [ ] Calling for an account with too little history returns a documented insufficient-data response, distinguishable from both an unsupported network and an empty result.
- [ ] Period or date range parameters follow the same conventions as the existing public analytics endpoints.
- [ ] Requesting an account the API key's workspace does not own returns a permission error and reveals nothing about the account.
- [ ] Standard error responses match the existing analytics endpoints: 401 unauthorized, 403 workspace permission denied, 422 validation error, 429 rate limited, 502 upstream analytics error.
- [ ] Rate limiting is applied consistently with the other public API analytics endpoints.
- [ ] The endpoint is covered by the public API's per-plan credit accounting the same way other analytics endpoints are.
- [ ] The response does not expose internal scoring metric names or query internals.

### Mock-ups

N/A. Public API endpoint, documented via OpenAPI.

### Impact on existing data

None. Read-only over analytics data that is already computed. No new storage.

### Impact on other products

- Public API: new endpoint for external consumers.
- Web app: unaffected, the Composer keeps using its internal path. If both are refactored onto one shared service, the Composer's behaviour must not change.
- Mobile apps and Chrome extension: unaffected.
- Every downstream surface in this epic builds on this contract.

### Dependencies

None. This story defines the contract everything else waits on.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A, backend only
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A, no UI
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- The computation already exists. `getTimeRecommendationQuery($period)` lives in `contentstudio-backend/app/Builders/Analytics/Analyze/FacebookBuilder.php`, `InstagramBuilder.php` and `LinkedinBuilder.php`. Those three are the only builders with it, which is where the supported-network list comes from.
- `contentstudio-backend/app/Http/Controllers/Analytics/Analyze/FacebookController.php::getTimeRecommendation()` is the thinnest existing call path: builder query, ClickHouse read, return rows.
- `contentstudio-backend/app/Http/Controllers/Analytics/Analyze/OverviewController.php::mergeTimeResponseData()` already does the reshaping this endpoint needs: it turns the raw rows into an hour-by-day matrix, applies the workspace timezone offset, derives the top three slots, and sets a flag for whether there was enough data. That flag is the natural basis for the insufficient-data response.
- The public API pattern to mirror is `contentstudio-backend/app/Http/Controllers/Api/V1/Analytics/FacebookAnalyticsController.php`: per-section endpoints, `ApiKeyHeader` security, `@OA` annotations, and the shared error schemas in `AnalyticsApiSchemas.php`. `BaseApiController.php` is the base for API-key endpoints.
- Route group goes in `contentstudio-backend/routes/api/v1.php` alongside the existing `workspaces/{workspace_id}/analytics/{network}` groups. Grepping the routes directory for "best" currently returns nothing, so the name is free.
- Worth deciding up front whether the endpoint is per-account only or also accepts a set of accounts. `OverviewController` already merges across platforms, so a multi-account form is available if wanted, and retrofitting it later is a breaking-ish change to the response shape.

---
---

# 2. [BE] Document Best Time to Post in the OpenAPI spec and API guide

### Description

As a developer evaluating whether to use ContentStudio's API for scheduling, I want the best-times endpoint fully documented, so I understand which networks support it, what the confidence value means, how timezones are handled, and what I get back when there isn't enough data, without having to test my way to the answer.

### Workflow

1. A developer opens the public API reference.
2. They find the best-times endpoint with its parameters, response schema and examples.
3. The docs state which networks are supported and what happens for the rest.
4. The docs state what the confidence value means and how the slots are ranked.
5. The docs state how timezones are resolved when the caller passes one and when they don't.
6. The agent skill definition reflects the same capability, so an agent using the skill knows the command exists.

### Acceptance criteria

- [ ] The endpoint appears in the generated OpenAPI spec with full parameter and response schemas.
- [ ] The API guide documents supported networks explicitly, and documents the unsupported-network response.
- [ ] The docs define the confidence value's meaning and range in plain terms, without referring to internal metrics.
- [ ] The docs explain timezone resolution for both the caller-supplied and default cases.
- [ ] The docs document the insufficient-data response and state the conditions that produce it.
- [ ] A worked request and response example is included, including one showing the optional full distribution.
- [ ] Rate limits and credit consumption for this endpoint are documented alongside the other analytics endpoints.
- [ ] The agent `SKILL.md` is updated so an agent using the skill knows best posting times are available and how to request them.
- [ ] The public API changelog notes the addition.

### Mock-ups

N/A.

### Impact on existing data

None.

### Impact on other products

- Public API documentation and the agent skill definition. No runtime behaviour changes.

### Dependencies

- Depends on **[BE] Expose Best Time to Post per connected account in the public API** for the frozen contract.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A, docs only
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A, API docs are English
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- `contentstudio-backend/app/Http/Controllers/Api/V1/SwaggerController.php` drives docs generation, so the new endpoint has to be reachable from there to show up in the reference.
- `AnalyticsApiSchemas.php` holds the shared error schemas already documented for the analytics endpoints; reusing them keeps the error docs consistent.

---
---

# 3. [CLI] Add best posting times to the CLI and the agent skill

### Description

As a developer or an agent using the ContentStudio CLI, I want a command that tells me the best times to post for a connected account, so I can pipe it into a scheduling script or let an agent pick a slot before creating a post.

### Workflow

1. A developer runs the CLI with their API key configured.
2. They list their connected accounts.
3. They run the best-times command for one of those accounts, optionally passing a period and a timezone.
4. They see the recommended slots in readable output, or structured output when they ask for it.
5. If the account's network isn't supported, the CLI says so plainly and exits without a stack trace.
6. An agent using the ContentStudio skill can call the same command and read the structured output.

### Acceptance criteria

- [ ] A best-times command exists following the CLI's existing command naming style.
- [ ] The command accepts an account, and optional period or date range and timezone arguments, matching the API's parameter names.
- [ ] Default output is human-readable: the ranked slots with day, time and confidence.
- [ ] Structured output mode returns the API payload inside the CLI's standard result envelope.
- [ ] An unsupported network produces a clear message naming the network and a non-zero exit code, not a raw API error dump.
- [ ] Insufficient data produces a clear message distinct from the unsupported-network message.
- [ ] Authentication, rate-limit and permission errors surface with the same wording style the CLI already uses for other commands.
- [ ] The command appears in the CLI's help output and in its README.
- [ ] The agent skill definition lists the command so an agent knows it exists, with an example of using it before creating a scheduled post.

### Mock-ups

N/A. Command-line output.

### Impact on existing data

None.

### Impact on other products

- CLI package and the agent skill repo. Feeds the ClawHub listing refresh.

### Dependencies

- Depends on **[BE] Expose Best Time to Post per connected account in the public API**.
- Feeds **[ClawHub] Refresh the published ContentStudio skill with the best-time commands**.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A, CLI
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A, CLI is English
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- The CLI is the `contentstudio-cli` npm package, binary `contentstudio`, with colon-style commands (`posts:create`, `accounts:list`), a structured `--json` result envelope and a `--dry-run` flag. A name in the `accounts:` or `analytics:` family fits the existing shape.
- The agent skill lives in the standalone skill repo installed via the `skills add` flow. Its `SKILL.md` declares the binary and the API key environment variable, and is the file that needs the new command listed.

---
---

# 4. [MCP] Add a best-time-to-post tool to the MCP server

### Description

As someone using ContentStudio through an MCP client, I want a tool that returns the best posting times for a connected account, so the assistant can choose a good slot when I ask it to schedule something rather than defaulting to now.

### Workflow

1. A user connects the ContentStudio MCP server with their API key.
2. They ask the assistant when they should post on one of their accounts.
3. The assistant calls the best-times tool with that account.
4. The assistant reports the recommended slots and, if asked, creates a scheduled post at one of them using the existing create-post tool.
5. If the network isn't supported, the tool returns that plainly so the assistant can say so instead of inventing an answer.

### Acceptance criteria

- [ ] A best-time-to-post tool is exposed by the MCP server, named and described consistently with the existing tools.
- [ ] The tool accepts an account identifier and optional period or date range and timezone.
- [ ] The tool description tells the model what the confidence value means, so it does not present a low-confidence slot as a firm recommendation.
- [ ] The tool returns ranked slots in a structure the model can read without further parsing.
- [ ] An unsupported network returns an explicit result the model can relay, not an error the model has to guess at.
- [ ] Insufficient data returns an explicit result distinct from unsupported network.
- [ ] Permission, authentication and rate-limit failures return actionable messages rather than raw HTTP errors.
- [ ] The tool is documented in the MCP server's README and tool listing.
- [ ] The tool works in the packaged desktop bundle as well as the npx invocation.

### Mock-ups

N/A.

### Impact on existing data

None.

### Impact on other products

- MCP server package and its packaged desktop bundle. Reachable through the automation platforms that proxy MCP.

### Dependencies

- Depends on **[BE] Expose Best Time to Post per connected account in the public API**.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- The MCP server is the `contentstudio-mcp` npm package, run via npx with the API key in the environment, and also shipped as a desktop bundle. The existing account-fetching tool is the closest shape to copy for parameter handling and workspace scoping.
- Tool descriptions are the only thing steering the model's use of the tool, so the confidence-value explanation belongs in the description text, not only in the docs.

---
---

# 5. [Claude Extension] Surface best posting times in the Claude extension

### Description

As a ContentStudio user working inside Claude, I want to ask for the best times to post on one of my accounts and get our recommendation, so I can schedule from the conversation without switching to the app to check.

### Workflow

1. A user has the ContentStudio Claude extension installed and connected with their API key.
2. They ask when they should post on a connected account.
3. The extension retrieves the recommended slots and reports them.
4. The user asks to schedule a post at one of those slots, and the extension creates it.
5. If the network isn't supported, the extension says which networks do support it.

### Acceptance criteria

- [ ] The extension can retrieve best posting times for a connected account.
- [ ] Optional period or date range and timezone are supported.
- [ ] The extension reports the ranked slots along with what the confidence means, so a weak recommendation is not presented as a strong one.
- [ ] An unsupported network is reported plainly, naming which networks are supported.
- [ ] Insufficient data is reported plainly and distinctly from an unsupported network.
- [ ] The user can go from a returned slot to creating a scheduled post in the same conversation.
- [ ] The extension's documented capability list mentions best posting times.
- [ ] Authentication and permission failures produce actionable messages.

### Mock-ups

N/A. Conversational surface.

### Impact on existing data

None.

### Impact on other products

- The Claude extension bundle only.

### Dependencies

- Depends on **[BE] Expose Best Time to Post per connected account in the public API**.
- Best done after **[MCP] Add a best-time-to-post tool to the MCP server**, since the extension is the packaged path to the same tools.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

# 6. [GPT App] Add a best-time-to-post action to the GPT app schema

### Description

As a ContentStudio user using our GPT app, I want to ask for the best posting times for an account, so I can pick a slot and schedule without leaving the conversation.

### Workflow

1. A user opens the ContentStudio GPT app and authenticates.
2. They ask for the best times to post on a connected account.
3. The app calls the best-times action and reports the ranked slots.
4. The user asks to schedule at one of those times and the app creates the post.
5. If the network isn't supported, the app says so and names the supported networks.

### Acceptance criteria

- [ ] A best-time-to-post action is added to the GPT app's action schema.
- [ ] The action accepts an account and optional period or date range and timezone.
- [ ] The action's description tells the model what the confidence value means.
- [ ] The response schema is defined so the model reliably reads the ranked slots.
- [ ] Unsupported network and insufficient data are both representable in the response and distinguishable by the model.
- [ ] The action respects the app's existing authentication flow.
- [ ] Rate-limit responses are handled so the app reports a retry rather than failing opaquely.
- [ ] The app's documented capabilities mention best posting times.

### Mock-ups

N/A.

### Impact on existing data

None.

### Impact on other products

- The GPT app action schema only.

### Dependencies

- Depends on **[BE] Expose Best Time to Post per connected account in the public API**.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

# 7. [Zapier] Add a best-time-to-post search step

### Description

As a Zapier user automating ContentStudio, I want a step that looks up the best posting times for an account, so my Zap can schedule a post at a recommended time instead of a hardcoded one.

### Workflow

1. A user builds a Zap and adds the ContentStudio best-times step.
2. They pick a workspace and a connected account from dropdowns.
3. They optionally set a period and a timezone.
4. They test the step and see the ranked slots in the sample output.
5. They map a returned slot into a later ContentStudio create-post step's scheduled time.

### Acceptance criteria

- [ ] A best-times search step is available in the ContentStudio Zapier app.
- [ ] Workspace and account are selected from dropdowns populated by the existing list endpoints, not typed by hand.
- [ ] Period or date range and timezone are optional inputs with sensible defaults.
- [ ] Output fields are flat enough to map directly into a later create-post step's scheduled time field.
- [ ] The top recommended slot is available as its own output field so simple Zaps do not have to index into a list.
- [ ] An unsupported network produces a clear step error naming the network, in Zapier-friendly wording.
- [ ] Insufficient data produces a clear step error distinct from unsupported network.
- [ ] Authentication and rate-limit errors surface in a form a Zapier user can act on.
- [ ] The step has sample data so it can be tested and mapped before the account has real history.

### Mock-ups

N/A. Zapier's own UI.

### Impact on existing data

None.

### Impact on other products

- The ContentStudio Zapier app only.

### Dependencies

- Depends on **[BE] Expose Best Time to Post per connected account in the public API**.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

# 8. [Make] Add a best-time-to-post module

### Description

As a Make.com user automating ContentStudio, I want a module that returns the best posting times for an account, so my scenario can schedule at a recommended time.

### Workflow

1. A user adds the ContentStudio best-times module to a scenario.
2. They pick a workspace and a connected account from dropdowns.
3. They optionally set a period and a timezone.
4. They run the module once and inspect the ranked slots.
5. They map a returned slot into a later create-post module's scheduled time.

### Acceptance criteria

- [ ] A best-times module is available in the ContentStudio Make.com app.
- [ ] Workspace and account are selected from dropdowns populated by the existing list endpoints.
- [ ] Period or date range and timezone are optional inputs.
- [ ] The module's output interface is defined so downstream mapping shows named fields rather than raw JSON.
- [ ] The top recommended slot is exposed as its own output field.
- [ ] An unsupported network and insufficient data both produce clear, distinguishable module errors.
- [ ] Authentication and rate-limit errors surface in a form a Make user can act on.
- [ ] The module is documented in the app's module list.

### Mock-ups

N/A. Make.com's own UI.

### Impact on existing data

None.

### Impact on other products

- The ContentStudio Make.com app only.

### Dependencies

- Depends on **[BE] Expose Best Time to Post per connected account in the public API**.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

# 9. [n8n] Add a best-time-to-post operation

### Description

As an n8n user automating ContentStudio, I want an operation that returns the best posting times for an account, so my workflow can schedule at a recommended time.

### Workflow

1. A user adds the ContentStudio node to a workflow and picks the best-times operation.
2. They select a workspace and a connected account.
3. They optionally set a period and a timezone.
4. They execute the node and see the ranked slots in the output panel.
5. They reference a returned slot in a later create-post operation.

### Acceptance criteria

- [ ] A best-times operation is available on the ContentStudio n8n node.
- [ ] Workspace and account are selectable from loaded options rather than typed identifiers.
- [ ] Period or date range and timezone are optional parameters.
- [ ] Output items are shaped so a later node can reference the recommended time by name.
- [ ] The top recommended slot is available without indexing into a list.
- [ ] An unsupported network and insufficient data both surface as clear, distinguishable node errors.
- [ ] Authentication and rate-limit errors surface in the node output.
- [ ] The operation is documented in the node's documentation.

### Mock-ups

N/A. n8n's own UI.

### Impact on existing data

None.

### Impact on other products

- The ContentStudio n8n node only.

### Dependencies

- Depends on **[BE] Expose Best Time to Post per connected account in the public API**.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

# 10. [ClawHub] Refresh the published ContentStudio skill with the best-time commands

### Description

As someone installing ContentStudio from ClawHub, I want the published skill to include the best-posting-times capability, so the agent I install it into can use it without me having to update anything manually.

### Workflow

1. A user installs the ContentStudio skill from ClawHub into their agent runtime.
2. They ask the agent when they should post on a connected account.
3. The agent uses the best-times command from the installed skill and answers.
4. The published listing shows the current capability set, including best posting times.

### Acceptance criteria

- [ ] The published ClawHub listing is updated to the skill version that includes the best-times command.
- [ ] The listing's description and capability summary mention best posting times.
- [ ] Installing from ClawHub into a supported agent runtime yields a working best-times command end to end against a real workspace.
- [ ] The listing's version pin matches the CLI and skill version that shipped the command.
- [ ] The refresh follows the documented release process, so the listing does not drift from the skill again.

### Mock-ups

N/A.

### Impact on existing data

None. Distribution only.

### Impact on other products

- The ClawHub listing only. No code changes in the product repos.

### Dependencies

- Depends on **[CLI] Add best posting times to the CLI and the agent skill**, since ClawHub republishes that skill.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- The ClawHub submission and refresh path was already researched for the standalone skill publish work; that research flagged the exact submission mechanism as an open question for engineering to confirm against current ClawHub docs. The same applies here.
- The refresh process itself was a deliverable of that earlier work. If it exists, follow it; if it was never written down, this story is the moment to write it.
