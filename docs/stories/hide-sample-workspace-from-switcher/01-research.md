# Research — Hide the sample workspace from the workspace switcher

Let users hide the **sample (demo) workspace** from the workspace-switcher dropdown, via a three-dot option on the sample workspace's tile in Manage Workspaces.

## Current State

**Sample workspace identity**
- `src/utils/workspaceKind.ts` + `src/utils/demoWorkspace.ts` — `isSampleWorkspace(...)`, `SAMPLE_WORKSPACE_KIND = 'sample'`. Reliable way to detect the sample workspace anywhere.

**Manage Workspaces tiles ("where all workspaces are shown")**
- `src/modules/setting/components/workspace/ManageWorkspacesMain.vue` → `Listing.vue` → **`reusable/WorkspaceActiveTile.vue`** (the per-workspace tile).
- The tile's three-dot "more options" `Dropdown` is **currently suppressed for the sample workspace**: `v-if="!isLocked && !isSample"` (WorkspaceActiveTile.vue ~line 99). Its items today: Pause/Resume posting, **Manage workspace** (settings), Remove workspace.
- ⇒ To add a "Hide from switcher" option on the sample tile, we must **un-suppress a menu for the sample workspace** (either relax the `v-if` and gate the existing items, or give the sample tile its own minimal menu).

**Workspace switcher dropdown ("the dropdown of workspaces")**
- `src/components/layout/WorkspaceSwitcherDropdown.vue` renders `filteredWorkspaces` and shows a **"Sample" status badge** for the sample workspace (`isSampleWorkspace(item.workspace)`).
- `filteredWorkspaces` is computed in **`useWorkspaceCore()`** and passed down via `TopHeaderBar.vue` / `DesktopNavigationRail.vue`. ⇒ The filter that excludes a hidden sample workspace belongs here.

**Persistence**
- There is **no existing hide/favorite/pin workspace mechanism.** Hiding needs a **new persisted flag**, so this is not purely frontend — a backend change is required to store the flag and return it on the workspace/list payload.

## What Needs to Change
1. **Manage Workspaces tile** — show a three-dot menu on the **sample workspace** tile (today it's hidden for sample) with a **"Hide from workspace switcher"** action, and its inverse **"Show in workspace switcher"** once hidden.
2. **Switcher dropdown** — exclude the sample workspace from `filteredWorkspaces` when it's marked hidden (it still appears in Manage Workspaces so it can be un-hidden).
3. **Persistence (BE)** — store the "hidden from switcher" flag and return it on the workspace payload so both surfaces honor it.

## Open questions (for the review gate)
- **Scope of the flag: per-user or org-wide?** Since the sample workspace is shared, hiding is most naturally a **per-user** preference (each teammate declutters their own switcher). Org-wide (hidden for everyone, owner-only action) is simpler but less flexible. **Recommend per-user.** This decides the BE shape.
- **Un-hide path:** from the same sample-workspace tile in Manage Workspaces (the tile always shows there). Confirm that's the intended way back.
- **Mobile:** the switcher UI and Manage Workspaces are web. Treat as **web-only** (the flag is returned by the API so mobile could honor it later)?

## Story split
Needs both BE (persist + expose the flag) and FE (tile menu + dropdown filter). Given the size and your preference, this fits **one `[Full Stack]` story** — confirm at the gate.

## Files Involved
- `src/modules/setting/components/workspace/reusable/WorkspaceActiveTile.vue` — un-suppress/gate the menu for the sample tile; add the hide/show action.
- `src/components/layout/WorkspaceSwitcherDropdown.vue` — (display only; already flags the sample workspace).
- `src/composables/useWorkspaceCore.ts` — `filteredWorkspaces` computed; apply the hidden filter here.
- `src/composables/useWorkspaceSwitcher.ts` / `src/utils/demoWorkspace.ts` — sample detection helpers to reuse.
- **Backend** — a workspace (or per-user workspace-preference) endpoint + flag on the workspace/list response. *(Confirm `contentstudio-backend` paths during BE work.)*
