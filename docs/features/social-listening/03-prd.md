# PRD: Listening

**Author:** Product Team
**Last Updated:** 2026-03-06
**Status:** Approved
**Target Release:** Q2 2026

---

## 1. Overview

Listening is a $49/month paid add-on for ContentStudio that monitors 18 social platforms, news, blogs, and discussion forums for brand mentions, competitor activity, and industry keywords — and surfaces them in a single, intelligent, feed-first inbox. Users create topics (up to 5 by default) with keywords and platform scope; the system ingests up to 10,000 mentions per month, applies AI-powered smart tagging (10 categories) and sentiment scoring, then presents everything in a unified global feed. Users can reply to mentions inline with AI-assist, save custom filtered views, set spike alerts, and export analytics reports. The feature is web-only in V1 and fully mobile-responsive.

---

## 2. Problem Statement

**What problem are we solving?**

Social media managers and brand owners currently have no native way inside ContentStudio to monitor what people are saying about their brand, competitors, or industry across platforms. They either use expensive standalone tools (Brandwatch, Mention, Brand24 — ranging from $49–$499/mo), manually search each platform, or simply don't monitor at all and discover PR issues after they've already spread. This is a critical gap: ContentStudio handles publishing, scheduling, inbox, and analytics but leaves brand monitoring entirely to external tools — forcing users to context-switch and duplicate their workspace setup.

**Who has this problem?**

- **Social media managers** (primary) — responsible for brand reputation, community engagement, and competitive intelligence. They check for mentions multiple times per day across multiple accounts.
- **Agency account managers** — managing brand monitoring for clients, need consolidated views and scheduled reports.
- **Founders / growth teams** — monitoring competitor activity and customer sentiment to inform product and marketing decisions.
- Estimated 30–40% of ContentStudio's customer base actively uses a third-party listening tool today (based on competitor comparison searches, support requests, and user interviews). This represents direct churn risk to specialized tools.

**What happens if we don't solve it?**

- Users keep paying two subscriptions (ContentStudio + a listening tool), which weakens ContentStudio's value proposition and increases churn risk.
- ContentStudio falls further behind Hootsuite (Streams), Buffer (Mentions), and Metricool (Mentions tab) — all of which have native listening capabilities.
- PR crises and viral negative mentions go undetected; users blame ContentStudio for leaving them blind.
- Agency users specifically require listening for client reporting — without it they cannot consolidate their workflow in ContentStudio.

---

## 3. Goals & Success Metrics

| Goal | Metric | Target | How We'll Measure |
|---|---|---|---|
| Add-on revenue | Monthly add-on subscribers | 500 subscribers within 90 days of launch | Billing data |
| Reduce tool fragmentation | % of subscribers who cancel a competing listening tool | 30% self-report cancellation | Post-activation survey (30-day) |
| Engagement | DAU / MAU ratio for Listening module | ≥ 40% (i.e., users check it more than once a week) | Product analytics (feature-level) |
| Onboarding success | % of new subscribers who complete setup + create ≥1 topic | ≥ 75% within 24h of activation | Funnel analytics |
| Retention | 3-month add-on retention rate | ≥ 80% | Billing churn data |
| Guard rail | No increase in overall ContentStudio churn due to pricing concerns | < 0.5% delta | Billing data |

---

## 4. Target Users

**Primary Persona:**
Social media manager at a mid-size brand or agency — manages 3–10 social accounts across multiple platforms. Checks ContentStudio daily for scheduling and inbox. Currently uses Brand24 or Mention separately. Wants brand monitoring in the same tool they live in. Not highly technical; expects intuitive setup with smart defaults.

**Secondary Persona:**
Agency account manager — runs ContentStudio workspaces for multiple clients. Needs scheduled PDF/CSV reports to send to clients weekly. Values custom views per client segment (brand vs. competitor) and the ability to set topic types (Own Brand / Competitor / Industry) for organizational clarity.

**Tertiary Persona:**
Founder / head of growth — monitors competitor launches, hiring signals (from GitHub/LinkedIn), and industry conversations. Uses the Analytics tab to track share of voice trends. Power user of advanced keyword filters (include ALL, negative authors, Reddit exclusions).

**Non-Users (explicitly out of scope):**
- Enterprise customers requiring Salesforce/CRM integration with mentions — out of scope for V1.
- Developers wanting API access to raw mention data — V1 is UI-only.
- Users on trial plans — Listening is paid-only; trials see the upsell gate.

---

## 5. User Stories / Jobs to Be Done

| ID | As a... | I want to... | So that... | Priority |
|---|---|---|---|---|
| US-1 | Social media manager | see all mentions of my brand across platforms in one feed | I never miss a conversation I need to respond to | Must Have |
| US-2 | Social media manager | reply to mentions directly from the feed | I don't have to open each platform separately | Must Have |
| US-3 | Social media manager | filter mentions by sentiment, platform, and tag | I can prioritize what to act on first | Must Have |
| US-4 | Social media manager | save custom filtered views (e.g., "Crisis Monitor", "Brand Love") | I can instantly switch to the segments that matter most to me | Must Have |
| US-5 | Social media manager | be alerted when mention volume spikes or sentiment turns negative | I catch a PR crisis before it spreads | Must Have |
| US-6 | Agency account manager | create separate topics for each client's brand and competitors | I can keep clients' data organized | Must Have |
| US-7 | Agency account manager | export analytics as PDF or schedule weekly email reports | I can deliver professional reporting to clients without manual work | Must Have |
| US-8 | Brand manager | see sentiment trends and platform breakdown over time | I can report on brand health to leadership | Should Have |
| US-9 | Power user | configure advanced keyword rules (include ALL, negatives, exact match) | I get high-signal mentions without noise | Should Have |
| US-10 | Social media manager | get AI-generated reply drafts when engaging with mentions | I reply faster and more consistently | Should Have |
| US-11 | Founder | use AI smart tags (Buy Intent, Bug Report, Competitor Mention) to identify high-value mentions | I can route the right mentions to the right team without reading everything | Should Have |
| US-12 | Agency manager | pause a topic during off-seasons or closed campaigns without deleting it | I don't lose the configuration when I restart monitoring | Should Have |
| US-13 | Social media manager | see which subreddits to exclude | I don't get spammed by off-topic subreddit noise | Nice to Have |
| US-14 | Power user | override AI-assigned sentiment on a mention | I can correct misclassifications and maintain accurate analytics | Nice to Have |
| US-15 | Social media manager | bookmark important mentions to reference later | I can compile examples for presentations or client reviews | Nice to Have |

---

## 6. Requirements

### 6.1 Must Have (P0)

**Topics & Monitoring**
- Users can create up to 5 topics (default limit). Each topic requires a name, ≥1 keyword, and ≥1 platform.
- Topics can be assigned a type: Own Brand, Competitor, Industry Term, or custom user-defined types.
- Topics can be paused (stops new ingestion, retains historical mentions) and resumed at any time.
- Topics can be deleted immediately; paused topics show confirmation modal before pausing.
- Monitoring covers all 18 platforms: Twitter/X, Instagram, Facebook, LinkedIn, TikTok, YouTube, Reddit, Bluesky, Pinterest, Threads, Hacker News, GitHub, DEV.to, Stack Overflow, Podcasts, Newsletters, News, Blogs.

**Feed**
- All mentions across all active topics surface in a single global feed ("All Mentions" view).
- Feed is filterable by: topic(s), sentiment, AI smart tags, minimum follower count, date range, and sort order.
- Each mention card shows: platform icon, author name, time, topic label, sentiment badge, AI tags, engagement count, and 3-line content preview with matched keyword highlighted.
- Unread mentions are visually distinguished; "Mark all read" clears all unread indicators.
- Hover reveals action pill: Reply, Bookmark, Open original, Mark as read, Mark as irrelevant, Copy link.

**Reply**
- Inline reply panel expands from the mention card (no page navigation).
- Account selector shows connected accounts for that platform (using existing CS OAuth tokens).
- If no accounts connected for that platform: orange banner with "Connect an account →" link; Reply button disabled.
- If account token expired: warning icon on account in dropdown, error message on send attempt.
- Reply is supported on: Twitter/X, Instagram, Facebook, LinkedIn, TikTok, YouTube, Reddit, Bluesky, Threads (9 platforms). Non-supported platform icons shown at reduced opacity with tooltip.

**Smart Tagging + Sentiment**
- AI automatically assigns ≥0 tags from 10 categories to each mention: Own Brand Mention, Competitor Mention, Industry Insight, Buy Intent, Bug Report, User Feedback, Promotional Post, Product Question, Event, Hiring.
- AI automatically assigns sentiment: Positive, Neutral, or Negative.
- User can override AI-assigned sentiment from the mention card (dropdown).

**Views**
- 2 system views (read-only): "All Mentions", "High Relevance" (Buy Intent + Own Brand Mention tagged).
- Users can create, rename, and delete custom views; each view is a saved combination of topic / platform / language / sentiment / tag / author reach filters.
- Active view is preserved across navigation.
- Deleting the active view falls back to the next available view.

**Alerts**
- Users can create alerts per view for: Volume Spike (% above 7-day rolling average, configurable) and/or Sentiment Spike (% negative exceeding threshold, configurable).
- At least 1 trigger type must be active and at least 1 email recipient must be added before an alert can be saved.
- Alerts can be toggled active/paused without deleting.
- Alert emails sent to specified recipients (workspace team members or manually entered email addresses).

**Analytics**
- Date range filter: All time, Last 24h, Last 7 days, Last 30 days, Last 90 days, Custom range.
- Topic filter: multi-select (defaults to all topics).
- KPI cards: Total Mentions, Positive Sentiment %, Avg. Daily Mentions, Topics Tracked.
- Charts: Mentions Over Time (line, per topic), Sentiment Trend (stacked area), Sentiment Distribution (donut), By Platform (donut + ranked list), By Tag (horizontal bar).
- Each chart has an "AI Insights" button showing 3 AI-generated contextual bullet points.

**Export**
- Download: PDF report or CSV data with date range selector. Immediate download on button click.
- Schedule: Weekly or monthly recurring reports emailed to specified recipients.
- Share: Read-only analytics link (30-day expiry).

**Settings**
- Global Filters tab: negative terms (exclude from all topics), blocked authors, excluded subreddits, language filter.
- Topic Types tab: view built-in types; create, edit, delete custom types.
- Global filter summary shown as read-only block inside TopicFormModal.

**Subscription / Pricing**
- Listening is a $49/month add-on billed to existing ContentStudio subscription.
- Default limits: 5 topics, 10,000 mentions/month.
- Usage bars visible in PrimaryNav (bottom) when subscribed: shows topics X/5 (red at limit) and mentions X.Xk/10k (amber at ≥90%, red at limit).
- "Need more? Add-ons available →" CTA in usage bar.
- Trial users see upsell gate; paid CS plan users see enable CTA; expired subscribers see re-enable CTA.

**Onboarding**
- Two-step setup wizard on first activation: (1) website URL input → AI brand/competitor detection → topic suggestions; (2) topic review → "Start Monitoring".
- Skip option if user prefers to add topics manually.

**Mobile Responsiveness**
- Full mobile support (<768px): PrimaryNav as hamburger Drawer, ViewsSidebar as Drawer, horizontally scrollable filter bars, stacked chart layouts.

### 6.2 Should Have (P1)

- Bookmarks: saved mentions with search + platform / sentiment / tag / date filters. Empty states for zero bookmarks and no-match states.
- AI-powered reply drafting: "Write with AI" generates a context-aware reply draft; "Rephrase", "Improve", "Shorten", "Lengthen", "Grammar Fix" transform existing draft text.
- Advanced keyword configuration per topic: Include ANY of these terms, Include ALL of these terms, Negative terms (topic-level), Negative authors (topic-level), Exact match toggle, Case sensitive toggle.
- Keyword conflict detection: surface errors when a keyword appears in both Keywords and Negative Terms (or other conflicting combinations); auto-open Advanced accordion to highlight the conflict.
- AI Context Hint per topic (200 chars): free-text field helping the AI understand ambiguous brand names.
- View icons: 7 icon options per custom view.
- View duplication and "Manage alerts" shortcut from view context menu.
- Sentiment override persistence to database (session-only is acceptable for V1 prototype, production must persist).

### 6.3 Nice to Have (P2)

- Platform-level exclusions per topic (beyond global subreddit filter).
- In-app spike notifications (currently alert is email-only).
- "Open in Inbox" from mention card to create a full Inbox thread.
- Share link revocation from Settings UI.
- Webhook delivery for alerts (email-only in V1).
- Custom topic color picker (8 swatches available in prototype).

### 6.4 Explicitly Out of Scope (V1)

- Share of Voice calculations (requires complex cross-brand aggregation).
- Predictive trend forecasting (ML dependency).
- Influencer discovery / author ranking by reach.
- Historical data tiers (30d / 90d / 1yr) — V1 ingests forward from activation.
- Competitor benchmarking charts (Compare tab from v1 architecture).
- AI weekly narrative digest email.
- Saved dashboards per brand/client (agency multi-workspace).
- Spam / bot auto-detection.
- Mobile app (iOS/Android) — web-only for V1.
- API access to raw mention data.
- CRM or Slack integration for mention routing.

---

## 7. User Flow (High Level)

**First-time activation (paid plan, not yet subscribed):**
1. User clicks "Listening" in the ContentStudio global top nav.
2. Landing page shown: feature highlights, $49/mo pricing card, "Enable Listening" + "Preview Demo" CTAs.
3. User clicks "Enable Listening" → billing confirmation (handled by existing billing flow) → redirected to Setup Wizard.
4. Step 1: User enters their website URL → clicks "Analyze Website" → AI detects brand + competitors (2-second animation) → pre-populates topic suggestions.
5. Step 2: User reviews suggested topics, edits types/keywords if needed, adds more topics via TopicFormModal, then clicks "Start Monitoring."
6. User lands in the main Feed with "All Mentions" selected in ViewsSidebar.

**Daily usage (active subscriber):**
1. User navigates to Listening → lands in Feed, last-used view active.
2. Scans mention cards — new unread mentions visible.
3. Clicks "Reply" on a high-priority mention → inline panel → AI drafts a reply → user edits → sends.
4. Clicks sentiment badge on a mislabelled mention → overrides to correct value.
5. Switches to Analytics → checks 30-day sentiment trend → AI Insights highlight a spike.
6. Exports weekly PDF report.

**Alert fired (email):**
1. Volume spike detected on "Brand Monitoring" view.
2. Alert email sent to configured recipients.
3. User clicks link in email → lands in Listening Feed with that view pre-selected.

---

## 8. Business Rules & Constraints

| Rule ID | Rule | Rationale |
|---|---|---|
| BR-1 | Each topic requires ≥1 keyword and ≥1 platform to be saved | A topic with no keywords or platforms would match nothing |
| BR-2 | A keyword cannot appear in both the keyword list and the negative terms list for the same topic | These filters would cancel each other out, matching nothing |
| BR-3 | An include term cannot also appear in the negative terms list | Same cancellation conflict |
| BR-4 | A term cannot appear in both Include ANY and Include ALL | Include ALL already implies Include ANY; this is a configuration error |
| BR-5 | Default topic limit is 5 per workspace | Controlled limit enables predictable pricing; add-on available for more |
| BR-6 | Default mention limit is 10,000/month per workspace | Ingestion stops at limit; user must purchase add-on to resume |
| BR-7 | Alerts require ≥1 trigger type active AND ≥1 recipient | An alert with no trigger or no recipient cannot fire usefully |
| BR-8 | Replies can only be sent from connected accounts for the matching platform | OAuth scope is platform-specific; cross-platform reply is technically impossible |
| BR-9 | Reply is not available on: Hacker News, GitHub, DEV.to, Stack Overflow, Podcasts, Newsletters, News, Blogs | These platforms do not expose write APIs accessible via OAuth |
| BR-10 | System views ("All Mentions", "High Relevance") cannot be edited or deleted | They are global defaults required for the product to function correctly |
| BR-11 | Custom topic types can be deleted, but built-in types (Own Brand, Competitor, Industry Term) cannot | Built-in types are referenced by the AI tagging engine |
| BR-12 | Trial plan users cannot enable Listening | Listening is paid-only; trial upgrade required |
| BR-13 | Topics can be paused without losing configuration or historical mentions | Pause is a billing/monitoring state, not destructive |
| BR-14 | Share links expire after 30 days | Prevents indefinite exposure of analytics data |
| BR-15 | Global filters apply to all topics; per-topic filters apply only to that topic | Two-level hierarchy; global filters always run first |

---

## 9. Open Questions

| Question | Options | Owner | Due Date | Decision |
|---|---|---|---|---|
| What happens at the mention limit — hard stop or grace period? | Hard stop (new mentions discarded) / Grace period (keep ingesting, notify) / Soft limit (overage billing) | Product / Engineering | Before dev start | Pending |
| How far back does historical data go on activation? | From activation only / Last 7 days backfill / Last 30 days backfill | Product / Data | Before dev start | Pending |
| Should keyword conflict detection block save or warn? | Block (current prototype) / Warn but allow | Product | Before dev start | Block (current) |
| Are topic limits per workspace or per user? | Per workspace (current assumption) / Per user seat | Product | Before dev start | Per workspace (assumed) |
| What is the retry/fallback behavior if a platform API is down? | Queue and retry / Discard / Show partial results | Engineering | Before dev start | Pending |
| Does the AI Context Hint field need moderation? | No (trust user input) / Character limit only (current: 200) / Content filter | Product | Before dev start | 200-char limit only (current) |
| Should deleting a topic with associated views cascade-delete those views? | Yes / No (views become "all topics" scoped) / Warn user | Product / Engineering | Before dev start | Pending |

---

## 10. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Platform API rate limits cause incomplete mention ingestion | High | High | Per-platform rate limit management, queue/retry with exponential backoff; surface "partial data" warning in UI when a platform is throttled |
| AI tagging accuracy too low → users lose trust | Medium | High | Show confidence scores in later versions; allow manual override (already in V1); include user feedback loop on tag corrections |
| Mention spam / low-quality sources overwhelm feed | Medium | Medium | Global filters (negative terms, blocked authors); per-topic negative terms; "Mark as irrelevant" action that trains filters over time |
| Users hit the 10,000 mention limit before month end | Medium | Medium | Amber warning at 90%; red warning at 95%; email notification at 100%; upsell modal with add-on purchase flow |
| Reply from expired OAuth token creates silent failure | Medium | High | Token expiry detection at account selection time (⚠ icon + tooltip); hard error on send attempt; link to reconnect flow |
| Webhook feature promised on landing page but not in V1 | Low | Medium | Landing page copy already says "email or webhook alerts" — update to "email alerts" for V1 launch to avoid expectation mismatch |
| Data privacy: monitoring mentions of individuals without their knowledge | Low | High | Monitor only public content; add Privacy Policy note in onboarding; comply with platform Terms of Service for each API |
| Latency: mentions appear hours after they're posted | Medium | Medium | Set clear expectations in UI ("mentions typically appear within 30–60 minutes"); for Twitter/X, aim for near-real-time via streaming API |

---

## 11. Dependencies

**Internal:**
- **Connected Accounts (OAuth):** Reply panel requires existing connected account tokens. Expired token detection reuses the same `isExpired` flag used by Publisher and Inbox.
- **Team Members directory:** Alert email recipients and export recipients are drawn from the existing workspace team member API.
- **Billing / Subscription system:** "Enable Listening" and "Add-on" purchase flows must integrate with ContentStudio's existing billing infrastructure.
- **AI / Content Generation pipeline:** "Write with AI" reply drafts + "AI Insights" per chart use ContentStudio's existing AI generation layer (see `contentstudio-ai-agents/`). Reuse the same caption generation prompts pattern for reply drafts.
- **Notification system:** Alert emails use existing transactional email infrastructure.

**External:**
- Platform APIs for 18 data sources. Each has its own rate limits, ToS, and data freshness characteristics:
  - Twitter/X: Streaming API (fast) — requires approved developer access.
  - Reddit: PRAW / Reddit API — rate-limited, ToS requires attribution.
  - Instagram / Facebook / LinkedIn / TikTok: Business API or scraping (platform-dependent availability).
  - Hacker News, GitHub, DEV.to, Stack Overflow: Public APIs, generally permissive.
  - Podcasts, Newsletters, News, Blogs: RSS + web crawling.
- **PDF generation library:** For PDF export (wkhtmltopdf, Puppeteer, or server-side PDF service).
- **Email delivery service:** Transactional email for alerts and scheduled reports (already in use for other CS notifications).

**Blockers:**
- Twitter/X developer API access must be secured before backend implementation begins.
- Billing system must support add-on SKUs (if not already).
- AI topic suggestion (website analysis in onboarding) requires the AI agent pipeline to be extended with a brand detection prompt.

---

## 12. Appendix

- **Prototype:** `cs-prototypes/app/features/listening2/` (Next.js 15 / Mantine UI 8 / Zustand)
- **Live prototype:** https://cs-prototypes.vercel.app/features/listening2
- **Prototype inventory:** `docs/features/social-listening/01-prototype-inventory.md`
- **Workflow document:** `docs/features/social-listening/02-workflow.md`
- **Shortcut Docs:**
  - Prototype Inventory: https://app.shortcut.com/contentstudio-team/write/IkRvYyI6I3V1aWQgIjY5YWEyMjZhLTFjNWQtNDBkNC1hMzNlLTJkNzk4YzljYTEyMyI=
  - Workflow Design: https://app.shortcut.com/contentstudio-team/write/IkRvYyI6I3V1aWQgIjY5YWEyMjdhLWY5N2YtNDQxMy04OTdlLTlhMmU1NTgzYTEyOCI=

---

## Changelog

| Date | Author | Changes |
|---|---|---|
| 2026-03-06 | Product Team | Full rewrite based on completed listening2 prototype — replaces v1 PRD |
