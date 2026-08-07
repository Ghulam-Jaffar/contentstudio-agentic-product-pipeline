# Story: Video clipping should respect user-selected duration and clip count

One bug ticket for reference. Web only (AI generation is web-only). Dev owns the investigation and fix.

---

## [BE] Video clipping ignores user-selected duration and clip count

### Description
In the video clipping feature, a user can pick a clip duration and the number of clips before generating. Today those manual selections are ignored: the AI still auto-generates clips with its own duration and count. Auto mode, where the user lets the AI decide, working this way is correct. But when the user has made explicit selections, the generated clips must match what they chose.

### Current behavior
Regardless of the duration and clip-count the user selects, the clipping generation decides the duration and number of clips on its own.

### Expected behavior
- Auto mode: the AI decides the duration and the number of clips (this is the current behavior and is fine).
- Manual selection: the generated clips honor the user's chosen duration and chosen number of clips.

### Steps to reproduce
1. Open the video clipping feature.
2. Choose manual settings and set a clip duration and a number of clips.
3. Generate the clips.
4. Notice the clips come back with an AI-chosen duration and count, not the ones that were selected.

### Acceptance criteria
- [ ] When the user sets a clip duration manually, every generated clip matches that duration.
- [ ] When the user sets the number of clips manually, the generation returns exactly that number of clips.
- [ ] When the user leaves the setting on Auto, the AI decides the duration and clip count as it does today.
- [ ] The user's manual duration and clip-count selections are carried through from the frontend to the clip generation and actually applied.

### Mock-ups
N/A. Behavior fix, no UI change.

### Impact on existing data
None. This corrects generation behavior, not stored data.

### Impact on other products
Web only. Video clipping (AI generation) is web-only, so no mobile or Chrome extension impact.

### Dependencies
None.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness, N/A, AI generation is web-only
- [ ] Multilingual support, N/A, no user-facing copy change
- [ ] UI theming support, N/A, no UI change
- [ ] White-label domains impact review, N/A
- [ ] Cross-product impact assessment (web only)

### Implementation references
*Pointers for reference, not a contract. Engineering owns the investigation.*
- The fix likely lives in the video clipping generation path: the manual duration and clip-count selected in the UI should be passed to and honored by the clip generation, and only Auto mode should leave those AI-decided.
- Dev to confirm where the selection is dropped: whether the frontend already sends the duration and clip count, and whether the generation service (backend or the AI clipping agent) reads and applies them or overrides them with its own auto logic.
