# Research: AI carousel templates, live preview and Carousel Maker

**Source:** Team meeting, 2026-09-30. Feedback on the AI chat carousel tool was good overall. Three additions requested:

1. Show a carousel slider with skeleton slides while it generates, and render each slide as soon as it's ready instead of all at once at the end.
2. Show the user the templates. When the user doesn't say how the carousel should look, the AI asks: pick a template (cover slides in a slider), upload a reference, let AI decide, or describe a custom style. When the prompt already describes the look, or a reference is attached, skip the question.
3. A **Carousel Maker** tool in AI Studio's sidebar: prompt, template picker, reference upload and options, calling the same carousel tool.

**Epic:** new epic, **AI Carousel 2.0: Carousel Maker and improvements** (PO, 2026-09-30). Separate from the in-progress Helpin epic "AI Carousel post generation" (`096b60bd-...`), which is the v1 build (`docs/stories/ai-agent-carousel-generation/`).

## How carousel generation works today (`contentstudio-ai-agents/`)

- Tools (`src/orchestration/carousel/tool.py`):
  - `plan_carousel(topic, slide_count)`: picks the template and outlines slides, parks the plan in a session row (`state.py`). No price, builds nothing.
  - `generate_carousel(slide_count)`: needs confirmation. The approval card shows price (`pricing.py`) and a preview (`preview.py`, built in `streaming_router.py` ~line 1232). Writes copy, renders slides, delivers to GCS (`delivery.py`, `storage.py`).
  - `edit_carousel(instruction)`.
- Design strategies:
  - **Library template, auto-selected** by the planner LLM (`writer.py plan_deck` returns `template_slug` + `why_this_template`). Templates in `src/carousel/templates/`: **9** of them: bold-editorial, brutalist-grid, clean-minimal, dark-premium, data-report, fluid-shapes, retro-print, soft-organic, warm-human. Each has `template.json` (slug, name, tagline, mood, tone, best_for/avoid_for, aspect_ratios, fonts, knobs, layouts incl. `cover`) and `theme.css`. Shared `_base/` (HTML, CSS, layouts, CONTRACT.md). Rendered as HTML by Playwright.
  - **Custom**, model-written CSS: slug `custom`, `designer.py design_theme(_checked)`. Only CSS is generated; markup, slots and character budgets are fixed. Falls back to `clean-minimal` on failure (`tool.py` ~895). A "Photo Editorial" fallback also exists for photo decks.
  - **Native, reference-matched** (`src/orchestration/carousel/native/`): the image model draws every slide, text included.
- **No cover preview images exist** for templates. `cover` in `template.json` is a layout, not an image. The picker needs rendered previews.
- References: `reference.py read_reference`, `MAX_REFERENCES = 2`. Up to 2 attachments are read as a `DesignSpec` and merged. The style is matched via native mode or the reference palette, or the reference backdrop is regenerated per slide.

## Streaming today

- Server: no per-slide progress. One tool block, kind `asset.carousel`, label `act_carousel_build` ("Building the slides"), through `tool.start` / `tool.end`. On finish the slides fan out into one block per slide (`_fan_out_assets`, `stream_v2.py` ~83-88, 165-166, 3814). An out-of-band channel exists (`mcp_progress_queue` / `_push_sse_event`, `streaming_router.py` ~1917-1963) but the carousel doesn't use it. `block.delta` can carry a `progress` dict.
- Order of work: cover photo first, then other photos in parallel (`imagery.py generate_set`, `asyncio.gather`). Copy and theme concurrently. **HTML render is sequential** (`src/carousel/render.py:533 render_slides`, one Playwright page, PNGs returned only at the end). Upload parallel (`storage.py put_many`). Native mode: cover first, rest in parallel (`gather_all`), then a parallel verify/fix pass.
- Making it stream: yield each screenshot from the `render_slides` loop. For native and photo paths, switch `gather` to `as_completed`. `_settle_legibility` and `deliver` currently run over the whole set and need to move per slide.
- Slide count is known once the plan exists (`plan_carousel`), before rendering starts, so the skeleton can be sized up front.
- Frontend (`contentstudio-frontend/src/modules/AI-tools/components/timeline/`): `AiCarousel.vue` is already a one-slide-at-a-time slider with arrows. `AiAssetGrid.vue:109` renders it when `isCarousel` and shows a pulse placeholder before any assets exist (line 104). `utils/blockAssets.ts:66-105` merges adjacent `asset.carousel` blocks and sorts by `data.index`. **The slider already copes with slides arriving one frame at a time** (cursor clamped as the count grows). Caption and note ride slide 0 only (`types/stream-frames.ts:419`). So the frontend change is mainly: show `N` skeleton slides from the plan, and swap each in as it lands.

## AI Studio tools

- Routes `src/modules/ai-studio/config/routes.ts` (`ai_studio_tool` = `tools/:toolKey`), shell `views/AiStudioShell.vue`, sidebar `components/AiStudioSidebar.vue` (categories, favourites, NEW badges via `useFeatureBadges`).
- Register in `composables/useAiToolCatalog.ts`: `TOOL_META` (line 138) or `DEDICATED_TOOLS` (~243: image-to-image, image-to-video, product-image, face-swap, upscale, motion-control). `views/AiStudioToolView.vue` routes dedicated tools to their panels or the generic `ToolPanel`.
- Carousel Maker is a dedicated tool with its own panel. It should drive the same carousel tool as chat (plan, confirm with price, generate), not a second pipeline.

## Open questions for the PO

- Should template cover previews be static (sample copy, template colours) or rendered with the user's brand colours and logo? Static is cheaper and instant. Branded looks better. Draft assumes static for v1.
- Carousel Maker results: shown in the tool's own feed like other AI Studio tools, or opened as a chat? Draft assumes the tool's feed, with "Continue in chat" to edit.
- Mobile: carousel handling in the Flutter chat is filed in the existing "AI Carousel post generation" epic (`docs/stories/flutter-ai-chat-carousel/`). A mobile template picker is not scoped yet.
- Developer surfaces: exposing templates as a list (for the MCP server / public API) would let external agents pick a template too. Not drafted; checklist item carries it.
