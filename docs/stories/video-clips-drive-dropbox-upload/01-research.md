# Research: Add Google Drive & Dropbox Upload Options to Video Clips

## Current State

The **Video Clips** feature (`AI Tools > Instant Video Clips`) has a media import step (`StepVideoImport.vue`) with a dropzone that currently supports only **2 upload methods**:

1. **Local file upload** — drag & drop or click to browse (file input accepts `.mp4, .avi, .mov, .m4v`)
2. **Media Library** — opens the global media library modal via `EventBus.$emit('show-media-library-modal', { source: 'instant-video-clips', sideTabIndex: 1 })`

Additionally, there's a **URL import** section below an "OR" divider for fetching videos from a direct URL.

### How the Composer Does It (Reference Implementation)

The **Social Composer** uses `MediaSelection.vue` which has the full set of upload options in a horizontal button row:
- **Upload** (file input)
- **Media Library** (emits `media-action` with type `openMediaLibrary`)
- **Google Drive** — `EventBus.$emit('show-media-library-modal', { source: 'common', sideTabIndex: 9 })`
- **Dropbox** — `EventBus.$emit('show-media-library-modal', { source: 'common', sideTabIndex: 8 })`
- _(separator)_
- Canva (discontinued/disabled), Vista Create, PostNitro — **not needed here**

The Drive and Dropbox options simply open the **same global media library modal** with the appropriate side tab pre-selected. The media library's `SideTabs.vue` confirms:
- **Tab index 8** = Dropbox (`DropBoxMediaTab.vue`)
- **Tab index 9** = Google Drive

So the integration is straightforward — just emit the same `show-media-library-modal` EventBus event with the correct `sideTabIndex`. The selected media comes back through the same `add-media-from-media-library` event the component already listens to.

## What Needs to Change

In `StepVideoImport.vue`:
1. Add a **Google Drive** button next to the existing Media Library button
2. Add a **Dropbox** button next to the Google Drive button
3. Both buttons open the global media library modal with the correct `sideTabIndex` (9 for Drive, 8 for Dropbox)
4. The existing `onVideoFromLibrary` handler already processes media from the library modal — it should work for Drive/Dropbox selections too since they go through the same event

## All Locations Using This Upload Component Pattern

| Location | Component | Upload Options | Has Drive/Dropbox? |
|---|---|---|---|
| Social Composer (caption area) | `composer_v2/components/MediaSelection.vue` | Upload, Media Library, Drive, Dropbox, Canva, Vista, PostNitro | Yes |
| Video Clips (import step) | `AI-tools/video-clips/components/StepVideoImport.vue` | Upload, Media Library | **No** (this story) |
| AI Chat input | `AI-tools/ChatBox.vue` | Upload only | No |
| Composer Lite Editor | `composer_v2/components/EditorBox/LiteEditorBox.vue` | Media Library (inline button) | No |
| Composer Carousel Editor | `composer_v2/components/EditorBox/EditorCarouselBox.vue` | Media Library (inline) | No |
| Blog Image Section | `publish/components/posting/blog/ImageSection.vue` | Dropzone (legacy) | No |
| Media Library Upload Tab | `publish/components/media-library/components/MediaTabs/UploadFilesTab.vue` | File upload | N/A (is the library) |

## Files Involved

- `contentstudio-frontend/src/modules/AI-tools/video-clips/components/StepVideoImport.vue` — add Drive & Dropbox buttons
- `contentstudio-frontend/src/locales/en/*.json` — add i18n keys for tooltips (if not reusing composer keys)
- Asset images already exist: `@assets/img/composer/google-drive.svg` and `@assets/img/composer/dropbox.svg`
