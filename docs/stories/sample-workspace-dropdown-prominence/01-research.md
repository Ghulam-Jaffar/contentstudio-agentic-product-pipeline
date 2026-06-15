# Research — Make the Sample Workspace prominent in the workspace switcher

> Quick-story pipeline · Step 1 · FE-only

## What the user asked for

Make the **Sample Workspace** stand out wherever a user picks a workspace:

1. **Always pin it to the top** of the workspaces dropdown (not buried in the alphabetical list).
2. In the dropdown, show a **"Sample Workspace"** label on the right of its row so it's clearly identified.
3. On the **"View all" workspaces screen**, give the Sample Workspace's pill a **solid colour** (vs the translucent status pills) so it's more prominent.

## Current State

**The workspaces dropdown (header)**
- Lives in [TopHeaderBar.vue](contentstudio-frontend/src/components/layout/TopHeaderBar.vue) and the rail variant in [DesktopNavigationRail.vue](contentstudio-frontend/src/components/layout/DesktopNavigationRail.vue).
- Each row = round logo + workspace name + (on the right) a green check for the active workspace, or a lock for restricted ones.
- Footer has **"+ Add workspace"** and a **"View all"** link that routes to `{ name: 'workspaces' }`.

**Ordering**
- Built in [useWorkspace.js](contentstudio-frontend/src/modules/common/composables/useWorkspace.js) → `filteredWorkspacesOwnedByMe()`. Workspaces are split into *owned-by-me* vs *owned-by-others*, and **each group is sorted alphabetically** by name (`sortBy(arrayToUse, (item) => [item.workspace.name.toLowerCase()])`, line ~225). So today the Sample Workspace lands wherever its name falls alphabetically.

**The "View all" screen**
- Route `name: 'workspaces'` → [ManageWorkspacesMain.vue](contentstudio-frontend/src/modules/setting/components/workspace/ManageWorkspacesMain.vue), which renders a grid of [WorkspaceActiveTile.vue](contentstudio-frontend/src/modules/setting/components/workspace/reusable/WorkspaceActiveTile.vue) tiles (the "pills").
- A tile is a **white card** (`bg-white`, `border-border-default`) with the logo, name, and a small status pill on the right. Existing status pills are **translucent** — e.g. "Default" uses `bg-cstu-blue-450/10`, locked/paused use `bg-cstu-red-500/10`. There's no special treatment for the Sample Workspace today.

## What Needs to Change

- **Identify** the Sample Workspace reliably (see dependency below).
- **Dropdown ordering:** pin the Sample Workspace to the very top, above the alphabetical list (in both the header dropdown and the navigation-rail dropdown).
- **Dropdown label:** add a "Sample Workspace" badge on the right of the Sample Workspace's row.
- **View all screen:** give the Sample Workspace tile a prominent **solid** "Sample Workspace" pill (filled primary colour, white text) instead of the translucent style used by the other status pills.

## ⚠️ Dependency / open question — how is the Sample Workspace identified?

There is **no `is_sample` / `is_demo` flag** anywhere today:
- Not on the workspace object returned to the frontend (`ActiveWorkspace` interface in [useWorkspaceStore.ts](contentstudio-frontend/src/stores/setting/useWorkspaceStore.ts) has `_id, name, slug, timezone, logo, on_hold, onboarding, …` — no sample flag).
- No "Sample Workspace" literal or `is_sample` field in the backend `Workspace` model.

The FE work needs a **reliable identifier** for "this is the sample workspace." Matching on the name string is fragile (breaks with translations, white-label renaming, or a user renaming it). The clean solution is a boolean flag on the workspace object (e.g. `is_sample: true`) exposed by the API.

**This makes the FE story depend on a small backend addition** if that flag doesn't already exist. Flagged for the user to confirm at the review gate.

## UI components available (from catalog)

- **`Badge`** (@contentstudio/ui) — the right component for the "Sample Workspace" label/pill in both surfaces. Need to confirm it exposes a **solid/filled** variant in the theme primary colour; if not, that's a small gap to flag.
- Colours must be **theme-aware** (`bg-primary-cs-500`, white text) per the theming rules — never hardcode a blue.

## Mobile Context

- The workspace switcher also exists in the iOS/Android apps, but this is a **web FE** request. Web-only this round (consistent with the other Sample Workspace work). Mobile prominence could be a later follow-up.

## Files Involved (grounding only — kept out of the story body)

- `contentstudio-frontend/src/modules/common/composables/useWorkspace.js` — `filteredWorkspacesOwnedByMe()` sort (pin sample first here)
- `contentstudio-frontend/src/components/layout/TopHeaderBar.vue` — header workspace dropdown rows
- `contentstudio-frontend/src/components/layout/DesktopNavigationRail.vue` — rail workspace dropdown rows
- `contentstudio-frontend/src/modules/setting/components/workspace/ManageWorkspacesMain.vue` — "View all" grid
- `contentstudio-frontend/src/modules/setting/components/workspace/reusable/WorkspaceActiveTile.vue` — the tile/pill
- `contentstudio-frontend/src/stores/setting/useWorkspaceStore.ts` — workspace object shape (no sample flag today)
