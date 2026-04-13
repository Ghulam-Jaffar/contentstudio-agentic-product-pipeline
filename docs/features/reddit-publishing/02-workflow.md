# Reddit Publishing — Workflow Design
*Feature Pipeline · Step 2 · Generated 2026-04-03*

---

## 1. Feature Placement

### Where Reddit Lives in ContentStudio

Reddit integrates into the same four areas as every other social platform:

| Area | Placement |
|---|---|
| **Settings → Social Accounts** | Reddit listed in "Connect Social Account" modal alongside Facebook, Instagram, LinkedIn, etc. Connected Reddit accounts appear in the accounts table with an orange Reddit icon and `Token valid / Token expired` status pills |
| **Composer (Publisher)** | When a Reddit account is selected in the Social tab: (1) Title field appears above the common text description box, (2) Source URL field appears below the common text description box, (3) a Reddit Settings section appears below the common message area with: Post Type radio selector (Text | Image/Video | Link), subreddit selector, flair dropdown, OC toggle, Spoiler toggle, and conflict warnings at bottom |
| **Planner / Calendar** | Reddit posts appear in the calendar view with Reddit's orange logo badge, same as all other platforms |
| **Analytics** | Reddit account analytics card: karma, post count, top posts by upvotes/comments |

### Navigation Entry Points
1. **Settings → Workspace Settings → Social Accounts** — Primary connection point
2. **Composer → Social tab → Account selector** — Reddit accounts appear here once connected
3. **Onboarding wizard** — Reddit listed in platform selection step (if applicable)
4. **EasyConnect** — Agency can share an EasyConnect link; client connects their Reddit account

---

## 2. Step 1 — Connecting a Reddit Account

### Happy Path

```
1. User goes to Settings → Social Accounts
2. Clicks "Connect Social Account" button (top-right of filter bar)
3. "Connect Social Accounts" modal opens — shows platform list
4. Reddit row is visible with orange Reddit icon, "(Profiles)" subtext, and "+" button
   → If already has connected accounts: shows count badge (e.g., "2")
5. User clicks "+" next to Reddit (or clicks anywhere on the Reddit row)
6. Reddit OAuth flow begins:
   a. ContentStudio opens Reddit's authorization page in a new tab/popup
      URL: https://www.reddit.com/api/v1/authorize?client_id=...&scope=identity+submit+read+flair+mysubreddits&response_type=code&state=...&redirect_uri=...&duration=permanent
   b. User sees Reddit's permission screen: "ContentStudio would like to: Know your Reddit username and signup date / Submit links and comments from your account / Read posts..."
   c. User clicks "Allow"
7. Reddit redirects back to ContentStudio callback URL with authorization code
8. ContentStudio backend exchanges code for access_token + refresh_token
9. Backend fetches Reddit profile: username, avatar, karma, account age
10. "Connected Successfully" state shown in the modal:
    → Reddit avatar, u/username, karma count, account age
    → "Connect Another" + "Done" buttons
11. User clicks "Done"
12. Modal closes
13. Reddit account appears as a new row in the Social Accounts table:
    → Name: u/username | Type: Reddit Profile
    → Token Status: "Token valid" (green pill)
    → Platform: Reddit
14. Toast: "u/username connected to Reddit successfully!"
```

### Connection Modal UI Detail
The Reddit connection modal (shown after clicking "+") follows the same pattern as Telegram's 3-step modal but is simpler because Reddit uses standard OAuth (no bot setup):

```
┌─────────────────────────────────────────────────────┐
│  Connect Your Reddit Account                     [✕] │
│  ─────────────────────────────────────────────────── │
│  You'll be redirected to Reddit to authorize         │
│  ContentStudio. Once connected, you can schedule     │
│  posts to any subreddit your account can access.     │
│                                                      │
│  What ContentStudio will be able to do:              │
│  ✓  Submit posts to subreddits                       │
│  ✓  Read your username and profile info              │
│  ✓  View subreddits you moderate/subscribe to        │
│  ✓  Read and apply post flairs                       │
│                                                      │
│  What ContentStudio cannot do:                       │
│  ✗  Vote on your behalf                              │
│  ✗  Access private messages                          │
│  ✗  Access NSFW communities                          │
│                                                      │
│         [ Cancel ]   [ Connect with Reddit ↗ ]       │
└─────────────────────────────────────────────────────┘
```

After OAuth completes and callback returns:

```
┌─────────────────────────────────────────────────────┐
│  Connect Your Reddit Account                     [✕] │
│  ─────────────────────────────────────────────────── │
│              ✅ (success ring)                        │
│         Account Connected!                           │
│  Your Reddit account is now linked to ContentStudio. │
│                                                      │
│  ┌──────────────────────────────────────────┐        │
│  │ [avatar]  u/brandname_official           │        │
│  │           Karma: 1,240 · 3 years old     │        │
│  │           [Reddit Profile] [✓ Connected] │        │
│  └──────────────────────────────────────────┘        │
│                                                      │
│   [ Connect Another ]          [ Done ]              │
└─────────────────────────────────────────────────────┘
```

---

## 3. Step 2 — Creating a Reddit Post in Composer

### Happy Path

```
1. User opens Composer (Publisher → Create Post)
2. In "Social" tab, user selects one or more Reddit account(s)
3. Title field appears ABOVE the common text description box:
   → Label: "Title" (if only Reddit and/or non-Pinterest platforms selected)
   → Label: "Title (Pinterest, Reddit)" (if both Pinterest and Reddit are selected)
   → Required, 300-char limit with live counter
4. User types in the common text description box (body text for text posts)
5. Source URL field appears BELOW the common text description box:
   → Label: "Source URL" (if only Reddit selected)
   → Label: "Source URL (Reddit, Pinterest)" (if both Reddit and Pinterest selected)
   → Auto-populated from any URL detected in the common message box
   → Only submitted to Reddit when post type = Link
6. A "Reddit Settings" section appears below the common message area:

   ┌──────────────────────────────────────────────────────────┐
   │ 🟠 Reddit Settings                    [ Customize ]      │
   │                                                          │
   │ Post Type                                                │
   │ ● Text   ○ Image/Video   ○ Link   ℹ                   │
   │                                                          │
   │ Subreddit *                                              │
   │ [🔍 Search subreddit (e.g. r/marketing)             ▾ ] │
   │                                                          │
   │ Flair  ℹ  (appears after subreddit is selected)          │
   │ [Select flair...                                    ▾ ] │
   │                                                          │
   │ ☐ Original Content (OC)   ☐ Spoiler                     │
   │                                                          │
   │ ─── Conflict warnings (shown when applicable) ────       │
   │ ⚠ [warning / error / info alert here]                    │
   └──────────────────────────────────────────────────────────┘

7. User selects post type from the "Post Type" radio row:
   → Default: Text

8. User selects "Text":
   → Body text from the common message box is used as Reddit body
   → User can click "Customize" to write Reddit-specific body (overrides common message)
   → Markdown supported

9. User selects "Image/Video":
   → Media uploaded in the common section is used (first image/video)
   → Body text is NOT submitted to Reddit (info in conflict warnings area if body has content)
   → Source URL is NOT submitted to Reddit (error in conflict warnings area if URL is filled)

10. User selects "Link":
    → Source URL field (below the text box) is used as the Reddit link URL
    → Body text is NOT submitted to Reddit (info notice if body has content)
    → Validation error shown below Source URL if it is empty

11. User types in Subreddit field:
    → Autocomplete search fires after 2 characters
    → Shows: subreddit name + subscriber count + lock icon if private
    → e.g.: "r/marketing · 1.2M members", "r/entrepreneur · 890K members"

12. After subreddit is selected:
    → Flair dropdown loads dynamically from Reddit API
    → ℹ tooltip on Flair: "Some subreddits require a flair before you can post — if yours does,
      you must select one here. If your subreddit doesn't require flair, you can skip this."
    → If subreddit has no flair: dropdown hidden entirely
    → If subreddit requires flair: "(Required)" shown, publish blocked without selection
    → Flair options show color swatches if subreddit uses colored flairs

13. User toggles OC and/or Spoiler if needed
14. User fills in title + selects subreddit + sets flair
15. User schedules or publishes
```

### Post-type specific behavior

**Text (kind: self):**
- Title: required, 1–300 chars (live counter)
- Body: from the common message box; user can override via Customize button; optional, 0–40,000 chars, Markdown supported
- Source URL: visible below the text box but NOT submitted to Reddit

**Image/Video (kind: image):**
- Title: required, 1–300 chars
- Image/video: pulled from media uploaded in common section (first attachment for v1)
- Body text: NOT submitted (Reddit API limitation — shown as non-blocking warning in conflict area if body has content)
- Source URL: NOT submitted — shown as blocking error in conflict area if Source URL is filled
- Supported: JPEG, PNG, GIF (up to 20MB)

**Link (kind: link):**
- Title: required, 1–300 chars
- Source URL: required — from the Source URL field below the text box (auto-populated from any URL in common message box)
- Body text: NOT submitted to Reddit (Reddit link posts don't support it — shown as non-blocking info notice if body has content)
- Reddit auto-generates thumbnail from the URL — shown as info note

---

## 4. Step 3 — Scheduling a Reddit Post

```
1. User composes Reddit post (Flow B above)
2. Clicks "Schedule" → date/time picker appears
3. Post is added to the publishing queue
4. In Planner/Calendar:
   → Reddit post appears with orange Reddit badge on the scheduled date/time
   → Clicking the post shows preview: title, post type, subreddit, flair (if set)
5. At scheduled time:
   → Backend publishes to Reddit via POST /api/submit
   → If flair was marked required but is missing: post fails → notification sent
   → If duplicate post detected: post fails → notification sent with reason
   → If subreddit karma requirement not met: post fails → notification sent
6. After publishing:
   → Post status updates in Planner: "Published"
   → Direct link to Reddit post stored: https://reddit.com/r/{subreddit}/comments/{post_id}
   → Analytics start populating after first data sync
```

---

## 5. Step 4 — Reconnecting / Token Expiry

```
1. Reddit access tokens expire every 1 hour
   → ContentStudio silently refreshes using the refresh_token (no user action needed)
   → Refresh tokens valid for 6 months
2. If refresh token expires or user revokes access on Reddit:
   → Account row in Social Accounts table shows "Token expired" (red pill)
   → Warning banner in Composer if that account is selected
   → Email/in-app notification sent (if notifications enabled)
3. User clicks "↺ Reconnect" next to the account
4. Same OAuth flow as initial connection (Flow A steps 6–14)
5. New access_token + refresh_token saved
6. Account status returns to "Token valid"
```

---

## 6. Step 5 — Viewing Reddit Post Analytics

```
1. User goes to Analytics section
2. Selects Reddit account
3. Dashboard shows:
   → Upvotes (karma score) per post
   → Number of comments per post
   → Total posts published
   → Top performing post (by upvotes)
4. Per-post detail: direct link to Reddit post
5. Data synced via periodic background job polling GET /api/info?id=t3_{post_id}
```

---

## 7. Alternative Flows & Edge Cases

| Scenario | Behavior |
|---|---|
| **User denies Reddit OAuth** | OAuth callback returns `error=access_denied`. Modal shows error: "Reddit authorization was denied. Please try again and click 'Allow' when prompted." |
| **Reddit account already connected** | On callback, backend checks if `platform_identifier` (Reddit user ID) already exists in workspace. Shows: "u/username is already connected to this workspace." |
| **Subreddit not found** | Autocomplete shows "No subreddits found" message. User cannot proceed without a valid subreddit. |
| **Subreddit is private/restricted** | Subreddit shown with 🔒 icon in autocomplete. User can select it, but post will fail if account is not a member. Error shown post-publish: "You don't have access to post in r/subreddit." |
| **Flair required but not selected** | Publish button disabled with tooltip: "This subreddit requires a flair. Please select one before publishing." |
| **Duplicate post detected** | Reddit returns `ALREADY_SUB` error. ContentStudio shows: "This URL was already submitted to r/subreddit recently. Reddit prevents duplicate posts. Try a different subreddit or wait before resubmitting." |
| **Post title empty** | Inline validation: "Post title is required for Reddit." Publish button disabled. |
| **Post title > 300 chars** | Character counter turns red. Publish button disabled. Inline error: "Reddit titles cannot exceed 300 characters." |
| **Image too large (>20MB)** | Standard media upload validation: "Max. file size for Reddit is 20MB." |
| **Account karma too low** | Reddit returns `NOT_WHITELISTED_BY_USER_IN_SUBREDDIT` or karma error. ContentStudio shows: "Your Reddit account doesn't meet this subreddit's minimum requirements. Check the subreddit rules." |
| **Token refresh fails mid-publish** | Post marked as failed in Planner. Notification sent. Account marked `Token expired`. |
| **Multiple Reddit accounts selected** | Each account publishes independently to its specified subreddit. Each has its own subreddit + flair selection (per-account fields) |
| **Common box used with Reddit selected** | Common message box text is used as Reddit body for text posts. User can click 'Customize' in the Reddit Settings header to write Reddit-specific body text instead. Source URL below text box auto-populates from any URL detected in the common message. |
| **Post type = Image/Video selected, Source URL is filled** | Blocking error shown at the bottom of Reddit Settings: "Image posts don't support links. Clear the Source URL or switch to Link post type." Publish button disabled. |
| **Post type = Link selected, Source URL empty** | Inline validation error below Source URL field: "A URL is required for Link posts." Publish button disabled. |
| **Scheduling Reddit post in the past** | Standard ContentStudio validation — "Please choose a future time." |

---

## 8. Key Design Decisions

### Decision 1: Per-post subreddit vs. per-account subreddit

**Option A — Subreddit selected per post (at compose time)**
The user picks which subreddit to post to every time they create a post.
- ✅ Flexible — same Reddit account can post to any subreddit
- ✅ Matches how Reddit-specific tools (Postpone, Social Rise) work
- ✅ Matches VistaSocial's approach
- ❌ Adds friction for users who always post to the same subreddit

**Option B — Default subreddit saved on the account**
User sets a default subreddit on the account, can override per post.
- ✅ Faster for users with a dedicated brand subreddit
- ❌ Reddit accounts often post to multiple subreddits; default may mislead

**→ Recommendation: Option A** (per-post subreddit selection). This is the standard pattern across all Reddit tools and aligns with how Reddit itself works. We can add a "recently used subreddits" shortcut list as a UX enhancement.

---

### Decision 2: Where does Reddit's "Title" field live?

Reddit requires a separate post title (1–300 chars) — unlike all other ContentStudio platforms which use a single message body.

**Option A — Use the common message box first line as the title**
Automatically split: first line = title, rest = body.
- ✅ Zero extra UI
- ❌ Users don't know this rule; creates confusion and titles easily exceed 300 chars

**Option B — Dedicated Title field above the common text description box**
A clearly labeled "Title" input appears above the common message textarea, following the same placement as Pinterest's Title field.
- ✅ Clear UX, matches Reddit's own interface
- ✅ Allows common message box to be used as-is (for cross-platform scheduling)
- ✅ Same pattern VistaSocial uses
- ✅ Label dynamically adapts: "Title (Pinterest, Reddit)" when both platforms are selected — single shared field

**→ Recommendation: Option B — BUT positioned **above the common text description box** (not inside the Reddit Settings section), following the same pattern as Pinterest's Title field. This keeps it prominently visible and allows the label to dynamically indicate which platforms are using it (e.g., 'Title (Pinterest, Reddit)' when both platforms are selected). The Reddit Settings section stays focused on Reddit-specific controls only.**

**Note: Source URL field placement** — Reddit's link post URL reuses the existing Source URL field that Pinterest already uses, positioned below the common text description box. Label is dynamic: "Source URL" when only Reddit is selected; "Source URL (Reddit, Pinterest)" when both are selected. This avoids adding a duplicate URL field and keeps the composer layout clean.

---

### Decision 3: Post type selection — explicit vs. auto-detect

**Option A — Auto-detect post type**
If media is attached → Image post. If URL is in message → Link post. Otherwise → Text post.
- ✅ Less clicks
- ❌ Ambiguous when both URL and text are present
- ❌ User may not realize it became a link post when they just wanted to mention a URL in text

**Option B — Explicit post type radio selector**
User explicitly picks: **Text** | **Image/Video** | **Link** — via a "Post Type" radio row inside the Reddit Settings section (same pattern as YouTube Settings).
- ✅ No ambiguity
- ✅ Matches Reddit's own "Create Post" interface with tabs
- ✅ VistaSocial uses explicit type selection

**→ Recommendation: Option B** (explicit post type selector — radio button row in Reddit Settings, same pattern as YouTube's "Post Type" row). Options: **Text | Image/Video | Link**. Default selection: Text. Conflict warnings shown at the bottom of Reddit Settings when selected type conflicts with added content, rather than blocking the user from adding content.

---

### Decision 4: Multi-subreddit posting

**Option A — One subreddit per post (v1)**
A post goes to one subreddit. To post to multiple subreddits, user duplicates/uses bulk scheduling.
- ✅ Simpler implementation
- ✅ Avoids Reddit spam detection (same post to multiple subreddits simultaneously triggers spam filters)
- ❌ Repetitive for agencies managing brand subreddits across regions

**Option B — Multiple subreddit selector with time staggering (v2)**
User can pick multiple subreddits; system staggers posts automatically (e.g., 30-min delay between each).
- ✅ Power user feature
- ✅ Staggering avoids Reddit's spam detection
- ❌ Complex UI and scheduling logic

**→ Decision: Both in v1.** Single subreddit is the default; multi-subreddit with staggered scheduling is also included in v1 scope. System auto-staggers posts (30-min delay between each subreddit) to avoid Reddit's spam detection.

---

## 9. Integration with Existing ContentStudio Features

| Feature | How Reddit Integrates |
|---|---|
| **Common message box** | Reddit body (text posts) pre-fills from common box. Reddit title is separate. |
| **Media Library** | Images/GIFs uploaded via media library can be used for Reddit image posts |
| **Link shortener** | Applied to URLs in Reddit link posts if workspace has link shortener enabled |
| **Content Categories** | Reddit posts can be tagged with content categories like any other platform |
| **Planner / Calendar** | Reddit posts appear in calendar with orange Reddit badge |
| **Bulk Scheduling / CSV** | Reddit posts included in bulk scheduler — title and subreddit as required CSV columns |
| **First Comment** | Not applicable for Reddit v1 (comment scheduling is a separate feature) |
| **UTM Parameters** | Appended to URLs in Reddit link posts |
| **AI Caption Generation** | Common AI caption box works for Reddit body text. Reddit-specific AI prompt can generate community-tone content (v2) |
| **Post Preview** | Reddit post preview shows formatted title, post type badge, subreddit name, and body/image/link |
| **EasyConnect** | Agency clients can connect their Reddit account via EasyConnect link |
| **Analytics Dashboard** | Reddit account card added with karma, comments, top posts |
| **Inbox / Listening** | Out of scope for v1. Reddit comments on published posts are not surfaced in Inbox until v2. |
| **Automations** | RSS-to-Reddit posting: when RSS feed publishes new article, auto-post as link post to specified subreddit (v2 via automation rules) |

---

## 10. Scope Recommendation

### v1 — Launch Scope

**Account Management:**
- [x] Connect Reddit account via OAuth 2.0 (web only, no mobile)
- [x] Multiple Reddit accounts per workspace
- [x] Account display in Social Accounts table (name, type "Reddit Profile", token status)
- [x] Token silent auto-refresh (every <1 hour)
- [x] Token expiry detection + "Reconnect" flow
- [x] EasyConnect support (agency link sharing)

**Composer / Publishing:**
- [x] Reddit account selection in Social tab
- [x] Explicit post type selector (radio row in Reddit Settings): Text | Image/Video | Link
- [x] Dedicated Title field above common text description box (1–300 chars with counter)
- [x] Body field for text posts — from common message box; Customize button to override per Reddit
- [x] Source URL field below text box shared with Pinterest (auto-populates from common message box; used by Reddit for Link post type)
- [x] Image support (JPEG, PNG, GIF, max 20MB)
- [x] Subreddit selector with autocomplete search
- [x] Dynamic flair loading per subreddit
- [x] Mandatory flair enforcement (block publish if required)
- [x] OC (Original Content) toggle (default: off)
- [x] Spoiler toggle (default: off)
- [x] Conflict warnings at bottom of Reddit Settings section (post type vs. content conflicts)
- [x] Inline validation (title required, title length, flair required, Source URL required for Link posts)

**Scheduling:**
- [x] One-time scheduling
- [x] Calendar / Planner visibility with Reddit orange badge
- [x] Published post link stored (https://reddit.com/r/{sub}/comments/{id})

**Error Handling:**
- [x] Duplicate post detection (`ALREADY_SUB`) with user-friendly message
- [x] Karma/eligibility errors surfaced to user
- [x] Token expiry mid-publish → retry and notify

**Analytics:**
- [x] Upvotes (karma) per post
- [x] Comment count per post
- [x] Total posts published from ContentStudio

---

### v2 — Deferred to Later

**Composer:**
- [ ] GIF post type (extension of image)
- [ ] Video post type (title + MP4 via Reddit video upload API)
- [ ] Gallery post (multi-image, up to 20)
- [ ] Poll creation (title + options + duration)
- [ ] Crosspost
- [ ] Rich text / Markdown editor for body (v1 uses plain text)
- [ ] "Recently used subreddits" quick-select
- [ ] Staggered multi-subreddit scheduling

**Intelligence:**
- [ ] Per-subreddit best-time recommendations
- [ ] Subreddit discovery (suggest subreddits by topic)
- [ ] Account karma health check per subreddit (eligibility indicator)
- [ ] AI Reddit-native content generation (community tone)

**Analytics:**
- [ ] Upvote ratio
- [ ] Award count
- [ ] View count (if available via API)
- [ ] Post benchmarking vs. subreddit average

**Inbox / Listening:**
- [ ] Reddit Smart Inbox (replies to published posts)
- [ ] Reddit mentions tracking
- [ ] Subreddit monitoring / keyword alerts

**Automation:**
- [ ] RSS-to-Reddit auto-posting rule

---

## 11. Pre-Requisites Before Development Starts

| # | Pre-requisite | Owner | Notes |
|---|---|---|---|
| 1 | **Apply for Reddit Commercial API Access** | Product / Business Dev | Reddit reviews applications manually. Apply at developers.reddit.com. Requires use case description, estimated API volume, and terms acceptance. Timeline: 2–6 weeks. This is the #1 blocker. |
| 2 | **Reddit Developer App Registration** | Engineering | Create app at reddit.com/prefs/apps — type: "web app", set redirect URI to ContentStudio callback URL. Obtain `client_id` and `client_secret`. |
| 3 | **Define OAuth Scopes** | Engineering | Minimum needed: `identity submit read flair mysubreddits`. Add `history` for analytics. Duration: `permanent` (for refresh tokens). |
| 4 | **ContentStudio Bot / Brand Reddit Account** | Marketing | Create/designate an official `u/contentstudio` Reddit account for the bot. ContentStudio should have karma > 0 before API launch (build organic presence first). |
| 5 | **Set Callback URL** | Engineering | Register `https://app.contentstudio.io/oauth/reddit/connect` (or equivalent) in the Reddit developer app settings. |
| 6 | **Design Reddit icon / badge** | Design | Frontend `icon-mapping.js` already has `'reddit': '__brand__Reddit'` — confirm the Reddit icon is in the `@contentstudio/ui` icon set. If not, add it. |
| 7 | **Review Reddit Developer Agreement** | Legal / Product | Reddit's Developer Terms prohibit certain use cases. Confirm ContentStudio's use case is compliant. |
