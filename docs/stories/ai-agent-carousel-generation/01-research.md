# Research — Carousel post generation in the AI agents platform

Request: the AI agent can already return several images in one response, but they are not a carousel. There is no title/cover slide, no body slides, no closing slide, no shared theme, and no single caption for the set. It should be able to generate a proper carousel post that can then be scheduled.

## Backlog check

No existing story covers carousel generation. `instagram-carousel-media-limit-gate` is about enforcing Instagram's media count limit in the composer, not about generating one. The AI Studio stories (`ai-studio-*`), `ai-powered-composer-customization` and `evergreen-ai-variations` all deal with single images or independent variations. This is new.

## What exists today

### Multi-image generation, but as independent images

`contentstudio-ai-agents/src/agents/image/image_generator.py` line 196:

```
def generate_multiple_image_prompts(self, content, image_style=None, max_images=5)
    """Generate multiple image prompts based on content analysis."""
```

The output schema is `contentstudio-ai-agents/src/agents/image/schemas.py`:

```
class ImagePromptItem(BaseModel):
    prompt: str      # description of the image
    filename: str    # kebab-case name
    alt: str         # alt text
    caption: str     # "Brief, informative caption (1-2 sentences)"

class ImagePromptOutput(BaseModel):
    images: list[ImagePromptItem]
```

Three things follow from that shape:

- **Every image gets its own caption.** There is no concept of one caption for the set, which is what a carousel post needs.
- **There are no slide roles.** Nothing marks an item as the cover, a body slide or the closing slide.
- **Order is incidental.** The list has an order but nothing gives it meaning or guarantees it survives generation.

### Nothing binds the set together visually

- `image_style` is a single free-text string applied to the batch (`prompt_builder.py` line 34), defaulting to `"professional and modern"`. It is a hint, not a lock.
- Brand styling exists and is applied per image: `prompt_builder.py` pulls `BrandVoiceParser.get_visual_style_description(brand_voice)` when the payload has no explicit style (lines 119, 153).
- Seeds work against cohesion rather than for it. `GeneratedImage` carries a `seed`, and the variation path at `image_generator.py` line 789 uses `seed=42 + i` — *a deliberately different seed per variation*. For a carousel the requirement is the opposite: hold enough of the generation constant that the slides look like one deck.

### A hard rule that currently blocks the cover slide

`prompt_builder.py` line 54, in the content-analysis prompt:

> **No Rendered Text**, but Keep Real Subjects: Do not bake words, letters, numbers, signs, or labels INTO the image.

That rule is sound for ordinary image generation, where model-drawn text is unreliable. But a carousel cover is defined by its title text, and body slides usually carry a headline. So the feature cannot be built by relaxing this rule and hoping the model spells correctly.

### There is no text layer. "Composite" means a second model pass

This corrects an earlier reading of `brand_pass.py`. Its docstring says it will *"composite the real logo and any explicit overlay text"*, which sounds like layer compositing. It is not.

Line 195: *"Composite model (logo + text) reads from `config.brand_composite_model` (default nano-banana-pro — strong text, multi-input, cheaper than gpt-image-2)."*

So the "composite" is **a second image-to-image model pass**, and the model still draws the glyphs. The prompt it builds (line 286) reads:

> Add this exact text, rendered verbatim with correct spelling: "{overlay_text}". Do not alter its wording or capitalization.

And because an i2i re-render can damage the base image, the text-only branch has to defend against it (line 315):

> Text-only edit: pin EVERYTHING incl. existing marks so the re-render can't corrupt them. Every existing logo, brand mark, sponsor decal, badge, face, colour, texture, and reflection in the base image must stay pixel-for-pixel identical.

**There is no programmatic text rendering anywhere in the service.** Pillow is a dependency (`pyproject.toml`) but is not used to draw text onto generated images. Grepping for `ImageDraw`, `ImageFont` and `truetype` across `src/` returns nothing relevant.

Two consequences, both significant for this epic:

1. **A wording change is not free today.** It costs a model call, and the underlying art can drift, which is exactly what all that pixel-for-pixel language is trying to prevent.
2. **Slide text is not data.** Once the model has drawn it, the string is baked into pixels. It cannot be re-rendered, translated, restyled, or corrected without going back to a model.

### How this is normally solved

Every carousel and template tool in this space (Canva, Predis, PostNitro, AdCreative, Placid, Bannerbear) separates the two layers: the image model produces **background art only**, and the application renders **text as a real text layer** over it, keeping the string as data on the slide.

| | Model-drawn text (today) | Application-rendered text layer |
|---|---|---|
| Spelling | Usually correct, sometimes not | Always correct |
| Changing the wording | Another image-to-image pass; art can drift | Re-render; milliseconds; art untouched |
| Brand typeface | An approximation of it | The actual font file |
| Same input twice | Varies | Identical |
| Translating a slide | Regenerate | Re-render |
| Text available downstream | No, it is pixels | Yes, still a string |

The trade-off is real and worth stating: model-drawn text can sit *inside* a scene, wrapping around a subject or falling in perspective on a surface. A rendered layer reads as a designed overlay. For carousels the overlay is the right answer, because decks want typographic clarity and reliable spelling far more than they want text painted into a photograph. That is why the tools above all work this way.

### We already run a rendering stack. It is just in the reports service

This corrects the paragraph above, which said we have never had this. **We have.** `contentstudio-social-analytics-go` operates all of it in production today, for analytics PDF reports:

- **A Node rendering sidecar.** `sidecar/server.js` with its own `package.json`, `Dockerfile`, and k8s manifests (`report-chart-sidecar-deployment.yaml`, `report-chart-sidecar-service.yaml`). It runs real ECharts and returns PNG, described in `charts/renderer.go` as *"pixel-matching the UI"*. Configured via `cfg.Reports.ChartSidecarURL` or `CHART_SIDECAR_URL`, with a dependency-free Go fallback renderer for when it is unavailable.
- **SVG to vector and raster.** `render/vector.go` calls a `/vectorize` endpoint backed by `rsvg-convert` / cairo, which *"writes `<text>` out as real"* text and subsets each font once per document.
- **Managed fonts.** `render/fonts/` ships `DejaVuSans.ttf` and `DejaVuSans-Bold.ttf`, with `render/fonts.go` alongside.
- **SVG composition primitives.** `render/svgcompose.go` nests one SVG document inside another so a whole page composes as a single SVG; `card_svg.go`, `ads_svg.go`, `insights_svg.go`, `pill.go`, `cover.go` and `icons.go` are the tile builders.

So the capability, the font pipeline, the rasterizer, the deployment and the operational experience of running a render sidecar all exist. What does not exist is any of it being reachable from the AI image path, or shaped for a 4:5 slide rather than an A4 page.

**That changes the size of story 2 substantially.** It is not "build a rendering stack from nothing". It is "make an existing rendering stack serve a second consumer".

### Carousel aspect ratios are already understood

`prompt_builder.py` lines 309 to 311 already describe carousel-appropriate dimensions:

- `portrait_3_4` — "Great for taller Instagram carousels"
- `landscape_3_2` — "Balanced landscape option for carousel covers"
- `landscape_16_9` — "Good for carousel posts"

So the dimension vocabulary is in place; nothing consumes it as a carousel.

### Publishing already accepts carousels

`contentstudio-ai-agents/src/integrations/contentstudio/toolkits/publishing_write.py` lists `carousel` and `carousel+story` among its valid post types (lines 38, 43), and the scheduling tool takes `post_type` plus `image_urls` as a list (lines 326 to 327), with an explicit instruction that images from earlier turns must be carried through via the `[Image handles]` block.

**So the second half of the request already works.** A set of image URLs can be scheduled as a carousel today. The missing piece is entirely on the generation side: producing a coherent, ordered, themed set with one caption, and handing it to that publishing path as a single post rather than as loose images.

## The gap, stated precisely

| Needed for a carousel | Today |
|---|---|
| Ordered slides with roles (cover, body, closing) | A flat list with no roles |
| One caption for the post | One caption per image |
| Consistent theme across slides | A free-text style hint per batch, and deliberately varied seeds |
| Title and headline text on slides | Model is explicitly forbidden from drawing text; a compositing path exists but has no slide-text concept |
| Delivered as one schedulable post | Publishing accepts it, but generation does not produce it as a unit |

## What needs to change

**Generation**
- A carousel becomes a first-class output: an ordered set of slides, each with a role, sharing one caption and one theme.
- The agent decides slide count and the narrative arc (what the cover promises, what the body delivers, what the closing asks for), rather than returning N loosely related images.
- Theme is locked across slides: palette, style, and whatever generation parameters give visual continuity, sourced from Brand Knowledge where the workspace has it.

**Slide text — v1 decision: reuse what exists**

The PO has scoped v1 to add **no new rendering, imaging or font infrastructure**. Slide text is drawn by the model through the existing `brand_pass` overlay path, and every slide edit is a regeneration.

- Wire carousel slides through the existing overlay pass so each slide gets its title or headline.
- Keep the generation step's no-rendered-text rule; text arrives in the second pass, as it does today.
- Cap text length per slide role, since length is the main lever on how reliably the model renders it.
- Store the text the assistant wrote on the slide, so the deck knows what each slide should say even though the pixels are what ship.

The known ceiling is accepted and recorded in the epic: text is re-rolled on every edit, the art can drift, every wording change costs a generation, the text is pixels rather than data, and the brand's actual typeface is approximated.

**Deferred, not discarded: the rendered text layer.** It is the fix for all five, and the company already runs the stack for it in the analytics reports service (a Node render sidecar on k8s, SVG composition, managed fonts, cairo rasterization). If the ceiling starts costing users, the work is extending an existing renderer rather than building one. The detail below is kept for that decision.

**Hand-off**
- The carousel reaches the composer and scheduler as one post: ordered image URLs, one caption, `post_type: carousel`.

**Editing**
- A carousel is a stateful artifact that is **patched in place**, not re-emitted. An edit returns the same carousel with a version bump rather than a second deck in the conversation, which is what keeps "schedule it" unambiguous.
- Edits route by cost rather than all being treated as regeneration: a wording change is a re-composite with no image generation, an imagery change regenerates one slide against the locked theme and keeps its text, and a caption change touches no slide at all.
- Slide identity has to be stable and independent of position, or a positional reference breaks the moment the user reorders the deck.

**Frontend**
- AI chat renders the result as a carousel the user can page through, not as N stacked images.
- The user can edit the caption, reorder or remove slides, and regenerate a single slide without losing the rest, before scheduling.
- An edit made by talking to the assistant updates the carousel already in the conversation rather than appending another one.

## Open questions for product

- **Slide count.** Agent-chosen from the content, user-specified, or both with a default? Instagram allows up to 20; a sensible default is likely 5 to 7.
- **Which networks.** Instagram is the obvious target. Carousels also exist on LinkedIn (as a document post, a different mechanism entirely) and Facebook. Confirm v1 scope before building, since LinkedIn document carousels are not the same feature.
- **Editing depth.** Can a user edit a slide's overlay text directly, or only regenerate the slide? Direct text editing is far more useful and materially more work.
- **Slide reference resolution.** Positional ("slide 3"), ordinal ("the last one") and role-based ("the cover") are cheap to resolve. Semantic ("the one about pricing") is not. Confirm which are in v1 rather than letting the unsupported kind fail oddly.
- **Templates.** Is there a defined set of slide layouts, or does the agent compose freely within a theme? A small template set is more predictable and more likely to look designed.
- ~~**Where slide rendering lives.**~~ **Deferred out of v1** by the no-new-infrastructure decision. Kept here for when the ceiling is revisited:
  - **Extend the existing render sidecar** with a slide endpoint. Reuses the fonts, the rasterizer, the k8s deployment, the fallback pattern and the team's experience of running it. `ai-agents` calls it over HTTP, exactly as the Go service already does. Recommended.
  - **Build SVG to raster inside `ai-agents`** (resvg, cairosvg). Keeps carousels in one repo, at the cost of a second font pipeline and a second rasterizer doing what we already run elsewhere.
  - **Pillow, inside `ai-agents`.** Least work, already a dependency, but no real text shaping, crude wrapping and poor complex-script support. Adequate for a watermark, wrong for a cover headline.
  If the sidecar is extended, a second question follows: **SVG or HTML/CSS for slide templates.** SVG reuses the existing librsvg path exactly. HTML/CSS gives better typography (wrapping, balancing, OpenType) and gives the design team templates in a language they already work in, but adds a browser-class renderer to a service that currently does not need one. The sidecar being Node makes either viable.
  Either way, the deciding property is unchanged: **re-rendering must not involve an image model.**
- **How much text a slide should carry.** The live v1 question, and the one with the most effect on quality. Model-drawn text degrades as it lengthens, so the cap is doing real work and belongs in the design story as a number.
- ~~**Does `brand_pass` also move?**~~ Moot in v1, since carousels use `brand_pass` rather than replacing it. It returns if a text layer is ever built.

## Files involved

`contentstudio-ai-agents`:
- `src/agents/image/image_generator.py` — `generate_multiple_image_prompts`, the generation path
- `src/agents/image/schemas.py` — `ImagePromptItem`, `ImagePromptOutput`, `GeneratedImage`
- `src/agents/image/prompt_builder.py` — the content-analysis prompt, the no-rendered-text rule, dimension vocabulary
- `src/agents/image/brand_pass.py` — the existing overlay/compositing pass
- `src/agents/image/logo_utils.py`, `dimensions.py`, `models_registry.py`
- `src/integrations/contentstudio/toolkits/publishing_write.py` — the scheduling path that already accepts carousels
- `src/orchestration/handlers/image.py`, `src/api/routers/image_router.py`

`contentstudio-frontend`:
- `src/components/dashboard/` — the AI chat surface that renders generated media
- `src/modules/common/components/social-previews/SocialPreviews/InstagramMultimediaPreview.vue` and `SwipeCarousel.vue` — the existing carousel preview components

## Mobile

Carousel **generation** is an AI generation feature and therefore web only. But AI chat itself is on mobile, so a carousel produced on web can appear in a conversation the user later opens on mobile. Mobile does not need to generate carousels, but it must not break when it encounters one in chat history. Worth confirming with the mobile team rather than assuming.
