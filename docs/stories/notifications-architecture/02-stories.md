# Epic + stories — Notifications architecture

**Scope of this doc:** the epic and **one** story, a research ticket. No implementation stories yet. Nothing is pushed to any tracker. The PO creates this by hand.

---

## Epic: Notifications architecture

ContentStudio sends notifications from several places that grew independently: the web app's in-app notification centre, email, browser and realtime updates, the social inbox, analytics reports, and mobile push. Each of those was built for its own feature, with its own naming, its own delivery path and its own idea of what a user is allowed to switch off. There is no single list of what notifications exist, who gets them, or on which channel. Adding a notification today means picking a pipeline and hoping it matches the others.

That already hurts us in two visible ways. First, users cannot reason about their own settings, because the notification preferences they see only cover part of what we actually send. Second, the mobile app receives almost nothing: the only push notification it handles is the publish reminder for platforms that need manual publishing. Everything else, including posts waiting for the user's approval, is web and email only.

This epic defines one notification architecture and one workflow for the whole product, and then starts using it. The first delivery target is **in-review posts on mobile**: when a post is waiting for a user's approval, that user should get a push notification and land on the post in the app. The rest of the mobile notification set comes later, once the architecture is agreed.

**This epic starts with research only.** The first ticket produces the picture and the target design. Implementation stories are written after that is reviewed.

### Out of scope for now

- Any implementation work. The architecture has to be agreed first.
- The full mobile notification set. Only in-review posts are in the first delivery target.
- Replacing the realtime provider. That work already exists separately and this epic sequences behind it.

### Stories

1. `[Research]` Define one notification architecture and workflow for ContentStudio, and the plan for in-review post notifications on mobile

---

## [Research] Define one notification architecture and workflow for ContentStudio, and the plan for in-review post notifications on mobile

### Description

As a product and engineering team, we want a single, agreed picture of every notification ContentStudio sends and one target architecture for sending them, so that we stop growing a new notification pipeline for every feature and can add notifications, including mobile push for in-review posts, in one predictable way.

### Workflow

1. The team collects every notification ContentStudio sends today and writes them down in one place: what triggers it, who receives it, and on which channels the recipient gets it (in-app, email, mobile push, realtime update in the browser).
2. The team records, for each notification, whether the user can currently turn it off, and where that setting lives.
3. The team identifies where the same notification behaves differently depending on which part of the product produced it, and where a notification a user would expect simply does not exist on a channel.
4. The team agrees one target model: a single notification catalogue, consistent naming, one set of channels, one place a user manages their preferences, and one delivery path new features plug into.
5. The team agrees the notification workflow end to end, from "something happened" to "the user saw it", including what happens when delivery fails and what the user should be able to see about their own notification history.
6. The team defines what has to be true for a user to get a push notification on the mobile app when a post is waiting for their approval, and what the user sees and where they land when they tap it.
7. The team writes down the sequence: what has to be built first, what can be delivered as the first slice, and what is deliberately deferred to a later phase.
8. The output is reviewed and approved. Implementation stories are written from it afterwards.

### Acceptance criteria

- [ ] A single catalogue lists every notification ContentStudio sends today, with its trigger, its recipients, and the channels it currently reaches: in-app, email, mobile push, realtime browser update.
- [ ] The catalogue covers all product areas that notify users today, including publishing and the planner, post approvals and approval workflows, comments and tasks, the social inbox, analytics reports, workspaces and team membership, billing, and white-label teams.
- [ ] For each notification, the document states whether a user can currently turn it off and where that control lives, and flags the ones with no control at all.
- [ ] Inconsistencies are named explicitly: notifications that reach one channel but not another where users would expect them, notifications named differently in different parts of the product, and duplicate notifications for the same event.
- [ ] The document states which notifications the mobile app receives today and which it does not.
- [ ] A proposed target architecture is documented: one notification catalogue, one naming convention, one agreed set of channels, one user-facing preferences model, and one path a new notification is added through.
- [ ] The target model covers per-user and per-workspace preferences, and says what happens for a user who belongs to several workspaces.
- [ ] The target model states what happens when a notification fails to deliver, and what visibility the team has into that.
- [ ] The end-to-end notification workflow is documented as a diagram plus a written walkthrough, from trigger to delivered notification, covering all channels.
- [ ] The document defines the first delivery target in full: a user with a post waiting for their approval receives a mobile push notification, and tapping it opens that post in the app.
- [ ] The in-review notification design states who is notified, what the notification says, when it is sent and when it is deliberately not sent, for example when the same user made the change themselves.
- [ ] The document states which later mobile notifications are in the intended set and confirms they are deferred, so this epic's first slice is only in-review posts.
- [ ] A sequencing plan lists what must be built first, what ships as the first slice, and what is deferred, with the dependencies on already-planned work called out by name so this epic does not conflict with it.
- [ ] Open questions and recommendations are captured for the review gate.
- [ ] The document is reviewed and approved before any implementation story is written.

### Mock-ups

None. This is a research ticket with no UI.

### Impact on existing data

None. Nothing is changed, migrated or deleted by this ticket. The document does record where notification data lives today and flags any duplication for the implementation phase to resolve.

### Impact on other products

This research covers the web app, the mobile app, the social inbox, and analytics reports together, because all of them notify users today. No product changes yet. The output determines the implementation stories for each of them.

### Dependencies

- Sequences alongside the existing realtime provider migration work, so the target architecture does not contradict it.
- Sequences alongside the existing analytics and inbox webhook event work, so notification events and webhook events are not defined twice.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, research ticket with no UI
- [ ] Multilingual support (the target model must state how notification copy is translated across channels)
- [ ] UI theming support — N/A, research ticket with no UI
- [ ] White-label domains impact review (notification copy, sender identity and links differ per white-label domain)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
