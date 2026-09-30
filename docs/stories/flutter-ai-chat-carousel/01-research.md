# Research: Carousel generation in the Flutter AI chat

**Source:** PO, 2026-09-30. Add a `[Flutter]` story to the **existing** Helpin epic **AI Carousel post generation** (`096b60bd-5376-4dc6-a87d-647fdf1f008c`) so carousel generation is handled in the Flutter app. Checked the epic on 2026-09-30: it has CONT-3376, 3377, 3378, 3379 (all In Review) and CONT-3380 (Design, Done). **No Flutter story exists**, and a workspace search for Flutter AI chat tasks returned nothing.

## Web behaviour to match (v1 epic)

- `[BE] Generate carousels as an ordered, themed slide set with one caption` (CONT-3376)
- `[BE] Put title and headline text on carousel slides using the existing overlay pass` (CONT-3377)
- `[BE] Edit a carousel by regenerating the slide the user names` (CONT-3378)
- `[FE] Review, edit and schedule a generated carousel from AI chat` (CONT-3379)
- Server tools: `plan_carousel` → approval card with price and preview → `generate_carousel` → `edit_carousel` (`contentstudio-ai-agents/src/orchestration/carousel/tool.py`). Finished slides fan out as one `asset.carousel` block per slide with `data.index`. Caption and note ride slide 0 only (web `types/stream-frames.ts:419`). Web merges adjacent `asset.carousel` blocks and sorts by index (`contentstudio-frontend/src/utils/blockAssets.ts:66-105`), rendered by `timeline/AiCarousel.vue`.

## Flutter today (`contentstudio-flutter/lib/features/ai_assistant/`)

- No carousel-specific handling: no `asset.carousel` references in the feature.
- `domain/timeline/timeline_layout.dart:261`: adjacent image sets draw as one swipeable carousel ("adjacency is the whole test"). There is no ordering by slide index, no single caption for the set, and no "schedule as carousel post" action.
- v1 carousel in `presentation/widgets/ai_message_bubble.dart:116` and suppression rules in `domain/timeline/timeline_resolver.dart:103-136`.
- Approval gate card `ai_preview_card.dart` (Generate / Cancel / Refine, credits line) already exists and should serve the carousel approval.
- Draft and Schedule route through the Composer's publishing options. The composer supports carousel post previews for Facebook and LinkedIn (`composer/presentation/preview/platforms/*`), and `composer/data/mappers/plan_hydrator.dart` handles carousel plans. Check that multi-image posts hand over in slide order as a carousel post type where the platform supports it.

## 2.0 follow-ons

The new epic **AI Carousel 2.0: Carousel Maker and improvements** adds slide-by-slide delivery and a template picker. The story below asks the app to cope with slides arriving one by one and to show skeletons when a slide count arrives, so it doesn't break when the 2.0 backend ships. The template picker on mobile falls back to plain text until a mobile picker is scoped.
