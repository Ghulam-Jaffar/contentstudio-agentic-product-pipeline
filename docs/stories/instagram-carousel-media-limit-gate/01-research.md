# Research — Block publishing an Instagram carousel with more than 10 media

## Feature request
Instagram allows a maximum of **10 media** in a carousel. Today ContentStudio lets the user post anyway and **silently drops the extras**. Instead, the **Post Now / Schedule** action should be **disabled** until the user reduces the media to 10, and the **bottom error banner** should tell them to "reduce the number of media to 10."

## Current State
- **Silent slicing to 10:** at publish/hydration, Instagram images are silently capped — `useComposerSharing.ts:342` and `components/main-composer/useMediaEditing.ts:220`: `images.length > 10 ? images.slice(0, 10) : images` (tests: "caps instagram/threads/telegram images at 10"). This is why the last items are dropped.
- **Over-limit is only a WARNING today (non-blocking):** `useComposerValidationSurface.ts` (~lines 751–800) pushes a `max_images_warning` when `totalImages > maxImages` (max from `channelConfig.warnings.image.max_images`). Warnings are **advisory and never gate publish**.
- **How gating actually works:** blocking errors live in `socialPostErrors` (computed in `useComposerValidationSurface.ts` ~965). The publish/schedule button disable flag (`disableScheduleButton`, ~944–953) is driven by `socialPostErrors.value.length > 0`. `disableScheduleButton` / `forceDisableButton` are draft-state fields (`useComposerDraftState.ts`) the footer CTAs consume. So: **put the over-limit check into `socialPostErrors` (an error, not a warning) and the button auto-disables + the message shows in the footer/bottom banner.**
- The per-platform media-limit rules live in `features/validation/mediaLimits.ts`; the current Instagram cap is **10** (matches the request; the "20" in the ask is a typo — confirmed by code + "drops the last 2").

## What Needs to Change (all FE, composer)
- When a selected **Instagram** account's post has **more than 10 media**, raise a **blocking error** (into `socialPostErrors`) instead of the current advisory warning — so the **Post Now / Schedule / Add to Queue** button becomes **disabled**.
- The **bottom error banner** shows a clear message: *"Instagram allows up to 10 media per post. Please reduce the number of media to 10."*
- Once the user reduces to ≤ 10, the error clears and the publish actions re-enable.
- With the gate in place, the silent slice-to-10 no longer masks the problem (the user can't reach publish with > 10).

## Mobile Context
Web composer only (SocialModal). The mobile apps' composers are not covered here; if they have the same silent-drop behavior it's a separate ticket.

## Files Involved
- `src/modules/composer/views/social-modal/useComposerValidationSurface.ts` — move the Instagram over-limit from the `max_images_warning` bucket into `socialPostErrors` (blocking); this auto-feeds `disableScheduleButton`.
- `src/modules/composer/features/validation/mediaLimits.ts` — Instagram media-count limit (10) / where an error rule fits.
- `src/modules/composer/views/social-modal/useComposerDraftState.ts` — `disableScheduleButton` / `forceDisableButton` flags the CTAs read.
- `src/modules/composer/views/social-modal/useComposerSharing.ts` (:342) & `components/main-composer/useMediaEditing.ts` (:220) — the current silent `slice(0, 10)`.
- `src/locales/*/composer.json` — the new/updated banner + disabled-button tooltip copy (all 8 locales).

## Note
Single **[FE]** story. The number (10) should come from the existing configured Instagram max rather than a hardcode, so it stays correct if Instagram's cap changes.
