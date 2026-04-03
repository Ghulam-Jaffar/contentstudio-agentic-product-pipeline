# Reddit Publishing — Research & Competitor Analysis
*Feature Pipeline · Step 1 · Generated 2026-04-03*

---

## 1. What Is Reddit Publishing for Social Media Tools?

Reddit is a network of communities (subreddits) where users share text, links, images, and videos. For brands, it's a high-intent discovery channel — users come to Reddit specifically to ask questions, evaluate products, and seek peer recommendations. Unlike Instagram or Twitter, Reddit's culture rewards genuine community participation over promotional content, which makes scheduling and tone critical.

**Why brands use Reddit:**
- Direct access to niche, high-intent audiences (subreddits are self-selected communities)
- Organic reach that doesn't decay as fast as algorithmic feeds
- AMA (Ask Me Anything) sessions drive brand awareness
- Early adopter and tech-savvy demographics
- Long content lifespan — top posts stay visible for days/weeks

**Unique challenges vs. other platforms:**
- Reddit's API v2 (June 2023) introduced commercial API pricing ($0.24/1,000 calls) and requires a formal application/approval — not self-serve
- NSFW content is completely blocked for third-party API consumers since July 2023
- Subreddit rules vary wildly — what's allowed differs per community (no promotional posts, minimum karma requirements, mandatory flair, etc.)
- No personal profile posting via API — subreddit-only submission
- Duplicate post detection — Reddit actively blocks near-identical posts to the same subreddit within a short window
- Reddit's API does not support image + text together in a single post (image posts have a title and image, not body text)

---

## 2. VistaSocial Deep Dive

VistaSocial is currently the most feature-complete SMM tool with native Reddit publishing. This section documents everything they offer.

**Reference:** https://support.vistasocial.com/hc/en-us/articles/4531883893659-Reddit-Publishing-with-Vista-Social

### Connection Flow
- Standard **OAuth 2.0** — user is redirected to Reddit's authorization page, clicks "Allow", and their profile is added to VistaSocial
- Reddit counts as **one profile slot** in their account limits
- Multiple Reddit accounts can be connected to the same workspace

### Supported Post Types
| Post Type | Supported? | Notes |
|---|---|---|
| Text post | ✅ | Title + body text (Markdown) |
| Link post | ✅ | Title + URL; Reddit auto-generates thumbnail |
| Image post | ✅ | Title + single image upload |
| GIF post | ✅ | Treated as image |
| Video post | ❌ | Not documented; likely API limitations |
| Gallery (multi-image) | ❌ | Not documented |
| Poll | ❌ | Not supported |
| Crosspost | ❌ | Not supported |

### Subreddit Targeting
- **Required field** — Reddit API doesn't support personal profile posting
- Subreddit autocomplete search from the composer
- Shows subreddit name and subscriber count in dropdown
- Only one subreddit per post (no multi-subreddit scheduling in a single post — must create separate posts)

### Reddit-Specific Post Fields
- **Flair:** Dynamic dropdown loaded per subreddit from Reddit API
  - If a subreddit enforces mandatory flair, VS detects this and blocks publishing without a flair selection
  - Flair options refresh when subreddit is changed
- **NSFW toggle:** Not implemented (Reddit blocked 3rd-party NSFW access July 2023)
- **Spoiler toggle:** Not documented
- **OC (Original Content) toggle:** Not documented
- **Send Replies toggle:** Not documented

### Scheduling
- One-time scheduling
- Queue/slot-based scheduling (drag-and-drop)
- Best time recommendations (audience analytics-based, not per-subreddit)
- Multi-time scheduling per post (added 2025 — same post queued for multiple time slots)
- No recurring/evergreen scheduling documented for Reddit specifically

### Analytics
- Upvote score (karma)
- Number of comments
- Follower count on connected account
- Shares/awards
- Export: CSV + PDF
- AI-generated report summaries (added 2025 — natural language insights)

### Account Health / Warnings
- Shows "bot removed" style warnings when token is revoked
- Token expiry notifications
- Reconnect flow available from social accounts settings

### Pricing Tier
- Reddit available on **all paid plans** ($79/mo+)
- Free plan: not available
- Reddit profile counts toward the account slot limit (same as other platforms)

### UX Description
- In the composer, after selecting a Reddit account, a Reddit-specific section appears below the post body
- Subreddit selector: text input with autocomplete search
- Flair selector: dropdown appears after subreddit is selected, loads dynamically
- Post type is auto-detected based on whether media is attached (text vs. image vs. link)
- Calendar view shows Reddit posts alongside other platforms with Reddit's orange icon

---

## 3. Competitor Analysis Table

| Tool | Has Reddit? | Post Types | Scheduling | Reddit-Specific Fields | Analytics | Pricing Tier | Notes |
|---|---|---|---|---|---|---|---|
| **VistaSocial** | ✅ Native | Text, Link, Image, GIF | One-time, Queue, Best Time, Multi-time | Flair (dynamic, mandatory enforcement) | Karma, comments, followers, awards | All paid ($79+) | Most complete implementation; formal API partnership |
| **Sprout Social** | ✅ (limited) | Text, Link | One-time, Queue | None documented | Basic | Enterprise ($249+) | Via formal Reddit API partnership; limited features |
| **Buffer** | ❌ | — | — | — | — | — | No Reddit support |
| **Hootsuite** | ❌ (Ads only) | — | — | — | — | — | Added Reddit Ads Dec 2025; no organic publishing |
| **Later** | ❌ | — | — | — | — | — | No Reddit support |
| **Loomly** | ⚠️ Notification | — | Reminder only | — | — | — | Mobile notification → user manually posts |
| **SocialBee** | ⚠️ Notification | — | Reminder only | — | — | — | Mobile notification → user manually posts |
| **Sendible** | ❌ | — | — | — | — | — | No Reddit support |
| **Metricool** | ❌ | — | — | — | — | — | No Reddit support |
| **Agorapulse** | ❌ | — | — | — | — | — | No Reddit support |
| **Publer** | ❌ | — | — | — | — | — | No Reddit support |

**Reddit-specialist tools (for reference):**
| Tool | Focus | Unique Features |
|---|---|---|
| Postpone | Reddit-only | Per-subreddit best time, karma tracker, post history |
| Later for Reddit | Reddit-only | Community analytics, scheduling |
| Social Rise | Reddit-only | Bulk scheduling, account warm-up |
| Delay for Reddit | Reddit-only | Lightweight, simple scheduling |

---

## 4. Common Patterns (What Most Reddit-Supporting Tools Do)

1. **OAuth 2.0 connection** — Standard Reddit authorization flow, no bot tokens
2. **Subreddit as required field** — No personal profile posting (Reddit API limitation)
3. **Flair support** — Dynamic loading per subreddit, mandatory flair enforcement
4. **Text + Link + Image** — The three core post types supported universally
5. **One post = one subreddit** — Multi-subreddit requires multiple posts
6. **Calendar integration** — Reddit posts appear in the publishing calendar
7. **Token validity tracking** — Reddit tokens expire (1 hour access, 6-month refresh)

---

## 5. Differentiators Worth Noting

- **VistaSocial:** Mandatory flair enforcement (prevents failed posts) + multi-time scheduling
- **Postpone:** Per-subreddit best-time recommendations (more precise than audience-based)
- **Social Rise:** Account warm-up flows (gradual karma building to meet subreddit minimums)
- **Sprout Social:** Formal Reddit API partnership (presumably higher rate limits)
- **Loomly/SocialBee notification approach:** Zero API risk but terrible UX — just a workaround

---

## 6. Reddit API Constraints

| Constraint | Detail |
|---|---|
| **Commercial API approval** | Required — not self-service; Reddit reviews and may decline; requires formal application |
| **Pricing** | $0.24 per 1,000 API calls; enterprise requires custom contract |
| **Rate limits** | 60–100 requests/minute authenticated; 10/min unauthenticated |
| **NSFW content** | Blocked for third-party apps since July 2023; ~20% of communities inaccessible |
| **Personal profile posts** | Not supported via API — subreddit-only |
| **Image + text** | Reddit API does NOT support image posts with body text; image posts have title only |
| **Duplicate posts** | Reddit detects and blocks duplicate URL submissions to the same subreddit |
| **Token lifespan** | Access token: 1 hour; Refresh token: 6 months (revoked on re-authorization) |
| **Required OAuth scopes** | `identity`, `submit`, `flair`, `read`, `mysubreddits`, `history` |
| **Post title** | Required; 1–300 characters |
| **Post body** | Optional for text posts; 0–40,000 characters |
| **Subreddit rules** | Each subreddit enforces its own rules — tools can only handle flair/NSFW; content rules are manual |

---

## 7. User Expectations

### Table Stakes (Must-Have for Launch)
- Connect Reddit account via OAuth — same familiar flow as LinkedIn/Twitter
- Select subreddit per post with autocomplete
- Support text posts, link posts, and image posts
- Show Reddit-specific flair selector (per subreddit)
- One-time scheduling + calendar visibility
- Basic analytics: upvotes + comments
- Token expiry warnings + reconnect flow
- Repost duplicate detection

### Delighters (Premium / v2)
- Per-subreddit best-time recommendations (based on when that community is most active)
- Subreddit discovery — suggest relevant subreddits based on topic/keyword
- Account karma display — show if account meets minimum karma requirements for a subreddit
- Staggered multi-subreddit scheduling (same post, multiple subreddits, time-delayed to avoid spam flags)
- Reddit Smart Inbox — view and reply to comments on posts from within ContentStudio's unified inbox
- AI-generated Reddit-native titles/body that avoids marketing-speak
- Post performance benchmarking vs. subreddit average

---

## 8. Existing ContentStudio Features That Can Be Reused for Reddit

### From Codebase Analysis

**Architecture Pattern (Strategy/Factory):**
- The existing `Connector` factory + `ConnectionInterface` can be extended with `RedditConnector.php` following the same OAuth 2.0 pattern as LinkedIn/Threads
- The `Posting` factory + `PostingInterface` can be extended with `RedditPosting.php` following `BlueskyPosting.php` as the most recent reference
- `SocialRepo` with `platform_type: 'reddit'` filter works without any changes

**Backend Files to Create:**
| File | Action |
|---|---|
| `app/Strategy/Integrations/RedditConnector.php` | CREATE — OAuth 2.0 flow |
| `app/Strategy/Planner/RedditPosting.php` | CREATE — post dispatch |
| `app/Strategy/Integrations/Connector.php` | MODIFY — add Reddit case |
| `app/Strategy/Planner/Posting.php` | MODIFY — add Reddit case |
| `app/Libraries/Publish/Posting/SocialPosting.php` | MODIFY — add dispatcher |
| `app/Models/Integrations/Platforms/SocialIntegrations.php` | MODIFY — add `reddit_*` fields |
| `routes/web/integrations.php` | MODIFY — OAuth callback + addProfile routes |
| `config/integrations/social_integrations/reddit.php` | CREATE — env config |

**Frontend Files to Create/Modify:**
| File | Action |
|---|---|
| `src/types/common/social-accounts.ts` | MODIFY — add `RedditAccount` type + `'reddit'` to `ChannelType` union |
| `src/modules/integration/components/dialogs/AddReddit.vue` | CREATE — OAuth connection modal |
| `src/modules/integration/components/dialogs/SaveSocialAccounts.vue` | MODIFY — add Reddit conditional |
| `src/modules/publisher/.../SocialTab.vue` | MODIFY — add `reddit: []` platform init |
| `src/utils/icon-mapping.js` | Already has `'reddit': '__brand__Reddit'` ✅ — no change |

**Reusable ContentStudio Features for Reddit Posts:**
- Media upload pipeline (images/GIFs already handled in common_sharing_details.media)
- Link shortener (if enabled in workspace settings — useful for link posts)
- AI caption generation (can generate Reddit-native body text)
- Post scheduling and calendar (platform-agnostic — Reddit slots in automatically)
- First comment scheduling (could post a comment immediately after — useful for adding extra links)
- UTM parameter builder (for link posts)
- Content categories (can tag Reddit posts same as others)
- Bulk scheduling / CSV import
- Analytics dashboard (with Reddit-specific metrics added to existing charts)

**Mobile Apps:**
- Social account connection is web-only (confirmed — no OAuth in iOS/Android)
- Mobile apps read connected accounts from the backend API
- Reddit posts will be visible in mobile composer once backend supports it
- Mobile post scheduling will work for Reddit without mobile-specific changes

---

## 9. Recommended Scope for ContentStudio

### v1 — Minimum Viable Reddit Integration

| Feature | Priority | Rationale |
|---|---|---|
| OAuth connection (multi-account) | P0 | Foundation |
| Text post type (title + body) | P0 | Most common brand use case |
| Link post type (title + URL) | P0 | Core brand sharing |
| Image post type (title + image) | P0 | Visual content |
| Subreddit selector with autocomplete | P0 | Required by Reddit API |
| Dynamic flair loading per subreddit | P0 | Mandatory flair enforcement prevents post failures |
| One-time scheduling | P0 | Core publishing feature |
| Calendar visibility (Reddit posts) | P0 | Parity with other platforms |
| Token expiry warnings + reconnect flow | P0 | Account health |
| Basic analytics: upvotes + comments | P1 | Minimum reporting |
| Repost duplicate detection | P1 | Prevents spam flags |
| Social Accounts settings UI | P0 | Same pattern as Telegram mockup |

### v2 — Differentiators

| Feature | Priority | Rationale |
|---|---|---|
| Per-subreddit best-time recommendations | P1 | Best-in-class vs VistaSocial |
| Staggered multi-subreddit scheduling | P1 | Unique, high user value |
| Account karma display + subreddit eligibility check | P1 | Reduces friction/failed posts |
| Subreddit discovery (suggest based on topic) | P2 | Reduces research burden |
| Reddit Smart Inbox (replies + mentions) | P2 | Unified inbox extension |
| AI Reddit-native content generation | P2 | Avoids marketing-speak |
| GIF post support | P1 | Easy extension of image support |
| Video post support | P2 | More complex; Reddit video API |
| Gallery post (multi-image) | P2 | Reddit-specific API |
| Poll creation | P2 | Limited brand use case |

### Out of Scope (v1 + v2)
- NSFW toggle (Reddit API blocked for 3rd parties)
- Personal profile posting (Reddit API limitation)
- Crosspost (low brand use case)
- Comment-only scheduling (separate product area)

---

## 10. Competitive Positioning

ContentStudio entering Reddit publishing in 2026 has a **clear market opportunity**:
- Only VistaSocial and Sprout Social offer native Reddit publishing
- Buffer, Hootsuite, Later, Sendible, Metricool have no Reddit support
- ContentStudio can differentiate with: per-subreddit timing, subreddit discovery, karma health checks, and staggered posting — none of which VistaSocial currently offers

The biggest risk is Reddit's commercial API approval process. ContentStudio should apply for API access early in the development cycle.
