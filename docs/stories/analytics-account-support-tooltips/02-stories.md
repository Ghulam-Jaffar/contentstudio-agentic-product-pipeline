# Story — Supported-account-type tooltips on the analytics account selector

One frontend story. Create with the **New Feature Template**. Nothing is pushed to Shortcut.

This is a single shared-component change: the analytics account selector
(`PlatformAccountSelect.vue`) is one component reused by every platform, so the tooltip
is added once and rendered per-platform from its `type` prop.

---

## [FE] Add "supported account type" tooltip beside the analytics account selector

### Description

As a user viewing a platform's analytics, I want a short explanation next to the
account picker telling me which account types that platform's analytics supports, so I
understand at a glance why (for example) my personal Facebook profile or personal
LinkedIn profile doesn't appear in the list — instead of assuming something is broken.

Today the account picker silently shows only the eligible accounts (e.g. Facebook shows
Pages only, LinkedIn shows Company Pages only), with no on-screen explanation of the
rule. This story adds a small info icon (ℹ) immediately to the right of the account
picker on each platform's analytics page that reveals a plain-language message on hover
or focus.

**In scope:** Facebook, Instagram, LinkedIn, Pinterest, TikTok, YouTube, Google Business
Profile.
**Out of scope:** X (Twitter) — it already has its own analytics info tooltip in the tab
header and has no account-type rule to explain — and the cross-platform Overview page,
which mixes platforms so no single rule applies.

### Workflow

1. The user opens the analytics page for one of the supported platforms (e.g. Publish
   → Analytics → Facebook).
2. Next to the account-selector dropdown in the header, the user sees a small info icon
   (ℹ).
3. The user hovers over (or, on touch/compact screens, taps) the icon.
4. A tooltip appears showing the message for that platform — e.g. on Facebook:
   "Analytics are only available for Facebook Pages. Personal profiles and Groups aren't
   supported."
5. The user now understands which account types are supported and moves on.

The icon sits in the header's left cluster, immediately to the right of the account
dropdown (for Pinterest, to the right of the second/board dropdown). It does **not**
appear on the X (Twitter) page or the Overview page.

### Acceptance criteria

- [ ] An info icon (ℹ) appears immediately to the right of the account-selector dropdown on each of these analytics pages: Facebook, Instagram, LinkedIn, Pinterest, TikTok, YouTube, Google Business Profile.
- [ ] On Pinterest (two dropdowns: profile → board), the icon appears to the right of the **board** dropdown, at the end of the group.
- [ ] No info icon appears on the **X (Twitter)** analytics page.
- [ ] No info icon appears on the cross-platform **Overview** analytics page.
- [ ] Hovering the icon (mouse) shows the tooltip; the tooltip is also reachable via keyboard focus.
- [ ] On compact / touch viewports, the tooltip opens on tap/click (matching the existing analytics info-tooltip behavior).
- [ ] The tooltip shows the exact copy per platform:
  - [ ] **Facebook:** "Analytics are only available for Facebook Pages. Personal profiles and Groups aren't supported."
  - [ ] **Instagram:** "Analytics are only available for Instagram Business and Creator accounts. Personal accounts aren't supported."
  - [ ] **LinkedIn:** "Analytics are only available for LinkedIn Company Pages. Personal profiles aren't supported." *(see note in Impact — confirm against live behavior before shipping)*
  - [ ] **Pinterest:** "Analytics are available for your Pinterest profile and each of its boards."
  - [ ] **TikTok:** "Analytics are available for all connected TikTok accounts."
  - [ ] **YouTube:** "Analytics are available for YouTube channels. You may need to reconnect periodically to keep analytics up to date."
  - [ ] **Google Business Profile:** "Analytics are available for your Google Business Profile locations."
- [ ] All copy is added to every locale (`en fr de it es el zh pl`); non-English locales show the translation or fall back to English (never a raw key string).
- [ ] The icon and tooltip use theme-aware styling (no hardcoded colors) and render correctly on white-label domains.
- [ ] Adding the icon does not shift or overlap the account dropdown, date-range picker, Share, or Export controls at any breakpoint.

### Mock-ups

No new screens. Placement within the existing analytics header:

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  [ 🅕  Acme Facebook Page  ▾ ] ⓘ      Last updated …   [ Date range ▾ ]  Share  Export │
│   └─ account selector ──────┘  ▲                                                     │
│                                └── supported-account-type tooltip (hover / focus)    │
└──────────────────────────────────────────────────────────────────────────────────┘

Pinterest (two dropdowns):  [ Profile ▾ ] [ Board ▾ ] ⓘ
```

### Impact on existing data

None. No schema, API, or data changes — the account-eligibility rules already exist in
the product; this only surfaces them as on-screen copy. The backend already provides all
analytics data.

**Note on LinkedIn copy:** the account selector currently surfaces **Company Pages
only**, which is what the default copy reflects. The analytics data pipeline can also
fetch personal LinkedIn Member Profiles (with limited data). Before shipping, confirm in
the live app whether a connected personal LinkedIn profile appears in the analytics
account dropdown. If it does, change the LinkedIn line to: "Analytics are available for
LinkedIn Company Pages and personal profiles. Personal-profile analytics are limited —
follower growth and audience demographics aren't available."

### Impact on other products

- **Mobile apps (iOS/Android):** N/A — analytics dashboards are a web surface.
- **Chrome extension:** N/A.
- **White-label:** Must use theme-aware color tokens so the icon/link colors follow the
  white-label primary color.

### Dependencies

None.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry point:**
- `contentstudio-frontend/src/modules/analytics/views/common/PlatformAccountSelect.vue` —
  the single account selector shared by all platforms (receives a `type` prop). Adding
  the icon here (immediately after the dropdown, wrapped in a flex container) covers every
  in-scope platform in one change. Guard it so it does **not** render when
  `type === 'twitter'`. For the Pinterest branch (profile + board), place it after the
  board dropdown.

**Established pattern to follow:**
- `contentstudio-frontend/src/modules/analytics/components/common/PlatformTooltip.vue` —
  the existing analytics info tooltip (info `Icon name="Info"` + `v-menu` popover, with
  hover-vs-click behavior for compact viewports). Either reuse this popover pattern or use
  the lighter `v-tooltip` directive (floating-vue) already used inside
  `PlatformAccountSelect.vue` (e.g. the reconnect warning and account-name tooltips). The
  `@contentstudio/ui` catalog has **no standalone Tooltip component**, so one of these two
  patterns is the right call — don't invent a new component.
- The existing `PlatformTooltip` covers **data caveats** (timezone, refresh cadence), not
  account types — this new copy is a separate, additive concern. Do not merge it into the
  `analytics.platform_tooltip.*` keys.

**Suggested i18n keys** (add to `src/locales/<locale>/analytics.json`, all 8 locales),
under `analytics.common` alongside the existing `profile_types` block:

```json
"account_support": {
  "facebook": "Analytics are only available for Facebook Pages. Personal profiles and Groups aren't supported.",
  "instagram": "Analytics are only available for Instagram Business and Creator accounts. Personal accounts aren't supported.",
  "linkedin": "Analytics are only available for LinkedIn Company Pages. Personal profiles aren't supported.",
  "pinterest": "Analytics are available for your Pinterest profile and each of its boards.",
  "tiktok": "Analytics are available for all connected TikTok accounts.",
  "youtube": "Analytics are available for YouTube channels. You may need to reconnect periodically to keep analytics up to date.",
  "gmb": "Analytics are available for your Google Business Profile locations."
}
```
Referenced as `t('analytics.common.account_support.facebook')`, keyed off the `type` prop.
No `twitter` / `overview` keys (both out of scope).

**Existing behavior to preserve (no change needed):**
- The eligibility filtering in
  `contentstudio-frontend/src/modules/analytics/components/common/composables/useAnalyticsUtils.ts`
  (`getPlatformAccounts`) already decides which accounts appear — this story only explains
  it, it must not change which accounts are shown.

**No analytics event:** this is a view-only informational tooltip; no new Usermaven event
is needed.

---

### Shortcut fields

- **Template:** New Feature Template
- **Story type:** feature
- **Project:** Web App
- **Group:** Frontend
- **Epic:** Q2 - 2026: Miscellaneous
- **Priority:** Low
- **Product Area:** Analytics
- **Skill Set:** Frontend
- **Estimate:** _(empty — devs estimate during sprint planning)_
- **Labels:** _(none)_
- **Iteration:** _(PO assigns the current/target sprint at creation)_
