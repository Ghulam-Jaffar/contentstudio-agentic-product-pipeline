# Listening — Workflow Design

**Feature:** Listening (Social Listening add-on)
**Pipeline Step:** 2 of 5 — Workflow Design
**Date:** 2026-03-06
**Prototype:** `cs-prototypes/app/features/listening2/`
**Live prototype:** https://cs-prototypes.vercel.app/features/listening2

---

## 1. Feature Placement

### In the ContentStudio Navigation

Listening is a **top-level module** in the ContentStudio global top navigation bar, alongside Home, Publisher, Analytics, Inbox, Discover, and Library. It is visible to **all users regardless of plan** — clicking it always works; non-subscribers and trial users land on the upsell/gate page rather than getting an error or a disabled item.

```
[Global Top Bar — ContentStudio]
  CS  My Workspace  |  Home  Publisher  Analytics  Inbox  [Listening]  Discover  Library  |  🔔  ⚙  JD
```

### Module-Level Navigation (Secondary Nav — Left Panel)

Inside Listening, a **dedicated left sidebar (PrimaryNav, 200px)** replaces the ContentStudio-wide sidebar. It contains:

```
[Listening logo + wordmark]
────────────────────────
  Feed
  Bookmarks
  Analytics
  Alerts
  Settings
────────────────────────
  TOPICS
  ● ContentStudio        ···
  ● Hootsuite            ···
  ○ Buffer (paused)      ···
  [+ Add topic]
────────────────────────
  [Usage bar — active users only]
  Topics:   X / 5   ▓▓▓▓▓░░░░░
  Mentions: X.Xk / 10k   ▓▓▓▓░░░░░
  Need more? Add-ons available →
```

### Feed Sub-Navigation (Views Sidebar — 210px)

When "Feed" is the active section, a second-level panel (ViewsSidebar) appears to the right of PrimaryNav:

```
[FEED]                          [×]
─────────────────────────────────
  FOR YOU
  ✦ All Mentions         5,272
  ✦ High Relevance         312

  VIEWS                         [+]
  ⛨ Crisis Management          38
  ⛨ Brand Monitoring          847
  ✕ Competitor Intel         2,135
  ♡ Brand Love                 201
  🛒 Buy Intent                 89
  [+ Add view]
```

This panel is collapsible (chevron button). On mobile it renders as a Drawer overlay.

---

## 2. User States & Entry Flows

Listening has five distinct user states, each producing a different experience on first visit.

### State A: Trial Plan

**Trigger:** `userState = 'trial'`

```
User clicks "Listening" in top nav
  └→ LandingPage renders (trial variant)
       ├→ Badge: "Add-on · $49/mo"
       ├→ Headline: "Never miss a mention that matters"
       ├→ Trial gate banner (orange):
       │    "Not available on trial plans"
       │    "Listening is a paid add-on. Upgrade to any ContentStudio plan to unlock it."
       │    [Upgrade Plan] →
       ├→ Feature highlights (3 cards): Feed-First · AI-Powered · Instant Alerts
       └→ Pricing card: $49/mo — feature list — [Upgrade Plan First] (orange, disabled)
```

**All CTAs lead to the plan upgrade flow.** No path into Listening itself.

---

### State B: Paid Plan, Not Subscribed

**Trigger:** `userState = 'locked'`

```
User clicks "Listening" in top nav
  └→ LandingPage renders (locked variant)
       ├→ Badge: "Add-on · $49/mo"
       ├→ Headline: "Never miss a mention that matters"
       ├→ No trial banner
       ├→ CTAs:
       │    [Enable Listening]   → setUserState('unlocked_new') → Setup Flow
       │    [Preview Demo]       → setUserState('unlocked')     → Main Feed (demo data)
       └→ Pricing card: $49/mo — [Add to ContentStudio]
```

---

### State C: Expired Subscription

**Trigger:** `userState = 'expired'`

```
User clicks "Listening" in top nav
  └→ LandingPage renders (expired variant)
       ├→ Badge: "Add-on Expired" (orange)
       ├→ Headline: "Re-enable Listening"
       ├→ Subtitle: "Your Listening add-on has expired. Re-enable it to resume monitoring
       │            your brand, competitors, and industry keywords across 18 platforms."
       └→ CTAs:
            [Re-enable Listening]  → setUserState('unlocked_new') → Setup Flow
            [Preview Demo]         → setUserState('unlocked')     → Main Feed (demo data)
```

---

### State D: First-Time Setup (New Subscriber)

**Trigger:** `userState = 'unlocked_new'`

Full onboarding wizard before entering the main product:

```
Step 1 — Website Input
  ┌─────────────────────────────────────────────────────────┐
  │  🌐                                                      │
  │  Set up your monitoring                                  │
  │  Enter your website and we'll detect your brand,         │
  │  competitors, and suggest topics automatically.          │
  │                                                          │
  │  [🌐 https://yourcompany.com                         ]   │
  │                                                          │
  │  [Analyze Website →]                                     │
  │  [Skip, I'll add topics manually →]                      │
  └─────────────────────────────────────────────────────────┘

  If URL entered → click "Analyze Website":
    Phase 1 (0.9s): "Detecting brand & competitors..."  (spinner)
    Phase 2 (1.1s): "Generating topic suggestions..."   (spinner)
    → auto-transition to Step 2 with AI-suggested topics pre-loaded
    → banner: ✦ "Our AI suggested these based on your website"

  If no URL → click "Analyze Website" OR "Skip":
    → jump to Step 2 with empty topics list
    → immediately opens TopicFormModal

Step 2 — Topics Review
  ┌─────────────────────────────────────────────────────────┐
  │  Your topics                                             │
  │  ✦ Our AI suggested these based on your website          │
  │                                                          │
  │  ●── ContentStudio  [Own Brand ▾]  ✏  🗑               │
  │       contentstudio  @contentstudio  contentstudio.io   │
  │                                                          │
  │  ●── Hootsuite  [Competitor ▾]  ✏  🗑                  │
  │       Hootsuite  @Hootsuite                              │
  │                                                          │
  │  [+ Add another topic]  (dashed button)                  │
  │                                                          │
  │  [Start Monitoring →]   (disabled if no topics)          │
  └─────────────────────────────────────────────────────────┘

  Each topic card shows:
    - Color bar (left border, type-colored)
    - Topic name + keyword pills (up to 4, +N more)
    - Type selector dropdown (Own Brand / Competitor / Industry Term / custom)
    - Edit ✏ → opens TopicFormModal pre-filled
    - Delete 🗑 → removes immediately

  "Add another topic" → opens TopicFormModal (empty)
  "Start Monitoring" → setUserState('unlocked') → Main Feed
```

---

### State E: Active Subscriber

**Trigger:** `userState = 'unlocked'`

User lands directly in the Feed with full navigation available. This is the primary ongoing experience.

---

## 3. Main User Flow (Active State)

### 3.1 Feed

```
Enter Listening → Feed section
  ├→ ViewsSidebar (left): select a view (All Mentions, High Relevance, or custom)
  └→ FeedShell (right):
       ├→ Filter bar: Topic | Sentiment | Tags | Min Followers | Date Range | Sort | Mark All Read
       └→ Mention card list (filtered + sorted)

Per mention card:
  ├→ Read content (3-line preview)
  ├→ See: author · platform · time · topic · sentiment · tags · engagement
  ├→ Hover → floating action pill appears (top-right of card):
  │    Reply · Bookmark · Open original · Mark as read · Mark irrelevant · Copy link
  │
  ├→ [Reply] (social platforms only)
  │    └→ Inline reply panel expands below card content:
  │         ├→ Account selector (connected accounts for that platform)
  │         │    ├→ No accounts: orange banner "No [Platform] accounts connected. Connect an account →"
  │         │    └→ Expired account: warning icon + tooltip on select; error on send attempt
  │         ├→ Textarea: "Write a reply..."
  │         ├→ AI toolbar:
  │         │    [✦ Write with AI] → generates context-aware draft (1.2s)
  │         │    (if text exists) [Rephrase] [Improve] [Shorten] [Lengthen] [Grammar Fix]
  │         └→ [Cancel] [Reply] (disabled if no text or no accounts)
  │              └→ Success: button turns green "Sent!" → panel collapses after 1.2s
  │
  ├→ [Bookmark] → saves to Bookmarks section (icon fills solid)
  ├→ [Open original] → opens platform URL in new tab
  ├→ [Mark as read] → removes unread indicator
  ├→ [Mark as irrelevant] → hides from feed
  ├→ [Copy link] → copies URL to clipboard
  └→ Sentiment badge (clickable) → dropdown: Positive / Neutral / Negative (override AI)
```

**Filter bar interactions:**
- Topic chip → multi-select popover (all topics with color dots)
- Sentiment chip → multi-select popover: Positive / Neutral / Negative
- Tags chip → multi-select popover (all 10 AI tags with color dots)
- Min Followers chip → single-select: Any reach / 500+ / 1k+ / 5k+ / 10k+ / 50k+ / 100k+
- Date range chip → presets (All time / 24h / 7d / 30d / 90d) + custom date picker
- Sort → "Newest" default (single-select)
- Mark all read → appears only when unread mentions exist; marks all as read

**Empty states:**
- No mentions (no data): "No mentions yet"
- No matches (filters active): "No mentions match your filters"

---

### 3.2 Views Management

```
ViewsSidebar → "FOR YOU" section (system views, read-only):
  All Mentions → shows everything unfiltered
  High Relevance → shows Buy Intent + Own Brand Mention tagged mentions

ViewsSidebar → "VIEWS" section (user-created):
  Click view → activates that view's filter combination in the feed
  Hover view → "···" context menu:
    Duplicate → creates copy with "(copy)" suffix
    Manage alerts → opens CreateAlertModal pre-filled with this view
    Delete → confirmation modal: "Are you sure? This cannot be undone."

[+ Add view] button → opens CreateViewModal:
  ├→ View name (required) + Icon picker (7 options)
  ├→ Filter by Topics (optional, multi-select)
  ├→ Filter by Platforms (optional, 18 platforms)
  ├→ Language (optional, 16 languages)
  ├→ Filter by Sentiment (optional, pill selector)
  ├→ Filter by Tags (optional, 10 AI tags)
  ├→ Author Reach (optional, min follower threshold)
  └→ [Save View] → added to sidebar, auto-selected

Collapse/expand:
  Active view shows PanelLeftOpen icon → click → sidebar collapses
  When collapsed, Feed icon in PrimaryNav shows expand hint → click → sidebar opens
```

---

### 3.3 Topic Management

**Add topic from PrimaryNav:**
```
Topics section header [+] button → opens TopicFormModal (empty)
  OR
PrimaryNav → topic "···" menu → Edit → opens TopicFormModal (pre-filled)
```

**TopicFormModal fields:**
```
[●] [Topic Name ________________] [Own Brand ▾]
    Color picker (8 swatches) on dot click

Keywords*    [tag input — type + Enter]
Platforms*   [dropdown — all 18, "Select all" shortcut]
AI Context Hint (optional, 200 chars)
  "e.g. ContentStudio is a social media tool, not a photo studio."

▼ Advanced Settings
  Include ANY of these terms  [tag input]
  Include ALL of these terms  [tag input]
  Negative terms              [tag input]
  Negative authors            [tag input]
  [○ Exact match]  [○ Case sensitive]

[Global filters also apply]  (read-only summary, link to Settings)

                              [Cancel] [Save Topic]
```

**Validation (on save attempt):**
- Name empty → "Topic name is required"
- No keywords → "Add at least one keyword"
- Keyword conflicts (3 types detected, accordion auto-opens):
  - Keyword in Negative Terms → "X is in both Keywords and Negative Terms — these would match nothing"
  - Include term in Negative Terms → "X appear in both Include Terms and Negative Terms — they cancel each other out"
  - Term in both Include ANY and Include ALL → "X appear in both Include ANY and Include ALL — Include ALL already implies Include ANY"

**Topic type system:**
- 3 built-in types: Own Brand / Competitor / Industry Term
- Custom types: can be created inline in the type selector ("+ New type")
- Custom types also managed in Settings > Topic Types

**Pause / Resume topic:**
```
PrimaryNav → topic "···" → Pause →
  Confirmation modal: "This will pause tracking for [name] and we'll stop
  looking for new mentions. Existing mentions will remain in your feed."
  [Cancel] [Pause keyword]

Paused topics:
  - Show gray dot in sidebar
  - Show PauseCircle icon
  - Topic "···" menu shows "Resume" instead of "Pause"
```

**Delete topic:**
```
PrimaryNav → topic "···" → Delete → immediate deletion (no confirmation)
  OR
SetupFlow / Settings → delete button → immediate
```

---

### 3.4 Bookmarks

```
PrimaryNav → Bookmarks
  ├→ Filter bar: Search (text) | Platform | Sentiment | Tags | Date Range
  └→ Mention card list (same MentionCard component, bookmarked mentions only)

Filtering: AND logic across all active filters
  Text search matches: mention content OR author name

Empty states:
  - No bookmarks: "No bookmarks yet" → explanation text
  - No matches: "No bookmarks match your filters" → "Try adjusting your filters or clearing the search."
```

---

### 3.5 Analytics

```
PrimaryNav → Analytics
  ├→ Filter bar: Topic filter (multi-select) | Date range chip | [Export] button
  │
  ├→ KPI cards (4):
  │    Total Mentions | Positive Sentiment % | Avg. Daily Mentions | Topics Tracked
  │
  ├→ Chart 1: Mentions Over Time (line chart, one line per topic, color-coded)
  │    [✦ AI Insights] button → 3 bullet points in popover
  │
  ├→ Chart 2: Sentiment Trend (stacked area chart, positive/neutral/negative by day)
  │    [✦ AI Insights] button
  │
  ├→ Side by side (stacked on mobile):
  │    Chart 3: Sentiment Distribution (donut, % positive in center + legend)
  │             [✦ AI Insights] button
  │    Chart 4: By Platform (donut + ranked list of platforms)
  │             [✦ AI Insights] button
  │
  └→ Chart 5: By Tag (horizontal bar chart, all 10 AI tags)
               [✦ AI Insights] button

Date filter: All time / Last 24h / Last 7 days / Last 30 days / Last 90 days / Custom range
Topic filter: multi-select; empty = all topics shown

AI Insights (per chart): 3 pre-generated bullet points contextualizing the chart data.
  Example (Volume chart): "Mention volume peaked 3.5× above average on day 15 — driven by
  the Hootsuite pricing announcement going viral."
```

---

### 3.6 Export

```
Analytics → [Export] button → ExportModal (3 tabs)

Tab 1: Download
  Format: [PDF Report] [CSV Data]  (segmented control)
  Date range: [date chip with presets + custom]
  [Generate & Download PDF/CSV]
  Success: green alert "Report ready! Your download should start automatically."

Tab 2: Schedule
  Frequency: [Weekly] [Monthly]
  Weekly → Send on: [day of week dropdown]
  Monthly → Send on day: [1st–28th dropdown]
  Time: [time selector, 06:00–18:00]
  Recipients: team member picker + email chip input
  [Save Schedule]  (disabled until ≥1 recipient)
  Success: "Schedule saved! Reports will be sent automatically."

Tab 3: Share
  Info: "Generate a read-only link. Anyone with the link can view analytics,
        but cannot make any changes."
  [Generate Share Link] → shows readonly URL + copy button
  Copy success: "Link copied to clipboard!"
  Footer: "Share links expire after 30 days."
```

---

### 3.7 Alerts

```
PrimaryNav → Alerts
  ├→ Header: "Alerts" + "Get notified when something unusual happens in your feed"
  ├→ [New Alert] button
  ├→ Active alerts section
  ├→ Paused alerts section
  └→ Empty state: "No alerts configured" → explanation + "Create your first alert"

Per alert card:
  - View name (which view this monitors)
  - Trigger pills: Volume Spike (amber) and/or Sentiment Spike (red) with threshold detail
  - Email recipients (avatars + names)
  - Toggle switch: active/paused
  - "···" menu: Edit / Delete

[New Alert] → CreateAlertModal:
  ├→ Monitor this view (select — all views)
  ├→ Alert triggers:
  │    Volume Spike card: "Unusual surge in mention volume"
  │      Toggle ON → "Notify when mentions are [50%] above 7-day average"
  │    Sentiment Spike card: "Rise in negative sentiment"
  │      Toggle ON → "Notify when negative mentions exceed [30%] of total"
  ├→ Email recipients: team picker + manual email input
  │    "Press Enter or comma to add multiple addresses"
  └→ [Save alert] (disabled until ≥1 trigger active AND ≥1 recipient)
```

---

### 3.8 Settings

```
PrimaryNav → Settings
  Two tabs: Global Filters | Topic Types

Tab: Global Filters
  ├→ Negative Terms
  │    "Mentions containing these terms will be excluded from all topics."
  │    Tag input: global keywords to always exclude (e.g. spam, bot, advertisement)
  │    Hint: "These apply across all topics. Per-topic negative terms can be set in the Topics tab."
  │
  ├→ Blocked Authors
  │    "Mentions from these authors will never appear in your feed, regardless of topic."
  │    Tag input: @handle or u/username
  │
  ├→ Reddit Settings
  │    "Control which subreddits to exclude from monitoring."
  │    Tag input: r/subredditname
  │
  └→ Language Filter
       "Only collect mentions in the selected languages. Leave empty to monitor all."
       MultiSelect: 9 languages (English default)

  [Save Global Filters] button

Tab: Topic Types
  ├→ Built-in types (read-only): Own Brand / Competitor / Industry Term — "built-in" badge
  └→ Custom types: editable/deletable + [Add Type] button
       Edit: inline TextInput + Save / Cancel
       Delete: immediate
       Create: inline form — "Type name..." → [Create] / [Cancel]
```

---

### 3.9 Usage Bar (PrimaryNav)

Shown at the bottom of PrimaryNav only when `userState = 'unlocked'`:

```
Topics      X / 5    [████░░]   ← red when at limit (X = 5)
Mentions    X.Xk / 10k  [████░]   ← amber when ≥ 90%, red when at 100%
Need more? Add-ons available →
```

---

## 4. Alternative Flows & Edge Cases

### 4.1 Reply — No Connected Accounts
```
User clicks Reply on a mention from a platform with no connected accounts
  └→ Reply panel opens
       └→ Orange banner: "No [Platform] accounts connected. Connect an account →"
            Reply button remains disabled
```

### 4.2 Reply — Expired Account Token
```
User selects an expired account from the account dropdown
  (expired accounts shown with ⚠ icon + tooltip "Access token expired. Reconnect to reply.")
  └→ User clicks Reply anyway
       └→ Error message (red): "This account's access token has expired. Please reconnect to reply."
            Reply is not sent
```

### 4.3 Reply — Non-Supported Platform
```
Mention is from Hacker News, GitHub, DEV.to, Stack Overflow, Podcasts, Newsletters, News, or Blogs
  └→ Reply icon shown at 35% opacity
       Tooltip: "Replies not available on [Platform]"
       Clicking has no effect
```

### 4.4 Keyword Conflict in Topic Form
```
User adds keyword "apple" to both Keywords and Negative Terms
  └→ Clicking Save Topic:
       1. Advanced Settings accordion auto-opens
       2. Red error banner inside accordion: "apple" is in both Keywords and Negative Terms —
          these would match nothing"
       3. Save is blocked until conflict resolved

  Three conflict types (all detected simultaneously):
    KW in Negative Terms
    Include term in Negative Terms
    Term in both Include ANY and Include ALL
```

### 4.5 Delete View with Active Selection
```
User deletes the currently active view
  └→ Confirmation modal → Delete
       → activeViewId falls back to next available view OR 'view-relevance' (High Relevance)
```

### 4.6 Pause Topic — Existing Mentions
```
User pauses a topic
  └→ Confirmation: "we'll stop looking for new mentions. Existing mentions will remain in your feed."
       → Topic goes gray in sidebar + PauseCircle icon
       → Monitoring stops; feed retains historical mentions for that topic
```

### 4.7 Topics at Limit (5/5)
```
Usage bar turns red (topics bar)
User can still click [+] and create a topic (prototype) — in production this would:
  a) Block creation with upsell modal, OR
  b) Allow but charge for add-on topic
```

### 4.8 Mentions Near Limit (≥ 9,000 / 10,000)
```
Mentions bar turns amber at 90% usage
"Need more? Add-ons available →" link in usage bar
In production: at 100% → new mentions stop being ingested; user notified
```

### 4.9 Sentiment Override
```
AI assigns sentiment to a mention
User disagrees → clicks sentiment badge on card → dropdown: Positive / Neutral / Negative
User selects different value → persists for the session
(In production: persists to DB and re-weights AI training)
```

### 4.10 Mobile Navigation
```
Mobile (< 768px):
  PrimaryNav: hidden by default → hamburger opens as full Drawer (left)
  ViewsSidebar: "Views" button in filter bar → opens as Drawer (left)
  Filter bars: horizontally scrollable
  Analytics side-by-side charts → stacked vertically
  Topic cards in Setup: wrap to new row
  CTAs: stacked vertically (column instead of row)
```

---

## 5. Key Design Decisions

### Decision 1: Feed-First vs. Topic-First Architecture

**Previous v1 approach:** Topic-centric — user picks a topic → sees tabs (Analytics, Mentions, Compare, Reports, Settings) within that topic.

**Chosen approach (v2):** Feed-first — global mention stream across all topics, organized by views/filters. Topics are a configuration layer, not a navigation destination.

**Rationale:** Brand managers check Listening like they check email — they want to see what happened across all their tracking in one place, not navigate topic by topic. The views system lets power users segment without sacrificing the "one inbox" mental model.

### Decision 2: Views as Saved Filters (Not Separate Inboxes)

Views are saved filter combinations, not separate data stores. All views draw from the same global mention pool. This means:
- A mention can appear in multiple views simultaneously
- Counts update automatically as new mentions arrive
- No sync issues between views

### Decision 3: Limits-Based Pricing ($49/mo + Add-ons)

**Chosen:** Flat $49/mo add-on with 5 topics + 10,000 mentions/month; +topics and +mentions available as add-ons.

**Rationale:** Transparent, predictable, easy for users to understand. Avoids per-day/per-mention confusion seen with competitors (Metricool). Per-topic is the clearest usage dimension in the market. Usage bars in the sidebar make limits visible before they become a surprise.

### Decision 4: Inline Reply vs. Handoff to Inbox

**Chosen:** Inline reply directly from the mention card (with AI compose tools).

**Alternative considered:** Open the full Inbox thread.

**Rationale:** For social listening the engagement action is high-frequency and lightweight — a quick "thanks for the mention!" or a bug report acknowledgement. The full Inbox is for managed, threaded conversations. Inline reply is faster. Power users who want full thread context can click "Open original" to go to the platform.

### Decision 5: Smart Tagging Labels (Not "AI Tagging")

**Chosen:** "Smart tagging" in all user-visible copy. "AI tags" / `aiTags` remains in code and data models.

**Rationale:** "AI tagging" sounds technical and slightly off-putting. "Smart tagging" communicates benefit, not mechanism. The 10 tags (Own Brand Mention, Competitor Mention, Buy Intent, Bug Report, etc.) are descriptive enough that users understand the feature without needing to know it's AI-driven.

### Decision 6: Global vs. Per-Topic Filters in Settings

**Two-level filter hierarchy:**
1. **Global filters** (Settings > Global Filters): Apply to all topics — negative terms, blocked authors, excluded subreddits, language. Set once, applies everywhere.
2. **Per-topic filters** (TopicFormModal > Advanced): Apply only to that topic — topic-level negative terms, include ANY/ALL, negative authors, exact match, case sensitive.

Global filters are always shown as a read-only summary inside TopicFormModal so users understand what's already filtering their results without needing to leave the modal.

---

## 6. Integration with Existing ContentStudio Features

### 6.1 Inbox
- Reply to mention → sends via connected account (same OAuth tokens as the Inbox)
- Future: "Open in Inbox" link to create a full thread from a mention

### 6.2 Connected Accounts
- Account selector in reply panel uses existing connected accounts from Settings > Accounts
- Expired token detection (same `isExpired` flag as Publisher/Inbox)

### 6.3 Analytics
- Listening Analytics is a standalone tab within the Listening module, not the main ContentStudio Analytics
- Future: Listening mentions volume surfaced as a widget in the main Analytics dashboard

### 6.4 Team Members
- Alert email recipients drawn from workspace team member list (same directory used by approvals/comments)
- Export recipients same list

### 6.5 Settings / Billing
- "Enable Listening" / "Re-enable Listening" CTAs link to billing flow
- "Add-ons available →" link in usage bar links to billing/add-ons
- "Cancel anytime from your billing settings." links to billing

---

## 7. Scope — V1 vs. Deferred

### In V1 (as prototyped)

| Area | What's included |
|---|---|
| **Topics** | Create / edit / pause / resume / delete. Built-in types (Own Brand, Competitor, Industry). Custom topic types. Per-topic keyword configuration with advanced filters. |
| **Keywords** | Basic keyword + context hint + include ANY/ALL + negative terms/authors + exact match + case sensitive |
| **Platforms** | All 18 monitored: twitter, instagram, facebook, linkedin, tiktok, youtube, reddit, bluesky, pinterest, threads, hackernews, github, devto, stackoverflow, podcasts, newsletters, news, blogs |
| **Feed** | Global mention stream. Views (saved filter combinations — system + user-created). Feed filters (topic, sentiment, tags, followers, date, sort). Mention cards with keyword highlighting. |
| **Mention actions** | Reply (with AI compose + improve), Bookmark, Open original, Mark read, Mark irrelevant, Copy link, Sentiment override |
| **Bookmarks** | Saved mentions with search + filters |
| **Analytics** | 4 KPI cards. 5 charts: Mentions Over Time (line), Sentiment Trend (stacked area), Sentiment Distribution (donut), By Platform (donut + list), By Tag (bar). AI Insights per chart. Date range + topic filters. |
| **Export** | Download (PDF/CSV), Schedule (weekly/monthly, email recipients), Share (read-only link, 30-day expiry) |
| **Alerts** | Volume spike (% above 7-day average) and sentiment spike (% negative) per view. Email recipients. Active/paused toggle. CRUD. |
| **Settings** | Global negative terms, blocked authors, excluded subreddits, language filter. Topic type CRUD. |
| **Pricing** | $49/month add-on. 5 topics / 10,000 mentions limits. Usage bars with warnings. "Need more? Add-ons available →" |
| **Mobile** | Fully responsive: drawer navs, scrollable filter bars, stacked layouts |
| **Onboarding** | Website-based AI topic suggestion wizard → topics review → Start Monitoring |
| **User states** | Trial gate, Locked (paid, no addon), Expired, Setup flow, Active |

### Deferred to V2+

| Feature | Rationale |
|---|---|
| **Share of Voice** | Requires calculating brand % of total industry conversation — complex aggregation. High value, phase 2. |
| **Predictive trend forecasting** | ML model dependency. Phase 2+. |
| **Influencer discovery** | Ranking by reach requires enriched author data. Phase 2. |
| **Historical data tiers** (30d / 90d / 1yr) | Requires tiered plan logic. V1 = flat 10k/month, no explicit date history limit shown. |
| **Webhook alerts** | Landing page mentions "email or webhook". Email-only in V1; webhook requires infrastructure work. |
| **Competitor benchmarking / SOV charts** | Compare tab from v1 architecture. Deferred. |
| **AI Trend Summary digest** | Weekly narrative email ("340 mentions, 72% positive..."). Phase 2. |
| **Saved dashboards per brand/client** | Agency multi-brand workspace feature. Phase 2. |
| **Spam/bot auto-detection** | Global filters handle manual blocking; automated detection needs ML. Phase 2. |
| **Revoke share links from Settings UI** | Basic share works; revocation management deferred. |
| **In-app spike notification** | Alert is email-only in V1; in-app notification tray integration deferred. |
| **"Open in Inbox" from mention card** | Creates a full Inbox thread. Deferred pending Inbox integration. |
| **Mobile app (iOS/Android)** | Web-only for V1. Listening is a web-only add-on; mobile app integration is V2. |
