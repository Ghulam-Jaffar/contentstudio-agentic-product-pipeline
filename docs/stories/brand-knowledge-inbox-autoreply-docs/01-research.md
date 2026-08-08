# Research — Manage Inbox auto-reply help docs from Brand Knowledge

Item 18 of the 7 Aug 2026 backlog batch.

## Current state

There are two separate document stores. A user who has already given us their documents in one has to upload them again to use them in the other.

### Store A: Brand Knowledge Source Materials

`contentstudio-frontend/src/modules/publisher/ai-content-library/components/brand-knowledge/SourceMaterialsTab.vue`

- Reads `props.profile.source_materials`.
- Source types are **website, social, document and text**. Documents are already a first-class Brand Knowledge source type.
- Each row shows name, type, last synced, added date, and per-row actions for sync and delete, plus an unreachable status.
- Adding sources goes through an add-source modal and appends non-destructively; the brand is enriched only on Sync.
- Sibling tabs: `BrandProfileTab.vue`, `BrandVoiceTab.vue`, `BrandStyleTab.vue`, `MediaAssetsTab.vue`, all under `BrandKnowledgeEditor.vue`.
- Queries: `modules/publisher/ai-content-library/queries/useBrandKnowledgeQueries.ts` (which already polls with `refetchInterval` while a profile is generating).

### Store B: Inbox auto-reply help docs

`contentstudio-frontend/src/modules/inbox-revamp/`

- `composables/useAutoReplies.ts` holds a separate `brandDocs` ref with its own API surface: `fetchBrandDocsApi`, `uploadBrandDocApi`, `deleteBrandDocApi`, and `fetchAffectedRulesApi` (which rules would break if a doc is deleted).
- Its own quota, tracked separately as `max_docs` / `current_docs`.
- `components/autoreplies/BrandHelpDocsModal.vue` is the management UI: upload with progress, file-size and time-left display, delete with a confirmation that warns which rules are affected. Accepted types: `.pdf, .txt, .md, .csv, .docx`.
- `components/autoreplies/DocSelector.vue` is the per-rule picker, with search and select-all.
- `composables/useAutoReplyForm.ts` stores the selection on the rule as `guidance_doc_ids`, gated behind a `showBrandDocs` toggle. It also reports `uses_brand_docs` in its tracking payload.
- Usermaven events already in place: `auto_reply_brand_doc_uploaded`, `auto_reply_brand_doc_deleted`.

### Existing copy

From `src/locales/en/inbox.json`:

- `inbox.auto_replies.form.brand_docs` = "Use Help Docs"
- `inbox.auto_replies.form.brand_docs_tooltip_with_docs` = "AI will reference your uploaded help documents — FAQs, policies, pricing guides — to give accurate, brand-specific answers."
- `inbox.auto_replies.form.brand_docs_tooltip_no_docs` = "No help documents uploaded yet. Upload your FAQs, policies, or product guides so AI can give more accurate answers."
- `inbox.auto_replies.form.brand_docs_manage_link` = "Manage Help Docs"
- `inbox.auto_replies.form.use_brand_docs` = "Use Brand Help Documents"
- `inbox.brand_help_docs.description` = "Manage documents that can be shared across all Auto-Replies rules. Documents use workspace media storage."

Note the two existing tooltips use em dashes, which is against current copy convention. Any story touching them should rewrite them.

### Where the docs are actually consumed

`social-inbox-manager/app/services/auto_reply/` is the engine: `engine.py`, `intent_checker.py`, `reply_executor.py`, `rule_cache.py`, `quota_checker.py`, plus `models.py` with the `EvaluationContext`. Rules are cached (`rule_cache.py`) and rule changes are propagated by `app/workers/auto_reply_rule_change_worker.py`, so whatever supplies the documents has to invalidate that cache when the document set changes.

## The gap

Brand Knowledge already accepts documents as a source type. Inbox auto-replies has its own upload, its own quota, its own management modal and its own delete-confirmation flow over a different set of documents. There is no relationship between the two. A user who set up Brand Knowledge with their policies and FAQs still sees "No help documents uploaded yet" in the auto-reply form.

## What needs to change

- Brand Knowledge documents become the source of truth for auto-reply help docs.
- Each Brand Knowledge document carries an explicit opt-in for Inbox auto-replies, because not every brand document should be quoted back to a customer in a public reply. The opt-in is per document, not global.
- The auto-reply rule form's document picker reads from Brand Knowledge documents that are opted in.
- The separate Inbox upload path either goes away or becomes a shortcut that creates a Brand Knowledge document source with the opt-in already set. The second is friendlier and keeps the Inbox flow intact.
- The auto-reply engine has to receive the opted-in document set, and the rule cache has to invalidate when the opt-in or the document set changes.
- Existing documents in the Inbox store need a migration decision, since users are actively relying on them today.

## Decisions needed before implementation

- **Migration.** Existing Inbox help docs must not stop working. Migrate them into Brand Knowledge as document sources with the opt-in set, or keep both paths readable during a transition? Migration is cleaner; a transition period is safer.
- **Quotas.** The Inbox store has its own `max_docs` limit and Brand Knowledge has its own source limits. One limit or two after unification, and does an opted-in document count against both?
- **Storage.** The existing Inbox description says documents use workspace media storage. Confirm Brand Knowledge documents use the same storage accounting so a unified document does not get counted twice.
- **Delete semantics.** The Inbox modal warns which rules are affected before deleting. After unification, deleting a Brand Knowledge document has to carry the same warning, since it is now also an Inbox dependency.
- **Whether the Inbox upload shortcut survives.** Recommended yes, since removing it makes the Inbox setup flow worse.

## Files involved

Frontend:
- `contentstudio-frontend/src/modules/publisher/ai-content-library/components/brand-knowledge/SourceMaterialsTab.vue`
- `contentstudio-frontend/src/modules/publisher/ai-content-library/components/brand-knowledge/BrandKnowledgeEditor.vue`
- `contentstudio-frontend/src/modules/publisher/ai-content-library/queries/useBrandKnowledgeQueries.ts`
- `contentstudio-frontend/src/modules/publisher/ai-content-library/types/brand-knowledge.ts`
- `contentstudio-frontend/src/modules/inbox-revamp/composables/useAutoReplies.ts`
- `contentstudio-frontend/src/modules/inbox-revamp/composables/useAutoReplyForm.ts`
- `contentstudio-frontend/src/modules/inbox-revamp/components/autoreplies/{BrandHelpDocsModal,DocSelector,AutoReplyForm}.vue`
- `contentstudio-frontend/src/locales/*/inbox.json`, `publisher.json`

Backend:
- Brand Knowledge source-material storage and the brand-docs endpoints the Inbox composable calls
- `contentstudio-backend/app/Jobs/AI/BrandKnowledgeGenerationJob.php` and the source enrichment path

Inbox service:
- `social-inbox-manager/app/services/auto_reply/{engine,intent_checker,reply_executor,rule_cache}.py`
- `social-inbox-manager/app/workers/auto_reply_rule_change_worker.py`

## Related existing work

- `docs/features/ai-auto-replies/` — the original auto-replies feature docs, including pricing.
- `docs/features/brand-knowledge-revamp/` — Brand Knowledge v2.
- `docs/stories/brand-knowledge-media-assets/` — the media-asset harvest epic, same Source Materials surface.

## Mobile

None. Brand Knowledge is web only, and auto-reply rule configuration is web only.
