# Research — Home hero: Brand Knowledge widget + heading polish

## Scope

Three frontend-only changes to the **Home** dashboard hero (AI Studio launcher). Web app only — no backend, no mobile.

1. **Replace** the two text links under the AI prompt input — **"Recent chats"** and **"My AI creations"** — with a compact **Brand Knowledge widget** (a "Setup Brand Knowledge" pill).
2. **Remove** the horizontal **Brand Knowledge info banner** near the top of Home.
3. **Heading polish** — add vertical spacing between the hero heading and the tab row below it, and increase the heading font size for more prominence.

> **Supersedes the earlier `home-brand-knowledge-nav-popover` direction.** That prior draft removed the same two links + banner but proposed a *popover anchored to the Brand Knowledge icon in the left nav rail*. The user has since changed the design: the setup nudge now lives **inline, where the two links were** (a pill widget under the prompt), not in the nav rail. Items 1 & 2 overlap; item 3 is new (heading typography vs. nav popover).

---

## Current State

> **Note on the working tree:** the requester made **local, throwaway edits** to `AIAssistantSection.vue` and `DashboardNew.vue` (removing the links/banner, bumping the heading/spacing) purely to preview how it looks. These are **not** the deliverable — the dev implements all three items from scratch. The real baseline is **committed `develop`**, described below, where the links, banner, and 22/24px heading all still exist.

### Baseline on `develop` (what the dev starts from)

- **`src/components/dashboard/AIAssistantSection.vue`** — the hero component contains all the pieces this work touches:
  - The two shortcut links below the suggestion chips:
    - "Recent chats" (`dashboard.ai_assistant.recent_chats`) → opens the chat-history modal.
    - "My AI creations" (`dashboard.ai_assistant.my_ai_creations`) → routes to the media-library AI folder.
    - Backed by script: `useChatsQuery`, `checkForAIFolders`, `hasAIFolder`, `openHistoryModal`, `viewAIGenerations`, `inject('root')`, and related imports.
  - The hero `<h2>` heading — currently `text-[22px] lg:text-[24px]` (item 3 target: ~28px).
  - The mode-tabs wrapper — currently `mt-4` (item 3 target: more spacing, e.g. `mt-6`).
- **`src/views/DashboardNew.vue`** — the wide **`CstAlert`** Brand Knowledge banner (`dashboard.brand_knowledge_banner_message` + `..._action`), gated by a `showBanner` computed and wired to `handleBrandKnowledgeSetup` via `useSetup()` (item 2 target: remove it; clean up or repurpose the orphaned code).

### The hero component

`src/views/DashboardNew.vue` renders `<AIAssistantSection />` inside the hero. `AIAssistantSection.vue` contains: the gradient heading, the `SegmentedControl` mode tabs (AI Chat / Featured / Image / Video / Content), the `ChatInput`, and the suggestion chips. The widget belongs in this component, directly below the suggestion chips (where the removed links were).

---

## What Needs to Change

- **Remove** the two shortcut links ("Recent chats" + "My AI creations") under the AI prompt input, plus the script that backs them.
- **Remove** the wide `CstAlert` Brand Knowledge banner from Home; clean up its now-orphaned `showBanner`/`handleBrandKnowledgeSetup`/`useSetup()` code (or repurpose the `useSetup()` wiring for the widget's visibility gate).
- **Increase** the hero heading font size to ~28px (from 22/24px) and add more vertical spacing between the heading and the tab row.
- **Add** the Brand Knowledge widget under the AI prompt input (below the suggestion chips), in place of the removed links:
  - A pill/badge reading **"Setup Brand Knowledge"** with a trailing **arrow (→)** icon.
  - **On hover:** the trailing → swaps to a **cross (✕)** icon that acts as a dismiss button.
  - **Click the pill body** → navigate to the Brand Knowledge module (`brand-settings` route).
  - **Click the ✕** → dismiss the widget permanently (does *not* navigate).
  - **Visibility:** shown only when brand knowledge isn't set up yet and hasn't been dismissed.

### Open design decisions (see review gate)

- **Visibility gate** — show the widget only to users who haven't set up brand knowledge yet, or always until dismissed? The removed banner used `getProfileAttempted && !AIUserProfile` (profile fetched **and** none exists). The "Setup Brand Knowledge" wording implies the not-set-up gate.
- **Dismiss persistence** — permanent per user (account-level), or session/local only?

---

## Reusable Patterns & Building Blocks

- **Brand-knowledge "not set up" signal:** `useSetup()` (`src/modules/publisher/ai-content-library/composables/useSetup.js`) exposes `AIUserProfile` and `getProfileAttempted`. The removed banner gated on `getProfileAttempted && !AIUserProfile`. Note: if the widget uses this gate, the profile must actually be fetched on the dashboard (the old banner's `useSetup` wiring in `DashboardNew.vue` was the fetch trigger; confirm it still runs or move the gate/fetch into `AIAssistantSection.vue`).
- **One-time dismissal (FE-only, no backend):** `src/components/common/NewNavigationModal.vue` is the established pattern — persist via `setPreferenceStatus(key, value)` from `useHelper` (`composables/useHelpers.ts`), backed by `setUserPreferencesApi({ key, value })` in `api/profile.ts`, plus a `localStorage` fallback flag. Preferences are a flexible key/value store, so a new key (e.g. `home_brand_knowledge_widget_dismissed`) needs **no backend change** (same mechanism as `new_left_navigation_modal`). `useFrillSurvey.ts` is a second live example.
- **Navigation target:** `brand-settings` route (`{ name: 'brand-settings', params: { workspace } }`) — resolves to "Brand Knowledge | Settings" (`src/modules/setting/config/routes/setting.js`). Same target the removed banner's action used and what `useAiToolLauncher.ts` uses.
- **Components (`docs/ui-components.md`):**
  - No dedicated Pill/Chip component exists — build the pill with Tailwind + theme-aware color classes (`bg-primary-cs-50`, `text-primary-cs-500`, etc.), matching the existing suggestion-chip style in `AIAssistantSection.vue`.
  - `Icon` (`@contentstudio/ui`) for the arrow/cross (e.g. `ArrowRight` ↔ `X`), bound with `:size` numeric.
  - `ActionIcon` (`@contentstudio/ui`) is a fit for the icon-only dismiss control.
  - The hover swap is a CSS `group` / `group-hover` pattern — no popover library needed (unlike the superseded nav-popover approach).

---

## Analytics (Usermaven)

- **No new event needed.** Clicking the pill is navigation to `brand-settings` (page views tracked globally); dismissing a nudge is a trivial UI interaction (skip per guidelines §19); actual setup completion already fires `brand_profile_created` from the setup wizard. Nothing new to instrument.

---

## Files Involved

- `contentstudio-frontend/src/components/dashboard/AIAssistantSection.vue` — remove the two links; bump heading font size + tab-row spacing; **add the widget here**.
- `contentstudio-frontend/src/views/DashboardNew.vue` — remove the `CstAlert` banner; clean up/repurpose the orphaned `showBanner`/`handleBrandKnowledgeSetup`/`useSetup()` code.
- New i18n key(s) for the widget label in **all 8 locale dirs** (`src/locales/*/dashboard.json`).
- Reference: `src/components/common/NewNavigationModal.vue` (dismiss/persist pattern), `src/modules/publisher/ai-content-library/composables/useSetup.js` (brand-knowledge state), `src/composables/useHelpers.ts` (`setPreferenceStatus`), `src/modules/setting/config/routes/setting.js` (`brand-settings` route).

---

## Mobile / Cross-product

- **No mobile impact.** This is the web dashboard hero; Brand Knowledge is an AI-generation setup feature (web-only per guidelines §3). No `[iOS]`/`[Android]` stories.
- **White-label:** copy must go through i18n; use theme-aware color classes only (no hardcoded blues) so the pill inherits white-label primary colors.

---

## Story Split

**One `[FE]` story.** All three changes are web-frontend, tightly coupled around the same Home hero, and share no backend work (dismissal reuses the existing preferences API).
