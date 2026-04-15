# PRD: Auto-Delete Scheduled Posts

**Author:** Product  
**Last Updated:** 2026-04-15  
**Status:** Draft  
**Target Release:** Q2 2026

---

## 1. Overview

Auto-Delete Scheduled Posts lets ContentStudio users set a date and time for a published post to be automatically removed from the live social media platform. Users configure the deletion schedule in the Composer alongside their publish time; ContentStudio's backend fires the deletion at the right moment via existing platform APIs and sends an in-app notification confirming the outcome. This closes the publishing loop for time-sensitive content — flash sales, event announcements, limited-time offers — without requiring users to manually log into each platform to clean up after a campaign ends.

---

## 2. Problem Statement

**What problem are we solving?**

Social media managers regularly publish time-sensitive content — limited-time promotions, event posts, seasonal campaigns — that becomes irrelevant or misleading once the window passes. Today, cleaning up these posts requires manually logging into each social platform and deleting them one by one. For teams managing multiple brands, dozens of accounts, and high post volumes, this cleanup is tedious, error-prone, and frequently skipped. Outdated posts left live damage brand credibility and confuse audiences (e.g., a "50% off today only" post still visible three days later).

**Who has this problem?**

- **E-commerce brands and agencies** running flash sales, promotional campaigns, and limited-time offers
- **Event-driven accounts** (venues, conferences, sports teams) publishing event announcements that expire after the event date
- **Regulated industries** (finance, healthcare, legal) with compliance obligations to remove content after a defined retention period
- These users skew toward Growth and Agency plan customers who manage multiple accounts at scale

**What happens if we don't solve it?**

- Users remain dependent on manual cleanup workflows that don't scale
- Publer — the only SMB-tier competitor with this feature — gains a clear differentiator for acquisition in the agency and e-commerce segment
- Support burden from users asking how to bulk-delete published posts continues
- Brand risk for customers whose outdated posts remain live unintentionally

---

## 3. Goals & Success Metrics

| Goal | Metric | Target | How We'll Measure |
|---|---|---|---|
| Drive adoption of the feature | % of new posts created with auto-delete enabled | 10% within 90 days of launch | Product analytics (plan creation events) |
| Reduce manual post cleanup burden | Support tickets related to deleting published posts | -20% within 60 days | Intercom/support data |
| Reliable deletion execution | % of scheduled deletions that succeed on first attempt | ≥95% | Backend logs / deletion job success rate |
| No increase in churn from post-deletion confusion | Monthly churn delta | <1% change | Billing data |

---

## 4. Target Users

**Primary Persona: The Campaign Manager**
A social media manager or agency account manager running time-boxed campaigns (flash sales, promotions, events). Posts to multiple platforms simultaneously. Works across 5–20+ social accounts. Needs content to go live at the right moment and disappear at the right moment without manual intervention.

**Secondary Persona: The Compliance-Aware Publisher**
Works in a regulated industry (finance, legal, healthcare) or manages white-label brands with contractual content retention requirements. Needs posts removed after a defined window as a business or legal requirement, not just a preference.

**Non-Users (explicitly out of scope):**
- Users publishing evergreen content with no expiry intent
- Mobile app users (web-only feature for V1; no iOS/Android stories)

---

## 5. User Stories / Jobs to Be Done

| ID | As a... | I want to... | So that... | Priority |
|---|---|---|---|---|
| US-1 | Campaign manager | set a deletion date/time when scheduling a post | my promotional post is automatically removed after the offer expires | P0 |
| US-2 | Campaign manager | see a visual indicator on my Planner calendar for posts with auto-delete scheduled | I know which posts will be removed and when, without having to open each one | P0 |
| US-3 | Social media manager | receive an in-app notification when a post is auto-deleted | I have a record that the deletion happened as planned | P0 |
| US-4 | Social media manager | receive an in-app alert if an auto-deletion fails | I can take manual action before the post causes problems | P0 |
| US-5 | Content creator | be warned at compose time if auto-delete is not supported for a selected platform | I know upfront which platforms require manual cleanup | P0 |
| US-6 | Campaign manager | cancel a scheduled auto-deletion after a post is published | I can keep a post live if plans change after publishing | P1 |
| US-7 | Campaign manager | edit the deletion date/time after a post is published | I can extend or shorten the live window without republishing | P1 |

---

## 6. Requirements

### 6.1 Must Have (P0)

- Toggle in the Composer ("Auto-delete this post") below the First Comment toggle, off by default
- When toggled on, a date/time picker appears ("Delete post on") with a minimum value of publish datetime + 1 hour and a default of publish datetime + 24 hours
- `scheduled_deletion_enabled` and `scheduled_deletion_date` fields stored on the Plans document
- Inline warning displayed at compose time for any selected account where API deletion is not supported: Instagram, TikTok, Facebook Groups, Facebook Stories, GMB Video — using the exact copy: *"Note: Due to API limitations, posts from Facebook Groups, Instagram, TikTok, Facebook Stories, and GBP video posts cannot be deleted directly from the social media platforms through ContentStudio. While you can remove these posts within ContentStudio, you will need to delete them manually on the respective platforms."*
- `DeleteScheduledPostsCommand` cron job that queries published plans where `scheduled_deletion_enabled = true` AND `scheduled_deletion_date <= now()`, calls existing platform delete APIs, then soft-deletes the plan in CS
- Supported platforms for auto-deletion: Facebook Pages, Twitter/X, LinkedIn, YouTube, Pinterest, Google Business Profile (non-video), Threads, Bluesky, Tumblr
- Auto-delete skipped for plans where `repeat_post = true` (repeat/evergreen posts)
- Auto-delete skipped for plans where `status != published`
- Clock/timer badge on Planner calendar and list view cards for posts with pending auto-delete; tooltip shows deletion datetime
- In-app notification `post_auto_deleted` on successful deletion
- In-app notification `post_auto_delete_failed` on failure
- Retry logic: up to 3 attempts at 5-minute intervals before final failure notification
- Platform 404 on deletion treated as success (post already manually deleted)

### 6.2 Should Have (P1)

- Cancel auto-delete from published post detail view in Planner (`scheduled_deletion_enabled` set to false)
- Edit deletion date/time from published post detail view in Planner (new datetime must be in the future)
- For multi-platform posts with partial support: delete from supported platforms, skip unsupported, send a combined notification indicating which platforms were deleted and which require manual action

### 6.3 Nice to Have (P2)

- Pre-deletion reminder notification ("Your post will be auto-deleted in 1 hour") — V2
- Clock badge shows countdown in tooltip (e.g., "Auto-deletes in 2 days, 4 hours") — V2

### 6.4 Explicitly Out of Scope

- "CS only" auto-delete option (no platform deletion) — the existing manual delete modal handles this use case
- Email notifications for deletion success/failure — deferred to V2
- `expired` Planner status and filter — deferred to V2
- Performance-based deletion conditions (delete if reach/engagement < threshold) — V2
- Bulk auto-delete assignment from Planner — V2
- Account-level expiry presets — V2
- Mobile (iOS/Android) support — web only for V1
- Auto-delete for repeat/evergreen posts

---

## 7. User Flow (High Level)

1. User opens the Composer and composes a post to one or more social accounts
2. User sets a publish date/time as normal
3. User toggles on **"Auto-delete this post"** (below First Comment toggle)
4. **"Delete post on"** date/time picker appears (default: publish time + 24h)
5. User sets the desired deletion datetime
6. If any selected account is on an unsupported platform, inline warning is displayed
7. User schedules/publishes the post — `scheduled_deletion_enabled` and `scheduled_deletion_date` saved to the plan
8. Post appears in Planner with a clock badge; tooltip: "Auto-deletes on [date] at [time]"
9. At the scheduled deletion time, `DeleteScheduledPostsCommand` runs
10. Platform delete API is called for each supported platform
11. Plan is soft-deleted in CS; in-app notification sent confirming deletion
12. If deletion fails: plan remains published; in-app failure notification sent; up to 3 retries

---

## 8. Business Rules & Constraints

| Rule ID | Rule | Rationale |
|---|---|---|
| BR-1 | Deletion datetime must be at least 1 hour after the scheduled publish datetime | Prevents accidental immediate deletion; ensures post has a minimum visibility window |
| BR-2 | Auto-delete only fires when `status = published` | A post that failed to publish should never trigger a deletion attempt |
| BR-3 | Auto-delete is skipped for plans where `repeat_post = true` | Deleting a repeat post would silently break the recurring post cycle |
| BR-4 | Platform 404 on deletion is treated as success | Post already deleted manually — the desired outcome is achieved; no failure alert needed |
| BR-5 | Maximum 3 retry attempts at 5-minute intervals on failure | Balances reliability against hammering a broken token or rate-limited API |
| BR-6 | For multi-platform posts, deletion proceeds independently per platform | A failure on one platform should not block deletion on others |
| BR-7 | Auto-delete is not available for Instagram, TikTok, Facebook Groups, Facebook Stories, GMB Video posts | Hard API limitation — these platforms do not expose a delete endpoint to third-party apps |
| BR-8 | The inline warning copy at compose time is fixed: "Note: Due to API limitations, posts from Facebook Groups, Instagram, TikTok, Facebook Stories, and GBP video posts cannot be deleted directly from the social media platforms through ContentStudio. While you can remove these posts within ContentStudio, you will need to delete them manually on the respective platforms." | Consistent with the existing manual delete modal language already familiar to users |
| BR-9 | Deletion datetime is stored and evaluated in UTC | Consistent with how `execution_time.date` is stored on the Plans model |

---

## 9. Open Questions

| Question | Options | Owner | Due Date | Decision |
|---|---|---|---|---|
| Should cancelling/editing auto-delete be available in V1 from the Planner post detail, or deferred to V2? | V1 (P1 requirement) / V2 | Product | Sprint planning | Included as P1 in this PRD — dev to assess effort |
| Should the clock badge be shown on all Planner views (calendar, list, feed) or only calendar and list? | All views / Calendar + list only | Product / Design | Design review | TBD |
| What is the exact in-app notification copy for partial deletion (some platforms succeeded, some failed)? | Single combined message / Separate per-platform messages | Product | Before dev starts | TBD — suggest single combined message |

---

## 10. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| OAuth token expires before deletion time (tokens last 60 days; deletions could be set months out) | Medium | High | Failure notification prompts user to reconnect account; retry after token refresh |
| Post is boosted as a Facebook ad at deletion time — Facebook API rejects deletion | Low | Medium | Treat as failure; send notification: "Auto-delete failed — this post may be associated with an active ad campaign. Please delete manually." |
| Cron job downtime causes missed deletion windows | Low | Medium | Query uses `<= NOW()` not `== NOW()` — missed windows are caught on the next cron run |
| User confusion: auto-deleted post disappears from Planner without clear indication | Medium | Medium | In-app notification provides confirmation; clock badge visible until deletion fires |
| High volume of simultaneous deletions (e.g., end of a campaign with many posts) causes rate limiting | Low | Low | Platform rate limits are per-account; deletions are per-plan so naturally distributed |
| User sets deletion time = publish time + 1 minute on a queued post; post publishes late due to queue backlog and deletion fires before post goes live | Low | Medium | BR-2 enforces `status = published` check before deletion fires — if post hasn't published yet, deletion is skipped |

---

## 11. Dependencies

**Internal:**
- `Plans` model (`contentstudio-backend/app/Models/Publish/Planner/Plans.php`) — requires two new fields in `$fillable`
- `PlansRepository` (`contentstudio-backend/app/Repository/Publish/Planner/PlansRepository.php`) — new `fetchScheduledDeletionPlans()` method
- `PlanPostingCommand` (`contentstudio-backend/app/Console/Commands/Planner/PlanPostingCommand.php`) — pattern for the new `DeleteScheduledPostsCommand`
- `PostNotification` (`contentstudio-backend/app/Notifications/Publish/PostNotification.php`) — reused for new notification types
- `config/notifications.php` — new `post_auto_deleted` and `post_auto_delete_failed` entries
- `MainComposer.vue` (`contentstudio-frontend/src/modules/composer_v2/components/MainComposer.vue`) — toggle insertion point
- `composerInitialState.js` (`contentstudio-frontend/src/modules/composer_v2/views/composerInitialState.js`) — new state fields

**External:**
- Facebook Graph API (`DELETE /{post-id}`) — requires `pages_manage_posts`
- Twitter/X API (`DELETE /2/tweets/{id}`) — requires `tweet.write`
- LinkedIn API (`DELETE /ugcPosts/{urn}`) — requires `w_member_social`
- YouTube Data API (`DELETE videos?id=`) — requires `youtube.force-ssl`
- Pinterest API v5 (`DELETE /pins/{pin_id}`)
- Threads API (delete support added March 6, 2025)
- Bluesky AT Protocol (`deletePost`)
- Tumblr API v2 (`DELETE /v2/blog/{id}/post`)

**Blockers:**
- None. All platform delete endpoints are already used by the existing manual delete flow in ContentStudio. No new API integrations required.

---

## 12. Appendix

- Research doc: `docs/features/auto-delete-scheduled-posts/01-research.md`
- Workflow doc: `docs/features/auto-delete-scheduled-posts/02-workflow.md`
- Competitor reference: Publer auto-delete — https://publer.com/help/en/article/how-to-auto-delete-posts-1t21ddd/
- Existing delete modal UI: `contentstudio-frontend` — Planner delete flow with two-option modal (reference for consistent platform limitation copy)
- Existing cron pattern: `PlanPostingCommand` — exact pattern for `DeleteScheduledPostsCommand`
- Original user request (Frill): https://contentstudio.frill.co/roadmap/allow-scheduling-of-automatic-deletion-for-scheduled-facebook-posts
- Shortcut placeholder story: https://app.shortcut.com/contentstudio-team/story/115204

---

## Changelog

| Date | Author | Changes |
|---|---|---|
| 2026-04-15 | Product | Initial draft |
