# Research: Planner Note Popover — Add Template Option via Split Button

## Current State

When a user clicks a note in the planner calendar view, a popover appears (built with raw HTML in `CalenderView.vue`). The popover shows:
- Note title and description
- Author info
- Action buttons: **Social Post** (primary), Duplicate, Edit, More Options (delete)

The **Social Post** button is a plain `<button>` that opens the composer with the note's content (title + description) pre-filled as the caption. It also passes the note's date and selected planner accounts.

**Key code:** `CalenderView.vue:2077-2080` — the button is raw HTML injected via template literal (not a Vue component). The click handler is at `CalenderView.vue:2358-2383`.

### How it works elsewhere

In **Publisher sidebar** (`SidebarMain.vue:13-37`) and **Dashboard WelcomeRow** (`WelcomeRow.vue:17-45`), the compose button uses:
- `SplitButton` from `@contentstudio/ui` — clicking the main button opens the composer, clicking the dropdown chevron reveals a menu
- `ComposeActionsDropdown` inside the dropdown slot — shows "Social Post" + "Use Template" (via `TemplateAttachment` component) + optionally Blog Post / Bulk Schedule

The `ComposeActionsDropdown` component (`publisher/components/ComposeActionsDropdown.vue`) already handles template selection via the `TemplateAttachment` sub-component and emits `use-template`.

## What Needs to Change

1. **Replace the plain "Social Post" button** in the note popover with a split button pattern that has a "Use Template" dropdown option
2. **Social Post (main click):** Keep current behavior — opens composer with note content as caption, passes note date and selected accounts
3. **Use Template (dropdown option):** Opens the template picker — does NOT pass note content (template has its own content)
4. **Challenge:** The note popover is raw HTML (not Vue components), so implementing a proper `SplitButton` + `ComposeActionsDropdown` requires either:
   - Converting the popover to a Vue component (preferred, aligns with codebase standards)
   - Or building a simpler dropdown with raw HTML/JS (hacky, not recommended)

## Files Involved

- `contentstudio-frontend/src/modules/planner_v2/views/CalenderView.vue` — note popover HTML (~line 2040-2115) and click handlers (~line 2356-2383)
- `contentstudio-frontend/src/modules/publisher/components/ComposeActionsDropdown.vue` — existing dropdown with Social Post + Template options (reusable)
- `contentstudio-frontend/src/modules/publisher/components/SidebarMain.vue` — reference for how SplitButton + ComposeActionsDropdown is used
- `contentstudio-frontend/src/modules/composer_v2/components/TemplateAttachment.vue` — template picker component
- Locale files: `src/locales/*/planner.json` — for new i18n keys
