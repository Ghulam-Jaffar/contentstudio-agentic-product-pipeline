# Social Mentions — Workflow Design

## 1. Feature Placement

**Navigation:** Social Mentions lives as a top-level section in ContentStudio's left sidebar, alongside Planner, Inbox, and Analytics. It is an add-on feature ($49/month); non-subscribers see a locked landing page.

**Entry points:**
- Left sidebar nav item: "Social Mentions" (with mention count badge when new mentions arrive)
- Onboarding checklist (new workspaces): "Set up mention monitoring"
- Upgrade prompts from relevant areas (e.g., when a user searches for a brand name in Inbox)

**URL structure:**
```
/social-mentions                   → Landing page (non-subscriber) or Feed (subscriber)
/social-mentions/setup             → Onboarding wizard
/social-mentions/feed              → Main feed (default view: All Mentions)
/social-mentions/bookmarks         → Saved mentions
/social-mentions/analytics         → Mentions analytics
/social-mentions/alerts            → Alert rule management
/social-mentions/settings          → Company context + global keyword filters
```

---

## 2. Subscription States

| State | Experience |
|---|---|
| Non-subscriber | Landing page with value prop, demo, and $49/mo add-on CTA |
| Trial (7 days) | Full access, trial banner at top with days remaining |
| Active | Full access |
| Expired/Churned | Feed visible but locked; re-subscribe prompt |

---

## 3. Onboarding Flow (First-Time Setup)

Triggered when a subscriber first accesses Social Mentions (no topics configured yet).

```
Step 1: Enter Company Website (optional)
─────────────────────────────────────────
Prompt: "Enter your company website to get AI-suggested topics"
Skip link: "Skip, I'll add topics manually →"

If URL provided:
  → AI analyzes the site (10–20 seconds)
     Progress: "Detecting brand & competitors" → "Generating topic suggestions"
  → Proceeds to Step 2 with pre-filled suggestions

If skipped:
  → Proceeds to Step 2 with empty topic list (user adds manually)

Step 2: Review & Configure Topics
────────────────────────────────────
If AI suggested: 4–6 topic cards shown, each pre-filled
If skipped: one empty "Add a topic" prompt shown

Each topic card shows:
  - Topic name (editable text input, e.g., "ContentStudio Brand")
  - Type badge: Own Brand (blue) / Competitor (orange) / Industry Term (gray) / Custom (gray)
  - Keywords list within the topic (chips, each editable/deletable)
  - "Add keyword" inline button per topic
  - Platform pills (all selected by default, toggleable per topic)
  - AI context hint (collapsed, expandable, pre-filled if AI ran)
  - Delete topic button

User can:
  - Edit topic name
  - Add/remove keywords within a topic
  - Toggle platforms per topic
  - Add new topics manually ("+ Add another topic")
  - Delete any suggested topic
  - Expand advanced settings per topic (include/exclude terms, exact match)

CTA: "Start Monitoring →"
→ Redirects to /social-mentions/feed (All Mentions view)
→ System begins fetching mentions in the background
→ Toast: "We're fetching your first mentions. This may take a few minutes."
```

---

## 4. Adding Topics & Keywords (Post-Onboarding)

**Add Topic:**
Accessible from:
- "Add topic" button in the Views sidebar (feed page)
- Settings → Topics tab → "New Topic" button

**New Topic Modal:**
```
1. Topic name (text input, e.g., "Hootsuite", "social scheduling")
2. Topic type: Own Brand / Competitor / Industry Term / Custom
3. Keywords within this topic (tag input — add multiple)
4. Platforms (all checked by default, toggleable)
5. Optional: AI context hint (textarea, 200 char max)
6. Optional: Expand "Advanced filters":
   - Include ANY OF
   - Include ALL OF
   - Negative terms
   - Negative authors
   - Wildcard negative terms
   - Exact match toggle
   - Case sensitive toggle
7. "Save Topic"
```

System immediately starts fetching mentions for all keywords in the new topic (7-day lookback, up to 100 results per platform per keyword).

---

## 5. Feed — Happy Path

```
User lands on /social-mentions/feed
→ "All Mentions" view is active by default
→ Feed shows mention cards in reverse chronological order
→ Left sidebar shows: Topics list + Views list

User scans the feed:
→ Each card shows: platform icon, author, timestamp, keyword chip, content excerpt,
  AI tags, sentiment badge, engagement stats, action buttons

User switches views:
→ Clicks "Buy Intent" in sidebar
→ Feed filters to mentions tagged "Buy Intent"
→ Card count badge updates

User bookmarks a mention:
→ Clicks bookmark icon on a card
→ Card gets a bookmark indicator
→ Mention appears in /social-mentions/bookmarks

User opens original post:
→ Clicks "Open ↗" on a card
→ New tab opens to original post on the platform

User views mention details (optional side panel):
→ Clicks on the mention card body
→ Side panel slides in: full post text, author profile stats, thread context if available
```

---

## 6. Creating a Custom View

```
User clicks "+ Create View" in the sidebar
→ Modal opens: "New View"

Fields:
  - View name (text input)
  - Filters:
    - Keywords (multi-select from tracked keywords)
    - Platforms (multi-select)
    - Sentiment (positive / neutral / negative)
    - AI Tags (multi-select from 10 tags)
    - Language (multi-select)
  - Advanced: AND/OR logic toggle between filter groups

User clicks "Save View"
→ View appears in sidebar
→ Feed updates to show matching mentions
→ Toast: "View saved. You can set up an alert for this view."
```

---

## 7. Alert Setup Flow

**Entry points:**
- "Add Alert" button (top-right of feed page, applies to active view)
- Alerts page → "Create Alert"

```
Step 1: Choose View & Review Filters
──────────────────────────────────────
Dropdown: select which view to alert on
Filter summary displayed (read-only): keywords, platforms, sentiment, tags
User can edit filters if needed

Step 2: Alert Type
────────────────────
Select trigger type:
  ○ New Mentions — alert whenever new mentions match this view
  ○ Volume Spike — alert when mention count is X% above 7-day average
      └ Threshold input: "Alert when X% above average" (default: 50%)
  ○ Sentiment Shift — alert when negative sentiment crosses threshold
      └ Threshold input: "Alert when negative mentions exceed X%" (default: 30%)
  ○ First Mention — alert on the very first mention of a new keyword

Step 3: Destination & Frequency
─────────────────────────────────
Destination (multi-select):
  ☐ Email — recipient list (comma-separated)
  ☐ Slack — workspace connection required; channel selector
  ☐ Webhook — URL input + optional secret header

Frequency (for "New Mentions" type only):
  ○ Real-time (immediate)
  ○ Hourly digest
  ○ Daily digest
  ○ Weekly digest

CTA: "Create Alert"
→ Alert rule saved and active
→ Toast: "Alert created. You'll be notified when this view triggers."
```

---

## 8. Analytics Flow

```
User navigates to /social-mentions/analytics

Page layout: date range picker (top-right) + keyword filter (top-left)

Charts displayed:
  1. Total Mentions Over Time — line chart, filterable by topic + platform
  2. Topic Comparison — grouped bar chart (each topic as a series, side-by-side by day/week)
  3. Platform Breakdown — donut chart + table (mention count per platform, % of total)
  4. Sentiment Trend — stacked area chart (positive/neutral/negative over time)
  5. AI Tag Breakdown — horizontal bar (count per tag type)

Each chart has:
  - Download button (PNG/CSV)
  - AI Insights button (sparkle icon) → popover with 2–3 AI-generated observations
```

---

## 9. Settings Flow

```
/social-mentions/settings — tabbed layout:

Tab 1: Global Filters
  - Global negative terms (apply to all topics/keywords)
  - Global negative authors
  - Allowed subreddits (allowlist)
  - Excluded subreddits
  - Excluded GitHub repositories

Tab 2: Topics
  Table: all tracked topics with type, keyword count, platform count, mention count, status
  Actions per row: Edit (opens topic settings drawer), Pause/Resume, Delete
  Topic settings drawer:
    - Topic name + type
    - Keywords (add/remove chips)
    - Platforms (toggleable pills)
    - AI context hint
    - Include ANY OF / ALL OF
    - Negative terms, Negative authors
    - Wildcard negative terms
    - Exact match toggle, Case sensitive toggle
```

---

## 10. Non-Subscriber Landing Page

Small, focused landing page at `/social-mentions` for users without the add-on:

```
Hero:
  Headline: "Never miss a mention that matters"
  Subtext: "Track your brand, competitors, and industry keywords across 18 platforms — all in one feed."
  CTAs: [Start Free Trial]  [See Demo →]

3-column features:
  Feed-First    — See every mention in real time, organized by what matters to you
  AI-Powered    — Auto-tagged, sentiment-scored, and noise-filtered by AI
  Instant Alerts — Get notified the moment something important happens

Pricing:
  Single add-on card:
    Social Mentions Add-on
    $49 / month
    ✓ 18 platforms monitored
    ✓ Unlimited keywords
    ✓ AI tagging + sentiment
    ✓ Custom views & alerts
    ✓ Email, Slack & webhook alerts
    [Add to ContentStudio]

Demo preview:
  Interactive demo button → switches prototype state to show the feed
```

---

## 11. Key Design Decisions

### Decision 1: Topic-centric model
**Chosen: Topic-centric** — each Topic is a named container with one or more keywords. All keyword-level settings (platforms, include/exclude, exact match, context hint) live at the keyword level within the topic.

Example:
- Topic: "ContentStudio Brand" → keywords: ["ContentStudio", "CS dashboard", "@ContentStudio"]
- Topic: "Hootsuite" → keywords: ["Hootsuite", "@Hootsuite"]
- Topic: "Social Media Scheduling" → keywords: ["social media scheduler", "post scheduler"]

Views filter across all topics (or by selected topics).

### Decision 2: Views scope — global vs. per-keyword
**Options:**
- A) Views are global (filter across all keywords)
- B) Views are per-keyword

**Recommendation: A — Global views.**
Consistent with Octolens. A "Buy Intent" view should surface buy-intent mentions across all tracked keywords. Per-keyword filtering is already available through the keyword chip filter on each view.

### Decision 3: Mentions feed pagination vs. infinite scroll
**Options:**
- A) Infinite scroll (Twitter/Reddit style)
- B) Paginated (table style)

**Recommendation: A — Infinite scroll** with a "Back to top" floating button. Matches the expected feed UX for social content consumption.

### Decision 4: Analytics placement
**Options:**
- A) Separate `/analytics` route (current plan)
- B) Inline stats panel within the feed sidebar

**Recommendation: A — Separate route.** Analytics are comprehensive enough (6 charts) to deserve a dedicated page. The feed should stay focused on consumption, not analysis.

---

## 12. V1 Scope vs. V2 Deferral

### V1 (this feature)
- All 18 platform sources
- AI-powered onboarding (optional website URL → topic suggestions)
- Topics with multiple keywords, full per-keyword settings (Octolens parity + extras)
- Global filters (negative terms, authors, subreddit rules)
- Feed with 6 default Views + unlimited custom
- Bookmarks
- Alerts: 4 types × 3 destinations (email, Slack, webhook)
- Analytics: 5 chart types
- $49/month add-on with 7-day trial

### V2 (post-launch)
- Reply to mention directly from feed (using connected CS social account)
- Assign mention to teammate + internal notes
- AI weekly summary (brand + competitor brief)
- Bulk actions (mark read, dismiss, export selection)
- Mobile app notifications for alerts
