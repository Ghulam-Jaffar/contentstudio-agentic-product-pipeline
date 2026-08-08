# Research — Extend the standardized post previews to Inbox and Analytics

Item 21 of the 7 Aug 2026 backlog batch.

## Current state

### The standard exists and works

`contentstudio-frontend/src/modules/common/components/social-previews/` is the shared preview system produced by **[FE] Rework post previews in Composer and Planner for platform fidelity and consistent sizing**:

- `PostPreview.vue` — the entry point
- `shared/PostChrome.vue`, `shared/CaptionBlock.vue`, `shared/MediaBlock.vue`, `shared/LinkCard.vue`, `shared/EngagementBar.vue` — the shared scaffold
- `SwipeCarousel.vue` — carousel handling
- `SocialPreviews/*.vue` — one per platform and post type: Facebook (plus Reel, Story, Background), Instagram (plus Reel, Multimedia, Multimedia Story), Twitter, LinkedIn, TikTok, Threads (plus Multi), Bluesky, Pinterest, YouTube, Tumblr, Telegram, GMB, and `NoSocialPreview.vue`
- `types.ts` for the shared shape

That story's acceptance criteria are explicit that no per-platform preview re-implements the chrome, caption, media, link card or engagement bar, and that the choice of preview variant for a given platform, post type and media shape lives in one shared helper so Composer and Planner cannot drift.

### Who uses it

Grepping for `social-previews` consumers returns `planner_v2` (`PlannerFeedCard.vue`, `PlannerPostPDF.vue`, `PlanPreviewPlatformSwitcher.vue`, `SocialMediaViewer/Instagram/MainScreen.vue`, `usePlannerFeedCard.ts`) and the composer path. That is it.

**Inbox and Analytics do not use it.** Both hand-roll their own.

### What Inbox renders today

`contentstudio-frontend/src/modules/inbox-revamp/components/PostView.vue` builds the post display inline with raw markup:

- Profile image with a fallback to a platform image
- Caption via `parseDescriptionHtml(...)` with a per-platform boolean for LinkedIn
- Attachments handled by branching on `postDetails.post_attachment.length`, with a no-preview-available placeholder image and tooltip
- Two separate inline carousels, one over `postAttachments.post_image_album` and one over `postAttachments['sub-attachments']`, both with element ids of `preview-carousel` and a class of `facebook-carousel-preview`
- Instagram gets its own layout ordering via `'order-last': post?.platform === 'instagram'`

Also relevant: `components/CommentBlock.vue` and `components/ChatView.vue`, which render the conversation side.

So Inbox has platform-specific branching inline, a Facebook-named carousel class used generally, and its own no-preview state. None of it shares the standard scaffold, so a platform's real-world appearance change has to be fixed here separately.

### What Analytics renders today

Several separate implementations:

- `components/common/AnalyticPreview.vue` — the general post preview modal. Branches heavily by platform and media type: an Instagram carousel loop keyed `insta-carousel_img_preview_*`, video detection by URL substring (`mediaUrl.includes('//video') || mediaUrl.includes('.mp4')`), `media_type` checks against `'VIDEO'`, `'REELS'`, `'IMAGE'`, its own caption truncation at 250 characters, its own default profile image, and its own resized-image URL handling with an onerror fallback
- `components/competitor/FacebookPublishedPostPreview.vue`
- `components/competitor/LinkedinPublishedPostPreview.vue`
- `components/competitor/InstagramPublishedPostPreview.vue`
- `components/competitor/PerformancePostPreviewModal.vue`
- `views/tiktok/components/TiktokPublishedPostPreview.vue`
- `views/twitter/components/TwitterPublishedPostPreview.vue`

That is at least six per-platform preview components plus a general one, all outside the shared system.

## The gap

The same post can appear in four places (Composer, Planner, Inbox, Analytics) and look different in the last two. Media handling, caption truncation, link cards, engagement display and the no-preview state are each implemented separately, and video detection in Analytics is done by looking for substrings in a URL. A platform's UI refresh now has to be applied in three places instead of one.

## The complication worth naming

Composer and Planner preview a post **before** it is published, from data the user is composing. Inbox and Analytics show a post **after** it is published, from platform API data. The shapes differ:

- Analytics has real engagement numbers, which the shared `EngagementBar` currently renders as static chrome
- Analytics has platform media types (`VIDEO`, `REELS`, `IMAGE`, `CAROUSEL_ALBUM`) rather than composer media objects
- Inbox has attachment structures per platform (`post_attachment`, `post_image_album`, `sub-attachments`) that came from the platform
- Neither has the composer's per-account overrides

So this is not a drop-in reuse. The shared components need an adapter per surface that maps published-post data into the shared shape, plus the ability to render real engagement counts rather than placeholder chrome. That is the substance of the work.

## What needs to change

- An adapter layer mapping published-post data (Inbox attachment structures, Analytics platform media types) into the shared preview shape.
- `EngagementBar` able to render real counts, so Analytics does not need its own.
- Inbox's `PostView.vue` rebuilt on the shared components, retaining its own conversation-side chrome.
- Analytics' preview components replaced with the shared ones, behind whichever modal or card wrapper each surface already uses.
- One no-preview state, replacing Inbox's placeholder image and Analytics' fallbacks.
- Video detection by media type rather than by URL substring.

## Scope check on networks

The shared system covers Facebook, Instagram, Twitter, LinkedIn, TikTok, Threads, Bluesky, Pinterest, YouTube, Tumblr, Telegram and GMB. Inbox supports Facebook, Instagram, LinkedIn, YouTube, GMB and X. Analytics supports Facebook, Instagram, LinkedIn, X, YouTube, TikTok, Pinterest and GMB, plus Facebook and Instagram competitor views. Every network either surface needs is already covered by the shared system, so no new platform preview components are required.

## Files involved

Shared:
- `contentstudio-frontend/src/modules/common/components/social-previews/` (all of it, plus `types.ts`)

Inbox:
- `contentstudio-frontend/src/modules/inbox-revamp/components/PostView.vue`
- `contentstudio-frontend/src/modules/inbox-revamp/components/{CommentBlock,ChatView}.vue`

Analytics:
- `contentstudio-frontend/src/modules/analytics/components/common/AnalyticPreview.vue`
- `contentstudio-frontend/src/modules/analytics/components/competitor/{Facebook,Linkedin,Instagram}PublishedPostPreview.vue`
- `contentstudio-frontend/src/modules/analytics/components/competitor/PerformancePostPreviewModal.vue`
- `contentstudio-frontend/src/modules/analytics/views/tiktok/components/TiktokPublishedPostPreview.vue`
- `contentstudio-frontend/src/modules/analytics/views/twitter/components/TwitterPublishedPostPreview.vue`

Note the repo's import rule: modules may import global, and no new module-to-module imports. `social-previews` sits in `modules/common`, which `planner_v2` already imports from, so the precedent exists.

## Mobile

None. These are web surfaces. The mobile apps have their own inbox and analytics screens with their own rendering.
