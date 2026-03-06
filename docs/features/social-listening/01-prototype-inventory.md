# Listening — Prototype Inventory

**Feature:** Listening (Social Listening add-on)
**Pipeline Step:** 1 of 5 — Research
**Date:** 2026-03-06
**Prototype:** `cs-prototypes/app/features/listening2/`

---

## Prototype Inventory

*Complete documentation of `cs-prototypes/app/features/listening2/` — all data models, UI copy, user flows, edge cases, tooltips, and business logic as prototyped.*

---

### 1. Data Models (`lib/types.ts`)

#### `UserState`
Controls which view the user sees. Values:
- `'trial'` — User is on a trial plan; Listening is not available. Shows landing page with trial gate banner.
- `'locked'` — User is on a paid plan but has not enabled Listening. Shows landing page with "Enable Listening" CTA.
- `'unlocked_new'` — User just purchased/enabled. Shows the Setup Flow (onboarding wizard).
- `'unlocked'` — Feature is active and onboarding complete. Shows the full Feed interface.
- `'expired'` — Add-on subscription has expired. Shows landing page with "Re-enable Listening" CTA.

#### `FeedSection`
Navigation sections: `'feed' | 'bookmarks' | 'analytics' | 'alerts' | 'settings'`

#### `Platform` (18 platforms)
`twitter | instagram | facebook | linkedin | tiktok | youtube | reddit | bluesky | pinterest | threads | hackernews | github | devto | stackoverflow | podcasts | newsletters | news | blogs`

- Social platforms (reply-supported): `twitter, instagram, facebook, linkedin, tiktok, youtube, reddit, bluesky, threads`
- All 18 are monitored; reply is only supported on social platforms

#### `Sentiment`
`'positive' | 'neutral' | 'negative'`

Colors: Positive `#10B981`, Neutral `#6B7280`, Negative `#EF4444`

#### `AiTag` (10 smart tags)
`'Own Brand Mention' | 'Competitor Mention' | 'Industry Insight' | 'Buy Intent' | 'Bug Report' | 'User Feedback' | 'Promotional Post' | 'Product Question' | 'Event' | 'Hiring'`

Each tag has a dedicated color (see `AI_TAG_COLORS` in constants.ts).

#### `Keyword`
```typescript
interface Keyword {
  id: string;
  value: string;              // the keyword text
  contextHint?: string;       // AI disambiguation hint ("ContentStudio is a social media tool, not a photography studio")
  includeAny?: string[];      // mention must also contain at least one of these
  includeAll?: string[];      // mention must contain all of these
  negativeTerms?: string[];   // exclude mentions containing these
  negativeAuthors?: string[]; // exclude all posts from these accounts
  exactMatch?: boolean;       // match exact phrase only
  caseSensitive?: boolean;    // case-sensitive matching
}
```

#### `Topic`
```typescript
interface Topic {
  id: string;
  name: string;
  type: TopicType;           // 'own_brand' | 'competitor' | 'industry' | custom type id
  color?: string;            // per-topic color override; falls back to type color
  keywords: Keyword[];
  platforms: Platform[];
  isActive: boolean;         // active vs. paused
  mentionCount: number;
}
```

**Built-in Topic Types:** Own Brand, Competitor, Industry Term
**Custom Topic Types:** Users can create their own (e.g., "Influencer", "Partner") from Settings > Topic Types

**Topic Type Colors:**
- Own Brand: `#2563EB`
- Competitor: `#EA580C`
- Industry Term: `#6B7280`
- Custom: `#7C3AED` (default, can be overridden)

#### `MentionView`
```typescript
interface MentionView {
  id: string;
  name: string;
  isDefault: boolean;        // true = system "For you" view, false = user-created
  icon: string;              // icon name from predefined set
  filters: {
    topicIds?: string[];
    platforms?: Platform[];
    sentiments?: Sentiment[];
    aiTags?: AiTag[];
    languages?: string[];    // ISO 639-1 codes
    minFollowers?: number;
    dateRange?: string;      // 'all' | '1d' | '7d' | '30d' | '90d' | 'custom:YYYY-MM-DD:YYYY-MM-DD'
  };
  mentionCount: number;
}
```

**System Views (isDefault: true, read-only):**
- All Mentions (icon: Layers) — 5,272 mentions
- High Relevance (icon: Sparkles) — Buy Intent + Own Brand Mention tags — 312 mentions

**Default User Views (isDefault: false, editable):**
- Crisis Management (icon: AlertTriangle) — Own Brand Mention + negative sentiment — 38 mentions
- Brand Monitoring (icon: Shield) — Own Brand Mention — 847 mentions
- Competitor Intel (icon: Crosshair) — Competitor Mention — 2,135 mentions
- Buy Intent (icon: ShoppingCart) — Buy Intent tag — 89 mentions
- Brand Love (icon: Heart) — Own Brand Mention + positive sentiment — 201 mentions

#### `AlertRule`
```typescript
interface AlertRule {
  id: string;
  viewId: string;
  viewName: string;
  volumeSpike: boolean;
  volumeThreshold: number;    // % above 7-day average
  sentimentSpike: boolean;
  sentimentThreshold: number; // % negative mentions
  emails: string[];
  isActive: boolean;
}
```

#### `Mention`
```typescript
interface Mention {
  id: string;
  platform: Platform;
  author: {
    name: string; handle: string; avatar: string;
    verified?: boolean; followerCount?: number;
  };
  content: string;
  timestamp: string;          // ISO 8601
  topicId: string;
  topicName: string;
  sentiment: Sentiment;
  aiTags: AiTag[];
  engagement: { likes: number; replies: number; shares: number };
  url: string;
  isBookmarked: boolean;
  isIrrelevant?: boolean;
}
```

---

### 2. Global Constants (`lib/constants.ts`)

| Constant | Value | Purpose |
|---|---|---|
| `TOPICS_LIMIT` | 5 | Max topics on default plan |
| `MENTIONS_LIMIT` | 10,000 | Max mentions/month on default plan |
| `MENTIONS_USED_MOCK` | 6,842 | Mock current usage (68.4% of limit) |

---

### 3. Zustand Store (`lib/store.ts`)

Complete state surface:
- `userState` / `setUserState` — controls landing vs. setup vs. feed
- `activeSection` / `setActiveSection` — current nav section
- `activeViewId` / `setActiveViewId` — currently selected feed view
- `viewsSidebarOpen` / `setViewsSidebarOpen` — views panel collapsed/expanded
- `mobileNavOpen` / `setMobileNavOpen` — mobile drawer state
- `topics`, `views`, `mentions`, `alertRules` — core data
- `bookmarkedIds` / `toggleBookmark` — bookmark state per mention ID
- `readIds` / `markAsRead` / `markAllRead` — read state per mention ID
- `customTopicTypes` + CRUD — custom topic category management
- `topicModalOpen` / `topicModalTopic` / `topicModalRequired` + CRUD — topic form modal
- `updateMentionSentiment` — override AI-detected sentiment per mention
- `viewModalOpen` + CRUD — create view modal
- `alertModalOpen` / `alertModalViewId` / `alertEditTarget` + CRUD — alert modal

---

### 4. Application Layout

**Global navigation (TopBar):** ContentStudio app-level bar. Navigation: Home, Publisher, Analytics, Inbox, **Listening** (active), Discover, Library. Right section: notifications bell, settings, user avatar.

**Listening module layout:** Two-level sidebar structure:
1. **PrimaryNav** (200px left panel): Listening logo, section tabs (Feed / Bookmarks / Analytics / Alerts / Settings), Topics list with per-topic context menus, usage bar at bottom (topics X/5, mentions X.Xk/10k).
2. **ViewsSidebar** (210px, Feed section only): "For you" system views + user-created "Views". Collapsible with chevron. On mobile: renders as a left-side Drawer overlay.
3. **Content area**: The selected section's content.

**Mobile layout:** Hamburger button in the white content bar opens `PrimaryNav` as a full-screen Drawer. Views panel also renders as Drawer.

---

### 5. User States & Flows

#### Flow 1: New User (Trial Plan)
**Entry:** User navigates to Listening in main nav.
**State:** `userState = 'trial'`
**Sees:** LandingPage with:
- Badge: "Add-on · $49/mo" (violet)
- Headline: "Never miss a mention that matters"
- Subtitle: "Track your brand, competitors, and industry keywords across 18 platforms..."
- **Trial gate banner** (orange): "Not available on trial plans — Listening is a paid add-on. Upgrade to any ContentStudio plan to unlock it." + "Upgrade Plan" CTA button
- Feature highlights: Feed-First, AI-Powered, Instant Alerts (3-card grid)
- Pricing card: "Listening Add-on" — $49/mo — feature list — "Upgrade Plan First" CTA (orange)

#### Flow 2: Locked User (Paid Plan, No Listening)
**Entry:** User navigates to Listening in main nav.
**State:** `userState = 'locked'`
**Sees:** LandingPage with:
- Badge: "Add-on · $49/mo"
- Headline: "Never miss a mention that matters"
- CTAs: "Enable Listening" (primary, violet) + "Preview Demo" (ghost)
- No trial gate banner
- Pricing card shows "Add to ContentStudio" CTA

**Clicking "Enable Listening":** Sets `userState = 'unlocked_new'` → leads to Setup Flow.
**Clicking "Preview Demo":** Sets `userState = 'unlocked'` → skips setup, enters main feed.

#### Flow 3: Expired User (Add-on Lapsed)
**Entry:** User navigates to Listening.
**State:** `userState = 'expired'`
**Sees:** LandingPage with:
- Badge: "Add-on Expired" (orange)
- Headline: "Re-enable Listening"
- Subtitle: "Your Listening add-on has expired. Re-enable it to resume monitoring your brand, competitors, and industry keywords across 18 platforms."
- CTA: "Re-enable Listening" (violet)
- Pricing card shows "Add to ContentStudio" CTA

#### Flow 4: First-Time Setup (Onboarding Wizard)
**State:** `userState = 'unlocked_new'`
**Step 1 — Website Input:**
- Globe icon, heading: "Set up your monitoring"
- Subtitle: "Enter your website and we'll detect your brand, competitors, and suggest topics automatically."
- Input: `https://yourcompany.com` placeholder
- CTA: "Analyze Website" (violet)
- Skip: "Skip, I'll add topics manually →" (ghost)
- If URL provided: shows loading state with two-phase copy ("Detecting brand & competitors..." → "Generating topic suggestions...") then auto-transitions to Step 2 with `aiSuggested = true`
- If no URL: clicking Analyze skips to Step 2 and opens TopicFormModal

**Step 2 — Topics Review:**
- Heading: "Your topics"
- Subhead (if AI-suggested): Sparkles icon + "Our AI suggested these based on your website"
- Subhead (if manual): "Edit, remove, or add more topics before starting."
- Lists all current topics as TopicRow cards with: color bar, topic name, keyword pills (up to 4, +N more), type selector dropdown, edit pencil, delete bin
- "Add another topic" dashed button → opens TopicFormModal
- CTA: "Start Monitoring" (violet, disabled if no topics) → sets `userState = 'unlocked'`

#### Flow 5: Main Feed (Active User)
**State:** `userState = 'unlocked'`
**Layout:** PrimaryNav + ViewsSidebar (Feed section) + FeedShell content

**FeedShell — Feed View:**
- Filter bar: Topic filter (compact multi-select) | Sentiment filter | Tags filter | Min followers filter | Date range chip + "Newest" sort | Mark all read button (appears when unread exist)
- Views trigger (mobile): hamburger + "Views" button to toggle views panel
- Mention cards list (sorted by newest, filtered by active view's criteria)
- Mention count display
- Empty states: "No mentions yet" (with subtext) when no data; "No mentions match your filters" when filtered out

#### Flow 6: Mention Card Interactions
Each MentionCard shows:
- Author avatar (DiceBear) with platform color badge
- Author name + verified badge (if verified) + handle + platform + time ago + topic name
- Content text (clamped to 3 lines when reply is closed; full when reply is open)
- Keyword highlighting in content (yellow highlight on matched keywords)
- Smart tag pills (up to 3 visible, colored per tag)
- Sentiment indicator (colored dot + label, clickable to override)
- Engagement stats: ❤ likes, 💬 replies, 🔁 shares

**Hover actions (floating pill, top-right):**
1. Reply (opens reply panel; disabled + tooltip on non-reply platforms)
2. Bookmark (toggles; filled icon when bookmarked)
3. Open original post (external link)
4. Mark as read (hidden once read)
5. Mark as irrelevant (hides from feed)
6. Copy link to post

**Reply panel (inline, below card content):**
- Account selector: connected accounts for that platform (with avatar, name, handle; expired accounts show warning triangle tooltip: "Access token expired. Reconnect to reply.")
- No accounts state (orange banner): "No [Platform] accounts connected. Connect an account →"
- Textarea: "Write a reply..." placeholder, 2-5 rows
- AI toolbar: "Write with AI" button (Sparkles icon) — generates context-aware draft
- Improve options (shown after text entered): Rephrase | Improve | Shorten | Lengthen | Grammar Fix (each with tooltip)
- Error state (expired token): "This account's access token has expired. Please reconnect to reply." (red inline error)
- Actions: Cancel + Reply (violet; disabled if no text or no accounts; shows "Sent!" green state after)

**Sentiment override:** Clicking the sentiment indicator opens a dropdown with Positive / Neutral / Negative options. Selected option updates the mention's sentiment in the store.

#### Flow 7: Views Management

**ViewsSidebar sections:**
- **For you** (read-only): All Mentions, High Relevance
- **Views** (user-created): Custom views with edit/delete/alert context menus

**Creating a new view (CreateViewModal):**
- Title: "Create View"
- Fields:
  - View name (required, text input) — tooltip: "Give this view a clear name so you can find it quickly, e.g. 'Bug Reports' or 'High Intent Mentions'."
  - Icon picker (7 options): Layers, Sparkles, Shield, Heart, AlertTriangle, ShoppingCart, Crosshair — tooltip: "Pick an icon to visually distinguish this view in the sidebar."
  - Filter by Topics (optional, multi-select) — tooltip: "Limit this view to mentions from specific topics only. Leave empty to include all topics."
  - Filter by Platforms (optional, 18 platforms multi-select) — tooltip: "Only show mentions from these platforms. Leave empty to include all 18 platforms."
  - Language (optional, 16 languages multi-select) — tooltip: "Only show mentions written in the selected languages. Leave empty to include all languages."
  - Filter by Sentiment (optional, pill selector) — tooltip: "Filter by the AI-scored emotional tone of each mention. Leave all unselected to include every sentiment."
  - Filter by Tags (optional, multi-select with color dots) — tooltip: "Filter by AI-detected intent or category. Useful for views like 'Bug Reports' or 'Buy Intent'."
  - Author Reach (optional, min followers selector: Any reach / 500+ / 1k+ / 5k+ / 10k+ / 50k+ / 100k+) — tooltip: "Only show mentions from accounts with at least this many followers. Useful for filtering out low-reach accounts or focusing on influencer-level mentions."
- Footer: Cancel + "Save View" (violet)
- Validation: name required

**View context menu (hover the "…" icon):**
- Duplicate
- Manage alerts
- Delete (with confirmation modal: "Delete view? Are you sure you want to delete [name]? This cannot be undone.")

#### Flow 8: Topic Management

**Topics in PrimaryNav:**
- Colored dot per topic (color based on type or override)
- Active topics: colored dot + topic name
- Paused topics: gray dot + grayed name + PauseCircle icon
- Hover shows "…" context menu: Edit / Pause (or Resume) / Delete

**Pausing a topic:** Confirmation modal: "This will pause tracking for [topic name] and we'll stop looking for new mentions. Existing mentions will remain in your feed." → Cancel | "Pause keyword"

**TopicFormModal — Create/Edit Topic:**
Modal title: "New Topic" or "Edit Topic"

Fields:
- **Topic row** (color swatch + name + type selector):
  - Color swatch: clickable, opens palette (8 colors: `#2563EB #7C3AED #EA580C #10B981 #EF4444 #F59E0B #EC4899 #6B7280`) — tooltip: "Override the default color for this topic"
  - Name input: placeholder "e.g. ContentStudio, Hootsuite…"
  - Type selector: Built-ins (Own Brand / Competitor / Industry Term) + custom types + "New type" inline creation
- **Keywords** (required, tags input): "Search terms used to find mentions. Add your brand name, product names, hashtags, or any phrase to track. Each keyword is matched independently." — placeholder "Type a keyword and press Enter…"
- **Platforms** (required): Platform picker dropdown showing all 18; "Select all 18 platforms" shortcut
- **AI Context Hint** (optional, textarea): "Give our AI a short description of what this topic means to help it reduce noise. Example: 'ContentStudio is a social media scheduling tool, not a photography studio.'" — 200 char limit, counter shown
- **Advanced Settings** (collapsible accordion):
  - Include ANY of these terms: "A mention is shown only if it also contains at least one of these extra terms. Useful for narrowing broad keywords."
  - Include ALL of these terms: "A mention is shown only if it contains every one of these extra terms."
  - Negative terms: "Any mention containing one of these words is automatically hidden, even if it matches your keyword."
  - Negative authors: "All posts from these specific accounts are excluded, regardless of content."
  - Exact match toggle: "Only matches the exact keyword phrase as typed. 'Content Studio' won't match 'ContentStudio' or 'Studio Content'." — most users leave off
  - Case sensitive toggle: "Makes matching case-sensitive. 'Apple' won't match 'apple'. Most users leave this off."
- **Global settings summary** (read-only, at bottom): Shows current global negative terms / blocked authors / excluded subreddits. Link: "Edit global settings →" (navigates to Settings > Global Filters)

**Validation errors:**
- Name: "Topic name is required"
- Keywords: "Add at least one keyword"
- Advanced conflict: "X is in both Keywords and Negative Terms — these would match nothing" / "[terms] appear in both Include Terms and Negative Terms — they cancel each other out" / "[terms] appear in both Include ANY and Include ALL — Include ALL already implies Include ANY"
  (Accordion auto-expands when conflict is detected)

Footer: Cancel + "Save Topic" (violet)

**Topic Type custom creation (inline in TopicTypeSelect):**
- Inline "+ New type" in the dropdown
- Text input: "Type name…" + Add / Cancel buttons
- Enter key creates; Escape cancels

#### Flow 9: Alerts

**AlertsPage sections:**
- Header: "Alerts" + subtext "Get notified when something unusual happens in your feed" + "New Alert" button
- Active alerts section (count)
- Paused alerts section (count)
- Empty state: Bell icon, "No alerts configured", "Create an alert to get notified by email when there's a volume spike or sentiment shift.", "Create your first alert" button

**AlertRow card:**
- View name (bold)
- Trigger pills: Volume Spike (amber, TrendingUp icon, "≥ X% above average") and/or Sentiment Spike (red, TrendingDown icon, "> X% negative")
- Email recipients: team member avatars with names or plain email address pills
- Toggle (active/paused) + "…" menu: Edit / Delete

**CreateAlertModal:**
Title: "Create Alert" / "Edit Alert"

Fields:
- **Monitor this view** (Select): all available views
- **Alert triggers** (two toggleable cards):
  - Volume Spike (amber card): "Unusual surge in mention volume" — when toggled on, shows: "Notify when mentions are [X%] above 7-day average" (NumberInput, 10–500%, default 50%)
  - Sentiment Spike (red card): "Rise in negative sentiment" — when toggled on, shows: "Notify when negative mentions exceed [X%] of total" (NumberInput, 5–100%, default 30%)
- **Email recipients**: Team member picker dropdown (shows team avatar + name + email) + email address text input (Enter/comma to add) + chip display of added recipients. Helper text: "Press Enter or comma to add multiple addresses"

Validation: At least one trigger must be active + at least one email recipient
Footer: Cancel + "Save alert" / "Update alert" (violet, disabled until valid)

#### Flow 10: Analytics

**Filter bar:** Analytics | All topics (multi-select) | Date range chip | Export button

**KPI row (4 cards):**
- Total Mentions (sum of filtered period)
- Positive Sentiment (% positive)
- Avg. Daily Mentions
- Topics Tracked (count)

**Charts (5 total):**
1. **Mentions Over Time** (LineChart): Daily volume per topic, color-coded lines. Each chart has an "AI Insights" button (Sparkles) in the header.
   - AI Insights: "Mention volume peaked 3.5× above average on day 15 — driven by the Hootsuite pricing announcement going viral." / "ContentStudio mentions have grown 18% over the selected period..." / "Tuesdays and Wednesdays consistently drive the highest mention volumes..."
2. **Sentiment Trend** (AreaChart, stacked): Daily positive/neutral/negative breakdown
   - AI Insights: "Positive sentiment dipped to 42% during the Hootsuite pricing controversy..." / "Negative sentiment has trended down for the last 7 days..." / "Neutral mentions spike on weekends..."
3. **Sentiment Distribution** (PieChart donut): Shows % positive in the center; legend shows Positive/Neutral/Negative counts + pcts
   - AI Insights: "58% positive sentiment is 12pp above the social listening industry benchmark..." / "Buy Intent mentions are 78% positive..." / "Negative mentions cluster around pricing and onboarding topics..."
4. **By Platform** (PieChart donut + list): Top 8 platforms by mention share
   - AI Insights: "X / Twitter drives 40% of all mentions — prioritise real-time engagement there." / "Reddit mentions carry 3× higher engagement per post than average..." / "LinkedIn mentions grew 34% this period..."
5. **By Tag** (horizontal BarChart): All 10 AI tags sorted by count
   - AI Insights: "Competitor Mention is the top tag (2,135)..." / "Buy Intent mentions (312) are a direct sales opportunity..." / "Bug Report volume (67) is under 1.3% of total mentions..."

Charts 3 and 4 are side-by-side on desktop, stacked on mobile.

**Date filters:** Presets: All time / Last 24h / Last 7 days / Last 30 days / Last 90 days; Custom range (from/to date pickers)

#### Flow 11: Export

**ExportModal** (3-tab modal titled "Export Report"):

**Download tab:**
- Format toggle: PDF Report | CSV Data
- Date range: uses DateFilterChip (same presets + custom)
- Success state: green Alert "Report ready! Your download should start automatically."
- CTA: "Generate & Download PDF/CSV"

**Schedule tab:**
- Frequency toggle: Weekly | Monthly
- Weekly: "Send on" day-of-week select (Mon–Fri)
- Monthly: "Send on day" select (1st–28th)
- Time select (06:00–18:00 in 1-hour increments)
- Email recipients (same chip+team picker pattern as alerts)
- Success state: "Schedule saved! Reports will be sent automatically."
- CTA: "Save Schedule" (disabled until at least one recipient)

**Share tab:**
- Info text: "Generate a read-only link to this report. Anyone with the link can view the analytics, but cannot make any changes."
- "Generate Share Link" button → shows readonly URL input + copy button
- Copy success: "Link copied to clipboard!" (green)
- Footer note: "Share links expire after 30 days. You can revoke access at any time from Settings."

#### Flow 12: Bookmarks

**Filter bar:** Bookmarks [count] | Search bookmarks text input (190px wide) | Platform filter | Sentiment filter | Tags filter | Date range chip

**Filtering logic:** AND across all active filters, text search matches content OR author name

**Empty states:**
- No bookmarks: Bookmark icon, "No bookmarks yet", "Bookmark mentions from your feed to save them here for later review or sharing with your team."
- No matches: "No bookmarks match your filters" + "Try adjusting your filters or clearing the search."

Cards: Same MentionCard component as in Feed

#### Flow 13: Settings

**Two-tab layout:** Global Filters | Topic Types

**Global Filters tab (four sections):**
1. **Negative Terms** — "Mentions containing these terms will be excluded from all topics. Useful for filtering out spam, unrelated content, or noise." Tag input: "Global negative keywords" (placeholder: "e.g. 'advertisement', 'sponsored'..."), hint: "These apply across all topics. Per-topic negative terms can be set in the Topics tab." Defaults: `spam, bot, advertisement, paid partnership`
2. **Blocked Authors** — "Mentions from these authors will never appear in your feed, regardless of topic." Tag input: "Blocked accounts" (placeholder: "@username or u/username..."). Defaults: `@spamaccount, u/deleted`
3. **Reddit Settings** — "Control which subreddits to exclude from monitoring." Tag input: "Excluded subreddits" (placeholder: "r/subredditname..."). Defaults: `r/spam, r/memes`
4. **Language Filter** — "Only collect mentions in the selected languages. Leave empty to monitor all languages." MultiSelect with 9 languages (English, Spanish, French, German, Portuguese, Italian, Dutch, Japanese, Chinese). Default: English selected.

Footer: "Save Global Filters" button

**Topic Types tab:**
- Card shows built-in types (Own Brand / Competitor / Industry Term) — labeled "built-in", not removable
- Custom types (editable/deletable): Pencil to inline-edit, Trash to delete
- "Add Type" button → inline form with TextInput + Create/Cancel

**Note:** Settings also shows topics configuration accessible via TopicSettingsRow — expandable accordion rows per topic showing keywords (with context/exact/include badges), monitored platforms MultiSelect, topic-level negative terms TagInput, AI Context Hint Textarea.

---

### 6. Usage Limits & Upsell (PrimaryNav)

**Shown only when `userState === 'unlocked'`** at the bottom of PrimaryNav:

- **Topics progress bar**: `X / 5` label + 4px height bar; color: red `#EF4444` at limit, violet `#7C3AED` under limit
- **Mentions this month**: `X.Xk / 10k` label + bar; color: amber `#F59E0B` at ≥90%, violet `#7C3AED` under 90%
- Link: "Need more? Add-ons available →" (violet, 11px)

---

### 7. Complete UI Copy Inventory

#### Badge / Labels
- "Add-on · $49/mo" (violet badge)
- "Add-on Expired" (orange badge)

#### Landing Page Headlines
- Default: "Never miss a mention that matters"
- Expired: "Re-enable Listening"

#### Landing Page Subtitles
- Default: "Track your brand, competitors, and industry keywords across 18 platforms — all in one intelligent feed, inside ContentStudio."
- Expired: "Your Listening add-on has expired. Re-enable it to resume monitoring your brand, competitors, and industry keywords across 18 platforms."

#### Trial Banner
- Title: "Not available on trial plans"
- Body: "Listening is a paid add-on. Upgrade to any ContentStudio plan to unlock it."
- CTA: "Upgrade Plan"

#### Feature Highlights (3 cards)
1. Feed-First — "Every mention from 18 platforms in one clean, prioritized feed."
2. AI-Powered — "Smart-tagged, sentiment-scored, and noise-filtered automatically."
3. Instant Alerts — "Get notified the moment something important happens — email or webhook."

#### Pricing Card
- Title: "Listening Add-on"
- Price: $49 / month
- Subtext: "Billed to your existing ContentStudio subscription"
- Feature list:
  - 5 topics / keywords included
  - 10,000 mentions / month
  - 18 platforms monitored
  - Smart tagging + sentiment analysis
  - Custom views & filters
  - Email & webhook alerts
  - Add-ons: +topics & +mentions available
- CTA (paid): "Add to ContentStudio"
- CTA (trial): "Upgrade Plan First"
- Footer (paid): "Cancel anytime from your billing settings."
- Footer (trial): "Upgrade to any paid plan to enable add-ons."

#### CTAs
- "Enable Listening" / "Re-enable Listening"
- "Preview Demo"
- "Start Monitoring"
- "Analyze Website"
- "Skip, I'll add topics manually →"
- "Add another topic"
- "Save Topic" / "Cancel"
- "Save View" / "Cancel"
- "Create Alert" / "Update alert" / "Cancel"
- "Export Report" / "Generate & Download PDF/CSV" / "Save Schedule" / "Generate Share Link"
- "Pause keyword" / "Cancel" (pause confirmation)
- "Delete view" / "Cancel" (delete confirmation)

#### Empty States
- Feed, no data: "No mentions yet" — no subtext (inferred from filtered empty state)
- Feed, filtered: "No mentions match your filters" — no subtext
- Bookmarks, empty: "No bookmarks yet" — "Bookmark mentions from your feed to save them here for later review or sharing with your team."
- Bookmarks, no match: "No bookmarks match your filters" — "Try adjusting your filters or clearing the search."
- Alerts, empty: "No alerts configured" — "Create an alert to get notified by email when there's a volume spike or sentiment shift." — "Create your first alert"
- Topics list, empty: "No topics yet" (PrimaryNav)
- Views list, empty: "+ Add view" (ViewsSidebar, inline button)

#### Tooltips (complete list)
| Tooltip target | Text |
|---|---|
| Reply (can reply) | "Reply" / "Close reply" |
| Reply (cannot reply) | "Replies not available on [Platform]" |
| Bookmark | "Save to bookmarks" / "Remove bookmark" |
| Open original | "Open original post" |
| Mark as read | "Mark as read" |
| Mark irrelevant | "Mark as irrelevant — hide from feed" |
| Copy link | "Copy link to post" |
| Views expand hint | "Show views panel" |
| Views collapse | "Collapse" |
| Add topic | "Add topic" |
| Add view | "Add view" |
| View "…" menu | (shown on hover only, no tooltip) |
| Topic color swatch | "Override the default color for this topic" |
| Account expired (reply) | "Access token expired. Reconnect to reply." |
| View name field | "Give this view a clear name so you can find it quickly, e.g. 'Bug Reports' or 'High Intent Mentions'." |
| View icon picker | "Pick an icon for this view" / "Pick an icon to visually distinguish this view in the sidebar." |
| Topics filter | "Limit this view to mentions from specific topics only. Leave empty to include all topics." |
| Platforms filter | "Only show mentions from these platforms. Leave empty to include all 18 platforms." |
| Language filter | "Only show mentions written in the selected languages. Leave empty to include all languages." |
| Sentiment filter | "Filter by the AI-scored emotional tone of each mention. Leave all unselected to include every sentiment." |
| Tags filter | "Filter by AI-detected intent or category. Useful for views like 'Bug Reports' or 'Buy Intent'." |
| Author Reach filter | "Only show mentions from accounts with at least this many followers. Useful for filtering out low-reach accounts or focusing on influencer-level mentions." |
| Topic keywords field | "Search terms used to find mentions. Add your brand name, product names, hashtags, or any phrase to track. Each keyword is matched independently." |
| Topic platforms field | "Choose which social networks and content sources to monitor for this topic. Pick a subset if you only care about specific channels." |
| AI Context Hint | "Give our AI a short description of what this topic means to help it reduce noise. Example: 'ContentStudio is a social media scheduling tool, not a photography studio.'" |
| Include ANY | "A mention is shown only if it also contains at least one of these extra terms. Useful for narrowing broad keywords." |
| Include ALL | "A mention is shown only if it contains every one of these extra terms." |
| Negative terms | "Any mention containing one of these words is automatically hidden, even if it matches your keyword." |
| Negative authors | "All posts from these specific accounts are excluded, regardless of content." |
| Exact match | "Only matches the exact keyword phrase as typed. 'Content Studio' won't match 'ContentStudio' or 'Studio Content'." |
| Case sensitive | "Makes matching case-sensitive. 'Apple' won't match 'apple'. Most users leave this off." |
| Global filters (topic modal) | "These rules are set globally in Settings and apply to all topics. They cannot be overridden per-topic." |
| AI Insights button | "AI Insights" |
| Write with AI (reply) | "Let AI draft a reply based on the mention context" |
| Rephrase | "Rewrite the reply with different wording while keeping the same meaning" |
| Improve | "Make the reply clearer, more engaging, and professional" |
| Shorten | "Make the reply more concise without losing key points" |
| Lengthen | "Expand the reply with more detail and context" |
| Grammar Fix | "Fix any spelling and grammar mistakes in the reply" |

---

### 8. Business Rules / Edge Cases

1. **Reply availability:** Only platforms in `REPLY_SUPPORTED_PLATFORMS` (`twitter, instagram, facebook, linkedin, tiktok, youtube, reddit, bluesky, threads`) show the Reply action. Others show the icon at 35% opacity with tooltip "Replies not available on [Platform]".

2. **Expired connected accounts:** In the reply panel account selector, expired accounts show a warning triangle tooltip. Clicking "Reply" with an expired account selected shows: "This account's access token has expired. Please reconnect to reply." (red inline error, reply does not send).

3. **No connected accounts for platform:** Shows orange banner in reply panel: "No [Platform] accounts connected. Connect an account →" (link placeholder). Reply button is disabled.

4. **Topic color precedence:** Per-topic color override > topic type color > fallback `#7C3AED`. Paused topics always show `#D1D5DB` (gray) regardless of type.

5. **Keyword conflict validation:** Three conflict types detected on save:
   - Keyword appears in both Keywords and Negative Terms → "X is in both Keywords and Negative Terms — these would match nothing"
   - Include term appears in Negative Terms → "X appear in both Include Terms and Negative Terms — they cancel each other out"
   - Term in both Include ANY and Include ALL → "X appear in both Include ANY and Include ALL — Include ALL already implies Include ANY"
   Advanced accordion auto-opens when conflict detected.

6. **Live validation:** Errors only appear after first save attempt; re-validates on every change after that.

7. **Topics limit:** At `topics.length >= TOPICS_LIMIT`, the usage bar turns red. Topics can still be added in the prototype (limit is visual only).

8. **Mentions usage warning:** At `MENTIONS_USED_MOCK >= MENTIONS_LIMIT * 0.9` (90%), mentions bar turns amber `#F59E0B`.

9. **Minimum one keyword:** Attempting to close the topic modal without any keyword value shows "Add at least one keyword" error.

10. **Minimum one platform:** The platform picker prevents deselecting the last platform (no-op if `platforms.length === 1`).

11. **Sentiment override:** User can manually override AI-detected sentiment per mention from the clickable dropdown on MentionCard. Change persists in Zustand store for the session.

12. **Keyword highlighting:** Matched keywords from the mention's topic are highlighted in the card content using yellow `#FEF3C7` background with brown `#92400E` text (case-insensitive, regex-escaped).

13. **Alert requires at least one trigger:** "Save alert" is disabled until at least one of volumeSpike or sentimentSpike is true AND at least one email is added.

14. **Alert create from view:** ViewRow context menu "Manage alerts" opens CreateAlertModal pre-filled with that view's ID.

15. **Export scheduling requires recipient:** "Save Schedule" button is disabled until at least one email recipient is added.

16. **Share link expiry:** Share links are noted as expiring after 30 days with revocation available from Settings.

17. **Help widget:** Floating circular button in bottom-right corner of the Setup Flow. (Component `HelpWidget` exists; content not detailed in prototype.)

18. **Mark all read:** Button appears in the feed filter bar only when unread mentions exist. Marks all current `mentions` array as read.

19. **View duplicate:** Creates a copy with "..." suffix and `isDefault: false`.

20. **Delete view with active selection:** If the deleted view was the activeViewId, the store falls back to the next available view or `'view-relevance'`.

---

### 9. Mock Data Summary

**5 Topics:** ContentStudio (own_brand, all platforms, 847 mentions), Hootsuite (competitor, social only, 1,243), Buffer (competitor, social only, 892), Social Media Scheduling (industry, all platforms, 2,156), Tool Alternatives (industry, selected platforms, 134)

**7 Views:** All Mentions (5,272), High Relevance (312), Crisis Management (38), Brand Monitoring (847), Competitor Intel (2,135), Buy Intent (89), Brand Love (201)

**3 Alert Rules:** Crisis Management (sentiment spike, 30% threshold, 2 recipients — active), Buy Intent (volume spike, 75% threshold, 1 recipient — active), All Mentions (both triggers, 2 recipients — paused)

**43 Mention records** spanning 30 days across 12 platforms with varied sentiment, tags, and engagement data (complete dataset documented in mock-data.ts)

**Analytics data:** 30-day volume series per topic with spike on day 15 (3.5×), 30-day sentiment series, platform breakdown (12 platforms), tag distribution (all 10 tags)

---

### 10. Backend Codebase Integration Points

*(From v1 analysis — retained)*

**Must Build from Scratch:**
- Mention ingestion engine — no system currently polls social platforms for brand/keyword mentions
- Sentiment analysis backend service — no NLP pipeline exists
- Brand/keyword listening configuration — no UI or API for users to define what to monitor
- Alert rules — no threshold-based alerting with spike detection
- Analytics dashboard backend — no aggregation pipeline for mention trends

**Reusable Existing Infrastructure:**
- `CustomTopicsModel` (MongoDB, `custom_topics` collection) — schema directly mirrors Topic data model
- Notification & alert infrastructure (Kafka, Redis, Laravel jobs, Pusher real-time)
- Facebook/Instagram webhooks already subscribed to `mentions` field
- `CacheHelper` pattern for Redis keyword caching
- `FeederSentimentIcon.vue` UI component (adapt for mention sentiment)

**New MongoDB Collections Needed:**
- `listening_topics` — monitored keywords/brands
- `mentions` — mention records (indexed on platform, keyword_id, timestamp, sentiment, workspace_id)
- `listening_alerts` — alert rules

**New Kafka Topics:**
- `contentstudio.mentions.ingest`
- `contentstudio.mentions.sentiment`
