# Story: Remove blog-related saved prompts from AI chat

The AI chat "Saved Prompts" panel (with "Search prompts...", "Add custom prompt", and "Show favorites only") still lists default prompts about blogs. Blog publishing is being retired, so these should be removed. They render on the frontend and are sourced from the default prompts; remove them wherever they are defined (backend default prompts and/or any frontend defaults). User-created custom prompts must not be touched.

---

## [BE] Remove blog-related default prompts from the AI chat prompts source

### Description

The AI chat default (built-in) prompts include blog-related prompts. Remove the blog-related ones from the default prompts source so they no longer appear in the Saved Prompts panel for any user.

### Acceptance criteria

- [ ] Blog-related prompts are removed from the default AI chat prompts source.
- [ ] Non-blog default prompts are unchanged.
- [ ] User-created custom prompts are not affected.
- [ ] After the change, no blog-related default prompt is returned by the prompts API.
- [ ] If any users already have blog default prompts stored/marked, confirm the cleanup approach so they stop appearing.

### Impact on existing data

Default prompts data changes (removal of blog entries). Custom prompts untouched. Confirm whether existing stored/default-marked blog prompts need a data cleanup.

### Impact on other products

The prompts API feeds the web AI chat and any other AI-chat surface (including mobile AI chat). Removal applies everywhere the default prompts are served.

### Global quality and compliance checklist

- [ ] Mobile responsiveness (N/A, backend-only story)
- [ ] Multilingual support verified (prompt strings localized or handled)
- [ ] UI theming supported (N/A, backend-only story)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

- Default prompts are managed in `contentstudio-backend/app/Repository/AiChat/AiPromptRepo.php`. Remove the blog-related default prompt entries here (and any seed).
- The frontend fetches these via `@api/ai-chat` and renders them as `state.defaultPrompts` in `SavedPrompts.vue`.

---

## [FE] Ensure no blog prompts remain in the AI chat Saved Prompts UI

### Description

Verify the AI chat Saved Prompts panel shows no blog-related prompts after the backend change, and remove any blog-related prompts that are hardcoded on the frontend if present.

### Workflow

1. User opens the AI chat and the Saved Prompts panel.
2. The user sees no blog-related default prompt.
3. All other prompts and the user's custom prompts remain.

### Acceptance criteria

- [ ] The Saved Prompts panel shows no blog-related default prompt.
- [ ] Any blog-related prompt that was hardcoded on the frontend (if any) is removed.
- [ ] Non-blog prompts and user custom prompts are unchanged and still work.
- [ ] "Add custom prompt", "Show favorites only", and prompt search continue to work.
- [ ] No blog prompt reappears after refresh or in a new chat session.

### UI copy

None (removal only).

### Impact on existing data

None on the frontend (renders backend-provided prompts).

### Impact on other products

AI chat also exists on mobile; if it renders the same default prompts source, it is covered by the backend change.

### Dependencies

Depends on: **[BE] Remove blog-related default prompts from the AI chat prompts source** (if the blog prompts are backend-sourced).

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only)
- [ ] Multilingual support verified (translations available or fallback handled)
- [ ] UI theming supported (default and white-label) — no dark mode
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

- Saved Prompts UI: `contentstudio-frontend/src/modules/AI-tools/SavedPrompts.vue` (renders `state.defaultPrompts` plus custom prompts) and `SavedPromptsModal.vue`. Prompts are fetched via `@api/ai-chat`.
