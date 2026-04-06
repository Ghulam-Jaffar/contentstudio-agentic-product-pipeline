# Reddit Publishing — Epic & Stories
*Feature Pipeline · Step 4 · Generated 2026-04-06*

---

## Epic

**Title:** Reddit Publishing Integration
**Epic ID (existing):** 115076
**Link:** https://app.shortcut.com/contentstudio-team/epic/115076
**Description:**
Reddit Publishing adds Reddit as a fully supported social channel in ContentStudio. Users can connect their Reddit accounts via OAuth 2.0 and schedule text posts, link posts, and image posts to any subreddit — complete with subreddit autocomplete, dynamic flair selection with mandatory enforcement, and multi-subreddit staggered scheduling to stay within Reddit's spam detection rules.

The feature integrates across the entire ContentStudio product surface: Composer, Planner/Calendar, Automations (RSS, Evergreen, Bulk CSV), Onboarding, Analytics dashboard, and the unified Inbox (web, iOS, and Android). Reddit poll post support is scoped to v2 but stories are written now.

Pre-requisite: Reddit Commercial API approval must be obtained before this feature launches (apply at developers.reddit.com).

---

## Stories (16 total)

---

## Story 1

### Title: [BE] Add Reddit OAuth 2.0 account connection, token management, and platform configuration

### Description:
As a social media manager, I want to connect my Reddit account to ContentStudio so that I can schedule and publish posts to subreddits from within ContentStudio just like I do with LinkedIn or Twitter.

This story implements all backend infrastructure for Reddit account connection: the OAuth 2.0 flow, token storage, silent token refresh, platform configuration, and the core architectural additions (Connector factory, config, routes, model fields) that all other Reddit stories depend on.

**Reference implementations:**
- `app/Strategy/Integrations/ThreadsConnector.php` — OAuth 2.0 connector pattern to follow
- `app/Strategy/Integrations/Connector.php` — factory to update
- `config/integrations.php` — social_integrations and social_channels arrays to update
- `app/Models/Integrations/Platforms/SocialIntegrations.php` — fillable fields to add

---

### Workflow:
1. User navigates to Settings → Social Accounts
2. User clicks "Connect Social Account" and selects Reddit from the platform list
3. ContentStudio redirects the user to Reddit's authorization page (`https://www.reddit.com/api/v1/authorize`) with scopes: `identity submit read flair mysubreddits history`, duration: `permanent`
4. User reviews Reddit's permission screen and clicks "Allow"
5. Reddit redirects back to ContentStudio's callback URL with an authorization code
6. ContentStudio exchanges the authorization code for an `access_token` + `refresh_token` via `POST https://www.reddit.com/api/v1/access_token`
7. ContentStudio fetches the user's Reddit profile (`GET /api/v1/me`): username, avatar, total karma, account age
8. ContentStudio saves the account to the `SocialIntegrations` collection and the user sees the connected account in their Social Accounts table

---

### Acceptance criteria:
- [ ] `RedditConnector.php` created at `app/Strategy/Integrations/RedditConnector.php` implementing `ConnectionInterface`
- [ ] `fetchAuthorizationLink()` builds the Reddit OAuth URL with: `client_id`, `response_type=code`, `state`, `redirect_uri`, `duration=permanent`, `scope=identity+submit+read+flair+mysubreddits+history`
- [ ] Authorization code is successfully exchanged for `access_token` + `refresh_token` via POST to `https://www.reddit.com/api/v1/access_token` using HTTP Basic Auth (client_id:client_secret)
- [ ] User profile fetched from `GET https://oauth.reddit.com/api/v1/me`: username, icon_img, total_karma, created_utc stored
- [ ] Account saved to `SocialIntegrations` with: `platform_type: 'reddit'`, `platform_identifier` (Reddit user ID), `platform_name` (e.g. "u/brandname"), `platform_logo`, `access_token` (encrypted), `refresh_token` (encrypted), `token_expires_at` (now + 3600 seconds), `validity: 'valid'`
- [ ] Token encryption uses existing `SocialHelper::encryptToken()` accessor pattern (same as all other platforms)
- [ ] Access token is silently refreshed when within 5 minutes of expiry using `refresh_token` via `POST https://www.reddit.com/api/v1/access_token` with `grant_type=refresh_token`
- [ ] When refresh token is invalid/revoked, account is marked `validity: 'expired'` and `validity_error` is set with the error message
- [ ] Attempting to connect a Reddit account that is already connected in the same workspace returns a clear error (no duplicate documents created)
- [ ] `case 'reddit': $this->connector = new RedditConnector();` added to `app/Strategy/Integrations/Connector.php`
- [ ] `reddit` config block added to `config/integrations.php` under `social_integrations`: `authorization_type: 'oauth2'`, `client_id`, `client_secret`, `redirect_uri`, `scope`, `authorization_url`, `token_url`, `api_url: 'https://oauth.reddit.com/'`
- [ ] `"reddit"` added to `social_channels` array in `config/integrations.php`
- [ ] `'reddit'` added to `post_update_disallowed_platforms` in `config/integrations.php` (Reddit posts cannot be edited after publishing)
- [ ] Reddit OAuth callback route registered in `routes/web/integrations.php`
- [ ] `reddit_username`, `reddit_karma`, `reddit_account_age_days` added to `SocialIntegrations` fillable fields
- [ ] Reddit entry added to `socialAccountsValidationConfig.php`: `["initializer" => "Reddit", "name" => "reddit", "key" => "platform_identifier"]`
- [ ] `REDDIT_CLIENT_ID`, `REDDIT_CLIENT_SECRET`, `REDDIT_REDIRECT_URI` added to `.env.example`

---

### Mock-ups:
N/A — backend only

---

### Impact on existing data:
New `SocialIntegrations` documents created with `platform_type: 'reddit'`. No changes to existing documents. New fields in fillable array are additive only.

---

### Impact on other products:
This is the foundation for all other Reddit stories. Once live: Reddit accounts become connectable in Social Accounts settings, selectable in the Composer, and available in Automations and Onboarding. No impact on existing platform integrations.

---

### Dependencies:
None — this is the foundation story for all Reddit work.

---

### Global quality & compliance:
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support — N/A, no user-facing strings in backend
- [ ] UI theming support — N/A, backend-only
- [ ] White-label domains impact review — OAuth redirect URI must be white-label domain aware if EasyConnect supports white-label domains
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension) — Reddit accounts will be returned in the social accounts API used by all clients

---

---

## Story 2

### Title: [BE] Implement Reddit post publishing (text, link, image) with multi-subreddit staggered scheduling

### Description:
As a social media manager, I want my scheduled Reddit posts to be published to the correct subreddits at the right time, with multiple subreddit targets staggered automatically so my posts aren't flagged as spam by Reddit.

This story implements the `RedditPosting` class and integrates it into ContentStudio's publishing pipeline. It handles all three v1 post types (text, link, image), flair application, multi-subreddit staggered scheduling (30 min per subreddit), and user-facing error messages for all Reddit API failure modes.

**Reference implementations:**
- `app/Strategy/Planner/BlueskyPosting.php` — newest posting class, follow this pattern
- `app/Strategy/Planner/Posting.php` — factory to update
- `app/Libraries/Publish/Posting/SocialPosting.php` — orchestrator to update (add after Bluesky conditional block)

---

### Workflow:
1. User has composed a Reddit post in the Composer: post type selected, title entered, subreddit(s) chosen, optional flair selected
2. User clicks "Schedule" (or "Publish Now") — ContentStudio saves the plan with `reddit_sharing_details` and `account_selection.reddit`
3. At the scheduled time, the publishing job fires
4. ContentStudio publishes to the first subreddit via `POST https://oauth.reddit.com/api/submit`
5. On success, the direct Reddit post URL is stored and the Planner shows "Published"
6. If multiple subreddits were selected, the second subreddit is queued 30 minutes later, the third 60 minutes later, and so on
7. If publishing fails, the user receives a notification with a specific, actionable error message

---

### Acceptance criteria:
- [ ] `RedditPosting.php` created at `app/Strategy/Planner/RedditPosting.php` implementing `PostingInterface`
- [ ] `case 'Reddit': $this->post = new RedditPosting($plan);` added to `app/Strategy/Planner/Posting.php`
- [ ] Reddit posting dispatcher added to `SocialPosting::processSocialPosting()`: `if (($plan['account_selection']['reddit'] ?? false) && is_array($plan['account_selection']['reddit']) && sizeof($plan['account_selection']['reddit'])) { ... }`
- [ ] `initializePosting()` reads from `plan['reddit_sharing_details']` or falls back to `plan['common_sharing_details']` when `common_box_status` is true
- [ ] `reddit_sharing_details` field added to `Plans` model fillable: `{ title, message, url, image, subreddits: [{name, flair_id, flair_text}], send_replies, oc }`
- [ ] Text posts published via `POST /api/submit` with `kind: 'self'`, `sr`, `title`, `text`, `sendreplies`, `nsfw: false`
- [ ] Link posts published via `POST /api/submit` with `kind: 'link'`, `sr`, `title`, `url`, `sendreplies`
- [ ] Image posts: image URL fetched from media, then `POST /api/submit` with `kind: 'image'`, `sr`, `title`, `url` (direct image link)
- [ ] `flair_id` and `flair_text` applied when present in `subreddits[].flair_id`
- [ ] `oc: true` flag applied when `reddit_sharing_details.oc` is true
- [ ] Multi-subreddit staggering: first subreddit posts at scheduled time; each subsequent subreddit is queued 30 minutes later via the existing job queue
- [ ] After each successful publish, `posting_response` updated with: `platform: 'reddit'`, `subreddit`, `posted_id` (Reddit post ID), `link` (full Reddit URL: `https://www.reddit.com/r/{sr}/comments/{post_id}`)
- [ ] `ALREADY_SUB` API error mapped to message: "This URL was already submitted to r/{subreddit} recently. Reddit prevents duplicate posts within a short window. Try a different subreddit or wait before resubmitting."
- [ ] `FLAIR_REQUIRED` error mapped to: "This subreddit requires a flair to be selected before posting."
- [ ] `SUBREDDIT_NOEXIST` / `SUBREDDIT_NOTALLOWED` error mapped to: "r/{subreddit} doesn't exist or is not accessible."
- [ ] Karma/eligibility error mapped to: "Your Reddit account doesn't meet r/{subreddit}'s posting requirements. The subreddit may require a minimum karma score or account age."
- [ ] `NOT_WHITELISTED_BY_USER_IN_SUBREDDIT` error mapped to: "r/{subreddit} is a private or restricted community. Your account must be approved by the moderators to post there."
- [ ] Token auto-refreshed before each publish attempt if within 5 minutes of expiry
- [ ] Rate limit (429) handled with exponential backoff and retry (max 3 attempts)

---

### Mock-ups:
N/A — backend only

---

### Impact on existing data:
- `Plans` collection gets new field `reddit_sharing_details` (additive, no migration needed)
- `posting_response` array on plan documents updated with Reddit results after publishing

---

### Impact on other products:
Reddit posts will appear in the Planner calendar after publishing. Analytics job (see **[BE] Implement Reddit post analytics sync job**) consumes the `posted_id` stored here.

---

### Dependencies:
Depends on: **[BE] Add Reddit OAuth 2.0 account connection, token management, and platform configuration**

---

### Global quality & compliance:
- [ ] Mobile responsiveness — N/A, backend-only
- [ ] Multilingual support — Error messages returned to frontend must be translatable strings; use i18n keys
- [ ] UI theming support — N/A, backend-only
- [ ] White-label domains impact review — Publishing pipeline is workspace-agnostic; no white-label impact
- [ ] Cross-product impact assessment — Reddit posts publish through the same job queue as all other platforms; queue load should be assessed at scale

---

---

## Story 3

### Title: [BE] Add Reddit subreddit search and flair fetching API endpoints

### Description:
As a social media manager, I want to search for subreddits and see their available flairs directly inside the ContentStudio composer, so I can target the right community and meet flair requirements without leaving the app.

This story exposes two lightweight internal API endpoints that proxy Reddit's subreddit search and flair APIs, with caching to minimize Reddit API call costs.

---

### Workflow:
1. User types in the Subreddit field in the Composer (after typing 2+ characters)
2. ContentStudio calls the subreddit search endpoint and returns matching subreddits with subscriber counts
3. User selects a subreddit
4. ContentStudio calls the flair endpoint for that subreddit and populates the flair dropdown
5. If the subreddit requires flair, the response indicates this so the frontend can block publishing until one is selected

---

### Acceptance criteria:
- [ ] `GET /api/reddit/subreddits/search?q={query}&workspace_id={id}` returns: `[{ name, title, subscribers, description, icon_img, subreddit_type, over18 }]`
- [ ] Subreddit search proxies `GET https://oauth.reddit.com/subreddits/search?q={query}&limit=10&include_over_18=false` using any valid Reddit account token from the workspace
- [ ] `GET /api/reddit/subreddits/{subreddit}/flairs?workspace_id={id}` proxies `GET /r/{subreddit}/api/link_flair_v2`
- [ ] Flair response: `[{ id, text, css_class, background_color, text_color, mod_only, allowable_content }]`
- [ ] Flair response includes `flair_required: true/false` — detected from subreddit's `subreddit_type` and whether `link_flair_select` is mandatory (via `GET /r/{subreddit}/about`)
- [ ] Subreddit search results cached per query per workspace for 10 minutes (Redis)
- [ ] Flair results cached per subreddit per workspace for 5 minutes (Redis)
- [ ] Both endpoints return empty arrays (not 500 errors) when no results found or subreddit has no flairs
- [ ] `over18: true` subreddits filtered out from search results (NSFW communities not supported)
- [ ] Endpoints are protected by workspace authentication middleware
- [ ] If no Reddit account is connected to the workspace, endpoints return 422 with message: "No Reddit account connected to this workspace."

---

### Mock-ups:
N/A — backend only

---

### Impact on existing data:
New Redis cache keys for subreddit search and flair data. No database changes.

---

### Impact on other products:
None — new internal endpoints only used by the Reddit composer section.

---

### Dependencies:
Depends on: **[BE] Add Reddit OAuth 2.0 account connection, token management, and platform configuration**

---

### Global quality & compliance:
- [ ] Mobile responsiveness — N/A, backend-only
- [ ] Multilingual support — N/A, no user-facing strings
- [ ] UI theming support — N/A, backend-only
- [ ] White-label domains impact review — API endpoints are workspace-scoped; no white-label impact
- [ ] Cross-product impact assessment — New endpoints only; no existing endpoints changed

---

---

## Story 4

### Title: [BE] Add Reddit to automation platform configuration (RSS, Evergreen, Bulk CSV)

### Description:
As a social media manager, I want to be able to include my Reddit accounts in RSS automations and Evergreen automations, so that I can automatically publish RSS feed articles to subreddits and recycle evergreen content to Reddit on a schedule.

This story adds Reddit to the backend automation platform configuration so that all three automation types (RSS, Evergreen, Bulk CSV) recognize and process Reddit account selections.

**Reference files:**
- `app/Repository/Publish/Automation/RssAutomationRepository.php` — add reddit platform query block
- `config/social_platforms.php` — add 'reddit' to platforms array

---

### Workflow:
1. User creates a new RSS Automation and opens the social account selector
2. User sees their connected Reddit account(s) alongside Facebook, Twitter, etc.
3. User selects a Reddit account and specifies a subreddit for RSS posts
4. When the RSS feed publishes a new article, ContentStudio automatically posts a link post to the configured subreddit
5. Same applies for Evergreen automations — user can select Reddit accounts and subreddits in the account selection step

---

### Acceptance criteria:
- [ ] `'reddit'` added to the platforms array in `config/social_platforms.php`
- [ ] `'reddit.platform_identifier'` added to `account_selection_fields` in `config/social_platforms.php`
- [ ] `RssAutomationRepository.php` updated to filter automation plans that include reddit accounts in `account_selection.reddit` (same `orWhere` pattern as other platforms)
- [ ] RSS automation posts to Reddit go through the same `RedditPosting` pipeline as manually scheduled posts
- [ ] Evergreen automation's account selection validation in `EvergreenAutomationRepository.php` (or equivalent) includes reddit
- [ ] Reddit in RSS automation creates link posts by default (RSS items have a URL) using the RSS item title as the Reddit post title
- [ ] `reddit` entry added to `socialAccountsValidationConfig.php` if not already present from Story 1
- [ ] Bulk CSV upload automation works with Reddit accounts (no additional backend changes needed beyond platform config)

---

### Mock-ups:
N/A — backend only

---

### Impact on existing data:
New `account_selection.reddit` field in automation documents (additive, no migration needed).

---

### Impact on other products:
RSS automations publishing to Reddit go through the same publishing queue. Increased API calls to Reddit should be factored into the commercial API usage estimate.

---

### Dependencies:
Depends on: **[BE] Add Reddit OAuth 2.0 account connection, token management, and platform configuration**
Depends on: **[BE] Implement Reddit post publishing (text, link, image) with multi-subreddit staggered scheduling**

---

### Global quality & compliance:
- [ ] Mobile responsiveness — N/A, backend-only
- [ ] Multilingual support — N/A, no user-facing strings
- [ ] UI theming support — N/A, backend-only
- [ ] White-label domains impact review — Automation publishing is workspace-scoped; no white-label impact
- [ ] Cross-product impact assessment — Automations now publish to Reddit; test that existing RSS/Evergreen automations for other platforms are unaffected

---

---

## Story 5

### Title: [BE] Implement Reddit post analytics sync job

### Description:
As a social media manager, I want to see how my Reddit posts are performing (upvotes, comments) inside ContentStudio's Analytics dashboard, so I don't have to switch to Reddit manually to check engagement.

This story implements a background job that polls Reddit's API for engagement data on posts published through ContentStudio and stores the metrics for display in the Analytics dashboard.

---

### Workflow:
1. User publishes a Reddit post via ContentStudio
2. ContentStudio stores the Reddit post ID from the API response
3. Every 6 hours, a background job fetches updated metrics for all published Reddit posts
4. In Analytics, the user sees upvote score and comment count per post
5. The user can click a post to go directly to Reddit

---

### Acceptance criteria:
- [ ] Background job queries all `Plans` with `posting_response[].platform = 'reddit'` and `posting_response[].posted_id` present
- [ ] Job fetches post data from `GET https://oauth.reddit.com/api/info?id=t3_{post_id}` for each post
- [ ] Metrics stored on the posting_response entry: `score` (upvotes), `num_comments`, `upvote_ratio`, `url`
- [ ] Job runs every 6 hours (configurable via env)
- [ ] Deleted posts (Reddit returns empty listing) are handled gracefully: marked as `deleted: true` in posting_response, polling stops for that post
- [ ] API requests distributed across the 6-hour window to avoid rate limit spikes (max 60–100 req/min)
- [ ] Analytics API endpoint exposes Reddit account metrics: `total_posts_published`, `total_upvotes`, `total_comments`, `top_posts` (sorted by score, top 10)
- [ ] Per-post metrics returned in the Planner post detail API response alongside the existing `posting_response` data

---

### Mock-ups:
N/A — backend only

---

### Impact on existing data:
Additional fields (`score`, `num_comments`, `upvote_ratio`) written to existing `posting_response` sub-documents in the Plans collection. Additive only.

---

### Impact on other products:
Analytics dashboard (see **[FE] Add Reddit analytics to the Analytics dashboard**) consumes this data. No impact on other platforms' analytics.

---

### Dependencies:
Depends on: **[BE] Implement Reddit post publishing (text, link, image) with multi-subreddit staggered scheduling**

---

### Global quality & compliance:
- [ ] Mobile responsiveness — N/A, backend-only
- [ ] Multilingual support — N/A, no user-facing strings
- [ ] UI theming support — N/A, backend-only
- [ ] White-label domains impact review — Analytics data is workspace-scoped; no white-label impact
- [ ] Cross-product impact assessment — Analytics sync job shares the job queue; assess queue load impact

---

---

## Story 6

### Title: [BE] Implement Reddit inbox strategy for post comments

### Description:
As a social media manager, I want to see comments left on my Reddit posts in ContentStudio's Inbox, so I can engage with my community and reply to discussions from the same place I manage all other platform conversations.

This story implements the `StrategyReddit.php` inbox strategy and a background job that ingests new comments from published Reddit posts into ContentStudio's Inbox.

**Note:** Reddit's 3rd-party API does not support DMs or private messages. Inbox support for Reddit is limited to **comments on posts published via ContentStudio**.

---

### Workflow:
1. User publishes a Reddit post via ContentStudio
2. Reddit community members comment on the post
3. A background job syncs new comments into ContentStudio's Inbox collection
4. In the Inbox, the user sees a new item with the commenter's username, comment body, and subreddit
5. User clicks the comment → sees the original post title and the full comment thread
6. User types a reply and clicks "Send" → ContentStudio posts the reply as a comment on Reddit

---

### Acceptance criteria:
- [ ] `StrategyReddit.php` created at `app/Strategy/Conversation/StrategyReddit.php` implementing `PlatformInterface`
- [ ] `replyToPost()` implemented: `POST https://oauth.reddit.com/api/comment` with `parent=t3_{post_id}` and `text={reply_text}` — posts a top-level comment on the Reddit post
- [ ] `replyToConversation()` implemented to reply to an existing comment: `POST https://oauth.reddit.com/api/comment` with `parent=t1_{comment_id}` and `text={reply_text}`
- [ ] `deleteSocialDetails()` revokes the Reddit token on workspace account disconnection
- [ ] DM/message thread attempts return a clear unsupported error: Reddit DMs are not available via 3rd-party API
- [ ] Background sync job fetches new comments for each published Reddit post: `GET https://oauth.reddit.com/r/{subreddit}/comments/{post_id}`
- [ ] Sync job runs every 30 minutes for posts published in the last 7 days, every 6 hours for posts older than 7 days
- [ ] Comments ingested into Inbox collection with fields: `platform: 'reddit'`, `type: 'comment'`, `post_id`, `comment_id`, `author` (Reddit username), `body`, `score`, `created_utc`, `subreddit`, `parent_post_title`
- [ ] Duplicate comment detection: already-ingested comment IDs are not re-inserted
- [ ] Inbox API endpoint includes reddit comments in the unified inbox feed
- [ ] Comments deleted on Reddit are marked as `deleted: true` in the Inbox collection

---

### Mock-ups:
N/A — backend only

---

### Impact on existing data:
New Inbox collection documents with `platform: 'reddit'`. No changes to existing inbox documents.

---

### Impact on other products:
Powers web Inbox (see **[FE] Add Reddit comments to the unified Inbox**), iOS Inbox (see **[iOS] Add Reddit post comments support in iOS Inbox**), and Android Inbox (see **[Android] Add Reddit post comments support in Android Inbox**).

---

### Dependencies:
Depends on: **[BE] Implement Reddit post publishing (text, link, image) with multi-subreddit staggered scheduling**

---

### Global quality & compliance:
- [ ] Mobile responsiveness — N/A, backend-only
- [ ] Multilingual support — N/A, no user-facing strings in backend
- [ ] UI theming support — N/A, backend-only
- [ ] White-label domains impact review — Inbox is workspace-scoped; no white-label impact
- [ ] Cross-product impact assessment — Inbox sync job adds Reddit API calls; assess Reddit API rate limit usage alongside analytics sync job

---

---

## Story 7

### Title: [FE] Add Reddit account connection modal and Social Accounts settings integration

### Description:
As a social media manager, I want to connect my Reddit account to ContentStudio through a clear and trustworthy OAuth flow, and see it managed alongside my other social accounts in the Social Accounts settings, so I have a consistent experience across all platforms.

---

### Workflow:
1. User goes to Settings → Workspace Settings → Social Accounts
2. User clicks "Connect Social Account" — the "Connect Social Accounts" modal opens
3. Reddit is listed in the platform list with an orange Reddit icon, "(Profiles)" subtext, and a "+" button
4. If the workspace already has Reddit accounts connected, a count badge shows the number (e.g., "2")
5. User clicks "+" next to Reddit (or anywhere on the Reddit row) — the Reddit connection modal opens
6. User sees what ContentStudio can and cannot do with their Reddit account, then clicks "Connect with Reddit ↗"
7. Browser opens Reddit's authorization page in a popup/new tab
8. User clicks "Allow" on Reddit's permission screen
9. Reddit redirects back; ContentStudio shows a success state in the modal with the connected account card
10. User clicks "Done" — modal closes, Reddit account row appears in the Social Accounts table with "Token valid" status
11. If the token expires later, the row shows "Token expired" (red pill) and the user can click "↺ Reconnect" to re-run the same OAuth flow

---

### Acceptance criteria:
- [ ] Reddit row visible in the "Connect Social Accounts" modal platform list with: orange Reddit icon (`__brand__Reddit` from icon mapping), label "Reddit", subtext "(Profiles)", count badge showing number of connected accounts, "+" add button
- [ ] Clicking "+" on Reddit row opens the Reddit connection modal
- [ ] Reddit connection modal title: "Connect Your Reddit Account"
- [ ] Modal subtext: "You'll be redirected to Reddit to authorize ContentStudio. Once connected, you can schedule posts to any subreddit your account can access."
- [ ] Modal lists what ContentStudio can do: "Submit posts to subreddits", "Read your username and profile info", "View subreddits you moderate and subscribe to", "Read and apply post flairs"
- [ ] Modal lists what ContentStudio cannot do: "Vote on your behalf", "Access your direct messages", "Post to age-restricted (NSFW) communities"
- [ ] Primary CTA button label: "Connect with Reddit ↗" — uses `Button` component (primary variant)
- [ ] Secondary button label: "Cancel" — uses `Button` component (secondary/ghost variant)
- [ ] After OAuth completes and callback returns, modal transitions to success state showing: ✅ success ring icon, title "Account Connected!", subtext "Your Reddit account is now linked to ContentStudio. You can select it in the Composer when creating posts."
- [ ] Success state account card shows: Reddit avatar (from `icon_img`), username ("u/brandname"), karma count formatted (e.g., "1,240 karma"), account age ("3 years old"), "Reddit Profile" pill, "✓ Connected" green pill
- [ ] Success state buttons: "Connect Another" (returns to platform list modal) and "Done" (closes all modals)
- [ ] After clicking "Done", connected Reddit account appears as a new row in the Social Accounts table
- [ ] Table row shows: Reddit avatar, "u/username" as name, "Reddit Profile" as type, "Telegram" → "Reddit" as platform column, "Token valid" green pill
- [ ] Toast notification on success: "u/[username] connected to Reddit successfully!"
- [ ] If Reddit account already connected in workspace: error toast "u/[username] is already connected to this workspace."
- [ ] If user denies OAuth on Reddit: modal shows error state with message "Reddit authorization was denied. Please try again and click 'Allow' when prompted."
- [ ] "Token expired" (red pill) shown when `validity: 'expired'`; "↺ Reconnect" button re-opens the connection modal and runs the OAuth flow again
- [ ] `RedditAccount` TypeScript interface added to `src/types/common/social-accounts.ts`: `{ channel_type: 'reddit', platform_identifier: string, reddit_karma?: number }`
- [ ] `'reddit'` added to `ChannelType` union in `src/types/common/social-accounts.ts`
- [ ] `RedditAccount` added to `SocialAccount` union type
- [ ] `AddReddit.vue` component created at `src/modules/integration/components/dialogs/AddReddit.vue`
- [ ] `SaveSocialAccounts.vue` updated with reddit cases in all platform-specific switch statements (modalHeader, firstSectionHeader, firstSectionItems)
- [ ] All user-facing strings use `$t()` i18n keys

**UI component usage:**
- `Modal` for both modals (connection + success)
- `Button` (primary variant) for "Connect with Reddit ↗"
- `Button` (secondary/ghost) for "Cancel"
- `Avatar` for Reddit profile avatar in success card
- `Badge` for count badge and "Token valid"/"Token expired" status pills
- `Icon` (`__brand__Reddit`) for Reddit logo

**Empty/loading states:**
- While OAuth popup is open and waiting for callback: loading spinner inside the modal with text "Waiting for Reddit authorization…"
- If popup is blocked by browser: inline `Alert` (warning variant): "Please allow popups for ContentStudio to connect your Reddit account, then try again."

---

### Mock-ups:
See `docs/features/telegram-integration/telegram-integration-mockup.html` for the reference pattern for connection modals (same UX pattern, adapted for Reddit OAuth).

---

### Impact on existing data:
New `SocialIntegrations` records with `platform_type: 'reddit'`. No changes to existing accounts.

---

### Impact on other products:
Reddit accounts become available in: Composer Social tab, Automation channel selectors, Onboarding platform list (after **[FE] Add Reddit to automation channel selectors and onboarding platform list** is shipped).

---

### Dependencies:
Depends on: **[BE] Add Reddit OAuth 2.0 account connection, token management, and platform configuration**

---

### Global quality & compliance:
- [ ] Mobile responsiveness — Modal and table must be responsive on smaller screens; test at 768px and 375px breakpoints
- [ ] Multilingual support — All labels, tooltips, error messages, and toast notifications must use `$t()` i18n keys; add all keys to all locale files under `src/locales/`
- [ ] UI theming support — Use `text-primary-cs-500`, `bg-primary-cs-50`, `border-primary-cs-200` for all primary color usage; never hardcode `text-blue-600` or Reddit's `#FF4500` brand color
- [ ] White-label domains impact review — Reddit brand color (#FF4500) must NOT be used directly; use the Reddit icon only; all interactive elements use theme variables
- [ ] Cross-product impact assessment — Social Accounts table and connect modal are shared across web; test that existing platform rows are unaffected

---

---

## Story 8

### Title: [FE] Build Reddit composer section with post type selector, title field, subreddit search, flair dropdown, and multi-subreddit support

### Description:
As a social media manager, I want a dedicated Reddit section in the ContentStudio composer that lets me write a title, choose a post type (text, link, or image), search for and select a subreddit, pick a flair, and optionally add multiple subreddits — so I can craft the perfect Reddit post without any guesswork.

---

### Workflow:
1. User opens the Composer and selects a Reddit account in the Social tab
2. A "Reddit" platform-specific section appears below the common message box
3. User selects a post type using the segmented control: "Text Post", "Link Post", or "Image Post"
4. User enters a title in the Title field (required, 300-char limit shown as a live counter)
5. User types in the Subreddit field — autocomplete shows matching subreddits with subscriber counts after 2 characters
6. User selects a subreddit; the flair dropdown loads for that subreddit
7. If the subreddit has no flairs, the flair field shows "No flair available" and is disabled
8. If the subreddit requires flair, the flair field is labeled "(Required)" and the publish button stays disabled until one is chosen
9. User picks a flair (or skips if not required)
10. User optionally clicks "+ Add Subreddit" to target more subreddits (each with its own flair selector); stagger notice is shown
11. For text posts: body comes from the common message box (editable within the Reddit section)
12. For link posts: URL field appears (auto-populated from any link in the common box)
13. For image posts: the image from the media section is shown with a note that body text is not supported
14. User completes the post and schedules or publishes

---

### Acceptance criteria:
- [ ] Reddit section renders in the composer whenever at least one Reddit account is selected in the Social tab
- [ ] Reddit section header: "🟠 Reddit" with the Reddit icon badge
- [ ] Post type selector uses the `SegmentedControl` component with three segments: "Text Post" | "Link Post" | "Image Post"; defaults to "Text Post"
- [ ] Title field label: "Title" with asterisk (*) for required
- [ ] Title field placeholder: "Write a compelling title (e.g., 'How we grew our newsletter to 10k subscribers')"
- [ ] Title field helper text: "Keep it clear and direct — Reddit users scroll fast. 300 characters max."
- [ ] Title character counter shows live count (e.g., "87 / 300"); counter turns `text-red-500` when over 300
- [ ] Title validation error (shown inline below field): "Post title is required for Reddit."
- [ ] Title max-length error: "Reddit titles cannot exceed 300 characters."
- [ ] Subreddit selector field label: "Subreddit" with asterisk (*) for required
- [ ] Subreddit placeholder: "Search for a community (e.g. r/marketing, r/entrepreneur)"
- [ ] Subreddit autocomplete fires after 2 characters typed; shows dropdown with: subreddit name (bold), subscriber count (e.g. "1.2M members"), lock icon (🔒) for private/restricted subreddits
- [ ] Subreddit search uses `SearchInput` component; dropdown uses `Dropdown` + `DropdownItem` components
- [ ] "No subreddits found" empty state: small text "No subreddits matched your search. Try a different name."
- [ ] After subreddit is selected, flair dropdown appears below
- [ ] Flair field label: "Flair"
- [ ] Flair label shows "(Required)" in `text-red-500` when subreddit enforces mandatory flair
- [ ] Flair dropdown placeholder: "Select a flair…"
- [ ] Flair dropdown uses `Dropdown` + `DropdownItem` components; each item shows flair color swatch (as a 12px circle) + flair text
- [ ] When subreddit has no flairs: flair dropdown is hidden entirely (not shown as disabled)
- [ ] Flair required validation error (shown inline): "This subreddit requires a flair. Please select one before publishing."
- [ ] Publish button disabled when: title is empty, title > 300 chars, no subreddit selected, or flair required but not selected
- [ ] ℹ tooltip on the flair field: "Some subreddits require you to label your post with a category (called a 'flair') before you can publish. For example, r/entrepreneur might require you to tag your post as 'Advice', 'Story', or 'Resource'. If no flair is needed, this field won't appear."
- [ ] "Send Replies" toggle below subreddit section; uses `Switch` component; default: ON
- [ ] "Send Replies" label: "Send Replies"
- [ ] "Send Replies" ℹ tooltip: "When on, Reddit will notify you when someone replies to your post. Turn this off if you don't want email notifications from Reddit for this post."
- [ ] "OC (Original Content)" toggle; uses `Switch` component; default: OFF
- [ ] "OC" label: "Original Content"
- [ ] "OC" ℹ tooltip: "Mark this post as original content you created yourself — not a repost or reshared link. Some subreddits require this for certain types of posts."
- [ ] **Text Post:** Body textarea pre-filled with content from the common message box; uses `Textarea` component; label "Body (optional)"; placeholder "Add context, details, or a story (Markdown supported)…"; 40,000 char limit
- [ ] **Link Post:** URL field appears; label "URL"; placeholder "https://your-article-or-page.com"; required for link posts; auto-populates from any URL detected in the common message box
- [ ] **Image Post:** Shows the first image from the media section as a preview thumbnail; `Alert` component (info variant) shown below: "Reddit image posts don't support body text. Your image will be posted with the title only." — uses `text-primary-cs-500` and `bg-primary-cs-50` for theming
- [ ] "+ Add Subreddit" button (ghost variant `Button`) below the first subreddit row; clicking it adds another subreddit row (each with its own subreddit selector and flair selector)
- [ ] When 2+ subreddits are added, a stagger notice appears: `Alert` component (info variant): "Posts to multiple subreddits will be published 30 minutes apart to comply with Reddit's community guidelines and avoid spam detection."
- [ ] Each subreddit row shows an "×" remove button to delete that row; first row cannot be removed if it's the only one
- [ ] `reddit: []` added to platform initialization in `SocialTab.vue` (or equivalent composer platform data structure)
- [ ] All user-facing strings use `$t()` i18n keys

**UI component summary:**
- `SegmentedControl` — post type selector
- `TextInput` — title field
- `SearchInput` — subreddit search
- `Dropdown` + `DropdownItem` — subreddit autocomplete + flair dropdown
- `Textarea` — text post body
- `Switch` — Send Replies and OC toggles
- `Button` (ghost) — "+ Add Subreddit"
- `Alert` (info) — image post note and multi-subreddit stagger notice
- `Icon` — ℹ info icons, Reddit logo

**Loading states:**
- Subreddit autocomplete: `Loader` spinner inside the dropdown while fetching results
- Flair dropdown: `Loader` spinner while loading flairs ("Loading flairs…" text)

**Error states:**
- Subreddit not found: "No subreddits matched your search."
- Flair load failure: "Couldn't load flairs. Please try again." with a retry link
- No Reddit accounts connected: section hidden; instead shown as a `Alert` (warning variant) "No Reddit accounts connected. Connect one in Social Accounts settings."

---

### Mock-ups:
See PRD section 7 and workflow document (`docs/features/reddit-publishing/02-workflow.md`) for detailed UI layout descriptions.

---

### Impact on existing data:
`reddit_sharing_details` object saved to Plans documents when Reddit accounts are selected in the composer. Additive.

---

### Impact on other products:
This composer section is web-only. Reddit post data flows to: Planner calendar (shows Reddit badge), publishing pipeline, and analytics. Mobile apps do not get a Reddit-specific composer UI in v1.

---

### Dependencies:
Depends on: **[BE] Add Reddit OAuth 2.0 account connection, token management, and platform configuration**
Depends on: **[BE] Add Reddit subreddit search and flair fetching API endpoints**
Depends on: **[FE] Add Reddit account connection modal and Social Accounts settings integration**

---

### Global quality & compliance:
- [ ] Mobile responsiveness — Reddit composer section must reflow correctly on tablet (768px) and small screens; `SegmentedControl` must not overflow
- [ ] Multilingual support — All labels, placeholders, tooltips, validation messages, and `Alert` copy must use `$t()` i18n keys; add to all locale files
- [ ] UI theming support — Use `text-primary-cs-500`, `bg-primary-cs-50`, `border-primary-cs-200` for all primary color elements; no hardcoded colors
- [ ] White-label domains impact review — Reddit icon and brand color must not bleed into theme; all interactive states use CSS variable-backed classes
- [ ] Cross-product impact assessment — Composer change affects all publishing flows; ensure no regression in other platform sections

---

---

## Story 9

### Title: [FE] Add Reddit to automation channel selectors, Evergreen automation, and onboarding platform list

### Description:
As a social media manager, I want Reddit to appear as a selectable channel in RSS automations, Evergreen automations, and the onboarding setup flow, so I can include Reddit in my automated workflows from the start and when setting up my workspace for the first time.

---

### Workflow:
1. User creates a new RSS Automation and opens the social accounts dropdown — Reddit accounts appear alongside Facebook, Twitter, etc.
2. User selects a Reddit account, enters a subreddit for RSS posts
3. When the RSS feed fires, ContentStudio publishes a link post to the specified subreddit
4. Same: in Evergreen automation setup, user selects Reddit account and subreddit for recycled posts
5. Onboarding: when setting up a new workspace, Reddit is listed in the "Connect your social accounts" step alongside all other platforms

---

### Acceptance criteria:
- [ ] Reddit added to `channels` array in `src/modules/common/constants/common-attributes.js`: `{ name: 'reddit', key: 'platform_identifier', getter: 'getRedditAccounts' }`
- [ ] `'reddit'` added to `socialChannelsNameArray` in `common-attributes.js`
- [ ] RSS Automation account selector shows connected Reddit accounts (no additional frontend changes needed once `common-attributes.js` is updated — `AccountFilterDropdown.vue` uses this array automatically)
- [ ] Evergreen automation `EvergreenAccountSelection.vue` updated to include a watcher for `getAccountSelection.reddit` with the same validation logic as other platforms (lines 30–83 in `EvergreenAccountSelection.vue`)
- [ ] Reddit added to `SUPPORTED_PLATFORMS` array in `src/modules/composer_v2/composables/useComposerHelper.js`: `{ name: 'reddit', label: 'Reddit', types: ['Profiles'], accounts: [] }`
- [ ] Reddit platform card renders correctly in `SocialConnect.vue` during onboarding — shows Reddit icon, label, and "Connect" button
- [ ] Clicking "Connect" for Reddit during onboarding triggers the same OAuth flow as connecting from Social Accounts settings
- [ ] Reddit accounts appear in the bulk CSV upload automation account selector
- [ ] In automations, when Reddit is selected and no subreddit is configured, a validation error is shown: "Please specify a subreddit for your Reddit account." (subreddit is required for all Reddit posts)
- [ ] Subreddit input field appears in RSS automation settings when a Reddit account is selected: label "Subreddit", placeholder "e.g. r/marketing", helper text "Posts from this RSS feed will be published to this subreddit."
- [ ] All i18n strings use `$t()` keys; add to all locale files

**UI components:**
- `TextInput` — subreddit field in RSS automation settings
- `Alert` (warning) — validation message when Reddit account selected but no subreddit entered

---

### Mock-ups:
See existing RSS Automation settings UI for reference — Reddit follows the same pattern as other platforms in the account selector dropdown.

---

### Impact on existing data:
Automation documents get `account_selection.reddit` field when Reddit is added. No changes to existing automations.

---

### Impact on other products:
Automation publishing to Reddit uses the backend pipeline from **[BE] Add Reddit to automation platform configuration (RSS, Evergreen, Bulk CSV)**. Onboarding change is additive — no impact on existing platform onboarding flows.

---

### Dependencies:
Depends on: **[FE] Add Reddit account connection modal and Social Accounts settings integration**
Depends on: **[BE] Add Reddit to automation platform configuration (RSS, Evergreen, Bulk CSV)**

---

### Global quality & compliance:
- [ ] Mobile responsiveness — Onboarding platform card and automation dropdowns must be responsive
- [ ] Multilingual support — All labels and helper text use `$t()` i18n keys; add to all locale files
- [ ] UI theming support — Reddit card in onboarding uses theme-aware classes; no hardcoded colors
- [ ] White-label domains impact review — Onboarding platform list is visible to white-label customers; ensure Reddit card follows theming
- [ ] Cross-product impact assessment — Changes to `common-attributes.js` and `useComposerHelper.js` affect all automation types and the onboarding flow; test regression across RSS, Evergreen, and Bulk CSV

---

---

## Story 10

### Title: [FE] Add Reddit analytics to the Analytics dashboard

### Description:
As a social media manager, I want to see upvote and comment performance for my Reddit posts in ContentStudio's Analytics dashboard, so I can understand how well my Reddit content performs without having to log into Reddit separately.

---

### Workflow:
1. User goes to Analytics and selects a connected Reddit account
2. A Reddit analytics card appears showing: total posts published, total upvotes, total comments, top posts by upvote score
3. User can click any post title to open it directly on Reddit
4. Metrics refresh every 6 hours automatically

---

### Acceptance criteria:
- [ ] Reddit account selectable in the Analytics account filter alongside other platforms
- [ ] Reddit analytics card shows: "Total Posts" (count), "Total Upvotes" (sum of all post scores), "Total Comments" (sum of num_comments), "Avg. Upvote Ratio" (percentage)
- [ ] Top posts table columns: Post Title (truncated to 60 chars), Subreddit (e.g. "r/marketing"), Upvotes, Comments, Published Date, direct link icon (opens Reddit post in new tab)
- [ ] "Last synced" timestamp shown below card: "Data last updated [X hours] ago"
- [ ] Empty state (no Reddit posts yet): `Alert` (info variant) with headline "No Reddit posts published yet", subtext "Once you publish posts to Reddit via ContentStudio, your performance data will appear here.", CTA "Compose a Reddit Post" (links to Composer)
- [ ] Loading state: skeleton loaders for all metric cards and table rows while data loads
- [ ] If Reddit API sync failed for a post: show `-` in metrics with `CstPopup` tooltip "Data unavailable — this post may have been deleted on Reddit."
- [ ] All metric numbers formatted with locale-aware number formatting (e.g., 1,240 not 1240)
- [ ] All i18n strings use `$t()` keys; add to all locale files

**UI components:**
- `Badge` — for metric count badges
- `Alert` (info) — empty state
- `Loader` / skeleton — loading state
- `CstPopup` — tooltip on unavailable metrics

---

### Mock-ups:
Follow existing Analytics platform card pattern (e.g., Twitter/Bluesky analytics card layout). Reddit-specific metrics replace platform-specific equivalents (upvotes ≈ likes, comments ≈ comments).

---

### Impact on existing data:
No new data — consumes metrics already stored by the **[BE] Implement Reddit post analytics sync job** story.

---

### Impact on other products:
Analytics dashboard only. No impact on other platforms' analytics.

---

### Dependencies:
Depends on: **[BE] Implement Reddit post analytics sync job**

---

### Global quality & compliance:
- [ ] Mobile responsiveness — Analytics card must reflow on tablet and mobile viewports
- [ ] Multilingual support — All labels, empty state copy, and tooltips use `$t()` keys; add to all locale files
- [ ] UI theming support — All metric highlights use `text-primary-cs-500`; no hardcoded colors
- [ ] White-label domains impact review — Analytics is visible in white-label setups; theming must apply
- [ ] Cross-product impact assessment — Analytics page only; no impact on other features

---

---

## Story 11

### Title: [FE] Add Reddit post comments to the unified Inbox

### Description:
As a social media manager, I want to see comments left on my Reddit posts in ContentStudio's unified Inbox, so I can engage with my Reddit community and reply to discussions from the same place I manage all other platform conversations.

---

### Workflow:
1. User opens the Inbox in ContentStudio
2. Reddit comment items appear in the feed alongside Facebook, Instagram, and other platform conversations
3. User can filter by "Reddit" using the platform filter pill
4. User clicks a Reddit comment item — the conversation panel opens showing: the original post title (as thread header), the commenter's username, comment body, time
5. User types a reply in the reply box and clicks "Send" — ContentStudio posts it as a Reddit comment
6. The reply appears in the thread immediately (optimistic UI update)

---

### Acceptance criteria:
- [ ] Reddit platform filter pill visible in Inbox platform filter bar with Reddit icon
- [ ] Reddit inbox items display: Reddit icon badge, commenter username ("u/redditor_name"), comment body (truncated to 2 lines), subreddit ("r/marketing"), time elapsed
- [ ] Clicking a Reddit inbox item opens the conversation panel
- [ ] Conversation panel header: Reddit icon + subreddit name (e.g. "r/marketing") + original post title (linked, opens Reddit in new tab)
- [ ] Each comment shown with: user avatar (Reddit default avatar if none), "u/username", comment body (Markdown rendered), upvote score, time
- [ ] Reply box at bottom of conversation panel; placeholder: "Reply to this comment as u/[your-reddit-username]…"
- [ ] Send button: "Reply" — uses `Button` (primary variant)
- [ ] Optimistic UI: reply appears immediately in the thread with a loading indicator; confirmed/failed state updated when API responds
- [ ] Error state on reply failure: `Alert` (error variant) shown inline: "Couldn't send your reply. Reddit may have rate-limited your account. Please try again in a few minutes."
- [ ] "Direct messages are not available for Reddit" notice shown in the platform info area: `Alert` (info variant): "Reddit Inbox shows comments on posts you've published via ContentStudio. Direct messages are not supported via Reddit's API."
- [ ] Empty state (no Reddit inbox items): headline "No Reddit comments yet", subtext "Comments on Reddit posts you publish via ContentStudio will appear here.", CTA "View Published Posts" (links to Planner filtered by Reddit)
- [ ] Loading state: skeleton loaders for inbox item list while fetching
- [ ] Reddit inbox items included in unread count badge
- [ ] All i18n strings use `$t()` keys; add to all locale files

**UI components:**
- `Button` (primary) — Reply button
- `Alert` (info) — DM not supported notice
- `Alert` (error) — reply failure
- `Badge` — unread count
- `Loader` / skeleton — loading state

---

### Mock-ups:
Follow existing Inbox conversation panel layout. Reddit conversation panel uses the same structure as Facebook/Instagram comments panel.

---

### Impact on existing data:
No new data — consumes Reddit inbox items created by **[BE] Implement Reddit inbox strategy for post comments**. No impact on existing inbox items.

---

### Impact on other products:
Inbox module only. Unread count badge in the main navigation is updated to include Reddit comments.

---

### Dependencies:
Depends on: **[BE] Implement Reddit inbox strategy for post comments**

---

### Global quality & compliance:
- [ ] Mobile responsiveness — Inbox conversation panel must be responsive; test on 768px and 375px breakpoints
- [ ] Multilingual support — All labels, empty states, and notices use `$t()` keys; add to all locale files
- [ ] UI theming support — All primary color usage uses `text-primary-cs-500` etc.; no hardcoded colors
- [ ] White-label domains impact review — Inbox is available in white-label setups; Reddit icon must not use hardcoded brand color
- [ ] Cross-product impact assessment — Inbox platform filter, unread count, and notification system are shared; test that existing platform items are unaffected

---

---

## Story 12

### Title: [iOS] Add Reddit post comments support in iOS Inbox

### Description:
As a social media manager using the ContentStudio iOS app, I want to see and reply to comments on my Reddit posts in the iOS Inbox, so I can stay on top of Reddit community engagement on the go.

---

### Workflow:
1. User opens the Inbox in the ContentStudio iOS app
2. Reddit comment items appear in the inbox list alongside other platform conversations
3. User taps a Reddit comment → conversation detail screen opens showing the original post title and comment thread
4. User types a reply and taps "Send" → reply posted to Reddit

---

### Acceptance criteria:
- [ ] Reddit platform filter option visible in iOS Inbox filter controls with Reddit icon
- [ ] Reddit inbox list item shows: Reddit icon badge, "u/[username]", comment preview (2 lines), subreddit, time
- [ ] Tapping a Reddit inbox item opens the conversation detail screen
- [ ] Conversation detail header: subreddit name + original post title
- [ ] Comments displayed with username, body text (Markdown rendered as plain text on iOS), upvote score, timestamp
- [ ] Reply text field visible at bottom: placeholder "Reply as u/[your-reddit-username]…"
- [ ] Send button posts reply via the existing ContentStudio API (which calls `StrategyReddit::replyToConversation()`)
- [ ] Error toast shown if reply fails: "Couldn't send your reply. Please try again."
- [ ] DM notice shown (non-intrusive banner): "Reddit Inbox shows comments on your published posts. Direct messages are not supported."
- [ ] Empty state: "No Reddit comments yet — comments on posts you publish via ContentStudio will appear here."
- [ ] Reddit items counted in Inbox badge on the tab bar

**Platform detection:** Add `else if platform == "reddit"` conditional blocks in `PlatformPostViewController.swift` and `ConversationViewController.swift` following the existing facebook/instagram/linkedin pattern.

---

### Mock-ups:
Follow existing iOS Inbox conversation UI. Reddit-specific: subreddit name in header instead of page name; "u/username" instead of profile name.

---

### Impact on existing data:
No new data — reads from the same Inbox API that the web app uses.

---

### Impact on other products:
iOS Inbox only. No impact on Android or web.

---

### Dependencies:
Depends on: **[BE] Implement Reddit inbox strategy for post comments**

---

### Global quality & compliance:
- [ ] Mobile responsiveness — N/A, native iOS layout
- [ ] Multilingual support — All strings must use iOS i18n (`NSLocalizedString`); add to all locale `.strings` files
- [ ] UI theming support — Use iOS system tint color and existing ContentStudio theme tokens
- [ ] White-label domains impact review — iOS app may have white-label variants; ensure Reddit icon color doesn't break theming
- [ ] Cross-product impact assessed — iOS inbox change is isolated to the platform-specific conditional blocks

---

---

## Story 13

### Title: [Android] Add Reddit post comments support in Android Inbox

### Description:
As a social media manager using the ContentStudio Android app, I want to see and reply to comments on my Reddit posts in the Android Inbox, so I can manage Reddit community engagement from my phone.

---

### Workflow:
1. User opens the Inbox in the ContentStudio Android app
2. Reddit comment items appear in the inbox list with a Reddit icon badge
3. User taps a Reddit comment → opens the chat/conversation screen showing the original post and comments
4. User types a reply and taps "Send" → reply posted to Reddit

---

### Acceptance criteria:
- [ ] Reddit inbox items displayed in `InboxFragmentTabs.java` with Reddit icon badge, commenter username, comment preview, subreddit, time
- [ ] Tapping a Reddit item opens `ChatActivity.java` with the full comment thread
- [ ] Chat header shows: subreddit name + original post title
- [ ] Reply input at bottom with placeholder: "Reply as u/[your-reddit-username]…"
- [ ] Reply sent via `body.addProperty("platform", "reddit")` in `ChatActivity.java` — uses existing send mechanism
- [ ] Error snackbar on reply failure: "Couldn't send your reply. Please try again."
- [ ] DM notice shown (info card): "Reddit Inbox shows comments on your published posts. Direct messages are not supported."
- [ ] Empty state message: "No Reddit comments yet. Comments on posts you publish via ContentStudio will appear here."
- [ ] Reddit item count included in Inbox tab badge

**Platform detection:** Add `"reddit".equals(platform)` conditional blocks in `ChatActivity.java` and `InboxDetails.java` model class following the existing platform handling pattern.

---

### Mock-ups:
Follow existing Android Inbox chat UI. Reddit-specific: subreddit name in toolbar instead of page name; "u/username" label style.

---

### Impact on existing data:
No new data — reads from the same Inbox API endpoint used by web and iOS.

---

### Impact on other products:
Android Inbox only. No impact on iOS or web.

---

### Dependencies:
Depends on: **[BE] Implement Reddit inbox strategy for post comments**

---

### Global quality & compliance:
- [ ] Mobile responsiveness — N/A, native Android layout; test on multiple screen densities (mdpi, hdpi, xhdpi)
- [ ] Multilingual support — All strings in `strings.xml`; add Reddit-specific strings to all locale resource folders
- [ ] UI theming support — Use existing ContentStudio Android theme attributes; no hardcoded colors
- [ ] White-label domains impact review — Android app may support white-label; ensure Reddit icon follows theming
- [ ] Cross-product impact assessed — Android inbox change is isolated to platform-specific conditionals

---

---

## Story 14

### Title: [Design] Reddit integration — connection modal, composer section, inbox thread, and analytics card UI designs

### Description:
As a product team, we need finalized UI designs for the Reddit integration across all four touchpoints (connection modal, composer section, inbox thread, analytics card) so that frontend engineers have accurate specs to build from and the implementation is consistent with ContentStudio's design system.

---

### Workflow:
1. Designer reviews the workflow document (`docs/features/reddit-publishing/02-workflow.md`) and PRD (`docs/features/reddit-publishing/03-prd.md`) for all user flows and copy
2. Designer creates Figma designs for all four areas
3. Designs reviewed with Product and Engineering
4. Approved designs handed off to FE engineers working on Stories 7, 8, 10, and 11

---

### Acceptance criteria:
- [ ] **Connection modal designs:** Pre-OAuth info screen (what CS can/cannot do), success state (account card), and error state (OAuth denied / already connected) — all states designed
- [ ] **Composer Reddit section:** All three post type states (Text, Link, Image), subreddit autocomplete dropdown, flair dropdown with color swatches, mandatory flair "(Required)" state, multi-subreddit rows with stagger notice, all validation error states, all toggle states (Send Replies, OC)
- [ ] **Inbox thread view:** Reddit inbox list item, conversation panel with comment thread, reply box, DM not-supported notice, empty state
- [ ] **Analytics card:** Metric cards (posts, upvotes, comments, upvote ratio), top posts table, empty state, loading skeleton
- [ ] All designs use `@contentstudio/ui` design system components — no custom component shapes that don't exist in the library
- [ ] All colors use CSS variable-backed theme tokens — no hardcoded hex values including Reddit's #FF4500
- [ ] All states documented in Figma: default, hover, focus, error, loading, empty, success
- [ ] Reddit icon usage confirmed to be the `__brand__Reddit` icon from the existing icon set (or a gap flagged if missing)
- [ ] Copy in designs matches exactly the copy specified in the FE stories above — no placeholder "Lorem ipsum" or "TBD" copy
- [ ] Designs reviewed and approved by Product before frontend implementation begins

---

### Mock-ups:
N/A — this story produces the mock-ups.

---

### Impact on existing data:
N/A — design story only.

---

### Impact on other products:
Design system: if any new component patterns are introduced that could benefit the component library, flag to the design system team.

---

### Dependencies:
No technical dependencies — should be kicked off in parallel with BE stories so designs are ready when FE stories begin.

---

### Global quality & compliance:
- [ ] Mobile responsiveness — N/A for design story
- [ ] Multilingual support — Designs should account for text expansion (German, French expand ~30%); don't use fixed-width text containers
- [ ] UI theming support — All Figma components must use design tokens, not hardcoded colors
- [ ] White-label domains impact review — Designs must not assume ContentStudio's primary blue; use token colors
- [ ] Cross-product impact assessed — Reddit designs must be consistent with existing platform patterns (Telegram, Bluesky)

---

---

## Story 15 (v2)

### Title: [BE] Implement Reddit poll post type publishing

### Description:
As a social media manager, I want to create Reddit poll posts from ContentStudio so I can run community polls and gather opinions from subreddit audiences without leaving my scheduling tool.

**Note: This is a v2 story.** Reddit polls are not in the v1 launch scope. The story is written now for future sprint planning.

---

### Workflow:
1. User selects "Poll" as the post type in the Reddit composer section
2. User enters a poll question as the post title
3. User adds 2–6 answer options
4. User sets the voting duration (1 day, 3 days, or 7 days)
5. User selects a subreddit and optional flair
6. User schedules or publishes — ContentStudio creates the poll on Reddit

---

### Acceptance criteria:
- [ ] `RedditPosting` extended to handle poll posts: `kind: 'poll'` (or Reddit's equivalent poll submission endpoint)
- [ ] Poll submission uses Reddit's poll API: `POST /api/submit` with `kind: 'poll'` (verify current endpoint — Reddit's poll API may have changed since 2023)
- [ ] Poll fields stored in `reddit_sharing_details`: `{ poll_options: string[], poll_duration: 1|3|7 }`
- [ ] 2–6 poll options validated (minimum 2, maximum 6)
- [ ] Poll duration options: 1 day, 3 days, 7 days
- [ ] Poll post title follows same 300-char limit and required validation
- [ ] Flair supported for poll posts (same as other post types)
- [ ] Error handling for poll-specific API errors
- [ ] Poll post URL stored in posting_response after successful publish

---

### Mock-ups:
N/A — backend only

---

### Impact on existing data:
`reddit_sharing_details` extended with `poll_options` and `poll_duration` fields (additive).

---

### Impact on other products:
Enables **[FE] Add Reddit poll post type to the composer** story.

---

### Dependencies:
Depends on: **[BE] Implement Reddit post publishing (text, link, image) with multi-subreddit staggered scheduling**

---

### Global quality & compliance:
- [ ] Mobile responsiveness — N/A, backend-only
- [ ] Multilingual support — N/A, no user-facing strings
- [ ] UI theming support — N/A, backend-only
- [ ] White-label domains impact review — No impact
- [ ] Cross-product impact assessed — Extends existing Reddit publishing pipeline; no regression risk to other post types

---

---

## Story 16 (v2)

### Title: [FE] Add Reddit poll post type to the composer

### Description:
As a social media manager, I want to create Reddit poll posts from the ContentStudio composer, so I can run community polls to engage subreddit audiences and gather opinions — all from my scheduling workflow.

**Note: This is a v2 story.** Requires **[BE] Implement Reddit poll post type publishing** to be shipped first.

---

### Workflow:
1. User selects a Reddit account in the composer
2. User selects "Poll" in the post type `SegmentedControl` ("Text Post" | "Link Post" | "Image Post" | "Poll")
3. Poll-specific fields appear: Question (title field), Answer Options (2–6 inputs), Voting Duration
4. User enters the poll question as the title
5. User adds answer options (min 2, max 6) — a "+ Add Option" button adds more
6. User sets the voting duration using the `SegmentedControl`: "1 Day" | "3 Days" | "7 Days"
7. User selects subreddit and optional flair
8. User schedules or publishes

---

### Acceptance criteria:
- [ ] "Poll" segment added to the post type `SegmentedControl` (4th option)
- [ ] Poll section shows: Question field (same as Title field, 300 char limit), Answer Options section, Voting Duration selector
- [ ] Answer Options: minimum 2 fields shown by default, labeled "Option 1", "Option 2", etc.
- [ ] "+ Add Option" `Button` (ghost) adds a new option input up to a max of 6
- [ ] Options beyond 2 show an "×" remove button
- [ ] Option field validation: each option max 25 characters (Reddit poll option limit); error: "Poll options cannot exceed 25 characters."
- [ ] Minimum 2 options validation: "Polls require at least 2 answer options."
- [ ] Maximum 6 options: "+ Add Option" button hidden/disabled when 6 options exist; tooltip: "Reddit polls support a maximum of 6 answer options."
- [ ] Voting Duration `SegmentedControl`: "1 Day" | "3 Days" | "7 Days"; default "3 Days"
- [ ] ℹ tooltip on Voting Duration: "Choose how long the poll stays open for votes. After this time, Reddit closes the poll and shows the final results. Example: choosing '3 Days' means the poll closes 3 days after it's published."
- [ ] Subreddit and flair selectors shown below poll fields (same as other post types)
- [ ] Publish button disabled until: title/question filled, minimum 2 options filled, subreddit selected
- [ ] All i18n strings use `$t()` keys; add to all locale files

**UI components:**
- `SegmentedControl` — poll type selector + voting duration
- `TextInput` — question field, option fields
- `Button` (ghost) — "+ Add Option"
- `Switch` / `ActionIcon` — remove option button

---

### Mock-ups:
See Reddit's own poll creation UI for reference layout. Design story (Story 14) should include poll designs.

---

### Impact on existing data:
`reddit_sharing_details.poll_options` and `reddit_sharing_details.poll_duration` saved to Plans when poll type selected.

---

### Impact on other products:
Extends the Reddit composer section. No impact on other platforms.

---

### Dependencies:
Depends on: **[BE] Implement Reddit poll post type publishing**
Depends on: **[FE] Build Reddit composer section with post type selector, title field, subreddit search, flair dropdown, and multi-subreddit support**

---

### Global quality & compliance:
- [ ] Mobile responsiveness — Poll option inputs must stack correctly on small screens
- [ ] Multilingual support — All labels, validation messages, tooltips use `$t()` keys; add to all locale files
- [ ] UI theming support — All primary color classes use `text-primary-cs-500` etc.; no hardcoded colors
- [ ] White-label domains impact review — Poll UI is part of the composer; uses same theme tokens
- [ ] Cross-product impact assessed — Extends Reddit composer section only; no other platform sections affected

---

## Story Summary

| # | Title | Team | Priority | Project | Product Area |
|---|---|---|---|---|---|
| 1 | [BE] Add Reddit OAuth 2.0 account connection, token management, and platform configuration | Backend | High (P0) | Web App | Integrations |
| 2 | [BE] Implement Reddit post publishing (text, link, image) with multi-subreddit staggered scheduling | Backend | High (P0) | Web App | Publishing |
| 3 | [BE] Add Reddit subreddit search and flair fetching API endpoints | Backend | High (P0) | Web App | Publishing |
| 4 | [BE] Add Reddit to automation platform configuration (RSS, Evergreen, Bulk CSV) | Backend | Medium (P1) | Web App | Automation |
| 5 | [BE] Implement Reddit post analytics sync job | Backend | Medium (P1) | Web App | Analytics |
| 6 | [BE] Implement Reddit inbox strategy for post comments | Backend | Medium (P1) | Web App | Inbox |
| 7 | [FE] Add Reddit account connection modal and Social Accounts settings integration | Frontend | High (P0) | Web App | Integrations |
| 8 | [FE] Build Reddit composer section with post type selector, title field, subreddit search, flair dropdown, and multi-subreddit support | Frontend | High (P0) | Web App | Composer |
| 9 | [FE] Add Reddit to automation channel selectors, Evergreen automation, and onboarding platform list | Frontend | Medium (P1) | Web App | Automation |
| 10 | [FE] Add Reddit analytics to the Analytics dashboard | Frontend | Medium (P1) | Web App | Analytics |
| 11 | [FE] Add Reddit post comments to the unified Inbox | Frontend | Medium (P1) | Web App | Inbox |
| 12 | [iOS] Add Reddit post comments support in iOS Inbox | Frontend | Medium (P1) | Mobile | iOS Mobile |
| 13 | [Android] Add Reddit post comments support in Android Inbox | Frontend | Medium (P1) | Mobile | Android Mobile |
| 14 | [Design] Reddit integration — connection modal, composer section, inbox thread, and analytics card UI designs | Design | High (P0) | Web App | Throughout Product |
| 15 (v2) | [BE] Implement Reddit poll post type publishing | Backend | Low (P2) | Web App | Publishing |
| 16 (v2) | [FE] Add Reddit poll post type to the composer | Frontend | Low (P2) | Web App | Composer |
