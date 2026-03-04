# Epic & Stories: Bulk Schedule — Image Upload Mode

---

## Epic

**Title:** Bulk Schedule — Image Upload Mode
**Description:**
Add a second mode to Bulk Schedule that lets users upload images directly instead of preparing a CSV spreadsheet. Users drag and drop up to 100 images, name their campaign and select accounts (same as the existing CSV flow), set a schedule, and on the final step generate AI captions for all posts at once or one at a time.

The CSV/XLSX mode is completely unchanged. Both modes share the same 4-step wizard shell, and reuse the existing Name & Accounts (Step 2) and Scheduling (Step 3) components. The image mode adds a new image upload Step 1 and a caption-generation-powered Step 4.

Entry points are updated in both the empty state (two-card mode selector) and the listing state (dropdown button in the top-right replacing the single CSV upload CTA).

---

## Story 1: [BE] Extend Bulk Schedule API for Image Upload mode and AI caption generation

### Description:
As a ContentStudio user scheduling posts via the Image Upload mode of Bulk Schedule, I want the system to accept my uploaded images, turn each one into a post in my draft, and generate AI captions for them on demand — so that the frontend can drive a seamless image-to-scheduled-posts workflow.

---

### Workflow:

1. User starts the Image Upload wizard — frontend calls the draft save endpoint with image files instead of a CSV
2. Backend validates each file (type: JPG/PNG/GIF/WebP; size: max 10 MB); rejects invalid files with a per-file error in the response
3. Valid images are uploaded to GCS (same `MediaLibrary::uploadOriginalsToGCSFromFile` pattern as CSV files, directory: `/bulk_images`)
4. GCS URLs and filenames stored in the `csv_processing` MongoDB document under a new `image_mode: true` flag and `images` array
5. When the user reaches Step 4, frontend fetches the post list — backend returns one post per image, each with status `pending_caption` and the image URL
6. When the user triggers caption generation (for one post or all), frontend calls the caption generation endpoint
7. Backend calls the AI caption service, passing the image URL and workspace Brand Settings (tone, industry, etc.)
8. Caption returned to frontend; backend deducts 1 AI text credit from the workspace
9. If credits are exhausted, endpoint returns a `402` with a translatable error key — frontend shows the appropriate toast
10. Caption is saved to the post's draft record on the backend when the user proceeds to scheduling

---

### Acceptance criteria:

- [ ] `POST /saveDraftCsvAutomation` extended to accept `image_mode: true` and an `images[]` array of image files (multipart)
- [ ] Each uploaded image validated: accepted types JPG/PNG/GIF/WebP; max 10 MB per file; invalid files return per-file error in response without failing the whole request
- [ ] Valid images uploaded to GCS under `/bulk_images/{workspace_id}/` and URLs stored in the draft document
- [ ] Draft document includes `image_mode: true` flag and `images` array (URL, filename, original_name, size, mime_type per image)
- [ ] `POST /fetchCsvPosts` (or equivalent) returns posts for image-mode drafts: one post per image, each with `id`, `image_url`, `caption` (null if not yet generated), `status` (`pending_caption` | `caption_ready`)
- [ ] New endpoint `POST /generateBulkCaption` accepts: `draft_id`, `post_id`, `workspace_id` — calls AI caption service with the post's image URL and workspace Brand Settings
- [ ] Caption generation deducts AI text credits after a successful response (not before)
- [ ] If workspace has 0 AI text credits, endpoint returns HTTP 402 with translatable error key `errors.no_ai_credits`
- [ ] If AI service fails (timeout, error), endpoint returns HTTP 500 with translatable error key `errors.caption_generation_failed`
- [ ] Caption is saved to the post's draft record when received; subsequent calls to `fetchCsvPosts` reflect the updated caption and `status: caption_ready`
- [ ] Existing CSV mode draft save endpoint is unchanged — `image_mode` flag defaults to `false`, backward compatible
- [ ] `POST /processCsv` extended to handle image-mode drafts: skips CSV parsing, builds posts directly from the stored `images` array
- [ ] All new API error responses use translatable message keys, not hardcoded English strings

---

### Mock-ups:
N/A — backend story

---

### Impact on existing data:

- `csv_processing` MongoDB documents gain two new optional fields: `image_mode` (boolean, default false) and `images` (array of image objects)
- No changes to existing CSV-mode documents or processing logic
- New GCS directory: `/bulk_images/{workspace_id}/`

---

### Impact on other products:

- Web only — no mobile impact
- AI credit deduction integrates with the existing credit system; no new billing tables

---

### Dependencies:

- Existing `MediaLibrary::uploadOriginalsToGCSFromFile` — reused for GCS upload
- Existing AI caption generation service (same service used in Composer)
- Existing AI credit tracking and deduction system

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend story
- [ ] Multilingual support — all API error messages use translatable keys in all 7 locales
- [ ] UI theming support — N/A, backend story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

## Story 2: [FE] Update Bulk Schedule entry points with mode selector (empty state + listing state)

### Description:
As a user navigating to Bulk Schedule, I want to clearly see that I can either upload a CSV or upload images directly, so I can choose the right method for my current task without confusion.

---

### Workflow:

**Empty state (no existing bulk schedules):**
1. User navigates to Publisher → Automations → Bulk Schedule
2. Page shows a full-page empty state with two side-by-side cards
3. Left card: "Schedule via CSV / XLSX" — for spreadsheet-based scheduling
4. Right card: "Schedule via Images" — for direct image upload with AI captions; has "NEW · AI-Powered" badge
5. User clicks a card CTA → the respective wizard launches inline on the page
6. "Learn more about Bulk Scheduling" text link retained below both cards

**Listing state (has existing bulk schedules):**
1. User navigates to Bulk Schedule — sees the existing listing table of past bulk schedules
2. Top-right button changes from "Upload CSV File" to "+ Bulk Schedule"
3. User clicks "+ Bulk Schedule" → a small dropdown menu opens with two options:
   - "Schedule via CSV / XLSX" — sub-text: "Schedule up to 500 posts from a spreadsheet"
   - "Schedule via Images" — sub-text: "Upload images and generate captions with AI — up to 100 posts" + small "NEW" badge
4. User clicks an option → that wizard launches inline, replacing the listing table
5. Clicking × in the wizard returns the user to the listing table

---

### Acceptance criteria:

**Empty state:**
- [ ] Page heading: "Get started with Bulk Scheduling"
- [ ] Sub-heading: "Choose how you'd like to create and schedule your posts"
- [ ] Two cards rendered side by side; equal width
- [ ] CSV card: icon FileSpreadsheet (Lucide), title "Schedule via CSV / XLSX", description "Schedule up to 500 posts by uploading a spreadsheet with dates, captions, and media links.", primary CTA "Upload CSV File", secondary CTA "Download Sample CSV"
- [ ] Image card: icon Image (Lucide), title "Schedule via Images", badge "NEW · AI-Powered" (`bg-primary-cs-500` gradient, white text), description "Upload up to 100 images and let AI generate captions instantly — then schedule in bulk.", primary CTA "Upload Images", no secondary CTA
- [ ] "Learn more about Bulk Scheduling" text link displayed below both cards
- [ ] Clicking "Upload CSV File" launches the CSV wizard (existing behaviour)
- [ ] Clicking "Upload Images" launches the Image Upload wizard

**Listing state:**
- [ ] Existing "Upload CSV File" button replaced by "+ Bulk Schedule" button (primary, top-right)
- [ ] Clicking "+ Bulk Schedule" opens a dropdown with two options
- [ ] Option 1: icon FileSpreadsheet (Lucide), main text "Schedule via CSV / XLSX", sub-text "Schedule up to 500 posts from a spreadsheet"
- [ ] Option 2: icon Image (Lucide), main text "Schedule via Images", "NEW" badge (small, `bg-primary-cs-500`), sub-text "Upload images and generate captions with AI — up to 100 posts"
- [ ] Clicking option 1 launches the CSV wizard inline
- [ ] Clicking option 2 launches the Image wizard inline
- [ ] Dropdown closes on outside click or option selection
- [ ] Wizard × button returns user to the listing table when launched from listing state
- [ ] Wizard × button returns user to empty state when launched from empty state

**Both states:**
- [ ] All new strings use i18n keys; translations added for all 7 locales (en, fr, de, es, it, el, zh)
- [ ] No hardcoded hex values — all colors via `primary-cs-*` and `gray-*` Tailwind classes

---

### Mock-ups:
N/A — no designs provided; implementation follows PRD §8.1.

---

### Impact on existing data:
None — UI change only. CSV wizard behaviour unchanged.

---

### Impact on other products:
Web only.

---

### Dependencies:
- **[BE] Extend Bulk Schedule API for Image Upload mode and AI caption generation** — needed before the image wizard can save drafts
- **[FE] Build Image Upload wizard — Step 1, mode routing, and Step 3 adaptation** — the wizard launched from these entry points

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — Bulk Schedule is desktop-only; N/A
- [ ] Multilingual support — all new strings in all 7 locale files
- [ ] UI theming support — all colors via `primary-cs-*` CSS variable classes; no hardcoded hex
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

## Story 3: [FE] Build Image Upload wizard — Step 1, mode routing, and Step 3 adaptation

### Description:
As a user choosing the Image Upload mode in Bulk Schedule, I want to drag and drop my images into the wizard, set my campaign name and accounts, configure my schedule, and proceed to the finalization step — so that the wizard feels fast and familiar while correctly handling images instead of a spreadsheet.

---

### Workflow:

1. User clicks "Schedule via Images" from either entry point — Image wizard launches
2. **Step 1 — Upload Images:**
   - User sees a dashed drop zone with instruction copy and a drag-and-drop area
   - User drags images onto the zone or clicks "Upload" to open file picker
   - Thumbnails appear in a preview grid as images are added; counter updates ("5 / 100 images selected")
   - User can remove individual images (× on each thumbnail) or click "Remove all"
   - Oversized files (> 10 MB) are skipped with a per-file toast
   - User clicks "Next" (disabled until ≥1 image is present)
3. **Step 2 — Name & Accounts:** Identical to the CSV wizard — unchanged
4. **Step 3 — Scheduling:**
   - Same as CSV wizard except the "Use times from CSV" tab is not shown
   - Default tab: "Regular Intervals"
   - Warning banner shown at the bottom before user clicks "Next"
   - Clicking "Next" triggers post processing (same as CSV mode); user cannot go back after this point

---

### Acceptance criteria:

**Step 1 — Upload Images:**
- [ ] Wizard header: "Bulk Schedule via Images" + "AI-Powered" badge (`bg-primary-cs-500` gradient, white text)
- [ ] Step indicator shows 4 steps: "Upload Images", "Name & Accounts", "Scheduling", "Finalizing Posts" — Step 1 active
- [ ] Instruction line 1: "Upload up to 100 images. Each image becomes a separate post. AI captions will be generated in the last step."
- [ ] Instruction line 2 (grey, smaller): "Supported: JPG, PNG, GIF, WebP · Max file size: 10 MB"
- [ ] Drop zone: dashed `border-primary-cs-200` border, centered image icon (Lucide Image), label "Drag & drop images here, or Upload" — "Upload" is a `text-primary-cs-500` hyperlink-styled trigger
- [ ] Drag-over state: border becomes `border-primary-cs-500`, background becomes `bg-primary-cs-50`
- [ ] File picker accepts: `image/jpeg, image/png, image/gif, image/webp`; multiple selection enabled
- [ ] Files > 10 MB are skipped; per-file toast: "'{filename}' is too large. Max file size is 10 MB." — warn style, 5s
- [ ] If user selects more than 100 images, first 100 are kept silently; counter shows "100 / 100 images selected"
- [ ] Accepted images appear in an auto-fill thumbnail grid (min 72px cells, `object-fit: cover`, `border-radius: 8px`, `border-gray-200`)
- [ ] Each thumbnail has an × remove button in the top-right corner; clicking removes that image from the selection
- [ ] Counter above grid: "{N} / 100 images selected" — bold number, grey label
- [ ] "Remove all" link (`text-red-500`) shown top-right of grid; clicking removes all images and resets the drop zone
- [ ] Grid has a max height of ~200px with vertical scroll when images overflow
- [ ] "Next" button disabled (grey) until at least 1 image is uploaded; enabled once images are present
- [ ] Clicking × on the wizard returns to the landing page / listing (whichever launched it)

**Step 2 — Name & Accounts:**
- [ ] Identical to CSV wizard — all existing behaviour and validation unchanged
- [ ] Step indicator shows Step 2 active

**Step 3 — Scheduling:**
- [ ] Scheduling tabs shown: "Regular Intervals", "Content Category", "Add to Queue" — "Use times from CSV" tab NOT shown
- [ ] Default active tab: "Regular Intervals"
- [ ] All scheduling options (Regular Intervals dropdowns, Posting Status, Content Category, Add to Queue) behave identically to CSV mode
- [ ] "Regular Intervals" label: "Post Every"; dropdowns: numeric (1, 2, 3, 4, 6, 8, 12, 24) and unit (Hour(s), Day(s), Week(s))
- [ ] Posting Status radio buttons: "Schedule" (default selected) and "Save as draft"
- [ ] Warning banner shown at the bottom of Step 3: "Note: Review your settings before proceeding. You won't be able to go back to previous steps after this." — amber/yellow (`bg-yellow-50`, `border-yellow-200`)
- [ ] "Next" button always enabled on Step 3; clicking triggers post processing and navigates to Step 4
- [ ] "Previous" button on Step 3 navigates back to Step 2
- [ ] After clicking "Next" on Step 3, the "Previous" button is not available on Step 4 (processing has begun)
- [ ] Step indicator shows Step 3 active
- [ ] All new strings use i18n keys; translations for all 7 locales

---

### Mock-ups:
N/A — no designs provided; implementation follows PRD §8.2–§8.4.

---

### Impact on existing data:
None — CSV mode wizard is unchanged. Image mode creates new draft documents with `image_mode: true` flag (see **[BE] Extend Bulk Schedule API for Image Upload mode and AI caption generation**).

---

### Impact on other products:
Web only. No mobile impact.

---

### Dependencies:
- **[BE] Extend Bulk Schedule API for Image Upload mode and AI caption generation** — Step 1 "Next" calls the image draft save endpoint
- **[FE] Update Bulk Schedule entry points with mode selector (empty state + listing state)** — provides the entry points that launch this wizard

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — Bulk Schedule wizard is desktop-only; N/A
- [ ] Multilingual support — all new strings in all 7 locale files
- [ ] UI theming support — all colors via `primary-cs-*` CSS variable classes; no hardcoded hex
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

## Story 4: [FE] Build Finalizing Posts step with AI caption generation for Bulk Schedule Image mode

### Description:
As a user who has uploaded images in the Bulk Schedule wizard, I want to see all my posts in a review table and generate AI captions for them — all at once, in bulk for selected posts, or one at a time — before scheduling, so that every post is ready to go without me having to write captions manually.

---

### Workflow:

1. User arrives at Step 4 after completing Step 3 — wizard card expands to max-width 1000px
2. Posts table shows one row per uploaded image; all posts show "Pending" status (yellow badge)
3. Info banner shown: "Your images are ready. Click 'Generate All Captions' to auto-generate captions for all posts at once, or generate individually per post."
4. Header row shows: post count ("{N} post(s)") and "Generate All Captions" button

**Generate All flow:**
5. User clicks "Generate All Captions" — button enters loading state ("Generating captions…" with animated dots)
6. Posts are processed sequentially — each post's thumbnail shows a pulsing `border-primary-cs-200` ring; status badge turns "Generating" (blue); content area shows animated dots + "Generating caption…"
7. As each caption completes: caption text appears in the content area; status badge turns "Ready" (green); pulsing ring disappears; "Generate caption" link disappears for that post
8. Once all posts have captions: "Generate All Captions" button disappears; info banner dismissed
9. User selects all rows → clicks "Schedule Selected" in bulk actions bar → posts scheduled

**Single-post generate flow:**
10. User sees a post with "Pending" badge and "Generate caption" link
11. User clicks "Generate caption" — that post enters generating state (ring + blue badge + animated dots)
12. Caption appears → post becomes "Ready"

**Zero credits flow:**
13. User clicks any generate trigger
14. Toast: "You've run out of AI text credits. [Upgrade your plan] to get more." — orange toast, 5s
15. No generation starts; page state unchanged; buttons stay enabled

**Partial credits flow:**
16. "Generate All" starts; credits are exhausted partway through
17. Completed posts show "Ready"; remaining posts stay "Pending"
18. Toast: "Some captions couldn't be generated — you ran out of AI text credits. [Upgrade your plan] to get more." — error toast, 7s

**Generation failure flow:**
19. AI service fails for a specific post
20. Content area shows: "Couldn't generate caption. [Try again]"
21. Status badge reverts to "Pending"; pulsing ring disappears
22. User clicks "Try again" → re-triggers single-post generation for that post

**Closing during generation:**
23. User clicks × while generation is in progress
24. Blocking modal appears: "Caption generation in progress"
25. User clicks "Stay & Wait" → modal dismissed, generation continues
26. OR user clicks "Leave Anyway" → generation cancelled, user returns to landing page / listing

**Bulk select and actions:**
27. User checks multiple rows → floating bulk actions bar appears at bottom-center of screen
28. Bar shows: "{N} Selected", "( Unselect all )" link, "Actions ▾" button
29. Clicking "Actions ▾" opens dropdown above the bar
30. User selects "Generate Captions" to generate for selected pending posts only
31. OR selects "Schedule Selected" to schedule selected posts
32. OR selects "Delete Selected" to remove posts from the batch

---

### Acceptance criteria:

**Step 4 layout:**
- [ ] Wizard card expands to max-width 1000px on Step 4
- [ ] Step indicator shows Step 4 active with Steps 1–3 marked complete (green checkmark)
- [ ] "Previous" button not shown (processing already started on Step 3 → Next)

**Header row:**
- [ ] Post count shown top-left: "{N} post(s)" — bold number, `text-gray-500` label
- [ ] "Generate All Captions" button shown in header (image mode only); tooltip on hover: "Uses your AI text credits."
- [ ] Generate All — active state: animated three-dot indicator + "Generating captions…" text; `bg-primary-cs-50` background, `border-primary-cs-200` border; cursor `not-allowed` (non-clickable while generating)
- [ ] Generate All — zero credits: clicking shows toast "You've run out of AI text credits. [Upgrade your plan] to get more." — orange, 5s; button stays enabled
- [ ] Generate All button disappears once all posts have captions (pendingCount === 0); no success message shown

**Info banner:**
- [ ] Banner shown in image mode before any captions are generated: "Your images are ready. Click 'Generate All Captions' to auto-generate captions for all posts at once, or generate individually per post."
- [ ] Banner style: `bg-primary-cs-50`, `border-primary-cs-200`, left-aligned icon
- [ ] Banner auto-dismissed when generation starts or any caption is populated; does not reappear

**Posts table columns:**
- [ ] Checkbox (row-level + header select-all)
- [ ] Post (thumbnail + content area — see below)
- [ ] Accounts (`SocialChannelsTooltip` with avatars of selected accounts)
- [ ] Status (Pending / Generating / Ready badge)
- [ ] Label (tag chips — same as CSV mode)
- [ ] Campaign (campaign badge — same as CSV mode)
- [ ] Actions (Schedule link, Edit (pencil icon), Delete (trash icon))

**Post column — thumbnail:**
- [ ] 56×56px, `border-radius: 8px`, `1px border-gray-200`, `object-fit: cover` — shows uploaded image
- [ ] Broken/failed image: generic placeholder icon fallback (Lucide ImageOff — cycles by post index); `onError` handler
- [ ] Generating state: pulsing `border-primary-cs-200` ring animation around the thumbnail

**Post column — content area states:**
- [ ] Has caption: caption text, max 2 lines with `text-ellipsis`
- [ ] No caption (image mode): Line 1 "No caption yet." (`text-gray-400`, italic); Line 2 "Generate caption" (`text-primary-cs-500`, dotted underline); hover: `text-primary-cs-700`; tooltip: "Uses your AI text credits."
- [ ] No caption (CSV mode): Line 1 "No caption yet." (`text-gray-400`, italic); Line 2 "Edit manually →" (`text-gray-400`, italic) — no generate link in CSV mode
- [ ] Generating: animated three-dot indicator + "Generating caption…" at fixed 52px height to prevent row jump
- [ ] Generation failed: "Couldn't generate caption." + "[Try again]" link (`text-primary-cs-500`); clicking re-triggers single-post generation

**Status badges:**
- [ ] Pending: "Pending" — `bg-yellow-100 text-yellow-700`; shown on posts with no caption (image mode only)
- [ ] Generating: "Generating" — `bg-primary-cs-50 text-primary-cs-600`; shown while AI is working on that post
- [ ] Ready: "Ready" — `bg-green-100 text-green-700`; shown once caption is present

**Generation triggers:**
- [ ] Generate All: generates captions for all posts with no caption; sequentially processed (~500ms stagger); disabled during active generation
- [ ] Single-post "Generate caption" link: visible only on posts with no caption and not currently generating; triggers generation for that post only; can be clicked while Generate All is running (post joins queue)
- [ ] Bulk generate (Actions menu): only visible when ≥1 selected post has no caption; generates for selected pending posts only

**Credit handling:**
- [ ] Zero credits — any trigger: toast "You've run out of AI text credits. [Upgrade your plan] to get more." — orange, 5s; no generation
- [ ] Partial credits — mid-generation: completed posts show "Ready"; remaining stay "Pending"; toast "Some captions couldn't be generated — you ran out of AI text credits. [Upgrade your plan] to get more." — error, 7s
- [ ] Generation failure — single post: inline error "Couldn't generate caption. [Try again]"; status reverts to "Pending"

**Bulk selection & actions bar:**
- [ ] Floating bar appears at bottom-center when ≥1 row selected; disappears when selection cleared
- [ ] Selected count: "{N} Selected" — white text on `bg-gray-800`
- [ ] Unselect all: "( Unselect all )" — `text-primary-cs-300`, clears selection
- [ ] "Actions ▾" button — `bg-gray-700`; opens dropdown above the bar
- [ ] Dropdown — "Generate Captions": image mode only; visible only when ≥1 selected post has no caption; sub-label "For {N} selected post(s) · uses AI credits"; tooltip "Uses your AI text credits."; hidden if all selected posts already have captions
- [ ] Dropdown — "Schedule Selected": available in both modes; sub-label "Schedule {N} post(s)"
- [ ] Dropdown — "Delete Selected": red text; sub-label "Remove {N} post(s) from this batch"; clears selection after

**Close-while-generating guard:**
- [ ] Clicking × while generation is in progress triggers a blocking modal (cannot be dismissed by clicking outside)
- [ ] Modal title: "Caption generation in progress"
- [ ] Modal body line 1: "AI is still generating captions for your posts. If you leave now, generation will stop and any captions not yet completed will be lost."
- [ ] Modal body line 2 (bold): "AI credits already used for completed captions will not be refunded."
- [ ] Primary button: "Stay & Wait" — primary style; dismisses modal, generation continues
- [ ] Destructive button: "Leave Anyway" — white button, `text-red-600`, `border-red-300`; cancels generation and navigates to landing page / listing

**General:**
- [ ] All new strings use i18n keys in `automation.csv_bulk_schedule` namespace; translations for all 7 locales
- [ ] Row height remains stable during all state transitions (no layout jump when content area changes)
- [ ] No hardcoded hex values — all colors via Tailwind theme classes

---

### Mock-ups:
N/A — no designs provided; implementation follows PRD §8.5–§8.10.

---

### Impact on existing data:
- Reads `ai_generated_caption` flag and caption data from posts returned by the backend (new fields added in **[BE] Extend Bulk Schedule API**)
- CSV mode Step 4 is unchanged — no AI generation controls shown in CSV mode

---

### Impact on other products:
Web only. No mobile impact.

---

### Dependencies:
- **[BE] Extend Bulk Schedule API for Image Upload mode and AI caption generation** — caption generation endpoint and post status responses
- **[FE] Build Image Upload wizard — Step 1, mode routing, and Step 3 adaptation** — Step 3 → Next navigates to this step

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — Bulk Schedule is desktop-only; N/A
- [ ] Multilingual support — all new strings in all 7 locale files
- [ ] UI theming support — all colors via `primary-cs-*` and `gray-*` Tailwind classes; no hardcoded hex
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)
