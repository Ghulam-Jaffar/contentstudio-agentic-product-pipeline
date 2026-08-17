# Research — Reopen an inbox conversation when a new inbound message arrives

## Current State

A conversation marked as Done never returns to the active tabs. Once `is_marked_done` is `true`, the only thing that clears it is a human action (Mark as pending / Mark as undone). New inbound activity does not clear it, so the conversation sits in **Marked as Done** and the new customer message is silently missed.

**Where the flag lives.** Every inbox element document carries `inbox_details.marked_done.is_marked_done` (conversations and reviews) or `comments_details.marked_done.is_marked_done` (post comments).

**Where tab membership is decided** — `social-inbox-manager/app/database/mongo/repository/inbox_details_repository.py:1499-1516`:

```
UNASSIGNED     -> is_assigned=False, is_marked_done=False, is_archived=False
MINE           -> is_assigned=True, assigned_to.user_id=me, is_marked_done=False, is_archived=False
ASSIGNED       -> is_assigned=True, is_marked_done=False, is_archived=False
MARKED_AS_DONE -> is_marked_done=True
ARCHIVED       -> is_archived=True, is_marked_done=False
```

Tab membership is derived purely from these flags, server-side. The same query builder feeds the sidebar counts (`inbox_details_repository.py:1738-1780`). So flipping `is_marked_done` back to `false` is all that is needed to move a conversation out of Marked as Done and back into whichever active tab it belongs to.

**The reopen code already exists and is commented out.** Two places, both with a live but useless `find_by_element_id(...)` call whose result is discarded:

- `social-inbox-manager/app/database/mongo/repository/inbox_messages_repository.py:67-94` — comment says _"Reset marked_done status when a new message is received"_, followed by a commented-out `update_many` that sets `inbox_details.marked_done.is_marked_done` to `False`. This is inside `add_update_message`, on the `"added"` (new message) branch only.
- `social-inbox-manager/app/database/mongo/repository/inbox_comments_repository.py:113-136` — same shape for post comments, targeting the parent post's `inbox_details.marked_done`.

Both were disabled, most likely because at that layer the repository has no `account_id` and therefore cannot tell an inbound customer message from the workspace's own outbound reply. Reopening on every message would fight the existing "reply and mark done" flow (`social-inbox-manager/app/api/v1/routes/sends.py:383-396` sends `is_marked_done` / `is_pending` with the reply), because the platform echo of that outbound message would immediately undo the Done the agent just set.

**The sync path preserves Done deliberately.** `_update_existing_element` in each platform strategy (`facebook_strategy.py:168-205`, and the same method in `instagram_strategy.py:175`, `linkedin_strategy.py:165`, `gmb_strategy.py:156`, `youtube_strategy.py:178`) copies the existing `inbox_details` forward untouched and only refreshes `element_details` + `updated_at`. `inbox_comments_repository.py:82-89` does the equivalent by deleting `comments_details` from the update payload. That is correct for re-syncs, but it means the webhook path (which re-fetches the whole conversation on every new message) can never reopen anything either.

**The inbound check already exists one layer up.** `add_update_messages` in the strategy knows `self.account_id` and already uses exactly the test we need:

- `facebook_strategy.py:884-901` — `status == "added" and payload["from"][0]["id"] != self.account_id and self.job_id is None`
- `instagram_strategy.py:706-724` — same, expressed as `not_self = sender_id != self.account_id`
- `facebook_strategy.py:787-797` — the comment equivalent, skipping notifications when `payload["from"][0]["id"] == self.account_id`

This is the right layer for the reopen call. Only Facebook and Instagram ingest DMs; LinkedIn and YouTube are comments only, GMB is reviews only.

**The "un-done" semantics to mirror.** `update_inbox_pending` (`inbox_details_repository.py:2046-2148`) is the existing Mark-as-pending action and defines what un-done means:

- `marked_done.is_marked_done = False`
- `marked_done.action_performed_by = None`
- `pending.is_pending = True`
- `updated_at` refreshed
- inserts a system entry into the thread with `action.type = "PENDING"` and pushes it as a `new_message` Pusher event

## Frontend behaviour (no change needed for the main flow)

`contentstudio-frontend/src/modules/inbox-revamp/` already routes a reopened conversation correctly, because it re-reads the flags off the payload:

- `views/InboxView.vue:1018-1031` binds the `new_message` Pusher event and calls `updateLastMessage`.
- `composables/useInboxChat.ts:790-839` — if the conversation is not in the currently rendered list, it calls `fetchTargetedConversation`, then `shouldUpdateConversation`, then `pushNewConversation` to insert it at the top.
- `composables/useInboxChat.ts:935-1020` — `shouldUpdateConversation` gates the Unassigned tab on `!assigned.is_assigned && !archived.is_archived && !markedDone.is_marked_done`.

So once the server clears `is_marked_done`, a user sitting on **Unassigned** sees the conversation reappear at the top in real time, with no frontend change.

### Known frontend gap (follow-up, not in this story)

`shouldUpdateConversation` returning `false` causes an early `return` — it never **removes** a row that no longer belongs to the current tab (`useInboxChat.ts:806-808`, `833-835`). Consequence: a user sitting on **Marked as Done** when a conversation gets reopened keeps seeing the stale row until they switch tabs or refresh. Sidebar counts have the same lag, since `fetchDetails` (`useInboxChat.ts:125-164`) only refreshes counts on an explicit list fetch, not on a realtime event. Cosmetic, and the opposite direction of the reported bug, so it is called out here rather than folded into the story.

## What Needs to Change

- On a **new inbound** message on a Facebook or Instagram conversation, clear the conversation's Done state before the notification fires.
- On a **new inbound** review update on a GMB review, do the same.
- Trigger on inbound only. Skip when the sender is the connected account itself (agent reply, auto-reply, platform echo of an outbound send), and skip during backfill jobs (`self.job_id is not None`), matching the guards already used for notifications.
- Mirror `update_inbox_pending`: clear `is_marked_done`, null out `marked_done.action_performed_by`, set `pending.is_pending = True`, refresh `updated_at`.
- Leave assignment untouched, so an assigned-and-done conversation returns to **Assigned** / **Mine**, not **Unassigned**.
- Leave `archived.is_archived` untouched. Archive is a deliberate suppression, and the tab queries already exclude archived from Marked as Done.
- Delete the two commented-out blocks and their dead `find_by_element_id` calls once the real implementation lands, so there is one code path.
- Post comments need no change: each new comment is inserted as its own row with `is_marked_done: False` (`inbox_comments_repository.py:141-151`), so it already lands in Unassigned. Only an edit/re-sync of an existing comment hits the preserve-status path, which is correct.

## Files Involved

| File | Change |
|---|---|
| `social-inbox-manager/app/social_sync/facebook_strategy.py` (`add_update_messages`, ~884) | call reopen on new inbound message |
| `social-inbox-manager/app/social_sync/instagram_strategy.py` (`add_update_messages`, ~706) | same |
| `social-inbox-manager/app/social_sync/gmb_strategy.py` (`add_update_inbox_details`, ~330) | reopen on new inbound review activity |
| `social-inbox-manager/app/database/mongo/repository/inbox_details_repository.py` | new reopen method mirroring `update_inbox_pending` semantics |
| `social-inbox-manager/app/database/mongo/repository/inbox_messages_repository.py:67-94` | remove commented-out block + dead lookup |
| `social-inbox-manager/app/database/mongo/repository/inbox_comments_repository.py:113-136` | remove commented-out block + dead lookup |

## Gotchas

- The reply-and-mark-done flow (`sends.py:383-396`) is the reason the inbound check is not optional. Without it, the platform echo of the agent's own reply reopens the conversation the instant it is closed.
- The webhook message path re-fetches the entire conversation (`facebook_strategy.py:431-468`) and pushes it through `_update_existing_element`. If the reopen writes to Mongo before that update lands, the preserved-`inbox_details` copy will stamp `is_marked_done: True` back on. Order the reopen after the conversation upsert, or have the reopen write only the specific `marked_done` / `pending` fields with `update_many` so it does not race a whole-document `$set`.
- `add_update_message` is also reached during backfill sync jobs, where every historical message is "new" to the database. The `self.job_id is None` guard prevents a sync from mass-reopening months of closed conversations.
