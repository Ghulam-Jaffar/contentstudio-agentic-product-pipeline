# 01 Research — Platform identifier consolidation

Date: 2026-08-23
Scope: the per-platform account identifier keys still in use across `contentstudio-backend`, `contentstudio-frontend` and `contentstudio-flutter`, and the repeated social-account queries on the backend.

This doc is the local grounding for the epic. Nothing here goes into the story body.

---

## 1. The state today: `platform_identifier` exists everywhere, and so do the old keys

`platform_identifier` is populated on every social account, and newer platforms use it exclusively. The legacy per-platform keys were never removed, so both sets are live at the same time.

The clearest single artefact of this is `contentstudio-backend/config/social_platforms.php` → `account_selection_fields`, which is the product's own list of "which field identifies an account on this platform":

```
facebook.facebook_id          meta_ads.facebook_id
twitter.twitter_id            linkedin.linkedin_id
pinterest.board_id            pinterest.profile_id
tumblr.name                   tumblr.platform_identifier
instagram.instagram_id        gmb.name
youtube.platform_identifier   tiktok.platform_identifier
threads.platform_identifier   bluesky.platform_identifier
telegram.platform_identifier
```

Six platforms use a legacy key, one platform is listed twice with two different keys, one uses `board_id`, two use `name`, and the rest use `platform_identifier`.

The frontend mirrors the same mixture in `contentstudio-frontend/src/modules/common/constants/common-attributes.ts` → `socialChannelsArray`:

| Platform | key in the frontend descriptor |
|---|---|
| facebook | `facebook_id` |
| instagram | `instagram_id` |
| twitter | `twitter_id` |
| linkedin | `linkedin_id` |
| pinterest | `board_id` |
| gmb | `name` |
| threads, telegram, youtube, tiktok, tumblr, bluesky | `platform_identifier` |

## 2. Scale of the change

Raw usage counts, current checkout:

| Codebase | legacy key occurrences | `platform_identifier` occurrences |
|---|---|---|
| `contentstudio-backend/app` | 1547 across `facebook_id` (359), `instagram_id` (350), `linkedin_id` (304), `twitter_id` (183), `tiktok_id` (171), `pinterest_id` (113), `youtube_id` (52), `gmb_id` (13), `threads_id` (2) | 526 |
| `contentstudio-frontend/src` | 629 across `facebook_id` (242), `instagram_id` (149), `twitter_id` (101), `linkedin_id` (67), `pinterest_id` (59), `youtube_id` (4), `gmb_id` (4), `tiktok_id` (3), spread over 143 files | 341 |
| `contentstudio-flutter/lib` | 38 across `facebook_id` (15), `instagram_id` (13), `linkedin_id` (5), `youtube_id` (3), `twitter_id` (1), `threads_id` (1) | 23 |

Note that `tiktok_id` and `youtube_id` still appear even though those platforms are configured to use `platform_identifier`, so some of the legacy usage is not even consistent with the platform's own configured key.

Highest-density frontend files: `src/modules/analytics/components/common/composables/useAnalyticsUtils.ts`, `src/modules/common/composables/useChannel.ts`, `src/composables/useChannel.ts`, `src/modules/analytics/composables/usePDFReports.ts`, `src/modules/common/store/common-methods.ts`, `src/composables/usePlatform.ts`, `src/modules/common/composables/useSocialChannels.ts`, `src/modules/analytics/types/competitor.ts`, plus a large block of analytics views and their tests.

Flutter files that carry legacy keys: `lib/features/social_channels/data/social_connector_parser.dart`, `lib/features/social_channels/application/social_connection_controller.dart`, `lib/features/inbox/data/inbox_channels_parser.dart`, `lib/features/inbox/domain/inbox_account.dart`, `lib/features/planner/data/dtos/social_account_dto.dart`, `lib/features/composer/data/mappers/plan_hydrator.dart`, `lib/features/push_notifications/data/dtos/notification_account_dto.dart`, `lib/core/network/mock_api_interceptor.dart`.

## 3. The repeated-query problem

Every per-platform account model on the backend points at the **same** Mongo collection:

- `app/Models/Integrations/Platforms/Social/FacebookAccounts.php` → `social_integrations`
- `app/Models/Integrations/Platforms/Social/LinkedinAccounts.php` → `social_integrations`
- also `InstagramAccounts`, `TwitterAccounts`, `PinterestAccounts`, `YoutubeAccounts`, `GoogleMyBusinessAccounts`, `TumblrAccounts`, and `SocialIntegrations` itself

So "get the accounts for this workspace" is implemented as one query per platform against a single collection, instead of one query filtered by platform. `FacebookAccounts` alone is referenced at 52 call sites across 19 files. Any screen that shows a mixed set of accounts, the composer account picker, the planner, analytics overview, the inbox channel list, pays for that fan-out. `app/Services/SocialIntegrations/UnifiedSocialSyncService.php` already documents that the product moved to the unified collection, so the per-platform models are the leftover of a migration that stopped short of the read path.

## 4. Consequences of the mixed state

1. **Every consumer needs a lookup table.** Both the backend config and the frontend descriptor array exist purely to answer "which field do I read for this platform".
2. **New platforms are a config edit in three codebases.** The Telegram and Reddit feature docs both show `account_selection_fields` and the frontend `channels` array as required steps.
3. **Comparisons silently fail.** Where one side reads `platform_identifier` and the other reads `facebook_id`, a match that should succeed does not, and the symptom is an account that appears to be missing rather than an error.
4. **Nothing enforces it.** There is no single accessor that all three codebases go through, so a new feature can reintroduce a legacy key without failing anything.

## 5. What the change must not break

- Existing stored data. Legacy keys are persisted on documents, not only used in code, and other collections reference accounts by them. Removing a field from code is safe; removing it from stored data is a migration and needs its own decision.
- Pinterest, which has both `board_id` and `profile_id` in the config: a Pinterest board and a Pinterest profile are different things, so this is not purely a rename.
- GMB and Tumblr use `name`. Confirm `platform_identifier` on those accounts is populated and stable before switching reads.
- The public API and webhooks. Any field appearing in a documented response is a contract. Related existing deliverables: `public-analytics-api-rollout`, `api-legacy-endpoint-validations`, `api-platform-overrides-parity`, `public-webhooks`.
- Shareable analytics report links and PDF reports, which read accounts through the analytics utility composables that carry the highest legacy-key density.

## 6. Recommended shape (for the implementation stories)

- Backend: one accessor for "the identifier of this account", one query path for "the accounts of this workspace, filtered by platform", both used everywhere. Keep writing legacy keys until every reader is migrated, then decide separately on the data cleanup.
- Frontend: retire the per-platform `key` in the channel descriptor once all readers use `platform_identifier`, so the descriptor stops being a lookup table.
- Flutter: same, in the account DTOs and parsers.
- A guard, for example a lint rule or a test, that fails when a legacy key is reintroduced. Without it this state returns.

## 7. Design story

No `[Design]` story is needed for this epic. The frontend work is a refactor behind existing screens with no new or changed UI. If any screen's behaviour visibly changes, that is a regression, not a design change.
