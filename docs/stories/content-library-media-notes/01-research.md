# Research — Media Notes for the Content Library upload flow

## Summary

Let users attach a short text note to each file **while uploading** it to the Content Library. The note is:
- Added/edited/deleted only during the upload step (from the file tile in the "Upload to Content Library" modal).
- **Locked (read-only) once the file is uploaded** — never editable or deletable afterward.
- Displayed read-only in the File Info panel when the uploaded image is opened later.

Because the same upload modal and the same file tile are reused everywhere media can be added from the Content Library (composer, campaigns, automation, discovery, blog, etc.), building the note into those shared components makes the feature work identically at every entry point automatically.

---

## Current State

### The upload modal (shared everywhere)
- **`contentstudio-frontend/src/modules/publish/components/media-library/components/UploadMediaModal.vue`** — the "Upload to Content Library" modal. This single component is mounted at every entry point (see Entry Points below).
- Inside it, the upload/drag-drop tab is **`.../components/MediaTabs/UploadFilesTab.vue`**. Selected files are held in a reactive `images` array and each is rendered with `v-for="(image, k) in images"` → `<Asset :info="image" type="secondary" :hide-delete="isUploading" @delete="deleteItem" />` (UploadFilesTab.vue ~line 98-111).
- **`.../components/MediaUploadFooter.vue`** holds the folder selector + Upload button; `@upload="uploadFiles"` kicks off the actual upload.

### The file tile (shared component)
- **`.../components/Asset.vue`** (636 lines) is the tile used both in the upload modal (`type="secondary"`) and the main library grid. Its `info` prop is a per-file object (`AssetInfo` interface, Asset.vue ~line 24-43, with an open `[key: string]: unknown`).
- Existing overlays on the tile:
  - `.asset-menu` — the three-dots menu (`Dropdown` + `Icon name="EllipsisVertical"`), positioned `absolute bottom-2.5 right-2.5`, revealed on hover (`.asset:hover .asset-menu`). Matches the spec's "three-dots menu on hover at bottom-right".
  - `.asset-check-button` / `.asset-post-count` — top-right zone (`top-2 right-2`). This is the corner where the **always-visible note icon** would live.
- So the note icon is a **new, always-visible** control in the top-right corner; the existing three-dots menu stays at bottom-right on hover — the two are separate, exactly as specced.

### The File Info panel
- **`.../components/PreviewMediaAssetModal.vue`** (613 lines) — opens when a user previews an uploaded asset. Right side has a "File info" tab with stacked rows (`div.file-property`, label 40% / value 60%): **name, type, size, created on, and a conditional Dimensions row** (`v-if="mediaDimension"`, PreviewMediaAssetModal.vue ~line 161-165). The Note card goes **directly below this Dimensions row**.
- The panel header already renders the uploader: `currentMedia?.user?.full_name` + "uploaded this file" + relative time (~line 97-121). Uploader identity is already available client-side.

### Media data + upload API
- Frontend API layer: **`contentstudio-frontend/src/api/media-library.ts`** and **`src/api/media.ts`** (typed, on `@api/client`). URLs in `src/config/api-utils.js` (`uploadByteFileUrl`) and `src/modules/publish/config/api-utils.js` (`uploadMediaAssetUrl`, `fetchMediaAssetsUrl`, etc.).
- **Per-file selected-file shape** during upload is `ImageItem` in `UploadFilesTab.vue` (`{ _id, name, size, type, needsConversion, url, file }`) — `_id` is a local sequential counter, **not** a server id. No note field exists on it.
- **Upload payload flow:** files go through the shared upload queue (`src/composables/useMediaUploadQueue.ts` → `addToQueue`) and the GCS path (`src/composables/useGCSUpload.ts` → `processUploadedMedia`, payload `{ workspace_id, files: [fileDetails], folder_id?, is_root? }`, endpoint `POST api/media_library/assets/process-uploaded-gcs-media`) or the white-label proxy path (`src/modules/common/composables/useFileUpload.js`). **A note captured at upload has nowhere to travel today** — `ImageItem`, `QueueItem`, and `FileDetails` would each need a new `note` field and the process-upload payload would carry it. **Extending the upload request shape is an API-contract change** (per frontend CLAUDE.md, that requires confirmation) — call it out in the BE story.
- Backend controller: **`contentstudio-backend/app/Http/Controllers/Storage/MediaLibrary/MediaLibraryAssetsController.php`** (2163 lines). Upload routes (`contentstudio-backend/routes/web/storage.php`): `/assets/upload`, `/uploadDocument`, `/uploadByLink`, `/uploadByBytes`, `/blogImageUpload`; fetch via `/assets/fetch`.
- **Uploader is already captured:** every upload path sets `$mediaDetails['user_id'] = Auth::id()` on the asset document (controller lines 160, 295, 463, 1168, 1315). The frontend resolves this into `currentMedia.user.full_name`. **No new attribution capture is needed — only a new `note` field.**
- Asset document assembly happens in **`contentstudio-backend/app/Libraries/Storage/MediaLibrary.php`** (`$mediaDetails[...]` with name/mime_type/size/etc.).

### Existing note/caption pattern
- **No per-file "note", "caption", or "description" field exists on media library assets today.** This is a net-new field on the media asset.
- **Closest precedents to reference (structure only):**
  - Per-image **alt text** on the composer post — `sharingDetails.alt_texts` as `[{ image, alt_text }]` (`composer_v2/components/EditorBox/EditorBox.vue`). This is per-image text metadata keyed by image, but it lives on the *post* payload, not the library asset.
  - AI **caption editor** UI — `composer_v2/components/EditorBox/SaveCaptions/AddEditCaption.vue` (textarea + save/edit form). Good structural template for the "Add a note" modal.

---

## What Needs to Change

**Backend ([BE]):**
- Accept an optional per-file `note` (≤ 500 chars) on all Content Library upload paths (file upload, byte upload, link upload, document upload).
- Persist `note` on the media asset document alongside the existing `user_id`.
- Return `note` in the media asset fetch responses so the File Info panel can display it.
- Treat the note as **write-once**: only set at creation; no endpoint to edit or delete a note after the asset exists. (Deleting the whole asset removes it, as today.)

**Frontend ([FE]):**
- Add an always-visible **note icon** to the top-right of each file tile in the upload modal (neutral outline when empty; filled yellow-500 background with grey-700 icon when a note exists).
- Build the centered **"Add a note" modal** (icon badge, title, one-line description, 500-char `Textarea`, Cancel / Save note, X to close; no file name shown).
- Store the note on the per-file object so it's included in the upload payload; allow add/edit/delete only before upload.
- After upload, the note is locked (the icon/modal are no longer editable).
- Render the read-only **Note card** in the File Info panel directly below the Dimensions row (yellow-50 bg, yellow-100 border, grey-500 text), with a "Note" label + blue info icon tooltip, and an "Added by [uploader] · at upload" caption below.

---

## Entry Points (where the upload modal appears)

`UploadMediaModal.vue` is imported/mounted in (confirmed by grep):
- `Home.vue`
- `components/GoogleDriveAuth.vue`
- `modules/discovery/components/DiscoveryNavigationLayout.vue`
- `modules/composer/components/blog/Blog.vue` ⚠️ **legacy `composer` module (feature-frozen)** — don't implement here; the active composer is `composer_v2`
- `modules/composer_v2/views/SocialModal.vue` (active composer)
- `Home.vue` also mounts the app-wide global instance (`modal-id="global-upload-media-modal"`)
- `modules/automation/components/csv/BulkUploadAutomationSave.vue`
- `modules/automation/components/evergreen/create/EvergreenMain.vue`
- `modules/publish/components/media-library/MediaLibraryMain.vue`
- `modules/publish/components/media-library/composables/useMediaLibraryFetch.ts`

Because all of these reuse the same `UploadMediaModal` → `UploadFilesTab` → `Asset` chain, building the note into those shared components delivers the "works identically everywhere" requirement without per-surface work.

---

## UI Components Available (from `docs/ui-components.md`)

- **`Modal`** (`@contentstudio/ui`, via `$cstuModal` / `ModalPlugin`) — for the "Add a note" modal (already used in the media library, e.g. `CreateFolderModal.vue`, `PreviewMediaAssetModal.vue`).
- **`Textarea`** (`@contentstudio/ui`) — for the 500-char note field.
- **`Button`** — Cancel / Save note.
- **`Icon`** — note icon on the tile, info icon in the File Info card, X close.
- **`Dropdown` / `DropdownItem`** — already used for the tile three-dots menu (no change).
- **Tooltip:** There's no standalone Tooltip *component*, but the codebase uses the **`v-tooltip` directive** for hover tooltips (e.g. `Asset.vue`, `MediaUploadFooter.vue`). Use `v-tooltip` for the File Info "Note" info-icon tooltip — no new component needed.
- **Modal access:** `CstuModal` via `$cstuModal` from `inject('root') as ModalPlugin` (the pattern already in `UploadMediaModal.vue`).

---

## Color / theming note

The spec uses fixed yellow/grey/blue hexes (yellow-50 `#FFF9EE`, yellow-100 `#FFEDCA`, yellow-500, grey-700 `#353535`, grey-500 `#4a4a4a`, blue-450 for the info icon). These are **semantic accent colors for the note affordance**, not the brandable primary — so they are intentionally fixed and do **not** use `primary-cs-*`. The FE story should map them to the nearest Tailwind tokens already in the palette (e.g. `bg-yellow-50`, `border-yellow-100`, `text-gray-700`) rather than hardcoding hex, per repo styling rules.

---

## Mobile Context

Out of scope for mobile stories. The upload modal, file tile, and File Info panel are all web (Vue) components; the spec is written entirely in web terms ("Upload to Content Library" modal, tiles, File Info panel). Because the note is stored on the asset (BE), the iOS/Android apps *could* surface the read-only note later, but that is a future enhancement, not part of this change. Noted under "Impact on other products" in the stories.

---

## Files Involved

**Frontend:**
- `src/modules/publish/components/media-library/components/Asset.vue` — add note icon overlay (upload context only)
- `src/modules/publish/components/media-library/components/MediaTabs/UploadFilesTab.vue` — wire note state onto per-file `images` objects + include in upload payload
- `src/modules/publish/components/media-library/components/UploadMediaModal.vue` — host the "Add a note" modal (or a new `AddNoteModal.vue` sibling)
- `src/modules/publish/components/media-library/components/PreviewMediaAssetModal.vue` — read-only Note card below Dimensions row
- `src/api/media-library.ts` — include `note` in upload payloads; surface `note` on fetched assets
- `src/locales/*/` — new i18n keys (all 8 locales)

**Backend:**
- `app/Http/Controllers/Storage/MediaLibrary/MediaLibraryAssetsController.php` — accept + validate `note` per file on all upload paths
- `app/Libraries/Storage/MediaLibrary.php` — persist `note` on the asset document; return it in fetch
- `routes/web/storage.php` — no new routes (reuse existing upload/fetch endpoints)
