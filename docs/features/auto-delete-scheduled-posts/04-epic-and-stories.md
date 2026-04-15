# Auto-Delete Scheduled Posts — Epic & Stories

---

## Epic

**Title:** Auto-Delete Scheduled Posts

**Description:**

ContentStudio users — particularly those running e-commerce campaigns, event-driven accounts, and agency clients — regularly publish time-sensitive content that must be removed after a deadline. Today there is no way to automate this cleanup: users must manually log into each platform and delete posts one by one, a process that is tedious, error-prone, and frequently skipped, leaving outdated content live and damaging brand credibility.

This epic introduces scheduled post deletion: a toggle in the Composer that lets users set a date and time for a post to be automatically removed from the live social platform after publishing. ContentStudio handles the deletion via existing platform APIs, then sends an in-app notification confirming the outcome. Posts are soft-deleted in ContentStudio so the record and any analytics data are preserved.

V1 supports deletion on Facebook Pages, Twitter/X, LinkedIn, YouTube, Pinterest, Google Business Profile (non-video), Threads, Bluesky, and Tumblr. Instagram and TikTok are excluded due to API limitations and are called out clearly in the UI at compose time.

---

## Stories

---

### Story 1: [BE] Add scheduled deletion fields and API support to the Plans model

**Description:**
As a ContentStudio backend, I need to store scheduled deletion settings on a plan so that the auto-delete cron job can identify and process posts that are due for deletion.

---

**Workflow:**

1. User creates or edits a post in the Composer with the auto-delete toggle switched on and a deletion date/time set.
2. User saves, schedules, or immediately publishes the post.
3. The plan is saved with `scheduled_deletion_enabled: true` and `scheduled_deletion_date` (UTC timestamp) stored on the Plans document.
4. The plan is returned via the API with these fields so the frontend can display the deletion schedule on the Planner and in the post detail view.
5. When the user cancels or edits the deletion schedule from the Planner post detail (P1), the plan is updated — `scheduled_deletion_enabled` is set to `false` or `scheduled_deletion_date` is updated to the new value.

---

**Acceptance criteria:**

- [ ] `scheduled_deletion_enabled` (boolean, default `false`) is added to `Plans.$fillable`
- [ ] `scheduled_deletion_date` (UTC datetime string, nullable) is added to `Plans.$fillable`
- [ ] Plan create API accepts and persists `scheduled_deletion_enabled` and `scheduled_deletion_date`
- [ ] Plan update API accepts and persists `scheduled_deletion_enabled` and `scheduled_deletion_date`
- [ ] Both fields are returned in plan API responses
- [ ] `PlansRepository::fetchScheduledDeletionPlans()` returns all published plans where `scheduled_deletion_enabled = true` AND `scheduled_deletion_date <= Carbon::now('UTC')` AND `repeat_post` is not `true`
- [ ] `fetchScheduledDeletionPlans()` excludes plans where `status != 'published'`
- [ ] `config/notifications.php` includes a `post_auto_deleted` entry with `title`, `description`, `email_description`, and `subject` keys
- [ ] `config/notifications.php` includes a `post_auto_delete_failed` entry with the same keys

---

**Mock-ups:** N/A — backend only.

**Impact on existing data:** Two new optional fields added to the Plans MongoDB collection. Existing plan documents without these fields default to `scheduled_deletion_enabled: false` — no deletion will fire for them.

**Impact on other products:** None. Fields are additive; no existing queries are affected.

**Dependencies:** None.

---

**Global quality & compliance:**
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support — notification config strings added to `config/notifications.php`; ensure they follow the existing localisation pattern used by `NotificationLocalizationService`
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 2: [BE] Implement auto-delete cron job with platform deletion and in-app notifications

**Description:**
As a ContentStudio user who has set an auto-delete date on a published post, I want ContentStudio to automatically delete the post from the social media platform at the scheduled time and notify me in-app so that I don't have to remember to do it manually.

---

**Workflow:**

1. A post was published with `scheduled_deletion_enabled: true` and a `scheduled_deletion_date` that has now passed.
2. The `DeleteScheduledPostsCommand` cron job runs every minute (same cadence as `PlanPostingCommand`).
3. The command calls `PlansRepository::fetchScheduledDeletionPlans()` to retrieve eligible plans.
4. For each plan, the command calls the existing platform-specific delete API for each platform the post was published to. Unsupported platforms (Instagram, TikTok, Facebook Groups, Facebook Stories, GMB Video) are skipped silently — no API call is made for them.
5. On full success: the plan is soft-deleted in ContentStudio. An in-app notification is dispatched via `PostNotification` with type `post_auto_deleted`.
6. On partial success (some platforms succeeded, some failed): the plan is soft-deleted in CS. An in-app notification is sent indicating which platforms deleted and which require manual action.
7. On full failure: the plan is NOT deleted. An in-app notification is dispatched with type `post_auto_delete_failed`.
8. The command retries failed platforms up to 3 times at 5-minute intervals before issuing the final failure notification.
9. If a platform returns 404 (post already manually deleted), this is treated as success — the desired outcome is achieved.
10. The command is registered in `app/Console/Kernel.php` to run `everyMinute()`.

---

**Acceptance criteria:**

- [ ] `DeleteScheduledPostsCommand` (`plan:delete-scheduled`) exists and is registered in `app/Console/Kernel.php` to run `everyMinute()`
- [ ] Command fetches eligible plans using `PlansRepository::fetchScheduledDeletionPlans()`
- [ ] Command calls the existing platform delete API for each supported platform the post was published to
- [ ] Unsupported platforms (Instagram, TikTok, Facebook Groups, Facebook Stories, GMB Video) are skipped without error
- [ ] On full success: plan is soft-deleted (`deleted_at` set); `post_auto_deleted` in-app notification is sent to the plan owner
- [ ] On full failure after 3 retries: plan remains in `published` state; `post_auto_delete_failed` in-app notification is sent
- [ ] On partial success: plan is soft-deleted; notification indicates which platforms succeeded and which failed
- [ ] Platform 404 response is treated as success
- [ ] Plans with `repeat_post = true` are never processed
- [ ] Plans with `status != 'published'` are never processed
- [ ] A `deletion_attempts` counter is tracked per platform per plan; deletion stops after 3 failed attempts per platform
- [ ] Command execution is logged using `LogsBuilder` (consistent with `PlanPostingCommand` pattern)
- [ ] `post_auto_deleted` notification includes `plan_id` so the user can navigate to the post in the Planner via the notification link
- [ ] `post_auto_delete_failed` notification includes `plan_id`

---

**Mock-ups:** N/A — backend only.

**Impact on existing data:** Soft-deletes plans on schedule using the existing soft-delete mechanism. Existing `PostNotification` class is reused with the two new notification types added in **[BE] Add scheduled deletion fields and API support to the Plans model**.

**Impact on other products:** In-app notifications appear in the notification centre on web. No impact on mobile apps or Chrome extension.

**Dependencies:** Depends on **[BE] Add scheduled deletion fields and API support to the Plans model** — the new fields and notification config must exist before this command can be built.

---

**Global quality & compliance:**
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support — notification dispatch calls `NotificationLocalizationService`; ensure `post_auto_deleted` and `post_auto_delete_failed` types are handled for user locale
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review — notifications must use white-label `app_url` and `business_name`; verify `PostNotification` correctly resolves these for the two new types
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 3: [FE] Add auto-delete scheduling UI to the Composer and Planner calendar badge

**Description:**
As a social media manager scheduling a time-sensitive post, I want to set an auto-delete date directly in the Composer and see a visual indicator on my Planner calendar so that I can be confident my promotional content will be removed automatically after it expires — without having to log into each platform manually.

---

**Workflow:**

1. User opens the Composer and creates a post to one or more social accounts.
2. User sets a publish date/time as normal.
3. Below the First Comment toggle, user sees the **"Auto-delete this post"** toggle (off by default).
4. User clicks the toggle to switch it on.
5. A **"Delete post on"** section expands below the toggle with a date/time picker pre-set to the publish datetime + 24 hours.
6. User selects their desired deletion date and time.
7. If any selected social account is on a platform that doesn't support API deletion (Instagram, TikTok, Facebook Groups, Facebook Stories, GBP Video), an inline warning appears below the picker.
8. User schedules or publishes the post. The deletion schedule is saved with the post.
9. In the Planner calendar and list view, the published post card shows a small clock badge. Hovering over it shows when the post will be deleted.
10. The post is deleted at the scheduled time. The card disappears from the Planner (same behaviour as a manually deleted post).

---

**UI Copy:**

**Toggle:**
- Label: `Auto-delete this post`
- Info icon (`ℹ`) tooltip (via `CstPopup` on hover, same pattern as First Comment toggle in `MainComposer.vue`):
  > *"Automatically remove this post from your social media profiles on a date you choose. Perfect for time-sensitive content like flash sales, limited-time offers, or event announcements that should disappear once they're no longer relevant. For example, if you're running a '50% off this weekend only' promotion, set the post to auto-delete on Monday morning."*

**Date/time picker section:**
- Section label: `Delete post on`
- Helper text below picker: `The post will be automatically removed from your connected social profiles at this date and time.`
- Picker: Reuse the existing date/time picker component used in the Composer scheduling flow. No new component needed.
- Default value on toggle activation: publish datetime + 24 hours
- Minimum selectable value: publish datetime + 1 hour (enforced by picker constraints)
- Validation error (datetime too close to publish): `"Deletion time must be at least 1 hour after the publish time."`
- Validation error (datetime in the past): `"Please select a future date and time."`

**Platform limitation inline warning:**
- Component: `Alert` from `@contentstudio/ui`, warning variant
- Copy: *"Note: Due to API limitations, posts from Facebook Groups, Instagram, TikTok, Facebook Stories, and GBP video posts cannot be deleted directly from the social media platforms through ContentStudio. While you can remove these posts within ContentStudio, you will need to delete them manually on the respective platforms."*
- Display condition: any selected account belongs to Instagram, TikTok, Facebook Groups, Facebook Stories, or GBP Video

**Planner clock badge:**
- Component: `Badge` from `@contentstudio/ui` with a clock/timer icon using the `Icon` component
- Shown on: post cards in Planner calendar view and list view
- Tooltip (via `CstPopup` on hover): *"This post will be automatically deleted on [Day, Month DD, YYYY] at [HH:MM AM/PM] [workspace timezone]."*
- Display condition: `scheduled_deletion_enabled = true` AND `status = published` AND `scheduled_deletion_date` is in the future

**In-app notification copy (for reference — rendered by backend, no FE work needed):**
- Success: `"Your post '[excerpt]' was automatically deleted from [platform(s)] as scheduled."`
- Failure: `"We couldn't automatically delete your post '[excerpt]' from [platform]. Please delete it manually."`
- Partial: `"Your post was deleted from [platforms]. It could not be deleted from [platforms] — please remove it manually."`

---

**Acceptance criteria:**

- [ ] "Auto-delete this post" toggle appears below the First Comment toggle in the Composer using `CstSwitch`, off by default
- [ ] Toggle has an `ℹ` info icon next to the label; hovering shows the tooltip via `CstPopup`
- [ ] Switching toggle on expands the "Delete post on" date/time picker section
- [ ] Date/time picker defaults to publish datetime + 24 hours on toggle activation
- [ ] Date/time picker enforces a minimum of publish datetime + 1 hour
- [ ] Validation error "Deletion time must be at least 1 hour after the publish time." appears if the constraint is violated
- [ ] Validation error "Please select a future date and time." appears if a past datetime is entered
- [ ] Switching toggle off collapses the date/time section and clears `scheduled_deletion_date` from the form state
- [ ] `Alert` warning appears when any selected account is on a non-supported platform
- [ ] `Alert` warning is NOT shown when all selected accounts are on supported platforms
- [ ] `scheduled_deletion_enabled` and `scheduled_deletion_date` are included in the plan payload sent to the API
- [ ] Planner calendar view shows a clock `Badge` on published posts with a future `scheduled_deletion_date`
- [ ] Planner list view shows the same clock `Badge`
- [ ] Clock badge tooltip shows the deletion datetime in the workspace timezone
- [ ] Clock badge is not shown after the deletion datetime has passed
- [ ] All user-facing strings are added to all locale files under `contentstudio-frontend/src/locales/`
- [ ] Toggle and date/time section are responsive on smaller screen widths

---

**Component gaps:**
- No standalone `DateTimePicker` in `@contentstudio/ui`. Use the existing date/time picker component from the Composer scheduling flow. If it cannot be cleanly reused, flag to Design for a `DateTimePicker` design system addition.
- No standalone `Tooltip` in `@contentstudio/ui`. Use `CstPopup` for hover tooltips (consistent with existing First Comment info popover in `MainComposer.vue`).

**Mock-ups:** Use the existing First Comment toggle section (`MainComposer.vue` lines 738–846) as the layout/style reference. The auto-delete toggle must be visually consistent — same grid/icon/label/`CstSwitch` pattern.

**Impact on existing data:** Composer form state gains two new fields. Existing posts and compose flows are unaffected — fields default to `false`/`null`.

**Impact on other products:** Clock badge visible in Planner calendar and list views. No impact on mobile apps (web-only), Chrome extension, or other modules.

**Dependencies:** Depends on **[BE] Add scheduled deletion fields and API support to the Plans model** — API must accept the new fields before the frontend can send them.

---

**Global quality & compliance:**
- [ ] Mobile responsiveness — toggle and date picker must display and function correctly on narrow viewports (tablet/mobile browser)
- [ ] Multilingual support — all new i18n keys added to all locale files under `src/locales/`; no hardcoded English strings in template
- [ ] UI theming support — use `text-primary-cs-500`, `bg-primary-cs-50` etc. for primary-coloured elements; use `@contentstudio/ui` components via props/variants only; no hardcoded colour values
- [ ] White-label domains impact review — `Alert` and `Badge` use theme-aware components; no hardcoded colours
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Summary

| # | Title | Group | Skill Set | Product Area | Project | Priority |
|---|---|---|---|---|---|---|
| 1 | [BE] Add scheduled deletion fields and API support to the Plans model | Backend | Backend | Publishing | Web App | High |
| 2 | [BE] Implement auto-delete cron job with platform deletion and in-app notifications | Backend | Backend | Publishing | Web App | High |
| 3 | [FE] Add auto-delete scheduling UI to the Composer and Planner calendar badge | Frontend | Frontend | Composer | Web App | High |
