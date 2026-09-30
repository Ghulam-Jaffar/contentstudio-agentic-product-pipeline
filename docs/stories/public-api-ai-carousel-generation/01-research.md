# Research: AI carousel generation in the public API, CLI and MCP server

**Source:** PO, 2026-09-30. Target epic: the existing **AI Carousel post generation** (`096b60bd-5376-4dc6-a87d-647fdf1f008c`). High priority, assigned to Hammad Jamil (member `fd17aca1-5b03-4cfb-8f5f-0eb98390e9d0`, already owner on the carousel BE stories).

Ask: images and videos are on the public API, CLI and MCP. Carousels must be too. A caller can fetch templates and pass one in, pass references, or pass only a prompt. **The options must be the same across the API, AI chat and the Carousel Maker tool.**

## Pattern to follow (live today)

`contentstudio-backend/routes/api/v1.php` ~925-990:
- AI routes sit in their own group: `['force.json', 'api.key', 'api.request.log', 'set.locale', 'throttle:ai-tools']` plus `PermissionMiddleware`.
- Images (`AiImageController`): `GET ai/images/tools`, `GET ai/images/models`, `GET ai/brand`, `POST ai/images/generate`, `POST ai/images/tools/{tool_key}`. Synchronous, returns media.
- Videos (`AiVideoController`): `tools`, `models`, `estimate` (line 429), `generate` (583) and `invokeTool` return a job with a status URL (`jobStatusUrl`, 810).
- Jobs (`AiJobController`): `GET ai/jobs`, `GET ai/jobs/{job_id}`, `DELETE ai/jobs/{job_id}`.
- Carousels take a minute or more, so they should follow the **video** pattern: estimate, then generate returns a job, then poll the job.
- Original specs: `docs/features/public-api-ai-image-generation/04-epic-and-stories.md` and `docs/features/public-api-ai-video-generation/`. The image epic split API, CLI, MCP, automation apps and docs into 5 stories. Here the PO asked for one ticket covering API, CLI and MCP.

## Carousel engine (`contentstudio-ai-agents/`)

- `src/orchestration/carousel/tool.py`: `plan_carousel(topic, slide_count)`, `generate_carousel(slide_count)` (priced confirmation), `edit_carousel(instruction)`.
- 9 library templates in `src/carousel/templates/` (bold-editorial, brutalist-grid, clean-minimal, dark-premium, data-report, fluid-shapes, retro-print, soft-organic, warm-human), plus `custom` (model-written CSS) and native reference-matched mode.
- References: `reference.py`, `MAX_REFERENCES = 2`.
- Pricing: `pricing.py`. Delivery to GCS: `delivery.py`, `storage.py`.
- Template cover previews don't exist yet. Created by **[BE] Ask for a carousel template before generating, with a cover preview for every template** (CONT-4191).

## Option contract (one definition, three callers)

| Option | Values | Notes |
|---|---|---|
| `prompt` | text | required |
| `slide_count` | engine range | default as chat |
| `size` | `square` / `portrait` | default portrait |
| `style.mode` | `auto` / `template` / `reference` / `custom` | default `auto` when nothing else is given |
| `style.template_id` | from the templates list | with `template` |
| `style.references` | up to 2 media IDs or URLs | with `reference` |
| `style.description` | text | with `custom` |
| `use_brand` | boolean | same semantics as the image API (`brand_applied` in the response) |

The API never asks the interactive "how should it look" question. No style given means `auto`.
