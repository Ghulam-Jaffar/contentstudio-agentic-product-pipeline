# Auto-Delete Scheduled Posts — Research Report

**Date:** 2026-04-15
**Feature Slug:** auto-delete-scheduled-posts

---

## What Is This Feature?

Auto-Delete Scheduled Posts allows users to set a future date and time at which a published post will be automatically removed from the live social media platform. The user configures the expiry at the time of scheduling (or before publication), and ContentStudio's backend triggers the platform API deletion when the scheduled deletion time arrives.

**Primary use cases:**
- Flash sales and limited-time offers (e.g., "50% off, ends midnight")
- Event announcements (delete after the event date passes)
- Holiday/seasonal promotions that become irrelevant after the holiday
- Compliance-driven content removal (regulated industries, contracted asset expiry)
- A/B-style posting (remove underperforming posts after a set window)
- Time-sensitive job listings, product launches, and giveaway announcements

**Why it matters for ContentStudio users:** Manually logging into each platform to delete expired posts across dozens of accounts is tedious and error-prone. Auto-deletion closes the loop on the publishing workflow and keeps brand feeds clean without human intervention.

---

## Competitor Analysis

| Competitor | Has Auto-Delete? | Key Capabilities | UX Approach | Unique Differentiator |
|---|---|---|---|---|
| **Publer** | Yes — full feature | Deletes from live platform; supports Facebook Pages, Twitter/X, LinkedIn, YouTube, Google Business Profile, Pinterest, Threads, Bluesky. NOT available for Instagram, TikTok, or Stories (API limitation). Also supports Auto-Hide for Facebook Page posts (not videos) and YouTube. | Icon in composer; "Edit Conditions" button for deletion timing + engagement/reach thresholds. Also configurable in Post Presets as account-wide defaults. Auto-delete icon visible on calendar posts. | Engagement-based conditions: delete if reach < N or engagements < N after X days. Performance-based deletion is a Business-tier-only feature. Keeps a copy in Publer after deletion. |
| **SocialBee** | Partial — "Expire Post" only stops re-queuing | Does NOT delete from live platform. Post is simply removed from the queue rotation, so it won't be republished. Post remains live. | Toggle in post settings; labeled "expire post." | Queue-management focused; no live platform deletion. |
| **Sprinklr** | Yes — enterprise | Deletes from live platform. Supported platforms: Facebook, Twitter/X, YouTube. Tied to Digital Asset Manager (DAM) — deletion triggered by asset contract expiry, not a standalone scheduler setting. | Configuration inside the Asset Manager, not directly inside the post composer. Positioned as a compliance/legal workflow. | Deep integration with DAM and content lifecycle management. Enterprise-grade compliance workflow rather than a user-facing self-service feature. |
| **Buffer** | No | No auto-delete feature confirmed. Manual deletion only. | N/A | N/A |
| **Hootsuite** | No | No auto-delete feature confirmed. | N/A | N/A |
| **Later** | No | No auto-delete feature confirmed. | N/A | N/A |
| **Sprout Social** | No | No auto-delete feature confirmed. | N/A | N/A |
| **Agorapulse** | No | Manual deletion of published posts is supported. No automatic expiry or auto-delete feature found in 2025 release notes or help documentation. | N/A | N/A |
| **Sendible** | No | Bulk manual deletion of scheduled (not yet published) posts is supported. No auto-delete from live platforms found in 2025 product roundup or documentation. | N/A | N/A |
| **Loomly** | No | Manual deletion from Loomly does NOT delete the post from the social media platform. No auto-delete feature found. | N/A | N/A |
| **Metricool** | No | No auto-delete from live platforms. Autolist posts are deleted from the queue after publication (internal queue cleanup), but the post stays live. | N/A | N/A |

**Summary:** Among SMB/mid-market social media management tools, **Publer is the only direct SMB-tier competitor with a live-platform auto-delete feature.** Sprinklr has it at the enterprise level, but as a compliance workflow tied to asset management, not a simple user-facing date-picker. All others (Buffer, Hootsuite, Later, Sprout Social, Agorapulse, Sendible, Loomly, Metricool) have confirmed no auto-delete capability.

---

## Platform API Support

| Platform | Delete via API? | Notes | ContentStudio Current Manual Delete Support |
|---|---|---|---|
| **Facebook Pages** | Yes | `DELETE /{post-id}`. Requires `pages_manage_posts` permission. Standard and well-documented. | Yes |
| **Twitter / X** | Yes | `DELETE /2/tweets/{id}`. Requires `tweet.write` OAuth 2.0 scope. | Yes |
| **LinkedIn** | Yes | `DELETE /ugcPosts/{postUrn}` or `DELETE /shares/{shareUrn}`. Requires `w_member_social` or `rw_organization_admin`. | Yes |
| **YouTube** | Yes | `DELETE videos?id={videoId}`. Requires `youtube.force-ssl` scope. | Yes |
| **Pinterest** | Yes | `DELETE /pins/{pin_id}` via Pinterest API v5. Official endpoint confirmed. | Yes |
| **Google Business Profile** | Partial | Local Post API supports `DELETE` for event/offer/standard posts. Video posts cannot be deleted via API. | Partial (video posts excluded) |
| **Threads** | Yes | Delete support added to Threads API on **March 6, 2025** (Meta changelog confirmed). Standard Meta API patterns. | Yes (recently added) |
| **Bluesky (AT Protocol)** | Yes | `deletePost(postUri)` function in AT Protocol SDK. Deletion is immediate in the app. Back-end hard-delete is performed periodically. | Yes |
| **Tumblr** | Yes | `DELETE /v2/blog/{blogIdentifier}/post` — official API v2 endpoint. JavaScript (`deletePost`) and Python (`delete_post`) SDKs confirmed. | Unknown |
| **Instagram** | No | Instagram Graph API does not support deleting published feed posts. Confirmed API limitation. | No (manual only) |
| **TikTok** | No | TikTok Content Posting API supports uploads and scheduling but does NOT support deleting published posts. Confirmed API limitation. | No (manual only) |
| **Facebook Groups** | No | Facebook Graph API does not expose deletion of group posts via third-party apps. | No |
| **Facebook Stories** | No | Stories deletion via third-party API not supported. | No |
| **GMB Video Posts** | No | Google Business Profile API does not support video post deletion. | No |

---

## Common Patterns

Based on how Publer, Sprinklr, and broader UX research approach this feature:

1. **Date-and-time picker at post creation:** The most common pattern. User sets an expiry datetime alongside the publish datetime. Simple and predictable — mirrors how scheduling itself works.

2. **Expiry configured separately from scheduling:** Some tools (Publer) surface auto-delete as an icon or secondary action in the composer, keeping the primary scheduling UI clean. Avoids overwhelming new users.

3. **Confirmation before permanent deletion:** Given the destructive nature of deleting live content, the best implementations surface a pre-deletion notification (e.g., email or in-app alert shortly before deletion fires).

4. **Keep a copy in the tool:** All implementations that delete from the live platform preserve the post record within the scheduling tool itself, with a status change (e.g., "Deleted" or "Expired"). This provides audit trail and allows users to re-publish or review performance before it was removed.

5. **Visual badge on calendar:** Publer shows a disabled eye icon on deleted posts in the calendar. Users have asked for a separate "hide deleted posts" filter — this is a documented UX gap in Publer.

6. **Platform-specific warnings at setup time:** For platforms where deletion is not supported (Instagram, TikTok), user-facing warnings are shown at the time of configuration, not silently at deletion time. This is critical to prevent user confusion.

7. **Account-level defaults (preset behavior):** Publer offers Post Presets for setting auto-delete on by default for specific accounts. This is high-value for power users managing many accounts with consistent time-sensitivity (e.g., an agency running flash-sale campaigns).

8. **Tiered access for advanced conditions:** Basic time-based expiry is available on lower plans; performance-based conditions (delete if reach < N) reserved for higher tiers (Publer's Business plan). This is a natural monetization lever.

---

## Differentiators

### Where ContentStudio Can Lead Over Publer

- **Calendar filter for expired posts:** Publer has a documented gap — users want a way to hide/filter deleted posts from the calendar view. ContentStudio can implement this cleanly from launch with a "Deleted/Expired" status filter in the Planner.
- **Expiry notification before deletion:** A proactive in-app or email reminder (e.g., "Your post on Facebook will be auto-deleted in 1 hour") is not explicitly documented in Publer. ContentStudio can differentiate by building this in.
- **Bulk expiry settings:** Allow applying an auto-delete date to multiple posts in one action (bulk edit in Planner). Publer does not highlight this capability.
- **Re-publish from expired post:** Surface a one-click "Re-schedule" option on expired posts in the Planner. Useful for evergreen content that was only time-sensitive once.
- **Expiry templates / presets per campaign type:** Allow users to save "Campaign Presets" that bundle a post type with a standard expiry window (e.g., "Flash Sale = delete after 24 hours").

### Where ContentStudio Is Already Ahead
- ContentStudio already supports manual deletion across more platforms (including Threads, which was just added in 2025), giving the backend delete capability that auto-delete simply needs to schedule. This makes the implementation largely a scheduling layer on top of existing infrastructure.

---

## User Expectations

Based on the use cases, competitor gaps (especially Publer's feedback board requesting "hide deleted posts"), and general UX research:

1. **Simplicity first:** Users expect to set an expiry date/time the same way they set a publish date/time — a familiar datetime picker in the composer. No complex conditions by default.

2. **Transparency:** Users expect to know which posts have an expiry set. A visual indicator on the calendar post card (e.g., a clock/timer icon) is strongly expected.

3. **No silent failures:** If a deletion fails (e.g., post was already deleted manually, token expired), users expect an in-app notification or email alert — not silent failure.

4. **Post remains in ContentStudio after deletion:** Users universally expect the post record to remain in ContentStudio with a status like "Expired" or "Auto-Deleted" — not to disappear entirely. This preserves analytics history and allows performance review.

5. **Undo / recovery window:** Users expect at minimum a warning that the deletion is permanent, and ideally a grace period or cancellation option (e.g., "cancel auto-delete" up until 5 minutes before scheduled deletion).

6. **Platform limitation clarity at setup:** Users expect to be told upfront (at composer time) whether auto-delete is supported for a given platform selection. They do not want to discover the limitation after the deletion fails.

7. **Planner filter:** Users managing large volumes of posts expect to be able to filter out expired/deleted posts from the Planner calendar view to keep it uncluttered.

8. **Mobile responsiveness:** The auto-delete configuration UI should work on mobile-sized screens since social media managers frequently work on mobile.

---

## Recommended Approach for ContentStudio

### Scope for V1

**Supported platforms (auto-delete from live platform):**
Facebook Pages, Twitter/X, LinkedIn, YouTube, Pinterest, Google Business Profile (non-video), Threads, Bluesky, Tumblr

**Unsupported (show warning, disable option):**
Instagram, TikTok, Facebook Groups, Facebook Stories, GMB Video

### Core UX Flow

1. In the Composer, below the scheduling date-time picker, add an optional "Auto-delete" toggle/section — collapsed by default.
2. When toggled on, show a second date-time picker labeled "Delete post on" with a minimum value of [publish datetime + 1 hour].
3. For any selected accounts where auto-delete is NOT supported (Instagram, TikTok), show an inline warning badge: "Auto-delete is not supported for [platform]. This post will remain live after publishing."
4. On the Planner calendar, posts with auto-delete scheduled show a small clock/timer icon badge. Tooltip: "Auto-deletes on [date/time]."
5. After deletion fires: post status changes to "Auto-Deleted" in Planner. Post card is visually dimmed/muted. All analytics data is retained.
6. Add "Auto-Deleted" as a filterable status in Planner (alongside Published, Scheduled, Draft, Failed).
7. Send an in-app notification (and optional email) when a post is auto-deleted: "Your [platform] post '[title excerpt]' was auto-deleted on [date/time]."
8. If deletion fails (API error, revoked token, post already removed), send an in-app alert: "Auto-delete failed for [platform] post. Please delete manually."

### V2 Considerations (post-launch)

- Performance-based deletion conditions (delete if reach < N or engagement < N) — premium/higher-tier feature
- Bulk auto-delete setting in Planner (select multiple posts, set the same expiry)
- Account-level default expiry presets (per social account)
- "Re-schedule" one-click action on auto-deleted posts
- Campaign-level expiry templates

### Implementation Notes

- ContentStudio already has manual delete infrastructure across all supported platforms. Auto-delete is architecturally a **scheduled job layer** on top of existing delete endpoints — the backend difference is a `delete_at` timestamp stored with the post and a background worker (cron/queue) that polls and fires deletions.
- The `delete_at` field needs to be surfaced in the API and stored in the post model.
- Background worker reliability is critical: if the worker misses a deletion window (e.g., queue backlog), users need to see a failure notification rather than a silent miss.
- Threads delete support (added March 2025) aligns exactly with ContentStudio's existing Threads manual delete capability — auto-delete can piggyback immediately.

---

## Sources

- [Publer Auto-Delete Help Center](https://publer.com/help/en/article/how-to-auto-delete-posts-1t21ddd/)
- [Publer Auto-Delete by Performance Blog](https://publer.com/blog/auto-comment-share-delete-posts-by-performance/)
- [Sprinklr Automatic Delete Posts on Asset Expiry](https://www.sprinklr.com/help/articles/advanced-capabilities/automatic-delete-posts-on-asset-expiry/645773750104980882a57968)
- [Threads API Changelog — March 6, 2025: Delete Support Added](https://www.threads.com/@threadsapi.changelog/post/DG4GAQtBRTU/-threads-api-updatemarch-6-2025-support-for-deleting-posts-has-been-added-see-de)
- [Meta Threads API — Posts Documentation](https://developers.facebook.com/docs/threads)
- [Bluesky AT Protocol API — deletePost](https://github.com/bluesky-social/atproto/blob/main/packages/api/README.md)
- [Bluesky Disappearing Posts Discussion](https://github.com/bluesky-social/atproto/discussions/2388)
- [Pinterest API v5 — Delete Pin](https://developers.pinterest.com/docs/api/v5/pins-delete/)
- [Tumblr API v2 — Delete Post](https://www.tumblr.com/docs/en/api/v2)
- [TikTok Content Posting API Overview](https://developers.tiktok.com/products/content-posting-api/)
- [ContentStudio — Delete Published Posts Help Article](https://docs.contentstudio.io/article/1041-delete-published-posts-from-social-media-platforms-through-contentstudio)
- [Agorapulse 2025 Release Notes](https://support.agorapulse.com/en/articles/11419511-agorapulse-release-notes-2025)
- [Sendible 2025 Product Round Up](https://www.sendible.com/insights/whats-coming-in-2025)
- [Threads Chief Proposes Auto-Archiving Posts After 30 Days — Social Media Today](https://www.socialmediatoday.com/news/threads-chief-proposes-archiving-user-posts-30-days/708853/)
