# Epics & Stories — Publishing API v1.17

Single API epic extending the Publishing API surface with full workspace lifecycle management. Exposed only through ContentStudio's first-party developer surfaces — the Publishing API itself, the ContentStudio MCP server, and the public CLI + agent skill — **not** through Zapier, Make.com, or n8n.

**Note on adjacent scope:** inviting team members *into* a workspace is already covered by `POST /workspaces/{workspace_id}/members` in v1.15 (team members management). v1.17 is about the workspace container itself — creating it, renaming it, deleting it, listing what exists.

---

## Epic 1: Publishing API v1.17 — Workspace Management

Add full CRUD for workspaces to the Publishing API so developers, agents, and CLI users can list, inspect, create, update, and delete workspaces programmatically. Today, every other v1 endpoint takes a `workspace_id` as input but there is no public way to discover what workspace IDs exist or to spin up a new workspace without going through the web app.

**Capabilities covered:**

- List all workspaces the API key has access to
- Get one workspace's details
- Create a new workspace
- Update an existing workspace (name, timezone, logo, and any other workspace-level settings the web app exposes)
- Delete a workspace

**Out of scope for v1.17:**

- Inviting team members into a workspace (already shipped in v1.15)
- Workspace transfer / ownership change
- Workspace archive / restore
- Billing-plan changes via API
- Custom-domain / white-label configuration via API

### Stories:

**Story 1.1: [BE] Add workspace management endpoints to Publishing API v1**

> **Description:**
> As an API consumer, I want to list, inspect, create, update, and delete workspaces programmatically so that I can manage the workspaces an organisation uses without going through the ContentStudio web app every time a new client, brand, or team needs its own space.
>
> Add the following endpoints to the Publishing API v1:
>
> - `GET /api/v1/workspaces` — list workspaces the API key has access to
> - `GET /api/v1/workspaces/{workspace_id}` — get one workspace
> - `POST /api/v1/workspaces` — create a new workspace
> - `PATCH /api/v1/workspaces/{workspace_id}` — update an existing workspace
> - `DELETE /api/v1/workspaces/{workspace_id}` — delete a workspace
>
> All endpoints follow existing Publishing API v1 conventions: API key authentication, JSON in and out, standard error envelope, pagination on the list endpoint. The accepted fields on create / update and the validation rules around create and delete should match the equivalent web-app flows so the API and the UI behave consistently.
>
> ---
>
> ### Workflow:
>
> 1. API consumer authenticates with their API key.
> 2. To discover available workspaces, consumer calls `GET /workspaces` and receives the list of workspaces this API key can act on.
> 3. To inspect one workspace, consumer calls `GET /workspaces/{id}` and receives the full workspace record (name, timezone, logo, plan, created_at, member count, and any other fields the web app surfaces).
> 4. To spin up a new workspace, consumer calls `POST /workspaces` with the same fields the web-app's "Add new workspace" flow accepts. On success, the API returns the new workspace record including its `id`.
> 5. To rename or change settings, consumer calls `PATCH /workspaces/{id}` with the fields to update. The API returns the updated record.
> 6. To delete, consumer calls `DELETE /workspaces/{id}`. On success, the API returns HTTP 204 and the workspace (and its associated data) is removed per the web app's existing delete behavior.
>
> ---
>
> ### Acceptance criteria:
>
> - [ ] `GET /workspaces` returns the list of workspaces the API key's owner has access to, with pagination using the standard `page` and `per_page` parameters used elsewhere in v1.
> - [ ] `GET /workspaces/{workspace_id}` returns one workspace's full record, or HTTP 404 when the ID is unknown or the API key has no access.
> - [ ] `POST /workspaces` creates a new workspace. The accepted request body, required fields, and validation rules match the web-app's "Add new workspace" flow. On success, the API returns the new workspace record including its `id`.
> - [ ] `PATCH /workspaces/{workspace_id}` updates an existing workspace. The fields that can be changed match the web-app's workspace-settings page. Partial updates are supported (only fields present in the body are changed).
> - [ ] `DELETE /workspaces/{workspace_id}` deletes a workspace. The permission rules, confirmation requirements, and downstream cleanup (members, posts, integrations, etc.) match the web-app's delete-workspace flow exactly.
> - [ ] Every endpoint enforces the same permission rules the web app uses — if the API key's owner cannot perform this action in the UI, the API returns HTTP 403.
> - [ ] Validation errors return HTTP 422 with a clear, user-readable message that matches the message the web app surfaces for the same error.
> - [ ] API documentation (Swagger/OpenAPI) is updated with all new endpoints, request/response shapes, and error cases.
>
> ---
>
> ### Mock-ups:
> N/A — backend API only.
>
> ### Impact on existing data:
> No schema changes. Operates on the existing workspace data model the web app already uses.
>
> ### Impact on other products:
> - Web app: no impact (web app continues to use its own internal endpoints).
> - Inbox: no impact.
> - Mobile apps: no impact.
> - Chrome extension: no impact.
>
> ### Dependencies:
> None.
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A, backend-only story
> - [ ] Multilingual support — API returns data; user-facing strings (validation messages) continue to honor existing locale handling where applicable
> - [ ] UI theming support — N/A, backend-only story
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 1.2: [BE] Expose workspace management via the ContentStudio MCP server**

> **Description:**
> As an AI-agent user, I want to manage my ContentStudio workspaces through the MCP server so that I can ask my agent to "list my workspaces", "create a new workspace called Acme Corp", or "rename this workspace" and have the agent execute it safely against the real account.
>
> Expose every workspace capability added in **[BE] Add workspace management endpoints to Publishing API v1** as a corresponding MCP tool, following the existing ContentStudio MCP conventions for naming, schema description, and error reporting.
>
> ---
>
> ### Workflow:
>
> 1. Agent operator has already authenticated their agent against the ContentStudio MCP server with an API key.
> 2. Operator asks the agent in natural language ("list my workspaces", "create a new workspace named Acme Corp", "rename workspace 123 to Acme HQ", "delete workspace 456").
> 3. The agent discovers the relevant workspace MCP tools, fills in the required inputs from the operator's request, and calls them.
> 4. The MCP server forwards the call to the Publishing API endpoint added in Story 1.1 and returns the response to the agent in the standard MCP envelope.
> 5. The agent surfaces the result to the operator (workspace list, the new workspace record, confirmation of rename, confirmation of deletion).
>
> ---
>
> ### Acceptance criteria:
>
> - [ ] MCP tool `list_workspaces` exists and returns the workspaces the API key has access to.
> - [ ] MCP tool `get_workspace` exists and returns one workspace. Inputs: `workspace_id`.
> - [ ] MCP tool `create_workspace` exists and creates a new workspace. The tool schema documents which inputs are required and which are optional, matching the web-app create flow.
> - [ ] MCP tool `update_workspace` exists and updates an existing workspace. Inputs: `workspace_id` plus any updatable fields.
> - [ ] MCP tool `delete_workspace` exists and deletes a workspace. Inputs: `workspace_id`.
> - [ ] Each MCP tool's schema documents required vs. optional parameters and notes when an action is destructive (create_workspace is significant, delete_workspace is destructive) so agents can warn the operator before executing.
> - [ ] Validation errors (HTTP 422) are surfaced verbatim from the underlying API so the agent can relay the exact message to the operator.
> - [ ] Permission (HTTP 403) and not-found (HTTP 404) errors are surfaced distinctly so the agent can handle them differently.
> - [ ] The MCP server's tool catalog (returned by the standard MCP `list_tools` call) includes the five new tools with accurate descriptions.
>
> ---
>
> ### Mock-ups:
> N/A — MCP only, no UI.
>
> ### Impact on existing data:
> None.
>
> ### Impact on other products:
> - Inbox: no impact.
> - Mobile apps: no impact.
> - Chrome extension: no impact.
> - Web app: no impact.
>
> ### Dependencies:
> Depends on: **[BE] Add workspace management endpoints to Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A, backend-only story
> - [ ] Multilingual support — N/A, agent-facing tool schemas are English
> - [ ] UI theming support — N/A, backend-only story
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 1.3: [BE] Add workspace management commands to the ContentStudio CLI and agent skill**

> **Description:**
> As a developer or technical operator, I want `contentstudio workspaces ...` commands in the public CLI so that I can list, inspect, create, rename, and delete workspaces from my terminal and from any shell-capable agent that has installed the ContentStudio skill.
>
> Add a `workspaces` command group to the public CLI mirroring the endpoints from Story 1.1. Update the bundled `SKILL.md` so agents discover the new commands. Follow existing CLI conventions: `--json` for machine-readable output, exit code 0 on success and non-zero on validation / permission / not-found errors with distinct codes for each so scripts can branch on them.
>
> ---
>
> ### Workflow:
>
> 1. User has installed the public CLI and has set `CONTENTSTUDIO_API_KEY`.
> 2. To see all workspaces: `contentstudio workspaces list`.
> 3. To inspect one: `contentstudio workspaces get --workspace <id>`.
> 4. To create a new one: `contentstudio workspaces create --name "Acme Corp" ...` (with whichever additional flags the create flow expects).
> 5. To rename or update: `contentstudio workspaces update --workspace <id> --name "Acme HQ"`.
> 6. To delete: `contentstudio workspaces delete --workspace <id>`.
> 7. `--json` on any of these returns machine-readable output suitable for piping or agent use.
> 8. The bundled agent skill manifest is updated so any agent that has installed the ContentStudio skill discovers the new `workspaces` commands without manual intervention.
>
> ---
>
> ### Acceptance criteria:
>
> - [ ] `contentstudio workspaces list` prints a human-readable table of workspaces with id, name, plan, member count, created_at.
> - [ ] `contentstudio workspaces list --json` prints a JSON array matching the API response shape.
> - [ ] `contentstudio workspaces get --workspace <id>` returns one workspace's full record.
> - [ ] `contentstudio workspaces create` accepts the same inputs the web-app create flow expects and prints the new workspace record (including its `id`) on success.
> - [ ] `contentstudio workspaces update --workspace <id>` accepts the same updatable fields the web-app settings page exposes and prints the updated record.
> - [ ] `contentstudio workspaces delete --workspace <id>` removes the workspace and exits 0.
> - [ ] Missing required flags on `create` print a clear usage hint and exit non-zero.
> - [ ] Validation errors (HTTP 422) are printed verbatim with a distinct non-zero exit code.
> - [ ] Permission errors (HTTP 403) print a clear permission message with its own distinct exit code so scripts can branch on it.
> - [ ] Not-found errors (HTTP 404) print a clear "workspace not found or you don't have access" message with its own distinct exit code.
> - [ ] The bundled `SKILL.md` includes the five new commands in its declared command surface with one-line descriptions and example invocations. The agent skill flags `delete` as destructive so agents that consume the manifest can prompt the operator for confirmation.
> - [ ] CLI help (`contentstudio workspaces --help`) shows all subcommands and links to the public docs.
> - [ ] CLI launch docs (the existing public docs site) are updated with a "Manage your workspaces" section showing copy-pasteable examples.
>
> ---
>
> ### Mock-ups:
> N/A — terminal commands.
>
> ### Impact on existing data:
> None.
>
> ### Impact on other products:
> - Standalone ContentStudio agent skill repo: the bundled SKILL.md update flows downstream to Clawhub and any other registry the skill is published to on the next sync.
> - No other product impact.
>
> ### Dependencies:
> Depends on: **[BE] Add workspace management endpoints to Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A, backend-only story
> - [ ] Multilingual support — N/A, CLI is English
> - [ ] UI theming support — N/A, backend-only story
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)
