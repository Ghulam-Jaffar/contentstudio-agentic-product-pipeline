# 01 — Research: Custom Video Thumbnail → choose from Media Library

## Problem (user report)

In the composer's **Set Custom Thumbnail** modal, the **Upload Image** tab only supports uploading a fresh file from the computer. When a user customizes a post per platform (e.g. Facebook, Instagram, LinkedIn separately), they must upload/generate the thumbnail again for each platform, and again every time the video is re-uploaded. The user wants to reuse an image already in their **Media/Content Library** (and Google Drive / Dropbox) as the custom thumbnail — the same media widget the composer uses everywhere else — instead of only a bare file upload.

## Current State

The **Set Custom Thumbnail** modal has three radio options: **Choose Suggested**, **Choose From Video**, **Upload Image**. The **Upload Image** tab is a plain `<label>` + `<input type="file">` dropzone (click-to-upload / drag onto the box only). There is no Media Library, Google Drive, or Dropbox source in this tab.

- **Modal:** `contentstudio-frontend/src/modules/composer_v2/components/CustomThumbnail/CustomThumbnailModal.vue`
  - Upload tab markup: the `<label>` block (~lines 116-155) with a single `<input type="file" accept="image/png,image/jpeg,...">`.
  - `uploadNewThumbnail(event)` captures the chosen File and builds a local preview; `uploadFrame(file)` runs per-platform size checks then uploads via `uploadFilesHelper` (from `useMediaHelper`) to storage and calls `setThumbnailURL(mediaURL)`; `handleApply()` routes by `type` (`suggested` / `frame` / `upload`).
  - Per-platform validations already live here: Instagram image `max_size` and YouTube thumbnail `max_size` (from `socialIntegrationsConfigurations`), with warning messages for YouTube Shorts / TikTok / Instagram-reel-only.
- **The media widget we want to reuse:** `contentstudio-frontend/src/modules/composer_v2/components/MediaSelection.vue`
  - Already provides: drag/drop + paste + click-to-upload, **Media Library**, **Google Drive**, **Dropbox** (plus design tools Canva/Vista/PostNitro that are NOT relevant to a thumbnail picker and are already hidden on white-label domains).
  - Emits `media-action` events (`upload`, `openMediaLibrary`, ...) to its parent; opens the library modal via EventBus `show-media-library-modal` (Google Drive = `sideTabIndex: 9`, Dropbox = `sideTabIndex: 8`).
  - Supports images + video + pdf; for the thumbnail use case it must be constrained to **images only**.
- **Media library selection flow:** the library modal returns picks via EventBus `add-media-from-media-library` (and platform-specific variants) — today those **attach media to the post**, not to a thumbnail picker. A thumbnail-scoped selection path is needed so a picked image sets the thumbnail instead of being added to the post.
- **Modal callers (context):** `EditorMediaBox.vue` and `SocialModal.vue` open the modal via EventBus `custom-thumbnail-modal-show` and receive the result through the `emitter` callback.

## What Needs to Change (frontend only)

- In the **Upload Image** tab, replace the bare file dropzone with the composer's media-picker experience: **drag & drop / click to upload + Content Library + Google Drive + Dropbox**, constrained to **images only** (no video/pdf, no Canva/Vista/PostNitro design tools).
- Picking from Content Library / Drive / Dropbox returns an already-hosted image URL → set it as the thumbnail directly (no re-upload for existing library items). A freshly uploaded/dragged file keeps the current upload-then-set path.
- Preserve the existing per-platform guards: Instagram image `max_size`, YouTube thumbnail `max_size`, and the YouTube-Shorts / TikTok / Instagram-reel-only warnings. For URL-based (library) selections, apply the equivalent size/type guard using the library asset's metadata rather than a `File` object.
- Keep selection **scoped to the thumbnail modal** — a chosen image must set the thumbnail for the current post/platform, not get attached to the post body.
- Keep the tab's existing constraints intact: the Upload Image option is already hidden for TikTok and gated by `showUploadImageOption` / per-account disabled states.

## UX Reference

This reuses an **existing** ContentStudio pattern — the composer media widget (`MediaSelection.vue`) already used in the editor. The target (per the user's second screenshot) is the compact, image-only variant: a drag-and-drop area with an "upload" link, a "Supported: JPG, PNG, GIF, WebP" line, and three source buttons — **Content Library**, **Google Drive**, **Dropbox**. No new interaction pattern is introduced; it is the same widget adapted to fit the modal's upload tab.

## Mobile Context

N/A. The per-platform custom video thumbnail editor (Suggested / From Video / Upload) is a **web composer** feature; the mobile apps do not expose this thumbnail editor. No iOS/Android/Flutter story.

## Files Involved

- `contentstudio-frontend/src/modules/composer_v2/components/CustomThumbnail/CustomThumbnailModal.vue` — replace the Upload Image dropzone; add library/Drive/Dropbox selection handling; adapt validations for URL-based picks.
- `contentstudio-frontend/src/modules/composer_v2/components/MediaSelection.vue` — reuse; likely needs an image-only / compact "thumbnail" mode (hide video/pdf + design tools, relabel to fit the modal).
- `contentstudio-frontend/src/modules/publish/components/media-library/composables/useMediaHelper` — existing `uploadFilesHelper`; reference for a thumbnail-scoped library selection callback.
- Locale files under `contentstudio-frontend/src/locales/*/composer.json` — any new/changed copy (namespace `composer`, existing `composer.custom_thumbnail_modal.*` and `composer.media_selection.*` keys).

## Notes / open considerations

- **Reuse vs. new compact widget:** preferred approach is to add an image-only/compact mode to `MediaSelection.vue` (hide video/pdf, hide Canva/Vista/PostNitro, constrain accepted types) and mount it inside the Upload tab, rather than building a parallel picker. Final layout to be confirmed in the companion design story.
- **Terminology:** the target screenshot labels the library button **Content Library**; the current widget says **Media Library**. Align on one label in the design story (they refer to the same library).
- **No backend change:** library assets and Drive/Dropbox imports already resolve to hosted URLs via existing flows; setting a thumbnail only needs the resulting URL. No new API.
