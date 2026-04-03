# Research: Public API v1.8 — Delete Posts from Native Platforms

> **Context:** This is ContentStudio's **public API** used by Zapier, Make, and other third-party integrations — not an internal endpoint.

## Current State

### Existing Delete Endpoint (v1 API)
- **Route:** `DELETE /api/v1/workspaces/{workspace_id}/posts/{post_id}` → `PostController@destroy`
- **File:** `contentstudio-backend/app/Http/Controllers/Api/V1/PostController.php:1423`
- **Behavior:** Calls `PlanController::removePlan()` which deletes the post from ContentStudio only (removes the plan record, associated comments, tasks, notifications, media links). Does **not** delete from native social platforms.

### Existing Internal "Remove Posting" (Web Planner)
- **Route:** `POST /planner/removePlanPosting` → `PlanController@removePlanPosting`
- **File:** `contentstudio-backend/app/Http/Controllers/Planner/PlanController.php:1727`
- **Job:** `PlanRemovePostingJob` — this is the internal job that already handles native platform deletion.
- **File:** `contentstudio-backend/app/Jobs/PlanRemovePostingJob.php`

### Native Deletion Support (already implemented in PlanRemovePostingJob)
Platforms that **support** native deletion:
- **Facebook** (Pages only, not Groups)
- **Twitter/X**
- **LinkedIn**
- **Pinterest**
- **Tumblr**
- **YouTube**
- **Google My Business**
- **Bluesky**
- **Threads**

Platforms that do **NOT** support native deletion:
- **Instagram** (API limitation)
- **TikTok** (API limitation)
- **Facebook Groups** (API limitation)

Logic is in `PlanRemovePostingJob::isAllowedRemovePost()` (line 135).

### How Native Deletion Works
The `PlanRemovePostingJob` receives a plan, workspace_id, posting_ids (array of `{id, delete_cs}` objects), and deleted_by. For each posting:
1. Checks if the platform allows native deletion via `isAllowedRemovePost()`
2. If allowed and `delete_cs` is false → calls the platform-specific `removePost()` method
3. If `delete_cs` is true or platform doesn't support it → just marks as deleted in CS
4. Updates the posting record with `deleted`, `deleted_message`, `deleted_by`

## What Needs to Change

- **New API endpoint:** `DELETE /api/v1/workspaces/{workspace_id}/posts/{post_id}/postings` — deletes specific postings from native platforms + ContentStudio
- **Accepts:** `posting_ids` array (each with posting ID and optional `delete_from_platform` flag), or a `delete_all` flag to delete all postings for the post
- **Reuses:** The existing `PlanRemovePostingJob` which already has all platform deletion logic
- **Response:** Returns which postings were queued for deletion and which platforms don't support native deletion
- **API docs:** Update API documentation for v1.8

## Files Involved

| File | Change |
|---|---|
| `contentstudio-backend/routes/api/v1.php` | Add new DELETE route for postings |
| `contentstudio-backend/app/Http/Controllers/Api/V1/PostController.php` | Add `destroyPostings()` method |
| `contentstudio-backend/app/Jobs/PlanRemovePostingJob.php` | Already handles native deletion — reuse as-is |
| API documentation | Document the new endpoint for Zapier/Make/third-party integrations |
