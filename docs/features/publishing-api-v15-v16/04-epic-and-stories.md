# Epics & Stories — Publishing API v1.15 & v1.16

Two sibling API epics extending the Publishing API surface. Unlike v1.9–v1.12, these capabilities are exposed only through ContentStudio's first-party developer surfaces — the Publishing API itself, the ContentStudio MCP server, and the public CLI + agent skill — **not** through Zapier, Make.com, or n8n.

---

## Epic 1: Publishing API v1.15 — Team Members Management

Add a full team-members management surface to the Publishing API so developers, agents, and CLI users can list workspace members, invite new members with a role, change an existing member's role, and remove members — all programmatically and without needing the web app.

**Capabilities covered:**

- List all members of a workspace (with their role, email, name, status)
- Invite a new member by email and assign a role at invitation time
- Update an existing member's role
- Remove a member from the workspace
- List the standard ContentStudio workspace roles (`super_admin`, `admin`, `collaborator`, `approver`) so an API consumer can discover valid role values before calling create/update

**Out of scope for v1.15:**

- Permission-set editing (the standard role-to-permission mapping is treated as fixed)
- Custom role creation
- Bulk member operations (one member per call in v1.15; can revisit if needed)

### Stories:

**Story 1.1: [BE] Add team members management endpoints to Publishing API v1**

> **Description:**
> As an API consumer, I want to programmatically list, invite, update, and remove team members of a workspace so that I can manage my team from scripts, automation, and admin tooling without using the ContentStudio web app.
>
> Add the following endpoints to the Publishing API v1, under the existing workspace-scoped namespace:
>
> - `GET /api/v1/workspaces/{workspace_id}/members` — list members
> - `GET /api/v1/workspaces/{workspace_id}/members/{member_id}` — get one member
> - `POST /api/v1/workspaces/{workspace_id}/members` — invite a new member (email + role)
> - `PATCH /api/v1/workspaces/{workspace_id}/members/{member_id}` — update a member's role
> - `DELETE /api/v1/workspaces/{workspace_id}/members/{member_id}` — remove a member
> - `GET /api/v1/workspaces/{workspace_id}/roles` — list available roles
>
> All endpoints follow the existing Publishing API conventions: API key authentication, JSON request/response, standard error envelope, pagination on the list endpoint.
>
> ---
>
> ### Workflow:
>
> 1. API consumer authenticates with their API key.
> 2. Consumer calls `GET /workspaces/{id}/members` and receives the current team list with each member's role, email, name, and invitation status.
> 3. To add a new teammate, consumer calls `POST /workspaces/{id}/members` with `email` and `role` in the body. The system sends the standard ContentStudio invitation email and returns the new member record with `status: "invited"`.
> 4. To change someone's role, consumer calls `PATCH /workspaces/{id}/members/{member_id}` with `role` in the body. The API returns the updated member record.
> 5. To remove someone, consumer calls `DELETE /workspaces/{id}/members/{member_id}`. The API returns 204 No Content. The removed user immediately loses access to the workspace.
> 6. If the consumer is unsure which role values are accepted, they call `GET /workspaces/{id}/roles` and receive the list of role identifiers and a one-line description of each.
>
> ---
>
> ### Acceptance criteria:
>
> - [ ] `GET /workspaces/{id}/members` returns the full member list with `id`, `email`, `name`, `role`, `status` (`active` or `invited`), `invited_at`, `joined_at`.
> - [ ] The list endpoint supports pagination using the same `page` and `per_page` parameters as other v1 list endpoints.
> - [ ] `GET /workspaces/{id}/members/{member_id}` returns a single member record or 404 if not found.
> - [ ] `POST /workspaces/{id}/members` accepts `email` (required) and `role` (required, one of `super_admin`, `admin`, `collaborator`, `approver`).
> - [ ] On a successful invite, the system sends the standard ContentStudio team-invitation email — identical to the one sent when inviting from the web app.
> - [ ] Inviting an email that is already a member of the workspace returns a clear validation error (HTTP 422) and does **not** send another invitation.
> - [ ] Inviting an email that already exists as a ContentStudio user adds them directly to the workspace with the chosen role (matches web-app behavior).
> - [ ] `PATCH /workspaces/{id}/members/{member_id}` accepts a new `role` value and returns the updated record.
> - [ ] Attempting to demote the last `super_admin` of a workspace returns a clear validation error (HTTP 422) — every workspace must keep at least one super admin.
> - [ ] `DELETE /workspaces/{id}/members/{member_id}` removes the member and returns HTTP 204. Removing the last super admin returns HTTP 422.
> - [ ] Attempting to remove yourself (the API-key owner) returns HTTP 422 with a clear message — users cannot delete their own membership through this endpoint.
> - [ ] `GET /workspaces/{id}/roles` returns the array of role identifiers (`super_admin`, `admin`, `collaborator`, `approver`) with a one-line `description` for each.
> - [ ] Invalid role values on `POST` or `PATCH` return HTTP 422 with a message listing the accepted values.
> - [ ] All endpoints require the API key's owner to have permission to manage members of the target workspace (matches web-app permission rules). Insufficient permission returns HTTP 403.
> - [ ] API documentation (Swagger/OpenAPI) is updated with all new endpoints, request/response shapes, and error cases.
>
> ---
>
> ### Mock-ups:
> N/A — backend API only.
>
> ### Impact on existing data:
> No schema changes. Operates on the existing workspace-team data model used by the web app.
>
> ### Impact on other products:
> - Inbox: no impact.
> - Mobile apps: no impact.
> - Chrome extension: no impact.
> - Web app: no impact (web app continues to use its own internal endpoints; this is an additive public surface).
>
> ### Dependencies:
> None.
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A, backend-only story
> - [ ] Multilingual support — invitation email continues to honor the recipient's existing locale; API response strings are English
> - [ ] UI theming support — N/A, backend-only story
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 1.2: [BE] Expose team members management via the ContentStudio MCP server**

> **Description:**
> As an AI-agent user, I want to manage my ContentStudio workspace team through the ContentStudio MCP server so that I can ask my agent to "invite Jane as a collaborator" or "list our current team" and have the agent execute it safely against the real workspace.
>
> Expose every team-members capability added in **[BE] Add team members management endpoints to Publishing API v1** as a corresponding MCP tool, following the existing ContentStudio MCP conventions for naming, schema description, and error reporting.
>
> ---
>
> ### Workflow:
>
> 1. Agent operator has already authenticated their agent against the ContentStudio MCP server with an API key.
> 2. Operator instructs the agent in natural language (e.g. "list my workspace team", "invite jane@example.com as an approver").
> 3. The agent discovers the relevant team-members MCP tools, fills in the required inputs from the operator's request, and calls them.
> 4. The MCP server forwards the call to the Publishing API endpoint added in Story 1.1 and returns the response to the agent in the standard MCP envelope.
> 5. The agent surfaces the result to the operator (member list, confirmation of invite sent, role change confirmation, removal confirmation).
>
> ---
>
> ### Acceptance criteria:
>
> - [ ] MCP tool `list_team_members` exists and returns the workspace team list. Inputs: `workspace_id`.
> - [ ] MCP tool `get_team_member` exists and returns one member. Inputs: `workspace_id`, `member_id`.
> - [ ] MCP tool `invite_team_member` exists and triggers the invitation flow. Inputs: `workspace_id`, `email`, `role`.
> - [ ] MCP tool `update_team_member_role` exists and changes a member's role. Inputs: `workspace_id`, `member_id`, `role`.
> - [ ] MCP tool `remove_team_member` exists and removes a member. Inputs: `workspace_id`, `member_id`.
> - [ ] MCP tool `list_team_roles` exists and returns the available roles plus their descriptions. No inputs other than `workspace_id`.
> - [ ] Each MCP tool's schema documents required vs. optional parameters and lists accepted role values in the parameter description so the agent does not guess.
> - [ ] When the underlying API returns 422 (e.g. last super admin protection, duplicate invite), the MCP tool surfaces the API's error message verbatim so the agent can relay it to the operator.
> - [ ] Permission errors (403) and not-found errors (404) are surfaced clearly and distinctly so the agent can handle them differently.
> - [ ] The MCP server's tool catalog (returned by the standard MCP `list_tools` call) includes the six new tools with accurate descriptions.
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
> Depends on: **[BE] Add team members management endpoints to Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A, backend-only story
> - [ ] Multilingual support — N/A, agent-facing tool schemas are English
> - [ ] UI theming support — N/A, backend-only story
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 1.3: [BE] Add team members management commands to the ContentStudio CLI and agent skill**

> **Description:**
> As a developer or technical operator, I want `contentstudio team ...` commands in the public CLI so that I can list, invite, update, and remove workspace members from my terminal and from any shell-capable agent that has installed the ContentStudio skill.
>
> Add a `team` command group to the public CLI mirroring the API endpoints from Story 1.1. Update the bundled `SKILL.md` so agents discover the new commands. Follow the existing CLI conventions: `--workspace` for the workspace flag, `--json` for machine-readable output, exit code 0 on success and non-zero on validation/permission errors.
>
> ---
>
> ### Workflow:
>
> 1. User installs (or already has installed) the public ContentStudio CLI and has set `CONTENTSTUDIO_API_KEY`.
> 2. User runs `contentstudio team list --workspace <id>` and sees the team members.
> 3. To invite a new member: `contentstudio team invite --workspace <id> --email jane@example.com --role approver`.
> 4. To change a role: `contentstudio team update --workspace <id> --member <member_id> --role admin`.
> 5. To remove a member: `contentstudio team remove --workspace <id> --member <member_id>`.
> 6. To see available roles: `contentstudio team roles --workspace <id>`.
> 7. Adding `--json` to any of these returns machine-readable output suitable for piping into another tool or for agent consumption.
> 8. The bundled agent skill manifest is updated so any agent that has installed the ContentStudio skill discovers the new `team` commands without manual intervention.
>
> ---
>
> ### Acceptance criteria:
>
> - [ ] `contentstudio team list --workspace <id>` prints a human-readable table of members with email, name, role, status.
> - [ ] `contentstudio team list --workspace <id> --json` prints a JSON array matching the API response shape.
> - [ ] `contentstudio team get --workspace <id> --member <id>` returns one member.
> - [ ] `contentstudio team invite --workspace <id> --email <email> --role <role>` sends the invitation and prints a success confirmation with the new member ID.
> - [ ] `contentstudio team update --workspace <id> --member <id> --role <role>` updates the role and prints the new record.
> - [ ] `contentstudio team remove --workspace <id> --member <id>` removes the member and exits with code 0.
> - [ ] `contentstudio team roles --workspace <id>` prints the list of available roles with their descriptions.
> - [ ] Missing required flags print a clear usage hint and exit non-zero.
> - [ ] API validation errors (422) are printed verbatim with a non-zero exit code; permission errors (403) print a clear "you do not have permission to manage this workspace's team" message with its own distinct exit code so scripts can branch on it.
> - [ ] The bundled `SKILL.md` includes the six new commands in its declared command surface with one-line descriptions and example invocations.
> - [ ] CLI help (`contentstudio team --help`) shows all subcommands and links to the public docs.
> - [ ] CLI launch docs (the existing public docs site) are updated with a "Manage your team" section showing copy-pasteable examples.
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
> - Standalone ContentStudio agent skill repo: the bundled SKILL.md update means the standalone skill repo (and downstream registries like Clawhub) automatically pick up the new commands on the next sync.
> - No other product impact.
>
> ### Dependencies:
> Depends on: **[BE] Add team members management endpoints to Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A, backend-only story
> - [ ] Multilingual support — N/A, CLI is English
> - [ ] UI theming support — N/A, backend-only story
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

## Epic 2: Publishing API v1.16 — Labels & Campaigns CRUD

Expose full create/read/update/delete capabilities for **labels** and **campaigns** through the Publishing API. v1.9 added the ability to *filter* posts by labels and campaigns; v1.16 lets developers, agents, and CLI users actually *manage* those labels and campaigns programmatically — create new ones, rename them, change their color (labels), reschedule them (campaigns), or delete the ones they no longer need.

**Capabilities covered:**

- **Labels:** list, get one, create, update, delete
- **Campaigns:** list, get one, create, update, delete

**Out of scope for v1.16:**

- Bulk-assigning labels or campaigns to many posts at once (existing per-post update endpoints already accept `label_ids` and `campaign_id` — that surface is unchanged)
- Sharing labels/campaigns across workspaces
- Label / campaign analytics (analytics endpoints remain a separate surface)

### Stories:

**Story 2.1: [BE] Add labels and campaigns CRUD endpoints to Publishing API v1**

> **Description:**
> As an API consumer, I want to list, create, update, and delete labels and campaigns programmatically so that I can keep my workspace's taxonomy in sync with my external systems (CRM, content brief tool, project tracker) without manual web-app upkeep.
>
> Add the following endpoints to the Publishing API v1, under the existing workspace-scoped namespace:
>
> **Labels**
> - `GET /api/v1/workspaces/{workspace_id}/labels` — list labels
> - `GET /api/v1/workspaces/{workspace_id}/labels/{label_id}` — get one label
> - `POST /api/v1/workspaces/{workspace_id}/labels` — create a label
> - `PATCH /api/v1/workspaces/{workspace_id}/labels/{label_id}` — update a label
> - `DELETE /api/v1/workspaces/{workspace_id}/labels/{label_id}` — delete a label
>
> **Campaigns**
> - `GET /api/v1/workspaces/{workspace_id}/campaigns` — list campaigns
> - `GET /api/v1/workspaces/{workspace_id}/campaigns/{campaign_id}` — get one campaign
> - `POST /api/v1/workspaces/{workspace_id}/campaigns` — create a campaign
> - `PATCH /api/v1/workspaces/{workspace_id}/campaigns/{campaign_id}` — update a campaign
> - `DELETE /api/v1/workspaces/{workspace_id}/campaigns/{campaign_id}` — delete a campaign
>
> All endpoints follow existing Publishing API conventions: API key auth, JSON in and out, standard error envelope, pagination on list endpoints.
>
> ---
>
> ### Workflow:
>
> 1. API consumer authenticates with their API key and chooses a workspace.
> 2. To see the current taxonomy: consumer calls `GET /workspaces/{id}/labels` or `GET /workspaces/{id}/campaigns` and receives the existing list.
> 3. To create a new label: consumer calls `POST /workspaces/{id}/labels` with `name` and `color` (hex code) in the body. The API returns the new label with its ID.
> 4. To rename a label or change its color: consumer calls `PATCH /workspaces/{id}/labels/{label_id}`.
> 5. To delete a label: consumer calls `DELETE /workspaces/{id}/labels/{label_id}`. The label is removed and detached from any posts it was tagged on (matching web-app behavior — the posts themselves are not deleted).
> 6. The same shape applies to campaigns, with the campaign-specific fields (`name`, `start_date`, `end_date`, `description`, `color`).
>
> ---
>
> ### Acceptance criteria:
>
> **Labels**
> - [ ] `GET /workspaces/{id}/labels` returns the full label list with `id`, `name`, `color`, `created_at`, `posts_count`.
> - [ ] The list endpoint supports `page` and `per_page` pagination.
> - [ ] `GET /workspaces/{id}/labels/{label_id}` returns a single label or 404 if not found.
> - [ ] `POST /workspaces/{id}/labels` accepts `name` (required, 1–50 chars) and `color` (required, valid hex like `#FF5733`). Returns the created label with its new `id`.
> - [ ] Creating a label with a name that already exists in this workspace returns HTTP 422.
> - [ ] `PATCH /workspaces/{id}/labels/{label_id}` accepts partial updates (`name` only, `color` only, or both) and returns the updated record.
> - [ ] `DELETE /workspaces/{id}/labels/{label_id}` removes the label and detaches it from every post it was tagged on. Returns HTTP 204. The posts themselves are not deleted.
>
> **Campaigns**
> - [ ] `GET /workspaces/{id}/campaigns` returns the full campaign list with `id`, `name`, `description`, `color`, `start_date`, `end_date`, `posts_count`, `created_at`.
> - [ ] The list endpoint supports `page` and `per_page` pagination, plus an optional `status` filter accepting `upcoming`, `active`, `completed` (computed from `start_date` / `end_date`).
> - [ ] `GET /workspaces/{id}/campaigns/{campaign_id}` returns a single campaign or 404.
> - [ ] `POST /workspaces/{id}/campaigns` accepts `name` (required, 1–100 chars), `description` (optional), `color` (required hex), `start_date` (required, ISO 8601), `end_date` (required, ISO 8601, must be ≥ start_date).
> - [ ] `end_date < start_date` on create or update returns HTTP 422 with a clear message.
> - [ ] `PATCH /workspaces/{id}/campaigns/{campaign_id}` accepts partial updates and returns the updated record.
> - [ ] `DELETE /workspaces/{id}/campaigns/{campaign_id}` removes the campaign and unassigns it from every post that referenced it. Returns HTTP 204. The posts themselves are not deleted.
>
> **General**
> - [ ] All endpoints require the API key's owner to have permission on the target workspace; otherwise HTTP 403.
> - [ ] Invalid hex colors return HTTP 422 with a clear validation message.
> - [ ] API documentation (Swagger/OpenAPI) is updated with all new endpoints, fields, validation rules, and error cases.
>
> ---
>
> ### Mock-ups:
> N/A — backend API only.
>
> ### Impact on existing data:
> No schema changes. Operates on the existing labels and campaigns data the web app already manages.
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
> - [ ] Multilingual support — N/A, API responses are data-only
> - [ ] UI theming support — N/A, backend-only story
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 2.2: [BE] Expose labels and campaigns CRUD via the ContentStudio MCP server**

> **Description:**
> As an AI-agent user, I want to manage labels and campaigns from inside an agent conversation ("create a 'Holiday 2026' campaign from Nov 15 to Dec 31", "rename the 'Q4 push' label to 'Q4 launch'") so that I can keep my workspace organised without leaving the conversation.
>
> Expose every label and campaign capability from **[BE] Add labels and campaigns CRUD endpoints to Publishing API v1** as a corresponding MCP tool, following existing ContentStudio MCP conventions.
>
> ---
>
> ### Workflow:
>
> 1. Operator has authenticated their agent against the ContentStudio MCP server.
> 2. Operator asks the agent to list / create / rename / delete labels or campaigns in natural language.
> 3. The agent discovers the relevant MCP tools, fills inputs from the operator's request, and calls them.
> 4. The MCP server forwards each call to the Publishing API endpoint added in Story 2.1 and returns the result in the standard MCP envelope.
> 5. The agent surfaces the result to the operator (taxonomy list, confirmation of create, updated record, removal confirmation).
>
> ---
>
> ### Acceptance criteria:
>
> **Labels**
> - [ ] MCP tools exist: `list_labels`, `get_label`, `create_label`, `update_label`, `delete_label`.
> - [ ] `create_label` requires `workspace_id`, `name`, `color`. The tool schema documents the hex-color format with an example.
>
> **Campaigns**
> - [ ] MCP tools exist: `list_campaigns`, `get_campaign`, `create_campaign`, `update_campaign`, `delete_campaign`.
> - [ ] `create_campaign` requires `workspace_id`, `name`, `color`, `start_date`, `end_date` and accepts optional `description`. The schema documents the ISO 8601 date format with an example.
> - [ ] `list_campaigns` accepts an optional `status` filter (`upcoming` / `active` / `completed`).
>
> **General**
> - [ ] Each MCP tool's schema documents required vs. optional parameters and their formats.
> - [ ] Validation errors (e.g. invalid hex, end_date before start_date, duplicate label name) are surfaced verbatim from the API.
> - [ ] Permission (403) and not-found (404) errors are surfaced distinctly so the agent can handle them differently.
> - [ ] The MCP server's tool catalog includes the ten new tools with accurate descriptions.
>
> ---
>
> ### Mock-ups:
> N/A.
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
> Depends on: **[BE] Add labels and campaigns CRUD endpoints to Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A, backend-only story
> - [ ] Multilingual support — N/A
> - [ ] UI theming support — N/A, backend-only story
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

**Story 2.3: [BE] Add labels and campaigns commands to the ContentStudio CLI and agent skill**

> **Description:**
> As a developer or technical operator, I want `contentstudio labels ...` and `contentstudio campaigns ...` commands in the public CLI so that I can manage my workspace's taxonomy from the terminal and from any shell-capable agent that has installed the ContentStudio skill.
>
> Add `labels` and `campaigns` command groups to the public CLI, mirroring the endpoints from Story 2.1. Update the bundled `SKILL.md` so agents discover the new commands. Follow existing CLI conventions (`--workspace` flag, `--json` flag, standard exit codes).
>
> ---
>
> ### Workflow:
>
> 1. User has installed the CLI and set `CONTENTSTUDIO_API_KEY`.
> 2. To see all labels: `contentstudio labels list --workspace <id>`.
> 3. To create a label: `contentstudio labels create --workspace <id> --name "Q4 launch" --color "#FF5733"`.
> 4. To rename a label: `contentstudio labels update --workspace <id> --label <label_id> --name "Q4 push"`.
> 5. To delete a label: `contentstudio labels delete --workspace <id> --label <label_id>`.
> 6. The same shape applies to campaigns: `contentstudio campaigns list / get / create / update / delete`, with the additional `--start-date`, `--end-date`, and `--description` flags on create/update.
> 7. `--json` on any command returns machine-readable output for piping or agent use.
> 8. The bundled agent skill manifest is updated so agents discover the new command groups automatically.
>
> ---
>
> ### Acceptance criteria:
>
> **Labels**
> - [ ] `contentstudio labels list --workspace <id>` prints a table with `id`, `name`, `color`, `posts_count`.
> - [ ] `contentstudio labels get --workspace <id> --label <id>` returns one label.
> - [ ] `contentstudio labels create --workspace <id> --name <name> --color <hex>` creates the label and prints the new record.
> - [ ] `contentstudio labels update --workspace <id> --label <id>` accepts `--name` and/or `--color` and prints the updated record.
> - [ ] `contentstudio labels delete --workspace <id> --label <id>` removes the label and exits 0.
>
> **Campaigns**
> - [ ] `contentstudio campaigns list --workspace <id>` prints a table with `id`, `name`, `start_date`, `end_date`, `status`, `posts_count`.
> - [ ] `contentstudio campaigns list --workspace <id> --status active` filters to active campaigns only.
> - [ ] `contentstudio campaigns get --workspace <id> --campaign <id>` returns one campaign.
> - [ ] `contentstudio campaigns create --workspace <id> --name <name> --color <hex> --start-date <iso> --end-date <iso>` creates the campaign and prints the new record. `--description` is optional.
> - [ ] `contentstudio campaigns update --workspace <id> --campaign <id>` accepts partial updates and prints the updated record.
> - [ ] `contentstudio campaigns delete --workspace <id> --campaign <id>` removes the campaign and exits 0.
>
> **General**
> - [ ] `--json` works on every command and prints the API response shape.
> - [ ] Missing required flags print a clear usage hint and exit non-zero.
> - [ ] Validation errors (422) are printed verbatim; permission errors (403) print a clear permission message with a distinct exit code.
> - [ ] The bundled `SKILL.md` lists all ten new commands with one-line descriptions and an example for each.
> - [ ] CLI help (`contentstudio labels --help`, `contentstudio campaigns --help`) shows all subcommands and links to public docs.
> - [ ] CLI public docs are updated with a "Manage your taxonomy" section showing copy-pasteable examples for labels and campaigns.
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
> - Standalone ContentStudio agent skill repo: the SKILL.md update flows downstream to Clawhub and any other registry the skill is published to.
> - No other product impact.
>
> ### Dependencies:
> Depends on: **[BE] Add labels and campaigns CRUD endpoints to Publishing API v1**
>
> ### Global quality & compliance (wherever applicable)
> - [ ] Mobile responsiveness — N/A, backend-only story
> - [ ] Multilingual support — N/A, CLI is English
> - [ ] UI theming support — N/A, backend-only story
> - [ ] White-label domains impact review
> - [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)
