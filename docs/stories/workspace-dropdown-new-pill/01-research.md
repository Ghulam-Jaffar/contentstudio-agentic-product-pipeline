# Research — Workspace dropdown: "+ New" pill & "All Workspaces" rename

## Current State

The workspace switcher dropdown opens from the workspace avatar at the top of the left sidebar. It has four regions, top to bottom:

1. **Search bar** — `SearchInput`, placeholder "Search a workspace…"
2. **Header row** — left: the label "Your Workspaces" followed immediately by a bare circular `+` icon button; right: a "View All" text link with a count `Badge` next to it
3. **Workspace list** — scrollable, each row = logo, name, role, optional "Sample Workspace" tag, checkmark on the active one
4. **Manage section** — a bordered card titled "Manage" holding three rows, each with icon + title + subtitle + chevron:
   - **Workspaces** — "View & manage all your workspaces" → `workspaces?tab=manage_workspaces`
   - **Team & roles** — "Manage members across workspaces" → `workspaces?tab=manage_team`
   - **Limits** — "Manage members, socials & other limits" → `workspaces?tab=manage_limits`

### Today's header row, precisely

- The create control is a 20px circular icon-only button rendering the `CirclePlus` Lucide icon at 18px, in primary colour, with a primary-50 hover fill. It carries the tooltip **"Create a new workspace"** and the aria-label **"Add New"**. Clicking it emits `create-workspace` to the parent, which opens the create-workspace flow.
- It is only rendered when the user has the `can_save_workspace` permission — **collaborators and other users without that permission never see it.**
- The count badge currently lives on the right, paired with the "View All" link, not with the "Your Workspaces" label.

### Manage section permissions

- **Workspaces** and **Limits** render only for users with `can_access_workspace_setting` whose role is not `collaborator`.
- **Team & roles** renders only for users with `can_view_team`.
- The whole Manage card is hidden when none of the three qualify.

## What Needs to Change

### 1. `+` icon → "+ New" pill
- Replace the bare circular icon button with a pill-shaped button labelled **"+ New"** — a plus icon followed by the word "New", on a light primary-tinted background with primary-coloured text and fully rounded corners.
- Per the mockup the pill sits at the **right end** of the header row, not tucked beside the "Your Workspaces" label as the icon is today.
- Same behaviour, same permission gate, same click action — only the presentation and position change.

### 2. Manage → "Workspaces" renamed to "All Workspaces"
- Title-only change on the first Manage row. Its icon, subtitle, destination, and permission gate are untouched.
- Renames the label users read; it does not affect the "View All" link elsewhere.

### 3. Header row restructure (PO-confirmed 2026-08-21)
The mockup's header row reads `Your Workspaces [71]` on the left and `[+ New]` on the right. Both knock-on changes are confirmed in scope:
- **The count badge moves** from the right (next to "View All") to sit directly after the "Your Workspaces" label.
- **The "View All" link is removed** from the header row. Its destination overlaps with the renamed "All Workspaces" Manage row, which now carries that navigation.

## UX Reference

Labelled pill actions ("+ New", "+ Add") in list headers are the common pattern in workspace/team switchers (Slack, Linear, Notion) because a bare `+` glyph is ambiguous next to a heading — the word carries the action where the icon alone relies on a hover tooltip that touch users never see.

## Localisation

`add_new` and `manage_workspaces_title` already exist in all 8 locale files (`en`, `de`, `el`, `es`, `fr`, `it`, `pl`, `zh`). Current `add_new` values are inconsistent — several locales hold only a fragment ("nuevo", "nouveau", "Νέου", "nuevo") that reads correctly only after a word the old design never showed. Because the pill now displays this string as standalone visible text rather than as a hidden aria-label, every locale's value needs a review pass, not just `en`. Same for `manage_workspaces_title` across all 8.

If "View All" is removed, its `view_all` key becomes unused in this component.

## Mobile

Not applicable. The Flutter app has its own workspace switcher; this is a web-only header-row presentation change and no mobile parity work is implied.

## Files Involved

- `contentstudio-frontend/src/components/layout/WorkspaceSwitcherDropdown.vue` — header row markup (create button, badge, View All link)
- `contentstudio-frontend/src/composables/useWorkspaceSwitcherMenu.ts` — `manageItems`, holds `manage_workspaces_title`
- `contentstudio-frontend/src/locales/{en,de,el,es,fr,it,pl,zh}/header.json` — `header.workspace.add_new`, `header.workspace.manage_workspaces_title`
- **No component gap.** `docs/ui-components.md` notes there is no dedicated pill/chip *component*, but that entry is about static chips. For a clickable pill the existing `@contentstudio/ui` `Button` already covers it — its typings expose `variant="light"`, `color="primary"`, `size="xs"|"sm"`, and `radius="full"`, which is exactly the mockup's light-tinted, fully-rounded pill. The same combination is already in use elsewhere in the app (e.g. `LabelAttachment.vue`). No `[Design]` story or library change is required.
