# Upload from Phone (Scan QR): Research

> **Sources:** the PO's product exploration chat (2026-09-23 to 2026-09-28), the finalized prototype canvas at https://claude.ai/artifact/FbSZdddf9VP63dHSUsBJGJ, and a codebase pass over `contentstudio-backend/` and `contentstudio-frontend/`.
>
> **Scope note:** the PO asked to start the pipeline at the workflow doc, so the competitor web sweep was not run. The competitive note from the exploration chat is recorded below as-is and still needs a quick check before anyone uses it in marketing.

---

# Part A: Product research

## 1. What is this feature?

A user is on ContentStudio on a laptop, and the photos and videos they want are on a phone. They open "From Phone" in the media upload modal. The modal shows a QR code and a link they can copy. They scan the code with the phone camera, choose files on a plain mobile web page, and a few seconds later the files appear in the modal on the laptop, already saved to the Content Library.

No app, no login on the phone, and no emailing files to themselves, using WhatsApp Web or syncing through Google Drive.

**Why it matters:** social media managers shoot most of their content on their phones (events, products, behind the scenes). Today they have to move those files to the desktop some other way before ContentStudio can use them. This removes that step.

**A second use case surfaced during design:** the link can be sent to someone else, for example a client, who uploads files straight into the workspace. That shaped the phone copy. It never says "your workspace" or "your library", and every error points the uploader at the person who shared the link.

## 2. Competitive note (unverified)

- **Canva** has an "upload from phone" QR flow in its Uploads panel, as far as we know.
- **Buffer, Hootsuite, Later, Sprout Social and Planable** rely on their mobile apps or cloud drives. None of them is known to have a desktop-to-phone QR bridge.
- This may be a small but noticeable advantage. Verify before claiming it in marketing.

## 3. Decisions settled during design

| Topic | Decision |
|---|---|
| Destination | The phone uploads **only** to the Content Library, into the folder picked in the modal's Folder selector. The web app decides what happens next. |
| Final action | It depends on where the modal was opened: Composer "Add N files to post", Content Library "Done", Automation "Use in automation", AI chat "Attach to chat". |
| Closing the modal | Changes nothing. Uploads keep going, a toast confirms, and reopening "From Phone" shows the same session. |
| Expiry | A fixed 10-minute window to connect a phone, with no selector. Once connected, the session stays open while in use. It ends on Done (phone), Disconnect (web), or after 15 minutes with no uploads. An upload in progress always finishes. |
| Link | A copyable short link sits under the QR code: "Can't scan? Copy the link and open it on any phone." It is the same one-time pass as the QR code. |
| Longer-lived share links | **Rejected.** The PO chose to keep it simple with the 10-minute window. The window may grow later. |
| Phone page copy | Written for anyone holding the link. It shows "Sending to [WORKSPACE NAME] / Shared by [SENDER NAME]", never a folder path or "your". |
| Role-based UI | None anywhere on the phone. We can't know who opened the link. |
| Security | Upload-only token, one active session per user, rate-limited per session, the same type and size checks as normal uploads. |
| Media handling | Remove GPS location data, apply EXIF orientation, convert HEIC to JPG and MOV to MP4, chunked resumable uploads. |
| Visibility | Hidden on mobile browsers and in the mobile app. Never hidden by role (PO, 2026-09-28). Shown in single-image pickers too. |
| Mobile app | No Flutter work in phase 1. |
| Out of scope | PIN on the phone page, phone-side editing or cropping, captions on the phone, a "trusted phone", opening the ContentStudio app from the QR code, a "Recent phone uploads" filter. |

## 4. Edge cases identified

| Case | Behavior |
|---|---|
| iPhone HEIC photos | Convert to JPG. iPhones save HEIC by default, so without this the feature fails for most iPhone users. |
| iPhone MOV videos | Convert to MP4. |
| Large video on mobile data | Chunked, resumable upload. |
| Phone screen locks or connection drops | The upload resumes from where it stopped. The phone shows "Connection lost" with Try again. |
| Code expires before anyone scans it | The web app shows "This code has expired" with "Get a new code". |
| Code expires mid-upload | The upload finishes. Expiry only blocks connecting a new phone. |
| Storage full | The phone blocks the files and tells the uploader to ask the sharer. The web app shows "Workspace storage is full" with Open Content Library. |
| More files than the platform allows | Save all to the library, attach up to the limit, and explain. |
| Mixed images and video where the platform doesn't allow it | The composer's existing validation handles it. |
| Modal or composer closed mid-session | Files still land in the Content Library. |
| Composer open in two tabs | The session belongs to the tab that created it. |
| Wrong person gets the link | Short expiry, upload-only access and Disconnect limit the risk. |

## 5. What to measure (from the exploration)

- Codes generated compared with codes scanned.
- The share of connected sessions that upload at least one file.
- Files and total size per session.
- Upload failure rate by device and browser.
- "From Phone" use compared with other media sources.

---

# Part B: Codebase analysis

Paths are relative to the repo root. FE is `contentstudio-frontend/`, BE is `contentstudio-backend/`.

## 1. The shared upload modal (FE)

- **Modal:** `FE/src/modules/publish/components/media-library/components/UploadMediaModal.vue`. It is a `CstuModal` with props `type` (`social` default, `library`, `evergreen`), `folders`, `limits`, `modalId` and `defaultFolderId`. The storage meter is at L136-171 and uses `getMediaLimitsHelper` from `composables/useMediaHelper.ts`.
- **Left menu:** `components/SideTabs.vue`.
  - Tabs are a hardcoded `tabItems` computed (L243+).
  - `tabValues` (L212-239) maps tabs to **index strings**, and the mapping differs by `type`: the `library` type has no Content Library tab, so later indices shift by one.
  - Callers open the modal by numeric `sideTabIndex`. **Append the new "From Phone" tab at the end** so no existing index moves.
- **Tab slots** are filled in `UploadMediaModal.vue` L24-129. Tab components live in `components/MediaTabs/` (`UploadFilesTab.vue`, `MediaLibraryTab.vue`, `DirectUploadFileTab.vue`, `FetchMediaUrlTab.vue`, `SearchMediaTab.vue`, `DropBoxMediaTab.vue`).
- **Folder selector:** `components/MediaUploadFooter.vue` wraps `components/UploadFolderSelect.vue`. The `v-model` value is `"<folderId>-<isRoot>"` or `uncategorized`.
- **Opening:** `EventBus.$emit('show-media-library-modal', { source, details, sideTabIndex, modalId })`, handled by `handleShowMediaLibraryModal` (`UploadMediaModal.vue` L297-323). The Content Library page calls `$cstuModal.show('upload-media-modal')` directly (`components/Header.vue:150`).
- **Mount points:**

  | Mount | Type | Modal id |
  |---|---|---|
  | `FE/src/Home.vue:1051` | social | `global-upload-media-modal` |
  | `FE/src/modules/composer/views/SocialModal.vue:256` | social | `upload-media-modal` |
  | `FE/src/modules/publish/components/media-library/MediaLibraryMain.vue:3` | library | default |
  | `FE/src/modules/automation/components/evergreen/create/EvergreenMain.vue:293` | evergreen | default |
  | `FE/src/modules/automation/components/csv/BulkUploadAutomationSave.vue:177` | evergreen | default |
  | `FE/src/modules/publisher/ai-content-library/components/brand-knowledge/MediaAssetsTab.vue:25` | library | `brand-upload-media-modal` |

- **Openers and `source` values:**
  - Composer: `useComposerDialogs.ts:97`, `useEditorBoxMedia.ts:430` (source is the platform or `common`), `useLiteEditorBox.ts:336`, `useEditorCarouselBox.ts:224` (`carousel`).
  - AI chat: `components/dashboard/ChatInput.vue:1948,2140` through `composables/useEditorBridge.ts:89` (`dashboard-chat`, `dashboard-chat-frame-start|end`).
  - Automation: `useBulkUploadAutomationSave.ts:1155` (`bulk-image-schedule`).
  - AI Content Library: `UploadsTab.vue:273`, `CustomGenerateForm.vue:341`.
  - Single-image pickers: `composables/useMediaLibraryPicker.ts` (workspace logo, brand logo, profile image, custom thumbnail).
  - AI Studio tools and video clips.
  - Inbox: nothing opens this modal.
- **Final action per context:** `MediaLibraryTab.vue` `handleInsert()` L675-950 routes by `source`. Composer sources fall through to `insertFile` in `composables/useMediaInsertion.ts:410`. The phone tab should reuse this routing, not duplicate it.
- **Today's Uploads tab** (`UploadFilesTab.vue:396-500`) doesn't attach anything. For `library` it closes and emits `refetch-folders`. Otherwise it switches to the Content Library tab so the user picks and inserts.
- **Single-image pickers** (logos, profile image, thumbnail) take one image. From Phone shows there too, with a "Use selected image" final button (see Part C).

## 2. Composer media box (FE)

- The chips row is `FE/src/modules/composer/components/MediaSourceToolbar.vue`: Content Library, Drive (`sideTabIndex: 9`), Dropbox (`8`), and a Canva/Vista dropdown hidden on white-label domains. It is used by `MediaSelection.vue:73`.
- A "From Phone" chip goes here and opens the modal at the new tab's index.

## 3. Media upload backend (BE)

- **Routes:**
  - `BE/routes/web/storage.php` (auth + `PermissionMiddleware`): `media_library/assets/{upload, uploadByLink, uploadByBytes, fetch, move, limits…}`.
  - `BE/routes/api.php:84-90`: `generate-signed-url`, `process-uploaded-gcs-media`, `upload-from-drive`.
  - `media_library/temp-upload` with `throttle:temp-upload`.
- **Controller:** `BE/app/Http/Controllers/Storage/MediaLibrary/MediaLibraryAssetsController.php`.
  - `uploadMedia` L55 is the multipart/proxy path, used on white-label domains.
  - `uploadMediaByBytes` L1697 converts HEIC at L1786.
  - `getMediaLimits` L1310.
  - `generateSignedUrls` L1927 checks quota and returns a GCS signed PUT URL, plus a resumable-init URL for files of 10 MB or more.
  - `processUploadedGCSMedia` L1995 resolves the folder, builds video thumbnails, saves via `MediaRepository::saveMedia` and dispatches `MediaAssetProcessingJob`.
- **Models:** `BE/app/Models/Utilities/Media.php` (collection `media`), `BE/app/Models/Storage/MediaLibraryFolders.php`, and `BE/app/Repository/Utilities/MediaRepository.php:77`.
- **Storage:** GCS through `GcsBucketResolver` and `GcsSignedUrlGenerator`. Paths are `media_library/{ws}/uncategorized/original/...` or `folder_route_gcs/{folderId}`.
- **Resumable uploads already exist:**
  - BE: `config/media.php:96` `resumable_upload_min_bytes`.
  - FE: `composables/useGCSUpload.ts` (resumable, chunk retry, proxy fallback) and `composables/useMediaUploadQueue.ts`.
- **Validation:** MIME allow-lists are in `BE/config/media.php` (jpeg, png, gif, heic, heif, avif, webp, mp4, quicktime, x-msvideo, x-m4v, pdf). **No per-file size cap was found.** Only the quota is checked.
- **Auth coupling:** both `generateSignedUrls` and `processUploadedGCSMedia` rely on `Auth::id()` for `user_id`. The phone path needs a session-token-authenticated variant that attributes files to the session creator (`user_id`, plus `super_admin_id` for global folders).

## 4. Media processing gaps

- **HEIC:** `converted_image_mimes` and `MediaLibrary::convertImageToJpegAndUpload` exist, but they are only called on the multipart and bytes paths. **The signed-URL path (`processUploadedGCSMedia`) stores HEIC raw.** The phone flow needs conversion on whichever path it uses.
- **Likely existing FE bug (not verified at runtime):** `UploadFilesTab.vue:356` has `selectedFiles[i].type || isHEICFile(...) ? 'image/heif' : …`. Operator precedence tags any file with a MIME type as `image/heif` in the preview item. Worth a separate look.
- **MOV to MP4:** `BE/app/Libraries/Media.php:368` `fetchMediaThumbnail` calls `TRANSCODE_API v2/transcode`, with `performCloudConvert` (L453) as the fallback. `.mov` is accepted. Whether it becomes MP4 depends on the external transcode service.
- **EXIF orientation:** `BE/app/Libraries/Helper.php:879` `correctImageOrientation` exists (used at L844) but not on the GCS path. Thumbnails come from imgproxy in `BE/app/Jobs/Storage/MediaAssetProcessingJob.php`.
- **GPS stripping:** **does not exist anywhere.** It must be built.
- **Virus scanning:** none for media today. The only VirusTotal use is for links (`BE/app/Libraries/Publish/LinkShortener.php:229`). "Same checks as normal uploads" therefore means type and size checks only.

## 5. Storage quota

- Computed in `BE/app/Libraries/Settings/SubscriptionLimits.php:222` `availableMediaLimits($workspaceId)`:
  - Plan `limits.media_storage` × stackable, plus addons × 1 GiB.
  - Usage is summed across all of the super admin's workspaces.
  - Returns `{used, available, total}`.
- It falls back to `Auth::user()`. A phone request with no login must pass the workspace id so the owner is resolved through `WorkspaceRepo::getWorkspaceBaseSubscription`.
- Enforced in `generateSignedUrls` (L1944-1952), which returns `{status:false, error:'workspace_limit_exceeds', …, data: limits}`.
- FE handling today (`UploadFilesTab.vue:441-462`): `planStore.setMediaStorageLimit`, the `media-storage-limits-exceeded-modal`, and an alert with `publisher.media_tabs.media_helpers.storage_full`.

## 6. Real-time

- The backend uses **Centrifugo**. Laravel broadcasting is unused.
- BE publishing: `BE/app/Libraries/Realtime/Realtime.php` `Realtime::broadcast($channel, $event, $data)`.
- Channel names: `BE/app/Libraries/Realtime/Channels.php`.
- Auth: `BE/app/Libraries/Realtime/CentrifugoAuthToken.php`, where `isChannelAllowedForUser` (L88-114) matches by namespace. The endpoint is `POST /realtime/auth`.
- FE: `FE/src/modules/common/services/realtime/` (`RealtimeService.ts` on `centrifuge` 5.2.2, `channels.ts`, `tokenManager.ts`, `subscriptionRegistry.ts`).
- **Closest precedent:** `FE/src/modules/publish/components/media-library/composables/useMediaImportRealtime.ts`, which consumes the `media-import:{ws}:{mediaId}` events published by `BE/app/Jobs/Storage/ImportDriveMediaJob.php:178-211`.
- **For this feature:**
  - Add a namespace such as `media-phone-upload:{ws}:{sessionId}` in both `Channels.php` and `channels.ts`.
  - Add an auth rule that allows only the session creator.
  - The phone never subscribes. It only calls the backend, which publishes.

## 7. Tokens, public routes, rate limiting

- **Encrypted, expiring token precedent:** `BE/app/Http/Controllers/Integrations/Platforms/ExternalLinkIntegrationController.php:137-139` does `encrypt(['id','user_id','expires_at'])` with a 30-minute TTL.
  - Public routes: `getLinkSecureStatus` and `verifyLinkPassword` (`BE/routes/web/integrations.php:62-65`).
  - FE route: `/easy-connect/:id`.
- **Other public flows:**
  - `/api-connect/:sessionToken` (`integrations.php:299-301`).
  - Planner share `/share/planner/:id` (`BE/routes/web/planner.php:22-31`).
  - Analytics share via `shareable.optional` / `shareable.auth` middleware.
- **FE guest routes** use `meta: { guest: true }` in `FE/src/router.ts` (the guard skips them at L861-864).
- **Laravel `signed` middleware** is registered (`BE/app/Http/Kernel.php:66`) but unused.
- **Rate limiters:** `BE/app/Providers/RouteServiceProvider.php`, including `temp-upload` L254 (keyed by user, then IP), `realtime-auth` L160 and `generic`. A new `phone-upload` limiter should key by session.
- **No short-link service for app URLs.** `LinkShortener` is for post links. A short code route such as `/u/<code>`, resolved by the session collection, is enough.

## 8. White-label

- FE detects white-label by hostname (`FE/src/config/api-utils.ts:1`). On white-label domains `apiUrl` is `origin + '/backend/'`.
- **Do not touch the synchronous bootstrap** in `FE/index.html:138-186` without asking.
- Branding comes from `useWhiteLabelApplication.ts` via `fetchWhiteLabelDomainDataApi(hostname)` (`FE/src/api/whitelabel.ts:201`). It is served by public `GET /whitelabel/getDomainDetails`, so it works on an unauthenticated page.
- Use `FE/src/composables/useBrandName.ts` for the product name in copy.
- **White-label uploads use the proxy path, not GCS signed URLs** (`UploadFilesTab.vue:420-430`). The phone page must follow the same branch, and the QR link must use the tenant's origin.

## 9. Permissions

- **There is no dedicated "can upload media" permission.**
  - BE `PermissionMiddleware` only checks workspace access and demo-workspace action blocks.
  - FE: `FE/src/modules/common/composables/usePermissions.ts` (role checks, approver branches at L199-227) and `FE/src/composables/usePermission.ts` (`WORKSPACE_ROLES`).
- So there is nothing to hide the option by. **PO decision (2026-09-28):** From Phone is never hidden by role. It is hidden only on mobile browsers and in the app.

## 10. Usermaven

- There are 51 `userMaven.track(` calls. The only media one is `media_note_added` (`MediaNoteModal.vue:149`).
- Nothing tracks uploads, Drive or Dropbox imports, or Content Library use. There is no event to reuse, so this feature defines new ones.

## 11. Mobile detection

- `FE/src/composables/useIsMobile.ts` is viewport-based (under 768px).
- User-agent checks are ad hoc (`useOpenApp.ts:21`, `useBulkUploadAutomationSave.ts:927`, `linkAffordance.ts:58`).
- Hiding "From Phone" on mobile browsers needs a shared user-agent check. Viewport width alone would also hide it for a narrow desktop window.

## 12. QR code

- There is no QR library in `FE/package.json`. Adding one is a new dependency and needs approval. It is also a component gap: `docs/ui-components.md` has no QR component.

## 13. Platform media limits

- The spec-driven engine is `FE/src/modules/composer/features/validation/` (`mediaLimits.ts` `checkMediaLimits`, `max_images` warnings L648-662, `validateAgainstSpec.ts`). Its backend source is `BE/config/socialAccountsValidationConfig.php`.
- `useMediaInsertion.insertFile` enforces one video at a time.
- The "attach up to the limit" behavior should read the same spec, not a new list.

## 14. Mobile app

- Not analyzed. The PO confirmed no Flutter work in phase 1, and the option is hidden inside the app.

---

# Part C: Conflicts and decisions for the workflow step

1. **Session model.** Build one new collection for upload sessions: token, short code, workspace, creator, folder, source context, status, `connect_by` and `last_activity_at`, plus counters. Nothing existing fits.
2. **Upload path for the phone.** The phone page should use the same two routes the web app uses (signed URL on contentstudio.io, proxy on white-label), authenticated by the session token instead of a login. Both need the HEIC conversion, orientation fix and GPS stripping that the signed-URL path lacks today.
3. **Where the phone option appears.** Every opener, including the single-image pickers. **PO decision (2026-09-28):** include them, with a "Use selected image" final button.
4. **Hiding by role.** **PO decision (2026-09-28):** no role-based hiding at all.
5. **Per-file size cap.** **PO decision (2026-09-28):** no cap, only the storage quota, matching desktop uploads.
6. **Virus scanning.** Normal uploads have none. Recommend keeping parity in v1 and noting the risk, since this is the first path where someone without a login can upload.
