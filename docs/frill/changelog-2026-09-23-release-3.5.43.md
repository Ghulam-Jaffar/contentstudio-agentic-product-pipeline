# Changelog draft — Release 3.5.43

Source: Helpin **CONT-4010** (Release 3.5.43, shipped 22-09-2026, Production).
Structure follows `docs/frill/changelog-template.md`.

**Status: DRAFT, not pushed.** On approval it goes to Frill with `published_at: null`.

## Proposed announcement metadata

| Field | Value |
|---|---|
| `name` | YouTube competitor analytics, AI Chat Skills, major API expansion |
| `category_idxs` | New Feature, Improvement, Fix, Integrations |
| `idea_idxs` | `idea_g6n0n08e` (composer clickable links) |
| `published_at` | `null` — draft |

---

## Body

### YouTube Competitor Analytics — Now Available

You already track how you stack up on Facebook and Instagram. Now YouTube joins them. See how your channel performs against the competition, spot what they are doing well, and bring that back into your own content plan without leaving ContentStudio.

**Here's what's included:**

- **Channel-level comparison.** Track your channel against competitors on subscribers, views, and publishing cadence, so you can see who is actually growing.
- **Content performance breakdown.** See which of their videos land, how often they post, and what formats they lean on.
- **Sits alongside your existing reports.** YouTube competitor reports work the same way as the Facebook and Instagram ones you already use.

**For more details, check:** [YouTube Competitor Analytics](https://docs.contentstudio.io/) <!-- TODO: PO to confirm the exact help-doc URL -->

---

### Skills in AI Chat — Now Available

Stop re-explaining yourself. If you have a task you run over and over, save the instructions once as a Skill and reuse it. Same input, same quality, every time, without rewriting the brief.

**Here's how it works:**

- **Save a set of instructions once.** Turn a prompt you keep retyping into a reusable Skill.
- **Reuse it any time.** Call a Skill straight from AI Chat and get consistent output without re-explaining the task.
- **Keep your team consistent.** Everyone runs the same Skill, so the output does not drift from person to person.

---

### API Updates — New Endpoints

This is the largest API release yet. Publishing, approvals, automations, media and workspace administration are now all reachable programmatically, so you can build ContentStudio into your own tooling rather than working around it.

**Here's what's new:**

- **Post Share Links:** Create, retrieve, update and delete share links, send invitations, and read back the comments, approvals and rejections submitted through each one.
- **Recurring Posts:** Set repeat scheduling on post create and update. Each occurrence is an independent post that still knows which original it came from.
- **Content Categories and Slots:** Create, retrieve, update and delete categories, manage their weekly posting slots, redistribute queued posts across available times, and read the next expected publishing time for anything added to the queue.
- **Approval Workflows:** Create, retrieve, update, delete and duplicate workflows, including their levels, approvers and rules.
- **Automations:** Manage RSS, evergreen and bulk CSV automations, including feed validation, RSS pull history and on-demand pulls. Bulk CSV files are submitted in one call, processed in the background, and their rows can be reviewed, approved, labelled or assigned to a campaign before scheduling.
- **Bulk Automation Webhook:** An `automation.bulk.completed` event fires when a bulk CSV batch finishes, so your integration can react instead of polling.
- **Pause and Resume Posting:** Pause all publishing for a workspace and resume later, choosing whether to restore original slots, re-spread posts from now, or leave them as drafts. The response reports how many posts were held, restored or left as drafts.
- **Media Library Management:** Delete, archive and move media, and create, rename and delete folders. Deletes that would affect scheduled posts or non-empty folders ask for explicit confirmation first.
- **Media Storage, Notes and Brand Assets:** Read plan storage usage and remaining capacity, attach per-asset notes, and flag brand assets. Uploads that would exceed your storage limit now fail with the limit and your current usage in the response.
- **Workspace Settings:** Read and write every workspace-level setting, including timezone, first day of the week, Instagram posting method and report branding. Timezone changes tell you how many scheduled posts are affected, and an invalid timezone is rejected rather than silently falling back to UTC.
- **Team Member Social Account Access:** Read and set which social accounts a team member can use, without touching their role. Posting, disconnect and reconnect now enforce those grants, matching the web app.
- **Plan Usage and Limits:** Read workspace plan allowances, current usage, remaining capacity, API rate limits and reset times.

**For more details, check:** [ContentStudio API](https://docs.contentstudio.io/articles/contentstudio-api-f955a6ab)

---

### Other Improvements and Fixes

- Links in the post composer are now clickable and open in a new tab, so you can check a link resolves before you schedule the post.
- The plan upgrade experience has been rebuilt, and the new upgrade modal is now shown to customers on older billing plans.
- AI Chat handles reconnects inline, so a dropped connection no longer interrupts what you were doing.
- Campaigns and labels now work correctly in AI Chat.
- AI Chat handles bulk analytics requests properly instead of stalling on large pulls.
- Fixed analytics overview and platform API responses returning incorrect data.

---

## Excluded from this changelog, with reasons

| Release item | Why it is not announced |
|---|---|
| **LinkedIn Profile Analytics** | Feature flagged. Release ticket says so, and `linkedin_profile_analytics` gates it per user in the backend. Only a few requesting users have it |
| **Threads Analytics** | **Feature flagged — the release ticket did not say so.** `threads_analytics` gates the routes, the schedule-report modal and the AI-insights surfaces in the frontend, 16 references. A code comment reads *"gated behind the `threads_analytics` feature flag while it rolls [out]"*. Announcing it would send most customers to a feature they cannot see |
| **Pusher to Centrifugo migration** | Internal technical debt, explicitly tagged as such. No customer-visible change |
| Database / schema / migrations / env vars / config / deployment | Internal release-ticket sections with no customer-facing outcome |

## Verification notes

- **YouTube competitor analytics is NOT gated.** `competitor.ts` registers `youtube_competitor_v3` and `youtube_competitor_overview_v3` unconditionally; the only feature-flag check in that file is for Threads. Safe to announce.
- **Complete frontend feature-flag list** is just four: `facebook_reel_collaborators`, `google_ads`, `threads_analytics`, `whatsapp`, plus LinkedIn via the `LINKEDIN_PROFILE_ANALYTICS_FLAG` constant. Nothing else in this release is gated.
- **Composer clickable links shipped** — `88bd1d4403 feat(CONT-3432)`, confirmed an ancestor of `origin/develop`, adding `EditorLinkAffordance.vue` and `linkAffordance.ts`. It landed in the shared `CstSocialEditor`, so it applies anywhere that editor is used, not only the composer. Originated from Frill idea `idea_g6n0n08e`.
- **Skills in AI Chat shipped** — multiple commits through PR #7010, and `skillMention.ts` is present in `modules/AI-tools/`. No feature flag.
- **Public API items verified** against backend routes: share-links, content categories, approval workflows, automations, plan, media all have route definitions, and `WebhookEventType.php` defines `automation.bulk.completed`.
- **Caveat on the local checkouts.** Both repos sit on a long-lived `features` branch, 636 commits ahead of and 15 behind `develop`. Per-item checks were therefore made against `origin/develop` rather than the working tree. An earlier pass against the working tree wrongly suggested the composer feature was missing.

## Open items for the PO

1. **Screenshots.** The house style puts an image under each headline feature. None are attached here; they need adding in the Frill editor.
2. **YouTube competitor help-doc URL.** The link above is a placeholder — I did not want to invent an article URL.
3. **Confirm the Threads exclusion.** This is the one judgement call that changes the announcement materially. If the flag is in fact on for everyone, Threads Analytics should be the lead item.
