# Stories — Workspace dropdown header refresh

---

## [FE] Replace the workspace dropdown "+" icon with a "New" pill and rename the Manage entry to "All Workspaces"

### Description:

As a ContentStudio user managing several workspaces, I want the workspace dropdown's create action to be a clearly labelled "New" button and the manage entry to say "All Workspaces", so that I can tell at a glance how to add a workspace and where to go to see every workspace I belong to — without hovering over an unlabelled icon to find out what it does.

Today the create-workspace control in the workspace dropdown is a bare circular "+" icon sitting immediately after the "Your Workspaces" label. Its meaning is only revealed on hover, which never happens on touch devices, and it competes visually with the "View All" link on the opposite side of the same row. At the same time the Manage section's first entry is titled simply "Workspaces", which reads as a duplicate of the "Your Workspaces" heading directly above it rather than as the way to reach the full workspace list.

This story restyles the header row of the dropdown and renames one Manage entry. No new capability is added — who can create a workspace, and where each link goes, stays exactly as it is today.

---

### Workflow:

1. User clicks their workspace avatar at the top of the left sidebar and the workspace dropdown opens.
2. User sees the header row above the workspace list: the heading **"Your Workspaces"** on the left with the total workspace count in a small rounded badge directly beside it, and a **"New"** pill button on the right-hand end of the same row.
3. The "New" pill shows a plus icon followed by the word "New", on a light tint of the workspace's primary colour with matching primary-coloured text and fully rounded ends.
4. User hovers the pill and sees a tooltip explaining what a workspace is for.
5. User clicks the pill and the existing create-workspace flow opens, exactly as clicking the old "+" icon did.
6. If the user does not have permission to create workspaces, the pill is not shown at all and the header row displays only the heading and its count badge.
7. User scrolls to the **Manage** card at the bottom of the dropdown and sees its first entry now titled **"All Workspaces"**, keeping its existing grid icon, its subtext "View & manage all your workspaces", and its chevron.
8. User clicks "All Workspaces" and lands on the same workspace management page it has always opened, and the dropdown closes.

---

### Acceptance criteria:

**Header row — "New" pill**

- [ ] The circular, icon-only "+" button no longer appears anywhere in the workspace dropdown
- [ ] A pill-shaped button labelled **"New"** with a leading plus icon appears at the right-hand end of the header row that contains the "Your Workspaces" heading
- [ ] The pill uses the `Button` component from `@contentstudio/ui` with `variant="light"`, `color="primary"` and `radius="full"` — it is not a hand-rolled button and its colours are not overridden with hardcoded Tailwind classes
- [ ] The pill's text and tint follow the workspace's primary theme colour, so a white-label domain with a non-blue primary colour renders the pill in that colour
- [ ] Hovering the pill shows the tooltip: *"Create a new workspace — a separate space for another brand or client, with its own social accounts, posts and team members."*
- [ ] The pill's accessible label reads "Create a new workspace" for screen-reader users
- [ ] Clicking the pill opens the same create-workspace flow that the old "+" icon opened
- [ ] The pill is shown only to users who are allowed to create a workspace; users without that permission (including collaborators) see the header row with no pill, and the row's spacing stays visually correct with the pill absent

**Header row — count badge**

- [ ] The workspace count badge appears immediately after the "Your Workspaces" heading, on the left of the header row
- [ ] The badge shows the user's total number of workspaces and updates when that total changes
- [ ] The badge is plain, non-clickable text — it no longer links anywhere

**Header row — "View All" removal**

- [ ] The "View All" link is removed from the header row of the workspace dropdown
- [ ] No empty space, stray separator, or misalignment remains where the link used to be
- [ ] Users can still reach the full workspace list from the "All Workspaces" entry in the Manage section

**Manage section rename**

- [ ] The first entry in the Manage section is titled **"All Workspaces"** instead of "Workspaces"
- [ ] Its subtext still reads "View & manage all your workspaces"
- [ ] Its icon, its chevron, its position as the first of the three Manage entries, and the page it opens are all unchanged
- [ ] The entry remains visible only to the users who can see it today — users without workspace-management access still do not see it
- [ ] The "Team & roles" and "Limits" entries are untouched in title, subtext, icon and destination

**Translations**

- [ ] The pill label, the pill tooltip, and the "All Workspaces" title all come from translation keys — no hardcoded English in the interface
- [ ] Every supported language (English, German, Greek, Spanish, French, Italian, Polish, Chinese) has a translation for the pill label that reads correctly as a standalone button, not as a fragment of a longer phrase — several languages currently store a partial value that only made sense when the text was hidden from view
- [ ] Every supported language has a translation for the renamed "All Workspaces" title
- [ ] In the language with the longest pill label, the header row still fits on one line inside the dropdown with no text wrapping, clipping, or overlap with the heading and its count badge

---

### Mock-ups:

The attached mockup shows the target header row: **"Your Workspaces"** with the count badge beside it on the left, and the light-tinted **"+ New"** pill on the right, with the Manage card below showing "All Workspaces", "Team & roles" and "Limits".

---

### Impact on existing data:

None. No data model, stored setting, or API response changes. This is a presentation and copy change to an existing dropdown, and every destination and permission rule behaves as it does today.

---

### Impact on other products:

- **Mobile app (Flutter):** No impact. The mobile app has its own workspace switcher and is unaffected by this web header-row change. No `[Flutter]` story is needed.
- **Chrome extension:** No impact — it does not render the workspace dropdown.
- **White-label domains:** In scope and must be verified. The pill's background tint and text colour are driven by the primary theme colour, so it must be checked on a white-label domain with a non-default primary colour to confirm the label stays legible against its own tint.

---

### Dependencies:

None. This story is self-contained and can ship on its own.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
