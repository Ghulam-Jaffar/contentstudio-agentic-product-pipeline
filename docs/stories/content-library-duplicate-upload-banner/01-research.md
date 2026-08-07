# Research — Duplicate detection when uploading to the Content Library

## The problem (user's words)

When you bulk-upload media to the Content Library and only some files go through (network drop, storage hiccup, etc.), you don't get a clear picture of *which* files made it and which didn't. If you then re-select the same batch to retry, there's no signal telling you which ones are already in your library — so you either re-upload everything (creating duplicates) or waste time cross-checking by hand. Painful when the batch is large.

**Goal:** when the user selects files that are already in their Content Library, flag those files in the "Upload files from your computer" screen so they can skip them (or upload anyway — their choice). The approved design shows:
- A **warning banner** at the top of the upload area summarising how many are duplicates, with **Skip duplicates** / **Keep all** actions.
- A **red "Duplicate" label with a warning triangle** under each duplicate tile, plus an amber ring around the tile.

## Current state

The Content Library upload lives in the media-library module (web app only):

- **`contentstudio-frontend/src/modules/publish/components/media-library/components/UploadMediaModal.vue`** — the modal shell ("Upload to content library"), with a left rail of sources (Upload files, Media library, From URL, Dropbox, Google Drive, etc.) and a storage-usage indicator.
- **`.../components/MediaTabs/UploadFilesTab.vue`** — the "Upload files from your computer" tab. Holds the selected-files array (`images`), the drag-drop dropzone, the "Add more files" tile, the folder selector, and the **Upload** button. On upload it hands files to `useMediaUploadQueue` (proxy upload for white-label) or `useGCSUpload` (signed URLs otherwise).
- **`.../components/Asset.vue`** — the individual media tile (thumbnail, file-type chip, delete button, size/name overlay). Selected-files use `type="secondary"`. This is where the per-tile "Duplicate" label attaches.
- **`src/composables/useMediaUploadQueue.ts`**, **`src/composables/useGCSUpload.ts`** — the actual upload pipeline.

**There is no duplicate detection anywhere in this flow today** — grep for `duplicate` / `dedup` / `hash` / `md5` in the media-library area and upload composables returns nothing. Every selected file is uploaded as-is, so retrying a partial batch silently creates second copies.

## What needs to change

1. **Backend:** a lightweight "check which of these files already exist" endpoint. Given a list of files the user has selected (name + size, or a content hash), return which ones already exist in the workspace's Content Library.
2. **Frontend:** when files are added to the "Upload files" selection, call that check, then:
   - Show the warning banner with counts + **Skip duplicates** / **Keep all**.
   - Show the red "Duplicate" label + amber ring on each matching tile.
   - Let the user skip all duplicates in one click, keep all, or remove tiles individually — then upload whatever remains.

## How a duplicate is decided

Simplest heuristic that's easy to explain and reliable enough: **a file is a duplicate when the Content Library already contains a file with the same name and same size.** Engineering may choose a content hash (e.g. MD5) instead/as well for higher precision — left as an implementation choice. Scope of the match is the **whole workspace library** (a duplicate is a duplicate regardless of which folder it landed in), not just the selected target folder.

Within a single selection, two identical picks (same name + size) are a client-side dedupe and don't need the backend.

## UX reference

This is the standard "we found N items already in your library" pattern used by file managers and asset tools (Google Drive's "keep both / replace / skip", Dropbox upload conflicts). We deliberately do **not** auto-skip or auto-replace — uploading anyway is a legitimate choice, so the flags are advisory and the user stays in control.

## Components available (from `docs/ui-components.md`)

- **`Alert`** (`@contentstudio/ui`) — inline alert; use the warning variant for the summary banner.
- **`Icon`** — renders the warning-triangle (`AlertTriangle`) used in the banner and the per-tile label.
- **`Button`** — for "Skip duplicates" / "Keep all".
- No dedicated pill/chip component; the per-tile "Duplicate" label is just an `Icon` + text (theme-aware classes), matching the approved mock.

## Scope / platform

- **Web only.** The Content Library upload modal is a web-app feature. Mobile apps aren't in scope for this change (no request to surface duplicate flags there). The new BE endpoint is additive and doesn't change existing upload behaviour.
- **First delivery scoped to the "Upload files from your computer" tab.** Other sources in the same modal (From URL, Dropbox, Google Drive, Pixabay/Giphy) can reuse the same check later — noted as a follow-up, not in this pair.

## Files involved

- `contentstudio-frontend/src/modules/publish/components/media-library/components/MediaTabs/UploadFilesTab.vue`
- `contentstudio-frontend/src/modules/publish/components/media-library/components/Asset.vue`
- `contentstudio-frontend/src/modules/publish/components/media-library/components/UploadMediaModal.vue` (if the banner sits at modal level)
- `contentstudio-frontend/src/composables/useMediaUploadQueue.ts`, `useGCSUpload.ts` (selection → upload)
- `contentstudio-frontend/src/config/api-utils.js` (new endpoint URL)
- `contentstudio-backend/` — media library assets controller/service (new "check duplicates" endpoint alongside the existing `media_library/assets/uploadByBytes`)
