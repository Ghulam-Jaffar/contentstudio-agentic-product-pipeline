# Story: Reuse the social composer editor in Evergreen automation

## [FE] Make the composer EditorBox and MediaBox reusable and use them in the Evergreen (recycle) modal

### Description

The post creator/editor inside the Evergreen (recycle) automation uses an older version of the editor than the one in the social composer. Extract the social composer's editor box and media box into a reusable, self-contained unit and use it inside the Evergreen create/edit modal, so creating, editing, and drafting posts there matches the main composer. Build it so the same unit can be dropped into other surfaces later with minimal wiring.

### Workflow

1. User opens the Evergreen automation and adds or edits a post.
2. The post editor is the same modern editor used in the social composer (same editing controls and media handling), instead of the older one.
3. The user creates, edits, or drafts the post with the same experience as the main composer.

### Acceptance criteria

- [ ] The social composer's editor box and media box are extracted into a reusable unit that can be mounted outside the main composer.
- [ ] The Evergreen create/edit modal uses this reusable editor and media box in place of the older editor.
- [ ] Creating, editing, and drafting a post in Evergreen works with the modern editor (text editing controls and media handling match the composer).
- [ ] The media box supports the same sources as the composer (upload, media library, and the other configured sources).
- [ ] The reusable unit is documented/structured so it can be embedded in another surface later by mounting the component and passing the post data in and out.
- [ ] Existing Evergreen behaviors (variations, scheduling/optimization, account selection) continue to work with the new editor.

### UI copy

Reuses the existing composer editor and media box copy. No new user-facing strings beyond what the composer already provides; any Evergreen-specific labels around the editor stay as they are today.

### Impact on existing data

None to the data model. Evergreen posts are created/edited as before; only the editing UI changes.

### Impact on other products

Web only. Mobile and Chrome extension: N/A.

### Dependencies

None. Related to the composer media widget work (the media box already supports library/Drive/Dropbox sources).

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only)
- [ ] Multilingual support verified (translations available or fallback handled)
- [ ] UI theming supported (default and white-label) — no dark mode
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

- Composer editor + media pieces to extract/reuse: `contentstudio-frontend/src/modules/composer_v2/components/EditorBox/EditorBox.vue`, `EditorBox/EditorMediaBox.vue`, and `MediaSelection.vue`.
- Evergreen create/edit surface to plug into: `contentstudio-frontend/src/modules/automation/components/evergreen/create/AddEvergreenPost.vue` (and `EvergreenMain.vue`).
- Per the composer_v2 module guide, `SocialModal.vue` and `MainComposer.vue` are large Options API monoliths. Extract the reusable editor as a focused `<script setup lang="ts">` component with clean props/emits (post content in, changes out) rather than pulling in the whole composer. Keep composer state coupling minimal so the unit is genuinely portable.
