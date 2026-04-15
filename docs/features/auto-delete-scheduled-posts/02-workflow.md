# Auto-Delete Scheduled Posts — Workflow Design

---

## 1. Feature Placement

**Primary entry point:** Composer — below the First Comment toggle in `MainComposer.vue` (lines 738–846).

**Secondary surface:** Published post detail in the Planner — cancel or edit the scheduled deletion of an already-published post.

**Visibility in Planner:** Posts with a scheduled deletion show a clock/timer badge on the calendar card and list view item.

**Notification surface:** In-app notification centre — on successful deletion and on failure.

---

## 2. Happy Path — Setting Auto-Delete at Compose Time

1. User opens the Composer and creates a post.
2. User sets a publish date/time as normal.
3. Below the First Comment toggle, user sees an **"Auto-delete this post"** toggle (off by default).
4. User switches the toggle on.
5. A **"Delete post on"** date/time picker appears below the toggle, pre-filled to [publish datetime + 24 hours].
6. User selects the desired deletion date and time.
7. If any selected social account does not support API deletion (Instagram, TikTok, Facebook Groups, Facebook Stories, GMB Video), an **inline warning** appears:
   > *"Note: Due to API limitations, posts from Facebook Groups, Instagram, TikTok, Facebook Stories, and GBP video posts cannot be deleted directly from the social media platforms through ContentStudio. While you can remove these posts within ContentStudio, you will need to delete them manually on the respective platforms."*
8. User schedules or publishes the post. `scheduled_deletion_date` and `scheduled_deletion_enabled: true` are stored on the plan.
9. In the Planner, the post card shows a clock icon badge. Tooltip: *"Auto-deletes on [date] at [time]"*.
10. At the scheduled deletion time, `DeleteScheduledPostsCommand` fires.
11. For each supported platform, the existing platform-specific delete API is called.
12. On success: plan is soft-deleted in CS. In-app notification sent: *"Your post was auto-deleted from [platform(s)] on [date]."*
13. On failure: In-app notification sent: *"Auto-delete failed for [platform]. Please delete the post manually."* Plan is NOT deleted — it remains published so the user can act.

---

## 3. Alternative Flows

### 3a. Cancelling Auto-Delete Before It Fires
1. User finds the published post in the Planner (clock badge visible).
2. User opens post details and clicks **"Cancel auto-delete"**.
3. `scheduled_deletion_enabled` set to `false`. Clock badge removed.

### 3b. Editing the Deletion Date Before It Fires
1. User opens published post details in the Planner.
2. User clicks **"Edit"** next to the deletion date.
3. Date/time picker opens — new datetime must be in the future.
4. User saves — `scheduled_deletion_date` updated.

### 3c. Multi-Platform Post — Partial Support
1. User schedules a post to Facebook + Instagram.
2. Inline warning shown for Instagram at compose time.
3. At deletion time: Facebook deleted via API. Instagram skipped.
4. Plan soft-deleted in CS. Notification: *"Your post was auto-deleted from Facebook. Instagram does not support automatic deletion — please delete manually."*

### 3d. Deletion Fails (Token Expired / Post Boosted)
1. Cron fires. Platform API returns error.
2. Plan NOT deleted from CS.
3. In-app notification: *"Auto-delete failed for [platform]. Please delete the post manually."*
4. Retry: up to 3 attempts with 5-minute intervals. After 3 failures, permanent failure notification sent.

### 3e. Post Already Manually Deleted on Platform
1. Cron fires. Platform API returns 404.
2. Treat as success — desired outcome is achieved.
3. Plan soft-deleted in CS. Success notification sent.

### 3f. Post Never Published (Failed to Publish)
1. Cron fires at scheduled deletion time.
2. Command checks `status == published`. Status is `failed` — deletion skipped.
3. No notification sent.

### 3g. Repeat/Evergreen Post
1. Cron fires. Command detects `repeat_post = true`.
2. Deletion skipped — auto-delete does not apply to repeat posts.

---

## 4. Key Design Decisions

### Decision 1: Single deletion path — no "CS only" option
Auto-delete always deletes from the live platform. There is no "CS only" path in the auto-delete composer toggle. The existing manual delete modal (with its two options) is unchanged and remains available for manual deletion use cases.

**Why:** The feature's entire value is removing time-sensitive content from the audience's view. "CS only" doesn't solve that. Users who want to clean up their CS calendar can manually delete at any time.

### Decision 2: No new `expired` status for V1
Auto-deleted posts use the existing soft-delete flow and disappear from the Planner the same way manually deleted posts do.

**Why:** Minimises backend change for V1. Adding a dedicated `expired` status + Planner filter is a clean V2 addition once usage patterns are understood.

### Decision 3: In-app notification only (no email for V1)
Success and failure notifications are in-app only. Email deferred to V2.

**Why:** `PostNotification` already supports both channels — email can be added in a single follow-up story. Keeps V1 lean.

### Decision 4: Minimum deletion window is publish time + 1 hour
The earliest a user can set the deletion datetime is 1 hour after the scheduled publish time. Enforced by the date/time picker.

---

## 5. Integration with Existing Features

| Feature | Integration |
|---|---|
| Composer scheduling | Auto-delete picker sits below First Comment toggle. Publish datetime is the minimum bound. |
| Planner calendar/list | Clock badge on posts with pending auto-delete. Deleted posts disappear as they do today (V1). |
| Manual delete modal | Unchanged. Still available for one-off manual deletions with the two existing options. |
| Post notifications | Reuses `PostNotification` + `PostMail`. Two new types in `config/notifications.php`. |
| Repeat/evergreen posts | Auto-delete skipped for plans where `repeat_post = true`. |
| Approval workflow | Auto-delete scheduling preserved through approval flow. Fires after publish regardless. |

---

## 6. V1 vs V2 Scope

### V1 (This Epic)
- Composer: toggle + date/time picker + inline platform warning
- Backend: `scheduled_deletion_enabled` + `scheduled_deletion_date` fields on Plans
- Backend: `DeleteScheduledPostsCommand` cron + existing platform delete API calls
- In-app notifications: `post_auto_deleted` and `post_auto_delete_failed`
- Planner: clock badge on posts with pending auto-delete
- Retry logic: 3 attempts on failure

### V2 (Post-Launch)
- Email notification on deletion success/failure
- `expired` status + Planner filter for auto-deleted posts
- Cancel/edit deletion date from Planner post detail view
- Performance-based deletion conditions (premium tier)
- Bulk auto-delete assignment from Planner
- Account-level expiry presets
- "Re-schedule" action on auto-deleted posts
