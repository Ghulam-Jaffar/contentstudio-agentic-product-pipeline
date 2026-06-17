# Desktop Sidebar — Customize (Show/Hide + Reorder) · Stories

**Platform:** Web only (desktop left rail). **Scope:** Frontend + Design. No backend (reuses the existing user-preferences endpoint), no mobile.

**Background:** Originally scoped as per-item hover pin/unpin. CEO feedback: the rail is too narrow for hover affordances. Instead, follow the customer.io pattern — a **"Customize sidebar"** entry in the **More** menu opens a panel where users **Show/Hide** items and **drag to reorder**. We improve on customer.io (which is show/hide only) with toggles + reordering + responsive overflow.

| # | Story | Priority |
|---|---|---|
| S-1 | [FE] Add a "Customize sidebar" panel to show/hide and reorder rail items | Medium |
| S-2 | [Design] Design the "Customize sidebar" panel and More-menu entry | Medium |

---

## Mockups

### More menu — Customize entry (below a divider)
```
┌─ More ────────────────────┐
│  📊  Analytics             │   ← currently-overflowed items
│  🧭  Discover              │
│  🖼   Library               │
│  ───────────────────────   │   ← divider
│  ⚙   Customize sidebar     │
└────────────────────────────┘
```

### Customize sidebar panel
```
┌─ Customize sidebar ──────────────────────────────────── ✕ ┐
│  Items you show appear in the sidebar in this order. If    │
│  there's not enough room, the lower ones move into More.   │
│                                                            │
│  ⠿   🏠  Home  (default)ⓘ   ●——  on (disabled)             │
│  ⠿   🗓   Publisher          ●——  Shown                      │
│  ⠿   📊  Analytics          ●——  Shown                      │
│  ⠿   📥  Inbox              ●——  Shown                      │
│  ⠿   🧭  Discover           ●——  Shown · in More on this screen│
│  ⠿   🖼   Library            ●——  Shown · in More on this screen│
│  ⠿   🔗  Social Accounts    ——○  Hidden                     │
│                                                            │
│                    [ Reset to default ]        [ Done ]    │
└──────────────────────────────────────────────────────────────┘
```
`⠿` = drag handle · toggle = `Switch` (Shown/Hidden). The row carrying a **(default)** label is the user's **default landing page** (here, Home, but not hardcoded — if their default is Analytics, that row gets it). Its toggle is **on but disabled** (it can't be hidden). Hovering the **(default)** label shows a tooltip. The row is **still draggable** — the user can reposition it anywhere in the order like the others.

### Responsive overflow — same settings, two screens
```
 SHORT SCREEN                          TALL SCREEN
 ┌────────┐   ┌─ More ───────────┐     ┌────────┐   ┌─ More ───────────┐
 │ Home   │   │ 🧭 Discover       │     │ Home   │   │ ⚙ Customize      │
 │ Publ.  │   │ 🖼  Library        │     │ Publ.  │   └──────────────────┘
 │ Anly.  │   │ ────────────────  │     │ Anly.  │
 │ Inbox  │   │ ⚙ Customize       │     │ Discvr │
 │  •••   │   └──────────────────┘     │ Library│
 │  More  │                            │  •••   │
 └────────┘                            │  More  │
 Discover + Library overflow to More    Everything fits; More holds only Customize
```
The rail fills top-down by the user's order until it runs out of height (reserving the More slot); whatever doesn't fit spills into More along with hidden items. No hard cap — adapts per screen size.

---

## S-1 · [FE] Add a "Customize sidebar" panel to show/hide and reorder rail items
**Project:** Web App · **Group:** Frontend · **Skill:** Frontend · **Product area:** Throughout product · **Priority:** Medium · **Type:** Feature

### Description
As a user, I want to choose which sections appear in my desktop left sidebar and in what order — without per-item hover controls that don't fit the narrow rail — so that the sections I use most are always one click away and the rest are tucked into More.

### Workflow
1. The user opens the **More** menu in the left sidebar.
2. Below the overflow items, separated by a divider, they see **⚙ Customize sidebar**.
3. Clicking it opens the **Customize sidebar** panel listing all sidebar sections.
4. Each section has a **Show/Hide** toggle and a **drag handle** to reorder. The user's **default landing page** carries a small **(default)** label, its toggle is **on but disabled** (can't be hidden), and it can still be dragged/repositioned like any other.
5. As they toggle and reorder, the sidebar reflects the changes; lower-priority shown items that don't fit the current screen move into More automatically.
6. They click **Done**; their layout is remembered next time, on any device. A **Reset to default** option restores the standard layout.

### Acceptance criteria
- [ ] The **More** menu shows a **⚙ Customize sidebar** entry at the bottom, below a divider.
- [ ] Clicking it opens a **Customize sidebar** panel listing the sidebar's navigation sections, each with a **Show/Hide `Switch`** and a **drag handle**.
- [ ] Panel title: **"Customize sidebar"**; subtext: **"Items you show appear in the sidebar in this order. If there's not enough room, the lower ones move into More automatically."**
- [ ] The user's **default landing page** (from Account Settings → Default Landing Page) **cannot be hidden** — its row shows a **(default)** label and its Show/Hide toggle is **on but disabled**.
- [ ] Hovering the **(default)** label shows a tooltip: **"This is your default landing page. You can change it in settings."**
- [ ] The default-landing-page row is **still draggable** — the user can reposition it in the order like any other section (it is not anchored to the top).
- [ ] If the user changes their default landing page in Account Settings, the locked **(default)** row follows the new choice (no hardcoded Home).
- [ ] The **More** control and system utilities (notifications, profile) are not listed in the panel — only navigable sections are customizable.
- [ ] Toggling a section **Hidden** keeps it available under **More**; toggling it **Shown** makes it eligible for the rail.
- [ ] Dragging a section reorders it; the order sets the rail priority (top of the list wins the visible slots).
- [ ] **Show-all is allowed:** the rail renders shown sections top-down by order until it runs out of vertical space (reserving the More slot); sections that don't fit spill into **More** automatically — there is no hard cap and the result adapts to screen size/resize.
- [ ] A shown section that is currently overflowing into More is indicated in the panel (e.g. **"Shown · in More on this screen"**).
- [ ] The sidebar reflects changes made in the panel (live, or on Done — see Design).
- [ ] **Discover** defaults to **Hidden** (in More), preserving today's default, but can now be shown via the panel.
- [ ] A **Reset to default** action restores the default shown/hidden set and order; tooltip/label: **"Restore the default sidebar layout."**
- [ ] Choices persist **per user across sessions and devices** and apply on next load.
- [ ] The existing height-based auto-overflow continues to work and never hides the More button when overflow exists.
- [ ] When the user **shows** a section, a `nav_rail_item_shown` Usermaven event fires with `{ item: '<id>' }`; **hiding** fires `nav_rail_item_hidden` with `{ item: '<id>' }`.
- [ ] When the user **reorders**, a `nav_rail_reordered` event fires; **reset** fires `nav_rail_layout_reset`.
- [ ] All new copy (panel title/subtext, Customize entry, overflow hint, Reset) is added across every locale directory under `src/locales/`, English first.

### Mock-ups
See the **Mockups** section above and the **[Design]** story.

### Impact on existing data
Adds one new key to the user's stored preferences (the sidebar layout: shown/hidden + order). No schema migration — uses the existing generic preferences store.

### Impact on other products
Desktop web rail only. No mobile or Chrome extension change. Theme/white-label safe (icons/labels already theme-aware).

### Dependencies
- **[Design] Design the "Customize sidebar" panel and More-menu entry**.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*
- `contentstudio-frontend/src/components/layout/DesktopNavigationRail.vue` — owns the rail render + the current height-based auto-overflow (`OVERFLOW_PRIORITY`, `overflowIds`, `visibleItems`, `overflowItems`, the More `Dropdown`). Reframe overflow to honor the user's saved order/shown-set first, then spill by height. Add the **Customize sidebar** entry in the More dropdown (divider + gear) and the Customize panel/modal.
- `contentstudio-frontend/src/components/layout/useHeaderNavigation.ts` — builds `primaryNavItems` (ids: `home`, `publisher`, `analytics`, `inbox`, `discover`, `media-library`, `social-accounts`, `brand-knowledge`, `more`). Good home for the saved layout state + show/hide/reorder/reset actions.
- Persistence: reuse `setPreferenceStatus(key, value)` from `contentstudio-frontend/src/modules/common/composables/useHelper.js` → `setUserPreferencesApi` → `preferences/setPreferences`. Suggested key e.g. `nav_sidebar_layout` holding `{ order: [...ids], hidden: [...ids] }`. Confirm the endpoint round-trips an object value; if it only persists scalars, a tiny BE follow-up to allow the value type is the only addition.
- Components: `@contentstudio/ui` `Modal` for the panel, `Switch` for Show/Hide, `Icon` (gear) for the entry; a drag-reorder approach consistent with existing drag usage in the app (e.g. the content-category slots/list reordering). Theme tokens only.
- `const { trackUserMaven } = useUserMaven()` for the new events.
- Default landing page: `contentstudio-frontend/src/modules/setting/composables/useProfilePage.ts` (`landingPage` / `selectedLandingPage` / `resolveLandingValue`) and `useProfileStore` hold the user's default landing page; the `landing_options` map (settings.json: `home`, `dashboard`, `publisher`, `analytics_overview_v3`, `inbox`, `discovery`, `media-library`, `listening`, `api`) maps to rail section ids — note the aliases (`analytics_overview_v3`→`analytics`, `discovery`→`discover`). Lock whichever rail section matches the user's default landing value; keep it always-shown but reorderable. The tooltip's "settings" link points to the Profile / Default Landing Page setting.
- Keep only **More** (and notifications/profile utilities) out of the reorderable/hideable set; every navigable section — including the default landing page — is draggable.

---

## S-2 · [Design] Design the "Customize sidebar" panel and More-menu entry
**Project:** Web App · **Group:** Design · **Skill:** Design · **Product area:** Throughout product · **Priority:** Medium · **Type:** Feature

### Description
As the team building sidebar customization, we need a clear design for the Customize entry and panel so that show/hide + reorder is obvious, the locked Home state is understandable, and the "this is shown but currently in More" nuance reads clearly.

### Workflow
1. Designer reviews the customer.io reference and the mockups above.
2. Designer delivers Figma for the More-menu Customize entry and the Customize panel (rows, toggle, drag handle, locked Home, overflow hint, Reset/Done).
3. Designer specs whether the sidebar updates live as the user edits, or on Done, and the panel type (modal vs right-side drawer).

### Acceptance criteria
- [ ] Design for the **⚙ Customize sidebar** entry in the More menu (divider + gear + label).
- [ ] Design for the **Customize sidebar** panel: row layout with section icon + name, a **Show/Hide `Switch`**, and a **drag handle**; plus title, subtext, **Reset to default**, and **Done**.
- [ ] Design the **default-landing-page** row: a **(default)** label with an info affordance, a **disabled "on" toggle**, and the hover tooltip ("This is your default landing page. You can change it in settings.") — while keeping the row **draggable**.
- [ ] Design the **"Shown · in More on this screen"** indicator for shown rows that are currently overflowing.
- [ ] Design the **drag state** (handle, drop indicator) for reordering.
- [ ] Recommend **panel type** — centered `Modal` vs right-side drawer — and whether the sidebar **updates live** while editing or applies on **Done**.
- [ ] All designs use existing `@contentstudio/ui` components and theme tokens (white-label safe); no dark mode, no RTL.

### Mock-ups
This story produces the final mock-ups; the ASCII mockups above are the starting point. Link the Figma file here on completion.

### Impact on existing data
N/A — design only.

### Impact on other products
N/A — design only. Desktop web rail only.

### Dependencies
None (informs the FE story).

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, design story (desktop rail only)
- [ ] Multilingual support — N/A, design story (copy provided by the FE story)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment — N/A, desktop web only
