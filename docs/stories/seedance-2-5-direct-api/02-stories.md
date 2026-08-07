# Stories: Seedance 2.5 via direct API

One story.

No design story. Seedance 2.5 appears in the model picker and uses the duration, resolution, aspect ratio and attachment controls exactly as every other video model does, so there is nothing new to design.

---

## [Full Stack] Add Seedance 2.5 as a video model over its direct API

### Description

Add ByteDance's Seedance 2.5 as a selectable video model. We have direct API access through BytePlus ModelArk rather than going through fal, which is what makes this more than a configuration change: every video model in the product today runs through fal, and the video pipeline has no way to talk to anything else.

For the user, nothing about this is unusual. Seedance 2.5 shows up in the model list alongside the others, with its own durations, resolutions and aspect ratios, and generating a video works exactly as it does for any other model. That is the point. The provider is an implementation detail and should stay one.

Two things make this different behind the scenes. ModelArk bills video by tokens rather than by the second, while our whole cost model is per-second USD, so the rates have to be translated rather than copied. And the account needs an active $30 resource package before Seedance 2.5 can be activated at all, which gates testing as much as it gates release.

### Workflow

1. The user opens video generation and picks Seedance 2.5 from the model list, which looks and behaves like every other entry.
2. The user writes a prompt and optionally attaches a starting image.
3. The user sets duration, resolution and aspect ratio from this model's own options.
4. The credit estimate updates before they generate, as it does for every model.
5. The user generates. Progress is reported the same way, and the finished video arrives in the same place.
6. If the user cancels partway, it behaves consistently with cancelling any other generation.

### UI copy

- Model name in the picker: `Seedance 2.5`
- Model description: `ByteDance's video model, built for long single-shot clips.` Adjust once the confirmed maximum duration and headline capability are known from the official documentation.
- Duration, resolution and aspect ratio options come from the model's own supported values, using the existing controls with no new labels.
- Audio indicator, only if the model generates audio: reuse the existing always-on-audio treatment and tooltip pattern from the other models that behave that way.
- Nothing else is new. Errors, progress, cancellation and insufficient-credit messaging all reuse the existing copy.

### Scope

**In scope**

- A path for the video pipeline to call a provider that is not fal, covering submit, status, result retrieval, cancellation and media probing
- Seedance 2.5 registered with its real options and rates
- Text-to-video and image-to-video
- Correct credit estimation and charging, including its per-mode rates
- Frontend exposure in the model picker with its own duration, resolution and aspect ratio options

**Out of scope, deliberately**

- **Reference-to-video.** Seedance 2.5 accepts reference images, videos and audio. The mechanism for reference video and audio inputs is being built in **[Full Stack] Add reference video and audio inputs to MiniMax H3**. Seedance 2.5 should consume that mechanism once it exists rather than build a second one. Add it as a follow-up.
- **Any relaxed content-safety setting**, if the model exposes one. That is a content-policy and legal decision, not a wiring task. Leave whatever the provider's default safety setting is.
- **Any optional billed extras**, such as search or enrichment charged separately from generation. Leave them off until there is a reason to turn one on, so the credit model stays one rate per generation.
- **Camera control, long-video or editing features**, if present. None map to anything in the product today.

The last three are written conditionally on purpose. They were flagged from unofficial coverage, and whether they exist on the surface we actually have is one of the things to establish from the official documentation. If any of them turn out not to exist, delete the line rather than build it.

### Acceptance criteria

**The model works**

- [ ] Seedance 2.5 appears in the video model picker and can be selected like any other model
- [ ] Text-to-video works from a prompt
- [ ] Image-to-video works from a prompt plus an image
- [ ] The durations, resolutions and aspect ratios offered match what the model actually supports, taken from the official BytePlus ModelArk documentation rather than from third-party sources
- [ ] Durations longer than any existing model are selectable and generate correctly, and nothing in the duration selector or validation assumes a 15 or 20 second ceiling
- [ ] If the model generates audio, that is reflected in the interface consistently with the other always-on-audio models
- [ ] The generated video is delivered, stored and made available for use in a post exactly as videos from fal models are

**The provider path**

- [ ] Video generation can submit to a provider that is not fal, without changing behavior for any existing fal model
- [ ] Job status and progress are reported to the user in the same way as for fal models, so the interface cannot tell the difference
- [ ] Cancelling an in-flight Seedance 2.5 generation behaves consistently with cancelling any other generation, using ModelArk's own cancel operation
- [ ] What a cancelled task does to billing is established from the documentation and reflected in what the user is charged
- [ ] A provider error, timeout, or rate limit surfaces to the user as a clear message rather than a raw API error, and is reported to error tracking
- [ ] Credentials for the direct API are configured the same way as other provider credentials and are never logged
- [ ] Every existing fal model continues to submit, poll, cancel and return identically, verified rather than assumed

**Credits**

- [ ] ModelArk's token-based rate is translated into the per-second-by-resolution shape our cost model uses, and the derivation is recorded so it can be rechecked when the rate changes
- [ ] The credit estimate shown before generating reflects the correct rate for the chosen mode and resolution
- [ ] Credit charges are verified against the provider's actual billed usage for real generations, not assumed from configuration. Token billing means a wrong translation is invisible until the invoice arrives
- [ ] A generation at the model's maximum duration and highest resolution is estimated and charged correctly
- [ ] Insufficient credits are caught before the generation starts
- [ ] No other model's rates change

### Mock-ups

N/A. Uses the existing model picker and generation controls.

### Impact on existing data

- No schema changes. A new model definition, its cost entries, and new provider credentials.
- Generation records may need to distinguish which provider produced them, if the existing record assumes a fal request id. Worth checking, since cancellation and status both key off fal identifiers today.
- Existing generations and credit history are untouched.

### Impact on other products

- Mobile apps: no impact. AI video generation is web only.
- Chrome extension: no impact.
- White-label domains: no impact.

### Dependencies

- **Account prerequisite, blocking:** Seedance 2.5 can only be activated on an account holding an available $30 resource package. Packages already bought for the Seedance 2.0 series count. This gates development and testing, not just release, so it needs to be in place before the story starts.
- **Before build, blocking:** the option matrix read out of the official BytePlus ModelArk documentation. The docs site is fully client-side rendered and the Lark documents are permissioned, so this needs a person with access rather than a link in a ticket. Specifically: exact model IDs, endpoint and auth scheme, supported durations, resolutions and aspect ratios with their real parameter names, whether audio is generated, the token rate per resolution so per-second equivalents can be derived, the cancel semantics and their billing consequence, and the rate and concurrency limits fal used to absorb on our behalf.
- **Coordinate with** `docs/stories/flux-3-video-models/` and `docs/stories/video-model-reference-audio-video-inputs/`. All three touch `model_registry.py`, `useVideoGeneration.ts` and `useVideoCreditCalculation.ts` and will conflict if run at the same time.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness tested
- [ ] Multilingual support verified (translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

**The video path has no provider seam. This is the bulk of the work.**

Every video model in the registry is fal, and `provider` is metadata for video rather than something that selects behavior. The only place it drives anything is `src/agents/image/models_registry.py:18`, which filters image models to `provider == "fal"`. Registry counts today are 47 fal, 4 openai, 2 anthropic, 2 groq, 1 google, and every non-fal entry is a text or image model.

The hardwired points:

- `src/agents/video/video_generator.py:23` imports `fal_client` at module level
- `src/agents/video/video_generator.py:96` calls `fal_client.submit(model_endpoint, arguments=parameters)` then blocks on `handler.get()`
- `src/agents/video/video_generator.py:107-118` reads `handler.cancel_url` and `handler.request_id` and stores them through `store_fal_cancel_url` and `store_fal_request_id`
- `src/jobs/cancellation.py:235-262` calls `fal_client.status_async` and `fal_client.cancel_async`
- `src/jobs/cancellation.py:381-387` maps fal's `InProgress`, `Completed` and `InQueue` objects onto our status strings
- `src/utils/video/media_probe.py:7` uses `ensure_fal_credentials`

`src/agents/image/providers/` is the closest existing example of a provider module and is worth looking at before inventing a structure, though it serves images rather than the async video job flow.

**Async shape, and cancellation exists.** ModelArk's video generation API documents four operations: create a task, retrieve a task, list tasks, and cancel or delete a task. That maps cleanly onto the existing Dramatiq flow in `src/jobs/`, which publishes progress to Redis and lifecycle events to Kafka via `src/events/publisher.py`. Polling retrieve keeps the current shape and is the cheaper first integration. If ModelArk also offers callbacks, using them would mean standing up a public endpoint and reconciling it with that flow, which is more work than a first integration needs.

**Pricing is token-based, and that is the interesting part.** ModelArk quotes video models in USD per million output tokens, not per second. For scale, the Seedance 1.0 family lists at 2.5 USD per million tokens for Pro, 1.0 for Pro Fast and 1.8 for Lite.

`api_cost` in `src/utils/model_registry.py` is keyed by resolution at model level, and `VIDEO_PRICING_MATRIX` in `contentstudio-frontend/src/composables/useVideoCreditCalculation.ts` mirrors it. Both are per-second USD. So a token rate has to be converted into per-second-by-resolution figures.

The conversion is already precedented in our own code. The registry comment above `seedance-2.0-fast` records Seedance's token formula, `(h * w * duration * 24) / 1024` tokens, and the existing fal per-second rates were derived from exactly that. Reusing the same formula with ModelArk's token rate keeps the two consistent and gives a derivation that can be rechecked when the rate moves. Record the working next to the numbers.

**The undercharge trap.** `useVideoCreditCalculation.ts` around line 388 falls back to the first key in `api_cost` when the requested resolution has no entry, silently and with nothing logged. Documented in `docs/stories/seedance-2-4k-resolution/`. With token billing the stakes are higher, because a bad conversion does not fail loudly anywhere. Reconcile charged credits against the provider's actual billed usage, not just against our own estimate.

**Do not build from the research doc's option matrix.** Earlier third-party figures were removed from it deliberately, because plausible unofficial numbers sitting beside official links is how wrong values get shipped. Two of those sources already disagreed on the resolution ladder. The options need reading out of the official documentation.

**A separate opportunity, not this story.** We buy Seedance 2.0 Fast through fal at $0.2419 per second at 720p. BytePlus direct is about $0.09 per second under the discount running to 7 September 2026, implying roughly $0.12 at list, which is around half the fal rate for very likely the same model. Seedance 2.0 mini, which we do not carry at all, is around $0.032 per second discounted. Once the provider seam from this story exists, moving the existing Seedance models across is worth its own ticket. Confirm the comparison is like for like first, since platform, quotas and output specs may differ.
