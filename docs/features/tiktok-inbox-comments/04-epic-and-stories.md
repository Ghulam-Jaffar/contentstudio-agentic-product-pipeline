# Epic & Stories — TikTok Business Account Inbox

**Date:** 2026-04-30

---

## Epic

**Title:** TikTok Business Account Inbox — Comments, Connection & Auto-Reply

**Description:**

ContentStudio's social inbox currently covers Facebook, Instagram, LinkedIn, YouTube, and Google My Business, but not TikTok. This epic adds full TikTok Business Account support across three areas: account connection (standard OAuth and easy connect shareable link), comment management in the unified inbox (view, reply, hide/unhide, saved replies, video context), and an auto-reply rule engine that fires responses or moderation actions automatically based on configurable triggers.

The auto-reply capability is a meaningful differentiator — both Hootsuite and Sprout Social lock this behind their most expensive plans. ContentStudio will make it available at standard tiers, giving social media managers and agencies a cost-effective way to manage TikTok comment automation.

The backend rule engine is built as platform-agnostic (with a `platform` field) so it can be extended to other platforms in future epics. The v1 UI exposes TikTok only.

---

## Stories

---

### Story 1: [BE] Extend TikTok OAuth to support Business Account scopes for comment management

**Description:**
As a ContentStudio user with a TikTok Business Account, I want my account to be connected with the right permissions so that ContentStudio can read my video comments, post replies, and hide comments on my behalf — all without me needing to manage any credentials manually.

Currently, ContentStudio's TikTok OAuth only requests publishing scopes. This story extends the authorization request to include Business Account scopes required for comment management.

---

**Workflow:**
1. User initiates TikTok Business Account connection from Settings → Integrations → Social Accounts
2. ContentStudio redirects the user to TikTok's OAuth consent screen requesting Business API scopes: comment read, comment write, comment hide
3. User reviews and grants the requested permissions on TikTok
4. TikTok returns an authorization code to ContentStudio's callback URL
5. ContentStudio exchanges the code for an access token (valid ~2 hours) and a refresh token (valid 1 year)
6. Tokens are stored securely against the connected social account record
7. ContentStudio verifies the granted scopes match the required set; if any are missing, the user is shown a warning on the account settings page
8. Background token refresh is registered — access token is silently renewed using the refresh token before expiry

---

**Acceptance criteria:**
- [ ] TikTok OAuth authorization request includes Business API scopes for comment read, write, and hide
- [ ] Authorization callback correctly exchanges the code for access + refresh tokens
- [ ] Access token and refresh token are stored securely against the connected account
- [ ] If the user declines required permissions, the connection is not saved and the user sees: *"TikTok connection was cancelled. All permissions are required to manage comments."*
- [ ] If Business API scopes are not granted (partial grant), the account is flagged with a warning state visible in Social Accounts settings
- [ ] Background token refresh runs before access token expiry — users are not prompted to reconnect while their refresh token is valid
- [ ] When the refresh token expires or is revoked, the account is marked as requiring reconnection (consumed by Story 9 and Story 10)
- [ ] Existing TikTok publishing connections are not broken — accounts connected for publishing only retain their existing scopes

---

**Mock-ups:** N/A — backend only

**Impact on existing data:**
Existing TikTok social account records will not have comment management scopes. A migration or soft flag (`has_inbox_scope: false`) should distinguish pre-existing publishing-only connections from new Business Account connections with inbox scopes.

**Impact on other products:**
Chrome extension does not manage social account connections — no impact. Mobile apps have their own OAuth flow — mobile TikTok Business Account connection is out of scope for this story.

**Dependencies:** None

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

**Implementation references**
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-backend/app/Http/Controllers/Integrations/Platforms/Social/TiktokController.php` — existing OAuth callback (`connectLocal`, `callbackHooks`) lives here; extend to request Business API scopes
- `contentstudio-backend/app/Helpers/Integrations/TiktokHelper.php` — `validateTiktokAccessToken()` already implements refresh logic; extend to handle Business Account token refresh

**Existing pattern to follow:**
- Instagram OAuth in `InstagramController` — shows how Business Account scopes (pages, instagram_basic, etc.) are requested and stored

**Gotcha:**
- TikTok's standard Login Kit and the Business API use different OAuth authorization endpoints and different scopes. The current `TiktokController` uses Login Kit. Business Account comment management requires the Business API authorization endpoint — confirm the correct endpoint and required app registration with TikTok before implementation.

---

---

### Story 2: [FE] Add TikTok Business Account connection UI in Social Accounts settings

**Description:**
As a social media manager, I want to connect my TikTok Business Account from the Social Accounts settings page in ContentStudio so that my TikTok video comments start appearing in my unified inbox alongside Facebook, Instagram, and other platforms.

---

**Workflow:**
1. User navigates to Settings → Integrations → Social Accounts
2. User sees TikTok Business listed under available platforms (with a TikTok icon and "Connect" button)
3. User clicks "Connect TikTok Business"
4. A TikTok authorization window opens in a popup — user logs into TikTok and grants the requested permissions
5. The popup closes; ContentStudio shows a "Select Accounts" modal listing all TikTok Business accounts available under the authenticated user
6. User selects one or more accounts to add and clicks "Add to Inbox"
7. Selected accounts are added to the Social Accounts list with a TikTok badge and a "Connected" status
8. A success toast appears: *"TikTok Business account connected. Comments will start appearing in your inbox within a few minutes."*
9. If the user connects an account with incomplete permissions (missing comment scopes), a warning banner appears on the account row: *"Some TikTok permissions are missing. Reconnect to enable comment management."* with a "Reconnect" button

---

**Acceptance criteria:**
- [ ] TikTok Business appears as a connectable platform in Social Accounts settings
- [ ] Clicking "Connect TikTok Business" opens the TikTok OAuth flow in a popup window
- [ ] After authorization, a modal lists all available TikTok Business accounts for the user to select
- [ ] User can select multiple accounts and add them in a single action
- [ ] Success toast appears with copy: *"TikTok Business account connected. Comments will start appearing in your inbox within a few minutes."*
- [ ] Connected TikTok Business accounts appear in the Social Accounts list with TikTok icon and "Connected" status
- [ ] If permissions are incomplete, a warning banner shows on the account row with copy: *"Some TikTok permissions are missing. Reconnect to enable comment management."* and a "Reconnect" button
- [ ] If the user cancels the OAuth popup, no account is added and a toast reads: *"TikTok connection was cancelled."*
- [ ] Loading state shown on "Add to Inbox" button while accounts are being saved

**UI Copy:**

*Select Accounts modal:*
- **Title:** "Select TikTok Business Accounts"
- **Subtext:** "Choose which TikTok Business accounts you'd like to add to your inbox. You can add more accounts later from Settings."
- **CTA:** "Add to Inbox" / "Cancel"

*Account row warning banner:*
- "Some TikTok permissions are missing. Reconnect to enable comment management." + "Reconnect" button

*Success toast:*
- "TikTok Business account connected. Comments will start appearing in your inbox within a few minutes."

*Cancellation toast:*
- "TikTok connection was cancelled."

*Empty state (no TikTok Business accounts found under the authenticated user):*
- **Headline:** "No TikTok Business accounts found"
- **Subtext:** "Make sure you're logged into TikTok with an account that has Business Account access. If you manage multiple TikTok accounts, try switching accounts in TikTok and reconnecting."
- **CTA:** "Try Again"

---

**Mock-ups:** See PRD section 7

**Impact on existing data:** None — adds new account records

**Impact on other products:** None for this UI change

**Dependencies:** Depends on **[BE] Extend TikTok OAuth to support Business Account scopes for comment management**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

**Implementation references**
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-frontend/src/modules/integration/components/platforms/social_v2/ExternalCloudConnect.vue` — TikTok is already listed here (line 85); review what changes are needed to distinguish Business Account connection from existing TikTok publishing connection
- `contentstudio-frontend/src/modules/integration/components/platforms/social_v2/SocialConnectModal.vue` — modal for platform selection
- `contentstudio-frontend/src/modules/integration/components/platforms/social_v2/SaveSocialAccounts.vue` — dialog for saving newly connected accounts

**Components to use:**
- `Modal` from `@contentstudio/ui` for the Select Accounts modal
- `Checkbox` from `@contentstudio/ui` for multi-account selection
- `Button` from `@contentstudio/ui` for "Add to Inbox" and "Cancel" CTAs
- `Loader` from `@contentstudio/ui` for loading state on the Add button

---

---

### Story 3: [BE] Extend Easy Connect to generate TikTok Business Account authorization links

**Description:**
As a ContentStudio workspace, I want the Easy Connect system to support TikTok Business Account connections so that agency managers can share a link with clients, allowing clients to authorize their TikTok Business Account without ever sharing their TikTok login credentials.

---

**Workflow:**
1. Agency manager requests a TikTok Business Account Easy Connect link from the ContentStudio settings
2. The system generates a unique, expiring link tied to the workspace and the TikTok Business platform
3. The manager shares the link with the account owner (client)
4. When the account owner opens the link, the system validates it is still active (not expired, not revoked)
5. The system redirects the account owner through the TikTok Business OAuth flow with the correct scopes
6. After the account owner grants permissions, the tokens are stored against the workspace
7. The manager is notified (in-app) that the account has been connected
8. The link is marked as used and cannot be used again

---

**Acceptance criteria:**
- [ ] Easy Connect link generation supports TikTok Business as a platform option
- [ ] Generated links are unique per request and expire after 7 days
- [ ] Links are single-use — once an account is connected via the link, the link cannot be used again
- [ ] When an account owner opens the link, the system validates it (not expired, not revoked) before redirecting to TikTok OAuth
- [ ] If the link is expired, the account owner sees an error page: *"This connection link has expired. Ask your workspace manager to generate a new one."*
- [ ] After the account owner successfully authorizes, the connected account appears in the workspace's Social Accounts list
- [ ] After the account owner successfully authorizes, the manager receives an in-app notification: *"[TikTok account name] has been connected to your workspace via Easy Connect."*
- [ ] If the account owner cancels OAuth, the link remains valid for another attempt until it expires
- [ ] If the TikTok account is already connected to the workspace, the system returns a clear error: *"This TikTok account is already connected to your workspace."*

---

**Mock-ups:** N/A — backend only

**Impact on existing data:**
Adds TikTok Business as a supported platform to the Easy Connect link records. Existing Easy Connect links for other platforms are unaffected.

**Impact on other products:** None

**Dependencies:** Depends on **[BE] Extend TikTok OAuth to support Business Account scopes for comment management**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

**Implementation references**
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-backend/app/Http/Controllers/Integrations/ExternalLinkIntegrationController.php` — existing Easy Connect controller; routes: `generateExternalIntegrationLink`, `getLinkDetails`, etc.

**Existing pattern to follow:**
- Existing Easy Connect links for other platforms — add TikTok Business as a new `platform` value in the link generation logic

---

---

### Story 4: [FE] Add TikTok Business Account to the Easy Connect shareable link flow

**Description:**
As an agency manager, I want to generate a shareable TikTok Business Account connection link from ContentStudio's Easy Connect settings so that I can send it to a client and have them authorize their account — without the client needing to log into ContentStudio or share any passwords.

---

**Workflow:**
1. User navigates to Settings → Integrations → Easy Connect
2. User sees the list of platforms available for Easy Connect — TikTok Business is now included
3. User selects TikTok Business and clicks "Generate Link"
4. ContentStudio generates a unique link and displays it with a "Copy Link" button and an expiry notice
5. User copies the link and shares it with the account owner (via email, Slack, etc.)
6. Account owner opens the link in a browser — sees a ContentStudio-branded authorization page explaining what permissions are being requested
7. Account owner clicks "Authorize with TikTok" and goes through TikTok OAuth
8. After authorization, the account owner sees a confirmation: *"Your TikTok Business account has been successfully connected to [Workspace Name]."*
9. The manager's ContentStudio receives an in-app notification that the account was connected
10. The connected account appears in the workspace Social Accounts list

---

**Acceptance criteria:**
- [ ] TikTok Business appears in the Easy Connect platform list
- [ ] Clicking "Generate Link" produces a unique shareable URL
- [ ] The generated link is displayed with a "Copy Link" button and expiry notice: *"This link expires in 7 days. Share it only with the account owner."*
- [ ] The account owner authorization page shows ContentStudio branding, the workspace name, and a plain-language explanation of what access is being granted
- [ ] "Authorize with TikTok" button on the authorization page initiates TikTok OAuth
- [ ] After successful authorization, account owner sees: *"Your TikTok Business account has been successfully connected to [Workspace Name]. You can now close this page."*
- [ ] Manager receives in-app notification: *"[TikTok account name] has been connected to your workspace via Easy Connect."*
- [ ] If the link is expired when the account owner opens it, the page shows: *"This connection link has expired. Ask your workspace manager to generate a new one."*
- [ ] Manager can generate a new link at any time (previous link becomes invalid)
- [ ] Manager can revoke an active link before it is used

**UI Copy:**

*Easy Connect platform list — TikTok Business entry:*
- **Platform name:** "TikTok Business"
- **Subtext:** "Let someone connect their TikTok Business account to your workspace without sharing their login details."

*Generated link panel:*
- **Headline:** "Your TikTok Business connection link is ready"
- **Subtext:** "Share this link with the TikTok account owner. They'll go through a quick authorization on TikTok's website. This link expires in 7 days — generate a new one if it expires."
- **CTA:** "Copy Link"
- **Secondary action:** "Revoke Link"

*Account owner authorization page:*
- **Headline:** "[Workspace Name] is requesting access to your TikTok Business account"
- **Subtext:** "This will allow [Workspace Name] to view comments on your TikTok videos, reply to comments, and hide comments — directly from ContentStudio. They will not be able to post videos, access your account settings, or see your personal information."
- **CTA:** "Authorize with TikTok"
- **Secondary:** "Cancel"

*Account owner success page:*
- **Headline:** "You're all set!"
- **Subtext:** "Your TikTok Business account has been successfully connected to [Workspace Name]. You can close this page."

*Expired link page:*
- **Headline:** "This link has expired"
- **Subtext:** "This connection link is no longer valid. Ask your workspace manager to generate a new one and share it with you."

---

**Mock-ups:** See PRD section 7

**Impact on existing data:** None

**Impact on other products:** None

**Dependencies:**
- Depends on **[BE] Extend Easy Connect to generate TikTok Business Account authorization links**
- Depends on **[BE] Extend TikTok OAuth to support Business Account scopes for comment management**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

**Implementation references**
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-frontend/src/modules/integration/components/platforms/social_v2/ExternalCloudConnect.vue` — existing Easy Connect UI; TikTok is already referenced here (line 85); extend to add TikTok Business as a selectable option

**Components to use:**
- `Button` from `@contentstudio/ui` for "Generate Link", "Copy Link", "Revoke Link"
- `Badge` from `@contentstudio/ui` for expiry status indicator

---

---

### Story 5: [BE] Build TikTok comment ingestion worker and inbox queue integration

**Description:**
As a ContentStudio user with a connected TikTok Business Account, I want my TikTok video comments to automatically appear in my inbox so that I can monitor and respond to them alongside comments and messages from all my other social platforms — without any manual refresh or separate app switching.

---

**Workflow:**
1. ContentStudio's inbox queue runs on a schedule and processes TikTok accounts alongside other platforms
2. For each connected TikTok Business Account, the system fetches recent videos and retrieves comments on each video
3. New comments are stored as inbox items (type: `post`) with commenter details, comment text, timestamp, and the associated video's thumbnail URL and title
4. Nested replies to comments are stored with a parent comment reference so threads are preserved
5. The inbox is updated in real time via Pusher — new TikTok comments appear in the user's inbox without a page refresh
6. TikTok API rate limits are respected — the system uses batching and exponential backoff if limits are hit
7. If a TikTok token cannot be refreshed during ingestion, the account is flagged as needing reconnection and the queue skips it gracefully (no crash, no silent data loss)

---

**Acceptance criteria:**
- [ ] TikTok Business Account comments are fetched and stored as inbox items in the existing `InboxDetails` store
- [ ] Each inbox item includes: commenter username, commenter avatar URL, comment text, timestamp, video ID, video thumbnail URL, and video title
- [ ] Nested comment replies are stored with a reference to their parent comment ID so threading can be displayed
- [ ] New comments appear in the inbox in real time (via Pusher broadcast) without requiring a page refresh
- [ ] The inbox queue processes TikTok accounts without breaking or degrading processing for other platforms
- [ ] If TikTok API rate limits are hit, the queue backs off and retries — the job does not fail or crash
- [ ] If a TikTok account's token cannot be refreshed during ingestion, that account is skipped gracefully and flagged for reconnection; other accounts in the queue are not affected
- [ ] Comment ingestion is incremental — only comments newer than the last-fetched timestamp are pulled on subsequent runs (not a full re-fetch every time)

---

**Mock-ups:** N/A — backend only

**Impact on existing data:**
New `InboxDetails` records created for TikTok comments. Existing records for other platforms are unchanged. Schema additions: `video_id`, `video_thumbnail_url`, `video_title`, `parent_comment_id` fields in the `element_details` embedded object.

**Impact on other products:**
Chrome extension does not access the inbox — no impact.

**Dependencies:**
- Depends on **[BE] Extend TikTok OAuth to support Business Account scopes for comment management**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

**Implementation references**
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-backend/app/Libraries/Inbox/HelperClasses/InstagramHelper.php` — closest existing pattern; shows `processComment()`, token validation, and `InboxDetails` storage
- `contentstudio-backend/app/Jobs/InboxQueueMasterJob.php` — add TikTok to the `get_posts` queue processing loop
- `contentstudio-backend/app/Models/Inbox/InboxDetails.php` — existing model; platform-agnostic
- `contentstudio-backend/app/Repository/Inbox/InboxDetailsRepository.php` — platform-agnostic storage
- `contentstudio-backend/app/Libraries/Inbox/InboxPusherBroadcast.php` — real-time broadcast pattern

**TikTok API:**
- Comment fetch: `GET /v1/video/comments/search/` — requires `video_id`; must iterate per video
- Comment ingestion must first fetch recent videos, then fetch comments per video

**Gotcha:**
- TikTok's comment API is per-video (not account-level). For accounts with many videos, fetching comments on all videos every cycle is too expensive. Limit to videos published in the last 30 days (or configurable) and only fetch comments since the last-seen timestamp.

---

---

### Story 6: [BE] Implement TikTok comment reply and hide/unhide API endpoints

**Description:**
As a social media manager, I want ContentStudio to send my replies to TikTok comments and execute hide/unhide actions on my behalf so that I can engage with and moderate my TikTok comment section entirely from the ContentStudio inbox without ever needing to open TikTok.

---

**Workflow:**
1. User submits a reply from the ContentStudio inbox for a TikTok comment
2. ContentStudio validates the reply is within TikTok's 150-character limit
3. ContentStudio sends the reply to TikTok's comment reply API using the connected account's access token
4. On success, the reply is stored in the inbox thread and the user sees it appear immediately
5. On failure, the user is notified and can retry

6. User clicks hide on a TikTok comment
7. ContentStudio calls TikTok's hide comment API — the comment is hidden from public view on TikTok
8. The inbox item is updated to reflect the hidden state

9. User clicks unhide on a previously hidden TikTok comment
10. ContentStudio calls TikTok's unhide API — the comment becomes visible on TikTok again
11. The inbox item's hidden state is cleared

---

**Acceptance criteria:**
- [ ] `POST /inbox/tiktok/reply` (or equivalent route) accepts a comment ID and reply text, validates length ≤ 150 characters, and posts the reply to TikTok
- [ ] Reply success: reply stored in inbox thread, returned to frontend for immediate display
- [ ] Reply failure (TikTok API error): error returned to frontend with a user-facing message
- [ ] Reply failure (comment no longer exists on TikTok): specific error returned with message: *"This comment no longer exists on TikTok."*
- [ ] `POST /inbox/tiktok/hide` accepts a comment ID and hides the comment on TikTok; updates the inbox item's hidden status
- [ ] `POST /inbox/tiktok/unhide` accepts a comment ID and unhides the comment on TikTok; clears the inbox item's hidden status
- [ ] All three endpoints use the connected account's access token (with background refresh)
- [ ] If the account token is expired and cannot be refreshed, the endpoint returns an appropriate error prompting reconnection
- [ ] Replies are posted as the connected TikTok Business Account (not as ContentStudio)

---

**Mock-ups:** N/A — backend only

**Impact on existing data:**
Inbox items gain a `is_hidden` boolean field to track hide/unhide state for TikTok comments.

**Impact on other products:** None

**Dependencies:**
- Depends on **[BE] Build TikTok comment ingestion worker and inbox queue integration**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

**Implementation references**
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-backend/app/Libraries/Inbox/HelperClasses/InstagramHelper.php` — `sendMessageToInstagram()` is the pattern to follow for `sendCommentReplyToTiktok()`
- `contentstudio-backend/app/Http/Controllers/Conversation/InboxSavedReplyController.php` — saved reply pattern

**TikTok API:**
- Reply to comment: `POST /v1/video/comment/reply/create/`
- Hide comment: `POST /v1/video/comment/hide/` (toggle)

---

---

### Story 7: [FE] Display TikTok comments in the unified inbox with video context

**Description:**
As a social media manager, I want to see my TikTok video comments appear in the ContentStudio inbox alongside comments and messages from all my other connected platforms, with the video they were posted on clearly shown, so that I have full context before I reply — without needing to switch to the TikTok app.

---

**Workflow:**
1. User opens Inbox from the left sidebar navigation
2. TikTok comments appear in the inbox list alongside Facebook, Instagram, and other platform items
3. Each TikTok comment card in the list shows:
   - TikTok platform icon/badge
   - Commenter's username and avatar
   - Comment text preview (truncated if long)
   - Time since the comment was posted
   - Video thumbnail (small) and video title
4. User can filter the inbox to show only TikTok comments using the channel/platform filter — TikTok appears as a filter option
5. User clicks a TikTok comment to open the conversation thread
6. The thread view opens and shows at the top:
   - The original video thumbnail (larger), title, and a "View on TikTok" link
   - The commenter's username and avatar
   - The original comment text
   - Any existing replies in the thread (with indentation for nested replies)
7. A "Hidden" badge is visible on any comment that has been hidden from TikTok
8. New TikTok comments appear in real time (via Pusher) without requiring a page refresh

---

**Acceptance criteria:**
- [ ] TikTok comments appear in the inbox list with: TikTok badge, commenter username + avatar, comment text preview, timestamp, video thumbnail, and video title
- [ ] TikTok is available as an option in the inbox channel/platform filter
- [ ] Filtering by TikTok shows only TikTok inbox items; all other platform items are hidden
- [ ] Clicking a TikTok comment opens the thread view with full video context at the top (thumbnail, title, "View on TikTok" link)
- [ ] Nested replies in the thread are displayed with visual indentation to show the reply hierarchy
- [ ] Hidden comments show a "Hidden" badge in both the inbox list and thread view
- [ ] New TikTok comments appear in the inbox in real time without a page refresh
- [ ] Loading state (skeleton) shown while inbox items are loading
- [ ] Empty state shown when no TikTok comments exist for the filtered view

**Empty state (TikTok filter active, no comments yet):**
- **Headline:** "No TikTok comments yet"
- **Subtext:** "Comments on your TikTok videos will appear here. New comments usually show up within a few minutes of being posted."
- No CTA needed (passive state)

**Empty state (inbox is empty, no TikTok account connected):**
- This is handled by the existing inbox empty state — no TikTok-specific variant needed here

**Loading state:** Use existing inbox skeleton loader pattern; ensure TikTok items use the same skeleton shape as other platform items

**Error state (failed to load TikTok comments):**
- Inline error within the inbox list: *"Couldn't load TikTok comments. Check your connection and try refreshing."* with a "Refresh" link

---

**Mock-ups:** See PRD section 7 and [Design] TikTok inbox UI design story

**Impact on existing data:** None — reads from existing InboxDetails records

**Impact on other products:** None

**Dependencies:**
- Depends on **[BE] Build TikTok comment ingestion worker and inbox queue integration**
- Depends on **[Design] Design TikTok inbox comment cards, video context panel, and auto-reply rule builder**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

**Implementation references**
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-frontend/src/modules/inbox-revamp/InboxView.vue` — main inbox page
- `contentstudio-frontend/src/modules/inbox-revamp/components/InboxListing.vue` — inbox item list + filters; add TikTok to channel filter options
- `contentstudio-frontend/src/modules/inbox-revamp/components/ConversationView.vue` — thread view; add video context panel at top for TikTok items
- `contentstudio-frontend/src/modules/inbox-revamp/components/PlatformPill.vue` — platform badge; add TikTok variant
- `contentstudio-frontend/src/modules/inbox-revamp/store/inbox-revamp.js` — Vuex store; ensure TikTok platform is not excluded from any existing platform filters

**Components to use:**
- `Avatar` from `@contentstudio/ui` for commenter avatars
- `Badge` from `@contentstudio/ui` for "Hidden" state badge and TikTok platform pill
- `Loader` from `@contentstudio/ui` for loading states

---

---

### Story 8: [FE] Add TikTok comment reply composer with 150-character limit, hide/unhide actions, and saved replies

**Description:**
As a social media manager, I want to reply to TikTok comments directly from the ContentStudio inbox — using saved reply templates for speed and with a clear character counter so I don't exceed TikTok's 150-character limit — and I want to be able to hide or unhide a comment with a single click, all without leaving ContentStudio.

---

**Workflow:**
1. User opens a TikTok comment thread in the inbox
2. The reply composer is shown at the bottom of the thread view
3. User types a reply — a live character counter shows remaining characters (e.g., "120/150")
4. As the user approaches the limit, the counter turns to an amber warning color at 20 remaining; at 0 it turns red and the Send button is disabled
5. User can click "Saved Replies" to browse and insert a pre-written template; the template populates the composer (user can edit before sending)
6. User clicks "Send" — reply is submitted; a loading indicator appears on the button; reply appears in the thread on success
7. If the send fails, an inline error appears: *"Your reply couldn't be sent. Check your TikTok account connection and try again."* with a "Retry" button
8. To hide a comment, user clicks the "⋮" (more actions) menu on any comment in the thread and selects "Hide Comment"
9. An optimistic update hides the comment in the UI immediately; a "Hidden" badge appears
10. To unhide, user clicks "⋮" → "Unhide Comment"; the badge is removed

---

**Acceptance criteria:**
- [ ] Reply composer is shown for TikTok comment threads with a character counter (format: "[used]/150")
- [ ] Character counter turns amber when ≤ 20 characters remain; turns red at 0
- [ ] Send button is disabled when the reply is empty or exceeds 150 characters
- [ ] "Saved Replies" button opens the saved replies panel; selecting a reply populates the composer
- [ ] Clicking "Send" submits the reply and shows a loading state on the button; reply appears in the thread on success
- [ ] On reply failure: inline error shown with copy *"Your reply couldn't be sent. Check your TikTok account connection and try again."* and "Retry" button
- [ ] On reply failure (comment deleted): error shown with copy *"This comment no longer exists on TikTok."*
- [ ] "⋮" menu on each comment in the thread includes "Hide Comment" option
- [ ] Clicking "Hide Comment" optimistically hides the comment in the UI and shows a "Hidden" badge; reverts if the API call fails with a toast: *"Couldn't hide this comment. Try again."*
- [ ] "⋮" menu on a hidden comment shows "Unhide Comment"; clicking it restores the comment (optimistic update)

**UI Copy:**

*Reply composer placeholder:*
- "Reply to this comment… (max 150 characters)"

*Character counter:*
- Normal: "[used]/150" in `text-gray-400`
- Warning (≤20 remaining): "[used]/150" in `text-yellow-500`
- Exceeded: "[used]/150" in `text-red-500`

*Send button:* "Send Reply"

*Saved Replies button tooltip:*
- "Use a saved reply to respond faster. You can edit it before sending."

*Reply failure toast:*
- "Your reply couldn't be sent. Check your TikTok account connection and try again." + "Retry"

*Comment deleted error:*
- "This comment no longer exists on TikTok."

*Hide success (implicit — "Hidden" badge appears):* No separate toast needed; badge is sufficient feedback

*Hide failure toast:*
- "Couldn't hide this comment. Try again."

*Unhide failure toast:*
- "Couldn't unhide this comment. Try again."

---

**Mock-ups:** See [Design] TikTok inbox UI design story

**Impact on existing data:** None

**Impact on other products:** None

**Dependencies:**
- Depends on **[BE] Implement TikTok comment reply and hide/unhide API endpoints**
- Depends on **[FE] Display TikTok comments in the unified inbox with video context**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

**Implementation references**
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-frontend/src/modules/inbox-revamp/components/MessageComposer.vue` — existing reply composer; currently handles character limits for other platforms; add TikTok-specific 150-char limit when the active inbox item is a TikTok comment
- `contentstudio-frontend/src/modules/inbox-revamp/components/SavedReplyListing.vue` — existing saved replies panel; no TikTok-specific changes needed
- `contentstudio-frontend/src/modules/inbox-revamp/store/inbox-revamp.js` — add TikTok reply and hide/unhide actions

**Components to use:**
- `Textarea` from `@contentstudio/ui` for the reply text field
- `Button` from `@contentstudio/ui` for Send Reply, Saved Replies, Retry
- `Dropdown` + `DropdownItem` from `@contentstudio/ui` for the ⋮ more actions menu
- `Badge` from `@contentstudio/ui` for "Hidden" state badge

---

---

### Story 9: [BE] Implement TikTok token refresh and reconnect notification system

**Description:**
As a ContentStudio user with a connected TikTok Business Account, I want my inbox and auto-reply to keep working without interruption — and when my TikTok connection does need attention, I want to know about it immediately rather than discovering it hours later when comments have gone unanswered.

---

**Workflow:**
1. A background job runs on a regular schedule to check TikTok access token validity
2. If an access token is within expiry threshold (e.g., within 30 minutes of expiring), the job uses the refresh token to get a new access token silently — the user is not notified
3. If the refresh token itself has expired or been revoked, the background job cannot renew the access token
4. In this case, the TikTok account is flagged as `requires_reconnection = true` in the social account record
5. An in-app notification is created for the workspace: *"Your TikTok Business account [account name] needs to be reconnected. Inbox and auto-reply are paused for this account."*
6. Any auto-reply rules tied to this account are automatically paused
7. When the user reconnects the account (via Settings → Integrations → Social Accounts), the flag is cleared and rules resume

---

**Acceptance criteria:**
- [ ] Access tokens are silently refreshed before expiry — users are not interrupted during normal operation
- [ ] If refresh fails (refresh token expired or revoked), the account is marked `requires_reconnection = true`
- [ ] In-app notification is created when an account is flagged: *"Your TikTok Business account [account name] needs to be reconnected. Inbox and auto-reply are paused."*
- [ ] All auto-reply rules associated with the flagged account are automatically paused (not deleted)
- [ ] Comment ingestion for the flagged account is skipped gracefully — other accounts continue processing normally
- [ ] When the account is successfully reconnected, `requires_reconnection` is cleared and auto-reply rules resume automatically
- [ ] The reconnection notification includes a direct link to the Social Accounts settings page

---

**Mock-ups:** N/A — backend only

**Impact on existing data:**
Social account records gain a `requires_reconnection` boolean field. Auto-reply rule records gain an `is_paused` boolean (can be paused both manually by user and automatically by this system).

**Impact on other products:** None

**Dependencies:**
- Depends on **[BE] Extend TikTok OAuth to support Business Account scopes for comment management**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

**Implementation references**
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-backend/app/Helpers/Integrations/TiktokHelper.php` — `validateTiktokAccessToken()` already implements token refresh; extend to handle hard failure and set the reconnect flag
- `contentstudio-backend/app/Models/Inbox/InboxCronMaster.php` — existing model tracks failed ingestion attempts (`has_failed_today`, `failed_type`); extend or use as pattern for reconnect tracking

**Existing pattern to follow:**
- `contentstudio-frontend/src/modules/inbox-revamp/composables/useInstagramReconnectBanner.js` — Instagram reconnect banner composable; the backend state that drives this is the reference pattern

---

---

### Story 10: [FE] Show TikTok reconnect banner in inbox when token expires

**Description:**
As a ContentStudio user, I want to see a clear, actionable reconnect notice inside the inbox when my TikTok Business Account connection has expired, so that I can fix it in seconds and resume receiving comments and auto-replies — rather than discovering hours later that my inbox went silent.

---

**Workflow:**
1. ContentStudio detects that a connected TikTok account has `requires_reconnection = true` (set by the backend token refresh system)
2. A yellow warning banner appears at the top of the inbox (above the conversation list) when viewing TikTok items or when TikTok is the only/primary connected platform
3. The banner reads: *"Your TikTok Business account [account name] needs to be reconnected. Comments and auto-replies are paused."* with a "Reconnect" button
4. User clicks "Reconnect" — taken directly to Settings → Integrations → Social Accounts, scrolled to the affected TikTok account
5. After successful reconnection, the banner disappears automatically

---

**Acceptance criteria:**
- [ ] A warning banner appears in the inbox when any connected TikTok account has `requires_reconnection = true`
- [ ] Banner copy: *"Your TikTok Business account [account name] needs to be reconnected. Comments and auto-replies are paused until you reconnect."*
- [ ] "Reconnect" button in the banner navigates the user to Settings → Integrations → Social Accounts (deep-linked to the affected account if possible)
- [ ] If multiple TikTok accounts need reconnection, the banner lists all affected accounts (or a summarized version for 3+: *"3 TikTok accounts need reconnecting"*)
- [ ] Banner disappears automatically after the account is successfully reconnected — no page refresh required
- [ ] Banner is dismissed-able (close button) — dismissing it does not fix the issue; the banner reappears on next page load until the account is reconnected

**UI Copy:**

*Single account reconnect banner:*
- **Icon:** Warning icon (`⚠`)
- **Text:** "Your TikTok Business account **[account name]** needs to be reconnected. Comments and auto-replies are paused until you reconnect."
- **CTA:** "Reconnect Now"
- **Dismiss:** "×" (close icon; banner returns on refresh)

*Multiple accounts reconnect banner:*
- "**[N] TikTok Business accounts** need to be reconnected. Comments and auto-replies are paused for these accounts."
- **CTA:** "Go to Settings"

---

**Mock-ups:** See [Design] TikTok inbox UI design story

**Impact on existing data:** None — reads `requires_reconnection` flag from social account records

**Impact on other products:** None

**Dependencies:**
- Depends on **[BE] Implement TikTok token refresh and reconnect notification system**
- Depends on **[FE] Display TikTok comments in the unified inbox with video context**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

**Implementation references**
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-frontend/src/modules/inbox-revamp/composables/useInstagramReconnectBanner.js` — direct pattern to follow; create a `useTiktokReconnectBanner.js` composable following the same structure
- `contentstudio-frontend/src/modules/inbox-revamp/InboxView.vue` — where the banner should be rendered

**Components to use:**
- `Badge` from `@contentstudio/ui` for warning state indicator
- `Button` from `@contentstudio/ui` for "Reconnect Now" CTA

---

---

### Story 11: [BE] Build auto-reply rule engine for TikTok comments

**Description:**
As a ContentStudio user, I want the system to automatically reply to or hide TikTok comments based on rules I configure, so that common questions are answered instantly and spam is moderated even when I'm not online — without me needing to monitor my inbox around the clock.

---

**Workflow:**
1. User creates an auto-reply rule via the UI (covered in the FE story); rule is stored in the backend
2. When a new TikTok comment is ingested (via webhook or polling), the system evaluates all active rules for the connected account
3. Rules are evaluated in priority order (user-defined ordering); the first matching rule wins — only one rule fires per comment
4. Rule trigger conditions evaluated:
   - **All new comments:** matches every incoming comment
   - **Comment contains keyword(s):** case-insensitive substring match; multiple keywords are OR logic (any match triggers the rule)
   - **First comment from a user:** matches if this is the commenter's first comment on any video for this connected account
5. For rules with action "Reply with template":
   - If multiple response variants are defined, one is selected at random
   - The reply is sent to TikTok via the comment reply API
   - A cooldown is enforced: the same auto-reply rule does not fire more than once per commenter per 24-hour window (to prevent spam)
6. For rules with action "Hide": the comment is hidden on TikTok
7. For rules with action "Reply + Hide": both actions are executed
8. A log entry is created for each auto-reply action (rule ID, comment ID, account, action taken, timestamp)
9. If the auto-reply API call fails, the failure is logged and the inbox item is flagged with `auto_reply_failed = true`

---

**Acceptance criteria:**
- [ ] `InboxAutoReplyRule` model stores: name, platform (`tiktok`), account IDs (can be multiple or all-in-workspace), trigger type, keywords (array), action type, response variants (array), priority order, `is_active` boolean
- [ ] Rules are evaluated in `priority` order for each incoming TikTok comment
- [ ] Only the first matching rule fires — subsequent matching rules are skipped
- [ ] Trigger: "All new comments" matches every comment
- [ ] Trigger: "Contains keyword" performs case-insensitive substring match; keywords are OR logic
- [ ] Trigger: "First comment from user" fires only if the commenter has no prior comments on the account in the system
- [ ] Cooldown enforced: the same rule does not fire more than once per `(commenter_id, rule_id)` pair per 24-hour window
- [ ] Response variant selection is random when multiple variants exist
- [ ] Reply action calls the TikTok comment reply API (same endpoint as Story 6)
- [ ] Hide action calls the TikTok hide comment API (same endpoint as Story 6)
- [ ] An auto-reply log entry is created for every fired action: rule ID, comment ID, account, action, timestamp, success/failure
- [ ] On API failure, inbox item is flagged `auto_reply_failed = true`; failure is logged
- [ ] If the account's TikTok token is invalid at the time the rule fires, the rule is skipped and failure is logged; account is flagged for reconnection (Story 9 handles the notification)
- [ ] Rules for a paused or disconnected account do not fire
- [ ] Platform field is stored as `tiktok` on all rule records — architecture supports extending to other platforms without schema changes

---

**Mock-ups:** N/A — backend only

**Impact on existing data:**
Two new models introduced:
- `InboxAutoReplyRule` — stores rule configuration (platform, trigger, action, variants, priority, active state)
- `InboxAutoReplyLog` — stores a record of every auto-reply action fired

**Impact on other products:**
Auto-reply rule engine is built platform-agnostic; v1 only exposes TikTok rules via the UI. No impact on other products.

**Dependencies:**
- Depends on **[BE] Implement TikTok comment reply and hide/unhide API endpoints**
- Depends on **[BE] Build TikTok comment ingestion worker and inbox queue integration**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

---

### Story 12: [FE] Build auto-reply rule builder and management UI for TikTok comments

**Description:**
As a social media manager, I want to create, manage, and toggle auto-reply rules for my TikTok comments from a dedicated Automation tab in the Inbox, so that I can set up smart, time-saving automations for my TikTok account — and adjust them at any time — without any technical knowledge.

---

**Workflow:**
1. User navigates to Inbox → Automation tab (new tab added to inbox navigation)
2. The Automation tab shows the list of existing TikTok auto-reply rules (or an empty state if none exist)
3. User clicks "Create Rule"
4. A rule builder modal opens with the following fields:
   - Rule name (text input)
   - Applies to (select which TikTok account(s) this rule covers — or "All TikTok accounts")
   - Trigger (segmented control or radio: "All new comments" / "Comment contains keyword(s)" / "First comment from a user")
   - Keywords (text input with tag-style entry — visible only when "Contains keyword" trigger is selected)
   - Action (radio: "Reply with template" / "Hide comment" / "Reply and hide comment")
   - Response templates (one or more text areas — visible when action includes "Reply"; user can add up to 5 variants via "+ Add another response")
5. User saves the rule
6. The rule appears in the list with its name, trigger summary, and an on/off toggle
7. Rules can be reordered by dragging (priority is determined by list order — top = highest priority)
8. User can edit or delete any rule from the list
9. When a rule is paused due to account reconnection needed, it shows a "Paused" badge with a "Reconnect account" link

---

**Acceptance criteria:**
- [ ] An "Automation" tab appears in the Inbox navigation area (alongside existing inbox filter tabs/navigation)
- [ ] Automation tab shows the list of TikTok auto-reply rules with: rule name, trigger summary, action summary, on/off toggle, edit button, delete button
- [ ] "Create Rule" button opens the rule builder modal
- [ ] Rule builder modal has all required fields: rule name, applies-to selector, trigger selector, keywords input (conditional), action selector, response template input(s)
- [ ] Keywords input supports tag-style entry (user types a keyword, presses Enter to add it as a tag; can remove tags with ×)
- [ ] User can add up to 5 response variants via "+ Add another response"; each variant is a separate text area
- [ ] Saving a rule with an empty rule name shows validation: *"Please give your rule a name."*
- [ ] Saving a "Contains keyword" rule with no keywords shows validation: *"Please add at least one keyword."*
- [ ] Saving a "Reply" action rule with no response text shows validation: *"Please add at least one reply template."*
- [ ] Rules can be toggled on/off inline without opening the editor
- [ ] Rules can be reordered by dragging — the order determines priority (top = fires first)
- [ ] Rules auto-paused by the system (account needs reconnection) show a "Paused" badge and "Reconnect account" link
- [ ] Deleting a rule shows a confirmation: *"Delete this rule? This can't be undone."* with "Delete" and "Cancel" buttons
- [ ] Empty state shown when no rules exist

**UI Copy:**

*Automation tab label:* "Automation"

*Empty state:*
- **Headline:** "No auto-reply rules yet"
- **Subtext:** "Set up rules to automatically reply to TikTok comments or hide them based on keywords — so you're always responsive, even when you're offline."
- **CTA:** "Create Your First Rule"

*Rule builder modal:*
- **Title:** "Create Auto-Reply Rule"
- **Rule name label:** "Rule name"
- **Rule name placeholder:** "e.g., Reply to price questions"
- **Applies to label:** "Applies to"
- **Applies to placeholder:** "Select TikTok account(s)"
- **Applies to helper text:** "Choose which TikTok accounts this rule monitors. Select 'All TikTok accounts' to apply it to every connected account."
- **Trigger label:** "When…"
- **Trigger options:**
  - "Any new comment is posted" — tooltip: *"This rule fires on every new comment your TikTok videos receive, no matter what the comment says."*
  - "Comment contains a keyword" — tooltip: *"This rule fires when a comment includes any of your chosen words or phrases. For example, add 'price' and 'cost' to catch pricing questions."*
  - "First comment from a user" — tooltip: *"This rule fires the first time someone comments on any of your TikTok videos. Great for welcoming new commenters."*
- **Keywords label:** "Keywords"
- **Keywords placeholder:** "Type a keyword and press Enter"
- **Keywords helper text:** "Add words or phrases to watch for. A comment only needs to match one keyword to trigger the rule. For example: 'price', 'shipping', 'available'."
- **Action label:** "Then…"
- **Action options:**
  - "Reply with a template" — tooltip: *"ContentStudio will automatically post a reply using one of your templates. If you add multiple templates, it will rotate between them to keep your replies feeling natural."*
  - "Hide the comment" — tooltip: *"The comment will be hidden from everyone except the commenter. They won't know their comment is hidden, but other viewers won't see it."*
  - "Reply and hide the comment" — tooltip: *"ContentStudio replies first, then hides the comment. Useful for handling spam or off-topic comments while still acknowledging the person."*
- **Response template label:** "Reply template"
- **Response template placeholder:** "Type your reply here… (max 150 characters)"
- **Add variant button:** "+ Add another response"
- **Add variant helper text:** "Add up to 5 different responses. ContentStudio will rotate between them randomly — so your replies don't all look the same."
- **Primary CTA:** "Save Rule"
- **Secondary CTA:** "Cancel"

*Rule list item — paused badge:*
- "Paused — account needs reconnecting" (amber badge) + "Reconnect" link

*Delete confirmation modal:*
- **Title:** "Delete this rule?"
- **Body:** "This rule will be permanently deleted. Any auto-replies it would have sent won't happen anymore."
- **CTA:** "Delete Rule" / "Cancel"

*Validation errors:*
- Empty rule name: *"Please give your rule a name so you can find it later."*
- No keywords: *"Please add at least one keyword for this rule to watch for."*
- No response template: *"Please add at least one reply template."*
- Response over 150 chars: *"This reply is too long. TikTok allows a maximum of 150 characters per comment reply."*

---

**Mock-ups:** See [Design] TikTok inbox UI design story

**Impact on existing data:** None — creates new rule records via backend API

**Impact on other products:** None

**Dependencies:**
- Depends on **[BE] Build auto-reply rule engine for TikTok comments**
- Depends on **[Design] Design TikTok inbox comment cards, video context panel, and auto-reply rule builder**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

**Implementation references**
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-frontend/src/modules/inbox-revamp/InboxView.vue` — add Automation tab to inbox navigation
- New components to create in `contentstudio-frontend/src/modules/inbox-revamp/components/automation/`:
  - `AutoReplyRuleList.vue` — list + empty state
  - `AutoReplyRuleItem.vue` — single rule row with toggle, edit, delete
  - `AutoReplyRuleBuilderModal.vue` — rule creation/edit modal

**Components to use:**
- `Modal` from `@contentstudio/ui` for rule builder and delete confirmation modals
- `SegmentedControl` or `Radio` from `@contentstudio/ui` for trigger and action selectors
- `TextInput` from `@contentstudio/ui` for rule name field
- `Textarea` from `@contentstudio/ui` for response template fields
- `Dropdown` from `@contentstudio/ui` for "Applies to" account selector
- `Button` from `@contentstudio/ui` for all CTAs
- `Badge` from `@contentstudio/ui` for "Paused" status badge
- `Checkbox` from `@contentstudio/ui` if multi-account selection uses checkboxes

**Note:** The keyword tag-entry input (type + Enter to add a tag) is not currently in `@contentstudio/ui`. Either a custom input will be needed, or this can be implemented using `TextInput` with a tag list rendered below it. Flag to Design for confirmation.

---

---

### Story 13: [Design] Design TikTok inbox comment cards, video context panel, and auto-reply rule builder

**Description:**
As a ContentStudio designer, I want to deliver finalized designs for the TikTok inbox comment cards, the video context panel shown in the thread view, and the auto-reply rule builder modal so that frontend engineers have a clear visual spec to implement from — with all states, variants, and UI copy defined.

---

**Workflow:**
1. Designer reviews this epic's PRD, all FE story specs (Stories 7, 8, 10, 12), and the workflow diagrams
2. Designer produces Figma designs covering:
   - TikTok comment card in the inbox list (default, hover, selected, unread states)
   - Video context panel at the top of the TikTok thread view (thumbnail + title + "View on TikTok" link)
   - "Hidden" badge on comment cards and in the thread view
   - TikTok reconnect banner in the inbox (single account and multi-account variants)
   - Automation tab in the inbox navigation
   - Auto-reply rule list (default, empty state, rule with "Paused" badge)
   - Auto-reply rule builder modal (all trigger/action combinations, validation error states, all 5 response variant slots shown)
   - Delete confirmation modal
3. Designer provides all UI copy as specified in FE stories — no placeholder copy in designs
4. Designs are reviewed and approved before frontend implementation begins

---

**Acceptance criteria:**
- [ ] Figma designs delivered for all screens listed in the workflow above
- [ ] All states designed: default, hover, selected, loading, empty, error, paused
- [ ] All UI copy from FE stories (7, 8, 10, 12) is embedded in the designs — no lorem ipsum or placeholder text
- [ ] Component usage follows `@contentstudio/ui` design system — no custom components unless flagged as gaps
- [ ] Keyword tag entry input design provided (this component does not exist in `@contentstudio/ui` yet — designer to confirm approach)
- [ ] Video context panel shows clear visual hierarchy: thumbnail → title → "View on TikTok" link
- [ ] Reconnect banner designed to be dismissible without requiring immediate action (non-blocking)

---

**Mock-ups:** N/A — this story produces the mock-ups

**Impact on existing data:** None

**Impact on other products:** None

**Dependencies:** None — design can begin in parallel with backend stories

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness — N/A, Design story
- [ ] Multilingual support — N/A, Design story
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment — N/A, Design story

---

## Story Summary

| # | Title | Team | Project | Priority | Product Area | Skill Set |
|---|---|---|---|---|---|---|
| 1 | [BE] Extend TikTok OAuth to support Business Account scopes for comment management | Backend | Web App | High | Integrations | Backend |
| 2 | [FE] Add TikTok Business Account connection UI in Social Accounts settings | Frontend | Web App | High | Integrations | Frontend |
| 3 | [BE] Extend Easy Connect to generate TikTok Business Account authorization links | Backend | Web App | High | Integrations | Backend |
| 4 | [FE] Add TikTok Business Account to the Easy Connect shareable link flow | Frontend | Web App | High | Integrations | Frontend |
| 5 | [BE] Build TikTok comment ingestion worker and inbox queue integration | Backend | Web App | High | Inbox | Backend |
| 6 | [BE] Implement TikTok comment reply and hide/unhide API endpoints | Backend | Web App | High | Inbox | Backend |
| 7 | [FE] Display TikTok comments in the unified inbox with video context | Frontend | Web App | High | Inbox | Frontend |
| 8 | [FE] Add TikTok comment reply composer with 150-character limit, hide/unhide, and saved replies | Frontend | Web App | High | Inbox | Frontend |
| 9 | [BE] Implement TikTok token refresh and reconnect notification system | Backend | Web App | High | Inbox | Backend |
| 10 | [FE] Show TikTok reconnect banner in inbox when token expires | Frontend | Web App | High | Inbox | Frontend |
| 11 | [BE] Build auto-reply rule engine for TikTok comments | Backend | Web App | High | Inbox | Backend |
| 12 | [FE] Build auto-reply rule builder and management UI for TikTok comments | Frontend | Web App | High | Inbox | Frontend |
| 13 | [Design] Design TikTok inbox comment cards, video context panel, and auto-reply rule builder | Design | Web App | High | Inbox | Design |
