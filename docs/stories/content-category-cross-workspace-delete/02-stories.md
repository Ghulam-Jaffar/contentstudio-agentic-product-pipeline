# Story — Delete global content-category posts across all workspaces

A single full-stack story: correlate the sibling posts that a global content category fans out, surface an opt-in "delete from all workspaces" option in the planner's delete confirmation, and propagate the delete across the other workspaces (respecting membership and permissions).

**Scope note:** The cross-workspace delete option applies to global content-category posts that are **not yet published** (scheduled, queued, draft, pending). Posts that have already gone live in other workspaces are left untouched — they can't be un-published. The option is offered **only when deleting from the workspace where the post was originally created**.

---

## [Full Stack] Delete global content-category posts across all workspaces from the planner

### Description
As a user who runs a **global content category** across several workspaces, I want deleting one of its posts to optionally clean up the matching posts in my other workspaces too, so that I don't have to switch into each workspace and delete the same post over and over.

Today, a post created from a global content category is auto-scheduled into every workspace where that category exists, but deleting it only removes it from the current workspace — every other workspace keeps its copy, and there's no way to identify or remove those copies together.

### Workflow

```mermaid
flowchart TD
    Start([User clicks Delete on a post]) --> Check{Global content-category post,<br/>deleted from THIS origin workspace,<br/>still existing in other workspaces?}
    Check -->|No| Confirm[Standard delete confirmation]
    Check -->|Yes| ConfirmPlus[Confirmation with extra option:<br/>"Also delete from all other workspaces"]
    ConfirmPlus --> Toggle{User turns the option on?}
    Toggle -->|No| DelOne[Delete only this workspace's post]
    Toggle -->|Yes| Siblings[Find matching posts in other<br/>workspaces sharing this global category]
    Confirm --> DelOne
    Siblings --> Perm{User has delete access there<br/>AND post not yet published?}
    Perm -->|Yes| DelSibling[Delete that post + refill its category slot]
    Perm -->|No| Skip[Leave that workspace's post untouched]
    DelOne --> Toast1[Toast: post deleted]
    DelSibling --> Toast2[Toast: deleted here + other workspaces]
    Skip --> Toast2
```

1. In the planner, the user clicks **Delete** on a post that was created from a global content category in the workspace where it was originally created.
2. The delete confirmation appears with an extra opt-in: a checkbox to also delete the post from all the other workspaces where this category exists. The checkbox is **off** by default, and the confirmation explains in plain language that this post also exists in the user's other workspaces and what turning the option on will do.
3. The user confirms:
   - **Option off** → only this workspace's post is deleted (today's behavior).
   - **Option on** → the system finds the matching posts created in the other workspaces from the same global content category and deletes them too.
4. For each of those other workspaces, the matching post is deleted **only if** the user is a member there with permission to delete posts and the post hasn't been published yet; otherwise that workspace's post is left as-is. Each deleted post frees up its content-category queue slot the same way a normal single-post delete does today.
5. While the delete is in progress, the confirm button shows a loading state and can't be clicked again.
6. On success, the user sees a toast: a simple "post deleted" when only this workspace was affected, or a message naming how many other workspaces the post was also removed from. If some workspaces were skipped because the user doesn't have delete access there, the toast says so.
7. If the delete fails, the user sees an error toast and the post stays in place.

### Acceptance criteria

**Where the option appears**
- [ ] For a post created from a **global content category in the current (origin) workspace** that still exists in other workspaces, the delete confirmation shows an "also delete from all other workspaces" option (a `Checkbox`), unchecked by default.
- [ ] The option is **not** shown for: posts not created from a content category; posts from a workspace-only (non-global) category; posts being deleted from a workspace other than the one they were created in; or global-category posts that no longer exist in any other workspace.

**Delete behavior**
- [ ] With the option left unchecked, confirming delete removes only the current workspace's post — behavior is unchanged from today.
- [ ] With the option checked, confirming delete removes this post and the matching **not-yet-published** posts (scheduled, queued, draft, pending) in the other workspaces the user has delete access to.
- [ ] Only workspaces where the user is a member **with permission to delete posts** are affected; workspaces where the user has no access (or no delete permission) keep their post, and those skips are reported back to the UI.
- [ ] Already-published matching posts in other workspaces are left untouched.
- [ ] Each cross-workspace deletion refills the freed content-category queue slot and runs the same post-removal cleanup that a single-workspace delete runs today, per affected workspace.
- [ ] A post with no matching posts in other workspaces (e.g., the category now exists only in this workspace) deletes cleanly with no error, affecting only the current workspace.
- [ ] The delete request remains backward-compatible: callers that don't send the new option (including mobile apps and public API) get today's single-workspace delete.

**Correlation & data exposed to the UI**
- [ ] Posts created together across workspaces by a single global-content-category posting run are correlated afterwards, so the matching posts in the other workspaces can be identified from any one of them.
- [ ] The post data the planner reads indicates (a) whether the post was created from a **global** content category, (b) whether the current workspace is the one it was **originally created in**, and (c) how many other workspaces still hold a matching not-yet-published post.

**UI states**
- [ ] The confirm/delete button is disabled and shows a loading state while the request is pending, and can't trigger a second request.
- [ ] On success with only the current workspace affected, a toast appears: **"Post deleted."**
- [ ] On success with other workspaces affected, a toast appears naming the count, e.g. **"Post deleted from this and 3 other workspaces."**
- [ ] If some workspaces were skipped for lack of access, a toast reflects it, e.g. **"Post deleted. Some workspaces were skipped because you don't have delete access there."**
- [ ] On failure, an error toast appears — **"We couldn't delete the post. Please try again."** — and the post remains in the list.
- [ ] The confirmation names the category and other-workspace count when available, and falls back to generic copy when they aren't.

### UI copy

**Delete confirmation (custom modal for global content-category origin posts)** — the current quick-confirm dialog can't host a checkbox, so this uses a small `Modal` with a `Checkbox`:

- **Title:** "Delete this post?"
- **Body (when category name + count available):** "This post was created from the global content category "{categoryName}", so a copy was also scheduled in {count} other workspace(s)."
- **Body (fallback):** "This post was created from a global content category, so copies were also scheduled in your other workspaces."
- **Checkbox label:** "Also delete this post from all other workspaces where this category exists"
- **Checkbox helper text (below the checkbox):** "We'll only remove posts that haven't been published yet — anything already live in another workspace stays as it is. Example: if this post is still scheduled in 3 workspaces, it's removed from all 3."
- **Primary button (option off):** "Delete"
- **Primary button (option on):** "Delete from all workspaces"
- **Secondary button:** "Cancel"

**Toasts:**
- Single-workspace success: "Post deleted."
- Multi-workspace success: "Post deleted from this and {count} other workspace(s)."
- Partial success (some skipped): "Post deleted. Some workspaces were skipped because you don't have delete access there."
- Error: "We couldn't delete the post. Please try again."

**States:**
- **Loading:** primary button shows a `Loader` / "Deleting…" and is disabled; Cancel remains available.
- **Error:** handled via the global alert/toast (error copy above); modal closes or stays open per existing delete-error handling.
- Empty state: N/A — this is a confirmation modal, not a new list/view.

### Mock-ups
N/A — no mockups provided. Copy and component choices are specified above.

### Impact on existing data
- Posts created via the global-content-category fan-out gain the data needed to correlate siblings and identify the origin workspace (e.g., a shared correlation identifier plus an origin marker). No existing fields are removed or repurposed.
- **Forward-looking:** posts created **before** this ships won't carry the correlation data, so the cross-workspace option can't find their siblings. Decide whether to backfill existing category posts or accept that the option applies only to posts created after release (see Implementation references).

### Impact on other products
- **Web only.** The option lives in the web planner's delete flow, and the new delete instruction is optional and backward-compatible — so **mobile apps (iOS/Android)** and the **public API** are unaffected (they continue to delete a single workspace's post as today).
- **Chrome extension:** unaffected (no content-category post management).
- **White-label:** cross-workspace deletion can span white-label workspaces the user belongs to; the permission check must gate this the same way workspace membership is gated elsewhere. The confirmation modal is built from `@contentstudio/ui` components (`Modal`, `Checkbox`, `Button`), so it inherits white-label theming automatically — no hardcoded colors.

### Dependencies
None — self-contained.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Backend — primary entry points:**
- `contentstudio-backend/app/Jobs/SocialContentCategoryPostJob.php` — the fan-out job. It resolves the `global_content_category_id`, fetches all linked workspace categories (`ContentCategoriesRepository::getByGlobalCategoryId`), and creates one plan per workspace (origin workspace keeps its real `planId`; others get `planId = ""`). This is where a shared correlation id and an origin marker would be stamped on each sibling plan in the loop.
- `contentstudio-backend/app/Http/Controllers/Planner/PlanController.php` → `removePlan` — delete entry point. Today it calls `PlansRepository::removePlans([... 'workspace_id' => ...])`, always scoped to one workspace, then refills category slots via `ContentCategoriesRepository::fillRemovedSlots` and dispatches `PlanRemoveJob`. Extend to loop the sibling workspaces when the opt-in is set.
- `contentstudio-backend/app/Repository/Publish/Planner/PlansRepository.php` → `removePlans`, `fetchContentCategoriesPlans`, `savePlanByCreate`.
- `contentstudio-backend/routes/web/planner.php` — `POST /removePlan` route.

**Backend — patterns to follow:**
- `contentstudio-backend/app/Http/Controllers/Settings/ContentCategories/GlobalContentCategoriesController.php` → `store` already resolves the user's workspaces via `WorkspaceTeamRepo::getWorkspaceIdsByUserId(Auth::id())` and builds per-workspace member permission with `new PermissionHelper(Auth::id(), $workspace_id)`. Reuse the same access/permission resolution to gate which sibling workspaces may be deleted from.
- The origin vs non-origin distinction already exists implicitly in `SocialContentCategoryPostJob` (`$category['workspace_id'] !== $store_payload['workspace_id']` → `planId = ""`), so the origin plan is identifiable at creation time.

**Frontend — primary entry points:**
- `contentstudio-frontend/src/modules/planner_v2/composables/usePlannerActions.ts` → `removePlan`. Non-published deletes currently use `$cstuModal.msgBoxConfirm` (copy keys `planner.planner_actions.delete_post_confirm` / `remove_post_title`) — this has **no checkbox slot**, so a small module-level confirmation component (`Modal` + `Checkbox` + `Button`) is needed for global-category origin posts; keep `msgBoxConfirm` for every other post.
- `contentstudio-frontend/src/modules/planner_v2/queries/usePlannerMutations.ts` → `useDeletePlanMutation` / `removePlansFromQueryData`, and `contentstudio-frontend/src/api/planner.ts` (`RemovePlanPayload`, `removePlanApi`) + `src/config/planner-api-utils.ts` (`removePlanURL`) — add the optional cross-workspace flag to the delete payload.
- `contentstudio-frontend/src/locales/*/planner.json` (namespace `planner`, keys under `planner.planner_actions.*`) — add all new copy to **all 8 locale dirs** in the same change.

**Existing behavior to preserve:**
- The published-post delete path (`useDeletePostWizard` / the `delete-post-modal`) is a separate account-level flow — the cross-workspace option is **out of scope** there for this story; keep that path unchanged.
- Optimistic cache update (`removePlansFromQueryData`) removes the row from the current workspace's list; the other workspaces aren't loaded in the current view, so no extra cache work is needed for them.

**Suggested names (not binding):**
- Plan fields: `content_category_post_group_id` (shared correlation id), `content_category_origin_workspace_id` (or a boolean `is_content_category_origin`).
- Plan payload flags for the client: `is_global_content_category_post`, `is_content_category_origin`, `other_workspaces_count`.
- Remove payload flag: `delete_across_workspaces` (boolean).

**Gotchas:**
- Siblings currently share only the *global* category, indirectly, each via its own workspace's `content_category_id` — there is **no** direct plan-to-plan link and **no** stored origin marker. Correlation must be introduced; matching purely on `global_content_category_id` would sweep up *every* category post ever created in those workspaces, not the specific sibling set.
- Sibling execution times differ per workspace (each category computes its own slot via `PublishTimeFactory`), so don't correlate on `execution_time`.
- Disable the confirm button on `isPending` (double-submit guard) — the existing quick-confirm doesn't have to, but a custom modal does.

---

## Shortcut fields

### [Full Stack] Delete global content-category posts across all workspaces from the planner
- **Template:** New Feature Template
- **Story type:** feature
- **Project:** Web App
- **Group:** Full Stack
- **Epic:** Q2 - 2026: Miscellaneous (id: `115078`) — *PO: confirm the current-quarter Miscellaneous epic; config lists Q2-2026.*
- **Priority:** Medium
- **Product Area:** Publishing
- **Skill Set:** Backend + Frontend
- **Estimate:** _(empty — set during sprint planning)_
- **Labels:** _(none)_
- **Iteration:** _(PO assigns current/target sprint)_
