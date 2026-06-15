# Brand Knowledge Revamp — Research

> **Feature:** Unify Brand Knowledge to **one brand per workspace**, restructure into 5 tabs (Brand Style, Brand Profile, Brand Voice, Source Materials, Brand Assets), add multi-source ingestion, integrate brand assets with the Media Library, replace the brand-voice/style dropdown with an on/off toggle across AI surfaces, and ship a migration notice + consolidation.
>
> **Locked product decisions (from PO):**
> 1. **Scope:** one brand **per workspace**.
> 2. **Migration rule:** auto-keep the **first-created** brand style + first-created brand voice per workspace; the rest are removed at rollout. The 7-day banner gives users time to clean up so their preferred one survives.
> 3. **No data export** — hard delete after the notice period.
> 4. **Brand Profile is new** — current product has no Brand Profile tab.
>
> **Design artifact (reference mockup, exemplary only):** https://claude.ai/design/p/c9cb3c18-d69f-48c8-9674-4d9a18413a8c?file=Brand+Knowledge.html

---

# Part A — Competitor & Industry Research

## 1. What is this feature?

A **Brand Knowledge / Brand Kit / Brand Voice** capability is a persistent store of a brand's identity — its visual elements (logo, colors, fonts), its written/verbal personality (tone, character, vocabulary), and its business context (positioning, audience, competitors) — that an AI content engine reads before generating anything, so output is "on-brand" by default rather than generic.

**Why users want it:**
- Generative AI defaults to bland, interchangeable output. The market-wide fix is **grounding the model in brand-specific data before generation**, so the user doesn't re-explain their brand every session.
- Removes per-post manual tone tweaking and enforces consistency across channels.
- Visual brand kits (logos/colors/fonts) eliminate repetitive design setup and keep assets reusable in one place.

**2026 trend:** AI content generation is now table stakes — the differentiator has shifted to *authenticity and brand fidelity*. Tools are racing to let AI *learn* a brand automatically (website scan, past-post analysis, document upload) rather than relying purely on manual setup. ContentStudio's revamp — unifying style + profile + voice + source ingestion + assets into one Brand Knowledge surface — is aligned with where the market is heading.

## 2. Competitor Analysis Table

| Competitor | Has Feature? | Single vs Multiple Brands | Source Ingestion (website / doc / social / text) | Brand Assets? (logo/color/font/media) | How Applied in AI Gen | Pricing Tier | Unique Differentiator |
|---|---|---|---|---|---|---|---|
| **Buffer** | Partial — AI Assistant with tone/voice *guidelines*, no formal brand kit | Per-channel guidelines (no distinct brand-voice objects) | Text/instructions only; channel-aware | No brand kit | Tone adjust + guidelines in composer; channel-aware | All plans incl. Free | Channel-aware rephrasing inside composer |
| **Hootsuite** | Yes — OwlyGPT/OwlyWriter adapts to brand | Tied to org/Hootsuite bio | Social (past posts + bio); real-time Talkwalker mentions | No reusable visual kit | Generates captions in brand voice; tone/style menus | Higher/enterprise tiers | Real-time social-listening data feeds the voice |
| **Publer** | **Yes — strong, dedicated Brand Voices** | **Multiple per workspace** (per client/account) | **Doc (up to 5 files), social account analysis incl. 30-day analytics, text** | No (voice-focused) | **Dropdown selector** to switch voices | Paid AI Assist | Trains on connected accounts' *analytics/performance* (30 days) |
| **Later** | Yes — Caption Writer learns tone | Tied to account/profiles | Social (learns from previous posts) | Visual brand kit framing for media | Casual/professional/custom tone selector | Paid plans | Auto-learns tone from prior posts; creator-focused |
| **Sprout Social** | Yes — AI Assist + brand voice settings | Account/profile level | Conversation history + text-defined voice | No reusable kit emphasized | Tone adjust; reply suggestions | Higher/enterprise tiers | Brand voice drives *inbox reply* suggestions |
| **Loomly** | No brand voice/kit; only white-label "Custom Branding" | N/A | N/A | Logo/favicon for white-label only | N/A (tone menus in optimizer) | Premium+ | "Brand kit" = white-label skin, not AI brand knowledge |
| **Sendible** | Partial — AI Assist keeps captions on-brand | Account level | Text-based | No reusable kit | On-brand suggestions in Smart Compose | Paid plans | Brand-consistency suggestions in compose box |
| **SocialBee** | **Yes — Copilot scans site to build brand profile** | **One brand per workspace** (workspaces gated by tier) | **Website URL scan** (name, tagline, services, tone, audience, language); Q&A | Content categories, not visual kit | Tone presets; per-post | Pro and up | Website scan generates *full strategy* + categories + starter posts |
| **Agorapulse** | Yes — Writing Assistant + Organization Context (beta) | Organization-level context | Text: brand voice, audience, **competitors** | No reusable visual kit | 7 tone options; spans Publishing + Analytics | Paid; Org Context beta | "Organization Context" includes competitors + spans analytics |
| **Metricool** | Yes — AI assistant learns/adapts tone | Account/brand level | Learns from your edits over time | No reusable kit | Tone selector; AI Pilot | Paid plans | Iterative learning — refines from your edits |
| **Jasper** | **Yes — flagship Brand Voice + Brand IQ** | **Multiple, tier-gated** (2 → unlimited) | **Upload writing samples**; style guide; knowledge base | Visual guidelines | Select voice per output; **flags off-brand tone** | Teams+ | Governance — *flags & corrects* off-brand output (Brand IQ) |
| **Writer.com** | **Yes — Voice profiles + Style Guide** | **Multiple** (per product/channel/audience) | **Paste text samples** (300+ words; 500+ recommended) | Terms lists + style guide (verbal) | Select voice profile; combines voice + terms + style | Enterprise | "Voices from examples beat manual descriptions" |
| **Copy.ai** | **Yes — Brand Voice + Infobase** | **Unlimited (Pro)** | **Paste text** (300+ words → "Analyze") | Infobase = company knowledge base | Select voice in Chat/Workflows; `@`-mention facts | Pro | Infobase: `@`-mention brand facts mid-generation |
| **Canva** | **Yes — Brand Kit + Builder (visual benchmark)** | **Up to 100 brand kits** | **Auto-extract from website URL or PDF** | **Yes — full visual kit:** logos, colors, fonts, photos, icons, templates | Applied in editor; Brand Controls restrict to approved assets | Pro+ | Best-in-class auto-extraction + Brand Controls |

## 3. Common Patterns

- **AI brand voice / tone control is universal** — firmly table stakes.
- **Tone presets in the composer** (friendly/professional/witty/funny) are the baseline UX.
- **Auto-learning is now expected** — leaders scan a **website URL** (SocialBee, Canva), analyze **connected social accounts / past posts** (Publer, Later, Hootsuite), or analyze **pasted/uploaded writing samples** (Jasper, Writer, Copy.ai). Pure manual entry is legacy.
- **"Voice from examples beats manual descriptions"** (Writer.com) — grounding-before-generation is consensus.
- **Voice applied via a persistent selector** — typically a dropdown to pick the active brand/voice, so it's "always on" for the active brand.
- **Visual brand kits are mostly the domain of design tools** (Canva). A social tool combining media library + visual identity *with* voice is comparatively differentiated.

## 4. Differentiators worth considering

- **Canva Brand Kit Builder** — auto-extracts logos, colors, fonts, *and* voice/guidelines from a URL/PDF in one step; **Brand Controls** lock designs to approved assets. Gold standard for auto-population + governance.
- **Jasper Brand IQ** — **flags off-brand output and recommends corrections**. Governance layer, not just a prompt prefix.
- **Publer** — trains brand voice on **connected accounts' actual performance analytics (30 days)**.
- **Copy.ai Infobase** — `@`-mentionable knowledge base of brand facts, separating *voice* (how it sounds) from *facts* (what's true) — maps to ContentStudio's Brand Voice vs Brand Profile split.
- **Writer.com** — granular voice profiles layered with terms + style guide.
- **Agorapulse Organization Context** — bakes **competitors** into brand context (relevant to our Brand Profile competitors field).

## 5. User Expectations

**Table stakes:** persisted brand voice the AI uses automatically; tone presets; at least one auto-learning source; voice applied across all AI surfaces; editable/refinable output.

**Delighters:** multi-source ingestion in one place (URL + doc + social + text — no single social competitor offers all four → best-in-class breadth for us); Canva-style full-kit auto-extraction from a URL; a true visual brand kit + reusable media library; off-brand flagging/governance; separation of voice from profile/facts; learning from performance data.

## 6. On the "one brand per workspace" decision

**Well aligned with the market — provided "workspace" is the brand-isolation unit.** Two market models:

1. **Workspace-scoped, one brand per workspace** (social-management norm): **SocialBee** is the direct match — "a workspace is designed for one brand," extra workspaces sold per tier. Buffer, Sprout, Later, Sendible, Metricool, Agorapulse effectively scope brand identity to account/org/workspace.
2. **Multiple brand-voice objects within one account** (AI-writing-tool / agency norm): Jasper, Copy.ai, Writer.com, and **Publer** (one voice per client *inside* a workspace).

ContentStudio already has **multiple workspaces** (like SocialBee), so "one brand per workspace" is *not* "one brand per account." An agency uses one workspace per client and gets one clean brand each — the standard pattern.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| Power users expecting Publer/Jasper-style multiple voices *within* one workspace feel constrained | Position **workspace = brand**; make creating/switching workspaces fast |
| Existing users with several brands in one workspace face consolidation | Guided migration + 7-day banner (this feature); keep first-created as canonical |
| Sub-brands / campaign voices | Address via tone adaptation under one voice, not separate brands (v2) |

**Net:** the single-brand-per-workspace model is the cleaner, more opinionated choice and matches the dominant pattern. Main exposure is the one-time migration — handled by the banner + consolidation story.

## 7. Recommended Approach for ContentStudio

1. **Keep "one brand per workspace"** — market-aligned (SocialBee precedent). Ship a guided migration (banner + auto-keep first-created).
2. **Map the five tabs to the proven voice / facts / visuals split:** Brand Voice = *how it sounds* (Jasper/Writer/Copy.ai); Brand Profile = *facts/context* (Copy.ai Infobase, Agorapulse Org Context); Brand Style + Brand Assets = *visual* layer (Canva benchmark).
3. **Make Source Materials the headline differentiator** — ship all four sources (website scrape + document + connected social + pasted text), each auto-populating Voice *and* Style. No single social competitor does all four. Default to *generate from sources, then refine* (Writer principle).
4. **Apply brand voice as an always-on, per-workspace default** — since there's one brand per workspace, a simple on/off toggle replaces the brand-switch dropdown (a UX win enabled by the one-brand model).
5. **Consider a governance/flagging layer (fast-follow)** — Jasper Brand IQ-style "this caption drifts from your brand voice."
6. **Don't over-build voice variants in v1** — one canonical voice with optional per-platform tone adaptation.

---

# Part B — Codebase Analysis

> **Headline:** the feature exists today as the **"AI Content Library"** (`contentstudio-frontend/src/modules/publisher/ai-content-library/`), persisted as `AiContentLibraryProfile` in MongoDB. It **already stores `styles[]` and `brand_voices[]` as arrays** (the "multiple brands" we're unifying), each item carrying an `id`, `name`, and `is_default` flag. Partial source-ingestion plumbing already exists. The AI agents already consume a single `brand_voice` object — so the agent side mostly needs call-site changes, not model rewrites.

## Existing Related Code

### Frontend (Vue 3 — `contentstudio-frontend/`)
- **Module:** `src/modules/publisher/ai-content-library/` — the Brand Knowledge subsystem.
- **Profile model & state:** `src/modules/publisher/ai-content-library/composables/useSetup.js` — manages `AIUserProfile` with `styles` (array) + `brand_voices` (array), plus website-URL and file-upload ingestion state.
- **API:** `src/api/ai-content-library.ts` — `fetchAiProfileApi()`, `analyzeBrandApi()`, `uploadBrandFileApi()`, `updateBrandingApi()`, `updateStrategyApi()`, `updateTopicsApi()`.
- **Selector (dropdown):** `src/modules/AI-tools/components/BrandVoiceSelector.vue` — two dropdowns (style + voice), supports multiple per workspace. Used in `AIChatMain.vue` / `ChatHeader.vue`.
- **Editors:** `components/editors/BrandVoiceEditor.vue`, `components/editors/StyleEditor.vue`, `components/modals/UpdateBrandVoiceModal.vue`, `components/form/PostSettingsForm.vue`.
- **Chat selection logic:** `src/composables/useAIChatActions.ts` (`setBrandVoiceId()`, `setStyleId()`, `resetBrandVoiceToDefaults()`), re-exported via `src/composables/useAIChat.ts`. Store: `src/stores/core/useAIChatStore.ts` (`selectedBrandVoiceId`, `selectedStyleId`, `defaultBrandVoiceId`, `defaultStyleId`).
- **Media Library:** `src/modules/publish/components/media-library/` (SideBar.vue, Folder.vue, Asset.vue) + `composables/useMediaLibrary.js`.

### Backend (Laravel 10, MongoDB — `contentstudio-backend/`)
- **Model:** `app/Models/Ai/AiContentLibrary/AiContentLibraryProfile.php` — `styles: array`, `brand_voices: array` (**currently multiple per workspace**). Each style: `id`, `name`, `logo`, `colors` (brand/background/text), `business_name`, `heading_font`, `body_font`, `is_default`. Each brand voice: `id`, `name`, `color`, `strategy` (voice, niche, audience, tone[], language[], emotion[], competitors[]), `topics`, `is_default`.
- **Controllers:** `app/Http/Controllers/AI/AiContentLibrary/AiContentLibraryProfileController.php`, `...AiContentLibraryPostController.php`. Request: `app/Http/Requests/AI/AiContentLibrary/UpdateProfileRequest.php`.
- **Helpers/Jobs:** `app/Helpers/Ai/ContentLibraryHelper.php`, `app/Helpers/Ai/AiChatHelper.php`, `app/Jobs/AI/BrandKnowledgeGenerationJob.php` (async brand-knowledge processing already exists).
- **Media Library:** `app/Models/Storage/MediaLibraryFolders.php` (folder model with `is_root`, `is_global`, `is_ai_folder`, `is_ai_video_folder`, `root_folder_id` flags), `app/Repository/Storage/MediaLibraryFoldersRepo.php`, `app/Http/Controllers/Storage/MediaLibrary/MediaLibraryAssetsController.php`.

### AI Agents (Python, Agno — `contentstudio-ai-agents/`)
- **Model:** `src/models/brand_voice.py` — `BrandVoice`, `BrandStrategy`, `BrandVoiceParser` with `to_prompt_context()`, `to_image_prompt_context()`, `extract_keywords()`, `get_color_palette()`, `get_visual_style_description()`.
- **Consumers:** `src/agents/content/caption_writer.py`, `src/agents/content/rss_post_generator.py` (enforces brand-compliance rules), `src/agents/image/image_generator.py`, `src/teams/router_team.py`, `src/memory/memory_agent.py`.
- **Source-ingestion agent:** `src/agents/tools/business_info_agent.py` — already analyzes brand info from documents/URLs.
- **Pattern:** brand voice is injected as a context block; brand takes priority over platform defaults. Agents already expect a **single** brand voice object.

## Reusable Components/Services
- **FE:** `BrandVoiceSelector.vue` (adapt to toggle), `useSetup.js` (ingestion state), `useAIChatActions.ts` (selection → toggle), media-library components and `useMediaLibrary.js`. Existing API helpers: `analyzeBrandApi()`, `uploadBrandFileApi()`.
- **BE:** `AiContentLibraryProfile` (MongoDB flexible schema — evolve arrays → single objects without a hard migration), controller/repo patterns, `MediaLibraryFolders` folder hierarchy (add a `is_brand_assets_folder` flag), `BrandKnowledgeGenerationJob` for async ingestion.
- **AI:** `BrandVoiceParser` + `to_prompt_context()` / `to_image_prompt_context()` ready to reuse; `business_info_agent.py` for source ingestion.

## Integration Points
1. **Brand Knowledge module** — new 5-tab structure; profile endpoint(s) move from arrays → single `style` + `brand_voice` + new `brand_profile` + `source_materials[]`. `useSetup.js` refactor.
2. **AI chat toggle** — `AIChatMain.vue` / `ChatHeader.vue` swap `BrandVoiceSelector` dropdowns for a single on/off toggle; `useAIChatActions.ts` adds `toggleBrandUsage()`; `useAIChatStore` adds `brandGuidanceEnabled: boolean`.
3. **Downstream consumers:** Inbox auto-replies (`src/modules/inbox-revamp/components/autoreplies/AutoReplyForm.vue`, `composables/useAutoReplyForm.ts`), Inbox message composer (`MessageComposer.vue`), AI Content Library post settings (`PostSettingsForm.vue`), Evergreen (`src/modules/automation/components/evergreen/create/GenerateVariationsModal.vue` — already a boolean flag), RSS (`rss_post_generator.py`).
4. **Media Library brand-assets folder** — `MediaLibraryFolders` flag + UI badge in `SideBar.vue`; `useMediaLibrary.js` filtering.

## Technical Considerations
- **Data model migration (multiple → one):** MongoDB flexible schema. Add single `style` / `brand_voice` / new `brand_profile` / `source_materials[]` fields; migrate **first-created** `styles[0]` / `brand_voices[0]` (per PO decision) into the singles; drop old arrays in a later release. No hard schema migration required, but a **backfill job** is needed to populate singles from the first-created array items.
- **Source ingestion:** website scrape (HTTP client/scraper → meta + text + images), document parsing (e.g. `smalot/pdfparser` for PDFs), connected-social-account analysis (fetch recent posts → analyze tone via agent). Reuse `BrandKnowledgeGenerationJob` for async processing + the "Update brand knowledge?" confirm modal.
- **AI memory for source materials (RESEARCH SPIKE):** today brand voice is structured JSON, **no vector storage**. Options: (a) Agno memory system (`src/memory/` — already integrated, lighter); (b) PostgreSQL `pgvector` (infra change). Open questions: chunking, embeddings model, retrieval at generation time, staleness/refresh, storage limits per workspace. → dedicated research ticket.
- **Caching:** profile cached in Pinia (FE) and Redis per workspace (BE, invalidate on update).

## Downstream Impact Inventory (every place that reads brand voice/style)

**Frontend**
| File | Usage | Change |
|---|---|---|
| `src/modules/AI-tools/components/BrandVoiceSelector.vue` | Two dropdowns (style + voice) | Replace with single on/off toggle |
| `src/modules/AI-tools/AIChatMain.vue`, `ChatHeader.vue` | Render selector | Use toggle |
| `src/composables/useAIChatActions.ts` (~386–411) | `setBrandVoiceId()`, `setStyleId()` | `toggleBrandUsage()` |
| `src/composables/useAIChat.ts` (~127–134) | Re-exports selection APIs | Update exports |
| `src/stores/core/useAIChatStore.ts` | `selected*`/`default*` ids | `brandGuidanceEnabled: boolean` |
| `src/modules/inbox-revamp/components/autoreplies/AutoReplyForm.vue` | Brand-voice dropdown | Use single profile brand voice |
| `src/modules/inbox-revamp/composables/useAutoReplyForm.ts` | Reads `brand_voices[]` | Read single `brand_voice` |
| `src/modules/inbox-revamp/components/MessageComposer.vue` | Finds default voice | Read single `brand_voice` |
| `src/modules/publisher/ai-content-library/components/form/PostSettingsForm.vue` | Style + voice dropdowns | Pre-populate from single brand; toggle/override |
| `src/modules/publisher/ai-content-library/composables/useSetup.js` | `styles[]`, `brand_voices[]` | `style{}`, `brand_voice{}`, `brand_profile{}`, `source_materials[]` |
| `components/editors/BrandVoiceEditor.vue`, `StyleEditor.vue` | Array CRUD | Single-brand editors |
| `src/modules/automation/components/evergreen/create/GenerateVariationsModal.vue` | `useBrandContent` flag | Mostly unchanged (already boolean) |
| `src/api/ai-content-library.ts` | Profile fetch/update | Return/accept single brand |

**Backend**
| File | Usage | Change |
|---|---|---|
| `app/Models/Ai/AiContentLibrary/AiContentLibraryProfile.php` | `styles[]`, `brand_voices[]` | Add single `style`, `brand_voice`, `brand_profile`, `source_materials[]`; deprecate arrays |
| `app/Http/Controllers/AI/AiContentLibrary/AiContentLibraryProfileController.php` | Profile CRUD | Single brand; new source-ingestion + brand-assets-folder endpoints |
| `app/Http/Requests/AI/AiContentLibrary/UpdateProfileRequest.php` | Validation | Single-brand rules |
| `app/Helpers/Ai/AiChatHelper.php`, `ContentLibraryHelper.php` | Build brand context | Single brand voice |
| `app/Models/Storage/MediaLibraryFolders.php`, `MediaLibraryFoldersRepo.php` | Folder model | Add `is_brand_assets_folder` |

**AI Agents**
| File | Usage | Change |
|---|---|---|
| `src/models/brand_voice.py` | `BrandVoice` model | Keep (single object); optional `source_materials` for memory |
| `src/agents/content/caption_writer.py`, `rss_post_generator.py`, `src/agents/image/image_generator.py` | Consume brand voice | Keep; update call sites to single brand |
| `src/teams/router_team.py`, caption router | Pass brand voice | Auto-fetch workspace brand if not provided |
| `src/agents/tools/business_info_agent.py` | Analyze docs/URLs | Reuse for source ingestion |

---

# Combined Takeaways for the Revamp

1. **We're collapsing arrays, not building from scratch.** The model, editors, AI consumers, and partial ingestion already exist — the work is (a) data-model unification + backfill, (b) the migration banner/notice, (c) new Source Materials + Brand Profile + Brand Assets tabs, (d) dropdown → toggle everywhere brand is consumed, (e) a memory research spike.
2. **The one-brand decision is market-validated** (SocialBee precedent) and actually *simplifies* the downstream UX (toggle instead of dropdown).
3. **Source Materials = our headline differentiator** — all four sources in one place beats every social competitor.
4. **AI agents are already single-brand-shaped** — least risky surface; mostly call-site changes.
5. **Memory/embedding for source materials is genuinely open** — Agno memory vs pgvector, chunking, retrieval, refresh — warrants a dedicated research ticket before committing to an approach.
6. **Mobile is out of scope** — AI brand features are web-only per platform rules.
