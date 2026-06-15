# Research — Polish the Sample Workspace demo content

> Quick-story pipeline · Step 1 · Keep it lean

## What the user asked for

Make the **Sample Workspace** feel like a real, lived-in account so people exploring it (new sign-ups, trial users, sales demos) see ContentStudio working with believable content instead of empty widgets and flat charts. Four areas:

1. **Planner → post detail modal — improve the comments UI** so the conversation on a post looks real and reads cleanly.
2. **Analytics graphs — make them more realistic** (believable trends, not flat/empty/obviously-fake lines).
3. **Notifications — populate the module** so the notification panel has realistic activity instead of being empty.
4. **Home page — populate the "To-Do List" and "Content Publishing"** widgets so they show real numbers and tasks instead of zeros/empty states.

> Note: "Sample Workspace" is the demo workspace a user lands in to explore the product. There's no literal `sample workspace` string in the frontend — these are the **live product surfaces** that look empty when the workspace has no real activity. The work is about filling those surfaces with realistic demo content (and, for the planner comments, tidying the UI while we're there).

## Current State (where each surface lives)

**1. Planner post detail — comments**
- The post detail view renders comments through [CommentsAndNotes.vue](contentstudio-frontend/src/modules/planner_v2/components/CommentsAndNotes.vue) and [CommentCard.vue](contentstudio-frontend/src/modules/planner_v2/components/CommentCard.vue).
- A comment card shows: avatar, name, relative time ("2 hours ago"), comment type tooltip, resolved state (dimmed), and edit mode. The structure is there; in the Sample Workspace it's empty or sparse, so the conversation doesn't look real.

**2. Analytics graphs**
- Analytics lives in [analytics_v3](contentstudio-frontend/src/modules/analytics_v3/views/) (`MainAnalytics.vue`, plus Facebook/Instagram views). With little/no data, charts render flat or empty.
- There's already a "dummy data" pattern for the competitor section — [CompetitorDummyGraphs.vue](contentstudio-frontend/src/modules/analytics_v3/views/common/CompetitorDummyGraphs.vue) — proving the product already shows placeholder graphs when real data is missing. The Sample Workspace should show **believable** numbers, not obviously-fake ones.

**3. Notifications**
- The header notification panel is [HeadNotificationSlider.vue](contentstudio-frontend/src/modules/common/components/header-notifications/HeadNotificationSlider.vue) (e.g. account reconnect, post-failure alerts). In the Sample Workspace this is empty, so the bell looks dead.

**4. Home page widgets**
- Home is [Home.vue](contentstudio-frontend/src/Home.vue) / [DashboardNew.vue](contentstudio-frontend/src/views/DashboardNew.vue); widgets live in [src/components/dashboard/](contentstudio-frontend/src/components/dashboard/).
- **To-Do List** → [ToDoCard.vue](contentstudio-frontend/src/components/dashboard/ToDoCard.vue): shows actionable items like "Need Approval" that deep-link into the planner. Empty in the Sample Workspace.
- **Content Publishing** → [ContentPublishingCard.vue](contentstudio-frontend/src/components/dashboard/ContentPublishingCard.vue) + [ContentPublishingDonut.vue](contentstudio-frontend/src/modules/dashboard/components/charts/ContentPublishingDonut.vue): a donut split into **Published / Scheduled / Partially failed / Failed** posts. Shows zeros in the Sample Workspace.

## What Needs to Change

- Fill the Sample Workspace with realistic demo content so these four surfaces look populated and believable:
  - **Comments:** a short, natural-looking comment thread on a sample post (a couple of teammates, a reply, one resolved) — plus a light polish of the comment card layout so it reads cleanly.
  - **Analytics:** believable trend lines and totals (followers, reach, engagement over time) instead of flat/empty charts.
  - **Notifications:** a few realistic sample notifications so the bell/panel has content.
  - **Home — To-Do List:** show real-looking tasks (e.g. posts awaiting approval) that match the sample posts.
  - **Home — Content Publishing:** show a believable donut (a mix of published / scheduled / partially failed / failed) that matches the sample posts.
- **Consistency matters:** the demo numbers should agree across surfaces — the posts shown in the planner should be the same ones counted in the Content Publishing donut and surfaced in the To-Do List, so nothing looks contradictory.

## Mobile Context

- Home, analytics, planner and notifications also exist in the iOS and Android apps. If the Sample Workspace is shown there too, the same "empty/unrealistic" problem applies. **Open question for the user:** is the Sample Workspace experience in scope for mobile, or web-only for now? (Default assumption: web-only for this round.)

## Open questions for the review gate

1. **Where does the sample content come from?** Is the Sample Workspace seeded with real demo records (backend/data) or shown via mocked/placeholder content (frontend)? This decides whether we need a backend data-seeding story alongside the frontend work. (Best guess: the surfaces are real and data-driven, so making them look populated most likely needs seeded demo data → a backend/data effort + the frontend comments polish.)
2. **Story split** — one combined "polish the Sample Workspace" story, or split (data-population vs comments-UI)?
3. **Epic** — you said "assign in the Sample Workspace feature." Is there an existing **Sample Workspace** epic in Shortcut to point these at (instead of the default Miscellaneous epic)?
4. **Mobile** — web-only, or include iOS/Android?

## Files Involved (for grounding only — kept out of the story body)

- `contentstudio-frontend/src/modules/planner_v2/components/CommentsAndNotes.vue`, `CommentCard.vue`
- `contentstudio-frontend/src/modules/analytics_v3/views/` (`MainAnalytics.vue`, `common/CompetitorDummyGraphs.vue`)
- `contentstudio-frontend/src/modules/common/components/header-notifications/HeadNotificationSlider.vue`
- `contentstudio-frontend/src/components/dashboard/ToDoCard.vue`, `ContentPublishingCard.vue`
- `contentstudio-frontend/src/modules/dashboard/components/charts/ContentPublishingDonut.vue`
- `contentstudio-frontend/src/Home.vue`, `contentstudio-frontend/src/views/DashboardNew.vue`
