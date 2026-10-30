# Helpin links: Public API: safe retries, account events, large uploads and usage history

Pushed 2026-10-26. Team: Engineering. Sprint: Nov 2 - Nov 15 - 2026. State: Ready. Priority: Medium. Unassigned.

**Epic:** [Public API: safe retries, account events, large uploads and usage history](https://app.helpin.ai/w/contentstudio/pm/epics/9b68e139-1d14-4a4c-9360-4b73592423e8)

| Key | Story | Label |
|---|---|---|
| [CONT-4627](https://app.helpin.ai/w/contentstudio/pm/epics/9b68e139-1d14-4a4c-9360-4b73592423e8?task=CONT-4627) | [BE] Prevent duplicate posts when a public API create request is retried | backend |
| [CONT-4628](https://app.helpin.ai/w/contentstudio/pm/epics/9b68e139-1d14-4a4c-9360-4b73592423e8?task=CONT-4628) | [BE] Prevent duplicate AI generations and credit charges when a public API request is retried | backend |
| [CONT-4629](https://app.helpin.ai/w/contentstudio/pm/epics/9b68e139-1d14-4a4c-9360-4b73592423e8?task=CONT-4629) | [BE] Send webhooks when a social account is connected, disconnected or needs reconnecting | backend |
| [CONT-4630](https://app.helpin.ai/w/contentstudio/pm/epics/9b68e139-1d14-4a4c-9360-4b73592423e8?task=CONT-4630) | [FE] Describe the account events in the webhook event picker | frontend |
| [CONT-4631](https://app.helpin.ai/w/contentstudio/pm/epics/9b68e139-1d14-4a4c-9360-4b73592423e8?task=CONT-4631) | [BE] Let the public API upload large media files directly to storage | backend |
| [CONT-4632](https://app.helpin.ai/w/contentstudio/pm/epics/9b68e139-1d14-4a4c-9360-4b73592423e8?task=CONT-4632) | [BE] Add daily API usage history to the public API | backend |
| [CONT-4633](https://app.helpin.ai/w/contentstudio/pm/epics/9b68e139-1d14-4a4c-9360-4b73592423e8?task=CONT-4633) | [BE] Bring safe retries, account events, large uploads and usage history to the CLI, MCP server and automation apps | backend |

**Docs (linked to the epic):**

| Doc | Collection | ID |
|---|---|---|
| Research | Research | e7dbbf95-4584-4195-bc77-19b13681eb57 |
| Workflow | PRDs & Feature Specs | eda5823b-0d81-4991-9c61-b705f40d6e9c |
| PRD | PRDs & Feature Specs | 47a40c14-7e02-4a47-b28b-4525302d57a8 |

Push notes: CONT-4633's body came back as raw markdown and was re-sent as HTML with update_task.
2026-10-26: CONT-4629 now explicitly opt-in, at the PO's request: delivered only to webhooks that ticked the event, and existing webhooks are not auto-subscribed.
2026-10-26: event renamed account.reauth_required → account.reconnect_required (PO). CONT-4629, 4630 and 4633 updated. CONT-4630 also corrected: the description is the checkbox hover tooltip, not text under the name. The Helpin Research, Workflow and PRD docs still say reauth_required.
