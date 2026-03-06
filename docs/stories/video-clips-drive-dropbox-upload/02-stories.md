# Stories: Add Google Drive & Dropbox Upload Options to Video Clips

---

## [FE] Add Google Drive and Dropbox upload options to Video Clips import step

### Description:

As a ContentStudio user, I want to import videos from Google Drive and Dropbox in the Video Clips tool so that I can clip videos stored in my cloud accounts without downloading them first.

Currently, the Video Clips import step (`src/modules/AI-tools/video-clips/components/StepVideoImport.vue`) only offers two upload methods: local file upload and Media Library. The Social Composer's `MediaSelection.vue` already supports Google Drive and Dropbox via the global media library modal. This story adds the same two options to the Video Clips import step.

---

### Workflow:

1. User navigates to AI Tools > Instant Video Clips
2. In the import step, user sees the upload dropzone with three buttons below the upload area: **Media Library**, **Google Drive**, and **Dropbox**
3. User clicks **Google Drive** — the media library modal opens with the Google Drive tab pre-selected
4. User authenticates with Google Drive (if not already connected) and browses their files
5. User selects a video file and confirms — the video loads into the Video Clips import area, showing the thumbnail, filename, and "Video ready" status
6. Alternatively, user clicks **Dropbox** — the media library modal opens with the Dropbox tab pre-selected
7. User selects a video from Dropbox — same result as above
8. User can remove the selected video and pick a different one from any source (upload, Media Library, Drive, or Dropbox)

---

### Acceptance criteria:

- [ ] A **Google Drive** button appears next to the existing Media Library button in the Video Clips import step
- [ ] A **Dropbox** button appears next to the Google Drive button
- [ ] Clicking Google Drive opens the global media library modal with the Google Drive tab selected (sideTabIndex: 9)
- [ ] Clicking Dropbox opens the global media library modal with the Dropbox tab selected (sideTabIndex: 8)
- [ ] Selecting a video from Google Drive loads it into the import area with thumbnail, filename, and "Video ready" status — same as selecting from Media Library
- [ ] Selecting a video from Dropbox works identically to Google Drive selection
- [ ] Both buttons are visually disabled (opacity + no pointer events) while a video is uploading
- [ ] Both buttons use the existing SVG icons (`@assets/img/composer/google-drive.svg` and `@assets/img/composer/dropbox.svg`)
- [ ] Button styling matches the existing Media Library button pattern (pill-shaped, icon + label, hover state)
- [ ] All three buttons (Media Library, Google Drive, Dropbox) are horizontally aligned and wrap gracefully on smaller screens
- [ ] Tooltip text appears on hover for each button
- [ ] All new UI strings use i18n keys — no hardcoded text
- [ ] i18n keys are added to all 7 locale files (en, fr, de, es, it, el, zh)

---

### Mock-ups:

N/A — follows the same button pattern already used in the Social Composer's `MediaSelection.vue`. The three buttons should appear in a horizontal row below the upload area, matching the existing Media Library button's styling.

**Button layout (left to right):**
- Media Library (existing) — `bg-cs-ultra-violet` background, media icon
- Google Drive (new) — `bg-cs-ultra-blue` background, Google Drive icon
- Dropbox (new) — `bg-cs-ultra-green` background, Dropbox icon

---

### UI Copy:

**Google Drive button:**
- Label: `"Google Drive"`
- Tooltip: `"Select a video from your Google Drive account"`

**Dropbox button:**
- Label: `"Dropbox"`
- Tooltip: `"Select a video from your Dropbox account"`

**Existing Media Library button tooltip** (no change): Already uses `t('publisher.ai_content_library.uploads_tab.media_library_tooltip')`

---

### Impact on existing data:

None. No data model or API changes. This only adds two new UI buttons that open the existing media library modal with different tab indices.

---

### Impact on other products:

- **Mobile apps:** No impact — Video Clips is a web-only AI feature
- **Chrome extension:** No impact
- **White-label:** No impact — uses existing CSS variable-backed color classes and existing media library modal which already supports white-label theming

---

### Dependencies:

None. The Google Drive and Dropbox integrations already exist in the media library modal (`SideTabs.vue` tab indices 8 and 9). This story only adds entry points to those existing tabs from the Video Clips import step.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only) — N/A, Video Clips is a modal-based AI tool, not a responsive page
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension) — N/A, web-only AI feature
