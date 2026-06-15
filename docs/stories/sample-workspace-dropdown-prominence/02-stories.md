# Stories — Make the Sample Workspace prominent in the workspace switcher

> Quick-story pipeline · Step 2 · Deliverable for the Product Owner to create in Shortcut manually. Nothing is pushed to Shortcut.
>
> **Scope decisions (confirmed):** one FE story · the Sample Workspace is identified via a backend `is_sample` flag (captured as a dependency) · assign to the existing **Sample Workspace** epic · web-only.

---

## [FE] Make the Sample Workspace stand out in the workspace switcher and "View all" screen

### Description

As a **user with multiple workspaces (especially a new or trial user)**, I want the **Sample Workspace** to be easy to spot — pinned to the top of the workspace switcher, clearly labelled, and visually prominent on the "View all" screen — so that I can quickly find the demo workspace to explore the product, and never confuse it with my own real workspaces.

Today the Sample Workspace is sorted alphabetically like any other workspace, so it can be buried in the list, and nothing marks it as the demo one. This makes it harder for new users to find the place where they're meant to explore the product, and it's easy to mistake it for a real workspace.

This story makes the Sample Workspace consistently recognisable in both places a user picks a workspace: the **header workspace dropdown** and the **"View all" workspaces screen**.

---

### Workflow

1. A user opens the **workspace switcher** in the header (or the navigation-rail variant).
2. The **Sample Workspace appears at the very top** of the list — above the normal alphabetical ordering of their other workspaces.
3. Its row shows a **"Sample Workspace" badge** on the right, so the user immediately knows it's the demo workspace.
4. The user clicks **"View all"** at the bottom of the dropdown and lands on the workspaces screen showing every workspace as a tile.
5. The Sample Workspace's tile carries a **solid-coloured "Sample Workspace" pill** (filled with the app's primary colour), making it stand out clearly against the other tiles, whose status pills are lighter/translucent.
6. The user can tell at a glance which workspace is the sandbox for exploring, and which are their real ones.

---

### Acceptance criteria

**Pin to top (dropdown ordering)**
- [ ] In the header workspace dropdown, the Sample Workspace always appears as the first item, above all other workspaces, regardless of its name's alphabetical position.
- [ ] The same pinning applies in the navigation-rail workspace dropdown.
- [ ] The user's other workspaces keep their existing alphabetical ordering below the pinned Sample Workspace.
- [ ] If the account has **no** Sample Workspace, the list behaves exactly as it does today (no empty pinned slot, no placeholder).

**"Sample Workspace" label in the dropdown**
- [ ] The Sample Workspace's row shows a "Sample Workspace" badge on the right side of the row.
- [ ] The badge appears only on the Sample Workspace row — all other workspace rows are unchanged.
- [ ] The badge does not break the row layout when the workspace name is long (name truncates, badge stays visible).

**Solid pill on the "View all" screen**
- [ ] On the "View all" workspaces screen, the Sample Workspace's tile shows a "Sample Workspace" pill that is **solid/filled** (primary colour background, white text) — visually more prominent than the translucent status pills (e.g. "Default") on other tiles.
- [ ] The pill appears only on the Sample Workspace's tile.
- [ ] The pill's colour comes from the theme (adapts on white-label domains) and is not a hardcoded colour.

**General**
- [ ] Identifying the Sample Workspace relies on the workspace's sample flag from the API (see Dependencies), not on matching the workspace name text.
- [ ] No change to how any non-sample workspace looks or is ordered in either surface.

---

### UI copy

- **Badge / pill text:** "Sample Workspace"
- **Badge tooltip (on hover, both surfaces):** "A demo workspace filled with example data so you can explore the product. Your own workspaces hold your real content." *(brand-neutral — works on white-label domains)*

---

### Mock-ups

N/A — no mock-ups provided. Reuses the existing dropdown rows and "View all" tiles; only adds the badge/pill and changes ordering.

---

### Impact on existing data

- Display-only change. No change to workspace data. Relies on the API exposing which workspace is the sample one (see Dependencies).

---

### Impact on other products

- **Mobile apps (iOS/Android):** Out of scope this round (web-only). The mobile apps also have a workspace switcher — making the Sample Workspace prominent there could be a later follow-up.
- **Chrome extension:** N/A.
- **White-label domains:** In scope — the solid pill and badge must use the theme's primary colour (via component variant / CSS variables), so they adapt to each white-label brand. No hardcoded blue.

---

### Dependencies

- **Depends on:** **[BE] Expose an `is_sample` flag on the workspace object** — the frontend needs a reliable boolean on each workspace (in the workspaces list / active-workspace API response) to know which one is the Sample Workspace. Without it, the only option is matching the workspace name string, which is fragile across translations, white-label renaming, and user renames. *(If this flag already exists on the workspace payload, this dependency is already satisfied — wire the FE to it.)*

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only) — verify the badge/pill and pinned ordering render correctly across breakpoints in both the dropdown and the "View all" grid
- [ ] Multilingual support (frontend + backend, translations available or fallback handled) — "Sample Workspace" badge text and tooltip go through i18n with keys added to all locale directories
- [ ] UI theming support (default + white-label, design library components are being used) — solid pill uses the theme primary colour, not a hardcoded value
- [ ] White-label domains impact review — badge/pill colour adapts per brand (see Impact)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension) — web-only this round; mobile noted as a possible follow-up

---

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-frontend/src/modules/common/composables/useWorkspace.js` — `filteredWorkspacesOwnedByMe()` currently does `sortBy(arrayToUse, (item) => [item.workspace.name.toLowerCase()])`. Pin the sample workspace ahead of this alphabetical sort (e.g. sample-first comparator) so both dropdowns get the ordering for free.
- `contentstudio-frontend/src/components/layout/TopHeaderBar.vue` and `contentstudio-frontend/src/components/layout/DesktopNavigationRail.vue` — the workspace dropdown rows (logo + name + right-aligned active-check). Add the "Sample Workspace" badge to the right of the row for the sample workspace.
- `contentstudio-frontend/src/modules/setting/components/workspace/reusable/WorkspaceActiveTile.vue` — the "View all" tile. Existing status pills here are translucent (e.g. `bg-cstu-blue-450/10`); add the solid "Sample Workspace" pill for the sample tile.
- `contentstudio-frontend/src/modules/setting/components/workspace/ManageWorkspacesMain.vue` — the "View all" grid that renders the tiles.

**Components:**
- Use the **`Badge`** component (`@contentstudio/ui`) for the label/pill in both surfaces. For the "View all" pill, use its **solid/filled** variant in the theme primary colour. **Gap to confirm:** if `Badge` doesn't expose a solid primary-colour variant, flag a small library/Design adjustment rather than overriding its styles with Tailwind.

**Existing behavior to preserve (no change needed):**
- Owned-by-me vs owned-by-others split and the alphabetical ordering within each group stay as-is for non-sample workspaces.
- The active-workspace green check and the locked/on-hold states in the dropdown and tiles are unchanged.

**Gotcha:**
- There is **no `is_sample` field today** on the `ActiveWorkspace` shape (`src/stores/setting/useWorkspaceStore.ts`) or the backend `Workspace` model — the FE work is blocked until the flag is available (see Dependencies). Do not fall back to name-matching as the permanent solution.

---

### Shortcut fields

- **Template:** New Feature Template (PO selects this when creating the story so the standard sections + 5 quality-checklist tasks are pre-populated)
- **Story type:** Feature
- **Project:** Web App
- **Group:** Frontend
- **Epic:** Sample Workspace *(existing epic — confirm/select it when creating the story; falls back to **Q2 - 2026: Miscellaneous** (id: 115078) only if that epic doesn't exist)*
- **Priority:** Medium
- **Product Area:** Onboarding *(the Sample Workspace exploration experience; this change touches the workspace switcher and the "View all" workspaces screen)*
- **Skill Set:** Frontend
- **Estimate:** — (left empty; devs estimate during sprint planning)
- **Labels:** none (team manages labels manually)
- **Iteration:** PO assigns the current/target sprint at creation time

---

> **Analytics events (Usermaven):** N/A — this is a display/ordering/prominence change. It introduces no new trackable user action (switching workspaces is existing behaviour).
