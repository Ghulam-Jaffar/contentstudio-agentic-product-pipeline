# Epic + stories — Media library management on the public API

**Priority: P2.** 3 stories. Nothing is pushed to any tracker.

---

## Epic: Media library management on the public API

The public API's media surface has exactly two operations: list media, and upload media. The product's media library has around forty, including folders, archive, move, delete, notes, brand-asset flagging, storage stats and limits, and several upload routes for different sources.

The practical effect is that an integration can put files into a customer's media library and then never touch them again. It cannot organize them into folders, cannot move them, cannot archive or delete them, and cannot find out how much storage is left before uploading. A high-volume integration therefore fills a customer's library with unstructured, undeletable files until the customer runs out of storage, at which point the integration gets an error it had no way to anticipate.

Two gaps stand out. **Delete** is the important one: an integration that can create but not remove is a one-way ratchet, and it is the single most requested capability of any upload API. **Limits** is the cheap one: one read endpoint that tells a caller how much storage is available turns an unpredictable failure into a check.

This epic gives the API the management side of the media library.

### Out of scope

- Media editing. The product's image editing routes are an interactive feature and are not exposed.
- Changing storage limits or how they are enforced.

### Stories

1. `[BE] Add media delete, archive, move and folder management to the public API`
2. `[BE] Add media metadata, storage stats and limits to the public API`
3. `[BE] Expose media library management on every developer surface`

---

## [BE] Add media delete, archive, move and folder management to the public API

### Description

As a developer uploading media through the API, I want to organize, move, archive and delete that media, so that my integration is not a one-way pipe that fills a customer's library with files nobody can manage.

### Workflow

1. The developer lists a workspace's media folders.
2. The developer creates a folder, renames it, or deletes it.
3. The developer uploads media into a chosen folder, or moves existing media between folders.
4. The developer archives media it no longer needs in the working set.
5. The developer deletes media permanently.
6. The developer lists media filtered by folder, type and archived state.

### Acceptance criteria

- [ ] A caller can list, create, rename and delete media folders.
- [ ] A caller can upload media directly into a chosen folder.
- [ ] A caller can move existing media between folders.
- [ ] A caller can archive and unarchive media.
- [ ] A caller can permanently delete media.
- [ ] Listing media supports filtering by folder, media type, and archived state, and the existing list endpoint keeps its current response shape and default behavior for callers who pass no filters.
- [ ] Deleting media that is used by a scheduled or published post behaves exactly as the product does, and the response makes clear what happened. A caller is never able to silently break a scheduled post.
- [ ] Deleting a folder that contains media behaves exactly as the product does, and the response states what happened to the contents.
- [ ] Authorization uses the existing permission model. A caller who cannot delete media in the product cannot delete it through the API.
- [ ] Every destructive operation is idempotent or clearly reports that the target no longer exists, rather than returning a generic error on a retry.
- [ ] Every endpoint appears in the OpenAPI document with examples, and the documentation is explicit about which operations are irreversible.

### Mock-ups

None. Backend-only story.

### Impact on existing data

Callers can now delete and archive media that previously could only be created through the API. This is the intended capability, and the safeguards above exist so it cannot silently break a scheduled post. No migration of existing media.

### Impact on other products

Media deleted or moved through the API must reflect correctly in the web app, the mobile app and the composer's media picker. The story must verify a post scheduled with media that is subsequently deleted behaves the same way it does when the deletion happens in the browser.

### Dependencies

- None for the endpoints.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (error messages must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Add media metadata, storage stats and limits to the public API

### Description

As a developer uploading media through the API, I want to know how much storage is available before I upload and to set metadata on what I upload, so that I can handle a full library gracefully instead of discovering it as a failed request.

### Workflow

1. The developer reads the workspace's storage limits and current usage before uploading.
2. The developer uploads media and sets a note on it.
3. The developer updates a note on existing media.
4. The developer flags media as a brand asset, or removes that flag.
5. When the library is near its limit, the developer's integration can warn or pause rather than fail.

### Acceptance criteria

- [ ] A caller can read the workspace's storage limit, current usage and remaining space.
- [ ] A caller can set and update a note on media.
- [ ] A caller can flag and unflag media as a brand asset.
- [ ] An upload that would exceed the storage limit is refused with a clear, specific error naming the limit and the current usage, not a generic failure.
- [ ] The storage figures returned match what the product shows for the same workspace, verified by comparison.
- [ ] Authorization uses the existing permission model.
- [ ] Plan gating is respected, and the limit reported reflects the workspace's actual plan including any addons.
- [ ] Every endpoint appears in the OpenAPI document with examples.

### Mock-ups

None. Backend-only story.

### Impact on existing data

None. Notes and brand-asset flags are set by callers as they would be in the product.

### Impact on other products

Notes and brand-asset flags set through the API must appear in the web and mobile apps. Brand assets feed AI generation, so the story must verify a brand asset flagged through the API is picked up there.

### Dependencies

- Related to the usage visibility epic, which is building a broader limits-and-consumption picture. The storage figures here should not become a second, divergent source of truth. Reuse whatever that epic establishes if it lands first.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (error messages must be translated)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Expose media library management on every developer surface

### Description

As a developer using ContentStudio through the CLI, an AI assistant, or a no-code tool, I want media management available there too, so that I can organize and clean up a customer's library from wherever I am already working.

### Workflow

1. The developer manages folders and media from the CLI, including bulk cleanup.
2. An AI assistant, through the MCP server, can find media, organize it, and report how much storage is left.
3. The developer builds a no-code automation that uploads media into a folder and cleans up old media on a schedule.
4. The developer finds all of it in the public documentation.

### Acceptance criteria

- [ ] Folder management, upload, move, archive, delete, note and brand-asset operations, and the storage read are available in the CLI.
- [ ] The agent skill covers the same capabilities.
- [ ] The MCP server exposes them as tools.
- [ ] The ChatGPT and Claude connectors expose the same capabilities, inherited through the MCP server rather than defined separately.
- [ ] Permanent deletion through an AI assistant requires the explicit confirmation the MCP epic defines for destructive actions, so an assistant cannot clear a customer's media library on a loose instruction.
- [ ] Zapier, Make and n8n expose at minimum upload into a folder, move, and archive. Permanent deletion in a no-code tool is either omitted or gated, and whichever is chosen is recorded with its reason.
- [ ] The public documentation covers the capability on every surface, and is explicit about which operations are irreversible.
- [ ] The parity check passes for media management across every surface.
- [ ] Every surface enforces the same authorization and storage limits as the REST endpoints.

### Mock-ups

None. No graphical UI in this story.

### Impact on existing data

None beyond what a caller deliberately deletes, which is the point of the capability.

### Impact on other products

Every developer surface is touched. Existing capabilities on those surfaces must be unaffected.

### Dependencies

- Depends on both endpoint stories in this epic.
- Depends on the developer surface parity contract epic.
- Depends on the MCP epic for how an assistant confirms a destructive action.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, no graphical UI in this story
- [ ] Multilingual support (surface-facing text and errors translated where supported)
- [ ] UI theming support — N/A, no graphical UI in this story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
