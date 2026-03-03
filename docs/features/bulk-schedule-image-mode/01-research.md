# Research: Bulk Schedule — Image Upload Mode

## Existing Codebase

### Frontend Architecture
- **Main wizard:** `src/modules/automation/components/csv/BulkUploadAutomationSave.vue` — Options API, ~1300 lines, handles all 4 wizard steps in one component
- **Listing page:** `src/modules/automation/components/csv/listing/CsvProcessListing.vue` — entry point, DataTable of past bulk schedules
- **Composables:** `composables/useCsvDraft.js` (draft save/load), `composables/useCsvProcesses.js` (process operations)
- **Store:** `src/modules/automation/store/recipes/csv.js` — Vuex, manages selection state, tab_status, listing
- **API config:** `src/modules/automation/config/api-utils.js`

### Wizard Step Structure
| Step | Name | Key Components |
|---|---|---|
| 1 | Upload CSV File | Drag-drop zone, file picker (CSV/XLSX/XLS), progress simulation, GCS upload |
| 2 | Name & Accounts | TextInput, `AccountSelection`, `ContentCategorySelection` |
| 3 | Scheduling | `AutomationScheduleOptions` (reusable) — tabs: Use times from CSV / Regular Intervals / Content Category / Add to Queue |
| 4 | Finalizing Posts | DataTable — columns: Name/content, Accounts, Total posts, Status |

### File Upload Flow (existing)
1. Frontend sends multipart FormData with `inputFile`
2. Backend (`CSVController::saveDraftAutomation`) receives → `MediaLibrary::uploadOriginalsToGCSFromFile('/bulk_csv_files', file)` → returns GCS public URL
3. URL + filename stored in MongoDB `csv_processing` collection
4. On Step 4, backend fetches file from GCS, parses CSV, processes posts

### Backend
- **Controller:** `app/Http/Controllers/Automation/CSVController.php`
- **Model:** `app/Models/Publish/Automation/CsvProcessing.php` — MongoDB collection `csv_processing`
- **Config:** `config/csvAutomation.php` — default automation structure
- **Image upload trait:** `app/Traits/TraitBulkImageUploader.php` — `bulkUploadInChunks(urls, workspace_id, user_id)`, uploads to `media_library/{workspace_id}/uncategorized/original/bulk_uploader/`
- **Key routes:**
  - `POST /saveDraftCsvAutomation`
  - `POST /processCsv`
  - `POST /fetchCsvPosts`
  - `POST /csvCheckProcess`
  - `POST /csvPostBulkAction`

### AI Caption (existing, NOT in bulk schedule)
- Lives in `composer_v2/components/AiCaptionModal.vue` — side-slide modal
- Credit tracking via `getCreditUsedLimit` / `getCreditSubscribeLimit` getters
- No existing AI integration in the bulk schedule wizard
- Toast pattern: `dispatch('toastNotification', { message, type: 'success'|'error'|'warning' })`

### Reusable Components Available
- `StepIndicator` — shared step progress bar, already used in wizard
- `AccountSelection` — Step 2, fully reusable
- `ContentCategorySelection` — Step 2, fully reusable
- `AutomationScheduleOptions` — Step 3, reusable (tab filtering needed for image mode)
- `DataTable` — Step 4 listing
- `SocialChannelsTooltip` — account icons in table
- `TraitBulkImageUploader` — backend image upload to GCS/media library

### Drag-Drop Pattern (existing in wizard)
Steps 1 already has `@dragenter`, `@dragover`, `@dragleave`, `@drop`, `@paste` handlers — the same pattern can be reused for image uploads.

---

## Key Spec Adjustments Applied

| Spec issue | Fix applied in PRD |
|---|---|
| Tooltip: "Uses AI text credits from your workspace." | → "Uses your AI text credits." (ContentStudio doesn't append "from your workspace") |
| Zero credits handling: "disable buttons / show modal" | → Show toast when user clicks generate: "You've run out of AI text credits. [Upgrade your plan] to get more." Buttons remain enabled. |
| Hardcoded colors (`#7C3AED`, `#4F46E5`, `#F8FAFF`, `#1F2937`, etc.) | → All replaced with `text-primary-cs-500`, `bg-primary-cs-50`, `border-primary-cs-200`, `bg-gray-800`, etc. |
| File > 10 MB: "rejected silently or inline error — confirm with engineering" | → Show inline toast: "'{filename}' is too large. Max file size is 10 MB." Skip the file and continue. |
| AI generation failure: "Confirm error copy with engineering" | → Defined: inline error in post column: "Couldn't generate caption. [Try again]" — retry triggers single-post generation |
| Closing during generation: body copy inconsistency | → Cleaned up and finalized (see PRD §8.1) |
| "Back to all options" tooltip on × button | → Removed; × button needs no tooltip — its behaviour is self-evident |

---

## Story Split

1. `[BE]` Extend Bulk Schedule API for Image Upload mode
2. `[FE]` Update Bulk Schedule landing page with mode selector
3. `[FE]` Build Image Upload wizard (Step 1 + mode routing + Step 3 adaptation)
4. `[FE]` Build Finalizing Posts step with AI caption generation (Step 4)
