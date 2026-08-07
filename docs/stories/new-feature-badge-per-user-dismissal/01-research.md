# Research: Per-user dismissal of "New" feature badges

**Date:** 2026-07-31
**Trigger:** "New" tags shown on newly added features never go away. They should be tracked per user and disappear once that user has actually seen the feature.

---

## 1. Backlog check

Searched `docs/stories/` (140+ slugs) and `docs/features/` (26 slugs). **No existing story or feature covers this.** Nearest neighbours, none overlapping:

- `desktop-nav-rail-pin-unpin` — customises which rail items show, not badges on them
- `filter-button-applied-count-badge` — a count badge in the planner filter, unrelated
- `settings-ui-revamp` — settings sidebar visual work, does not touch the New tags there

This is net-new scope.

---

## 2. How "New" tags work today

There is **no system**. Every New tag is a hand-placed `<img src="@src/assets/img/common/new_tag.svg">` in a component template. There is no registry, no user state, no expiry, and no telemetry.

### Full inventory of New-tag surfaces (web frontend)

| # | Surface | File | How it is gated today |
|---|---|---|---|
| 1 | Left nav rail items | `contentstudio-frontend/src/components/layout/DesktopNavigationRail.vue:439` | `item.isNew` boolean set in `useHeaderNavigation.ts` (AI Studio, Social Listening) |
| 2 | Analytics sidebar platforms | `contentstudio-frontend/src/modules/analytics_v3/components/AnalyticsSidebar.vue` (6 render sites) | `platform.isNew` from `useAnalyticsRoutes.js`. Meta Ads uses a second asset, `new_tag_no_shadow.svg` |
| 3 | AI Studio tool list | `contentstudio-frontend/src/modules/ai-studio/components/AiStudioSidebar.vue:99` | `tool.isNew` from `useAiToolCatalog.ts` |
| 4 | Settings sidebar | `contentstudio-frontend/src/modules/setting/components/SettingSidebar.vue` (4 render sites: SSO Domains, Approval Workflows, each duplicated for locked/unlocked) | **No flag at all.** Hardcoded in the template, shows forever |
| 5 | Dashboard integration cards | `contentstudio-frontend/src/components/dashboard/IntegrationCard.vue:49` | `integration.newTag` |
| 6 | AI image/video model selector | `contentstudio-frontend/src/modules/AI-tools/components/MediaGenerationOptions.vue:538` | Hardcoded `NEW_MODEL_VALUES` array of 7 model ids |
| 7 | Approval sidebar "Workflows" tab | `contentstudio-frontend/src/modules/approval-workflows/components/SendForApprovalSidebar.vue:1230` | **No flag.** Hardcoded |
| 8 | Inbox | `contentstudio-frontend/src/modules/inbox-revamp/views/InboxView.vue`, `components/TagsDropdown.vue` | **No flag.** Hardcoded |
| 9 | Home settings dropdown | `contentstudio-frontend/src/components/layout/HomeSettingsDropdown.vue` | **No flag.** Hardcoded |
| 10 | CSV automation listing | `contentstudio-frontend/src/modules/automation/components/csv/listing/CsvProcessListing.vue` | **No flag.** Hardcoded |
| 11 | Blog selection | `contentstudio-frontend/src/modules/publish/components/posting/blog/BlogSelection.vue` | **No flag.** Blog publishing is sunset, this one should just be deleted |
| 12 | Legacy analytics sidebar | `contentstudio-frontend/src/modules/analytics/components/AnalyticsMain.vue:145` | Hardcoded `<span class="nav-beta new-feature-available">NEW</span>`, a different visual entirely |

**Consequences of the current state:**

- A badge only disappears when a developer opens a PR to delete it. In practice nobody does, so tags like SSO Domains and Approval Workflows have been lit for months.
- A user who signed up last week sees NEW on features that shipped a year before they joined.
- A user who has used AI Studio daily for two months still sees NEW on it.
- Nine of twelve surfaces have no flag at all, so there is not even a toggle to flip.
- Two different visual treatments exist (`new_tag.svg`, `new_tag_no_shadow.svg`, plus the legacy `NEW` span).
- No instrumentation, so there is no evidence the badges drive discovery at all.

---

## 3. Existing per-user preference infrastructure (reusable as-is)

ContentStudio already has everything needed to persist per-user badge state. Nothing new is required on the API.

**Write path:**

```
POST /preferences/setPreferences   { key, value }
  → UserPreferencesController::setPreferences        (app/Http/Controllers/Settings/UserPreferencesController.php:222)
  → UsersRepository::setPreferences($userId, $key, $value)   (app/Repository/Account/UsersRepository.php:896)
  → User::where('_id', $userId)->update(['preferences.'.$key => $value])
```

The Mongo dot-path write means **any key is accepted** with **any JSON value**. `UserPreferencesRequest` validates only `key => required|string` and `value => required`. There is no whitelist.

**Frontend wrapper:** `useHelper().setPreferenceStatus(key, value)` in `contentstudio-frontend/src/modules/common/composables/useHelper.js:64`.

**Read path:** the profile response carries `profile.preferences`, exposed through `useProfileStore` (`src/stores/core/useProfileStore.ts:82-112`).

### Closest existing precedents

**Structured object preference — AI Studio favourites** (`src/modules/ai-studio/composables/useAiToolFavorites.ts`):

```ts
export const AI_STUDIO_FAVORITES_KEY = 'ai_studio_favorite_tools'
const payload = { tools }                      // wrapped, not a bare array
profileStore.setAiStudioFavoriteTools(payload) // optimistic local write
void setPreferenceStatus(AI_STUDIO_FAVORITES_KEY, payload)
```

Its inline comment documents a gotcha we must respect: *"the backend's `required` rule on `value` rejects a bare empty array, so wrapping keeps an emptied list saveable."* Any badge state must therefore be stored as an object, e.g. `{ seen: {...} }`, never a bare array.

**One-time announcement with signup cutoff — left navigation modal** (`src/views/DashboardNew.vue:48-84`, `src/components/common/NewNavigationModal.vue`):

```ts
const isNewUserAfterNavigationLaunch = computed(() => {
  const createdAt = profileData.value?.created_at
  return Date.parse(createdAt) >= NEW_NAVIGATION_MODAL_CUTOFF_TIME
})

shouldShowNewNavigationModal = profileQuery.isSuccess
  && !isNewUserAfterNavigationLaunch      // users who joined after launch never see it
  && !localStorage.getItem(DISMISSED_KEY) // instant local suppression
  && (preferences.new_left_navigation_modal ?? true)
```

Two of the three mechanisms here are exactly what badges need: a **localStorage mirror for instant suppression** and a **server preference as the cross-device source of truth**. The badge system generalises those from one hardcoded announcement to a registry of them.

The **signup cutoff** is the exception. It is correct for this particular modal, which explains a navigation change that only pre-existing users lived through, but it is the wrong rule for feature badges. Copying it would hide current announcements from users who signed up in the days after a release, which is the opposite of what a New badge is for. A release-date window achieves the intended effect without that side effect. See the design decision in `02-stories.md`.

**Other structured preferences already in the store:** `nav_sidebar_layout: { order, hidden }`, `meta_ads_table_preferences: { hidden_columns, column_order }`, `planner_selected_columns`.

---

## 4. Backend DTO drift (finding, not a blocker)

`app/Data/Auth/UserPreferencesData.php` enumerates preference fields for TypeScript generation. It is **already out of date** — `ai_studio_favorite_tools`, `new_left_navigation_modal`, `nav_sidebar_layout`, `calendar_fade_published_posts` and others are live in the product but absent from the DTO.

`ProfileResponseData` (which wraps `UserData` → `UserPreferencesData`) is **not referenced by any controller**, so the profile response is not actually serialised through it and unknown keys are not stripped. The DTO functions purely as a TypeScript type source.

**Conclusion: no backend story is required.** The write and read paths both work today for an unregistered key, exactly as they do for `ai_studio_favorite_tools`. Adding the field to the DTO is a one-line consistency fix worth doing alongside, noted in the story's implementation references.

---

## 5. Scope boundaries

- **Web frontend only.** The user's request was explicitly about the frontend. Native mobile has no equivalent New-tag system, and the apps are mid-migration to Flutter, so there is nothing to keep in sync.
- **Blog surface is deleted, not migrated.** Blog publishing is sunset.
- **The legacy `AnalyticsMain.vue` `NEW` span** is in the old analytics module. Migrate it to the shared component or delete it, depending on whether that view is still routable.
- **No new API, no schema migration.** One additional key inside the existing `preferences` object.

---

## 6. Design system notes

From `docs/ui-components.md`: there is **no Badge/Pill component suited to this**. `Badge` exists for status/count indicators, but the New tag is a bespoke SVG asset with its own shape and positioning. The new `FeatureBadge` wrapper is app-level, not a `@contentstudio/ui` addition, so no design-system release is blocked. The two asset variants (`new_tag.svg`, `new_tag_no_shadow.svg`) plus the legacy text span need consolidating, which is the Design story.

---

## 7. Recommended split

Three stories. No backend story, no mobile story.

| # | Story |
|---|---|
| S-1 | `[FE] Build a per-user "New" badge system with automatic dismissal` |
| S-2 | `[FE] Migrate all existing "New" tags to the shared badge system` |
| S-3 | `[Design] Define the "New" badge visual and its lifecycle states` |
