# Research: LinkedIn Profile Analytics revival

**Date:** 2026-09-07
**Trigger:** LinkedIn Profile Analytics was built and QA-deployed earlier in the year (tracker stories CONT-220, CONT-408, CONT-407, CONT-1687, CONT-1708, CONT-1717, CONT-1727, CONT-1728, CONT-2003, CONT-2037, CONT-2038). It never reached production and is no longer reachable from the frontend. This research grounds a follow-up epic to bring it back behind a per-user feature flag.

---

## 1. Current state, by layer

### Ingestion (contentstudio-social-analytics-go): built and running

There is a complete, production-shaped Profile lane parallel to the Page lane.

| Stage | Where | Profile support |
|---|---|---|
| Scheduler | `src/cmd/jobs/fetcher/linkedin.go:41` | Maps account type `Profile` to the Kafka topic `work-order-linkedin-profile-batch`. Same 6-hour `last_analytics_updated_at` freshness gate as pages. |
| Scheduler invocation | `src/services/unified/account-fetcher/main.go:128` | Already passes `[]string{"page", "profile"}`, so **every** LinkedIn profile account in Mongo is being enqueued today. |
| Scheduler default | `src/cmd/jobs/fetcher/linkedin.go:56` | When no account types are passed, only `Page` is processed. So the standalone job is page-only, the unified fetcher is not. |
| Fetcher | `src/services/linkedin/linkedin-fetcher/profile.go:97` | Real implementation. Calls `memberCreatorPostAnalytics` for 5 query types (`IMPRESSION`, `MEMBERS_REACHED`, `RESHARE`, `REACTION`, `COMMENT`) plus `memberFollowersCount`, in parallel via `errgroup`, with token-error detection, `RecordProcessingError`, and `last_analytics_updated_at` updates. |
| Fetcher scope note | `src/services/linkedin/linkedin-fetcher/profile.go:93` | Comment is explicit: "Profile accounts only fetch insights data (no posts or org details)." |
| Parser | `src/utils/parsing/linkedin_parser.go:400` | Branches on `entityType == "profile"` into `parseProfileInsightsDaily`. |
| Analytics sink | `src/services/linkedin/linkedin-analytics-sink/main.go:43-44` | Consumes `raw-linkedin-profile-posts` and `raw-linkedin-profile-insights` on a dedicated consumer group. |
| ClickHouse sink | `src/services/linkedin/linkedin-clickhouse-sink/main.go:44-45`, `run.go:130-153` | Consumes `parsed-linkedin-profile-posts` and `parsed-linkedin-profile-insights`. Four consumers total. |
| Storage | `src/db/clickhouse/linkedin.go:119` | Profile rows are written to the **same** `linkedin_insights` table as pages. |

### Read and API layer: page-only

- Go query service exposes page methods only: `GetSummary`, `GetAudienceGrowth`, `GetPageViews`, `GetPublishingBehaviour`, `GetTopPosts`, `GetPostsPerDay`, `GetHashtags`, `GetFollowersDemographics` (`src/services/analytics/linkedin/service.go:30-38`). Nothing surfaces member impressions or members-reached as first-class metrics.
- Laravel legacy web controller `contentstudio-backend/app/Http/Controllers/Analytics/Analyze/LinkedinController.php` queries ClickHouse directly with page-shaped SQL (`app/Builders/Analytics/Analyze/LinkedinBuilder.php`).
- Laravel v1 API controller `contentstudio-backend/app/Http/Controllers/Api/V1/Analytics/LinkedInAnalyticsController.php` proxies to the Go service. Its OpenAPI tag description says "LinkedIn **page** analytics".

### Schema gap

`linkedin_insights` has no account-type discriminator. Columns are page-shaped: `organization_name`, `page_views`, `unique_visitors`, `page_views_by_country`, `page_views_by_seniority`, and so on (`src/db/clickhouse/linkedin.go:119-155`). Profile rows populate only the shared subset: `impressionCount`, `reach`, `repost`, `comments`, `reactions`, `engagement`, `totalFollowerCount`, `daily_follower_count`. Existing tracker stories CONT-1708, CONT-1727 and CONT-2003 own this work.

### OAuth scopes: analytics scope is not in production

`contentstudio-backend/app/Strategy/Integrations/LinkedinConnector.php:44-70`

- Production scope set: `r_basicprofile`, `r_emailaddress`, `rw_organization_admin`, `w_organization_social`, `w_member_social`, `r_ads`, `rw_ads`, `r_liteprofile`, `r_ads_reporting`, `r_organization_social`, `r_1st_connections_size`.
- A second scope set exists behind `if (env('APP_ENV') === 'qa-features')` and is the only place `r_member_profileAnalytics` and `r_member_postAnalytics` appear.

So live production tokens cannot call the member analytics APIs. Every existing LinkedIn profile account needs a fresh authorization pass before the ingestion lane can return anything but errors.

Single app credential set today: `LINKEDIN_APP_ID`, `LINKEDIN_APP_SECRET`, `LINKEDIN_CALLBACK` (`LinkedinConnector.php:31`, `Integrations/Platforms/Social/LinkedinController.php:54-55`).

### No per-account record of granted permissions

`contentstudio-backend/app/Models/Integrations/Platforms/Social/LinkedinAccounts.php:28-60` has `validity`, `validity_error`, `validity_status`, `token_expires_at`, `token_issued_at`, `state`, but nothing describing which scopes the stored token actually carries. There is no way today to tell a posting-only profile token apart from an analytics-capable one.

### Frontend: profiles are filtered out of analytics

`contentstudio-frontend/src/modules/analytics/components/common/composables/useAnalyticsUtils.ts:234-272`

The LinkedIn branch of `getPlatformAccounts` pushes an account only when `typeof item.linkedin_id !== 'string'`. Pages are stored as integers and profiles as strings in the same array (confirmed by comments at `contentstudio-backend/app/Repository/Integrations/Platforms/SocialRepo.php:28` and `app/Repository/Settings/WorkspaceTeamRepo.php:484`). That single type check is what hides personal profiles from the analytics account selector. It is the one line that has to change for profiles to reappear.

Note: the nearby `account.type === 'Page'` filters in `useAnalyticsAccountSelector.ts:90-93`, `useAnalyticsUtils.ts:151/187/195`, `TryTemplateModal.vue:176` and `looker-studio/MainComponent.vue:81` are **Facebook-specific**, not LinkedIn. Do not touch them.

---

## 2. Existing infrastructure we can reuse

### Per-user feature flags

`contentstudio-frontend/src/stores/core/useProfileStore.ts:150-170`

```
const hasFeatureFlag = (name: string): boolean => { ... }
```

- Reads `profile.feature_flags: string[]`.
- Internal staff pass automatically by email domain: `@contentstudio.io`, `@d4interactive.io` (`INTERNAL_EMAIL_DOMAINS`, line 53).
- Exposed as `useAccount().featureFlag(name)` (`src/composables/useAccount.ts:152`) and via `useComposerHelper().featureFlag`.
- Safe to call in router guards.
- Live flag names in use: `facebook_reel_collaborators`, `google_ads`. Convention is `snake_case`, so `linkedin_profile_analytics` fits.

Backend side: `feature_flags` is a plain array on the user document, declared in `app/Models/Account/User.php:106,125`, typed in `app/Data/Auth/UserData.php:80`, and returned to the client via `app/Http/Controllers/Accounts/AccountController.php:83`.

**Gap:** there is no admin UI, endpoint or artisan command anywhere in the backend that writes `feature_flags`. Granting a flag today means editing the Mongo `users` document by hand. If support is expected to turn this on for users who ask, a grant mechanism is part of the work.

**Decision needed:** the flag is per **user**, but analytics accounts are per **workspace**. A flagged user sees profile accounts, their unflagged teammates in the same workspace do not. That is acceptable for a request-access beta but should be a conscious choice.

### Reconnect / token-expired banner

Already wired for LinkedIn analytics, so the "token expired" half of this epic is largely existing behavior:

- `contentstudio-frontend/src/modules/analytics/views/linkedin/MainComponent.vue:96-99` computes `isReconnectRequired` from `['invalid', 'expired'].includes(selectedAccount.validity)`.
- Line 190-205 renders `CstAlert` with `type="token-expired"` and a "Reconnect now" button that opens the `social-connect-modal`.
- Copy keys already exist: `analytics.linkedin.main.actions.reconnect_now`, `analytics.linkedin.main.banner.*`.
- `src/modules/analytics/components/common/utils/analyticsSyncMessage.ts` maps a `Failed` sync with `error_type: 'token_invalid'` to `validity: 'invalid'`, which is what raises the banner. This is the hook a "scopes missing" state could reuse or sit beside.

**Important distinction for this epic:** `validity: 'invalid'` means the token is dead and posting is broken too. A profile that was connected with the old scope set has a perfectly **valid** token, posting works, and only analytics is unauthorized. That is a different state and needs its own field, not a reuse of `validity`.

### Empty state component

`contentstudio-frontend/src/modules/analytics/components/common/AnalyticsOnboardingEmptyState.vue`

- Props: `platform`, `showConnect`, `showLearnMore`.
- Per-platform config carries a logo, a feature icon, and a `helpArticleId` for the Learn more link.
- Copy resolves from `analytics.onboarding.<platform>.*` i18n keys.
- LinkedIn already renders it when the workspace has zero LinkedIn accounts (`views/linkedin/MainComponent.vue:173-178`).

A "connected but analytics not authorized" state is a new variant of this card, not a new component from scratch.

### Account type label

`useAnalyticsUtils.ts:103-115` already has `profileTypeMap.profile` backed by the `analytics.common.profile_types.profile` i18n key, so the selector can label a profile account correctly with no new copy plumbing.

### Locales

`contentstudio-frontend/src/locales/{en,es,fr,pl,el}/analytics.json`. New copy needs an `en` entry at minimum, with the other locales falling back.

---

## 3. Risks and gotchas

1. **The unified account fetcher is already enqueueing every LinkedIn profile.** `src/services/unified/account-fetcher/main.go:128` passes `"profile"`. With production tokens lacking `r_member_profileAnalytics`, those work orders can only fail. Before or alongside the flag rollout, the scheduler should skip profiles that are not analytics-authorized, otherwise we burn LinkedIn API quota and fill Sentry with 403s across the whole profile population. Worth checking current error volume in the fetcher logs.

2. **No profile posts.** The profile lane fetches insights only. Top posts, publishing behaviour, posts-per-day and hashtags have no data source for profiles. Those widgets must be hidden, which is what the existing tracker story CONT-1727 covers.

3. **No account-type column in ClickHouse.** Profile and page rows are indistinguishable except by `linkedin_id`. Any read query that needs to branch has to resolve the type from Mongo.

4. **Backfill depth is unverified.** The fetcher computes ranges from `calculateDateRanges(order.SyncType)` and the testable helper uses a 12-month window, but whether LinkedIn actually returns 12 months of member creator analytics on first authorization is untested. Copy should not promise a history window until this is confirmed. Currently phrased as "it can take up to 24 hours for your first numbers to appear", which is safe either way.

5. **`r_1st_connections_size` and the two scope sets have drifted.** The qa-features set drops `r_ads`, `rw_ads`, `r_ads_reporting`, `r_emailaddress` and `r_liteprofile`. If production moves to the qa-features scope list wholesale, Meta-style ads reporting and email-based flows could break. The production set should gain the analytics scopes rather than be replaced by the qa list.

6. **Account limits must not move.** Reconnecting a profile for analytics goes through the same OAuth callback that runs `processPagesAndGroups` / `processBulkReconnection` (`Integrations/Platforms/Social/LinkedinController.php:71-82`). Reconnect must resolve to the existing account document by `linkedin_id`, not create a second one, or workspace account counts will drift.

---

## 4. Suggested naming (for devs, not for the story body)

- Feature flag: `linkedin_profile_analytics`
- New field on the LinkedIn account document: `analytics_authorized` (bool) plus `analytics_authorized_at` (date). Add both to `LinkedinAccounts::$fillable` and to whatever social-account payload the analytics store consumes.
- Grant mechanism: `php artisan feature-flag:grant {email} {flag}` and a matching `:revoke`, since nothing exists today.
- Go scheduler gate: filter the profile branch of `processLinkedinBatches` on `analytics_authorized == true`.

---

## 5. Open questions for the PO

1. Per-user flag or per-workspace? Existing mechanism is per-user only. Per-workspace would need new plumbing.
2. Second LinkedIn app, or add scopes to the existing app? The user-facing story is the same either way ("reconnect once"), but a second app means new `LINKEDIN_ANALYTICS_APP_*` env keys and a second callback route.
3. Does the reconnect for analytics also need to cover LinkedIn **pages**, or profiles only? Pages already work with `rw_organization_admin`, so profiles only, unless we want member post analytics on page posts too.
4. Confirm the historical backfill window LinkedIn returns on first authorization so the copy can be firmed up.
