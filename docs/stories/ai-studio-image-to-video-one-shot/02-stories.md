# Stories: Convert Image to Video into a one-shot tool

We recently converted Image to Image into a one-shot (dedicated) tool. Do the same for Image to Video so it runs as a single dedicated tool rather than the older chat-driven flow.

---

## [BE] Add Image to Video as a dedicated one-shot tool

### Description

Expose Image to Video as a dedicated one-shot tool in the AI backend, mirroring the existing Image to Image dedicated tool. It takes an input image (and prompt/options) and returns a generated video as an async job.

### Acceptance criteria

- [ ] Image to Video is available as a dedicated tool with its own key and endpoint, listed by the dedicated-tools listing.
- [ ] The tool accepts an input image and the supported generation options and returns a video via the async video job flow (`output_type: video`).
- [ ] The tool is credit-metered using the existing video credit rules.
- [ ] Missing or invalid inputs return a clear, structured error.
- [ ] Behavior matches the Image to Image dedicated tool's request/response contract, adapted for video output.

### Impact on existing data

If the old chat-driven image-to-video path is retired, confirm migration/handling for any in-flight jobs. Otherwise additive.

### Impact on other products

Web-only AI generation. Mobile and Chrome extension: N/A.

### Dependencies

None. Pairs with the frontend story below.

### Global quality and compliance checklist

- [ ] Mobile responsiveness (N/A, backend-only story)
- [ ] Multilingual support verified (status/error messages localized or fallback handled)
- [ ] UI theming supported (N/A, backend-only story)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

- Dedicated (one-shot) tool framework: `contentstudio-ai-agents/src/tools/dedicated/` (`engine.py`, `route_factory.py`, `specs.py`, `models.py`), with `image_to_image.py` as the closest pattern to mirror.
- Router wiring: `contentstudio-ai-agents/src/api/routers/dedicated_tools_router.py` (add the new tool key, POST endpoint, and its capability to `list_tools` / `get_tool`).
- The base request schema already supports `output_type: video` (async job), so the video path exists.

---

## [FE] Surface Image to Video as a one-shot tool in AI Studio

### Description

Show Image to Video as a one-shot tool in AI Studio, matching how Image to Image now works, so users run it directly instead of through the older flow.

### Workflow

1. User opens AI Studio tools.
2. User picks "Image to Video".
3. User provides an image (and any options) and runs it.
4. The generated video appears when the job completes, with progress shown while it runs.

### Acceptance criteria

- [ ] Image to Video appears as a one-shot tool in AI Studio, consistent with the Image to Image tool.
- [ ] The user can provide an input image and run the tool in one shot.
- [ ] Progress/loading is shown while the video job runs, and the result appears on completion.
- [ ] Errors show a clear, user-friendly message.
- [ ] The older image-to-video entry point is replaced by (or routed to) the one-shot tool.

### UI copy

- Tool name: "Image to Video"
- Loading state: "Generating your video..."
- Error state: "We could not generate the video. Please try again."

### Impact on existing data

None (frontend surfacing).

### Impact on other products

Web only. Mobile and Chrome extension: N/A.

### Dependencies

Depends on: **[BE] Add Image to Video as a dedicated one-shot tool**.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only)
- [ ] Multilingual support verified (translations available or fallback handled)
- [ ] UI theming supported (default and white-label) — no dark mode
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

- Follow the frontend pattern established for the Image to Image one-shot tool in `contentstudio-frontend/src/modules/AI-tools/` (tool entry, run flow, and result handling). `components/MediaGenerationOptions.vue` and the video job composable `composables/useChatVideoJobs.ts` are relevant for options and job progress.
