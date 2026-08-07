# Stories: Seedance 2.0 4K resolution support

One story.

No design story. This adds an option to an existing resolution selector with no new layout or states.

---

## [Full Stack] Add 4K resolution to Seedance 2.0 with correct credit pricing

### Description

Seedance 2.0 can produce 4K video. We only offer 480p, 720p, and 1080p, so a user who wants the model's highest quality output cannot get it from ContentStudio even though the model supports it and we already pay for access to it.

This story adds 4K as a selectable resolution for Seedance 2.0 and prices it correctly.

The pricing half is not a formality. 4K costs roughly four times what 1080p costs per second, so a 15 second 4K video is a meaningfully expensive generation. The credit charge has to reflect that, and there is a specific way this goes wrong: the credit calculator falls back to the cheapest configured resolution when the selected one has no price entry. Exposing 4K in the selector without adding its cost in both the frontend pricing matrix and the backend model definition would charge users 480p credits for 4K video, silently. That is roughly a twentyfold undercharge per second, with no error raised.

### Workflow

1. The user opens video generation and selects Seedance 2.0.
2. The resolution selector offers 4K alongside 480p, 720p, and 1080p.
3. The user selects 4K. The credit estimate updates to reflect the real cost of a 4K generation, before they generate.
4. The user generates. If they do not have enough credits, they are told before the generation starts, as with any other resolution.
5. The video is produced at 4K and delivered through the existing flow.
6. Other models' resolution options and credit costs are unchanged.

### UI copy

- Resolution option label: `4K`
- No new tooltips, hints, or empty states. The existing credit estimate and insufficient-credit messaging cover this, since they are already driven by the calculated cost.

If product decides to gate 4K by plan, that needs its own copy and is not covered here. Flagged under Dependencies.

### Acceptance criteria

- [ ] 4K is selectable as a resolution for Seedance 2.0 in text-to-video mode
- [ ] 4K is selectable in image-to-video and reference-to-video modes, or is deliberately excluded there because fal does not accept it on those endpoints, with the finding recorded in the pull request
- [ ] The resolution value sent to the model matches the string fal expects, in the exact casing fal uses
- [ ] The credit estimate shown before generating reflects 4K's actual per-second cost, not the cost of any other resolution
- [ ] A 4K generation is charged credits proportional to its real cost, verified against a 1080p generation of the same duration on the same model
- [ ] The 4K per-second cost is confirmed against fal's published rate before release, not shipped from an internal derivation
- [ ] Existing 480p, 720p, and 1080p generations on Seedance 2.0 are charged exactly what they are charged today
- [ ] No other model gains a 4K option
- [ ] Seedance 2.0 Fast is unchanged, since it does not support resolutions above 720p
- [ ] Insufficient-credit handling works for 4K the same way it works for other resolutions, refusing before the generation starts rather than failing partway

### Mock-ups

N/A. One new option in an existing selector.

### Impact on existing data

- No schema changes. Resolution support and per-second costs live in model configuration.
- Existing generations, their stored resolutions, and their historical credit charges are untouched.

### Impact on other products

- Mobile apps: no impact. AI video generation is web only.
- Chrome extension: no impact.
- White-label domains: no impact.

### Dependencies

- **Before build:** confirm fal's published 4K per-second rate and confirm whether `4k` is accepted on the image-to-video and reference-to-video endpoints, not just text-to-video.
- **Product decision:** whether 4K should be available on all plans or gated. At roughly four times the 1080p rate, a single 15 second 4K video is a significant credit spend. If gating is wanted, that is separate work with its own upgrade messaging.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness tested
- [ ] Multilingual support verified (translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

**Two places carry the same numbers and both need updating:**

- `contentstudio-ai-agents/src/utils/model_registry.py`, the `seedance-2.0` entry from line 1547: `supported_resolutions` currently `["480p", "720p", "1080p"]`, and `api_cost` currently `{ "480p": 0.1346, "720p": 0.3034, "1080p": 0.6804 }` in USD per second. The `mode_requirements` blocks for image-to-video and reference-to-video carry their own `supported_resolutions` overrides with the same three values, so model level alone is not enough.
- `contentstudio-frontend/src/composables/useVideoCreditCalculation.ts`, `seedance-2.0` at line 289, which mirrors the same `api_cost` map. `ModelApiCostResolution` at line 39 declares `'480p'`, `'720p'`, `'1080p'` explicitly plus an index signature, so adding `'4k'` will not raise a type error and a casing typo will not either.
- `contentstudio-frontend/src/composables/useVideoGeneration.ts` holds the Seedance model definitions and their `supportedResolutions` used by the selector.

**The cost figure:**

The registry documents Seedance's own pricing formula in the comment above `seedance-2.0-fast`: `(h * w * duration * 24) / 1024` tokens, at a per-1000-token rate. Solving for the standard tier rate from the known 1080p price gives $0.014 per 1000 tokens exactly (1920 x 1080 gives 48,600 tokens/s, and $0.6804 / 48.6 = $0.014). It checks out at 480p too. Applying it to 4K at 3840 x 2160 gives 194,400 tokens/s and **$2.7216 per second**, exactly 4x the 1080p rate.

Treat that as a cross-check, not the source of truth. Confirm fal's published 4K rate, and confirm 4K actually means 3840 x 2160, since fal's schema only says `4k`.

**Casing:** fal's enum value is lowercase `4k`. Seedance's other keys are lowercase so it fits, but this is per-model rather than global: MiniMax H3 uses `2K` and `768P`. The key must match fal exactly, for the reason below.

**The gotcha that makes the cost entry non-optional.** `useVideoCreditCalculation.ts` around line 388 resolves the per-second cost like this:

```
} else if (apiCost[resolution] !== undefined) {
    costPerSecond = apiCost[resolution] as number
} else {
    // Fixed pricing (single resolution model)
    const availableResolution = Object.keys(apiCost)[0]
    costPerSecond = apiCost[availableResolution] as number
}
```

A resolution with no entry in `api_cost` does not error. It falls back to whichever key is first in the object, which for `seedance-2.0` is `480p` at $0.1346 per second. So a missing or mis-cased `4k` key charges 480p credits for a 4K generation, silently and with nothing in the logs. Against a real 4K rate of roughly $2.72 per second that is about a twentyfold undercharge. The type system will not catch it either, because `ModelApiCostResolution` has a permissive index signature. Verify the charged credits on a real 4K generation before release rather than trusting the config to have been picked up.

**Also touched:** `src/utils/video/cost_calculator.py` and `src/utils/credit_validation/validation.py` on the AI agents side, and `src/modules/AI-tools/components/MediaGenerationOptions.vue` for the selector.

**Unrelated drift found nearby, worth a separate look:** `useVideoCreditCalculation.ts` prices `seedance-2.0-fast` at `1080p: 0.5443`, while the model registry lists that model's resolutions as `["480p", "720p"]` with a comment saying 1080p was dropped and confirmed. Harmless today because the option cannot be selected, but the two lists disagree.
