# Stories: Content Category Validation in Post Composer

---

## Story 1

**Title:** `[FE] Show inline validation errors when "Add to Content Category" is selected without a category`

---

### Description:

As a content creator using the Post Composer, I want clear, in-context error feedback when I select "Add to Content Category" in the Posting Schedule but haven't chosen a category yet, so that I immediately know what's missing and where to fix it — without having to hover over a small icon to discover the problem.

This story replaces the existing tooltip-based error icon (shown to the right of the "Posting Schedule" heading in `MainComposer.vue`) with a more visible, inline error banner. It also adds a red border to the Content Category dropdown in the sidebar, a clickable error pill in the footer bar, and disables the "Add to Category" action button until a category is selected.

---

### Workflow:

1. User opens the Post Composer (social modal).
2. User clicks "Add to Content Category" in the Posting Schedule section — the radio button with value `content_category` in `PostingSchedule.vue`.
3. User has **not** selected a category from the Content Category dropdown at the top of the left-hand sidebar (`AccountSelectionAside.vue`).
4. The system immediately shows the following validation feedback:
   - **Sidebar dropdown:** The Content Category `CstDropdown` gets a `border border-red-500 rounded` style, making its outline visibly red.
   - **Inline error banner:** An `Alert` (variant `danger`) appears directly below the `<PostingSchedule>` component inside the collapsible, spanning the full width of the card. The banner reads: _"No content category selected. Please pick one from the top-left dropdown to continue."_
   - **Footer error pill:** A red pill badge appears to the left of the "Send for approval" button in the footer bar. It shows `! (1)`. Clicking this pill opens a `Modal` with the error details.
   - **"Add to Category" button:** The main action button in the footer (which reads "Add to Category" when `time_type === 'content_category'`) becomes visually dimmed (`opacity-50 cursor-not-allowed`) and non-clickable.
5. User clicks the red error pill in the footer → a modal opens titled **"Action Required"** with the message: _"No content category selected. Please pick one from the top-left dropdown to continue."_ and a single CTA: **"Got it"** (closes the modal).
6. User selects a category from the top-left dropdown in the sidebar.
7. All error states clear automatically:
   - Red border on the dropdown disappears.
   - Inline error banner disappears.
   - Footer error pill disappears.
   - "Add to Category" button becomes active and clickable again.

---

### Acceptance criteria:

- [ ] When `time_type` is `content_category` and `selectedContentCategory` is `null`, the Content Category `CstDropdown` in the sidebar displays a red border (`border border-red-500`) around it.
- [ ] When `time_type` is `content_category` and `selectedContentCategory` is `null`, a red inline `Alert` banner appears immediately below the Posting Schedule collapsible content (below `<PostingSchedule>`), full-width, with the message: _"No content category selected. Please pick one from the top-left dropdown to continue."_
- [ ] The existing error icon tooltip (the `v-menu` with `error-icon.svg`) shown to the right of the "Posting Schedule" heading is removed entirely. It is replaced by the new inline banner described above.
- [ ] When `time_type` is `content_category` and `selectedContentCategory` is `null`, a red error pill badge reading `! (1)` appears in the footer bar, to the left of the "Send for approval" button.
- [ ] Clicking the red error pill opens a modal titled **"Action Required"** with the message: _"No content category selected. Please pick one from the top-left dropdown to continue."_ The modal has a single button: **"Got it"** that closes it.
- [ ] When `time_type` is `content_category` and `selectedContentCategory` is `null`, the "Add to Category" action button in the footer has `opacity-50 cursor-not-allowed` styling and is non-clickable (`:disabled="true"`).
- [ ] All four error states (red border, inline banner, footer pill, disabled button) clear automatically the moment the user picks a category from the sidebar dropdown.
- [ ] When a category is already selected before the user clicks "Add to Content Category", none of the error states appear.
- [ ] When the user switches away from `content_category` to another posting option (e.g., "Schedule", "Draft"), all error states are hidden.
- [ ] The error pill does not appear when `time_type` is not `content_category`.

---

### UI Copy

**Inline error banner (below Posting Schedule card):**
> No content category selected. Please pick one from the top-left dropdown to continue.

**Footer error pill label:**
> ! (1)

**Error modal:**
- **Title:** Action Required
- **Body:** No content category selected. Please pick one from the top-left dropdown to continue.
- **CTA button:** Got it

**Tooltip on the footer error pill (hover):**
> 1 error in Posting Schedule

---

### Component Usage

- **Inline error banner:** Use `Alert` from `@contentstudio/ui` with `type="danger"` (or `variant="danger"` — check component props). Set `:show-close-icon="false"`. Full-width, placed inside the `CstCollapsible` below `<PostingSchedule>`.
- **Footer error pill:** Use `Badge` from `@contentstudio/ui` with a danger/red variant (check props). If `Badge` does not support a `danger` variant with custom styling, use a Tailwind inline element: `<span class="flex items-center gap-1 bg-red-50 border border-red-300 text-red-600 rounded-full px-2.5 py-1 text-xs font-medium cursor-pointer">! (1)</span>`. Flag this gap clearly during implementation.
- **Error modal:** Use `Modal` from `@contentstudio/ui`. Single "Got it" (`Button` variant `primary`) to close.
- **Red border on sidebar dropdown:** Apply conditional Tailwind classes to the wrapping `<div>` of the `CstDropdown` in `AccountSelectionAside.vue`: `class="w-full"` → `:class="{ 'ring-1 ring-red-500 rounded': isCategoryError }"`. Do not override `CstDropdown` internals.

---

### Mock-ups:

N/A — behavior is fully specified in the Workflow and UI Copy sections above.

---

### Impact on existing data:

None. This is a purely visual/interaction change. No data model or API changes are involved.

---

### Impact on other products:

- **Mobile apps (iOS/Android):** The Post Composer is a web-only feature. No mobile impact.
- **Chrome extension:** The Composer is not available in the Chrome extension. No impact.
- **White-label:** All colors use theme-safe classes (`red-500`, `red-600`, `red-50`, `red-300`) — these are neutral error colors, not primary brand colors, so they are white-label safe.

---

### Dependencies:

None.

---

### Implementation notes (for developers):

**Prop flow for the error state:**

The error condition (`time_type === 'content_category' && selectedContentCategory === null`) is already computed as `postingScheduleErrors` in `SocialModal.vue`. The changes required:

1. **`SocialModal.vue`:** In the `footerData` computed property (~line 2970), add `postingScheduleErrors: this.postingScheduleErrors` so it's available to `MainComposerFooter`.

2. **`MainComposer.vue`:**
   - Remove the `v-menu` block (~lines 1279–1304) that shows the tooltip error icon next to the heading.
   - Add `<Alert>` component below `<PostingSchedule>` inside the `CstCollapsible` slot, conditionally shown when `postingScheduleErrors.length`.
   - Pass a new prop `is-category-error` (boolean) to `AccountSelectionAside` based on `postingScheduleErrors.length > 0`.

3. **`AccountSelectionAside.vue`:** Accept a new `isCategoryError` boolean prop. Apply conditional ring/border styling to the `CstDropdown` wrapper `<div>` when `isCategoryError` is true.

4. **`MainComposerFooter.vue`:**
   - Accept `postingScheduleErrors` as a new prop (Array, default `[]`).
   - Compute `hasCategoryError` = `postingScheduleErrors.length > 0 && publishTimeOptions.time_type === 'content_category'`.
   - Show the error pill `<span>` to the left of the "Send for approval" button when `hasCategoryError` is true; bind click to open the modal.
   - Add the `Modal` for the error details.
   - In the button class binding for the main action button, add `hasCategoryError` to the `opacity-50 cursor-not-allowed` condition and add it to `:disabled` as well.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
