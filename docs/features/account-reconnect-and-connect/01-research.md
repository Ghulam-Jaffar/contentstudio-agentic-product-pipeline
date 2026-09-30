# Research: Reconnect and connect accounts from wherever they're used

**Date:** 2026-09-29
**Scope decision:** the PO chose to go straight from design to epic + stories, so there is no competitor research, workflow doc or PRD for this epic. The design canvas and this codebase research are the reference.

**Design canvas:** https://claude.ai/artifact/APzGjGQbTCJn8rYJ5EHtur ("Account Reconnect & Connect"), iterated with the PO on 2026-09-28/29.

- Row 1: "Reconnect required" popover in Composer, Analytics, Inbox, with Reconnect now
- Row 2: `+` and settings in the Planner filter and the Analytics account dropdown
- Row 3: Connect Social Accounts modal with Publish / Inbox / Analytics columns, the Facebook `+` menu, and the modal opened from Inbox

---

## 1. What the PO asked for

1. **One message for an expired account, everywhere.** Today three surfaces say three different things about the same state. The copy must use the term *access token*, lead with the heading **Reconnect required**, and say in plain words what stops working in *that* module.
2. **A next action next to the message.** Wherever an account is flagged, show **Reconnect now**, which starts the reconnect for *that* account directly, plus a help link to the Helpin article `how-to-refresh-token-expiry-723358d4`, opened the way docs open in the app today (the Helpin widget).
3. **Connect from any account picker.** A `+` (opens the Connect Social Accounts modal) and a settings gear (goes to Settings > Social Accounts) in every account list, the same pattern as the Label and Campaign dropdowns.
4. **The connect modal keeps its layout** but shows where each platform works (Publish / Inbox / Analytics), explains what each Facebook / Instagram option unlocks, and puts the relevant platforms first when opened from a module's `+`. The PO rejected a two-step redesign as too many steps.

## 2. Today: every place an expired account is flagged

Paths relative to `contentstudio-frontend/src`. There is **no shared helper or component**: `['invalid','expired'].includes(account.validity)` is repeated in ~8 files, and the composer uses `usePlatform().disableAccount()` (`composables/usePlatform.ts:245-267`), which also disables Twitter-not-allowed accounts and IG Personal / FB Group accounts with no mobile device.

| Surface | File | Locale key | Current copy |
|---|---|---|---|
| Composer account aside | `components/UI/CheckBox/CstAccountCheckBox.vue:84-90` (via `modules/composer/features/account-selection/aside/AccountsListPanel.vue:55`) | `composer.account_checkbox.messages.token_invalidated` | "The access token for this social account has been invalidated, please reconnect to continue posting." |
| Composer first-comment / carousel dropdown | `modules/composer/components/AccountSelectionDropdown.vue:423-443` (only checks `invalid`) | `composer.account_selection.token_invalidated` | same |
| Composer publish validation toast | `composables/usePublishSocialValidation.ts:233-252` | `common.publish_mixin.access_token_invalid` | "Access token of the {integration} {type} <b>{name}</b> is invalid. To renew your token, please re-connect the account to ContentStudio." (**hardcoded brand**) |
| Automations (RSS, Evergreen, Bulk CSV) | `modules/publish/components/posting/social/AccountSelection.vue:84-91,344,1156` | `common.account_selection.token_invalidated_warning` | same as composer |
| Planner filter sidebar | `components/common/SocialAccountListItem.vue:25-32` | `planner.filter_sidebar.social_account_list_item.token_expired` | "Your Account Token is Expired" |
| Schedule / queue modal | `modules/common/components/schedule-post/AccountPickerPanel.vue:37,105-115`, `AccountPickerRow.vue:32` | reuses planner key | "Your Account Token is Expired" |
| Planner Update Post modal | `modules/planner_v2/components/UpdatePostModal.vue:135` | `planner.update_post_modal.tooltips.access_token_expired` | "Your access token has expired, please reconnect to continue posting." |
| Planner bulk edit | `planner_v2/components/bulk-edit/BulkEditSocialAccounts.vue:111` | `planner.bulk_edit.invalid_token_tooltip` | composer copy |
| Inbox filter sidebar | `modules/inbox-revamp/components/ChannelListItem.vue:15-22` | `inbox.channel.token_expired` | "Your Account Token is Expired" |
| Inbox chat send error | `inbox-revamp/components/ChatBubble.vue:387` | `inbox.chat.error.reasons.reconnect` | "Your account needs reconnecting (go to Settings)." |
| Analytics single-account dropdown | `modules/analytics/views/common/PlatformAccountSelect.vue:307-314` | `analytics.common.filter_bar.reconnect_tooltip` | "To view your updated data, you must reconnect your account." |
| Analytics multi-account select | `analytics/views/common/MultipleAccountSelect.vue:139-146` | `analytics.account_select.reconnect_tooltip` | same |
| Analytics per-platform banners | `analytics/views/<platform>/MainComponent.vue` (facebook:153,222 · instagram:144,207 · linkedin:163,232 · pinterest:211-214,322 · gmb:192,291 · meta_ads:199,298 · google_ads:200,301 · tiktok:59 · youtube) | `analytics.<platform>.main*.{banner,status}.token_expired`, `actions.reconnect_now` | "Your {Platform} account token has expired! Your analytics were last updated on {date}. You need to reconnect your account." GBP differs. |
| Analytics overview alert | `analytics/views/overview/MainComponent.vue:218-232` | `analytics.overview.main.alerts.token_expired_multiple/_single` | "Access tokens for some of your social accounts have expired…" |
| Dashboard accounts card | `components/dashboard/SocialAccountsCard.vue:41-48` | `dashboard.social_accounts_card.token_expired` | "Access token for '{name}' has expired." |
| Header notification slider | `modules/common/components/header-notifications/HeadNotificationSlider.vue:27,124,207,228` | `header.notifications.*` | "Access Token Expired – Reconnect Required" (em dash) |
| Settings > Social Accounts | `modules/integration/components/platforms/social_v2/components/SocialAccountsDatatable.vue` | `settings.integrations.social.table.*` | already has its own Reconnect action, out of scope |
| Settings content categories | `modules/setting/components/content-categories/dialogs/AddCategory.vue:221,253` | `settings.add_category.tooltips.token_invalidated` | composer copy |

**Where "Reconnect Now" goes today:** most analytics banners, the Inbox IG banner and the dashboard card open the *generic* `social-connect-modal`, not a reconnect of the flagged account. Meta Ads / Google Ads banners navigate to Settings. The header banner goes to Settings `?filter=expired`.

**Not in scope, different problem:** Threads analytics permission missing (`accountSelectorHelpers.ts:128-143`) and LinkedIn profile "Analytics not connected" (`PlatformAccountSelect.vue:294-305`). Those are missing-permission prompts, not expired tokens.

The warning icon is an inline `<Icon name="TriangleAlert">` with `v-tooltip`, orange in some files and red in others. The only shared row is `components/common/SocialAccountListItem.vue`.

## 3. The reconnect flow that already exists (reuse it)

- Entry: `handleReconnectAccount` in `modules/integration/components/platforms/social_v2/composables/useSocialAccounts.ts:1413-1553`, called from the Settings datatable.
- Per platform: Bluesky opens `bluesky-reconnect-modal`; Telegram does nothing; Instagram via FB / IG login `:1429-1456`; meta_ads `:1459`; Facebook Profile `facebook-profile` `:1481`; Twitter custom app `:1505`; everything else (FB pages/groups, LinkedIn, Pinterest, GBP, YouTube, TikTok, Threads, Tumblr as `tumblr_social`) `:1514-1545`.
- It posts `{ process: 'reconnect', connector_id, type?, callback_url: location.href + '#platform' }` to BE `/getAuthorizationUrl` (`contentstudio-backend/app/Http/Controllers/Integrations/IntegrationController.php:466-505`), then does a **full-page redirect** to the platform.
- **Return-to already works.** `IntegrationBuilder::getRedirectLink()` (`contentstudio-backend/app/Strategy/Integrations/IntegrationBuilder.php:189-218`) sends the user back to `callback_url` when present. `SocialConnectHost.vue` is mounted globally (`Home.vue:1120`) and finishes the reconnect on whatever page the user lands on, then shows the `account_reconnected_success` toast.
- Precedent for returning to the originating page: LinkedIn analytics reconnect in `modules/analytics/views/linkedin/components/LinkedinProfileAnalyticsPrompt.vue:26-54`.
- **Composer risk:** autosave runs every 30s (`modules/composer/views/social-modal/useComposerAutosaveTimer.ts`), `minimizeComposer()` saves then hides (`useComposerConductor.ts:686-697`, EventBus `minimize-composer`). There is **no `beforeunload` anywhere in FE**, so a full-page OAuth redirect from the composer loses anything not yet autosaved. Label / Campaign gear already confirm-then-minimize before navigating (`components/common/LabelAttachment.vue:780-806`, `CampaignAttachment.vue:525-550`, copy `common.label_attachment.composer_warning_title/_message`). Possible bug: `if (isDraftComposer)` tests the Ref itself, always truthy (uncertain).

**Suggested:** extract a `useReconnectAccount(account, { source })` composable from `handleReconnectAccount` so every surface calls the same thing, and a single `ReconnectRequiredPopover` component with a `module` prop that picks the consequence sentence.

## 4. Account pickers that get `+` and settings

| Surface | Component |
|---|---|
| Composer account selection | `modules/composer/features/account-selection/AccountSelectionAside.vue`, header `aside/AccountSelectionHeaderBar.vue` |
| Planner filter sidebar | `modules/planner_v2/components/SocialAccountsDrawer.vue` |
| Schedule post modal | `modules/common/components/schedule-post/AccountPickerPanel.vue` |
| Analytics | `analytics/views/common/PlatformAccountSelect.vue` (single), `MultipleAccountSelect.vue` (overview) |
| Inbox filter | `modules/inbox-revamp/components/FilterDrawer.vue` |
| Automations | `modules/publish/components/posting/social/AccountSelection.vue` (RSS `SaveRSSAutomation.vue:115`, Evergreen `EvergreenAccountSelection.vue:214`, Bulk CSV `BulkUploadAutomationSave.vue:592`) |
| Left out on purpose | first-comment / carousel dropdown, planner bulk edit, `CstSocialAccountPicker` (AI chat input, team access, report modal), custom-view selector: small pickers where a connect shortcut is noise |

Label / Campaign precedent: `LabelAttachment.vue` `+` = `CirclePlus` (`:184-193`), gear = `Settings` icon hidden for approvers (`v-if="!isApprover"`, `:214-224`), tooltip `common.label_attachment.manage_labels`, route `{ name: 'miscellaneous', params: { id: 'labels' } }`.

Settings social accounts route: name `'social'`, path `/:workspace/settings/social/:id?` (`modules/setting/config/routes/setting.ts:122,195-201`). `?filter=expired` supported.

## 5. The Connect Social Accounts modal

- `modules/common/components/dialogs/SocialConnectModal.vue`, id `social-connect-modal`, mounted once in `Home.vue:1119`. Opened anywhere with `$cstuModal.show('social-connect-modal')` (~37 call sites). Props: `title`, `isEasyConnect`, `isTwitterAllowed`, `token`, `externalLinkId`.
- `modules/common/composables/useSocialConnectHighlight.ts`: global `targetPlatform` ref, when set before `show()` the modal scrolls to and highlights that row (used by `AnalyticsOnboardingEmptyState.vue:284`). **Natural hook for context-aware ordering.**
- Row: `modules/account/views/onboarding/SocialPlatform.vue` (labels, subtitles, `+` menus hardcoded in the template). Connect: `modules/account/composables/useSocialAccountsModal.ts:264-326`.
- Platform list: `CONNECTABLE_PLATFORMS` (`modules/composer/shared/constants/platforms.ts:84`) = `{ name, label, types, accounts }`, **no capability metadata anywhere**. Capability knowledge is scattered across inbox (`inbox-revamp/composables/useInboxUI.ts:161`), analytics (`analytics/components/common/composables/useAnalyticsRoutes.ts:66`), social-inbox-manager (`app/api/v1/routes/sync.py:158`) and BE (`app/Helpers/GlobalHelpers.php:284`). **Suggested:** add `capabilities` to `CONNECTABLE_PLATFORMS` so the modal has one source.
- Drift found: `SocialConnectBody.vue:119-123` hides WhatsApp without the `whatsapp` flag, `SocialConnectModal.vue:141-147` only hides it in EasyConnect. Unflagged users probably see WhatsApp in the live modal. Not fixed in this epic, flag for the devs.

### Capability matrix (what the columns show)

From code, 2026-09-28. `(?)` = not confirmed, verify before release.

| Platform / type | Publish | Inbox | Analytics |
|---|---|---|---|
| Facebook Page | Yes | Comments + messages | Yes |
| Facebook Profile / Group | Mobile notification only | No | No |
| Instagram (via FB or direct) | Yes | Comments + messages | Yes (direct vs via FB difference not found (?)) |
| Threads | Yes | Comments (SIM worker unmerged, `origin/feat/cont-threads-inbox` (?)) | Behind `threads_analytics` flag |
| X | Yes (needs CS app / custom app / add-on) | No | Only when analytics enabled on the X app |
| LinkedIn Page | Yes | Comments | Yes |
| LinkedIn Profile | Yes | No | Behind `linkedin_profile_analytics` + extra approval |
| Pinterest | Boards only | No | Yes (keyed by profile) |
| GBP | Yes | Reviews | Yes |
| YouTube | Yes | Comments | Yes |
| TikTok | Yes | No | Yes (Personal vs Business difference not found (?)) |
| Tumblr | Yes | No | No |
| Bluesky | Yes | No | Yes |
| Telegram | Yes | No | No |
| Meta Ads / Google Ads | No | No | Ads performance only (`ad_analytics` plan feature) |
| WhatsApp | No | Messages (`whatsapp` flag) | No |

## 6. Permissions and limits

- Connect / reconnect permission: BE `save_social`, FE `can_save_social`. Super admin, admin: yes. Collaborator: only with `permissions.addSocial`. Approver: no, and approvers are blocked from settings routes (`composables/usePermission.ts:268-294`, `router.ts:957-971`).
- **BE checks permission only at OAuth callback time** (`contentstudio-backend/app/Http/Requests/Integrations/ConnectAccountsRequest.php:21-37`, `IntegrationBuilder.php:120-145`). `/getAuthorizationUrl` does no check (`GetAuthorizationUrlRequest.php:14-17` returns true), so a user without permission gets sent through the whole platform sign-in and only fails on return.
- **`callback_url` is validated only as `required|string`**, no domain allowlist found. Once every picker and popover sends one, this becomes a real open-redirect surface. Worth fixing in this epic.
- Account limit is checked in BE on save (`app/Repository/Integrations/Platforms/SocialRepo.php:464`), reconnecting an existing account skips it. FE handles `limitExceed` after the fact via `modules/billing/composables/useAutoScaleConnect.ts` / `showUpgradeModal()`. Reconnect is unaffected by the limit.

## 7. White-label

`openHelpinArticle` returns early on white-label (`services/helpin.ts:147`), so a `data-helpin-article` link silently does nothing there. The "Why did this happen?" link must be **hidden** on white-label, not left dead. `isWhiteLabelDomain()` in `config/api-utils.ts:1-7`. The publish toast hardcodes "ContentStudio", which the new copy drops.

## 8. Analytics events

`connected_social_accounts` already fires for both connect and reconnect with `{ platform }` (`useSocialAccountsModal.ts:171`, `useSocialAccounts.ts:723,870`). There is no way to tell a reconnect from a new connection, or where it started. Proposed: keep the event, add `process: 'connect' | 'reconnect'` and `source` (where the user started). Flutter already has `social_account_reconnected` (`contentstudio-flutter/lib/core/analytics/analytics_event.dart:243`).

## 9. Developer surfaces

Public API already has `POST /v1/workspaces/{workspace_id}/connect/{platform}` with `process=connect|reconnect` and `return_url`, under `save_social` (`contentstudio-backend/routes/api/v1.php:823-826`, `Api/V1/ConnectController.php`). Nothing in this epic changes it. Its `return_url` is intentionally external (API clients return to their own apps).

## 10. Mobile

Out of scope, the request was web. Flutter has its own expiry check (`lib/features/planner/domain/social_account.dart:109-119`) and reconnect action (`lib/features/social_channels/presentation/reconnect_action.dart`). If the PO wants the same copy in the app, that is one follow-up `[Flutter]` story.

## 11. Open questions for the devs

1. Threads inbox comments: is the social-inbox-manager worker merged by the time this ships? If not, the Threads Inbox cell shows "No".
2. Instagram direct vs via Facebook, and TikTok Personal vs Business: any capability difference?
3. Reopening the Composer with the saved post after the OAuth return: the return lands on the same URL, but the Composer is a modal, so the draft id needs to ride along (hash or query) for it to reopen automatically.
