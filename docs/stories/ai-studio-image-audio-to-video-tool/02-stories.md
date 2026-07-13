# Stories: Add Image + Audio to Video one-shot tool

A new AI Studio one-shot tool that takes an image plus an audio track and produces a video (for example a talking avatar), using the `ltx-2.3/audio-to-video` and `longcat-single-avatar/image-audio-to-video` models.

---

## [BE] Add Image + Audio to Video as a dedicated one-shot tool

### Description

Add a new dedicated one-shot tool that accepts an input image and an audio file and returns a generated video, mirroring the existing dedicated-tool pattern. It uses the `ltx-2.3/audio-to-video` and `longcat-single-avatar/image-audio-to-video` models.

### Acceptance criteria

- [ ] Image + Audio to Video is available as a dedicated tool with its own key and endpoint, listed by the dedicated-tools listing.
- [ ] The tool accepts an input image and an input audio file and returns a video via the async video job flow.
- [ ] It supports the `ltx-2.3/audio-to-video` and `longcat-single-avatar/image-audio-to-video` models.
- [ ] The tool is credit-metered using the existing video credit rules.
- [ ] Missing image, missing audio, or unsupported file types return a clear, structured error.
- [ ] The request/response contract matches the other dedicated tools, adapted for two inputs (image + audio).

### Impact on existing data

None. Additive new tool.

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

- Dedicated (one-shot) tool framework: `contentstudio-ai-agents/src/tools/dedicated/` (`engine.py`, `route_factory.py`, `specs.py`, `models.py`), with `image_to_image.py` as the closest existing tool to mirror. Add a request schema that takes both `image_url` and an `audio_url`/audio file, with `output_type: video`.
- Register the two models (`ltx-2.3/audio-to-video`, `longcat-single-avatar/image-audio-to-video`) in the model registry / provider layer (`src/agents/image/models_registry.py`, `src/agents/image/providers/fal.py`).
- Router wiring: `contentstudio-ai-agents/src/api/routers/dedicated_tools_router.py` (new key, POST endpoint, capability in `list_tools` / `get_tool`).

---

## [FE] Add Image + Audio to Video tool in AI Studio

### Description

Add the Image + Audio to Video one-shot tool to AI Studio so users can upload an image and an audio file and generate a video.

### Workflow

1. User opens AI Studio tools and picks "Image + Audio to Video".
2. User provides an image and an audio file.
3. User optionally selects the model.
4. User runs the tool; progress shows while the video job runs, and the result appears on completion.

### Acceptance criteria

- [ ] "Image + Audio to Video" appears as a one-shot tool in AI Studio.
- [ ] The user can upload both an image and an audio file before running.
- [ ] Run is disabled until both an image and an audio file are provided.
- [ ] Progress/loading shows while the job runs; the generated video appears on completion.
- [ ] Errors show a clear, user-friendly message.
- [ ] If more than one model is available, the user can choose between them.

### UI copy

- Tool name: "Image + Audio to Video"
- Image input label: "Image"
- Audio input label: "Audio"
- Loading state: "Generating your video..."
- Missing input hint: "Add both an image and an audio file to continue."
- Error state: "We could not generate the video. Please try again."

### Impact on existing data

None.

### Impact on other products

Web only. Mobile and Chrome extension: N/A.

### Dependencies

Depends on: **[BE] Add Image + Audio to Video as a dedicated one-shot tool**.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only)
- [ ] Multilingual support verified (translations available or fallback handled)
- [ ] UI theming supported (default and white-label) — no dark mode
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

- Follow the frontend one-shot tool pattern in `contentstudio-frontend/src/modules/AI-tools/` used for Image to Image and (once built) Image to Video. Reuse the media-upload widget `src/components/common/MediaImportDropzone.vue` for the image and audio inputs, and the video job composable `composables/useChatVideoJobs.ts` for progress.
