# Stories: Planner Note Popover — Add Template Option via Split Button

---

## Story 1: [FE] Add template option to planner note popover compose button

### Description:

As a ContentStudio user, I want to use a template when composing from a planner note so that I can start from a pre-built template instead of always using the note's content as the caption.

Currently, when I click a note in the planner calendar view, the popover shows a "Social Post" button that opens the composer with the note content pre-filled. Every other compose entry point in the app (Publisher sidebar, Dashboard, calendar day cell) offers both "Social Post" and "Use Template" via a split button dropdown — but the note popover does not.

This story converts the plain "Social Post" button in the note popover into a split button with a dropdown containing the template option, matching the pattern already used in the Publisher sidebar (`SidebarMain.vue`) and Dashboard (`WelcomeRow.vue`).

**Key behavior difference:** When the user clicks "Social Post", the note's title and description are passed as the composer caption (existing behavior). When the user picks "Use Template", the note content is **not** passed — the template supplies its own content.

**Implementation context:**
- The note popover is currently raw HTML built via template literal in `contentstudio-frontend/src/modules/planner_v2/views/CalenderView.vue` (~line 2040-2115), with click handlers attached via `addEventListener` (~line 2356-2383)
- The existing `SplitButton` from `@contentstudio/ui` and `ComposeActionsDropdown` from `contentstudio-frontend/src/modules/publisher/components/ComposeActionsDropdown.vue` are the reference pattern
- The `ComposeActionsDropdown` already includes the `TemplateAttachment` component and emits `use-template`
- Ideally, the popover action bar should be refactored to use Vue components (or at minimum, the split button dropdown behavior should be replicated to work within the existing raw HTML popover)

---

### Workflow:

1. User clicks a note in the planner calendar view
2. The note popover opens showing the note title, description, author, and action buttons
3. User sees the compose button now has a small dropdown arrow next to it (split button pattern)
4. User clicks the main button area ("Social Post") — the composer opens with the note's content pre-filled as caption, same as today
5. Alternatively, user clicks the dropdown arrow to reveal a menu with a "Use Template" option
6. User clicks "Use Template" — the template picker opens
7. User selects a template — the composer opens with the template's content instead of the note's content
8. The popover closes after either action

---

### Acceptance criteria:

- [ ] The "Social Post" button in the note popover is converted to a split button with a dropdown chevron
- [ ] Clicking the main button area opens the composer with the note content (title + description) pre-filled as caption — existing behavior preserved
- [ ] Clicking the dropdown chevron reveals a menu with a "Use Template" option
- [ ] Clicking "Use Template" opens the template picker
- [ ] When a template is selected, the composer opens with the template's content — the note's content is not passed
- [ ] The split button visually matches the compose split button pattern used in the Publisher sidebar and Dashboard
- [ ] The popover closes after the user clicks either "Social Post" or selects a template
- [ ] The note date and selected planner accounts are still passed correctly when composing via "Social Post"
- [ ] The split button only appears for non-Approver roles (same visibility rule as the current "Social Post" button)
- [ ] The dropdown closes if the user clicks outside of it

---

### Mock-ups:

N/A — follow the exact same split button + dropdown pattern used in the Publisher sidebar compose button (`SidebarMain.vue:13-37`). The main button shows a compose icon + "Social Post" label, and the dropdown chevron reveals a menu with the template option.

---

### UI Copy

- **Main button label:** `Social Post` (existing, no change)
- **Main button tooltip:** `Compose a new social post using this note's content` (existing key: `planner.calendar_view.note_popover.actions.compose_tooltip` — update if needed)
- **Dropdown option — Use Template:**
  - Label: `Use Template`
  - Subtext: `Start from a saved template`

---

### Impact on existing data:

No data or schema changes. This is a UI-only change to the note popover.

---

### Impact on other products:

- Web App: Planner calendar view note popover only
- Mobile apps: No impact
- Chrome extension: No impact

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (split button and dropdown should remain usable on smaller widths within the popover)
- [ ] Multilingual support (new dropdown option label and subtext use i18n — add keys to all locale directories)
- [ ] UI theming support (reuse existing themed split button pattern — use CSS variable-backed colors, no hardcoded hex)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
