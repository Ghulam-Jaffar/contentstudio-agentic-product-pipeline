# Publishing API v1.8 — Delete Posts from Native Platforms

## Epic Description

Extend the Publishing API v1 to support deleting posts from native social platforms (not just from ContentStudio). Currently the `DELETE /posts/{post_id}` endpoint only removes the post from ContentStudio's planner. This epic adds a new endpoint that allows deleting individual postings from native platforms (Facebook Pages, Twitter/X, LinkedIn, Pinterest, Tumblr, YouTube, GMB, Bluesky, Threads) and from ContentStudio in a single call. Once the core API endpoint is ready, update all integration apps (Zapier, Make.com, n8n, GPT app, Claude extension, MCP) to surface the new deletion capability.

Platforms that **do not** support native deletion due to API limitations: Instagram, TikTok, Facebook Groups.

---

## Story 1: [BE] Add delete postings endpoint to Publishing API v1

### Description:

As an API consumer (Zapier, Make.com, or direct API user), I want to delete specific postings from native social platforms and/or from ContentStudio so that I can manage published content across platforms without logging into ContentStudio.

Add a new endpoint `DELETE /api/v1/workspaces/{workspace_id}/posts/{post_id}/postings` that accepts an array of posting IDs with a `delete_from_platform` flag. The endpoint should dispatch the existing `PlanRemovePostingJob` to handle native platform deletion — all platform-specific deletion logic already exists.

**Key changes:**
- `contentstudio-backend/routes/api/v1.php` — add the new DELETE route
- `contentstudio-backend/app/Http/Controllers/Api/V1/PostController.php` — add `destroyPostings()` method that validates the request and dispatches `PlanRemovePostingJob`
- `contentstudio-backend/app/Jobs/PlanRemovePostingJob.php` — already handles all native deletion logic, reuse as-is
- `contentstudio-backend/storage/api-docs/api-docs.json` — update Swagger/OpenAPI docs

The internal `PlanRemovePostingJob` already supports per-posting deletion with platform detection and the `isAllowedRemovePost()` check. This story just exposes that capability through the public API.

---

### Workflow:

1. API consumer sends `DELETE /api/v1/workspaces/{workspace_id}/posts/{post_id}/postings` with a JSON body containing the posting IDs to delete
2. The API validates that the post belongs to the workspace and the posting IDs belong to the post
3. For each posting, if `delete_from_platform` is true and the platform supports it, the post is deleted from the native platform and from ContentStudio
4. For each posting, if `delete_from_platform` is false (or the platform doesn't support native deletion), the post is only removed from ContentStudio
5. The API returns a response indicating which postings were queued for deletion and which platforms don't support native deletion

---

### Acceptance criteria:

- [ ] `DELETE /api/v1/workspaces/{workspace_id}/posts/{post_id}/postings` endpoint exists and requires API key authentication
- [ ] The request body accepts `postings` — an array of objects with `posting_id` (string, required) and `delete_from_platform` (boolean, optional, defaults to true)
- [ ] When `delete_from_platform` is true and the platform supports it (Twitter/X, Facebook Pages, LinkedIn, Pinterest, Tumblr, YouTube, GMB, Bluesky, Threads), the post is deleted from the native platform
- [ ] When `delete_from_platform` is true but the platform doesn't support deletion (Instagram, TikTok, Facebook Groups), the posting is still deleted from ContentStudio and the response indicates native deletion was not possible
- [ ] When `delete_from_platform` is false, the posting is only deleted from ContentStudio (same as current behavior)
- [ ] Invalid `post_id` returns 404
- [ ] Invalid `posting_id` (doesn't belong to the post) returns 400 with a clear error message
- [ ] Swagger/OpenAPI documentation is updated with the new endpoint, request body schema, and response schema
- [ ] Existing `DELETE /posts/{post_id}` endpoint continues to work unchanged (backwards compatible)
- [ ] Rate limiting applies to this endpoint consistently with other API endpoints

---

### Mock-ups:

N/A — backend API only.

---

### Impact on existing data:

Postings deleted from native platforms cannot be recovered. The posting record in ContentStudio is marked with `deleted: true`, `deleted_message`, and `deleted_by` — same as the existing web planner deletion behavior.

---

### Impact on other products:

- Enables all integration apps (Zapier, Make.com, n8n, GPT app, Claude extension, MCP) to offer post deletion with native platform support.
- No impact on the web app, mobile apps, or Chrome extension — the existing planner deletion flow is unchanged.

---

### Dependencies:

None — this is the foundational endpoint that all integration stories depend on.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend API only
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend API only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---
---

## Story 2: [BE] Update Zapier app to support delete post action with native platform deletion

### Description:

As a Zapier user, I want a "Delete Post" action in the ContentStudio Zapier app so that I can delete posts from native social platforms and ContentStudio as part of my automated workflows.

Add a new action to the Zapier app that calls the new `DELETE /api/v1/workspaces/{workspace_id}/posts/{post_id}/postings` endpoint. The action should let users specify which postings to delete and whether to delete from native platforms.

---

### Workflow:

1. User selects "Delete Post Postings" as an action in their Zapier workflow
2. User selects the workspace and provides the post ID (from a previous step or search)
3. User provides posting IDs and chooses whether to delete from native platforms
4. Zapier calls the API and reports the result back to the workflow

---

### Acceptance criteria:

- [ ] "Delete Post Postings" action is available in the ContentStudio Zapier app
- [ ] Action accepts workspace ID, post ID, posting IDs, and a `delete_from_platform` toggle
- [ ] Action calls the new delete postings API endpoint
- [ ] Successful deletion returns the API response to the Zapier workflow for downstream steps
- [ ] Errors from the API (404, 400, etc.) are surfaced clearly in Zapier

---

### Mock-ups:

N/A — Zapier action configuration UI is auto-generated from the action schema.

---

### Impact on existing data:

None — this is a new action, existing Zapier workflows are unaffected.

---

### Impact on other products:

None — Zapier app only.

---

### Dependencies:

Depends on **[BE] Add delete postings endpoint to Publishing API v1**.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend integration
- [ ] Multilingual support — N/A, Zapier app
- [ ] UI theming support — N/A, Zapier app
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---
---

## Story 3: [BE] Update Make.com app to support delete post module with native platform deletion

### Description:

As a Make.com user, I want a "Delete Post Postings" module in the ContentStudio Make.com app so that I can delete posts from native social platforms and ContentStudio as part of my automated scenarios.

Add a new module to the Make.com app that calls the new delete postings endpoint. The module should let users specify which postings to delete and whether to remove from native platforms.

---

### Workflow:

1. User adds a "Delete Post Postings" module to their Make.com scenario
2. User maps the workspace ID, post ID, and posting IDs from previous modules
3. User chooses whether to delete from native platforms
4. Make.com executes the module and passes the result to downstream modules

---

### Acceptance criteria:

- [ ] "Delete Post Postings" module is available in the ContentStudio Make.com app
- [ ] Module accepts workspace ID, post ID, posting IDs, and a `delete_from_platform` toggle
- [ ] Module calls the new delete postings API endpoint
- [ ] Successful deletion returns the API response for downstream modules
- [ ] Errors from the API are surfaced clearly in Make.com

---

### Mock-ups:

N/A — Make.com module configuration UI is auto-generated from the module schema.

---

### Impact on existing data:

None — this is a new module, existing Make.com scenarios are unaffected.

---

### Impact on other products:

None — Make.com app only.

---

### Dependencies:

Depends on **[BE] Add delete postings endpoint to Publishing API v1**.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend integration
- [ ] Multilingual support — N/A, Make.com app
- [ ] UI theming support — N/A, Make.com app
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---
---

## Story 4: [BE] Update n8n node to support delete post action with native platform deletion

### Description:

As an n8n user, I want a "Delete Post Postings" operation in the ContentStudio n8n node so that I can delete posts from native social platforms and ContentStudio as part of my n8n workflows.

Add a new operation to the n8n node that calls the new delete postings endpoint.

---

### Workflow:

1. User selects the "Delete Post Postings" operation in the ContentStudio n8n node
2. User provides workspace ID, post ID, and posting IDs
3. User chooses whether to delete from native platforms
4. n8n executes the operation and passes the result to the next node

---

### Acceptance criteria:

- [ ] "Delete Post Postings" operation is available in the ContentStudio n8n node
- [ ] Operation accepts workspace ID, post ID, posting IDs, and a `delete_from_platform` toggle
- [ ] Operation calls the new delete postings API endpoint
- [ ] Successful deletion returns the API response for downstream nodes
- [ ] Errors from the API are surfaced clearly in n8n

---

### Mock-ups:

N/A — n8n node UI is auto-generated from the node schema.

---

### Impact on existing data:

None — new operation, existing n8n workflows are unaffected.

---

### Impact on other products:

None — n8n node only.

---

### Dependencies:

Depends on **[BE] Add delete postings endpoint to Publishing API v1**.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend integration
- [ ] Multilingual support — N/A, n8n node
- [ ] UI theming support — N/A, n8n node
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---
---

## Story 5: [BE] Update GPT app action schema to support post deletion with native platform deletion

### Description:

As a user of the ContentStudio GPT app, I want to be able to delete posts from native platforms through the GPT conversation so that I can manage my published content without leaving ChatGPT.

Update the GPT app action schema to include a delete post postings action that calls the new API endpoint.

---

### Workflow:

1. User asks the GPT app to delete a specific post from a platform
2. The GPT app calls the delete postings endpoint with the appropriate parameters
3. The result is returned to the user in the conversation

---

### Acceptance criteria:

- [ ] GPT app action schema includes a "Delete Post Postings" action
- [ ] Action maps to the new delete postings API endpoint
- [ ] Action accepts workspace ID, post ID, posting IDs, and delete_from_platform flag
- [ ] GPT app can execute the action and return the result to the user

---

### Mock-ups:

N/A — GPT app action schema.

---

### Impact on existing data:

None — new action, existing GPT app behavior is unaffected.

---

### Impact on other products:

None — GPT app only.

---

### Dependencies:

Depends on **[BE] Add delete postings endpoint to Publishing API v1**.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend integration
- [ ] Multilingual support — N/A, GPT app
- [ ] UI theming support — N/A, GPT app
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---
---

## Story 6: [BE] Update Claude extension to support post deletion with native platform deletion

### Description:

As a user of the ContentStudio Claude extension, I want to be able to delete posts from native platforms through the Claude conversation so that I can manage my published content without leaving Claude.

Update the Claude extension to include a delete post postings tool that calls the new API endpoint.

---

### Workflow:

1. User asks Claude to delete a specific post from a platform
2. The Claude extension calls the delete postings endpoint with the appropriate parameters
3. The result is returned to the user in the conversation

---

### Acceptance criteria:

- [ ] Claude extension includes a "Delete Post Postings" tool
- [ ] Tool maps to the new delete postings API endpoint
- [ ] Tool accepts workspace ID, post ID, posting IDs, and delete_from_platform flag
- [ ] Claude can execute the tool and return the result to the user

---

### Mock-ups:

N/A — Claude extension tool.

---

### Impact on existing data:

None — new tool, existing Claude extension behavior is unaffected.

---

### Impact on other products:

None — Claude extension only.

---

### Dependencies:

Depends on **[BE] Add delete postings endpoint to Publishing API v1**.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend integration
- [ ] Multilingual support — N/A, Claude extension
- [ ] UI theming support — N/A, Claude extension
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---
---

## Story 7: [BE] Update MCP server tools to support post deletion with native platform deletion

### Description:

As a user of the ContentStudio MCP server, I want a delete post postings tool so that I can delete posts from native platforms through any MCP-compatible client.

Update the MCP server to include a delete post postings tool that calls the new API endpoint.

---

### Workflow:

1. MCP client calls the delete post postings tool with workspace ID, post ID, posting IDs, and delete_from_platform flag
2. The MCP server calls the delete postings API endpoint
3. The result is returned to the MCP client

---

### Acceptance criteria:

- [ ] MCP server includes a `delete_post_postings` tool
- [ ] Tool maps to the new delete postings API endpoint
- [ ] Tool accepts workspace ID, post ID, posting IDs, and delete_from_platform flag
- [ ] Tool returns the API response including which postings were deleted and which platforms didn't support native deletion

---

### Mock-ups:

N/A — MCP server tool.

---

### Impact on existing data:

None — new tool, existing MCP server behavior is unaffected.

---

### Impact on other products:

None — MCP server only.

---

### Dependencies:

Depends on **[BE] Add delete postings endpoint to Publishing API v1**.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend integration
- [ ] Multilingual support — N/A, MCP server
- [ ] UI theming support — N/A, MCP server
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---
---
