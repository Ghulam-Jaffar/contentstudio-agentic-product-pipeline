# Research — Home cleanup + Brand Knowledge setup popover

## Scope

Three frontend-only changes on the **Home** dashboard / left navigation:

1. Remove the two links under the AI textbox on Home — **"Recent chats"** and **"My AI creations"**.
2. Remove the **Brand Knowledge info banner** from Home.
3. Add a **setup popover** anchored to the **Brand Knowledge icon** in the left navigation rail, with a heading, descriptive text, and two buttons — **Skip** and **Setup Brand**.

All three are web-app frontend changes. No backend, no mobile.

---

## Current State

### 1 & 2 — Already implemented in the working tree (uncommitted)

The frontend repo (`contentstudio-frontend`, branch `develop`) already has **uncommitted edits** that perform changes 1 and 2:

- **`src/components/dashboard/AIAssistantSection.vue`** — the block under the suggestion chips containing the two links was deleted:
  - "Recent chats" link (`$t('dashboard.ai_assistant.recent_chats')` = *"Recent chats"*) → opened the chat-history modal.
  - "My AI creations" link (`$t('dashboard.ai_assistant.my_ai_creations')` = *"My AI creations"*) → routed to the media library AI folder.
  - The now-unused script (`useChatsQuery`, `checkForAIFolders`, `hasAIFolder`, `openHistoryModal`, `viewAIGenerations`, related imports) was also removed.
- **`src/views/DashboardNew.vue`** — the `CstAlert` info banner (`$t('dashboard.brand_knowledge_banner_message')` + `..._action`) was deleted from the template.
  - **Leftover dead code:** `showBanner` computed (line ~54) and `handleBrandKnowledgeSetup` (line ~124) are no longer referenced in the template. These should be cleaned up as part of this work.

> These edits are the user's in-progress work — they will not be reverted. The story documents the full intended change (all three items) so the PO has one complete record to create in Shortcut.

### 3 — Brand Knowledge nav item (target for the new popover)

- **`src/components/layout/useHeaderNavigation.ts`** (~line 377) defines the nav item:
  `id: 'brand-knowledge'`, icon `Lightbulb`, tooltip key `header.profile.brand_knowledge`, routes to `brand-settings`. Only shown when `!isApiCentricPlan` (hidden for API-centric plans).
- **`src/components/layout/DesktopNavigationRail.vue`** renders it icon-only near the bottom of the rail (`SECONDARY_ITEM_IDS = ['social-accounts', 'brand-knowledge']`). It currently shows a right-side label tooltip via the `v-tooltip.right` directive.

### Brand-knowledge "not set up" state (show condition)

The removed banner used this condition (in `DashboardNew.vue`, via `useSetup()`):

```
showBanner = getProfileAttempted && !AIUserProfile
```

i.e. the brand/AI profile has been fetched **and** the user has not created one yet. The new popover should use the same "brand knowledge not set up yet" signal so it only nudges users who still need to complete setup.

---

## What Needs to Change

- **Remove** the "Recent chats" + "My AI creations" links under the Home AI textbox (done in WT — story documents it).
- **Remove** the Brand Knowledge info banner from Home (done in WT — story documents it), and **delete the leftover dead code** (`showBanner`, `handleBrandKnowledgeSetup`, and any now-unused `useSetup` bindings in `DashboardNew.vue`).
- **Add** a popover anchored to the Brand Knowledge nav icon:
  - Heading: **"Setup Brand Knowledge"**
  - Body: **"Add your brand voice, colors, and guidelines here. AI Studio uses it so every generation sounds like you, not generic AI."**
  - Buttons: **Skip** (dismiss) and **Setup Brand** (→ `brand-settings` route).
  - Show only when brand knowledge isn't set up yet **and** the user hasn't already dismissed it.
  - Persist dismissal so it doesn't reappear.

---

## Reusable Patterns & Building Blocks

- **Popover library:** `floating-vue` (v5.2.2) is installed and registered in `main.js` (`FloatingVue` app plugin + globally-registered `VMenu`). This is the same library powering the existing `v-tooltip` directives in the rail — the natural fit for an anchored popover with an arrow and arbitrary content (heading, text, buttons). Note: `@contentstudio/ui` has **no** standalone Popover/Tooltip component (per `docs/ui-components.md` → "What's NOT Available"); legacy `CstPopup` also exists as a fallback.
- **One-time dismissal:** `src/components/common/NewNavigationModal.vue` shows the established pattern — persist via `setPreferenceStatus(key, value)` (from `useHelper`, backed by the generic `setUserPreferencesApi({ key, value })`), plus a `localStorage` fallback flag. A new preference key (e.g. `brand_knowledge_popover_dismissed`) needs **no backend change** — preferences are a flexible key/value store (same mechanism as `new_left_navigation_modal`, calendar prefs, etc.).
- **Setup destination:** the existing `brand-settings` route (`params: { workspace }`) — the same target the removed banner's action used.
- **Buttons:** `Button` from `@contentstudio/ui` (primary for "Setup Brand", `subtle`/secondary variant for "Skip").

---

## Files Involved

- `contentstudio-frontend/src/components/dashboard/AIAssistantSection.vue` — links removed (done in WT).
- `contentstudio-frontend/src/views/DashboardNew.vue` — banner removed (done in WT); remove leftover dead code.
- `contentstudio-frontend/src/components/layout/DesktopNavigationRail.vue` — anchor + render the new popover on the Brand Knowledge item.
- `contentstudio-frontend/src/components/layout/useHeaderNavigation.ts` — Brand Knowledge nav item definition (reference).
- New popover component (e.g. `src/components/layout/BrandKnowledgeSetupPopover.vue`) + i18n keys in all 8 locale dirs.
- Reference: `src/components/common/NewNavigationModal.vue` (dismiss pattern), `src/modules/publisher/ai-content-library/composables/useSetup.js` (brand-knowledge state).

---

## Mobile / Cross-product

- **No mobile impact.** The nav rail is `DesktopNavigationRail.vue` (web). Brand Knowledge is an AI-generation setup feature — web-only per story guidelines.
- **White-label:** brand knowledge nav item is already hidden for API-centric plans; the popover inherits that visibility. Copy must go through i18n and use theme-aware color classes (no hardcoded blues).

---

## Story Split

**One `[FE]` story.** All three changes are web-frontend, tightly coupled around the same Home / Brand Knowledge onboarding UX, and share no backend work.
