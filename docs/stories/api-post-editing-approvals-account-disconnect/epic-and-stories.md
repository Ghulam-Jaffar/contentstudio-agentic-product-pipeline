# Epic A — Public API publishing lifecycle: edit posts, approvals, account disconnect

## Problem

The public API can create, list, delete, and approve/reject posts, and connect social accounts, but it cannot:
- **Edit an existing post.** There is no update endpoint in `routes/api/v1.php` (only index/store/destroy/approval). Users who create a post via the API cannot change it, even while it is still unpublished.
- **Fully manage approvals.** Approval config is only set at creation, and the multi-level approval workflow is not represented in the public API.
- **Disconnect a social account.** There is a connect endpoint but no disconnect/remove.

## Goal

Let API users edit any post that is not published or partially-published, across the full editable payload (content, media, scheduling, accounts, and the approvers / approval workflow), with the multi-level approval workflow handled in the API. Add the ability to disconnect/remove a connected social account. Then propagate all of this to every developer surface.

## Scope (surfaces)

Backend public API first, then: **CLI (`contentstudio-cli`), MCP (`contentstudio-mcp`), Zapier, n8n, Make**.

## Rules

- **Editable only when not published.** Published and partially-published posts cannot be edited. Everything else (draft, scheduled, queued, in-review, rejected, failed) can be edited.
- **Full payload editable**, including `accounts`, media, scheduling, and the approval block (approvers, approve option, stages/levels, notes).
- **Reuse the existing v1 PUT + permission pattern** (workspaces, labels, campaigns, team-members already use it) and the existing create-post validation/transform.
- Keep the OpenAPI spec, the API guide, and the agent `SKILL.md` in sync with the new surface.

## Stories (see `02-stories.md`)

1. `[BE]` Edit non-published posts via the public API (full editable payload)
2. `[BE]` Handle the multi-level approval workflow in the public API
3. `[BE]` Disconnect / remove a connected social account via the public API
4. `[BE]` Sync OpenAPI spec, API guide, and the agent `SKILL.md`
5. `[CLI]` Add `posts:update`, approval editing, and `accounts:disconnect`
6. `[MCP]` Add update-post, approval, and disconnect tools
7. `[Zapier]` Add update-post and disconnect-account actions
8. `[n8n]` Add update-post and disconnect-account operations
9. `[Make]` Add update-post and disconnect-account modules

## Sequencing

Stories 1-3 (backend) first and can run in parallel; 4 closes out the backend contract. 5-9 (surfaces) depend on the backend contract being frozen. CLI + MCP first (we own them and they feed the agent skill), then Zapier/n8n/Make.

## Out of scope

- Editing published or partially-published posts.
- Analytics/inbox changes.
- New approval-workflow *management* (creating/editing workflow definitions) via the public API — this epic handles applying and advancing approvals on posts, not authoring workflow templates. Flag separately if needed.



---

---


# Epic A stories — Public API publishing lifecycle

## [BE] Edit non-published posts via the public API

### Description
Add an endpoint to update an existing post through the public API. Any post that is not published or partially-published can be edited, across the full payload (content, media, scheduling, accounts, labels, campaign, first comment, per-platform options, and the approval block). Published and partially-published posts return a clear error.

### Workflow (developer)
1. Developer sends an update request for a post id with the fields to change.
2. If the post is published or partially-published, the API rejects it with a clear error.
3. Otherwise the post is updated and the response returns the updated post.

### Acceptance criteria
- [ ] A public API endpoint updates an existing post by id within a workspace (follows the existing v1 `PUT` + permission pattern).
- [ ] The full editable payload is accepted: content/text, media, scheduling (publish type, time, timezone), accounts, labels, campaign, first comment, per-platform options, and the approval block.
- [ ] Editing is allowed for posts that are not published or partially-published (for example draft, scheduled, queued, in-review, rejected, failed).
- [ ] Editing a published or partially-published post returns a structured error with a clear message and correct HTTP status.
- [ ] Validation matches the create-post rules (reuse the same request validation and payload-to-internal transform).
- [ ] The response returns the updated post in the same shape as create.
- [ ] Changing `accounts` re-targets the post correctly (adding/removing channels behaves like the composer).

### Impact on existing data
No schema change expected; reuses the post model. Confirm status checks cover all non-publishable states.

### Impact on other products
Public API. Downstream surfaces (CLI/MCP/Zapier/n8n/Make) build on this. Web app and mobile use their own internal edit paths, unaffected.

### Global quality and compliance checklist
- [ ] Mobile responsiveness (N/A, backend-only story)
- [ ] Multilingual support verified (error messages localized or fallback handled)
- [ ] UI theming supported (N/A, backend-only story)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research, not a contract. Engineering may choose a different approach.*
- Add the route in `contentstudio-backend/routes/api/v1.php` next to the existing posts routes (index/store/destroy/approval). Mirror the existing v1 `PUT` + `PermissionMiddleware` pattern used by workspaces/labels/campaigns/team-members.
- Reuse `app/Http/Requests/Api/V1/PostStoreRequest.php` (or a sibling update request) and `PostController`'s `transformApiPayloadToInternal()`; add the status guard (block published + partially-published) before applying changes.
- Confirm the post status/enum values that count as published vs partially-published from the plan/post model.

---

## [BE] Handle the multi-level approval workflow in the public API

### Description
Represent and apply the multi-level approval workflow through the public API: set/replace approvers and stages when creating or editing a post, expose the current approval state, and make the approve/reject action advance multi-level workflows correctly.

### Acceptance criteria
- [ ] Create and edit accept a multi-level approval configuration (approvers, approve option, ordered stages/levels, notes).
- [ ] The approve/reject action advances a multi-level workflow correctly (a post is only fully approved when all required levels approve; a rejection sets the rejected state).
- [ ] The post response exposes the current approval state (pending level, who approved, overall status).
- [ ] Approval actions respect the acting API user's permission and the workspace's approval rules.
- [ ] Invalid approval configs (unknown approvers, empty required levels) return a structured error.

### Impact on existing data
Uses the existing approval-workflow model. No new schema expected.

### Impact on other products
Public API. Mirrors the internal multi-level approval behavior the web app already has.

### Global quality and compliance checklist
- [ ] Mobile responsiveness (N/A, backend-only story)
- [ ] Multilingual support verified (messages localized or fallback handled)
- [ ] UI theming supported (N/A, backend-only story)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research, not a contract. Engineering may choose a different approach.*
- Existing pieces: `PostController@approvalAction` (approve/reject) and the create-post `approval` block (`approvers[]`, `approve_option`, `notes`). Extend these for ordered multi-level stages.
- Internal multi-level workflow logic lives behind `contentstudio-backend/routes/api/approval-workflows.php` and its controllers/models (guarded by internal auth). Reuse that engine so API and web behavior match; do not fork the logic.

---

## [BE] Disconnect / remove a connected social account via the public API

### Description
Add an endpoint to disconnect (remove) a connected social account from a workspace through the public API, complementing the existing connect endpoint.

### Acceptance criteria
- [ ] A public API endpoint removes/disconnects a connected social account from a workspace by account id.
- [ ] It requires the appropriate permission (same class of permission that gates connecting/removing accounts).
- [ ] After disconnect, the account no longer appears in the workspace accounts listing and can no longer be targeted by posts.
- [ ] Attempting to disconnect an unknown or already-removed account returns a structured error.
- [ ] The response confirms the disconnect result.

### Impact on existing data
Removes the account connection for the workspace as the internal disconnect flow does. Confirm handling of scheduled posts that referenced the removed account.

### Impact on other products
Public API. Uses the same underlying disconnect behavior as the web app.

### Global quality and compliance checklist
- [ ] Mobile responsiveness (N/A, backend-only story)
- [ ] Multilingual support verified (messages localized or fallback handled)
- [ ] UI theming supported (N/A, backend-only story)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research, not a contract. Engineering may choose a different approach.*
- Existing connect route: `POST workspaces/{workspace_id}/connect/{platform}` (`ConnectController@connect`) in `routes/api/v1.php`. Add a disconnect route (for example `DELETE workspaces/{workspace_id}/accounts/{account_id}`) following the existing DELETE + permission pattern (see the team-member remove route).
- Reuse the internal account-removal/disconnect service the web app uses.

---

## [BE] Sync OpenAPI spec, API guide, and the agent SKILL.md

### Description
Update the API documentation and the agent skill so the new edit, approval, and disconnect capabilities are discoverable and correct.

### Acceptance criteria
- [ ] The OpenAPI spec includes the new update-post, multi-level approval, and disconnect-account operations with request/response schemas.
- [ ] The API guide (api.contentstudio.io/guide) reflects the new operations.
- [ ] The agent `SKILL.md` in the `contentstudio-agent` repo lists the new capabilities and keeps its command/tool contract in sync.
- [ ] Examples are provided for editing a post, setting a multi-level approval, and disconnecting an account.

### Global quality and compliance checklist
- [ ] Multilingual support (N/A, docs)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references
- OpenAPI annotations live in `contentstudio-backend/app/Http/Controllers/Api/V1/`; regenerate the spec after adding annotations. Keep the `contentstudioio/contentstudio-agent` `SKILL.md` and README aligned.

---

## [CLI] Add posts:update, approval editing, and accounts:disconnect

### Description
Extend `contentstudio-cli` with commands to edit a post, manage multi-level approvals, and disconnect an account, matching the new API.

### Acceptance criteria
- [ ] `posts:update` edits an existing non-published post (supports the same fields as create, including the approval block).
- [ ] Approval commands support the multi-level workflow (set approvers/levels; approve/reject already exist and advance stages correctly).
- [ ] `accounts:disconnect` removes a connected account from a workspace.
- [ ] All new commands support `--json` and `--dry-run`, follow the existing envelope, and confirm the target workspace before mutating (per the skill rules).
- [ ] Editing a published/partially-published post surfaces the API's structured error clearly.

### Global quality and compliance checklist
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references
- `contentstudioio/contentstudio-agent` repo (`src/commands/posts.ts`, `connect.ts`), colon-style commands, `--json`/`--dry-run` conventions.

---

## [MCP] Add update-post, approval, and disconnect tools

### Description
Add MCP tools to `contentstudio-mcp` for editing a post, multi-level approvals, and disconnecting an account.

### Acceptance criteria
- [ ] An update-post tool edits a non-published post via the API.
- [ ] Approval tools support the multi-level workflow.
- [ ] A disconnect-account tool removes a connected account.
- [ ] Tools return the same structured envelope as the existing MCP tools and error clearly when the API rejects (for example editing a published post).

### Implementation references
- `github.com/d4interactive/contentstudio-mcp` — mirror the existing `create_post` / `delete_post` tool pattern; call the new API endpoints.

---

## [Zapier] Add update-post and disconnect-account actions

### Description
Extend the ContentStudio Zapier app with an update-post action and a disconnect-account action.

### Acceptance criteria
- [ ] A Zapier action updates a non-published post via the API.
- [ ] A Zapier action disconnects a connected account.
- [ ] Actions surface the API's field errors and the published-post block in a Zapier-friendly way.
- [ ] Auth reuses the existing ContentStudio API key connection.

### Implementation references
- Extend the existing ContentStudio Zapier integration; reuse the shared public API auth.

---

## [n8n] Add update-post and disconnect-account operations

### Description
Extend the ContentStudio n8n node with update-post and disconnect-account operations.

### Acceptance criteria
- [ ] The n8n node supports updating a non-published post.
- [ ] The n8n node supports disconnecting a connected account.
- [ ] Errors (validation, published-post block) are surfaced in the node output.
- [ ] Auth reuses the existing ContentStudio API key credential.

### Implementation references
- Extend the community `n8n-nodes-contentstudio` node; reuse the `X-API-Key` credential.

---

## [Make] Add update-post and disconnect-account modules

### Description
Extend the ContentStudio Make app with an update-post module and a disconnect-account module.

### Acceptance criteria
- [ ] A Make module updates a non-published post.
- [ ] A Make module disconnects a connected account.
- [ ] Errors (validation, published-post block) map to clear Make module errors.
- [ ] Auth reuses the existing ContentStudio API key connection.

### Implementation references
- Extend the existing ContentStudio Make integration; reuse the shared public API auth.
