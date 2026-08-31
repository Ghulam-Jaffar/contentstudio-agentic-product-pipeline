# Telegram Analytics — Research Report

**Feature:** Telegram Analytics for ContentStudio
**Date:** August 27, 2026
**Pipeline Step:** 01 — Research

> **Prior art:** Telegram *publishing* already shipped (`docs/features/telegram-integration/`, April 2026) and is live in the codebase. That epic explicitly deferred analytics to "Phase 2" and assumed it would be easy. This research revisits that assumption against the current Telegram API — and it does not hold.

---

## Executive summary — read this first

**Telegram does not expose post views or channel statistics to bots. Not in any version, including the current one.**

ContentStudio's Telegram integration is bot-token based. Bot API **10.3 (24 Aug 2026)** still has no statistics method, and the `Message` object still carries no `views` or `forwards` field. Every metric a Telegram channel admin actually cares about — views per post, forwards/shares, subscriber growth curve, traffic sources, mute rate — lives behind the **MTProto client API**, which requires a *human user session* (phone-number login), not a bot.

That forces a product decision before anything else:

| Path | What we can show | What it costs us |
|---|---|---|
| **A. Bot-API only** (recommended for v1) | Subscribers, reactions, comments, link clicks, posting behaviour, top posts | No new auth, no new cost. **No views.** |
| **B. MTProto user session** | Everything Telegram's native stats screen shows, including views | Phone login per workspace, full-account-access session storage, ban risk, ToS exposure |
| **C. Third-party data provider (TGStat etc.)** | Views + forwards for **public** channels only | Per-call cost, RU-centric coverage, private channels excluded, ToS problem sharing customer data |
| **D. Hybrid** | A now, B as opt-in later | Two auth models to maintain |

**Decision (locked by the PO, 27 Aug 2026): path A — Bot-API only for v1.** *(One late addition: Telegram's terms restrict AI/ML use of platform data — see §3a-bis. This does not affect the path choice, but it does put the dashboard's AI insights section under legal review.)* The Telegram dashboard ships without view counts, with an explicit in-product explanation of why. Rationale in §7.

The second finding that shapes scope: **we receive no ongoing Telegram updates at all today** — no webhook route, no scheduled job, no worker. The one `getUpdates` call in the codebase fires on demand during the connect flow and nothing more. Reactions, comments, and native (non-ContentStudio) posts all arrive as *updates*, so building a receiving path is the single biggest engineering item in this feature — and registering a webhook would break those two on-demand connect-flow calls (§11a).

---

# Part A: Competitor & Industry Research

## 1. What Is This Feature?

Telegram Analytics means giving a ContentStudio user a per-channel performance dashboard for their connected Telegram channels and groups — sitting alongside the existing Facebook, Instagram, LinkedIn, X, TikTok, YouTube, Pinterest and GBP dashboards — so they can see audience growth, post performance, and engagement without leaving ContentStudio for Telegram's own app.

**Why users want it:**

- Telegram publishing shipped in April 2026. Users who connected channels can schedule to them but get **zero performance feedback** in ContentStudio. Every other network in the product has a dashboard; Telegram is the only publish-but-no-analytics network alongside Bluesky and Threads.
- Telegram's native statistics screen is **mobile/desktop-app only, admin-only, single-channel, and non-exportable**. There is no way to put Telegram numbers into a client report next to Instagram and LinkedIn numbers. Agencies feel this acutely.
- Channels are a broadcast medium with no algorithmic suppression, so admins optimise on a small number of levers — post timing, format, and hook. They need historical comparison to do that, and Telegram's native screen gives a fixed rolling window with no cross-period comparison.
- ContentStudio's scheduled-report and Looker Studio exports currently silently omit Telegram, which makes a multi-network report look incomplete.

**Core use cases:**

1. A media publisher comparing which daily digest formats hold subscribers vs. which cause unsubscribes.
2. An e-commerce brand checking whether the promo post that ran Tuesday actually beat the one that ran Sunday.
3. An agency putting Telegram performance into the monthly PDF client report next to the other networks.
4. A community manager tracking whether a linked discussion group is getting comment activity on channel posts.
5. A growth marketer measuring the subscriber effect of a cross-promotion or a Telegram Ads campaign.

---

## 2. Competitor Analysis Table

Telegram *analytics* is a much thinner market than Telegram *publishing*. Of the ten standard ContentStudio competitors, **not one offers Telegram analytics**. The only tools that do are the low-cost broad-platform schedulers and the dedicated Telegram-native analytics vendors.

### 2a. ContentStudio's standard competitor set

| Competitor | Telegram publishing? | Telegram analytics? | Notes |
|---|---|---|---|
| **Buffer** | No | **No** | No native Telegram at all; Zapier only |
| **Hootsuite** | No | **No** | No native Telegram; notable gap at ~$99/mo |
| **Publer** | Yes — full | **Partial — the only real one** | Markets "Post Stats" (views, engagement, reach) and "Ad Analytics" for Telegram, plus fake-subscriber detection. Depth strongly implies a non-Bot-API data source (MTProto or a data vendor) |
| **Later** | No | **No** | Visual-first; Telegram not strategic |
| **Sprout Social** | No | **No** | No native Telegram at ~$199/mo entry |
| **Loomly** | No | **No** | Zapier only |
| **Sendible** | No | **No** | Agency-focused but no Telegram |
| **SocialBee** | Semi-manual (push reminder) | **No** | Universal Posting workaround; no data comes back |
| **Agorapulse** | No | **No** | Third-party automation only |
| **Metricool** | No | **No** | Confirmed Aug 2026: connects 8–10 networks, Telegram is not among them, despite being the most analytics-led competitor in the set |

### 2b. Adjacent SMM tools that do support Telegram

| Tool | Telegram analytics offering | Approach |
|---|---|---|
| **Postly** | Publishing-focused; analytics is thin | Bot-based; strongest on Telegram-native *publish* controls (silent mode, inline buttons, content protection), not measurement |
| **Nuelink** | Generic dashboard applied to Telegram — engagement, reach, growth, post performance | Unclear data source; likely limited to what the bot sees |
| **OnlySocial** | Markets "real-time analytics", post engagement, audience demographics for Telegram | Marketing claims outrun what Bot API can supply; treat with scepticism |
| **Zoho Social** | Inbox only, no analytics | Telegram in unified inbox since 2024 |

### 2c. The Telegram-native analytics market — and the split that decides our scope

The dedicated Telegram analytics vendors are the real reference set, and they divide cleanly into **three camps by data source**. That split maps exactly onto what ContentStudio can and cannot build, because our access model is fixed: an **admin bot**, on channels and groups the customer already connected.

#### Camp 1 — Public-channel crawlers *(we cannot replicate this)*

These index public Telegram at scale and never touch the admin's account. They are the only reason anyone sees "views" in a third-party tool.

| Tool | What it offers | Scale / pricing | Scope |
|---|---|---|---|
| **TGStat** | Subscriber growth, views, ER, reach, content performance, ad efficiency, mentions, rankings, citation index, CSV export, paid JSON API | 3M+ channels, 62B+ posts; free + paid | Public channels & groups |
| **Telemetr.io** | Channel analytics, ad-creative intelligence, keyword monitoring, channel rankings, post search, advertiser research | 11M+ channels, 72B+ posts indexed | Public channels |
| **Popsters** | Multi-channel content analysis, performance dynamics, engagement, competitor comparison, report generation — explicitly **no admin access required** | Limited free tier | Public channels |
| **LiveDune** | Average post reach, engagement trends, optimal post times, competitor analysis, user insights | Paid, higher tier for advanced | Public channels & chats |
| **Telega.io** | Engagement, reach, member growth, content performance — built around its ad marketplace | Limited free tier | Public channels & groups |
| **Zelkaa** | Member count, visits, comments, post performance, content analysis | Not stated | Public channels |
| **Brand24** | Mention monitoring, sentiment, reputation, competitor analysis | Paid, higher tier for advanced | Public mentions |

**Why we can't:** replicating this means building or buying a Telegram-wide crawler. It is a *competitive-intelligence* product, not first-party analytics — it cannot see private channels, and its numbers are crawled estimates that will not reconcile with what the admin sees in Telegram. Buying it (path C in §3c) inherits the per-call cost and ToS problems already covered.

*Adjacent note:* Brand24-style public-mention monitoring is the one item here that maps to an existing ContentStudio surface — **Social Listening** — rather than to analytics. Telegram is **not** currently a listening source (the listening service covers 18 sources; Telegram is not among them). That is a separate feature, not part of this one.

#### Camp 2 — Admin-bot instrumentation *(this is exactly our access model)*

These add a bot to a group, and report on what the bot can legitimately observe. **This is the camp ContentStudio belongs to**, and it is proof that a genuinely useful bot-based Telegram analytics product exists.

| Tool | What it offers | Pricing | Scope |
|---|---|---|---|
| **Combot** | Message frequency, most active members, peak activity hours, per-user activity scoring, chat growth over time, content trends, anti-spam, custom commands | Free under 200 members; paid above | **Groups only** |
| **TeleMe** | Member management, engagement tracking, group activity reports, anti-spam | Free tier for every group with the bot added | **Groups only** |
| **Metricgram** | Group analytics plus monetisation, AI chatbot, gamification, member permissions, full community management | $23–$116/mo, 5-day trial | **Groups only** |
| **MnD Analytics** | Community health visualisation, sentiment analysis, community scoring, recommendations | Not stated | **Groups only** |
| **InviteMember** | Subscription management, user interaction tracking, payment history, data export | Not stated | **Groups only** (paid access) |

**The finding that matters:** every mature bot-based Telegram analytics tool is a **group/community** tool. Not one of them does channel analytics — because a bot in a channel sees posts and reactions but *not views*, and views are the only metric channel admins rank on. The market has already discovered exactly the constraint in §3a and routed around it by going after groups instead.

#### Camp 3 — Support / ticket analytics *(different product surface)*

| Tool | What it offers | Maps to |
|---|---|---|
| **Mava** | Total support requests, % tickets resolved, median first-response time, median resolution time, CSAT, team analytics — across Telegram and other channels | ContentStudio's **Social Inbox**, not analytics |

### 2d. Mapped against what ContentStudio can actually build

Every metric the three camps offer, scored against a bot-token connection:

| Metric | Who offers it | Can ContentStudio offer it? | How |
|---|---|---|---|
| Subscriber / member count | All camps | ✅ **Yes** | `getChatMemberCount` |
| Subscriber growth over time | TGStat, Telemetr, Combot, Telega.io | ✅ **Yes, forward-only** | Our own daily snapshots; no backfill, no join/leave split |
| Posts published + cadence | TGStat, Popsters, Telemetr | ✅ **Yes** | `channel_post` updates + our own publishing records |
| Post reactions (by emoji) | Rarely broken out anywhere | ✅ **Yes — and better than most** | `message_reaction_count` |
| Comments per post | Zelkaa, group tools | ✅ **Yes, conditionally** | Linked discussion group; needs bot admin there + privacy mode off |
| Link clicks per post | Nobody | ✅ **Yes — unique to us** | Existing `LinkShortener` (Replug / Cstu / Firebase) |
| Top / least performing posts | TGStat, Popsters, LiveDune | ✅ **Yes** | Derived from reactions + comments + clicks |
| Best time to post | LiveDune, Combot | ✅ **Yes** | Derived from our own engagement history |
| Cross-network reporting (PDF, Looker) | **Nobody** | ✅ **Yes — our structural advantage** | Existing reports pipeline |
| Group member activity / top contributors | Combot, TeleMe, Metricgram | ⚠️ **Possible, groups only** | Bot must be admin with privacy mode off; see open question 6 |
| Peak activity hours (groups) | Combot, Metricgram | ⚠️ **Possible, groups only** | Same |
| **Post views** | TGStat, Telemetr, Popsters, LiveDune, Publer | ❌ **No** | Bot API exposes nothing; crawler or MTProto only |
| **Forwards / shares** | TGStat, Telemetr | ❌ **No** | Same |
| **Reach / ERR (views ÷ subs)** | TGStat, LiveDune, Telega.io | ❌ **No** | Derived from views |
| **Subscriber traffic sources** | TGStat, Telemetr | ❌ **No** | MTProto only |
| **Join / leave churn split** | TGStat, Combot (groups) | ❌ **Channels: no** | Channel member events are not delivered to bots |
| **Competitor channel comparison** | TGStat, Telemetr, Popsters, LiveDune | ❌ **No** | Requires crawled public data |
| **Sentiment analysis** | TGStat, MnD, Brand24 | ⚠️ **Deferrable** | Would need comment text + AI; possible later via existing AI insights |
| **Ad / promo effectiveness** | TGStat, Telega.io, Publer | ❌ **No** | Requires views + traffic sources |
| **Fake subscriber detection** | TGStat, Publer | ❌ **No** | Requires crawled audience data |
| **Moderation / anti-spam** | Combot, TeleMe, Metricgram | ❌ **Out of scope** | Not an analytics feature |
| **Support ticket metrics** | Mava | ❌ **Different surface** | Belongs to Social Inbox, not analytics |

**Net:** roughly **nine of the twenty-one** metrics the Telegram analytics market offers are reachable on our current connection, plus two group-only extras. The nine include one metric — **link clicks** — that nobody else offers at all, and one structural capability — **Telegram inside a multi-network client report** — that no tool in any camp has.

**Read of the market:** the analytics-led competitors we benchmark against have *no* Telegram story at all. The vendors that do have one have already split themselves along the exact constraint we face: if you want views, you crawl public Telegram; if you only have an admin bot, you go after groups. ContentStudio sits in the bot camp — but with two things nobody in that camp has: **first-party publishing data** (we know which post is ours, what it contained, and what links it carried) and **a multi-network reporting pipeline** to put Telegram next to Instagram and LinkedIn. That combination, not view counts, is where the win is.

---

## 3. Telegram's analytics surfaces and what each one actually gives us

This section is the decisive input to scope. Telegram has **three** distinct data surfaces with three different access models.

### 3a. Bot API — what we already have access to

ContentStudio's Telegram accounts each carry a **user-supplied bot token**, and the bot is an administrator of the channel or group. Verified against Bot API **10.3, released 24 Aug 2026**:

**What the Bot API gives us:**

| Capability | Method / update | What we get |
|---|---|---|
| Subscriber / member count | `getChatMemberCount` | A single integer, **point-in-time snapshot**. No history, no backfill. |
| Channel metadata | `getChat` → `ChatFullInfo` | Title, username, description, photo, `linked_chat_id` (tells us whether a discussion group exists) |
| Every post published to the channel | `channel_post` / `edited_channel_post` updates | Bot admins receive these for **all** posts, including ones published natively in the Telegram app outside ContentStudio. Only from the moment we start consuming updates — no backfill. |
| Aggregate reactions per post | `message_reaction_count` update (added Bot API 7.0, Dec 2023) | Total count per emoji, **anonymous** in channels. Bot must be admin **and** must list `message_reaction_count` in `allowed_updates`. Updates are batched and can lag by minutes. |
| Comment / reply activity | `message` updates in the linked discussion group | Channel posts auto-forward into the linked group with `is_automatic_forward: true` and `forward_from_message_id` = the channel message id; replies to that thread are the comments. Requires the bot to also be admin **in the discussion group** with **privacy mode disabled**. |
| Boosts (per user) | `getUserChatBoosts` | Only answers "is this specific user boosting" — not an aggregate boost count. Low product value. |

**What the Bot API does NOT give us — confirmed, not an oversight:**

- ❌ **Post views.** No `views` field on `Message`, no method. This is the headline Telegram metric and we cannot get it.
- ❌ **Forwards / shares per post.**
- ❌ **Subscriber growth over time**, join/leave events, or churn.
- ❌ **Traffic sources** (where new subscribers came from: search, links, channels, ads).
- ❌ **Notification/mute rate**, top hours, subscriber languages.
- ❌ **Any historical backfill whatsoever** — every series starts the day we begin collecting.

There is no statistics method in the Bot API. `getChatStatisticsURL` exists only in **TDLib** (the client library), is restricted to the channel owner, and returns a link to a rendered stats page rather than machine-readable data.

### 3a-bis. A terms-of-service constraint we missed: AI and machine learning

Telegram's terms restrict what may be done with platform data and **artificial intelligence**, and the clause is broad. It appears in both the API Terms of Service and the Bot Platform Developer Terms:

> *"...prohibited from using, accessing or aggregating data obtained from the Telegram platform to train, fine-tune or otherwise engage in the development, enhancement or deployment of artificial intelligence, machine learning models and similar technologies."*

The Bot developer terms name as always-prohibited *"any form of data collection aimed at creating large datasets, machine learning models and AI products, such as scraping public group or channel contents."*

**There is a consent exception**, and it is narrow: permitted *"where all relevant users individually provide explicit, informed, affirmative and continued consent that is strictly limited to the specific content and chat, channel, or non-global context window for which it was requested"* — and consent in one context is explicitly non-transferable.

**Why this matters to us:** every other network dashboard in ContentStudio has an **AI insights** section, and the PRD carries one for Telegram. What we would be doing — read-time inference over one customer's own aggregate channel figures, shown back to that same customer — is a long way from dataset-building or model training, and arguably sits inside the consent exception. But "or deployment of artificial intelligence" is drafted broadly enough that this is a legal judgement, not an engineering one.

**Recommendation:** treat AI insights for Telegram as **blocked pending legal review**, not as a normal P1. It is the one part of the dashboard with terms-of-service exposure, and it is cheap to defer — everything else in the dashboard is unaffected. Note also that this constraint applies regardless of which data path we take: it is not a reason to prefer or avoid MTProto.

### 3b. MTProto client API — where the real numbers live

`stats.getBroadcastStats` and `stats.getMessageStats` return exactly what Telegram's in-app Statistics screen shows: period, followers, **views per post**, **shares per post**, reactions per post, plus graphs for growth, followers, mute, top hours, interactions, **views by source**, **new followers by source**, languages, and reactions by emotion. TDLib exposes the same through `getChatStatistics`.

**The catch, and it is a big one:** MTProto requires an authorised **user** session — an `api_id`/`api_hash` from my.telegram.org plus a phone-number login for a human who administers the channel. A bot token cannot call these methods.

Consequences if we went this way:

- **Auth UX:** every workspace would have to log a real Telegram account in — phone number, SMS/app code, and 2FA password if set. This is a materially heavier and scarier connection flow than anything else in ContentStudio.
- **Security and liability:** an MTProto session string is *full account access* — read all chats, send as the user, and it survives password changes until explicitly revoked. Storing those for customers is a different risk class from storing an OAuth token, and it invites a breach headline.
- **Ban risk:** Telegram places unofficial clients using third-party `api_id`s under automatic observation. Unusual traffic patterns get flagged, and a flagged account is *the customer's personal Telegram account*, not ours.
- **ToS:** the API Terms bar disclosing data obtained through Telegram to third parties without user authorisation, and treat artificial counter inflation and flood patterns as permanent-ban triggers. A compliant implementation is possible but needs deliberate rate discipline and legal review.
- **Coverage floor:** Telegram only enables channel statistics above a subscriber threshold (**~50 subscribers** currently; it was 500, and earlier 1,000). Below that, the API returns nothing useful — so small channels see an empty dashboard regardless.

### 3c. Third-party data vendors (TGStat and similar)

TGStat sells a JSON API over crawled public-channel data — subscriber history, views, ER, post data, mentions, rankings.

- ✅ Gives views and forwards without any new auth from the user.
- ❌ **Public channels only.** Private channels — a large share of the brand/community use case — are invisible.
- ❌ Crawled and sampled, so numbers will not reconcile with what the admin sees in Telegram's own screen. Explaining a discrepancy to a customer is worse than showing nothing.
- ❌ Per-call cost with RU-centric coverage bias.
- ❌ Sending customer channel identifiers to an external analytics vendor needs a ToS/privacy review, and Telegram's API terms are explicit about third-party disclosure.

**Precedent to apply here:** ContentStudio's standing rule on paid platform data (established for X) is that **per-call API cost is passed to the customer, never absorbed** — see `docs/features/x-analytics-metering/` and `docs/features/x-pay-per-use-credits/`. If we ever take path C, it inherits that rule and needs the same metering/credit machinery. That is a large amount of work to bolt onto a network where the alternative is free.

---

## 4. Common Patterns Across Tools

1. **Nobody in our competitive set does this.** There is no established SMM-tool pattern to copy — the patterns come from Telegram-native analytics products.
2. **Views are the currency.** Every Telegram-native analytics product leads with views and view-rate (views ÷ subscribers, "ERR"). Admins do not think in impressions or reach; they think in views per post.
3. **Public-channel crawling is the standard workaround** for tools that show views without a user session. It is why those tools cannot handle private channels.
4. **Subscriber growth is shown as a cumulative curve with join/leave breakdown.** Bot-API-only tools can only show the cumulative curve from their own polling, never the join/leave split.
5. **Reaction breakdown by emoji is a Telegram-specific expectation.** 👍/❤️/🔥 distribution is treated as a first-class metric, not a footnote — this one *is* reachable via Bot API.
6. **Comment counts come from the linked discussion group**, and every tool that shows them requires being admin in both.
7. **Honest empty states are rare and valued.** The tools that market "real-time Telegram analytics" without a user session are overstating what they have; reviewers notice.

---

## 5. Differentiators Worth Considering

1. **Be the only mid-market SMM tool with a Telegram dashboard at all.** Buffer, Hootsuite, Sprout, Metricool, Later, Sendible, Loomly, Agorapulse — none of them have one. This is a straight competitive checkbox win.
2. **Reaction-emoji breakdown as a headline chart.** Telegram is one of very few networks with a rich, per-emoji reaction model, and it maps to sentiment far better than a like count. Nobody in our set visualises it.
3. **Native-post capture.** Because a bot admin receives `channel_post` for *every* post, ContentStudio can report on posts the user published directly in Telegram — not just ones scheduled through us. Most networks make us choose between the two; here we get both for free.
4. **Comment health for linked discussion groups.** Channel + discussion-group pairing is a Telegram-specific structure, and reporting comment volume per post is genuinely useful to community managers. No competitor surfaces it.
5. **Click tracking via the existing shortener.** ContentStudio already has `LinkShortener` with Replug / Cstu / Firebase backends. Telegram gives us no click data, but our own shortener does — and click-through is arguably a better business metric than views. This turns a gap into a differentiator, and it works identically for public and private channels.
6. **Put Telegram in the multi-network report.** The moment Telegram appears in the Overview, PDF reports and Looker Studio export, agencies stop having to hand-assemble it. That is the actual agency pain.
7. **Transparent "why is this missing" UX.** Explicitly telling the user *"Telegram does not share view counts with third-party tools"* — with a link to their native stats — builds more trust than a competitor's vague dashboard. It is also the honest thing to do.

---

## 6. User Expectations

**Table stakes — what a user will assume exists on day one:**

- Telegram appears in the analytics network switcher alongside the other networks
- Subscriber count and how it has changed over the selected period
- A list of posts in the period with per-post engagement, sortable
- Top-performing and least-performing posts
- Posting-behaviour view (how much did I post, when)
- The standard analytics date-range picker, comparison period, and account switcher behave the same as every other network
- Telegram included in scheduled/exported PDF reports
- A clear empty state for a newly connected channel

**Delighters:**

- Reaction breakdown by emoji, with the distribution over time
- Comment volume per post from the linked discussion group
- Link clicks per post via the ContentStudio shortener
- Best-time-to-post derived from our own engagement history
- AI insights on Telegram performance (every other network dashboard has an `ai_insights` section)
- Channel-vs-channel comparison for users running several channels

**Will be asked for and we must have an answer ready:**

- **"Where are my view counts?"** — this *will* be the first support ticket. The in-product explanation needs to be written before launch, not after.
- "Why does my subscriber chart start on the day I connected?" — no backfill is possible.
- "Why don't my numbers match Telegram's Statistics screen?" — different data source, different definitions.

---

## 7. Recommended Approach for ContentStudio

### v1 — Bot-API-only, first-party analytics ✅ **LOCKED**

Ship a dashboard built entirely on data our existing bot connection can legitimately collect, plus data we already own:

| Section | Metric | Source |
|---|---|---|
| Summary | Subscribers, net change over period, posts published, total reactions, total comments, total link clicks | `getChatMemberCount` polling + our own collected data |
| Audience growth | Subscriber count over time, from connection date forward | Daily `getChatMemberCount` snapshot |
| Publishing behaviour | Posts per day, by content type (text / photo / video / album / document) | `channel_post` updates + our own publishing records |
| Post performance | Per-post reactions (total + per emoji), comments, link clicks | `message_reaction_count` + discussion-group replies + shortener |
| Top / least posts | Ranked by an engagement score we define and document | Derived |
| Reaction breakdown | Distribution across emoji | `message_reaction_count` |
| AI insights | Same treatment as every other network | Existing AI insights service |
| **Views** | **Not available** — replaced by an explicit, well-written explanation panel | — |

**Why this and not the alternatives:**

- It requires **no new authorisation from the user**. Every already-connected Telegram account starts producing data as soon as the update consumer ships. That is a genuinely rare property for an analytics feature.
- It carries **no per-call cost**, so no metering, no credits, no billing work — unlike the X analytics path.
- It works identically for **public and private** channels and groups, which the crawling approach does not.
- Reactions, comments and clicks are real engagement signals. A view on Telegram is closer to an impression; a reaction is closer to intent. The dashboard is thinner than Telegram's own, but the metrics on it are not junk.
- It avoids storing full-account MTProto sessions for customers, which is the single largest risk item in this whole space.

- The bot-instrumentation camp (§2c) has already validated this shape commercially — Combot, TeleMe and Metricgram run real businesses on nothing more than what an admin bot can see. Metricgram charges $23–$116/mo standalone for it.

**The cost of this choice, stated plainly:** no views, no forwards, no reach/ERR, no traffic sources, no join/leave split, no competitor comparison, and no history before the day collection starts. We must design for that rather than paper over it. §2d is the full ledger of what is in and what is out.

### v2 — Opt-in advanced statistics (path D)

If demand justifies it, add an explicitly opt-in **"Connect advanced Telegram statistics"** flow using an MTProto user session, unlocking views, forwards, subscriber-source breakdown, and the native graph set — gated behind a clear consent screen, a legal review, and a decision on whether we are willing to hold user sessions at all. This should be scoped as its own feature, not smuggled into v1.

### Explicitly not recommended

**Path C (third-party data vendor)** — the combination of public-channel-only coverage, numbers that will not reconcile with Telegram's own screen, per-call cost requiring the full X-style metering apparatus, and the ToS question around forwarding customer data makes it a poor trade against a free first-party alternative.

---

# Part B: Codebase Analysis

## 8. Existing Related Code

### 8a. Telegram integration — already shipped

Telegram publishing is live. Relevant surface:

| Layer | File |
|---|---|
| API client | `contentstudio-backend/app/Libraries/Integrations/Platforms/Social/Telegram/TelegramApiClient.php` |
| Exception type | `contentstudio-backend/app/Libraries/Integrations/Platforms/Social/Telegram/TelegramApiException.php` |
| Connector | `contentstudio-backend/app/Strategy/Integrations/TelegramConnector.php` |
| Posting strategy | `contentstudio-backend/app/Strategy/Planner/TelegramPosting.php` |
| Controller | `contentstudio-backend/app/Http/Controllers/Integrations/Platforms/Social/TelegramController.php` |
| Requests | `app/Http/Requests/Integrations/Telegram/{ValidateTelegramBot,DiscoverTelegramChats,ValidateTelegramChat,AddTelegramChats}Request.php` |
| Connect modal (FE) | `contentstudio-frontend/src/modules/integration/components/dialogs/AddTelegram.vue` |
| Connect composable (FE) | `contentstudio-frontend/src/modules/integration/composables/useAddTelegram.ts` |
| Platform visuals | `contentstudio-frontend/src/utils/platformVisuals.ts`, `src/config/platformConfigs.ts` |
| Plan gating | `contentstudio-frontend/src/modules/billing/constants/plansComparison.ts:147` |

**Two facts from this code that drive the whole feature:**

1. **Connection is per-account bot token, not a shared ContentStudio bot.** `TelegramConnector::validateToken()` and `TelegramController::validateBot()` take a `bot_token` from the user; the token is stored per social account. So each connected channel has its own bot with its own rate-limit budget and its own update stream — good for scale, but it means update-receiving must be set up *per bot*, not once globally.
2. **`TelegramApiClient` already exposes the two read methods we need** — `getChat()`, `getChatMemberCount()` — plus `getUpdates()` and a generic `request()` escape hatch. No new HTTP client work is needed for the polling half of this feature.

**What we already store per published post** (`TelegramPosting.php:311-321`): `posted_id` (message id), `platform_id` / `chat_id`, `message_ids[]` for albums, and `posted_uri` (the `t.me` permalink). Post identity for ContentStudio-published posts is therefore already solved.

### 8b. Analytics platform — where a new network gets added

Analytics is **not in Laravel**. It is the Go pipeline in `contentstudio-social-analytics-go/`, and the Laravel `app/Builders/Analytics/` code is legacy being migrated away from (see `docs/stories/analytics-php-to-golang-migration/`).

Per-platform ingestion is a 4–5 service chain, cleanest and smallest in Pinterest:

```
contentstudio-social-analytics-go/src/services/pinterest/
  pinterest-fetcher/              # reads accounts from MongoDB, calls the platform API, emits to Kafka
  pinterest-parser/               # normalises raw payloads
  pinterest-immediate-processor/  # post-publish immediate refresh path
  pinterest-analytics-sink/       # writes to ClickHouse
```
(LinkedIn additionally has a `linkedin-clickhouse-sink`.)

Supporting pieces per platform:

| Concern | Location |
|---|---|
| HTTP handlers | `src/api/analytics/<platform>/` |
| Route registration | `src/api/analytics/router.go` |
| Kafka message models | `src/models/kafka/<platform>.go` |
| ClickHouse table DDL | `src/deployments/clickhouse/schema/<platform>_schema.sql` |
| ClickHouse models + conversions | `src/models/db/clickhouse/<platform>.go`, `.../conversions/<platform>_clickhouse.go` |
| Read queries | `src/db/clickhouse/analytics-get-queries/<platform>/` |
| Account source | MongoDB, read directly by each fetcher (`src/db/mongodb`) |

**There is no Telegram anywhere in this repo.** The only `telegram` string matches are an unrelated YouTube sharing-service enum (`src/models/kafka/youtube.go:241,295`).

The section vocabulary a platform is expected to expose, taking LinkedIn as the reference (`router.go:54-63`): `summary`, `audienceGrowth`, `pageViews`, `publishingBehaviour`, `topPosts`, `postsPerDays`, `hashtags`, `getTopPosts`, `followersDemographics`, `ai_insights`.

### 8c. Analytics frontend

`contentstudio-frontend/src/modules/analytics/` — one module, documented in its own `CLAUDE.md`, covering per-platform dashboards, competitor analytics, and reports.

- Per-platform views: `src/modules/analytics/views/<platform>/MainComponent.vue` plus `components/` and `composables/` subfolders
- Route registration: `src/modules/analytics/config/routes/analytics.ts` — each platform is a lazy-loaded route (`path: 'pinterest/:accountId?'`, `name: 'pinterest_analytics'`)
- Dashboard shell: `components/AnalyticsMain.vue`
- Shared building blocks: `components/common/` — `AnalyticsCardWrapper.vue` (standard card + skeleton), `AnalyticsLoading.vue`, `AnalyticsSectionError.vue`, `PlatformTooltip.vue`, `SyncDateRangeModal.vue`, plus `composables/` for account selection, lazy section loading and manual sync
- Queries: `queries/useAnalyticsQueries.ts`, `queries/keys.ts` (TanStack Query)

There is **no Telegram view** — Telegram appears in the frontend only in integration, composer, planner and billing contexts.

---

## 9. Reusable Components / Services

**Reusable as-is:**

- `TelegramApiClient` — `getChat()`, `getChatMemberCount()`, and `request()` cover the polling half with no changes
- `LinkShortener` + `Integrations/Shorteners/{Replug,Cstu,Firebase}.php` — **link *creation* needs no new work; reading clicks back is new work.** See §9a
- The whole analytics-go scaffolding: Kafka client, ClickHouse client, MongoDB account reader, telemetry, the per-account semaphore/rate-limit pattern in the fetchers
- Analytics frontend shell: `AnalyticsMain.vue`, `AnalyticsCardWrapper.vue`, the loading/error/empty components, date-range picker, account selector, manual-sync composable
- AI insights service (`src/services/ai/`) — every platform dashboard has an `ai_insights` section and Telegram should follow
- Reports/PDF and Looker Studio export paths — platform-agnostic once the sections exist

**Adapt, don't rewrite:**

- `services/pinterest/*` → `services/telegram/*` — Pinterest is the smallest, most recent template. The **fetcher** shape changes most: instead of one API sweep per account per window, Telegram needs a light periodic `getChatMemberCount` poll *plus* a continuous update consumer (§11)
- `api/analytics/pinterest/` → `api/analytics/telegram/` — same handler-per-section shape
- `deployments/clickhouse/schema/pinterest_schema.sql` → `telegram_schema.sql`
- `views/pinterest/MainComponent.vue` → `views/telegram/MainComponent.vue`, with Telegram-specific cards (reaction breakdown) replacing the ones we cannot populate

**Genuinely new, no precedent in the codebase:**

- **A Telegram update consumer.** Nothing like it exists for any network. Every other platform's *analytics* data is fetched on a schedule by the Go services; Telegram's engagement data can only be pushed to us. Note this is a genuinely new receiving path — Telegram itself is not polled today either (§11a).
- **A reaction-by-emoji storage model and chart.** No other network has this shape.
- **A "metric not available on this network" UI treatment.** Distinct from an empty state or an error state, and it will be reused later for Bluesky/Threads.

---

### 9a. Correction — link clicks are half-built, not built

An earlier draft of this document claimed click data needed no new work. That is wrong, and the distinction matters for scoping:

**What already exists.** ContentStudio shortens links at **composition time**, not at publish time — `LinkShortener::fetchShortLinks()` is called from `HelperController`, `EvergreenBulkBuilder` and `CSVBuilder`, so the shortened URL is already sitting in the post body before any platform's posting strategy runs. Telegram therefore inherits shortening for free: `TelegramPosting` just sends whatever text the plan holds. **No Telegram-specific link work is needed.**

**What does not exist.** Nothing in ContentStudio reads click counts *back*. `Cstu` exposes only `fetchShortLinks`, `fetchLongLinks` and `getShortLink` — there is no click-retrieval method on any of the three shorteners, and the Go analytics service ingests no shortener data at all (its only `clicks` references are platform-native metrics: Facebook `post_clicks`, LinkedIn share statistics, Google Ads). So per-post click counts require a **new click-retrieval and attribution layer**.

**A third constraint.** Shortening is **opt-in at composition**. A post published without a shortened link has no click data at all — so the metric is sparse by nature, not merely delayed. The dashboard must treat "no shortened links in this post" as a distinct state, not as zero clicks.

**Scoping decision (28 Aug 2026): keep it, and promote it to P0.** Link clicks is the one metric on the dashboard that is neither free nor Telegram-constrained — ordinary ContentStudio work that happens to be new. The case for keeping it is that without views, the remaining dashboard is subscribers, posts, reactions and comments; comments are unavailable for any channel with no linked discussion group, which leaves reactions as the sole engagement signal. **Link clicks is then the only business-outcome metric on the dashboard, and per §2d the only metric no competitor offers.** It also directly offsets the highest-likelihood risk in the PRD — that the dashboard reads as incomplete because views are absent.

Two design requirements follow from its sparseness: a post with no shortened link must show **"no tracked links"** rather than zero, and any click total must be shown **with the number of posts it covers**. If scope has to give somewhere, **comments** is the cheaper cut — it is P1 and unavailable for most channels regardless.

---

## 10. Integration Points

1. **Analytics network switcher** — Telegram tile/route added to `analytics.ts` routes and the account selector; accounts come from the same MongoDB social-accounts source the other fetchers read.
2. **New `telegram` ingestion services** in `contentstudio-social-analytics-go/src/services/telegram/` following the Pinterest chain.
3. **New handler package + routes** at `src/api/analytics/telegram/` registered in `router.go`, exposing `summary`, `audienceGrowth`, `publishingBehaviour`, `topPosts`, `getTopPosts`, `ai_insights`, and a Telegram-specific `reactions` section. `pageViews`, `hashtags` and `followersDemographics` have no Telegram equivalent and should be omitted rather than stubbed.
4. **New ClickHouse tables** for channel snapshots, post inventory, and reaction/comment/click counters.
5. **A new update-receiving endpoint in the Laravel backend** — Telegram webhooks must terminate somewhere that holds the bot tokens, and that is `contentstudio-backend`. Received updates get normalised and pushed onto Kafka for the Go pipeline. This is a **new** route; none exists today.
6. **Post inventory** — reconcile `channel_post` updates against `posted_id` / `posted_meta.message_id` on our own plans so ContentStudio-published posts and natively-published posts land in one list.
7. **Link shortener** — attribute click counts to the Telegram post that carried the link.
8. **Reports** — Telegram sections in scheduled PDF reports, the cross-network Overview, and the Looker Studio connector.
9. **Billing / entitlements** — Telegram already exists in `plansComparison.ts`; confirm whether analytics access needs separate gating.

---

## 11. Technical Considerations

### 11a. The update consumer — sized

Reactions, comments and native posts are **push-only**. Telegram offers exactly two ways to receive updates, and **they are mutually exclusive per bot**:

- **`getUpdates` polling** — would need a persistent worker per bot token running on a loop. Does not scale to thousands of connected channels. Fallback, not the plan. **We do not do this today** — see below.
- **`setWebhook`** — Telegram POSTs updates to a URL we own. Scales. Registered per bot token with a per-bot secret, so we can route by token.

**What we do today: nothing continuous.** There is no scheduled task, no queued job, and no worker for Telegram anywhere in the backend. `TelegramApiClient::getUpdates()` is called **on demand during the connect flow only** — a single request, with no `timeout` parameter, so not even a long poll. It is used purely as a lookup while the user is connecting an account, and then nothing runs until the next connect.

So the accurate framing is **not** "migrate Telegram from polling to webhooks" — there is no polling to migrate. It is **"add an update-receiving path where none exists"**, and the only existing code that collides with it is two lookup calls.

**How much existing code this actually touches: two call sites in one file.**

`getUpdates` is called in exactly two places, both in `TelegramConnector.php`, and both run only while a user is connecting an account:

| Call site | Purpose | Impact of moving to webhooks |
|---|---|---|
| `discoverChats()` (line 184) | Mines recent updates to find which chats the bot has been added to, then enriches each via `getChat` / `getChatMember` / `getChatMemberCount` | **Replaced, and improved** — see below |
| `fetchChatMigrationMap()` (line 555), used by `validateChat()` | Builds a group→supergroup migration map | Migration events arrive as updates; we persist them instead of re-reading the buffer |

Everything else in the connect flow — `getMe`, `getChat`, `getChatMember`, `getChatMemberCount`, `inspectChatAccess`, and all of `validateToken` / `validateChat` / `addChats` — is unaffected. Webhook mode only changes how *updates* are delivered; every direct API call keeps working. **Publishing is entirely unaffected** (`sendMessage`, `sendPhoto`, `sendMediaGroup`, edits, deletes are all direct calls).

**The webhook is a better discovery mechanism than what we have.** `discoverChats()` currently mines `getUpdates` for `my_chat_member` events — but Telegram only retains updates for **24 hours**. If a user added the bot to their channel three days before connecting, discovery silently misses it. That is a latent limitation in the shipped flow today. A webhook consumer receives `my_chat_member` continuously and persists it, so discovery becomes "every chat this bot was ever added to" instead of "whatever happens to still be in the buffer". We are not rebuilding a working thing — we are replacing a fragile one.

**What is genuinely new work:**

- A public HTTPS webhook endpoint in `contentstudio-backend`, with per-bot secret-token validation, idempotent handling (Telegram retries), and a Kafka producer to hand updates to the Go pipeline
- Webhook lifecycle: register on connect, deregister on disconnect, re-register on token change
- A one-off backfill job calling `setWebhook` for every already-connected Telegram account — no user action required, and **no disconnection or data loss**: accounts live in MongoDB (`platform_identifier`, `access_token`, `workspace_id`), publishing reads them via `SocialRepo` and makes direct API calls, and none of that is touched by update delivery
- **Backfill safety rule:** call `getWebhookInfo` before `setWebhook` on every account. If a webhook is already registered to a non-ContentStudio URL, **do not overwrite it** — flag the account and surface it, because overwriting silently breaks whatever else the customer is running on that bot
- Persisting `my_chat_member` and migration events so discovery reads from our store, not the Telegram buffer

**The one genuinely fiddly bit — and the only real risk in the whole migration:** a bot can only have one update consumer. Because customers supply their own bot tokens, some may already use that bot elsewhere. Calling `setWebhook` would overwrite their existing webhook, or starve their polling loop — silently breaking *their* tooling, not their ContentStudio account. `getWebhookInfo` detects this before we touch anything, so it must be checked both at connect time and in the backfill, with a clear message instead of a silent takeover.

**What is *not* at risk:** no account is disconnected, no stored data is deleted, no customer action is needed, and publishing keeps working throughout. Analytics history starts at zero — but that is because Telegram never gave us reactions or views in the first place, not because anything is being lost.

**Other constraints:**
- `allowed_updates` must explicitly include `message_reaction_count`, or those updates never arrive
- Reaction updates are batched and can lag by minutes — not a real-time metric
- **No backfill.** Reactions and posts predating webhook registration are gone permanently.

### 11b. Subscriber history

`getChatMemberCount` is a snapshot. The growth chart is built from **our own periodic polling**, at whatever cadence we choose (daily is the sensible default; the existing per-account semaphore pattern in the fetchers gives us the rate discipline). Consequences: the series starts at connection/collection date, there is no join/leave split, and same-day granularity is the floor.

### 11c. Comment counting

Only works when: the channel has a linked discussion group (`ChatFullInfo.linked_chat_id`), **and** the same bot is an admin in that group, **and** privacy mode is disabled for the bot. All three are user-side configuration we cannot set for them. The dashboard needs a distinct state for "comments unavailable — no discussion group linked" versus "0 comments".

### 11d. Rate limits

Bot API: ~30 requests/sec overall, ~1 message/sec per chat. `getChatMemberCount` has no published limit but is documented as ban-triggering when hammered. With per-account bot tokens the budget is per-customer rather than pooled, which is favourable — but a daily poll across all connected channels still needs the existing semaphore/backoff pattern. `429` responses carry `Retry-After` and must be honoured.

### 11e. Data model sketch (ClickHouse)

- `telegram_channel_snapshots` — chat_id, workspace_id, subscriber_count, captured_at
- `telegram_posts` — chat_id, message_id, published_at, content_type, is_contentstudio_post, plan_id, permalink
- `telegram_post_reactions` — chat_id, message_id, emoji, count, updated_at
- `telegram_post_comments` — chat_id, message_id, comment_count, updated_at

Retention and observability should follow whatever `docs/stories/analytics-observability-and-data-retention/` establishes rather than inventing a Telegram-specific policy.

### 11f. Deliberately out of scope for a Bot-API v1

Views, forwards/shares, subscriber traffic sources, join/leave churn, mute rate, top-hours heatmap sourced from Telegram, subscriber languages, and any historical data predating collection. All of these require §3b or §3c.

---

## 12. Mobile Impact

**Cannot be assessed in this workspace.** `contentstudio-flutter/` — the single source of truth for all mobile work per the repo's `CLAUDE.md` — **is not currently mounted here**. (`contentstudio-ios-v2/` and `contentstudio-android-v2/` are present on disk as untracked directories, but they are the retired native apps and must not be used to ground new stories.)

Provisional read: Telegram analytics is a web-first feature. If the Flutter app has analytics dashboards, adding a Telegram one is a follow-on `[Flutter]` story rather than v1 scope. **This needs the Flutter repo mounted before it can be answered properly** — flagged as an open question below.

---

## 13. Open Questions for the Product Owner

**Decided 27 Aug 2026:**
- ✅ **Views** — we ship without them (path A, Bot API only), with an in-product explanation.
- ✅ **Third-party data vendors** — ruled out on cost, coverage and ToS grounds.

**Still open:**

1. **MTProto as a stated v2.** Do we keep path B alive as a future opt-in "advanced statistics" flow, or rule it out now on security/liability grounds? Storing customer Telegram *user sessions* is a materially different risk posture than anything currently in the product. This does not block v1 — it only affects what we say publicly about the roadmap.
2. **Connection-flow change.** Adding webhooks changes how the already-shipped Telegram connect flow works (§11a). Are we comfortable modifying a live integration, and does it need a migration for channels already connected?
3. **Mobile.** Should `contentstudio-flutter/` be mounted so mobile scope can be assessed, or is this explicitly web-only for now?
4. **Groups vs. channels — now a sharper question.** §2c surfaced something counterintuitive: **every mature bot-based Telegram analytics tool is a group tool, not a channel tool**, because a bot in a group can see far more than a bot in a channel. Combot, TeleMe, Metricgram and MnD all report member activity, top contributors, peak hours and message volume — none of which have a channel equivalent. So our *group* analytics could genuinely be richer than our *channel* analytics. Three options:
   - **Channels only in v1** — simplest, matches the publishing use case, defers groups
   - **Both, with a reduced section set for groups** — member count + posting behaviour only
   - **Both, with a distinct group dashboard** — adds Combot-style community metrics; more work, more differentiation, and requires privacy mode disabled on the bot
5. **Sentiment on comments.** Camp 1 and MnD both sell sentiment. We could derive it from discussion-group comment text using the existing AI insights service. v1 or defer?
6. **Bluesky and Threads precedent.** Those two are also publish-without-analytics with research stories pending (`docs/stories/analytics-network-research-bluesky-threads/`). Telegram is the first of the three to get a real dashboard — should the "network with limited analytics" UI patterns we invent here be built as reusable for those?
7. **Telegram in Social Listening.** Separate from this feature, but it came up: Telegram is not currently a listening source, and public-mention monitoring is the one Camp 1 capability that maps to an existing ContentStudio surface. Worth a backlog item?

---

## 14. Sources

- [Telegram Bot API (v10.3, 24 Aug 2026)](https://core.telegram.org/bots/api)
- [Telegram Bot API Changelog](https://core.telegram.org/bots/api-changelog)
- [Telegram Bots FAQ](https://core.telegram.org/bots/faq)
- [Telegram API Terms of Service](https://core.telegram.org/api/terms)
- [Telegram Bot Platform Developer Terms of Service](https://telegram.org/tos/bot-developers)
- [MTProto `stats.getBroadcastStats`](https://core.telegram.org/type/stats.BroadcastStats)
- [TDLib `getChatStatistics`](https://core.telegram.org/tdlib/docs/classtd_1_1td__api_1_1get_chat_statistics.html)
- [Telegram Channel statistics API docs](https://blogfork.telegram.org/api/stats)
- [Telegram Discussion groups](https://core.telegram.org/api/discussion)
- [Telegram Message reactions](https://core.telegram.org/api/reactions)
- [Telegram Channels FAQ](https://telegram.org/faq_channels)
- [python-telegram-bot — MessageReactionCountUpdated](https://docs.python-telegram-bot.org/en/v21.7/telegram.messagereactioncountupdated.html)
- [grammY — Reactions guide](https://grammy.dev/guide/reactions)
- [Obtaining a Telegram api_id](https://core.telegram.org/api/obtaining_api_id)
- [TGStat Statistics API](https://tgstat.ru/en/api/stat)
- [TGStat analytics](https://tgstat.com/analytics)
- [12 Best Telegram Analytics Tools — MetaCRM](https://www.metacrm.inc/blog/12-best-telegram-analytics-tools-native-dashboard-more)
- [Telegram Channel Analytics: Best Free & Paid Tools for 2026 — Collaborator](https://collaborator.pro/blog/telegram-analytics-tools)
- [Best Telegram Analytics Tools — MangoAds](https://mangoads.com/blog/for-channel-owners/best-telegram-analytics-tools)
- [Telegram analytics tools comparison — CRMChat](https://crmchat.ai/blog/telegram-analytics-tools-comparison)
- [Built-in statistics for Telegram channels — Onlypult](https://onlypult.com/blog/built-in-statistics-for-telegram)
- [Telegram Channel Statistics: Where to Find Them and What They Miss — TGuard](https://tguard.pro/en/blog/telegram-subscription-analytics/)
- [What is Metricool — supported networks](https://metricool.com/what-is-metricool/)
- [Social media tools with Telegram support — Nuelink](https://blog.nuelink.com/social-media-tools-with-telegram-support/)
- [Telegram Scheduler — OnlySocial](https://onlysocial.io/platforms/telegram/)
- [How to get post view count — Telegram.Bot issue #327](https://github.com/TelegramBots/telegram.bot/issues/327)
- [Top 10 Tools to Track Telegram Group or Channel Analytics — Mava](https://www.mava.app/blog/top-10-tools-to-track-telegram-group-or-channel-analytics-a-comprehensive-comparison)
- [Best Combot Alternative for Telegram Groups — Metricgram](https://metricgram.com/alternatives/combot)
- [Best Telemetr.io Alternative for Telegram Groups — Metricgram](https://metricgram.com/alternatives/telemetr)
- [Telegram Group Analytics & Metrics — Metricgram](https://metricgram.com/telegram-group-analytics)
- [Metricgram pricing](https://metricgram.com/pricing)
- [Best Telegram Group Management Tools Compared (2026) — Metricgram](https://metricgram.com/blog/telegram-group-management-tools-compared)
- [TeleMe pricing](https://teleme.io/pricing)
- [Combot bot directory entry](https://tdirectory.me/bot/combot.dhtml)
- [Telemetr.io — Bellingcat Online Investigation Toolkit](https://bellingcat.gitbook.io/toolkit/more/all-tools/telemetrio)
- [Statistics and analytics on Telegram — Popsters](https://popsters.com/blog/post/statistics-and-analytics-on-telegram)
- [13 Best Telegram Marketing Tools — SendPulse](https://sendpulse.com/blog/telegram-marketing-tools)
- [Best Telegram channel monitoring tool — Statiko](https://statiko.io/blog/best-telegram-channel-monitoring-tool)
