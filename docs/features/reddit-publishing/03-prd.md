# PRD: Reddit Publishing Integration

**Author:** Product Team
**Last Updated:** 2026-04-03
**Status:** Draft
**Target Release:** Q2 2026

---

## 1. Overview

Reddit Publishing adds Reddit as a fully supported social platform in ContentStudio, allowing users to connect their Reddit accounts via OAuth 2.0 and schedule or publish text posts, link posts, image posts, and (in v2) polls directly to any subreddit — all from within the ContentStudio composer, planner, and analytics dashboard.

Reddit is the only major social network not yet supported by the mainstream SMM tools (Buffer, Hootsuite, Later, Metricool, Publer have no native Reddit). Only VistaSocial and Sprout Social have launched it. This is a clear market gap. For ContentStudio's agency and brand users, Reddit represents a high-intent audience channel — users come to Reddit specifically to research products, ask questions, and evaluate tools — making it valuable for B2B and B2C brands alike.

---

## 2. Problem Statement

**What problem are we solving?**

ContentStudio users who manage brands with a Reddit presence have no way to schedule or publish Reddit content from ContentStudio. They are forced to maintain a separate Reddit-specific workflow (manual posting, or using a Reddit-only tool like Postpone) on top of their ContentStudio workflow, fragmenting their social media operation. This means lost time, inconsistent posting schedules, and a reason to consider competitor tools that do support Reddit.

**Who has this problem?**

- **Social media managers** at B2B SaaS companies, tech brands, gaming companies, and developer tools — all industries with strong Reddit communities
- **Digital agencies** managing multiple client brands, where even one or two clients have Reddit audiences
- **Content marketers** who run link-building campaigns or distribute blog content via relevant subreddits
- **Community managers** who maintain brand subreddits and post regular updates

Based on the competitive landscape, Reddit support is a missing feature that surfaces in user forums, review sites (G2, Capterra), and support requests from users who compare ContentStudio to VistaSocial.

**What happens if we don't solve it?**

- Users evaluating ContentStudio vs. VistaSocial cite Reddit support as a differentiator for VistaSocial
- Agency clients who use Reddit lose confidence in ContentStudio as their "all-in-one" tool
- Risk of churn to VistaSocial specifically, which markets Reddit as a headline feature
- Missed opportunity to be first among Buffer/Hootsuite/Later/Metricool — all of whom have not launched Reddit native publishing

---

## 3. Goals & Success Metrics

| Goal | Metric | Target | How We'll Measure |
|---|---|---|---|
| Drive Reddit account connections | Number of Reddit accounts connected across workspaces | 500 accounts within 60 days of launch | Product analytics (social integrations DB) |
| Drive Reddit post volume | Number of posts published to Reddit per month | 2,000 posts/month by month 3 | Publishing logs |
| Reduce competitive churn risk | Reduction in churn mentions of VistaSocial as alternative | 20% reduction in VistaSocial mentions in churn surveys | Churn survey data |
| Zero critical publish failures due to our bugs | Reddit post failure rate due to ContentStudio bugs (excluding Reddit API errors) | < 1% failure rate | Error logs / Sentry |
| Flair enforcement prevents Reddit-side rejections | Posts rejected by Reddit due to missing required flair | 0 occurrences (flair enforced before submission) | Reddit API error logs |

---

## 4. Target Users

**Primary Persona — The Social Media Manager**
A full-time social media manager at a mid-size B2B or consumer brand. Posts 5–20 times per week across 4–8 platforms. Uses ContentStudio as their primary scheduling tool. Has a brand Reddit account and posts to 1–3 subreddits regularly (e.g., r/marketing, r/entrepreneur, brand's own subreddit). Currently handles Reddit manually. Wants Reddit in ContentStudio so everything is in one calendar.

**Secondary Persona — The Digital Agency Account Manager**
Manages 5–20 client workspaces in ContentStudio. Some clients are in industries with strong Reddit communities (tech, gaming, finance, developer tools). Uses EasyConnect to onboard client social accounts. Wants to offer Reddit scheduling as part of the agency's social media package without adding separate tooling.

**Non-Users (out of scope):**
- Personal Reddit users — ContentStudio is for business/brand accounts, not personal use
- Users wanting to monitor/listen to Reddit conversations (that's a Listening feature, not publishing)
- Users wanting to post to NSFW communities (blocked by Reddit API for 3rd parties)

---

## 5. User Stories / Jobs to Be Done

| ID | As a... | I want to... | So that... | Priority |
|---|---|---|---|---|
| US-1 | Social media manager | Connect my brand's Reddit account to ContentStudio | I can manage Reddit from the same tool as all other platforms | Must Have |
| US-2 | Social media manager | Schedule a text post to a subreddit with a title and body | I can share thought leadership content to relevant communities at the right time | Must Have |
| US-3 | Social media manager | Schedule a link post to a subreddit | I can share blog articles, product pages, and announcements to Reddit audiences | Must Have |
| US-4 | Social media manager | Schedule an image post to a subreddit | I can share visual content, infographics, and brand imagery on Reddit | Must Have |
| US-5 | Social media manager | Search for and select a subreddit when composing a Reddit post | I can target the right community without leaving the composer | Must Have |
| US-6 | Social media manager | See which flairs are available for a subreddit and be required to pick one if the subreddit enforces flair | My posts aren't rejected by Reddit for missing a required flair | Must Have |
| US-7 | Social media manager | Post the same content to multiple subreddits with automatic time staggering | I can reach multiple communities without triggering Reddit's spam detection | Must Have |
| US-8 | Social media manager | See my Reddit posts in the ContentStudio calendar view | I have a unified content calendar across all platforms including Reddit | Must Have |
| US-9 | Social media manager | Receive a clear error message when a Reddit post fails (duplicate, karma, flair) | I understand why it failed and can fix it without guessing | Must Have |
| US-10 | Social media manager | Reconnect my Reddit account when the token expires | My scheduled posts aren't blocked due to a stale connection | Must Have |
| US-11 | Agency account manager | Share an EasyConnect link so clients can connect their own Reddit account | I can onboard Reddit accounts without needing client credentials | Must Have |
| US-12 | Social media manager | See upvotes and comments on my Reddit posts in ContentStudio analytics | I can measure Reddit performance without switching to a separate tool | Should Have |
| US-13 | Social media manager | Create a Reddit poll post with custom options and duration | I can run community engagement polls from ContentStudio | Nice to Have (v2) |
| US-14 | Social media manager | Schedule a video post to a subreddit | I can share video content on Reddit from ContentStudio | Nice to Have (v2) |
| US-15 | Social media manager | Schedule a gallery post with multiple images | I can share image sets (e.g., product launches) as a Reddit gallery | Nice to Have (v2) |

---

## 6. Requirements

### 6.1 Must Have (P0)

**Account Connection**
- Connect a Reddit account via standard OAuth 2.0 flow (redirect to Reddit authorization page)
- Support multiple Reddit accounts per workspace
- OAuth scopes requested: `identity submit read flair mysubreddits history`
- OAuth duration: `permanent` (obtains refresh token for long-lived access)
- Silent access token refresh (Reddit tokens expire every 1 hour; system refreshes using refresh token without user action)
- Store encrypted access token + refresh token in `SocialIntegrations` collection
- Detect token expiry / revocation and mark account as `validity: expired`
- Display Reddit accounts in Social Accounts settings table with: username, "Reddit Profile" type, token status pill (valid / expired), reconnect action
- Reconnect flow re-runs OAuth and replaces stored tokens

**Post Composition**
- Reddit account selectable in the Composer's Social tab
- **Title field** — rendered **above the common text description box** (same placement as Pinterest's title field). Required, 1–300 characters, live character counter. **Label is dynamic:**
  - Reddit selected (alone or with any non-Pinterest platform) → "Title"
  - Reddit + Pinterest both selected → "Title (Pinterest, Reddit)" — single shared field for both platforms
  - Pinterest only (no Reddit) → "Title" (existing Pinterest behavior unchanged)
- **Source URL field** — rendered **below the common text description box** (same placement as Pinterest's Source URL field — reuses the same field). **Label is dynamic:**
  - Reddit selected (alone or with any non-Pinterest platform) → "Source URL"
  - Reddit + Pinterest both selected → "Source URL (Reddit, Pinterest)" — single shared field for both platforms
  - Pinterest only (no Reddit) → "Source URL" (existing Pinterest behavior unchanged)
  - For Reddit: Source URL is only submitted to Reddit when post type = **Link**; when post type is Text or Image/Video, the Source URL field is visible but its value is not sent to Reddit
- When Reddit account is selected, a **Reddit Settings** section renders below the common message area with:
  - **Post type selector** — "Post Type" label row at the top of the Reddit Settings section (same pattern as YouTube's "Post Type" radio row) with three radio button options: **Text**, **Image/Video**, **Link**. Default: Text. User selects one; the selection determines exactly what is submitted to Reddit regardless of other content present.
  - **Subreddit selector** — text input with autocomplete search (min 2 chars to trigger), showing subreddit name + subscriber count; supports multiple subreddits (staggered)
  - **Flair selector** — dropdown that loads dynamically from Reddit API after subreddit is selected; hidden entirely if subreddit has no flairs. ℹ tooltip: "Some subreddits require a flair before you can post — if yours does, you must select one here. If your subreddit doesn't require flair, you can skip this." Labeled "(Required)" and blocks publishing when the subreddit enforces mandatory flair.
  - **OC (Original Content) toggle** — default OFF; marks post as original content
  - **Spoiler toggle** — default OFF; marks post content with Reddit's spoiler blur
- **Customize button** — When Reddit is selected alongside other platforms, a "Customize for Reddit" toggle is available in the Reddit section header. When ON, the body text area becomes Reddit-specific and overrides the common message body for Reddit only. When OFF, Reddit uses the common message body. Reddit-specific fields (post type selector, title, subreddit, flair, toggles) are always visible regardless of Customize state.
- **Conflict warnings** shown at the **bottom of the Reddit Settings section** when the selected post type conflicts with added content:
  - Post type = **Text**, image attached → `Alert` (warning): "Your attached image won't be included — text posts don't support images. Switch to 'Image/Video' to include it." Non-blocking.
  - Post type = **Image/Video**, body text written → `Alert` (warning): "Body text won't be included — image posts don't support body text." Non-blocking.
  - Post type = **Image/Video**, Source URL filled → `Alert` (error, blocks publish): "Image posts don't support links. Clear the Source URL or switch to 'Link' post type."
  - Post type = **Link**, body text written → `Alert` (info): "Body text won't be included — link posts only publish the title and URL." Non-blocking.
  - Post type = **Link**, Source URL empty → inline validation error (shown below the Source URL field in the common area): "A URL is required for Link posts."
  - Post type = **Image/Video**, no image/video attached → inline validation error inside Reddit Settings: "Please attach an image or video."
- Inline validation: title required, title ≤ 300 chars, flair required if subreddit enforces it, subreddit required

**Multi-Subreddit Posting**
- User can add multiple subreddit targets for the same post (via "+ Add Subreddit" control in the Reddit section)
- Each subreddit target has its own flair selector (loaded independently)
- System staggers publishing between subreddits with a 30-minute delay per subreddit (first subreddit posts at scheduled time; second posts 30 min later; third posts 60 min later; etc.)
- Stagger delay is automatic and shown to the user: "Posts will be staggered 30 minutes apart to comply with Reddit's guidelines"

**Scheduling & Publishing**
- One-time scheduling supported
- Published Reddit posts appear in Planner / Calendar with Reddit's orange badge icon
- After successful publish: direct Reddit post URL stored (`https://www.reddit.com/r/{subreddit}/comments/{post_id}`)
- Post status updated in Planner: "Published" with timestamp

**Error Handling — User-Facing Messages**
- `ALREADY_SUB` (duplicate post): "This URL was already submitted to r/{subreddit} recently. Reddit prevents duplicate posts within a short window. Try a different subreddit or wait before resubmitting."
- `FLAIR_REQUIRED`: "This subreddit requires a flair. Please select one before publishing." (Also blocked in UI before submission)
- `SUBREDDIT_NOEXIST`: "r/{subreddit} doesn't exist or has been banned."
- `SUBREDDIT_NOTALLOWED` / karma error: "Your Reddit account doesn't meet r/{subreddit}'s posting requirements. Check the subreddit rules (minimum karma or account age may apply)."
- `NOT_WHITELISTED_BY_USER_IN_SUBREDDIT`: "r/{subreddit} is a private or restricted community. Your account needs to be approved to post there."
- Token expired at publish time: "Reddit account u/{username} is disconnected. Please reconnect it in Social Accounts settings."
- Generic Reddit API error: "Reddit returned an error: {error_message}. Please try again or contact support."

**EasyConnect**
- Reddit included in EasyConnect link flow so agencies can share a connect link for clients to authorize their Reddit accounts
- EasyConnect link for Reddit follows the same OAuth flow as direct connection; the generated link redirects the client to Reddit's OAuth authorization page with ContentStudio's parameters pre-filled
- Connected Reddit account appears in the client's workspace Social Accounts table after EasyConnect authorization completes
- Reddit displayed in the EasyConnect platform list with the Reddit icon, label "Reddit", and subtext "(Profiles)"

### 6.2 Should Have (P1)

**Analytics**
- Reddit account analytics card in Analytics section showing:
  - Total posts published from ContentStudio
  - Total upvotes (karma) across posts
  - Total comments across posts
  - Top 5 posts by upvote score (with direct Reddit link)
- Analytics data synced via background job polling `GET /api/info?id=t3_{post_id}` every 6 hours
- Post-level analytics shown in Planner post detail: upvote score, comment count, direct link

**Duplicate Post Detection (pre-publish)**
- Before scheduling/publishing a link post, check if the same URL was recently posted to the same subreddit by the same account
- If detected, show a warning (not a block) in the composer: "This URL may have already been posted to r/{subreddit} recently. Reddit may reject it."

**Post Preview**
- Reddit post preview in the composer preview pane: shows Reddit-style card with title, post type badge, subreddit name, username, and body/image/link

### 6.3 Nice to Have (P2)

**Poll Posts (v2)**
- Post type: Poll — title + 2–6 answer options + voting duration (1 day / 3 days / 7 days)
- Available as a `kind: poll` post via Reddit's poll creation API
- Still create the Shortcut story for this even though it ships in v2

**Video Posts (v2)**
- Post type: Video — title + MP4 upload
- Requires multi-step Reddit video upload API (upload to S3 → DASH encoding → submit)

**Gallery Posts (v2)**
- Post type: Gallery — title + up to 20 images with optional per-image captions
- Uses separate Reddit endpoint `/api/submit_gallery_post.json`

**Per-Subreddit Best Time (v2)**
- Recommend optimal posting time based on subreddit activity patterns
- Shown as a suggested time in the scheduler

**Reddit Smart Inbox (v2)**
- Surface comments and replies on published Reddit posts inside ContentStudio's unified Inbox
- Allow replying to comments from ContentStudio

**AI Reddit Content Generation (v2)**
- AI caption generator mode tuned for Reddit: community-first tone, avoids marketing-speak, generates relevant title + body for a given subreddit

### 6.4 Explicitly Out of Scope

- **NSFW content** — Reddit API blocks 3rd-party access to NSFW communities since July 2023
- **Personal profile posting** — Reddit API does not support posting to a user's profile via the submit API (subreddit-only)
- **Comment-only scheduling** — Not part of the publishing feature; requires separate Inbox/Listening feature
- **Reddit Ads management** — Out of scope; entirely separate API and use case
- **Crosspost** — Low brand use case; deferred indefinitely
- **Account warm-up / karma building flows** — Out of scope; ContentStudio is a publishing tool, not a Reddit growth tool
- **Subreddit moderation tools** — ContentStudio does not moderate subreddits on behalf of users
- **Mobile apps (iOS/Android)** — Account connection is web-only (OAuth flow); Reddit posts are schedulable from web only. Mobile app will display Reddit posts in the calendar view but not support creating Reddit-specific fields.
- **Dark mode / RTL** — ContentStudio does not support either

---

## 7. User Flow (High Level)

### Connecting a Reddit Account
1. User goes to Settings → Social Accounts → clicks "Connect Social Account"
2. Selects Reddit from the platform list (shown with orange Reddit icon, "Profiles" subtext)
3. ContentStudio opens Reddit OAuth authorization page
4. User clicks "Allow" on Reddit's permission screen
5. Reddit redirects back; ContentStudio exchanges code for tokens, fetches profile
6. Success modal shows: u/username, karma, account age
7. User clicks "Done" → account appears in Social Accounts table

### Creating and Scheduling a Reddit Post
1. User opens Composer → selects Reddit account in Social tab
2. **Title field appears above the common text description box** — label reads "Title (Pinterest, Reddit)" if Pinterest is also selected, otherwise "Title"
3. User enters post title (required)
4. A **Reddit Settings** section appears below the common message area with a Customize button in the header
5. User selects post type from the "Post Type" radio row (defaults to Text)
6. User adds content appropriate for the selected post type: writes body text in the common message box for Text, attaches image/video for Image/Video, enters URL in the Source URL field below the text box for Link (label reads "Source URL (Reddit, Pinterest)" if Pinterest also selected)
7. If added content conflicts with the selected post type, a warning appears at the bottom of the Reddit Settings section
8. User optionally clicks "Customize for Reddit" to write Reddit-specific body text that overrides the common message
9. User searches for and selects a subreddit; flair loads if the subreddit supports it
10. User selects flair if available or required
11. Optionally adds more subreddits (multi-subreddit with auto-stagger)
12. User schedules or publishes immediately
13. Post appears in calendar; after publish, direct Reddit URL stored

### Token Expiry & Reconnect
1. Refresh token expires (6 months) or user revokes access on Reddit
2. Account marked "Token expired" in Social Accounts table
3. User clicks "↺ Reconnect" → OAuth flow re-runs
4. New tokens stored; account returns to "Token valid"

---

## 8. Business Rules & Constraints

| Rule ID | Rule | Rationale |
|---|---|---|
| BR-1 | A Reddit post MUST have a subreddit — no personal profile posting | Reddit API limitation; `/api/submit` requires `sr` parameter |
| BR-2 | A Reddit post MUST have a title (1–300 chars) — publish button disabled until title is provided | Reddit API requires title for all post types |
| BR-3 | User explicitly selects post type via the "Post Type" radio row in Reddit Settings: **Text**, **Image/Video**, or **Link**. The selected type determines what is submitted to Reddit regardless of what other content is present in the common message area. | Gives users explicit control; avoids unintended post type changes when composing for multiple platforms where only one uses an image |
| BR-3a | Post type = **Text** AND image attached: image excluded from Reddit submission; non-blocking warning at bottom of Reddit section: "Your attached image won't be included — text posts don't support images. Switch to 'Image/Video' to include it." | Reddit API `kind: self` does not accept media attachments |
| BR-3b | Post type = **Image/Video** AND Source URL field is filled: publish is blocked until Source URL is cleared or post type is switched to Link. Blocking error at bottom of Reddit Settings: "Image posts don't support links. Clear the Source URL or switch to 'Link' post type." | Reddit API `kind: image` does not support a URL alongside an image |
| BR-3c | Post type = **Link** AND body text written: body text excluded from Reddit submission; non-blocking info notice shown. Post type = **Image/Video** AND body text written: body text excluded; non-blocking warning shown. | Reddit API does not support body text on `kind: link` or `kind: image` posts |
| BR-4 | Flair is mandatory only when the target subreddit has `flair_required: true` (detected via `/r/{subreddit}/about`). For subreddits where flair is optional, the flair selector is shown but not required. | Flair must be mandatory when Reddit enforces it (or the post is rejected); optional flair improves post discoverability but should not block publishing |
| BR-5 | Multi-subreddit posts are staggered 30 minutes apart per subreddit | Avoids Reddit's spam detection which flags the same content posted to multiple subreddits simultaneously |
| BR-6 | Flair options must be re-fetched each time the subreddit changes in the composer | Flair is per-subreddit; stale flair from a previous subreddit selection causes API errors |
| BR-7 | Reddit access tokens are refreshed silently before each publish attempt if the token is within 5 minutes of expiry | Prevents publish failures from expired access tokens (access tokens expire every 1 hour) |
| BR-8 | Reddit account connection is web-only — not available in iOS or Android apps | Reddit OAuth requires a browser redirect; mobile apps cannot handle this flow |
| BR-9 | ContentStudio must apply for and receive Reddit Commercial API approval before launch | Reddit requires formal application for commercial API usage; without it, API access is rate-limited to the free tier which may not support production volume |
| BR-10 | Posts to Reddit cannot be edited after publishing via ContentStudio (Reddit allows text post body edits, but ContentStudio will not support this in v1) | Simplifies implementation; edit flow deferred to v2 |
| BR-11 | NSFW communities are not accessible via 3rd-party API; ContentStudio must not attempt to post to them | Reddit API restriction since July 2023; attempting to post to NSFW subreddits returns a permission error |

---

## 9. Open Questions

| Question | Options | Owner | Due Date | Decision |
|---|---|---|---|---|
| Has Reddit Commercial API approval been applied for? | Apply now / Already in progress / Not yet started | Product / Business Dev | 2026-04-10 | Pending — this is the #1 blocker |
| What is the stagger delay between multi-subreddit posts? | 15 min / 30 min / 60 min / user-configurable | Product | 2026-04-10 | Defaulting to 30 min; configurable in v2 |
| Should the subreddit selector show subreddits the user moderates at the top? | Yes (preferred) / No (flat search) | Product | 2026-04-10 | Pending — requires `GET /subreddits/mine/moderator` call |
| Is the Reddit icon already in `@contentstudio/ui` icon set? | In icon set / Needs to be added | Engineering / Design | 2026-04-10 | `icon-mapping.js` has `'reddit': '__brand__Reddit'` — confirm icon asset exists |
| Should ContentStudio warn users about subreddit rules (e.g., "No promotional posts")? | Yes — fetch subreddit rules / No — out of scope for v1 | Product | 2026-04-17 | Pending |
| What is the analytics sync frequency? | Every 1h / 6h / 24h | Engineering | 2026-04-17 | Suggesting 6h to balance API cost vs. freshness |

---

## 10. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Reddit Commercial API application is denied or takes longer than expected | Medium | High — entire feature is blocked without API access | Apply immediately; engage Reddit through a direct partnership contact if possible; build feature in parallel with approval process |
| Reddit API cost ($0.24/1,000 calls) exceeds budget at scale | Medium | Medium | Implement aggressive caching (subreddit search, flair lists cached per session); minimize polling frequency for analytics; monitor cost per workspace |
| Reddit rate limits (60–100 req/min) cause publish failures during peak batch publishing | Low | High | Queue Reddit publishing jobs with rate limit awareness; implement exponential backoff on `429` responses |
| Users' posts fail because subreddit has minimum karma requirement that their account doesn't meet | High (common on Reddit) | Medium | Surface subreddit requirements check before scheduling if possible; display clear error message post-failure; long-term: add karma health indicator in v2 |
| Reddit token refresh fails mid-publish (refresh token revoked by user) | Medium | Medium | Detect `401` on refresh → mark account expired → notify user via in-app + email notification → re-queue post for manual retry |
| Duplicate post detection (`ALREADY_SUB`) frustrates users cross-posting to multiple subreddits | Medium | Low | Stagger posting by 30 min per subreddit; document duplicate detection window in UI tooltip; show pre-publish warning for link posts |
| Reddit deprecates or changes the `/api/submit` endpoint (as happened with the v2 API migration in 2023) | Low | High | Monitor Reddit developer announcements; abstract the Reddit client into a single service class (`RedditConnector`) so changes are isolated |
| User connects a Reddit account with very low karma (new account) and all posts fail | High | Medium | Show account karma in Social Accounts table; add info tooltip: "Some subreddits require minimum karma to post. If posts fail, check the subreddit's rules." |

---

## 11. Dependencies

**Internal:**
- `app/Strategy/Integrations/Connector.php` — Add Reddit connector case
- `app/Strategy/Planner/Posting.php` — Add Reddit posting case
- `app/Libraries/Publish/Posting/SocialPosting.php` — Add Reddit publishing dispatcher
- `app/Models/Integrations/Platforms/SocialIntegrations.php` — Add Reddit-specific fillable fields
- `config/integrations.php` — Add Reddit OAuth config and social_channels entry
- `src/types/common/social-accounts.ts` — Add `RedditAccount` type and `'reddit'` to `ChannelType`
- Analytics dashboard — Reddit account card and post metrics
- Media upload pipeline — Reused for image posts (existing `sharing_media_details` flow)
- EasyConnect — Reddit included in the supported platforms list

**External:**
- **Reddit API v2** — Formal commercial API approval required. Endpoints used:
  - `POST /api/submit` — Create posts
  - `GET /api/v1/me` — Fetch user profile on connection
  - `GET /r/{subreddit}/api/link_flair_v2` — Fetch flair options
  - `GET /subreddits/search` — Subreddit autocomplete
  - `GET /api/info?id=t3_{id}` — Fetch post analytics
  - `POST /api/v1/access_token` — Token exchange and refresh
- **Reddit Developer Portal** — App registration required for `client_id` / `client_secret`
- **Reddit API Terms of Service** — Legal review required before launch

**Blockers:**
1. Reddit Commercial API approval (must apply before development starts; approval gate before launch)
2. Reddit Developer App registered with ContentStudio's OAuth callback URL
3. Reddit icon confirmed in `@contentstudio/ui` icon set

---

## 12. Appendix

- **Research & Competitor Analysis:** `docs/features/reddit-publishing/01-research.md`
- **Workflow Design:** `docs/features/reddit-publishing/02-workflow.md`
- **Telegram Integration Mockup (UI reference):** `docs/features/telegram-integration/telegram-integration-mockup.html`
- **Reddit API Reference:** https://www.reddit.com/dev/api/
- **VistaSocial Reddit Publishing (primary competitor reference):** https://support.vistasocial.com/hc/en-us/articles/4531883893659-Reddit-Publishing-with-Vista-Social
- **Backend architecture reference:** `contentstudio-backend/app/Strategy/Planner/BlueskyPosting.php` (newest posting class)
- **Frontend type reference:** `contentstudio-frontend/src/types/common/social-accounts.ts`

---

## Changelog

| Date | Author | Changes |
|---|---|---|
| 2026-04-03 | Product Pipeline | Initial draft |
