# Stories: Social Listening Data Availability Banner

## Story 1 — [FE] Add data availability info banner to the Social Listening module

---

### Description:

As a Social Listening user, I want to see a clear notice at the top of the Listening module explaining how far back my data goes, so that I understand the 30-day historical window and I'm not confused by missing mentions older than that.

When a user connects a brand to Social Listening, ContentStudio fetches up to 30 days of historical mention data retroactively from that connection date. Mentions older than 30 days before the connection date are not available. Without a visible explanation, users may assume the feed is broken or incomplete. This banner sets the right expectation on every visit.

---

### Workflow:

1. User navigates to the Social Listening module.
2. At the top of the main view — above the mention feed, analytics, or any tab content — an informational banner is displayed.
3. The banner reads:

   > **"When you connect a brand to Social Listening, we fetch up to 30 days of mention history from your connection date. Data older than 30 days before you connected is not available."**

4. The banner is styled using the `CstAlert` component with `type="info"` — consistent with the informational banners in the Analytics module (e.g., LinkedIn Analytics, Facebook Analytics).
5. The banner is persistent — it remains visible on every visit. It is not dismissible (no close button), as it communicates a permanent data constraint rather than a temporary state.
6. The banner appears on all sub-pages and tabs within the Listening module (feed, analytics, bookmarks, alerts, settings) — placed at the shell/layout level so it is always visible regardless of which tab is active.

---

### Acceptance criteria:

- [ ] An informational banner is visible at the top of the Social Listening module on every visit, across all tabs
- [ ] The banner copy reads exactly: **"When you connect a brand to Social Listening, we fetch up to 30 days of mention history from your connection date. Data older than 30 days before you connected is not available."**
- [ ] The banner uses the `CstAlert` component with `type="info"` styling — it must match the visual language of the Analytics module info banners
- [ ] The banner is not dismissible — there is no close/× button
- [ ] The banner renders correctly at all viewport sizes (desktop, tablet, mobile-width)
- [ ] The banner copy is fully internationalised — all locale directories under `contentstudio-frontend/src/locales/` have a matching translation key
- [ ] The banner does not appear in any shared/read-only export view (public share links), as external readers have no connection-date context

---

### Mock-ups:

N/A — follow the existing `CstAlert type="info"` visual style used in the Analytics module platform views (LinkedIn and Facebook analytics `MainComponent.vue` files for reference).

---

### Impact on existing data:

None. Purely visual — no new data fields, no schema changes, no API calls.

---

### Impact on other products:

- **Mobile apps:** Social Listening is web-only in V1. No mobile impact.
- **Chrome extension:** No impact.
- **White-label:** `CstAlert` from `@contentstudio/ui` respects the theme's info color variables — white-label themes are automatically supported.

---

### Dependencies:

**[FE] Listening module shell — navigation, routing, and user state management** (sc-113295) — the banner must be added within the module shell that this story builds. This story should be implemented after or alongside the module shell.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension) — N/A for mobile and Chrome; web only

---

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

**Existing pattern to follow:**
- `contentstudio-frontend/src/modules/analytics/views/linkedin_v2/MainComponent.vue` — uses `<CstAlert type="info" class="text-left mx-5">` inside a `<template v-slot:alert>` slot within `TabsComponent`. Mirror this pattern.
- `contentstudio-frontend/src/modules/analytics/views/facebook_v2/MainComponent.vue` — same pattern.

**Suggested i18n key:**
- Namespace: `social_listening` (new namespace) or `common`
- Key suggestion: `social_listening.module.data_availability_banner`
- English value: `"When you connect a brand to Social Listening, we fetch up to 30 days of mention history from your connection date. Data older than 30 days before you connected is not available."`

**Placement note:**
The banner should sit at the shell/layout level (outside tab content slots) so it renders on every sub-page — consistent with how Analytics module banners sit above the `TabsComponent` tabs.
