# Stories — Post preview platform fidelity & responsive sizing (Composer + Planner)

**Pipeline:** local `/story` (no Shortcut push)
**Platform:** Web (frontend only)
**Type:** Visual fidelity + responsive consistency pass

> **Design dependency — read first.** This story is the engineering scope and the cross-platform consistency pattern. The exact per-platform visual specs (header layouts, fonts, spacing, icon sets, link-card styling, engagement-bar treatments) must be **provided by a product designer** before implementation begins. The AC below is framed around observable parity with the designer's spec, not pre-described pixel values.

| # | Story |
|---|---|
| 1 | [FE] Rework post previews in Composer and Planner for platform fidelity and consistent sizing |

---

## Story 1 — [FE] Rework post previews in Composer and Planner for platform fidelity and consistent sizing

### Description
As a ContentStudio user composing a post in Composer or reviewing a scheduled post in Planner, I want each social-platform preview to look like the actual network it represents and to look right in both views and at every viewport size, so that I can trust what I see in the preview before I publish — without second-guessing whether the real Facebook / Instagram / X / LinkedIn / Threads / TikTok / YouTube / Pinterest / GMB / Bluesky / Tumblr / Telegram post will look the same.

The fidelity gap today comes from each platform's preview component being hand-built once, then drifting from the real platform UI as those platforms refreshed. There is also no shared visual scaffold across the preview set — so spacing, avatar sizes, fonts, engagement bars, and link cards all look inconsistent across platforms even within the same Composer or Planner screen. And the previews don't scale gracefully between the narrow Composer sidebar and the wider Planner modal.

This story refreshes every platform preview against an authoritative design spec (provided by the product designer), introduces a small shared-scaffold layer so future updates stay consistent, and fixes responsive sizing so the same components look correct in both the Composer pane and the Planner detail modal at every viewport size.

### Workflow
1. The user opens the **Composer**. As they pick social accounts and edit their post, the right-hand preview pane renders one preview card per account-family (Facebook, Instagram, X, LinkedIn, Threads, etc.) — each looking like the actual network's post chrome.
2. The user toggles the Composer preview between **web** and **mobile** layouts using the existing toggle. Each platform preview switches to its mobile presentation and remains accurate to that platform.
3. The user switches the post type (e.g. Facebook → Reel; Instagram → Story; Threads → Multi). The right variant is rendered automatically — no stale variant left on screen.
4. The user resizes the browser window or opens Composer on a smaller laptop. The previews stay readable and proportional — no clipping, no overflowing engagement bars, no truncated avatars.
5. The user opens **Planner**, finds a scheduled post, and opens the post-detail modal. The same preview components render in the wider modal column. Spacing, fonts, media aspect ratios, and engagement bars look correct for the wider container — not an upscaled or stretched version of the narrow Composer layout.
6. The user opens a shared / publicly-accessible plan link. The previews still look correct in that container.
7. Across both Composer and Planner, all platforms (Facebook feed, Background, Reel, Story; Instagram feed, Multimedia, MultimediaStory, Reel; Threads feed, Link, Multi, Multimedia; X, LinkedIn, Bluesky, TikTok, YouTube, Pinterest, GMB, Tumblr, Telegram) and the empty-state (no-social-selected) render with the same visual rhythm — avatar size, header spacing, divider treatment, engagement-bar height, link-card aspect — except where the platform itself differs.

### Acceptance criteria

**Platform fidelity (each preview matches the designer's spec for the real platform):**
- [ ] **Facebook** feed-post, Background post, Reel, and Story previews each match the designer's reference for facebook.com / Facebook mobile as of the spec date.
- [ ] **Instagram** feed-post, Multimedia (carousel), MultimediaStory, and Reel previews each match the designer's reference for instagram.com / Instagram mobile.
- [ ] **Threads** feed-post, Link, Multi, and Multimedia previews each match the designer's reference for threads.net.
- [ ] **X / Twitter** preview matches the designer's reference for x.com.
- [ ] **LinkedIn** preview matches the designer's reference for linkedin.com.
- [ ] **Bluesky** preview matches the designer's reference for bsky.app.
- [ ] **TikTok** preview matches the designer's reference for tiktok.com (full-bleed vertical video, right-side action stack, caption overlay).
- [ ] **YouTube** preview matches the designer's reference (video thumbnail, title, channel row, description).
- [ ] **Pinterest** preview matches the designer's reference for pinterest.com (tall pin, title, board row).
- [ ] **GMB (Google Business Profile)** preview matches the designer's reference for the post types Update / Event / Offer.
- [ ] **Tumblr** preview matches the designer's reference for tumblr.com.
- [ ] **Telegram** preview matches the designer's reference for the Telegram channel post UI.
- [ ] Every preview is reviewed and approved by the designer against their own spec before sign-off.

**Shared visual scaffold:**
- [ ] A shared set of internal primitives is introduced and used by every per-platform preview component:
  - A post-chrome scaffold component covering the platform header (avatar, name slot, supporting-line slot, timestamp slot, verified-badge slot, more-menu slot) and the engagement-bar slot at the bottom.
  - A caption block covering text + truncation + "see more" + mention/hashtag/URL highlighting.
  - A media block covering single image, image grid, carousel, and video with platform-specific aspect ratios.
  - A link card component used by every platform that renders an OG card, accepting platform-specific styling via props.
  - An engagement bar component accepting per-platform icon sets and counts/labels.
- [ ] No per-platform preview component re-implements the chrome, caption, media, link card, or engagement bar from scratch — each composes the shared primitives.
- [ ] Each primitive renders with semantic tokens / CSS variables from the existing design system — no hardcoded hex colors, no hardcoded font sizes, no magic px values. Per project rule: use `text-primary-cs-500`, `bg-primary-cs-50`, etc., never raw colors like `text-blue-600`.

**Responsive sizing:**
- [ ] Every preview component renders correctly at four reference container widths:
  1. **Composer sidebar — desktop** (the post-preview pane in Composer v2 at typical desktop widths)
  2. **Composer sidebar — mobile toggle** (when the existing web/mobile toggle is set to "mobile")
  3. **Planner post-detail modal** (wider column than the Composer sidebar)
  4. **Shared plan public view** (whatever the existing share-plan link renders in)
- [ ] Sizing inside preview components uses relative units (rem / em / `%` / container-relative) or the existing breakpoint utilities. No fixed-px widths inside the preview components.
- [ ] At intermediate viewport widths (small laptops, narrow browser windows) every preview remains readable — nothing clips, nothing overflows the container, no avatars shrink to unreadable sizes, no engagement bar wraps awkwardly.
- [ ] The existing Composer web/mobile preview toggle still works for every platform — toggling switches each platform's chrome to its mobile presentation.

**Variant routing:**
- [ ] The decision of which preview variant to render for a given `(platform, postType, mediaShape)` lives in a single shared helper used by both Composer and Planner — Composer and Planner cannot drift on which variant they pick.
- [ ] Switching post type in Composer (e.g. enabling Facebook Reel, Instagram Story, Threads Link) immediately switches the rendered preview to the correct variant. No stale variant left on screen.
- [ ] A regression covering each variant route is documented (test plan or QA script) so future changes can be verified quickly.

**Behavior preserved (no functional regressions):**
- [ ] Selecting / deselecting social accounts in Composer adds / removes the corresponding preview cards — same behavior as today.
- [ ] First-comment preview, location chip, mention rendering, hashtag highlighting, link-preview / OG card, video thumbnail playback, carousel arrows, story progress bars, Reel sound badge — every existing affordance still works.
- [ ] The Composer "info / disclaimer" banner (`composer.post_preview.disclaimer`) still appears in the same conditions.
- [ ] The Composer empty state (`NoSocialPreview`) still appears when no social account is selected, and is also refreshed against the designer's spec.
- [ ] The Planner detail modal continues to show the same surrounding UI (comments / notes, approval status, approval workflow badge, label / campaign attachments, action buttons) — only the per-platform preview block changes visually.

**Internationalization / theming / copy:**
- [ ] All user-facing text continues to come from existing locale files. Any new copy introduced (e.g. a new "Sponsored" / "Promoted" label if a platform adds one in the spec) is added to **every** locale directory under `src/locales/` with English source + per-locale fallbacks per the existing localization policy. No hardcoded English strings.
- [ ] White-label workspaces render previews using the same per-platform real-world colors / iconography — preview chrome reflects the social network's brand, not the workspace's white-label theme. (Existing behavior — preserved.)

**No mobile apps in scope:**
- [ ] No changes to the iOS app (`contentstudio-ios-v2/`) post-preview surface. No changes to the Android app (`contentstudio-android-v2/`).

### Mock-ups
**TBD — Product designer to supply.** Implementation cannot start until the designer has delivered per-platform reference specs (mockups, Figma frames, or screenshots from the real platforms annotated with our chrome) for every platform variant listed in the AC.

### Impact on existing data
None. No schema changes, no API changes, no migration. Pure visual refactor of UI components that consume already-existing post data.

### Impact on other products
- **Backend / API:** no changes.
- **iOS:** no changes (iOS post-preview is a separate codebase — see the Q2 2026 iOS improvements stories).
- **Android:** no changes (the new Android post-preview is being built separately under the Q2 2026 Android improvements stories).
- **Chrome extension:** no changes.
- **White-label:** no behavioral changes; previews continue to render with each platform's own brand chrome, not the white-label theme.

### Dependencies
- **Designer-provided per-platform specs** — engineering blocked on receiving these for every active platform variant before implementation.
- Existing `@contentstudio/ui` icon set — used where icons map. Where a platform-native glyph is needed (e.g. X's repost icon, TikTok's right-side action stack), a local SVG asset is acceptable.
- No story dependencies on other in-flight stories.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories) — verify every preview at narrow Composer sidebar, mobile-toggle, and Planner modal widths.
- [ ] Multilingual support (frontend + backend, translations available or fallback handled) — verify every locale renders correctly with the refreshed previews; any new strings added to all locales.
- [ ] UI theming support (default + white-label, design library components are being used) — previews continue to render social-platform brand chrome (not white-label theme); design library components used for shared scaffolding.
- [ ] White-label domains impact review — verified: previews show the social network's chrome, not the white-label theme; no behavior change for white-label workspaces.
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension) — verified none — see Impact on other products above.

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Per-platform preview components (refactor — same files):**
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/FacebookPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/FacebookBackgroundPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/FacebookReelPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/FacebookStoryPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/InstagramPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/InstagramMultimediaPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/InstagramMultimediaStoryPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/InstagramReelPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/TwitterPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/LinkedinPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/ThreadsPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/ThreadsLinkPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/ThreadsMultiPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/ThreadsMultimediaPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/BlueskyPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/TikTokPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/YoutubePreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/PinterestPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/GmbPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/TumblrPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/TelegramPreview.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/NoSocialPreview.vue`

**Hosts (light edits — container responsiveness, variant routing):**
- `contentstudio-frontend/src/modules/composer_v2/components/PostPreview.vue` (589 lines — Composer sidebar)
- `contentstudio-frontend/src/modules/planner_v2/components/PlannerPostPreview_v2.vue` (2505 lines — Planner detail modal; already imports the same SocialPreviews components — see imports near line 1187)
- `contentstudio-frontend/src/modules/planner_v2/composables/usePostPreview.js`

**Suggested new shared primitives (engineering may name / split differently):**
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/_shared/PostChrome.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/_shared/CaptionBlock.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/_shared/MediaBlock.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/_shared/LinkCard.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/_shared/EngagementBar.vue`
- `contentstudio-frontend/src/modules/composer_v2/composables/usePreviewVariant.ts` (centralized variant routing)

**Gotchas:**
- The Planner host (`PlannerPostPreview_v2.vue`) is 2505 lines and does a lot beyond just rendering previews (comments, approval workflow, plan status, label / campaign attachments). Touch only the parts that pass props into the SocialPreviews components — leave the rest alone in this story.
- The Composer host has a single `togglePreview` boolean controlling mobile vs web. Keep that contract — don't break the toggle UI / store wiring. Move responsive behavior *inside* the platform components.
- Facebook carousel currently routes through a separate `carousel` / `carouselAccount` prop only on `FacebookPreview`. Other platforms fold carousel handling into their Multimedia variants. The new `MediaBlock` should normalize this so any platform can render a carousel without a side channel of props on the host.
- Some preview components only render correctly with specific post-type combinations (e.g. an Instagram post with no media must not render the Reel variant). Centralizing variant routing in `usePreviewVariant` fixes the silent-wrong-variant class of bug.
- Legacy parallel preview set at `contentstudio-frontend/src/modules/publish/components/posting/social/previews/` has zero external consumers (per grep) — out of scope here, flag as a follow-up cleanup story.
- Blog post-type previews are out of scope (blog publishing has been sunset).
- `@contentstudio/ui` icons cover most generic icons but not platform-native glyphs (X's repost, TikTok's right-rail icons, Threads' specific reaction). Local SVG assets are acceptable for those; group them under `contentstudio-frontend/src/assets/img/social-previews/<platform>/` for findability.
- All shared primitives and refactored previews should use `<script setup lang="ts">` per `contentstudio-frontend/CLAUDE.md`.
- Any new user-facing strings must be added to every locale directory under `src/locales/` — though most copy already exists today and should be reused.
