### YouTube Competitor Analytics: Now Available

You already track how you stack up on Facebook and Instagram. Now YouTube joins them. See how your channel performs against the competition, spot what they are doing well, and bring that back into your own content plan without leaving ContentStudio.

**Here's what's included:**

- **Channel-level comparison.** Track your channel against competitors on subscribers, views, and publishing cadence, so you can see who is actually growing.
- **Content performance breakdown.** See which of their videos land, how often they post, and what formats they lean on.
- **Sits alongside your existing reports.** YouTube competitor reports work the same way as the Facebook and Instagram ones you already use.

---

### Skills in AI Chat: Now Available

Stop re-explaining yourself. If you have a task you run over and over, save the instructions once as a Skill and reuse it. Same input, same quality, every time, without rewriting the brief.

**Here's how it works:**

- **Save a set of instructions once.** Turn a prompt you keep retyping into a reusable Skill.
- **Reuse it any time.** Call a Skill straight from AI Chat and get consistent output without re-explaining the task.
- **Keep your team consistent.** Everyone runs the same Skill, so the output does not drift from person to person.

---

### API Updates: New Endpoints

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

---

### Other Improvements and Fixes

- Links in the post composer are now clickable and open in a new tab, so you can check a link resolves before you schedule the post.
- The plan upgrade experience has been rebuilt, and the new upgrade modal is now shown to customers on older billing plans.
- AI Chat handles reconnects inline, so a dropped connection no longer interrupts what you were doing.
- Campaigns and labels now work correctly in AI Chat.
- AI Chat handles bulk analytics requests properly instead of stalling on large pulls.
- Fixed analytics overview and platform API responses returning incorrect data.
