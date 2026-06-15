# Stories — Polish the Sample Workspace demo content

> Quick-story pipeline · Step 2 · Deliverable for the Product Owner to create in Shortcut manually. Nothing is pushed to Shortcut.
>
> **Scope decisions (confirmed):** one combined story · assign to the existing **Sample Workspace** epic · **web-only** (no mobile stories this round).

---

## [Full Stack] Polish the Sample Workspace so it looks lived-in (Home, Analytics, Notifications & Planner)

### Description

As a **new user, trial user, or someone being shown a demo**, I want the **Sample Workspace** to look like a real, active account — with a populated home page, believable analytics, real-looking notifications, and a planner post that has an actual conversation on it — so that I immediately understand what ContentStudio does and what my own workspace could look like, instead of seeing empty widgets, flat charts, and a dead notification bell.

Right now the Sample Workspace surfaces are technically working but visually empty: the home page shows zeros, the analytics charts are flat, the notification panel is empty, and posts in the planner have no comment activity. This makes the demo feel hollow and undersells the product on a user's very first impression.

This story fills those four surfaces with realistic, **consistent** demo content so the Sample Workspace tells a believable story end to end.

---

### Workflow

1. A user lands in the **Sample Workspace** (e.g. right after signing up, or when exploring the product before connecting their own accounts).
2. On the **Home page**, they see populated widgets instead of zeros:
   - The **To-Do List** shows real-looking tasks they can act on (e.g. posts awaiting approval, a post that needs attention).
   - The **Content Publishing** donut shows a believable mix of Published / Scheduled / Partially failed / Failed posts.
3. They open **Analytics** and see believable charts — follower growth, reach, and engagement trending over time with natural ups and downs — instead of flat or empty graphs.
4. They open the **notification bell** and see a few realistic recent notifications (a post published, posts scheduled, an approval, a report ready) instead of an empty panel.
5. They open a **sample post in the Planner** and, in the post detail view, see a short, natural comment thread between sample teammates — including a reply and one resolved comment — laid out cleanly and easy to read.
6. Everything **agrees with everything else**: the posts shown in the planner are the same ones counted in the Content Publishing donut and surfaced in the To-Do List, and the analytics reflect a workspace that's clearly been active. Nothing looks contradictory or obviously fake.

---

### Acceptance criteria

**Home page — To-Do List**
- [ ] In the Sample Workspace, the To-Do List shows real-looking actionable items instead of an empty state (e.g. "2 posts need your approval", "1 post needs attention", "3 drafts ready to schedule").
- [ ] Each To-Do item links to the matching place in the Sample Workspace (e.g. clicking "posts need approval" opens the planner filtered to those sample posts).
- [ ] The counts in the To-Do List match the actual sample posts present in the workspace (no item points to content that doesn't exist).

**Home page — Content Publishing**
- [ ] The Content Publishing donut shows a believable spread across Published, Scheduled, Partially failed, and Failed (e.g. Published 24 · Scheduled 8 · Partially failed 1 · Failed 1) instead of all zeros.
- [ ] The donut totals match the sample posts in the workspace and the To-Do List counts — the numbers are consistent across the home page.

**Analytics**
- [ ] Analytics charts show believable demo data with natural variation over time (follower growth, reach, engagement) instead of flat lines or empty states.
- [ ] Headline numbers (e.g. total followers, total reach/impressions, engagement rate) are realistic for a small-but-active account, not round placeholder values like "0" or "100/100/100".
- [ ] Any "top posts" / best-performing content shown reflects the same sample posts that appear in the planner.

**Notifications**
- [ ] The notification panel shows a few realistic recent notifications instead of an empty state.
- [ ] Notifications reference the same sample content elsewhere in the workspace (e.g. a "post published" notification names a post that exists in the planner).
- [ ] Notifications have believable, varied timestamps (e.g. "1h ago", "yesterday", "2 days ago") so they read as recent activity.

**Planner — post detail comments**
- [ ] A sample post in the planner has a short, natural comment thread (at least 3 comments) between sample teammates, including at least one reply and one resolved comment.
- [ ] The comment thread is laid out cleanly and is easy to read: each comment clearly shows who wrote it, when, and the resolved comments are visually distinct (dimmed/marked resolved) from open ones.
- [ ] The comments read like a real conversation (natural language, on-topic to the post), not lorem-ipsum or filler text.

**Consistency & safety**
- [ ] Demo content is consistent across all four surfaces — the same sample posts/accounts are referenced everywhere and no surface contradicts another.
- [ ] Demo content appears **only** in the Sample Workspace and never leaks into a user's real workspaces, real analytics, or real notifications.
- [ ] Demo content reads naturally on white-label domains — it does not hardcode "ContentStudio"-specific branding that would look wrong for a white-label customer's demo.

---

### Suggested demo content (copy)

This is concrete sample content the team can use as-is or adapt. Keep it generic and brand-neutral so it works on white-label domains.

**Planner — comment thread on a sample post** (e.g. a post titled *"Summer sale is live ☀️"*):
- **Sarah (Content Manager)** · 2 hours ago — "Love this one! Can we swap the hero image for the brighter version? 🙌"
- **Ali (Designer)** · 1 hour ago (reply) — "Updated — the brighter image is in now. Take a look 👀"
- **Sarah (Content Manager)** · 45 minutes ago — "Perfect, approving it now." *(marked resolved)*

**Notifications** (most recent first):
- "Your post *'Summer sale is live ☀️'* was published to Instagram and Facebook." — 1h ago
- "3 posts are scheduled to go out tomorrow." — 3h ago
- "Maria approved 2 posts in the *Q3 Campaign*." — yesterday
- "Your weekly analytics report is ready to view." — 2 days ago

**Home — To-Do List:**
- "2 posts need your approval"
- "1 post needs attention" (matches the 1 failed post in the donut)
- "3 drafts ready to schedule"

**Home — Content Publishing donut:** Published 24 · Scheduled 8 · Partially failed 1 · Failed 1

**Analytics (believable targets):** followers trending up from ~1,200 to ~1,580 over the last 30 days; reach and engagement lines with natural day-to-day variation; an engagement rate in a realistic range (e.g. ~3–5%).

> Exact values are illustrative — the point is "small but active and believable," and that the numbers stay consistent with each other across surfaces.

---

### Mock-ups

N/A — no mock-ups provided. The "Suggested demo content" above defines the content; existing layouts are reused.

---

### Impact on existing data

- The Sample Workspace gains demo content (posts, comments, analytics, notifications). This content must be **scoped to the Sample Workspace only** and must never appear in, or be counted toward, a user's real workspaces, real analytics, or real notification history.
- No change to the data of existing real workspaces.

---

### Impact on other products

- **Mobile apps (iOS/Android):** Out of scope for this round (web-only, confirmed). The same home/analytics/planner/notification surfaces exist on mobile — if the Sample Workspace is later shown there, a follow-up is needed so it looks populated too.
- **Chrome extension:** N/A — the Sample Workspace exploration experience is in the web app.
- **White-label domains:** In scope. The Sample Workspace also appears on white-label domains, so the demo content must stay brand-neutral and theme-aware (no hardcoded ContentStudio branding, theme-aware colours).

---

### Dependencies

- None blocking. The four surfaces should be populated together so the demo stays consistent — partial population (e.g. populated home but empty analytics) would look worse than today.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only) — web layouts reused; verify the populated widgets/charts/comment thread render correctly across breakpoints
- [ ] Multilingual support (frontend + backend, translations available or fallback handled) — demo content strings should go through the normal translation path or have a sensible fallback
- [ ] UI theming support (default + white-label, design library components are being used) — demo content and charts must render correctly on white-label themes
- [ ] White-label domains impact review — demo content must be brand-neutral and theme-aware (see AC)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension) — web-only this round; mobile noted as a possible follow-up

---

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Surfaces to populate (entry points):**
- Home widgets — `contentstudio-frontend/src/components/dashboard/ToDoCard.vue` (To-Do List) and `ContentPublishingCard.vue` + `contentstudio-frontend/src/modules/dashboard/components/charts/ContentPublishingDonut.vue` (donut: Published / Scheduled / Partially failed / Failed)
- Analytics — `contentstudio-frontend/src/modules/analytics_v3/views/` (`MainAnalytics.vue` and the Facebook/Instagram views)
- Notifications — `contentstudio-frontend/src/modules/common/components/header-notifications/HeadNotificationSlider.vue`
- Planner comments — `contentstudio-frontend/src/modules/planner_v2/components/CommentsAndNotes.vue` and `CommentCard.vue` (card already supports avatar, author, relative time, resolved/dimmed state — comments polish is light)

**Existing pattern to follow:**
- There's already a "show believable placeholder graphs when real data is missing" pattern in `contentstudio-frontend/src/modules/analytics_v3/views/common/CompetitorDummyGraphs.vue` — the analytics demo content can follow the same idea so it looks like real data, not an obvious placeholder.

**Key design decision for engineering (kept out of the user-facing body):**
- The cleanest path is likely **seeding the Sample Workspace with real demo records** (posts, comments, analytics, notifications) so the existing data-driven widgets simply render them and stay consistent automatically — rather than mocking each widget independently, which risks the surfaces drifting out of sync. Whether this is done via backend seed data or a frontend demo-content layer is an engineering call; either way, the **consistency** and **no-leak-into-real-workspaces** ACs are the hard requirements.

**Gotcha:**
- Keep demo content out of real analytics aggregation and real notification streams — the Sample Workspace's posts/notifications must not be counted toward, or surface in, any real user's data.

---

### Shortcut fields

- **Template:** New Feature Template (PO selects this when creating the story so the standard sections + 5 quality-checklist tasks are pre-populated)
- **Story type:** Feature
- **Project:** Web App
- **Group:** Full Stack
- **Epic:** Sample Workspace *(existing epic — confirm/select it when creating the story; falls back to **Q2 - 2026: Miscellaneous** (id: 115078) only if that epic doesn't exist)*
- **Priority:** Medium
- **Product Area:** Onboarding *(the Sample Workspace is the first-run / exploration experience; this change also touches Dashboard, Analytics, and Planner)*
- **Skill Set:** Frontend, Backend *(plus light Design input on the comments layout, if desired)*
- **Estimate:** — (left empty; devs estimate during sprint planning)
- **Labels:** none (team manages labels manually)
- **Iteration:** PO assigns the current/target sprint at creation time

---

> **Analytics events (Usermaven):** N/A — this story populates demo content and polishes a layout; it introduces no new trackable user action. Viewing the Sample Workspace is read-only navigation, already covered by global pageview tracking.
