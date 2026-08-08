# Research — Social Listening fixes and Blog and Website Mentions to production

Item 22 of the 7 Aug 2026 backlog batch. Requested explicitly as **one ticket**.

## Part A: delete-topic confirmation

`contentstudio-frontend/src/modules/listening/composables/useListeningTopics.ts`:

- `deleteTopic(topic)` at line 222 calls `deleteMutation.mutateAsync(...)` at line 233, then fires `trackUserMaven('listening_topic_deleted')` and shows a success toast (`listening.toasts.topic_deleted`).
- Nothing in that path asks the user to confirm.

The delete is triggered from two places, both of which emit straight through:

- `components/topics/ListeningTopicListItem.vue:96` — `@click="$emit('delete', topic)"`
- `components/nav/ListeningViewsListContent.vue:179` — `@click="$emit('delete-view', view.id)"`

So a single click in a menu deletes a topic and everything gathered under it. Confirmation patterns exist elsewhere in the app to copy from, including the listening module's own removal confirmations elsewhere in the product.

## Part B: topic sidebar scrolling

`components/nav/ListeningViewsSidebar.vue`:

- Line 41: the sidebar root sets its width and collapse behaviour, with `desktop:overflow-hidden` when collapsed.
- Line 43: a fixed-height header, `h-[4.25rem] min-h-[4.25rem]`.

`components/nav/ListeningViewsListContent.vue`:

- Line 2: the list container is `flex min-w-[200px] flex-1 flex-col overflow-y-auto px-3 pt-3 space-y-2`.

The list itself has `overflow-y-auto` and `flex-1`, which is the right idea, but the reported symptom (topics cut off at the bottom and the whole sidebar scrolling) points at the ancestor chain not constraining height. A `flex-1` child with `overflow-y-auto` only scrolls internally if every ancestor between it and the viewport-height container has a definite height or `min-h-0`. Nothing in the sidebar root at line 41 establishes that, and flex children default to `min-height: auto`, which lets the list grow past its parent instead of scrolling inside it.

This is the well-known flexbox overflow trap, and the fix is almost always adding `min-h-0` at the right level rather than changing the list's own overflow.

## Part C: Blog and Website Mentions billing and production

Web mentions are **already built**, end to end:

- `contentstudio-frontend/src/modules/listening/constants.ts:124` — `{ id: 'web', label: 'Blogs & sites' }` is in `LISTENING_PLATFORM_OPTIONS`, with no feature gate, alongside Facebook, Instagram, Threads, Reddit, Twitter / X and TikTok.
- `contentstudio-frontend/src/modules/listening/utils/mentionContent.ts` handles the `web` platform specially at lines 219 and 257, since a web mention's author is a source domain rather than a social handle. There is a dedicated spec, `mentionContent.web.spec.ts`.
- `components/feed/ListeningMentionCard.vue:109` — engagement metrics are hidden for `web` mentions, because blogs and sites do not report them.
- `contentstudio-social-analytics-go/src/services/listening/fetcher/web.go` — the fetcher exists.

So the feature works. What is unfinished is the commercial side.

### Existing listening billing model

- `contentstudio-backend/app/Data/Listening/ListeningSubscriptionUsageData.php` carries `topics_count`, `topics_limit`, `mentions_this_month`, `mention_limit_monthly`, `approaching_limit`, `mentions_limit_reached`, `status` and `reset_date`.
- `contentstudio-backend/app/Libraries/Settings/SubscriptionLimits.php` counts `listening_topics` across all of a user's workspaces for the addon catalog, and sets `social_listening_lock` false plus `social_listening_addon` true when `has_social_listening_subscription` is set.
- `contentstudio-backend/app/Libraries/Account/IncreaseLimitsAddon.php`, `app/Services/PaddleBillingService.php`, `app/Services/Listening/TopicService.php` and `app/Http/Controllers/Listening/ListeningTopicsController.php` are the other billing-adjacent touchpoints.
- Per the feature docs, Listening is a $49/month add-on with 5 topics by default and 10,000 mentions per month.

### The billing questions to settle

- **Do web mentions count against the same monthly mention pool as social mentions, or their own?** Web crawling has a different cost profile from social API reads, and a blog-heavy topic could exhaust a shared 10,000 pool quickly.
- **Is Blogs and sites included in the base add-on or priced separately?** The website's Add Ons section lists Social listening as one add-on with no web-specific line, which suggests included, but that needs confirming.
- **What happens at the limit?** The existing model has `approaching_limit` and `mentions_limit_reached` flags, so the mechanism exists. Whether web mentions stop independently of social mentions depends on the pool decision.
- **Is there a per-topic or per-domain cap** for web sources, given a single broad keyword can match a very large number of blog posts?

## The concern with one ticket

Parts A and B are small frontend fixes with no commercial consequence. Part C is a billing implementation plus a production rollout, with pricing decisions, metering, limit enforcement and a launch checklist. Bundling them means the two small fixes cannot ship until the billing decisions are made, and it makes the ticket hard to estimate or hand to one person.

**Recommended split** (for the PO to accept or reject):

1. `[FE]` Add delete confirmation and fix the topic sidebar scroll containment
2. `[BE]` Meter and enforce Blog and Website Mentions against the listening subscription
3. `[FE]`+`[BE]` Blog and Website Mentions production readiness

The story below is authored as the single combined ticket that was requested. The split above is the alternative, not a substitute.

## Related existing work

- `docs/features/social-listening/` — the full feature docs, including the pricing strategy and the 18-platform scope. Its workflow doc section 6.5 covers Settings and Billing.
- `docs/stories/social-listening-module-launch-ui/` — the launch surface work: new badge, billing modal copy, announcement banner, login carousel, hover tooltips.
- `docs/stories/social-listening-data-banner/` — the data banner.

## Files involved

Frontend:
- `contentstudio-frontend/src/modules/listening/composables/useListeningTopics.ts`
- `contentstudio-frontend/src/modules/listening/components/topics/ListeningTopicListItem.vue`
- `contentstudio-frontend/src/modules/listening/components/nav/ListeningViewsListContent.vue`
- `contentstudio-frontend/src/modules/listening/components/nav/ListeningViewsSidebar.vue`
- `contentstudio-frontend/src/modules/listening/constants.ts`
- `contentstudio-frontend/src/locales/*/listening.json`

Backend:
- `contentstudio-backend/app/Data/Listening/{ListeningSubscriptionUsageData,ListeningTopicUsageDetailsData}.php`
- `contentstudio-backend/app/Libraries/Settings/SubscriptionLimits.php`
- `contentstudio-backend/app/Services/Listening/TopicService.php`
- `contentstudio-backend/app/Http/Controllers/Listening/ListeningTopicsController.php`
- `contentstudio-backend/app/Jobs/Listening/TriggerTopicSyncJob.php`

Go service:
- `contentstudio-social-analytics-go/src/services/listening/fetcher/web.go`
- `contentstudio-social-analytics-go/src/services/listening/` mentions and parser packages

## Mobile

None. Social Listening is web only in v1.
