# Research — Media Notes follow-ups (CL tile menu + composer surfacing)

Follow-up changes to the shipped **Media Notes** feature. **All frontend.** "Anyone can edit" is already done by the dev — out of scope here.

## Current State

**Content Library tile** — `src/modules/publish/components/media-library/components/Asset.vue`
- A note button renders whenever the `showNote` prop is true (bottom-left, `bottom-2.5 left-2.5`). It shows in **both** the library grid (`MediaLibraryMain.vue`, `MediaLibraryTab.vue` pass `:show-note="true"`) and the pre-upload modal (`UploadFilesTab.vue` passes `:show-note="!isUploading"`).
- Styling: **filled yellow** (`bg-yellow-500`) when `hasNote`, **dark translucent** (`bg-[#2F3942]/40`) when there's no note — i.e. the icon is visible even with no note today. `hasNote = !isEmpty(info.note)`.
- The tile's **3-dot dropdown** (`Dropdown` at `bottom-2.5 right-2.5`) currently has items like "Add to composer", "Restore", etc. — **no note item**.
- The note modal is `MediaNoteModal.vue`; when persisting immediately (library) it calls `updateMediaNoteApi` → `media_library/assets/updateNote`. Editing already works and (per your note) is open to anyone.

**Content Library rich preview** — `PreviewMediaAssetModal.vue` (the media + right-hand File Info panel, including the read-only Note card). This is the modal we want the composer to reuse.

**Composer media tile** — `src/modules/composer/features/media/editor-media/MediaThumbnailItem.vue`
- Has a yellow **audio/music badge** (top area): `Icon name="Music"`, `bg-[#FFC555]` when `audio_track_id` is set, emits `audio-preview`. This is the badge the new note icon should sit next to.
- Clicking a media opens the **basic** preview via `useMediaLightbox.ts` (image → `vue-easy-lightbox`; file → `display-file-modal`) — **no info panel, no note.**
- The composer media object needs the asset's `note` (and asset `_id` / uploader) to (a) show the indicator and (b) feed the rich preview — this must be carried in when media is selected from the Content Library.

## What Needs to Change (all FE)

1. **CL tile — icon only when a note exists.** In the library grid, render the note icon **only when `hasNote`**; drop the always-visible "no note" state icon.
2. **CL tile — Add/Edit in the 3-dot menu.** Add a dropdown item: **"Add note"** when there's no note, **"Edit Note"** when one exists (opens the existing `MediaNoteModal`).
3. **Composer tile — note indicator.** On `MediaThumbnailItem.vue`, when a Content-Library-sourced media has a note, show a **yellow note icon next to the audio/music badge**.
4. **Composer preview = CL preview.** Clicking a media in the composer opens the **`PreviewMediaAssetModal`-style rich preview** (info panel + **read-only** note), not the basic lightbox.
5. **Plumbing.** Carry the note (+ asset `_id`/uploader) into the composer's media object when selected from the Content Library, so #3 and #4 have the data.

## Mobile Context
N/A — web-only. Notes aren't on the mobile apps.

## Files Involved
- `Asset.vue` — note-icon `v-if` (gate on `hasNote`), new "Add note"/"Edit Note" `DropdownItem` in the 3-dot menu.
- `MediaNoteModal.vue` — reused from the dropdown (no change expected beyond invocation).
- `MediaThumbnailItem.vue` (composer) — note indicator next to the audio badge.
- `useMediaLightbox.ts` / `useMediaItemActions.ts` (composer) — swap the preview trigger to the CL rich preview.
- `PreviewMediaAssetModal.vue` — reused by the composer preview (read-only note).
- Composer media-select flow — carry `note` + asset `_id`/uploader onto the media object.
- Locales (`publisher.media_tabs.*`) — the two new menu labels, in all 8 dirs.

## Suggested split
Two small FE stories: **(1) Content Library tile — note icon + menu**, **(2) Composer — note indicator + CL-style preview**. Can be combined into one FE story if you'd prefer.
