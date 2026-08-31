# Design prompt — X (Twitter) Competitor Analytics mockups

> Paste everything below the line into a fresh Claude conversation. It is written to be self-contained — the receiving session has no context on ContentStudio.

---

Build me a **single self-contained HTML artifact**: a high-fidelity UI mockup deck for a new feature called **X (Twitter) Competitor Analytics** in a product called **ContentStudio**.

This is a **review artifact for a Product Owner**, not a working app. It should read as a designed spec document: framed product screens with short annotations explaining each decision, so a non-technical stakeholder can approve the direction before engineering starts.

## 1. Product context you need

**ContentStudio** is a social media management platform (scheduling, publishing, inbox, analytics). Inside its **Analytics** module there is a **Competitors** section where a user builds a *report* — a saved set of up to **5 social profiles** — and compares them side by side.

Competitor reports exist today for **Facebook** and **Instagram**. We are adding **X (Twitter)** as a third platform. The report is normally set up as *your own brand + up to 4 rivals*, all compared in one view.

**The one thing that makes X different, and which the whole design has to solve:**

Facebook and Instagram competitor data is free for us to fetch, so those reports refresh automatically every day. **X charges us per piece of data read.** Reading one competitor's recent posts costs real money, every single time. So the X report **cannot auto-refresh** — the user has to deliberately trigger a sync, and they must see what it will cost **before** they spend.

ContentStudio already has this pattern elsewhere: X publishing and X account analytics are already metered. Users are on one of two payment models:

- **X Wallet** (newer workspaces) — a prepaid dollar balance. Costs shown in **$**.
- **X Credits** (long-standing workspaces) — a monthly credit allowance shared with X publishing. Costs shown in **credits**.

Your mockups must show **both**, because the PO needs to see that the experience works for each.

## 2. What to produce

One HTML file containing a vertically scrolling review page, roughly **1080px content width**, with:

- A page header: eyebrow label `DESIGN REVIEW`, title **"X Competitor Analytics"**, a one-paragraph lede, and small pill tags for context (e.g. `Analytics → Competitors`, `New platform: X`, `Metered — pay per sync`, `Desktop web`).
- A short **legend** explaining any visual conventions you use (e.g. what a callout marker means).
- **Numbered sections**, one per screen below. Each section = a kicker label, an `<h2>`, a one-or-two-sentence subtitle explaining *why* the screen looks like this, then the framed screen.
- Each product screen sits inside a **browser chrome frame** (three dots + a fake URL bar showing e.g. `app.contentstudio.io/analytics/competitors/x`).
- Where a screen has multiple states, put a **segmented control above the frame** that swaps between them with a little JavaScript (e.g. `Wallet cohort` / `Credits cohort`, or `Valid handle` / `Not found` / `Private account`).

## 3. Visual language (match our existing mockup decks)

Use these exact tokens so this deck sits alongside our previous ones:

**Page shell:** background `#e9edf3`, panels `#ffffff`, ink `#18212e`, muted text `#5a6575`, borders `#d7dee8`, accent `#157fff`.
**Inside the product screens:** app background `#f4f6fa`, surface `#ffffff`, ink `#17202e`, secondary ink `#3c4658`, muted `#6b7688`, borders `#e7ebf2`, accent `#157fff`, accent hover `#0e63cc`, accent tint `#eaf3ff`, success `#15934f` on `#e8f6ed`, warning `#b56a09` on `#fdf2e1`, danger `#d24343` on `#fdecec`, **X brand black `#0f1419`**.

**Type:** system sans stack (`ui-sans-serif, -apple-system, "Segoe UI", Roboto, Helvetica, Arial, sans-serif`). Tabular numerals for all figures.
**Shape:** cards `border-radius: 12px`, screens `18px`, pills `999px`. Soft shadows, generous whitespace, 1px hairline borders.

**Hard constraints:**
- **Light mode only.** ContentStudio has no dark mode — pin `color-scheme: light` and do not add a dark variant.
- **No RTL.**
- **Fully self-contained** — no CDN, no external fonts, no external images. Draw all icons and charts as **inline SVG** or CSS. Use lettered avatar circles instead of photos.
- Charts must be static, hand-built SVG (bars, scatter, lines) — do **not** pull in a chart library.
- Desktop-first, but nothing should overflow horizontally; wide tables scroll inside their own container.

## 4. The screens to design

### Screen 1 — Competitor reports list, X selected, empty
The Competitors landing page. A platform switcher with **Facebook · Instagram · X** (X selected, X logo in brand black). Below: a grid of report tiles. For X there are none yet, so show a single dashed empty tile.

Copy:
- Headline: **Create your first X competitor report**
- Subtext: **Compare your X profile against up to 4 rivals — followers, engagement, impressions and top posts, side by side.**
- CTA button: **Create X report**
- A small info note under the CTA: **X charges for every profile and post we read, so X reports refresh only when you sync them.**

### Screen 2 — Create report / Manage competitors modal
Modal titled **Create X competitor report**.

- Field: **Report name** — placeholder `e.g. Social tools benchmark`
- Field: **Add profiles by X handle** — placeholder `@handle`, helper text **Add up to 5 profiles including your own. Enter the exact handle — we'll verify it exists.**
- A **Verify** button next to the input.
- Added profiles list, each row = avatar, display name, `@handle`, follower count, remove ✕.
- Counter: **3 of 5 profiles added**
- Footer buttons: **Cancel** / **Create report**

Show these states via the segmented control:
1. **Verified** — green check row: `Buffer · @buffer · 198,540 followers`
2. **Not found** — inline error: **We couldn't find @bufffer on X. Check the spelling and try again.**
3. **Private account** — inline error: **@example is a protected account. X only shares data for public profiles, so it can't be tracked.**
4. **Suspended** — inline error: **@example is suspended on X and has no data to track.**
5. **Already added** — inline error: **@buffer is already in this report.**
6. **Limit reached** — input disabled, note: **You've added 5 of 5 profiles. Remove one to add another.**

### Screen 3 — Sync cost preview modal (the most important screen)
Triggered by a **Sync now** button. Modal titled **Sync X competitor data**.

Content:
- Explanatory line: **X charges for each profile and post we read. Choose how many recent posts to pull for each profile.**
- Dropdown: **Posts per profile** with options `10, 20, 30, 50, 80, 100` — **20 selected by default**.
- A summary block: **5 profiles × 20 posts = 100 posts**
- The cost line, large and prominent.
- Balance line underneath.
- Footer: **Cancel** / **Sync now**
- Small print: **You're only charged for data we successfully fetch. Nothing is charged if a sync fails.**

Two cohort states via the segmented control:
- **Wallet:** cost shows **$0.66**, balance line: **X Wallet balance: $24.80** with a **Top up** text link.
- **Credits:** cost shows **34 X credits**, balance line: **X credits remaining this month: 112 of 150** with a note **Shared with your X publishing credits.**

Add a third state:
- **Not enough balance** — cost `$0.66`, balance `$0.12`, the **Sync now** button disabled, and a warning band: **Your X Wallet doesn't have enough balance for this sync. Top up to continue.** with a **Top up wallet** button. (Credits variant: **You don't have enough X credits left this month. Buy more credits to continue.** → **Buy credits**.)

### Screen 4 — Syncing in progress
Same report shell with a progress state: **Syncing 5 profiles…** and a per-profile checklist ticking through (`@contentstudioio ✓`, `@buffer ✓`, `@hootsuite ⋯`, `@latersocial`, `@metricool`), plus **This usually takes under a minute.**

### Screen 5 — The X competitor report (the main dashboard)
The full report screen, populated. Header row: report name **Social tools benchmark**, an X logo chip, a **Last synced: 18 Aug 2026, 12:04** timestamp, a **Sync now** button, plus **Export** and **Share** secondary buttons.

Use this exact dataset across every widget so the numbers stay consistent:

| Profile | Handle | Followers | Change since last sync | Posts | Engagements | Impressions | ER by followers | ER by impressions |
|---|---|---|---|---|---|---|---|---|
| ContentStudio *(you)* | @contentstudioio | 28,412 | +312 | 20 | 1,842 | 412,300 | 0.32% | 0.45% |
| Buffer | @buffer | 198,540 | +1,204 | 20 | 9,860 | 2,140,000 | 0.25% | 0.46% |
| Hootsuite | @hootsuite | 7,912,000 | −2,140 | 18 | 4,320 | 1,980,000 | 0.00% | 0.22% |
| Later | @latersocial | 412,300 | +842 | 20 | 6,140 | 1,120,000 | 0.07% | 0.55% |
| Metricool | @metricool | 62,180 | +1,980 | 20 | 5,240 | 890,000 | 0.42% | 0.59% |

Mark the ContentStudio row as "you" with a subtle accent tint so it's findable at a glance.

Widgets to lay out, in this order:

1. **Overview cards** — four summary tiles: `Profiles tracked 5`, `Posts analysed 98`, `Total engagements 27,402`, `Total impressions 6.54M`.
2. **Comparative table** — every column from the dataset above, plus **Bookmarks** and **Reposts** columns. Heat-shade the best and worst cell per column (green tint / red tint). Include a footnote: **Engagement rate by impressions is only available on X — Facebook and Instagram don't share it for competitors.**
3. **Followers vs change** — grouped bar chart, total followers on one axis, change on the other. **Because this is the first sync, include an inline notice inside the card:** **Follower growth is tracked from your first sync onward. X doesn't share historical follower counts, so this chart fills in over time.**
4. **Engagement rate comparison** — horizontal bars, with a toggle above the chart: **By followers | By impressions**.
5. **Posting activity by post type** — stacked/grouped bars per profile across X post types: **Text, Image, Video, GIF, Poll, Link**.
6. **Performance comparison** — scatter plot, posts on X-axis vs engagement rate on Y-axis, one labelled bubble per profile, bubble size = followers.
7. **Post engagement breakdown** — per profile, a stacked bar splitting engagement into **Likes, Reposts, Replies, Quotes, Bookmarks**.
8. **Engagement over time** — small multiples: one mini line chart per profile across the synced period.
9. **Most engaged hashtags** — table: Hashtag | Used by | Times used | Total engagement. Use plausible rows (`#socialmedia`, `#contentmarketing`, `#smm`, `#AI`, `#creators`).
10. **Bio analysis** — table: Profile | Bio | Location | Website | Joined | Verified | Listed count. Use short realistic bios.
11. **Top performing posts** — per profile, a row of 2–3 post cards. Each card: avatar + handle, post text snippet, a media thumbnail placeholder for image/video posts, then a metric strip — **♥ likes · ↻ reposts · 💬 replies · 🔖 bookmarks · 👁 impressions**.
12. **Least performing posts** — same card layout, collapsed by default behind a section header.

### Screen 6 — Per-profile problem states inside a synced report
The comparative table again, but showing how failures appear **without breaking the report**:
- A row with a warning chip: **Protected — @example made their account private. No new data since 12 Aug 2026.**
- A row greyed out with: **Suspended on X**
- A row with: **No posts in this sync**

### Screen 7 — Export modal
Modal titled **Export report**. A checklist of report sections to include (the widget names from Screen 5, all checked by default), a **Language** dropdown (English selected; also list German, Spanish, French, Italian, Polish, Greek, Chinese), and buttons **Cancel** / **Download PDF**.

### Screen 8 — Shared read-only view
The same report viewed through a public share link: no **Sync now** button anywhere, a subtle banner reading **You're viewing a shared report. Data was last synced on 18 Aug 2026.** This exists to make the point that a shared link can never trigger a paid sync.

## 5. Copy rules — important

- Write for a **non-technical marketer**. No jargon, no API talk. Never use the words "API", "endpoint", "read", or "fetch" in the product UI itself.
- Say **repost**, never "retweet". Say **post**, never "tweet". The platform is **X**, not "Twitter".
- **Never show our cost or our margin inside the product UI.** The user sees only the price they pay (`$0.66`, `34 credits`) — never a per-post rate, never a percentage, never what X charges us.
- If you want to explain the economics to the reviewer, do it in the **annotation text outside the browser frame**, clearly separated from the UI.

## 6. Explicitly do NOT design these

They're impossible on X's public data and would mislead the reviewer:

- Any **audience demographics** (age, gender, location breakdown of followers)
- **Reaction-type breakdowns** like Facebook's love/haha/wow — X has no reactions beyond likes
- **Follower history charts that predate the first sync**
- **Profile visits, link clicks, or video completion rates** for competitors
- **Share of voice, mentions of a competitor, or sentiment** — out of scope for v1
- Any **auto-refresh / "updates daily"** messaging for X

## 7. Closing section

End the deck with a short **"Open questions for the PO"** panel listing, as a plain checklist:

1. Should the default be 20 posts per profile, or lower to keep the first sync cheap?
2. Should reposts and replies count as "posts", or only original posts?
3. Should X competitor reports also offer a scheduled auto-sync in v1, or on-demand only?
4. Should workspaces on the older X credits model get this at launch, given one sync can use a third of a monthly allowance?

Make it look considered and shippable — this is going in front of a decision-maker.
