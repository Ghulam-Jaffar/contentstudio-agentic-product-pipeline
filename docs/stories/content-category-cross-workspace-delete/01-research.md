# Research — Delete global content-category posts across all workspaces

## Feature request

A post created from a **global content category** is automatically scheduled into **every workspace** where that category exists. When a user deletes such a post from the workspace where it was originally created, they should be offered the option to **also delete the corresponding sibling posts from all the other workspaces** where it was created.

---

## Current State

### How global content categories fan out
- A **global content category** lives in the `global_content_categories` collection (`App\Models\Settings\GlobalContentCategories`). Each workspace where it exists gets its **own** `content_categories` document linked back via `global_content_category_id`.
- `GlobalContentCategoriesController::store` creates the global category, then loops over **every workspace the user has access to** (`WorkspaceTeamRepo::getWorkspaceIdsByUserId`) and creates one workspace-level `content_categories` doc each, all sharing the same `global_content_category_id`.

### How one authoring action becomes N posts
- `App\Jobs\SocialContentCategoryPostJob` is the fan-out job. Given the category the user posted into, it resolves the `global_content_category_id`, fetches **all linked workspace categories** (`ContentCategoriesRepository::getByGlobalCategoryId`), and creates one plan (post) **per workspace**.
- The **origin** workspace (the one in the job payload) keeps its real `planId`; every other workspace gets `planId = ""`, so a brand-new plan is created there.
- Each sibling plan is stored with its **own workspace's** `content_category_id`, its own `execution_time` (each category computes its own slot), and `publish_time_options.time_type = 'content_category'`.

### How a post is deleted today (single workspace only)
- **Frontend:** `usePlannerActions.removePlan` (planner_v2) → `useDeletePlanMutation` → `removePlanApi` → `POST publisher/planner/removePlan`.
  - Non-published posts: a simple confirm via `$cstuModal.msgBoxConfirm` (copy keys `planner.planner_actions.delete_post_confirm` / `remove_post_title`). **This dialog has no checkbox slot.**
  - Published posts: a dedicated `delete-post-modal` (account-level delete wizard, `useDeletePostWizard`).
- **Backend:** `PlanController::removePlan` → `PlansRepository::removePlans(['plan_ids' => [...], 'workspace_id' => ...])`. **Every delete query is scoped to a single `workspace_id`.** It also refills content-category queue slots (`ContentCategoriesRepository::fillRemovedSlots`) and dispatches `PlanRemoveJob`.

### The core gap
- **Nothing links the sibling posts together.** Siblings share only the *global category* (indirectly, each via its own workspace `content_category_id`). There is **no shared batch/group id** on the plans and **no stored marker for which workspace was the origin**. So today there is no reliable way to say "these N plans across N workspaces were created by the same authoring action."
- The frontend plan payload exposes `content_category` / `content_category_id`, but **no flag** telling it the category is *global* or that the current workspace is the *origin*.

---

## What Needs to Change

**Backend**
- Give sibling posts a reliable correlation so a delete can find them: stamp a shared identifier (and an origin marker) on every plan `SocialContentCategoryPostJob` creates in the same run.
- Extend the remove-post flow so a caller can opt to also delete the sibling posts in the **other** workspaces where the global category exists — deleting **only** in workspaces the user is a member of / has post-delete permission in.
- Expose on the plan (detail/list) whether it is a **global content-category post** and whether the current workspace is the **origin**, so the UI can decide when to show the option.

**Frontend**
- In the delete confirmation for a post that is a global-category post **created in the current (origin) workspace**, show an extra opt-in: "Also delete this post from all other workspaces where this category exists." Pass the choice to the delete API.
- The current `msgBoxConfirm` has no checkbox — this needs a small custom confirmation modal (`Modal` + `Checkbox`, both available).

---

## Mobile Context
Not applicable. Global content categories (and the cross-workspace fan-out) are a web-only power feature configured in web Settings, and the planner delete-confirmation option lives in the web planner. The new delete parameter is **optional and backward-compatible**, so existing mobile/API callers keep today's single-workspace delete behavior. No iOS/Android stories.

---

## Files Involved

**Backend (`contentstudio-backend/`)**
- `app/Jobs/SocialContentCategoryPostJob.php` — fan-out creation; where a shared group id + origin marker would be stamped
- `app/Http/Controllers/Planner/PlanController.php` (`removePlan`) — delete entry point to extend
- `app/Repository/Publish/Planner/PlansRepository.php` (`removePlans`, `savePlanByCreate`, `fetchContentCategoriesPlans`) — single-workspace-scoped delete + plan persistence
- `app/Repository/Settings/ContentCategoriesRepository.php` (`getByGlobalCategoryId`, `fillRemovedSlots`) — resolves linked workspace categories
- `app/Models/Settings/GlobalContentCategories.php`, `app/Repository/Settings/WorkspaceTeamRepo.php` — global-category ↔ workspace membership
- `routes/web/planner.php` — `removePlan` route

**Frontend (`contentstudio-frontend/`)**
- `src/modules/planner_v2/composables/usePlannerActions.ts` (`removePlan`) — confirm + mutation trigger
- `src/modules/planner_v2/queries/usePlannerMutations.ts` (`useDeletePlanMutation`) — delete mutation
- `src/api/planner.ts` / `src/config/planner-api-utils.ts` (`removePlanURL`, `RemovePlanPayload`) — API contract
- `src/modules/planner_v2/composables/useDeletePostWizard.ts` — published-post delete wizard (parallel path)
- `src/locales/*/planner.json` — confirmation copy
