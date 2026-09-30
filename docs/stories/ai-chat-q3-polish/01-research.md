# Research: AI chat polish (Q3 AI updates)

**Source:** Team meeting, 2026-09-30. Target epic: **Q3 - 2026: AI updates** (`6961f057-9483-4b7e-b5fb-db225bc6850f`). That epic is currently marked **completed** in Helpin with a 2026-09-30 deadline; the PO chose to file into it anyway. Pushing new stories into it may need the epic reopened. Confirm at push time.

Four asks, plus one `[Design]` story covering the visual ones:

1. Live elapsed time on the running AI step
2. Transparent logos, placed with enough contrast (the PO merged "remove background" and "contrast-aware placement" into one story)
3. Premium typing and streaming feel
4. "What's new" popup in AI chat with Try now

## 1. Live step timer

- Timeline: `contentstudio-frontend/src/modules/AI-tools/components/timeline/AiBlockTimeline.vue` (header: running step title / "Working" / "Worked for {duration}", lines ~355-368; header duration = `formatMs(props.ms)`, the whole-turn ms).
- Step row: `timeline/AiStepRow.vue`. `duration = formatMs(props.block.ms)` (line 42), shows "running" while there is no `ms` (line 85).
- `formatMs` in `src/utils/blockStatusVisual.ts:~97`: < 100 ms shows nothing, < 60 s `x.xs`, else `Xm Ys`.
- Duration is **never measured on the client**. `block.ms` comes from the server on `block.end`; turn ms from `run.finished` `meta.ms` (`utils/reduceStreamFrame.ts:126`).
- Server (`contentstudio-ai-agents/src/api/helpers/stream_v2.py`): every frame carries `ts` (stream clock). `block.end` carries `ms` (Agno `metrics.duration` or now minus `Block.started_at`). `started_at` is not put on the wire as its own field, but the `block.start` frame's `ts` is.
- So this can be FE-only: tick from the `block.start` frame's `ts`, swap to the server `ms` on `block.end`. If clock skew between `ts` and the browser clock is a problem, compare against the first frame's `ts` rather than `Date.now()`. If that proves unreliable, a tiny BE change to put `started_at` on `block.start` is the fallback.
- Resume/reload mid-run: timer must be based on the start timestamp, not on component mount, so it doesn't reset to 0s.

## 2. Logos

**Where the logo comes from**
- Website extraction: `contentstudio-ai-agents/src/api/routers/business_info/business_info_router.py` `fetch_website_content` (Firecrawl v2 `branding.logo`), `src/agents/tools/business_info_agent.py` (`logo` field).
- Backend orchestration: `contentstudio-backend/app/Services/AI/BrandAnalysisService.php` (first account avatar as seed line 48, `ContentLibraryHelper::materializeLogoDataUrl` line 166, logo.dev fallback line 449, `sanitizeLogoReferenceUrl` line 535). Stored as brand profile `style.logo` (`app/Jobs/AI/BrandKnowledgeGenerationJob.php:167`).
- Manual upload: Brand Knowledge > Style tab, `contentstudio-frontend/src/modules/publisher/ai-content-library/components/brand-knowledge/BrandStyleTab.vue` using `components/common/LogoUploader.vue`, API `updateBrandLogoApi` (`src/api/ai-content-library.ts:322`).
- Brand asset of category LOGO: carousel falls back to it (`ai-agents/src/orchestration/carousel/brand.py:74`).
- Onboarding shows the extracted logo in `modules/onboarding/components/OnboardingBrandReviewCards.vue`.

**How it's used today**
- Images: `ai-agents/src/agents/image/brand_pass.py` `apply_brand_pass` -> `_resolve_logo` (`logo_utils.py prepare_logo_reference_url`; SVG/non-raster dropped). LLM decides treatment `product_integration | in_scene | overlay_mark | none` with free-text placement (BrandPlanner in `src/agents/brand/induction.py`, or `_decide`).
- `overlay_mark` composited by `src/agents/image/logo_stamp.py`: `_corner`, `_scale`, `_opacity`, and `_unpad`, a rough flood-fill of a flat white backdrop to alpha when the logo has no transparency. **No contrast check** against the image beneath.
- Other treatments go through a model edit with the logo as a second input; the model gets no instruction about backgrounds or contrast.
- Carousel: `src/carousel/logo.py` `inspect` detects transparency; `settle_plate` adds a contrast plate only when a **transparent** logo lacks contrast (MIN_CONTRAST 3.0). An opaque white-background logo gets no plate and shows as a white box. That's the reported bug.
- Video: `src/tools/dedicated/image_to_video.py:223` brand object is "styling only (no logo merge)". **No logo is added to videos today**, so the video case the PO described likely comes from the source image already carrying the logo. Open question below.

**Direction**
- Clean the logo once, at intake (fetch and upload): detect a flat, solid background (white or any uniform colour touching the edges) and store a transparent PNG next to the original. Keep the original so a bad cut can be reverted.
- Generation always uses the transparent version. Before placing, measure contrast between the logo and the area it will sit on. If too low: use a light or dark outline/plate, or a single-colour (white or black) version of the mark. Carousels already have `settle_plate`; images need the same check in `logo_stamp`, and model-edit treatments need the instruction in their prompt.

**Open questions for the PO**
- Video: should this story also start adding the logo to generated videos, or only fix videos whose source image is branded? Written as the latter for now.
- Should the user see and approve the cleaned logo (a before/after in Brand Knowledge), or is it silent? Written as silent with the original kept.

## 3. Typing and streaming feel

- Input: `components/dashboard/ChatInput.vue` embeds `modules/AI-tools/components/CstTextEditor.vue` (Tiptap, StarterKit + Placeholder), with `extensions/imageMention.ts` (`@`) and `extensions/skillMention.ts` (`/`) suggestion plugins rendering `ImageMentionDropdown.vue` / `SkillSlashMenu.vue` through VueRenderer.
- Output: `components/StreamingMarkdown.vue` uses `streamdown-vue` `StreamMarkdown` with `parseIncompleteMarkdown`, a `useRafFn` loop revealing text at 120 chars/s, and a `rehypeAnimateWords` plugin fading words in. **The whole markdown string is re-parsed on every reveal tick**, so cost grows with message length. That's the most likely cause of jank on long answers.
- Suspects to profile: full re-parse per tick, word-fade animation spans multiplying DOM nodes, auto-scroll fighting the user, Tiptap reactivity on every keystroke (v-model round-trips into Vue state), suggestion plugins running on every keystroke, layout shift when code blocks and tables close.
- Research deliverable first (spike inside the story), then fixes. Compare against Claude, ChatGPT and Grok (input latency, reveal cadence, scroll behaviour, how incomplete markdown is shown).

## 4. "What's new" popup

- Existing plumbing to reuse:
  - Registry `src/config/feature-announcements.ts` (`FEATURE_ANNOUNCEMENTS`: `releasedOn`, `surface`, `windowDays`, `children`).
  - `src/composables/useFeatureBadges.ts`: 30-day window, seen state in `profile.preferences.feature_badges.seen` via `setPreferenceStatus`. `components/common/NewBadge.vue`.
  - Other preference flags in `stores/core/useProfileStore.ts` (`discovery_dismissed_recommendations`, `nav_rail_expanded`), so dismiss state can live in user preferences with no new endpoint.
  - Frill changelog widget (`composables/useFrillWidget.ts`) is the app-wide changelog. This popup is chat-specific and action-driven, not a changelog.
- "Try now" actions map onto what the chat input already supports: insert text into the editor, open the `/` skill menu, open the `@` mention menu, focus the input.
- Capabilities to seed the rotation for existing users: carousels (this quarter), Skills (`/`, live since 2026-09-29), `@` attachment mentions, analytics questions, scheduling from chat, image/video generation.

## Mobile

The mobile streaming counterpart lives in the separate Flutter AI chat epic (`docs/features/flutter-ai-chat/`), per the PO on 2026-09-30. No Flutter timer story: the PO expects the app to pick it up once web and backend land. "What's new" has no mobile story yet.

## PO decision on "What's new" (2026-09-30)

Prototype: https://claude.ai/artifact/YRL6ukhZXD1QG12YgyjG8p. The PO rejected a Grok-style header popover (nothing in our header to point at) and the multi-tip card with image, counter and arrows. Chosen: a banner docked above the chat box, the same in the side panel, with badge, headline, subtext, Try now and close only. Closing hides that tip for good; the banner only returns when a new tip ships, deliberately with no "don't show again". Launch: NEW = Skills (`/`), Did you know = `@` image mentions.
