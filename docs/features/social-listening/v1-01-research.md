# Social Mentions — Research & Analysis

## Feature Overview

**Social Mentions** is a keyword-based, feed-first social listening tool built into ContentStudio. Users track brand names, competitor names, and industry terms across social platforms and the web. Matching posts appear in a real-time mentions feed. Saved filter presets ("Views") let users organize mentions by topic, intent, or sentiment. Alerts can be triggered from any View.

**Why now:** Management has validated the concept with a prototype but redirected scope — away from per-topic analytics dashboards, toward a simple, high-value feed experience. Primary inspiration is Octolens (a B2B-focused social listening tool), but ContentStudio goes further in platform coverage, alert depth, and publishing integration.

---

## Primary Reference: Octolens

Octolens is the closest model for what ContentStudio is building. Key design patterns worth adopting:

### Onboarding
- User enters company website URL
- AI analyzes site → extracts brand name, description, competitors
- Suggests 4–6 keywords automatically: own brand, competitor names, industry terms
- Each keyword pre-loaded with platform selection and AI context hint
- One-step confirm → land in feed

### Keyword Settings (per-keyword)
| Setting | Description |
|---|---|
| Platforms | Which platforms to monitor for this keyword |
| Keyword Context | Up to 200-char AI hint ("Our brand X is a SaaS tool, ignore generic usage") |
| Include ANY OF | Posts must contain keyword + any of these extra terms |
| Include ALL OF | Posts must contain keyword + ALL of these extra terms |
| Negative Terms | Exclude posts containing these terms |
| Negative Authors | Exclude posts from these authors (exact match) |
| Wildcard Negatives | Pattern exclusions (e.g., `beta.*` matches `beta.0.1`, `beta.test`) |
| Exact Match | Standalone word only vs. appears within another string |
| Case Sensitive | Enforce uppercase/lowercase matching |

### Global Keyword Settings (workspace-level)
- Global negative terms (apply to all keywords)
- Global negative authors
- Allowed/excluded subreddits
- Excluded GitHub repositories

### Views (Saved Filter Presets)
Default views pre-created:
- **High Relevance** — Buy Intent + Own Brand Mention + Competitor negative mentions
- **Brand Monitoring** — Own Brand Mention tag
- **Brand Love** — Own Brand + Positive sentiment
- **Crisis Management** — Own Brand + Negative sentiment
- **Buy Intent** — Buy Intent tag
- **Competitor Intelligence** — Competitor Mention tag

Custom views: any combination of keyword + platform + language + tag + sentiment filters, with AND/OR operators.

### AI Relevance & Tags
Every incoming mention is evaluated by AI against company context. Tags applied:
- Own Brand Mention
- Competitor Mention
- Industry Insight
- Buy Intent
- Bug Report
- User Feedback
- Promotional Post
- Product Question
- Event
- Hiring

Irrelevant posts are marked and can be filtered out.

### Alerts
Tied to Views. Steps:
1. Select a View (filters pre-filled from that view)
2. Choose destination: Email / Slack / Webhook
3. Choose frequency: Realtime / Hourly digest / Daily digest / Weekly

### Analytics (Octolens scope)
- Total mentions over time (line or bar chart)
- Mentions by Relevance
- Mentions by Platform

### AI Summaries (Octolens Scale tier)
Weekly briefs: brand summary + competitor summary. Volume/trend by source, highlights/lowlights, sentiment change, linked source posts.

### Tracked Platforms (Octolens)
X/Twitter, Reddit, Bluesky, Podcasts, HackerNews, GitHub, DEV.to, YouTube, Newsletters (200+), Stack Overflow, TikTok, LinkedIn, News (Scale only)

---

## Where ContentStudio Wins Over Octolens

| Dimension | Octolens | ContentStudio |
|---|---|---|
| Social platforms | No Instagram, Facebook, Pinterest, Threads | All major social platforms included |
| Alert types | "New mentions" only | + Volume spike, Sentiment shift, First mention |
| Publishing integration | None | Reply to mentions using connected CS accounts |
| Team collaboration | Basic | Assign mentions, add notes, team workflows |
| AI weekly summary | Scale tier only | Included from lower tiers |
| Platform coverage | Niche (B2B developer tools focus) | Broad SMB/mid-market social media managers |

---

## Tracked Platforms (ContentStudio V1)

ContentStudio monitors all major social and web platforms from day one:

| Platform | What's Monitored |
|---|---|
| X / Twitter | Keyword in tweet or reply (no retweets) |
| Instagram | Public posts and captions mentioning keyword |
| Facebook | Public posts and page content |
| LinkedIn | Public posts mentioning keyword |
| TikTok | Video descriptions mentioning keyword |
| YouTube | Video titles and descriptions |
| Reddit | Post title, body, and comments |
| Bluesky | Posts and replies |
| Pinterest | Pin descriptions and titles |
| Threads | Public posts mentioning keyword |
| HackerNews | Posts and comments |
| GitHub | Issues mentioning keyword |
| DEV.to | Article headings and descriptions |
| Stack Overflow | Question titles and bodies |
| Podcasts | Transcript mentions |
| Newsletters | Subject and body content (200+ tracked newsletters) |
| News | Articles from major news publications |
| Blogs & Web | Articles and blog posts |

---

## Recommended Approach for ContentStudio

**Positioning:** "Octolens for social media managers" — simpler than Brandwatch/Sprout, wider platform coverage than Octolens, tightly integrated with CS publishing workflow.

**V1 Scope:**
- Keyword tracking across all 18 platforms above
- AI tagging (10 tag types) + sentiment analysis
- Feed with Views (6 default + unlimited custom)
- Alerts: email + Slack + webhook, 4 trigger types (new mentions, volume spike, sentiment shift, first mention)
- AI-powered onboarding from company website URL
- Analytics: mentions over time, keyword comparison, platform breakdown, sentiment trend, AI tag breakdown
- Bookmarks

**V2 (post-launch):**
- Reply from feed using connected CS account
- Team features: assign mention to teammate, add internal note
- AI weekly summaries (brand + competitor brief)
- Bulk actions on mentions

**Pricing:** $49/month add-on (aligns with Octolens $49–$79/month range).
