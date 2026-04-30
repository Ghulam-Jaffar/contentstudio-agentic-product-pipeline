# iOS AI Studio response action placement

## Current State
- In the iOS AI Studio response card, the `Copy` and `Replace` buttons are rendered below the follow-up action pills.
- The buttons appear separated from the text response and feel disconnected from the response card itself.
- The current layout makes the action buttons look like they belong to the follow-up pills area rather than the generated text block.

## What needs to change
- Move the `Copy` and `Replace` buttons into the footer area of the AI text response card.
- Keep the follow-up action pills above the buttons, with a clear visual separator between the pill area and the card footer actions.
- Keep the buttons aligned to the bottom-right of the response card and maintain a consistent margin from the card edge.
- Preserve tappable area and spacing to avoid accidental taps when the text response is long.

## Mobile Context
- This is an iOS-only UI refinement for AI Studio response cards in the mobile app.
- No backend changes are required.

## Files involved
- `contentstudio-ios-v2/ContentStudio/Controllers/Composer/View/Composer.storyboard`
- `contentstudio-ios-v2/ContentStudio/Controllers/Nav Menu VCs/Composer/ComposerViewController.swift`
- `contentstudio-ios-v2/ContentStudio/Controllers/Composer/View/ComposerVC.swift`
- Any custom response cell or view used by the AI Studio text response card in the iOS composer flow
