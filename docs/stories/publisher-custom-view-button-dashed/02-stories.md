# Publisher Sidebar — Dashed-Border "Create New View" Button · Story

**Platform:** Web only. **Scope:** Frontend (styling). No backend, no mobile.

| # | Story | Priority |
|---|---|---|
| S-1 | [FE] Restyle the "Create New View" button in the publisher sidebar with a dashed border | Low |

---

## S-1 · [FE] Restyle the "Create New View" button in the publisher sidebar with a dashed border
**Project:** Web App · **Group:** Frontend · **Skill:** Frontend · **Product area:** Planner · **Priority:** Low · **Type:** Feature

### Description
As a user managing custom views in the publisher sidebar, I want the "Create New View" action to look like a distinct "add" affordance — a dashed-border button — so that it clearly stands apart from my existing saved views instead of blending in as just another row.

Today "Create New View" renders as a plain text row with a small plus icon, visually similar to the saved-view rows above it, so it's easy to miss.

### Workflow
1. The user opens the publisher sidebar and scrolls past their list of saved custom views.
2. Below the separator, the user sees the **"Create New View"** action presented as a **dashed-border button** that clearly reads as "add something new."
3. The user clicks it and the create-custom-view drawer opens (unchanged behavior).

### Acceptance criteria
- [ ] The "Create New View" action in the publisher sidebar is styled as a **dashed-border** button so it visually stands out from the saved-view rows.
- [ ] The existing label (**"Create New View"**) and the plus icon are retained; no copy changes.
- [ ] The dashed border, its color, hover state, and the icon/text color use **theme-aware classes** (e.g., `border-primary-cs-*`, `text-primary-cs-*`, `bg-primary-cs-50`) — the existing hardcoded values (`text-[#2294ff]`, `hover:bg-[#E8F4FF]`) are replaced so the button adapts to white-label themes.
- [ ] Clicking the button still opens the create-custom-view drawer (behavior unchanged).
- [ ] The "Learn more" help icon next to the action is preserved and still opens its help article.
- [ ] Hover/focus states are visible and the dashed treatment remains legible on hover.
- [ ] The change is contained to the "Create New View" action — the saved-view rows above it are not restyled.

### Mock-ups
N/A — exact dashed-border treatment (weight, radius, spacing, color) to be confirmed with the Product Designer before implementation.

### Impact on existing data
None — presentational change only.

### Impact on other products
None. Desktop web publisher sidebar only; no mobile or Chrome extension impact. Uses theme tokens so white-label domains are unaffected.

### Dependencies
- **Consult the Product Designer** to confirm the exact dashed-border style (border weight, radius, color, padding, hover state) before implementing.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support — N/A, no copy change (reuses existing `planner.planner_custom_view.sidebar.create_new_view`)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry point:**
- `contentstudio-frontend/src/modules/publisher/components/SidebarMain.vue` — the "Create New View" action lives in the "Separator and Create New View" block (around the `planner.planner_custom_view.sidebar.create_new_view` label, with the `CirclePlus` icon and the `handleCreatePlannerCustomView` click handler). Apply the dashed-border styling to this row's container.

**Theming cleanup to fold in:**
- Replace the hardcoded `text-[#2294ff]` on the icon and `hover:bg-[#E8F4FF]` on the row with theme-aware classes per the frontend CLAUDE.md (no hardcoded brand hex).

**No new behavior:**
- `handleCreatePlannerCustomView` and the learn-more beacon already work — this is style-only.
