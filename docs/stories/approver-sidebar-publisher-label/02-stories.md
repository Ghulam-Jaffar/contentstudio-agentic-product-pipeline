# Approver Sidebar — Rename "View Content" to "Publisher" · Story

**Platform:** Web only. **Scope:** Frontend (label/copy). No backend, no mobile.

| # | Story | Priority |
|---|---|---|
| S-1 | [FE] Rename the approver desktop sidebar item from "View Content" to "Publisher" | Low |

---

## S-1 · [FE] Rename the approver desktop sidebar item from "View Content" to "Publisher"
**Project:** Web App · **Group:** Frontend · **Skill:** Frontend · **Product area:** Throughout product · **Priority:** Low · **Type:** Feature

### Description
As an approver, I want the desktop sidebar item that takes me to the planner to be called **"Publisher"** — the same name everyone else sees — so that the navigation is consistent and I'm not confused by a different label ("View Content") for the same destination.

Today, approver-only users see a single sidebar item labeled "View Content"; it routes to the planner just like the standard "Publisher" item does for other roles.

### Workflow
1. An approver opens the app and looks at the desktop left sidebar.
2. The single navigation item now reads **"Publisher"** (previously "View Content").
3. Clicking it opens the planner exactly as before.

### Acceptance criteria
- [ ] For approver-only users, the desktop sidebar navigation item label reads **"Publisher"** instead of "View Content".
- [ ] The item's hover tooltip also reads **"Publisher"**.
- [ ] Clicking the item still opens the planner (destination/behavior unchanged).
- [ ] The label change reuses the existing "Publisher" navigation copy so it stays consistent across all locales.
- [ ] No change for non-approver roles (they already see "Publisher").

### Mock-ups
N/A — copy change only.

### Impact on existing data
None — label/copy change only.

### Impact on other products
None. Desktop web sidebar only; no mobile or Chrome extension impact. The "Publisher" label is already translated, so all locales stay consistent.

### Dependencies
None.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support — reuses the existing translated "Publisher" key; no new strings
- [ ] UI theming support — N/A, copy-only change
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry point:**
- `contentstudio-frontend/src/components/layout/useHeaderNavigation.ts` — the approver-only branch of `basePrimaryNavItems` (rendered when `!canAccessPrimaryNavigation`) builds the `view-content` item with `labelKey: 'header.nav.view_content'`. Point its `labelKey` to the existing `header.nav.publisher` ("Publisher") so label and tooltip both update. The `id`, icon, and `to` (planner route) stay as-is.

**Note:**
- `header.nav.publisher` already exists and equals "Publisher" in every locale — no new i18n keys needed. The now-unused `header.nav.view_content` key can be left or removed.
