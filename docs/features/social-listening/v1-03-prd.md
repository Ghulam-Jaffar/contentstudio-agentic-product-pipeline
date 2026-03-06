# PRD: Social Mentions

**Author:** Product Team
**Last Updated:** 2026-03-05
**Status:** In Review
**Target Release:** Q2 2026

---

## 1. Overview

Social Mentions is a keyword-based, feed-first social listening add-on for ContentStudio. Users create **Topics** — named groups of keywords representing their brand, competitors, or industry terms — and ContentStudio monitors 18 platforms in real time, surfacing matching posts in a unified mention feed. AI automatically scores relevance, detects sentiment, and tags each mention (Buy Intent, Bug Report, Own Brand, etc.) so users can filter, organize, and act without the noise. Saved filter presets called **Views** let users slice the feed any way they need. Alerts notify teams via email, Slack, or webhook when something important happens. At $49/month, it's the easiest way for ContentStudio's SMB and mid-market users to know what's being said about them across the internet — without switching tools.

---

## 2. Problem Statement

**What problem are we solving?**

ContentStudio users currently have no way to monitor what's being said about their brand, competitors, or industry across social platforms and the web — without leaving ContentStudio. They rely on Google Alerts (limited, slow, email-only), manual platform searches (time-consuming, incomplete), or expensive dedicated tools like Brandwatch or Sprout Social's listening module (overkill for SMBs). As a result, they miss buying signals, competitor complaints, crisis moments, and engagement opportunities every day.

**Who has this problem?**

- **Social media managers** at SMB and mid-market companies (ContentStudio's core user base) who need to stay on top of brand reputation without enterprise budgets
- **Founders and marketing leads** at SaaS companies who want to catch competitor mentions and buying intent signals ("looking for alternatives to X")
- **Agency managers** running brand monitoring for multiple clients
- Estimated: 40–60% of ContentStudio's subscriber base would benefit; 15–25% would pay for the add-on

**What happens if we don't solve it?**

- Users who need social listening will adopt a dedicated tool (Octolens, Brand24, Mention.com) and reduce ContentStudio to a scheduling-only tool — increasing churn risk
- Competitive gap: Hootsuite, Sprout Social, and Agorapulse all include social listening; ContentStudio's absence is a known objection in sales
- Missed upsell: $49/mo add-on with strong retention characteristics (monitoring is habitual, stickier than scheduling)

---

## 3. Goals & Success Metrics

| Goal | Metric | Target | How We'll Measure |
|---|---|---|---|
| Drive add-on revenue | Monthly add-on subscribers | 500 paying workspaces within 6 months | Billing data |
| Prove retention value | Add-on churn rate | <5%/month | Billing data |
| Demonstrate engagement | Weekly active users (visited feed) | 60% of add-on subscribers | Product analytics |
| Ensure data quality | Mention relevance score (AI accuracy) | >85% precision on relevance tags | Manual audit sample (n=200/month) |
| Guard rail | Core platform churn | No increase vs. control group | Cohort analysis |

---

## 4. Target Users

**Primary Persona: The Social Media Manager**
Works at a 10–200 person company, manages 3–10 social profiles, posts daily. Wants to know when the brand is mentioned, catch complaints before they escalate, and spot engagement opportunities. Not a data analyst — needs a clean feed, not a dashboard. Currently uses Google Alerts and manual searches. Budget-conscious: will pay $49/mo if it saves 1 hour/week.

**Secondary Persona: The Founder / Growth Marketer**
At a SaaS company, actively monitors competitors and tracks "looking for alternatives" posts as sales leads. Wants Buy Intent alerts in Slack. May track 5–10 competitor keywords. Values precision over volume.

**Secondary Persona: The Agency Manager**
Manages listening for 3–10 client brands. Needs separate topics per client (workspaces handle this), clean export, and reliable alerts to show clients. Values professionalism of the tool — will use prototype screenshots in client pitches.

**Non-Users (out of scope):**
- Enterprise users needing historical data beyond 7 days backfill, custom taxonomy, or SLA guarantees — this is not a Brandwatch replacement
- Users needing Instagram/Facebook public mention search — platform API limitations prevent this at launch; only connected account mentions for those platforms in V2

---

## 5. User Stories / Jobs to Be Done

| ID | As a... | I want to... | So that... | Priority |
|---|---|---|---|---|
| US-1 | Social media manager | Set up monitoring for my brand in under 5 minutes | I don't need to configure anything complex to start seeing value | P0 |
| US-2 | Social media manager | See all mentions of my brand in a single feed | I don't miss any conversations happening about us | P0 |
| US-3 | Social media manager | Filter mentions by platform, sentiment, and AI tag | I can focus on what matters right now without scrolling through noise | P0 |
| US-4 | Growth marketer | Track competitor keywords alongside my own brand | I can spot opportunities when someone complains about a competitor | P0 |
| US-5 | Founder | Get alerted on Slack when there's a mention spike or a Buy Intent post | I can act quickly without checking the tool manually | P0 |
| US-6 | Social media manager | Create saved Views with my own filter combinations | I can instantly see "crisis mentions" or "buy intent" without re-filtering every time | P0 |
| US-7 | Agency manager | Bookmark specific mentions | I can save important posts to review or share with clients later | P1 |
| US-8 | Growth marketer | See a chart of mention volume over time per topic | I can track whether brand awareness is growing | P1 |
| US-9 | Social media manager | Set per-keyword negative terms and include filters | I don't get irrelevant mentions cluttering my feed | P1 |
| US-10 | Social media manager | Pause a topic without deleting it | I can temporarily stop monitoring a keyword during a campaign | P2 |

---

## 6. Requirements

### 6.1 Must Have (P0)

**Onboarding**
- AI-powered setup: user optionally enters company website URL → system suggests 4–6 topics (own brand, competitors, industry terms) with keywords pre-filled
- Website URL is optional; users can skip and add topics manually
- No company info confirmation step — go straight to topic review
- Topics displayed as cards in onboarding: name, type badge, keywords, platform toggles, optional AI context hint
- "Start Monitoring" CTA triggers 7-day backfill (up to 100 results per platform per keyword)

**Topics & Keywords**
- Create, edit, pause, and delete topics
- Each topic has: name, type (Own Brand / Competitor / Industry Term / Custom), one or more keywords, platform selection
- Per-keyword settings: AI context hint (200 chars max), Include ANY OF, Include ALL OF, Negative terms, Negative authors, Wildcard negative terms, Exact match toggle, Case sensitive toggle
- Platform selection per topic (all 18 platforms available; user toggles)
- Global keyword settings (workspace-level): global negative terms, global negative authors, allowed subreddits, excluded subreddits, excluded GitHub repos

**Feed**
- Unified mention feed in reverse chronological order (infinite scroll)
- Mention card displays: platform icon, author name + handle + avatar, post timestamp, topic chip, content excerpt (3 lines), AI tags, sentiment badge (positive/neutral/negative), engagement stats (likes, replies, shares/reposts), Bookmark button, Open original post button, More actions menu
- More actions: Mark as irrelevant, Copy link
- AI tags shown as colored chips on each card
- Irrelevant mentions (scored by AI) hidden by default; toggle to show
- Feed filters (inline bar): search text, date range, platform, sentiment, AI tag

**Views**
- 6 default Views pre-created (see workflow for definitions): High Relevance, All Mentions, Brand Monitoring, Brand Love, Crisis Management, Buy Intent, Competitor Intelligence
- Create custom Views: filter by topic, platform, sentiment, AI tag, language; AND/OR logic
- Rename, edit, delete custom views
- View displayed in sidebar with mention count badge
- Active view highlighted in sidebar

**AI Tagging & Sentiment**
- Every mention AI-scored for relevance to the workspace
- AI tags applied (up to 3 per mention): Own Brand Mention, Competitor Mention, Industry Insight, Buy Intent, Bug Report, User Feedback, Promotional Post, Product Question, Event, Hiring
- Sentiment: Positive / Neutral / Negative

**Alerts**
- Create alert from active View or from Alerts page
- Alert types: New Mentions (any match), Volume Spike (X% above 7-day average), Sentiment Shift (negative % crosses threshold), First Mention (first time a keyword fires)
- Destinations: Email (recipient list), Webhook (URL + optional secret)
- Frequency options (New Mentions only): Realtime, Hourly digest, Daily digest, Weekly digest
- Alert rule management: list, toggle active/inactive, edit, delete

**Subscription & Access**
- $49/month add-on
- 7-day free trial on first activation
- Non-subscribers see locked landing page with value prop, feature highlights, $49/mo pricing card, and demo access
- Trial users see banner with days remaining + upgrade CTA
- Expired subscribers see feed in locked state with re-subscribe prompt

### 6.2 Should Have (P1)

**Analytics**
- Total Mentions Over Time — line chart, filterable by topic + platform + date range
- Topic Comparison — grouped bar chart (each topic as a series, by day/week/month)
- Platform Breakdown — donut chart + table with percentage of total
- Sentiment Trend — stacked area chart (positive/neutral/negative over time)
- AI Tag Breakdown — horizontal bar chart (count per tag type)
- Date range picker: Last 7 days / 30 days / 90 days / Custom
- Download each chart (PNG and CSV)
- AI Insights per chart (sparkle button → popover with 2–3 AI-generated observations)

**Bookmarks**
- Bookmark any mention from the feed (bookmark icon on card)
- Dedicated `/social-mentions/bookmarks` page with same card layout
- Empty state with CTA to return to feed

**Settings**
- Global Filters tab: global negative terms, global negative authors, allowed/excluded subreddits, excluded GitHub repos
- Topics tab: table of all topics with type, keyword count, mention count (last 30 days), status; edit/pause/delete actions; opens topic settings drawer

### 6.3 Nice to Have (P2)

- Slack alert destination (connected workspace + channel selector)
- Pause/resume individual topics without deleting
- Language filter on Views
- Export mentions as CSV from feed (filtered by current view)
- Mention count badge on left nav item

### 6.4 Explicitly Out of Scope

- Reply to mentions from ContentStudio (V2)
- Assign mentions to teammates / internal notes (V2)
- AI weekly summary briefs (V2)
- Bulk actions on mentions (V2)
- Mobile app support for Social Mentions
- Historical data beyond 7-day backfill
- Instagram and Facebook public keyword search (platform API limitations; connected account mentions only in V2)
- White-label / custom branding of the listening interface
- Sentiment score on KPI cards (mentioned feed, not analytics focus)
- Multiple workspaces / client-separated listening within one account (handled by CS workspace model already)

---

## 7. User Flow (High Level)

**First-time subscriber:**
1. User activates Social Mentions add-on → lands at `/social-mentions/setup`
2. Optionally enters company website URL → AI suggests 4–6 topics (or skips)
3. Reviews/edits topic cards → clicks "Start Monitoring"
4. Lands at `/social-mentions/feed` (All Mentions view, empty or loading)
5. Toast: "We're fetching your first mentions. Check back in a few minutes."

**Daily use:**
1. User opens `/social-mentions/feed` → sees mention cards in active view
2. Switches views from sidebar (e.g., "Buy Intent") → feed filters instantly
3. Bookmarks an interesting mention → appears in Bookmarks
4. Clicks "Open ↗" on a card → original post opens in new tab

**Alert setup:**
1. User activates a view → clicks "Add Alert" (top right of feed)
2. Selects alert type → sets threshold if spike/sentiment
3. Selects destination (email/Slack/webhook) + frequency
4. Saves → alert is active

---

## 8. Business Rules & Constraints

| Rule ID | Rule | Rationale |
|---|---|---|
| BR-1 | Each topic must have at least one keyword | A topic with no keywords would never match anything |
| BR-2 | Maximum 50 topics per workspace on the $49 plan | Prevents API cost overrun; higher limits possible in future tiers |
| BR-3 | Maximum 10 keywords per topic | Keeps query complexity manageable per platform API call |
| BR-4 | Keyword context hint is capped at 200 characters | Matches Octolens standard; longer prompts add cost without meaningful accuracy gain |
| BR-5 | 7-day backfill on topic creation (up to 100 results per platform per keyword) | Matches industry standard; prevents users from seeing an empty feed on day one |
| BR-6 | Mention deduplication: same post_id per platform stored once | Prevents duplicate cards when a post matches multiple keywords |
| BR-7 | AI relevance scoring runs on all fetched mentions before they surface | Users should never need to scroll through clearly irrelevant posts |
| BR-8 | Irrelevant mentions are hidden by default; user can toggle to show | Keeps feed clean without permanently discarding data |
| BR-9 | Alert "Volume Spike" baseline uses rolling 7-day average | Requires at least 3 days of data before spike detection activates |
| BR-10 | Trial period is 7 days, one trial per workspace | Prevents trial abuse |
| BR-11 | Paused topics stop fetching new mentions but retain historical data | Users can resume without losing past mentions |
| BR-12 | Global negative terms apply to all topics; per-keyword settings apply after global filters | Global is the outer gate; keyword-level filters refine further |

---

## 9. Open Questions

| Question | Options | Owner | Due Date | Decision |
|---|---|---|---|---|
| How do we handle Instagram/Facebook mentions? | Connected account mentions only (V1) vs. skip entirely until V2 | Product + Engineering | Sprint planning | Pending |
| Slack integration method | OAuth app vs. incoming webhook URL (user-provided) | Engineering | P2 sprint | Deferred to P2 |
| AI classification service | Use existing contentstudio-ai-agents platform vs. third-party (AWS Comprehend) | Engineering | Sprint planning | Pending |
| Reddit data source | Official Reddit API vs. third-party aggregator (Pushshift alternative) | Engineering | Sprint planning | Pending |
| Platform fetch cadence per plan tier | Same cadence for all $49 users vs. tiered by volume | Product | Before dev | Pending |
| Mention retention policy | How long do fetched mentions stay in DB? (30 days? 90 days? Unlimited?) | Product + Engineering | Before dev | Pending |

---

## 10. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Platform API rate limits cause missed mentions | High | High | Per-platform rate limit management; queue with backoff; show "data may be partial" indicator when limits hit |
| AI tagging accuracy is too low, feed feels noisy | Medium | High | Target >85% precision; add "Mark as irrelevant" feedback loop to improve model |
| Twitter/X API cost spikes due to volume | Medium | High | Cap results per keyword per run; use search/recent endpoint efficiently; monitor spend |
| Users create too many broad keywords, overwhelming the system | Medium | Medium | Keyword count limits (BR-3); onboarding guidance ("Be specific, not broad"); enforce per-workspace topic cap (BR-2) |
| LinkedIn mentions largely unavailable via public API | High | Medium | Set user expectations clearly in platform selection UI ("LinkedIn shows connected account mentions only") |
| Low trial-to-paid conversion | Medium | High | Show value in first session via 7-day backfill; send "You got X mentions" email on day 1 |
| Mention deduplication failure causes duplicate cards | Low | Medium | Unique index on post_id + platform in DB; dedup before insert |

---

## 11. Dependencies

**Internal:**
- AI agents platform (`contentstudio-ai-agents`) — must support batch mention classification endpoint for relevance scoring, AI tagging, and sentiment analysis
- Notification infrastructure — email and in-app channels already exist; Slack channel requires new integration work
- Social platform integrations (`SocialIntegrations` model) — existing OAuth tokens used for platforms that require authenticated access (LinkedIn, Instagram)
- Billing system — add-on activation, trial management, and expiry enforcement

**External:**
- Twitter/X API v2 (search/recent endpoint) — requires paid API tier for production volumes
- Reddit API — limited official search; may require third-party data provider
- Platform-specific rate limits and terms of service compliance for each of the 18 monitored sources

**Blockers:**
- AI classification service endpoint must be scoped and ready before mention pipeline can be tested end-to-end
- Reddit data strategy (official API vs. third-party) must be decided before sprint 1

---

## 12. Appendix

- Research & Competitor Analysis: `docs/features/social-mentions/01-research.md`
- Workflow Design: `docs/features/social-mentions/02-workflow.md`
- Primary Reference: Octolens (https://octolens.com) — full product docs reviewed
- V1 Social Listening Prototype: `cs-prototypes/app/features/listening/` (do not modify)
- V2 Social Mentions Prototype (new): `cs-prototypes/app/listening2/` (planned)

---

## Changelog

| Date | Author | Changes |
|---|---|---|
| 2026-03-05 | Product Team | Initial draft |
