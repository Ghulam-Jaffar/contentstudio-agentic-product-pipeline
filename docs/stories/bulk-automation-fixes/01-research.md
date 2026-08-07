# Research — Bulk Automation fixes (Publisher)

Three fixes in the Publisher **Bulk Scheduling / Bulk Automation** flow (CSV + image upload wizard).

---

## Fix 1 — Wrong banner wording in the image-upload flow

### Current State
- The bulk save wizard shows a note/warning banner on the **Scheduling** step:
  `contentstudio-frontend/src/modules/automation/components/csv/BulkUploadAutomationSave.vue:715` — a `CstAlert` rendering `note_prefix` + `note_message`.
- The copy lives at `src/locales/en/automation.json:227-228`:
  - `note_prefix`: `"Note:"`
  - `note_message`: `"Review your settings before proceeding, CSV will be processed into posts and you won't be able to move back to any previous steps."`
- The banner is rendered **unconditionally** — it always says "CSV" even when the user is in the **image-upload** flow. The flow is already distinguished in the component by an `isImageMode` flag (imported at `BulkUploadAutomationSave.vue:33`), so a conditional message is straightforward.

### What Needs to Change
- Show **image-flow-appropriate wording** when `isImageMode` is true (talk about "images" being turned into posts, not "CSV").
- **Improve the copy** for both flows (clearer, friendlier).
- Purely a **frontend + i18n** change (add a second message key, choose it by `isImageMode`). Keys must land in all 8 locale dirs.

---

## Fix 2 — Finalizing Posts step: image preview + date collide with the background on scroll

### Current State
- Step 4 is **"Finalizing Posts"** (`automation.json:99` → `steps.finalizing`).
- It renders a horizontally + vertically scrollable table inside `BulkUploadAutomationSave.vue` (grid starts ~line 851, `v-dragscroll`).
- The table uses layered sticky positioning and z-indexes:
  - Sticky header row: `sticky top-0 ... z-40 bg-[#FBFBFB]` (line 858)
  - Sticky first column (checkbox): `sticky left-0 z-35 / z-30 bg-white` (lines 859, 918)
  - Sticky last column (actions): `sticky right-0 z-35 / z-30 bg-white` (lines 893, 1253)
  - A fixed bottom gradient overlay: `fixed ... z-20` (line 1407)
- The **middle scrollable cells** — including the **image thumbnail preview** and the **date** cell — do not carry the same solid-background / stacking treatment as the sticky header/columns. When the user scrolls, the image preview and date visually **bleed into / overlap** the sticky header or background layers (a z-index / opaque-background stacking bug).

### What Needs to Change
- Correct the **stacking/background layering** in the Finalizing Posts table so the image-preview thumbnail and the date cell stay cleanly behind the sticky header and don't collide with the background when scrolled.
- Purely a **frontend (CSS/z-index)** fix in `BulkUploadAutomationSave.vue`.

---

## Fix 3 — Edit (pencil) option for automations that completed with 0 posts

### Current State
- The automations list is `src/modules/automation/components/csv/listing/CsvProcessListing.vue`.
- Each row item has `is_draft`, `approved_posts` (total-posts count, shown as `item.approved_posts || 0`, line 176), and a status column: `is_draft` → **"Draft"**, else → **"Completed"** (lines 182-192).
- Action column (lines 209-235):
  - **Drafts** (`v-if="item.is_draft"`): a **pencil Edit** icon (`editDraft(item)`, tooltip `edit_draft`) and a Delete icon.
  - **Completed** (`v-if="!item.is_draft"`): only a **"View Posts"** link (opens the posts in Planner via `csv_id`).
- `editDraft(item)` (lines 409-418) navigates to route `save-bulk-csv-automation` with `query: { draft_id: item._id, mode: 'image' (if item.image_mode) }`.
- Loading for edit: `useCsvDraft.ts` → `loadDraftForEditing(draftId)` → `fetchDraft` (TanStack `getCsvDraftOptions`) → restores `file_url`, images, settings, and sets `SET_CSV_IS_DRAFT(true)`.
- Backend routes (`contentstudio-backend/routes/web/automation.php`): `fetchCSVAutomations` (list), **`fetchCsvAutomation`** (single by id — generic, not draft-only), `saveDraftCsvAutomation`, `deleteDraftCsvAutomation`, `fetchCsvPosts`.

### The scenario
- When a bulk automation is processed and **every post ends in error**, the automation still lands in the list as **status "Completed" with 0 posts** (`approved_posts === 0`). It only offers "View Posts" (which shows nothing useful) — there's no way to fix and re-run it.

### What Needs to Change
- Show the **pencil Edit** action (same affordance drafts get) for list items that are **completed AND have 0 posts** (`!is_draft && approved_posts === 0`) — and only that case; normal completed automations (with posts) keep just "View Posts".
- Clicking it reopens the automation in the bulk save wizard so the user can fix settings/content and re-process.
- **Open dependency (needs user + possibly BE):** the draft edit path forces `is_draft = true` and relies on the automation still holding its **source data** (uploaded CSV/images + settings + rows). It's unconfirmed whether a fully-errored *completed* automation retains that source and can be loaded back via `fetchCsvAutomation` into an editable state. If it can, fix #3 is FE-only; if not, a **[BE] story is required** to make completed-0-post automations reopenable/editable (retain source + allow loading into the edit flow). → **Clarify with user.**

---

## Story Split (proposed)

- **[FE]** — one story covering all three fixes (all live in `BulkUploadAutomationSave.vue` + `CsvProcessListing.vue`, one PR / one QA pass on the bulk flow): banner wording (image vs CSV), Finalizing-step scroll overlap, and the edit pencil for completed-0-post automations.
- **[BE]** — only if fix #3 needs backend support to reopen/edit completed-0-post automations (pending clarification).

Within the `/story` limit (≤4 stories).

---

## Mobile Context
Not impacted. Bulk Scheduling / CSV & image bulk automation is a **web-only** Publisher feature — no equivalent flow in the iOS/Android apps. No mobile stories.

---

## Files Involved

**Frontend:**
- `src/modules/automation/components/csv/BulkUploadAutomationSave.vue` — banner (fix 1), Finalizing table stacking (fix 2)
- `src/modules/automation/components/csv/listing/CsvProcessListing.vue` — edit pencil for completed-0-post items (fix 3)
- `src/modules/automation/components/csv/composables/useCsvDraft.ts` — edit-loading path (fix 3, if reused)
- `src/locales/en/automation.json` (+ all 8 locale dirs) — `csv_bulk_schedule.save.scheduling.note_message` and a new image-flow variant; edit tooltip for completed-0-post

**Backend (only if fix 3 needs it):**
- `contentstudio-backend/app/Http/Controllers/.../` CSV automation controller behind `fetchCsvAutomation` / `saveDraftCsvAutomation` — allow reopening a completed-0-post automation for editing
