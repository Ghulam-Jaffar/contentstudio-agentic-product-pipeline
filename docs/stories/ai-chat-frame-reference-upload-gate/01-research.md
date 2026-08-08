# Research — AI Chat upload gate blocks reference images when no frame is set

Item 25 of the 7 Aug 2026 backlog batch.

## Reported behaviour

In AI Chat, the upload button becomes disabled with a message saying start/end frames and reference images cannot be used together, even when the user has not uploaded a first or last frame.

## Root cause

The bug is one message serving two different causes. Everything below is in `contentstudio-frontend/src/components/dashboard/ChatInput.vue`.

### The gate

```
const genericUploadsAllowed = computed(() => {
  const support = startEndFrameSupport.value
  if (!support) return true
  return (
    support.allowReferences &&
    (support.referencesWithFrames || !hasFrames.value)
  )
})
```
(line 1436)

That expression is false for **two unrelated reasons**:

1. `support.allowReferences === false` — the model does not accept reference images at all, regardless of frames. **No frame is involved.**
2. `hasFrames === true && !support.referencesWithFrames` — a frame is set and this model cannot combine the two.

### The message

The upload dropdown is disabled off that single flag and shows one tooltip regardless of which cause fired:

```
v-tooltip="
  genericUploadsAllowed
    ? ''
    : $t('dashboard.chat_input.frames_references_exclusive')
"
...
:disabled="!genericUploadsAllowed"
```
(lines 441 to 454, and the same pattern in the compact toolbar via `:can-upload="genericUploadsAllowed"` at line 433)

`dashboard.chat_input.frames_references_exclusive` is "Start/end frames and reference images can't be used together."

So on a frames-only model, cause 1 fires, uploads are disabled, and the user is told the two cannot be combined. There is nothing to un-combine: no frame is set, and none can help. The message points them at a fix that does not exist.

### The same shape in two more places

**Frame slot tooltip** (line 116 to 127):

```
const framesDisabled = computed(() => {
  const support = startEndFrameSupport.value
  if (!support) return false
  return (
    hasReferences.value &&
    (!support.allowReferences || !support.referencesWithFrames)
  )
})
```
(line 1448)

Two causes again, one message, same string.

**Frame validation message** (line 1395 to 1413):

```
const frameReferenceConflict = computed(() => {
  const support = startEndFrameSupport.value
  if (!support || !hasReferences.value) return false
  if (!support.allowReferences) return true
  return hasFrames.value && !support.referencesWithFrames
})
```
(line 1429)

`if (!support.allowReferences) return true` returns a conflict **without checking `hasFrames` at all**. `frameSelectionInvalid` then goes true and `frameValidationMessage` returns the same exclusive string.

## The model support table

`contentstudio-frontend/src/composables/useVideoGeneration.ts` defines `startEndFrameSupport` per model with three flags (documented at lines 61 to 71): `bothRequired`, `allowReferences`, and optional `referencesWithFrames`.

Three groups exist today:

| Group | Flags | Models |
|---|---|---|
| Frames only, references not accepted | `allowReferences: false` | `kling-video-v2.5-turbo-pro`, `flux-3` |
| Frames and references both accepted together | `allowReferences: true, referencesWithFrames: true` | `kling-video-v3-standard`, `kling-video-o1` |
| Frames and references mutually exclusive | `allowReferences: true`, `referencesWithFrames` unset | the remainder, including the `bothRequired: true` models |
| No frame support at all | `startEndFrameSupport` absent, so `support` is null | ordinary video models, uploads always allowed |

The bug bites the first group: `kling-video-v2.5-turbo-pro` and `flux-3`. For those models the correct statement is "this model uses start and end frames instead of reference images", not "they can't be used together".

## Worth noting: the disabling itself may be correct

For a frames-only model, references genuinely are not supported, so blocking the upload is defensible. `imageRequiredButMissing` (line 1363) already treats a frame as satisfying the image requirement, with a comment naming "Kling o1 with first frame, no references". So the frame *is* the image input for those models.

The defect is therefore primarily the message, plus the `frameReferenceConflict` computed reporting a frames-versus-references conflict when no frame exists. A user should be told what this model actually wants, not given a rule about a combination they never attempted.

## Related normalisation that already works

`contentstudio-frontend/src/composables/useAIChatMessage.ts` `updateVideoModel` (line 248) already normalises on model switch: it drops frames the new model cannot accept, and at line 272 drops references when moving to a model with `allowReferences: false` ("Drop references on frames-only models"). So switching models cleans up correctly. The problem is purely the steady-state gate and its wording.

## What needs to change

- Separate the two causes wherever they are currently collapsed: the upload gate, the frame-slot disabled state, and the frame validation message.
- Show the exclusive message only when a frame is actually set and the model cannot combine the two.
- Show a distinct, accurate message when the model simply does not accept reference images.
- Fix `frameReferenceConflict` so it does not report a frames-versus-references conflict with no frame set.
- Verify across all four support groups and every model in each.

## Existing copy

`contentstudio-frontend/src/locales/*/dashboard.json`:

- `chat_input.start_frame` = "Start Frame"
- `chat_input.end_frame` = "End Frame"
- `chat_input.frames_both_required` = "Add both a start and end frame, or neither."
- `chat_input.end_frame_needs_start` = "Add a start frame to use an end frame."
- `chat_input.frames_references_exclusive` = "Start/end frames and reference images can't be used together."
- `chat_input.frame_upload_failed` = "Frame upload failed. Please try again."
- `chat_input.tooltip_upload_image` (upload button tooltip)

One new key is needed for the references-not-supported case.

## Files involved

- `contentstudio-frontend/src/components/dashboard/ChatInput.vue`
- `contentstudio-frontend/src/components/dashboard/ChatInputCompactToolbar.vue` (receives `can-upload`)
- `contentstudio-frontend/src/composables/useVideoGeneration.ts` (the support table, read only)
- `contentstudio-frontend/src/composables/useAIChatMessage.ts` (model-switch normalisation, likely unchanged)
- `contentstudio-frontend/src/locales/*/dashboard.json`

## Mobile

AI chat exists on mobile, but the frame and reference-image video inputs are a web AI-generation surface. No mobile story. If the mobile app ever exposes frame inputs, the same distinction applies there.

## Related existing work

- `docs/stories/video-model-reference-audio-video-inputs/` — added reference audio and video inputs for MiniMax H3, which introduced `referenceFiles` and `referenceClips`, both of which feed `hasReferences`. Worth reading before changing that computed.
- `docs/stories/flux-3-video-models/` — Flux 3, one of the two models this bug affects.
