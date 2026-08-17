# Research: Flutter Composer Parity with the Web Composer

**Supersedes:** `docs/stories/mobile-composer-posting-options-parity/` (written against the retired native iOS and Android apps, with stale web paths).

**Ask:** one single generic `[Flutter]` story a dev can pick up, covering composer parity overall rather than one story per platform.

---

## Current State

### The Flutter composer is already built, and already at iOS parity

`contentstudio-flutter/lib/features/composer/` is 69 Dart files across `domain/`, `data/`, `application/`, `presentation/`. Its own feature doc (`contentstudio-flutter/docs/features/composer.md`) reports **Phases 0 to 11 complete**, ~160 composer tests green, and "**~Full functional parity** with the iOS composer for the supported scope."

This changes the shape of the work fundamentally versus the old research. The old story assumed the mobile composer was a stub (iOS `PostSettingVC.swift` was a 40-line placeholder). It is not. **The remaining gap is not iOS parity, it is web parity**, because the Flutter app was ported from iOS, and iOS itself never reached web parity.

Key files:
- `lib/features/composer/domain/model/post_settings.dart` (per-platform settings model, 380 lines)
- `lib/features/composer/domain/settings/post_settings_rules.dart` (which options apply per post type / media profile)
- `lib/features/composer/presentation/widgets/composer_settings_sheet.dart` (per-platform collapsible sections)
- `lib/features/composer/domain/payload/publish_payload_builder.dart` (byte-match golden fixtures against the real payload)
- `lib/features/composer/domain/platform/composer_platform.dart` (supported network list, char limits)

### The web composer moved too

The web composer was refactored since the old research: options now live in `contentstudio-frontend/src/modules/composer/features/platform-settings/ChannelOptions/` (the old `composer_v2/components/ChannelOptions/` path in the superseded research no longer exists). Validation is per-platform under `features/validation/rules/`.

---

## What Needs to Change (the web parity gap)

### 1. Two networks are missing entirely

| | Networks |
|---|---|
| Web (`shared/constants/platforms.ts`) | facebook, instagram, threads, twitter/X, linkedin, pinterest, gmb/GBP, youtube, tumblr, tiktok, **bluesky**, **telegram** (12) |
| Flutter (`composer_platform.dart` `kComposerPlatforms`) | facebook, instagram, threads, twitter/X, linkedin, pinterest, gmb, youtube, tiktok, tumblr (10) |

Bluesky and Telegram can be selected and published on web but do not exist in the Flutter composer at all. Web has `EditorBlueskyBox.vue`, `TelegramOptions.vue`, and `validation/rules/telegram.ts` (album max 10 items).

### 2. TikTok is missing its compliance fields

This is the highest-risk gap, because TikTok requires these for direct posting. Nothing matching `brand`, `aigc`, or `commercial` exists anywhere in the Flutter composer (verified by grep across `lib/features/composer/`).

`TikTokSettings` has: `postType`, `privacyLevel`, `allowComments`, `allowDuet`, `allowStitch`, `publishingMethod`, `autoAddMusic`.

Missing versus web (`ChannelOptions/Tiktok/useTiktokOptions.ts`):
- Disclose branded content (`disclose_commercial_content`)
- Your brand (`brand_organic_toggle`)
- Third party brand (`brand_content_toggle`)
- AI-generated content (`is_aigc`)
- Carousel title (max 90 chars, carousel only)
- Post type is in the model but `copyWith` does not accept it, so video vs carousel is not user-selectable

### 3. Model fields exist but are not editable

Several settings hydrate correctly on edit (`data/mappers/plan_hydrator.dart` reads them) and serialize correctly, but there is no UI to set them, so a post created on mobile can never carry them. These are cheap wins:

| Platform | Field | State |
|---|---|---|
| Instagram | `collaborators` | hydrated from `instagram_collaborators`, absent from `copyWith`, no UI |
| Instagram | `postingOption` (API vs mobile device) | hydrated from `instagram_posting_option`, absent from `copyWith`, no UI |
| Facebook | video title | web has it (max 100 chars on video posts), no field on `FacebookSettings` |
| Facebook | `backgroundId`, `postedAs` | on the model, absent from `copyWith` |
| X/Twitter | `postType` | on the model, absent from `copyWith` |
| YouTube | `playlist` | round-trips on edit, no picker (needs a per-account playlist fetch) |

### 4. LinkedIn has no poll

Web has `components/LinkedinPollModal.vue` (question, up to 4 options, duration) and `LinkedInCarouselBox.vue`. `LinkedInSettings` covers the document carousel (`isCarousel`, `documentAdded`, `document`) but has no poll at all.

### 5. Threads has a model but no UI

`ThreadsSettings` supports `hasMultiThreads` + `threadSegments` and the payload builder serializes `multi_threads`, but the settings sheet has no Threads section. It was deliberately removed during the iOS port because iOS has none. Web supports it.

### 6. Settings sheet section coverage

`composer_settings_sheet.dart` `_sectionFor` has sections for facebook, instagram, twitter, youtube, tiktok, gmb, linkedin. Pinterest is handled inline on the canvas (`composer_pinterest_fields.dart`). Threads and Tumblr have no section.

### 7. Carousel builder UI is deferred

`CarouselSettings` and the payload support carousels; the dedicated builder UI is explicitly deferred to a standalone epic in the Flutter composer doc. Web has it. **Recommend keeping this out of the parity story** and letting it stay its own epic, otherwise this story has no end.

### 8. Validation depth

Flutter's own doc flags these as known follow-ups: per-post-type Instagram rules (feed = 1 media, reel = 1 video), YouTube Shorts (60s max plus aspect ratio), and GMB required event/offer fields still need per-`PostSettings` validation threaded into the validators. Web enforces them via `features/validation/rules/`.

### 9. Custom video thumbnail

Web has `components/CustomThumbnail/CustomThumbnailModal.vue`. No thumbnail selection exists in the Flutter composer (only thumbnails read back from hydrated plans).

---

## Explicitly Out of Scope

- **Write with AI / AI caption assist.** Deliberately excluded from the Flutter composer (stubbed and hidden), and consistent with the product rule that AI *generation* is web-only. The AI assistant lives in `lib/features/ai_assistant/` and is a separate surface.
- **Carousel builder UI.** Already carved out as its own epic on the Flutter side.
- **Instagram manual publish coordinator.** Belongs to the Instagram publish epic, not the composer.
- **Backend work.** Mobile posts through the same `POST /processSocialShare` endpoint as web and the API already accepts every field above. The old research confirmed this and nothing suggests it changed. **No `[BE]` story needed.**

---

## Story Shape

One `[Flutter]` story, as requested, framed as a **parity audit plus close the gaps**, with the gap list as acceptance criteria grouped by network. That works as a single ticket because:
- It is one codebase, one screen, one settings sheet, and one payload builder.
- The dev doing it will be in the same three files for nearly every item.
- The backend needs nothing, so there is no cross-team sequencing.

Risk to flag to the PO: this is a **large** single story. If the dev wants it split during sprint planning, the natural seams are (a) the two missing networks, (b) TikTok compliance fields, (c) the not-editable model fields, (d) validation depth. Called out in the story's Dependencies section so planning has the option.

---

## Files Involved

**Flutter (the work):**
- `lib/features/composer/domain/model/post_settings.dart`
- `lib/features/composer/domain/settings/post_settings_rules.dart`
- `lib/features/composer/domain/platform/composer_platform.dart`
- `lib/features/composer/domain/payload/publish_payload_builder.dart`
- `lib/features/composer/domain/validators/platform_validator.dart`
- `lib/features/composer/presentation/widgets/composer_settings_sheet.dart`
- `lib/features/composer/data/mappers/plan_hydrator.dart`
- `assets/i18n/*.json` (all 8 locales, parity test enforces it)

**Web (reference only, no changes):**
- `src/modules/composer/shared/constants/platforms.ts`
- `src/modules/composer/features/platform-settings/ChannelOptions/` (Tiktok, Facebook, Instagram, Youtube, Gmb, Telegram)
- `src/modules/composer/components/LinkedinPollModal.vue`
- `src/modules/composer/features/validation/rules/`

**Flutter conventions the dev must follow** (`contentstudio-flutter/CLAUDE.md`):
- Riverpod 2, feature-first layering, `domain/` stays pure Dart
- Every user-facing string localized into all 8 locale files (`translations_parity_test` gates it)
- Payload builders and validators get tests before UI work
- `CSColors` / `CSType` / `CSFonts`, reuse shared widgets, `docs/ui-consistency.md` is the UI guide
- Branch from `develop-cs` as `feat/<task>`
