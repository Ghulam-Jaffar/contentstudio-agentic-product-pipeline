# Research — Post preview platform fidelity & responsive sizing (Composer + Planner)

**Date:** 2026-05-18
**Pipeline:** `/story` (local docs only — not pushed to Shortcut)
**Type:** Frontend visual refactor — fidelity + responsive layout
**Surfaces:** Composer (post preview pane) and Planner (post-detail / preview modal) in `contentstudio-frontend`

---

## Current State

Per-platform post previews are rendered by a single shared component set under [composer_v2/components/SocialPreviews/](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/). Both the Composer's preview pane and the Planner's post-detail modal consume these same components — so visual issues in any one component show up in both places.

### Canonical preview component set — `composer_v2/components/SocialPreviews/`

| File | Lines |
|---|---|
| [FacebookPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/FacebookPreview.vue) | 727 |
| [FacebookBackgroundPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/FacebookBackgroundPreview.vue) | 81 |
| [FacebookReelPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/FacebookReelPreview.vue) | 145 |
| [FacebookStoryPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/FacebookStoryPreview.vue) | 316 |
| [InstagramPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/InstagramPreview.vue) | 339 |
| [InstagramMultimediaPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/InstagramMultimediaPreview.vue) | 388 |
| [InstagramMultimediaStoryPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/InstagramMultimediaStoryPreview.vue) | 552 |
| [InstagramReelPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/InstagramReelPreview.vue) | 278 |
| [TwitterPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/TwitterPreview.vue) | 552 |
| [LinkedinPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/LinkedinPreview.vue) | 465 |
| [ThreadsPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/ThreadsPreview.vue) | 187 |
| [ThreadsLinkPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/ThreadsLinkPreview.vue) | 58 |
| [ThreadsMultiPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/ThreadsMultiPreview.vue) | 168 |
| [ThreadsMultimediaPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/ThreadsMultimediaPreview.vue) | 125 |
| [BlueskyPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/BlueskyPreview.vue) | 471 |
| [TikTokPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/TikTokPreview.vue) | 349 |
| [YoutubePreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/YoutubePreview.vue) | 338 |
| [PinterestPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/PinterestPreview.vue) | 216 |
| [GmbPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/GmbPreview.vue) | 298 |
| [TumblrPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/TumblrPreview.vue) | 182 |
| [TelegramPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/TelegramPreview.vue) | 328 |
| [NoSocialPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/NoSocialPreview.vue) | 198 |

22 files, ~6800 lines of Vue templates + scoped styles. Each one re-implements the platform's post chrome (header, action row, timestamps, link card, etc.) from scratch — no shared "post-chrome scaffold" / "media block" / "caption block" abstraction across them.

### Hosts

- **Composer side:** [composer_v2/components/PostPreview.vue](contentstudio-frontend/src/modules/composer_v2/components/PostPreview.vue) (589 lines) — sidebar pane with a `togglePreview` switch between **web** and **mobile** layouts (`mobile-preview w-11/12 absolute` class). Width is dictated by the composer's right-sidebar column.
- **Planner side:** [planner_v2/components/PlannerPostPreview_v2.vue](contentstudio-frontend/src/modules/planner_v2/components/PlannerPostPreview_v2.vue) (**2505 lines** — monolith) — imports the same `composer_v2/components/SocialPreviews/*Preview.vue` files (see imports at lines 1187–1197). Width is dictated by the planner detail-modal column, which is different from the composer sidebar.
- Composable: [planner_v2/composables/usePostPreview.js](contentstudio-frontend/src/modules/planner_v2/composables/usePostPreview.js)

### Legacy parallel preview set (still in repo, mostly unused)

[publish/components/posting/social/previews/](contentstudio-frontend/src/modules/publish/components/posting/social/previews/) — 12 older components (`FacebookPreview`, `FacebookCarouselPreview`, `InstagramPreview`, `LinkedinPreview`, `TwitterPreview`, `TwitterThreadedTweetsPreview`, `TumblrPreview`, `PinterestPreview`, `GmbPreview`, `TiktokPreview`, `YoutubePreview`, `SocialPreview`). Grep shows the only consumers are inside this same folder (`SocialPreview.vue` referencing `TwitterPreview.vue`) — no other modules import them. Effectively dead code candidates; cleanup is out of scope for this story but worth flagging.

---

## What's "out of shape" (problems observed)

This is the user's complaint, not strict findings — engineering + design will confirm during planning. Most likely issues, grounded in the code:

1. **Fidelity drift from real platforms.** Each `*Preview.vue` re-implements platform chrome by hand and has not been kept in sync as Facebook / Instagram / X / LinkedIn / TikTok have refreshed their UIs. Engagement-bar icons, font weights, header layouts, line heights, link-card aspect ratios, and reaction emojis are stale on multiple platforms.
2. **Inconsistent chrome across platforms.** Because there's no shared scaffold (no `<PostChrome />`, no `<MediaBlock />`, no `<CaptionBlock />`, no `<EngagementBar />` primitive), each platform's avatar size, header spacing, divider treatment, and timestamp format is one-off. Visual rhythm across platforms looks uneven inside the same screen.
3. **Different sizing in Composer vs Planner.** The Composer sidebar gives previews a narrow column (~360–420 px depending on viewport); the Planner modal renders them in a wider column (~520–640 px). The previews don't scale gracefully — fonts/spacing/media block proportions look correct in one container and broken in the other.
4. **Responsive breakpoints.** Composer's `togglePreview` flips between "web" (desktop) and "mobile" layouts via a single boolean (`mobile-preview w-11/12 absolute`). The actual breakpoints used inside each `*Preview.vue` (`@media` rules in `<style scoped>`) are inconsistent — some use fixed px widths, some use rem, some don't react at all. Result: at intermediate viewport sizes (small laptops, narrow browser windows), previews clip / overflow / stack awkwardly.
5. **Story sub-types missing/inconsistent.** Facebook has separate `Background`, `Reel`, `Story` variants; Instagram has separate `Multimedia`, `MultimediaStory`, `Reel`; Threads has `Link`, `Multi`, `Multimedia`. The selection logic that routes a single post to the right variant lives in the host (`PostPreview.vue` / `PlannerPostPreview_v2.vue`) — easy to forget a state. Some posts render in the wrong variant (e.g. a Reel landing in the feed-post layout).
6. **Carousel handling.** `carousel` and `carouselAccount` are passed only to `FacebookPreview` (line 64 of `PostPreview.vue`). Carousel rendering in Instagram and LinkedIn is folded into the Multimedia variants instead — inconsistent abstraction.
7. **Per-platform caption truncation.** Different `*Preview.vue` files cap caption length / "see more" expansion differently. Some show full text, some truncate at the wrong count.
8. **Link preview / OG card.** Card layout (image position, title weight, domain row) differs across platforms — not always matching what the real network renders.

---

## What Needs to Change

The story should describe **a visual fidelity + responsive consistency pass** across all 22 `*Preview.vue` files, **driven by design specs**. The user explicitly asked to **consult with a product designer for the per-platform specs** — so the story body is the *what* and *where*; the designer owns the *exact pixels per platform*.

Concrete changes:

- **Audit + spec collection (design-led):** Designer provides reference specs / mockups for every active platform variant (Facebook feed / Reel / Story / Background, Instagram feed / Reel / Story / multimedia, Threads feed / Link / Multi / Multimedia, X, LinkedIn, Bluesky, TikTok, YouTube, Pinterest, GMB, Tumblr, Telegram). Specs cover: header layout, avatar size, font sizes, line heights, spacing, media aspect ratios, engagement bar, link card, caption truncation, color tokens.
- **Shared scaffold primitives (FE):** Introduce small shared building blocks under `composer_v2/components/SocialPreviews/_shared/`:
  - `PostChrome.vue` — wraps the platform header (avatar, name, timestamp, verified badge slot, more-menu slot) and engagement bar slots
  - `CaptionBlock.vue` — text + truncation + mention/hashtag/link highlighting (already partially exists scattered)
  - `MediaBlock.vue` — image grid / single image / video / carousel handling with platform-specific aspect ratios
  - `LinkCard.vue` — OG card with platform-specific styling slots
  - `EngagementBar.vue` — like / comment / share / save with per-platform icon set
  Each platform component composes these instead of re-implementing the chrome.
- **Responsive sizing strategy (FE):** Replace the `togglePreview` boolean with container-query / breakpoint-driven sizing inside each platform component so they look correct at any of the four practical widths: **Composer sidebar (narrow)**, **Composer sidebar mobile-toggle**, **Planner modal (wider)**, and the **share-plan public view** (if applicable). All sizing in rem/em or container-relative units — no fixed px widths inside preview components.
- **Per-platform fidelity refresh (FE):** Update each platform's chrome, fonts, icons, spacing, and link-card to match the current real-world platform UI as of the design spec. Use `@contentstudio/ui` icons where they map, fall back to local SVGs for platform-native glyphs.
- **Variant routing fix (FE):** Centralize the "which `*Preview` variant to render" logic in a single helper (e.g. `composer_v2/composables/usePreviewVariant.ts`) so Composer and Planner cannot drift. Pass that helper a `(platform, postType, mediaShape)` and get back the component to render.
- **Cleanup (FE, opportunistic):** Delete the unused `publish/components/posting/social/previews/` set if engineering confirms zero callers — or leave with a TODO. Out of scope to require.

---

## UX Reference

Visual reference is the real platforms themselves (Facebook, Instagram, X, LinkedIn, Threads, Bluesky, TikTok, YouTube, Pinterest, GMB, Tumblr, Telegram) as of the design spec date. No external UX research needed — these are widely-known platform UIs. Designer's mockups are the authoritative spec.

---

## Mobile Context

Not applicable. This story is **web frontend only**. The native iOS post preview (`PostPreviewView.swift`, 3334 lines) and the planned Android post preview (see [Q2 2026: Android improvements](docs/stories/q2-2026-android-improvements/02-stories.md)) are separate codebases. If the design refresh produces specs that should also apply on mobile, that's a follow-up — not part of this story.

---

## Files Involved

**Per-platform preview components (refactor):**
- All 22 files under `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/`

**Hosts (light edits — variant routing + responsive container handling):**
- `contentstudio-frontend/src/modules/composer_v2/components/PostPreview.vue`
- `contentstudio-frontend/src/modules/planner_v2/components/PlannerPostPreview_v2.vue`
- `contentstudio-frontend/src/modules/planner_v2/composables/usePostPreview.js`

**New shared primitives (introduced by this story):**
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/_shared/PostChrome.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/_shared/CaptionBlock.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/_shared/MediaBlock.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/_shared/LinkCard.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/_shared/EngagementBar.vue`
- `contentstudio-frontend/src/modules/composer_v2/composables/usePreviewVariant.ts` (or `.js`)

**Locales (may need new keys for any new labels — most copy already exists):**
- `contentstudio-frontend/src/locales/*` — only if the redesign introduces new strings; mostly N/A.

**Legacy folder (out of scope — flagged for follow-up):**
- `contentstudio-frontend/src/modules/publish/components/posting/social/previews/` (appears unused outside its own folder)

---

## Out of Scope

- Mobile apps (iOS / Android) — separate codebases, separate stories
- Cleanup / deletion of the legacy `publish/.../previews/` folder (flagged here, follow-up story)
- Changes to how posts are composed, scheduled, or published — preview only
- Backend / API changes — none
- New platform integrations — only the platforms ContentStudio already supports
- Dark mode — not supported in ContentStudio
- RTL layout — not supported in ContentStudio
- Blog post-type previews — blog publishing has been sunset
