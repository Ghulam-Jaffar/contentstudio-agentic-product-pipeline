# Brand Knowledge on Developer Surfaces — Research

> **Feature:** Expose ContentStudio's Brand Knowledge (Brand Style, Brand Profile, Brand Voice, Source Materials, Brand Assets) as a programmable resource — full CRUD on the public REST API, read access on the MCP server and CLI, a brand toolkit in `contentstudio-ai-agents`, and brand-aware answers in the in-product AI chat.
>
> **Researched:** 2026-09-20. ~35 web searches plus direct doc fetches for the market half; three codebase sweeps across `contentstudio-backend/`, `contentstudio-ai-agents/`, `contentstudio-frontend/` and `contentstudio-flutter/` for the technical half.
>
> **Prior art in this repo that constrains this epic:**
> - `docs/features/brand-knowledge-revamp/` — locks the product model: **one brand per workspace**, five tabs, async ingestion behind a consent gate (BR-4), Brand Assets as a Media Library folder (BR-5), AI brand features web-only (BR-8).
> - `docs/stories/public-api-parity-contract/` — a **P0** epic requiring every new capability to be defined once and propagated to all developer surfaces, rather than hand-copied per surface.

---

# Part A — Competitor & Industry Research

## 1. What is this feature?

**Brand knowledge as a programmable resource** means treating the things that make AI output sound and look like *your* brand as first-class API resources with stable IDs, CRUD semantics and permission scoping — rather than settings buried in a web form.

It decomposes into five resource families, which map almost exactly onto ContentStudio's revamped 5-tab model:

| Family | What it holds | Market analogue |
|---|---|---|
| **Brand profile** | Company facts, products, audiences, positioning | Jasper *Products* + *Audiences*; Copy.ai *Infobase*; Typeface *Brand Hub* |
| **Brand voice** | Tone, vocabulary, sample writing | Jasper *Voices*; Writer *Voice* + *Style guide*; Copy.ai *Brand Voice* |
| **Brand style** | Colors, fonts, logo usage rules | Canva *Brand Kit*; Frontify guidelines; Figma *Variables*; DTCG tokens |
| **Source materials** | Website crawl, uploaded docs, connected accounts, ingested into a retrievable store | Jasper *Knowledge*; Writer *Knowledge Graph*; Copy.ai *Infobase* |
| **Brand assets** | Logos, media, fonts as binaries | Canva *Assets*; Bynder / Frontify DAM |

**Why developers and agents want it:**

1. **Grounding.** An agent writing a post needs the voice, the claims it may not make, and the logo. Without an API a human pastes it in every time. Jasper markets directly against this: *"Jasper's MCP … connects your brand intelligence — voice, compliance, audiences — to the AI tools your teams already use"* ([jasper.ai/mcp](https://www.jasper.ai/mcp)).
2. **Provisioning at scale.** An agency onboarding 200 client workspaces cannot hand-fill 5 tabs × 200. This is the strongest argument for CRUD writes.
3. **Sync from a system of record.** Brand guidelines usually already live in Frontify / Bynder / Figma / Notion. An API lets ContentStudio consume that source of truth instead of becoming a fourth copy of it.
4. **Governance and audit.** "Which posts were generated under which voice version?" is only answerable if voices have IDs.

The 2026 industry framing is **"brand as code"** — guidelines shifting from PDFs to executable rules exported in formats AI systems can consume. Frontify calls its version *"machine-readable brand intelligence"* ([frontify.com/en/blog/frontify-mcp](https://www.frontify.com/en/blog/frontify-mcp)).

## 2. Competitor Analysis

### Angle 1 — Social media management platforms

| Competitor | Brand kit / voice feature? | Public API? | Brand exposed? | MCP? | CLI? | Tier gate | Notable |
|---|---|---|---|---|---|---|---|
| **Buffer** | Light AI Assistant, no brand object | Yes, GraphQL | **No** — no brand/voice node found | **Yes**, `mcp.buffer.com/mcp`, OAuth, **free on every plan** | Not verified | API free | MCP as acquisition wedge |
| **Hootsuite** | Yes — OwlyWriter → **Wisdom**, learns voice from past posts | Yes, REST + OAuth2 | **No** — dashboard-only | **Yes**, announced 24 Jun 2026 for ChatGPT/Claude/Gemini/Copilot | No | REST needs manual app approval | Brand context lives inside Wisdom, not a resource |
| **Publer** | AI Assist, no brand object | Yes, REST | **No** — narrowest surface of any competitor | None found | No | Business + Enterprise | Users, Workspaces, Accounts, Posts only |
| **Later** | Yes in-product | Effectively no — Influence reporting only | **No** | No | No | Influence product | No general scheduling/brand API |
| **Sprout Social** | Yes — AI Assist | Yes, REST | **No** | **Yes** — remote MCP + **Trellis** agent, Nov 2025 | No | Advanced (~$399/seat/mo) | AI brand features absent from API |
| **Loomly** | Yes (markets as "brand success platform") | **No public API at all** | — | No | No | — | Strongest brand language, zero developer surface |
| **Sendible** | Brand/client management for agencies | Yes, REST | **No** | No | No | Not verified | Dated auth (token in query param) |
| **SocialBee** | Yes — AI content, categories | **No public API**, stated explicitly | — | No | No | — | Zapier/Make only |
| **Agorapulse** | Brand-level org model | Yes but **read-only** analytics | **No** | **Yes**, `@agorapulse/mcp` beta | npm launcher | REST = Custom/enterprise; **MCP = all paid** | **MCP is strictly more capable than REST** — creates drafts, manages inbox; deliberately cannot auto-publish |
| **Metricool** | "Brands" = workspaces | Yes, REST | **No** — "brand" is an account container | **Yes**, `mcp-metricool` on PyPI | Python package | Advanced + Custom | — |

> **Angle 1 verdict: not one social media management platform exposes brand voice, brand kit or AI settings over a public API.** Several have the feature in-product; every one keeps it inside the dashboard. This is a genuinely open space in the category.

### Angle 2 — AI content & design platforms (the real benchmark)

| Platform | Brand feature | Exposed via API? | MCP? | Tier | Notable |
|---|---|---|---|---|---|
| **Jasper** | **Deepest — Jasper IQ:** Voices, Audiences, Style Guides, Knowledge, Products | **Yes, the reference implementation.** Voices **full CRUD**; Knowledge **CRUD + natural-language search**; Products full CRUD; Style Guides + Audiences **read-only**. Every object has an external ID usable in API calls | **Yes**, `mcp.jasper.ai`, **OAuth 2.0 or `X-API-KEY`** | Business / Tech Partners | **Voice creation accepts URL crawl, PDF/DOCX upload, or up to 8 pasted text samples** — exactly ContentStudio's Source Materials. Also ships an **API usage statistics** endpoint |
| **Writer.com** | Voice profiles + Style Guide | **Split.** Knowledge Graph API full CRUD + query; File API with **retry-failed-processing**. Style guides/terminology are **legacy** endpoints. **Voice profiles are UI-only** | Not verified | Enterprise | The *knowledge* half is programmable, the *voice* half is not |
| **Copy.ai** | Brand Voice + Infobase (RAG, 10 MB/item) | **Workflows only** — you can invoke a workflow that uses them, nothing more | Not verified | Enterprise | **The cautionary tale**: real brand features with no resource, no ID, no CRUD |
| **Canva** | Brand Kit (logos, colors, fonts, templates) | **Split, revealingly.** Connect API has Assets, Brand templates, Folders, Exports — **no brand-kits group** | **Yes**, `mcp.canva.com/mcp`. **`list-brand-kits` exists on MCP and nowhere else** — read-only | Pro+ | **The single most important finding: Canva shipped brand kits to MCP before REST.** Per-op limits: reads 100/min, writes 20/min, uploads 30/min |
| **Figma** | Variables = design tokens | **Yes** — query/create/update/delete | Yes (Dev Mode MCP) | Enterprise full seat | **Writes require an explicit publish step** before other files see them. Aligns with **DTCG v2025.10**, the first stable design-token spec |
| **Frontify** | Brand portal is the whole product | **Yes**, GraphQL — 19 queries, **139 mutations**, 640 types, OAuth 2.0 + webhooks | **Yes (beta)** — **52 tools in packs**; default "discovery" pack = **25 read-only**; writes are opt-in packs; **respects each user's existing permissions**; no extra cost | Free with app enabled | **Best-designed MCP surface found.** Exposes tone-of-voice rules and style guides alongside assets, and returns **copyright/licensing info in search results** |
| **Bynder** | DAM + brand guidelines | **Yes**, REST v4/v5, OAuth 2.0, dedicated Upload Assets API | Ambiguous — could not verify | Enterprise | Publishes `llms.txt` and serves any doc page as Markdown via a `.md` suffix |
| **Typeface** | Brand Hub / "Arc Graph" — continuously-updated brand knowledge store | Claims REST for identity discovery + programmatic asset ingestion; **endpoints could not be verified** | Claimed, unverified | Enterprise, sales-gated | Closest conceptual match to "brand knowledge" as a product |
| **Adobe Express / Firefly** | Custom Models trained on brand imagery | Generation endpoints; **no brand-kit CRUD resource verifiable** | Not verified | Enterprise | Brand consistency via **fine-tuning**, not a structured resource — a road not taken |
| **zeroheight** | Design-system docs, tokens, guidelines | Yes | **Yes**, read-only remote MCP, OAuth + SSO + RBAC | Trial/paid | Cleanest statement of the publish gate: *"hidden content stays hidden — agents work only from published, approved guidelines"* |

## 3. Common Patterns

**a) Brand knowledge is a flat set of named, ID-addressable sub-resources under a workspace — never one monolithic `/brand` blob.** Jasper splits into Voices / Audiences / Style Guides / Knowledge / Products; Frontify into brands → guidelines / libraries / assets. The consistent rule: **each sub-resource has its own ID, and generation calls reference IDs, never inline copies.**

**b) Read is universal, write is selective — and the split is by *risk*, not by resource.** Config-ish brand knowledge (voices, knowledge items, products) gets CRUD; governance-ish brand knowledge (style guides, approved guidelines) is read-only. Jasper: Voices/Knowledge/Products full CRUD, Style Guides/Audiences read-only. Frontify: 25 read-only by default, writes opt-in. zeroheight: read-only entirely.

**c) Auth: OAuth 2.0 for multi-tenant, API key for server-to-server, increasingly both.** Jasper's MCP accepts either. Buffer's is OAuth-only. ContentStudio's `X-API-Key` is fine for REST but is the weaker option for a remote MCP.

**d) Binary uploads are always a separate, asynchronous, multi-step flow — never a field on the brand object.** Canva returns a job; Writer returns `status: in_progress`. **Nobody accepts a logo as base64 inside a brand PUT.**

**e) Ingestion is a job with a status enum, and the vocabulary is near-identical everywhere.** Canva: `in_progress | success | failed`. Writer adds a **retry-failed-processing** endpoint. The generic pattern: `202 Accepted` + `Location` header + poll a status resource, optionally with `Retry-After` and percent complete.

**f) Voice creation accepts *sources*, not just structured fields.** Jasper takes a URL to crawl, a PDF/DOCX, or up to 8 pasted samples. **The contract is: POST a source → get a job → poll → the derived artifact appears.** This is industry-standard and is exactly ContentStudio's Source Materials tab.

**g) Per-plan gating is the norm and it is aggressive — except on MCP.** Publer Business+, Sprout Advanced, Agorapulse Custom, Jasper Business, Copy.ai Enterprise, Figma Enterprise. But Buffer's MCP is free on every plan, Frontify's costs nothing extra, and Agorapulse's MCP is on all paid plans while its REST is enterprise-only. **MCP is being used as acquisition, REST as monetization.**

**h) Rate limits are per-operation, with reads roughly 5× writes.** Canva MCP: 100/min reads, 20/min writes, 30/min uploads. ContentStudio's flat ~100/min is coarse by comparison.

**i) `llms.txt` is table stakes for developer docs.** Canva, Writer, Jasper and Bynder all publish one.

## 4. Differentiators worth considering

1. **Canva shipped brand kits to MCP before REST.** `list-brand-kits` is an MCP tool with no REST equivalent. The implicit reasoning: a brand kit's primary consumer is a *generator* needing a handle, not an integration needing to sync records. **Direct precedent for shipping brand read-only on MCP first and deferring REST writes.**
2. **Frontify's tool-pack architecture** — 52 tools, 25 read-only by default, writes as opt-in packs. The best answer found to tool-count explosion.
3. **Agorapulse's MCP creates drafts but cannot auto-publish.** A safety rail worth copying.
4. **zeroheight's publish gate** — draft brand knowledge is invisible to agents by default.
5. **Figma's explicit publish step** after writes — API changes don't affect consumers until published.
6. **Frontify returns copyright/licensing inside asset search results**, so an agent knows whether it may use a logo before it does. Usage rules should ride with the asset.
7. **Jasper's API usage statistics endpoint.** ContentStudio already meters API credits — exposing consumption is a cheap differentiator.
8. **Copy.ai as the cautionary tale.** Brand Voice and Infobase are substantial features reachable *only* by invoking a workflow. **That is where ContentStudio ends up if brand knowledge only ever becomes an implicit input to an existing generation endpoint** — which is precisely what the current `GET /ai/brand` two-boolean endpoint is.

## 5. MCP and agent-surface findings

### 5.1 Brand context ships as **tools**, essentially never as MCP **resources**

- Jasper MCP: **9 tools, zero resources** — `get-jasper-brand-voices`, `get-jasper-audiences`, `search-knowledge-base`, `get-jasper-style-guides`, `get-jasper-products`, `get-jasper-agents`, `run-jasper-agent`, `generate-content`, `upload-attachment`.
- Canva MCP: tools only. Frontify MCP: 52 tools in packs, no resources.

The published guidance says the opposite of what vendors do. The decision rule is about *control*: tools are for what **the model** decides to invoke; resources are for context **the user or host** deliberately attaches. By that rule a brand voice is exactly a resource. **Vendors ship tools anyway because MCP resource support across clients is uneven** — tools work everywhere.

> **Recommendation: ship tools as primary; optionally mirror as resources for hosts that support them. Never ship resources alone.**

### 5.2 Brand context is **explicitly fetched, not auto-injected**

**No vendor auto-injects brand context into every agent turn over MCP.** Every one requires an explicit tool call. Auto-injection happens only *inside* the vendor's own product — Hootsuite Wisdom, Copy.ai's `@`-mention, Writer linking a style guide to a voice.

> **Implication: the in-product AI chat and the external MCP server should behave differently.** In-product chat can auto-load the workspace's single brand (one brand per workspace removes all disambiguation). External MCP agents must ask.

The documented downside of the "brand context folder" pattern: *"context windows fill up with voice instructions, leaving less room for content."*

### 5.3 Token-budget handling — the most actionable guidance found

- **Search, don't list.** Anthropic's guidance is to implement search-focused tools rather than list-all tools. Jasper follows it precisely: small enumerable sets get `get-*`, the knowledge base gets **`search-knowledge-base`**.
- **Bound every response.** Default `max_results` small — 10 for search, 3-5 for content extraction.
- **Cursor pagination is in the MCP spec** — opaque `cursor` in, `nextCursor` out, default page 10-20 with `has_more`.
- **Progressive retrieval** — summaries first, full content only on request for a specific item.
- **Response-size caps are the server's job today.** A negotiated `max_response_bytes` capability is only a proposal.
- **Format matters** — CSV is ~29% cheaper than JSON for tabular data (~1,380 tokens saved at 100 rows). Relevant for color and font lists.
- **Offer verbosity controls** — `fields=[...]`, `format: compact`, `per_page`.
- **Prefix tools with their domain** (`brand_*`), and build evals before shipping.

### 5.4 Permission inheritance is the expected security model

Frontify: the MCP *"respects each user's existing permissions."* zeroheight: SSO, RBAC, unpublished content stays hidden. Canva: operations match the user's access level. **No vendor gives the MCP server its own elevated identity.**

### 5.5 Writes over MCP are deliberately fenced

Jasper's IQ context is entirely read-only over MCP **even though the REST API has full voice CRUD**. Frontify hides writes behind opt-in packs. Agorapulse creates drafts but cannot publish.

> **A very direct precedent: full CRUD on REST, read-only on MCP.**

## 6. Asset and file handling over API

**Canva — the pattern to copy most closely:**
- `POST /v1/asset-uploads`, `Content-Type: application/octet-stream` (raw body, not multipart)
- `Asset-Upload-Metadata: {"name_base64": "..."}` — name Base64-encoded to survive emoji in an HTTP header, max 50 chars unencoded
- Returns a **job ID**; poll → `in_progress | success | failed`
- Success returns id, type, dimensions, timestamps, **thumbnail URL**
- Failure returns typed codes: `file_too_big`, `import_failed`, `fetch_failed`
- 30 req/min per user; MCP also offers `upload-asset-from-url` as a URL-fetch alternative

**Writer — standard HTTP file semantics plus retry:**
- `POST /v1/files` with `Content-Disposition: attachment; filename=…`
- Optional `?graphId=` binds the file to a Knowledge Graph at upload time
- Response carries `status: in_progress` — async processing after upload
- Accepts PDF, DOC, DOCX, PPT, PPTX, JPG, PNG, EML, HTML, SRT, CSV, XLS, XLSX, MP3, MP4
- **An explicit retry-failed-processing endpoint** — files fail to parse often enough to warrant it

**Copy.ai:** 10 MB per Infobase item. **Presigned-URL three-step** is the generic alternative (request URL → PUT bytes to storage → finalize with etag); keeps bytes off the API tier at the cost of three round-trips.

## 7. User expectations

### Table stakes

| Expectation | Evidence |
|---|---|
| Every brand sub-object has a **stable ID usable in generation calls** | Jasper's entire design; Canva's `list-brand-kits` → `generate-design` handoff |
| **Read/list for all five families** over REST | Jasper, Frontify, Canva, Writer |
| **Workspace/tenant scoping** on every call | Metricool, Canva, Frontify |
| **Permission inheritance** — the API identity sees exactly what the user sees | Frontify, zeroheight, Canva state this explicitly |
| **Async upload with a pollable job** for any binary or ingestion | Canva, Writer |
| **Typed error codes** (`file_too_big`, `unsupported_format`, `fetch_failed`) | Canva |
| **Documented size limits and accepted MIME types** | Writer's list; Copy.ai's 10 MB |
| **Pagination with cursors**, small default page sizes | MCP spec |
| **An MCP server that can read brand context** | 9 of the vendors researched shipped one within ~18 months |
| **`llms.txt`** in the docs | Canva, Writer, Jasper, Bynder |

### Delighters

| Delighter | Who does it |
|---|---|
| **Create a brand voice from a URL / document / pasted samples via API** | **Jasper only.** Highest-value single feature to copy — it is precisely the Source Materials tab |
| **Natural-language search over brand knowledge** as a tool | Jasper, Writer, Typeface |
| **Usage rules travelling with the asset** (clear space, min size, licensing) | Frontify |
| **Draft/publish gate so agents only see approved brand content** | zeroheight, Figma |
| **Tool packs / progressive tool disclosure** | Frontify |
| **Both OAuth and API-key auth on MCP** | Jasper |
| **API usage/credit statistics endpoint** | Jasper |
| **Webhooks on brand changes** | Frontify, Copy.ai |
| **Export brand style in a portable standard** (DTCG) | Figma/DTCG ecosystem — **nobody does it for brand kits yet, open space** |
| **CLI that can also launch the MCP server** | Canva (`@canva/cli`) |
| **MCP free on all plans as an acquisition wedge** | Buffer, Frontify, Agorapulse |

### Anti-patterns to avoid

- Brand knowledge reachable only by invoking a workflow, with no addressable resource → **Copy.ai**.
- A brand-voice feature that exists only in the dashboard while the API ships posts → **the entire social category**, which is why this is the opportunity.
- Read-only REST while the MCP can write → **Agorapulse** (surface asymmetry in the wrong direction).
- A `get_all_brand_knowledge` mega-tool that dumps a 40-page guideline doc into context.

## 8. Recommended approach for ContentStudio

### 8.1 REST resource model

One brand per workspace is a large simplification: the brand ID is implied by the workspace, so paths can be singular and there is no collection to paginate. Mirror the five tabs exactly so API, UI and docs share one vocabulary.

> Paths below are adjusted to ContentStudio's actual convention (`workspaces/{workspace_id}/…`), which the market research did not know.

```
GET    /api/v1/workspaces/{workspace_id}/brand                    # whole brand, nested summaries
PATCH  /api/v1/workspaces/{workspace_id}/brand                    # enabled flag, name

GET|PUT /api/v1/workspaces/{workspace_id}/brand/profile           # singleton
GET|PUT /api/v1/workspaces/{workspace_id}/brand/style             # singleton
GET|PUT /api/v1/workspaces/{workspace_id}/brand/voice             # singleton (one brand per workspace)

GET    /api/v1/workspaces/{workspace_id}/brand/source-materials
POST   /api/v1/workspaces/{workspace_id}/brand/source-materials   # 202 + Location: .../brand/jobs/{id}
GET    /api/v1/workspaces/{workspace_id}/brand/source-materials/{id}
DELETE /api/v1/workspaces/{workspace_id}/brand/source-materials/{id}
POST   /api/v1/workspaces/{workspace_id}/brand/source-materials/{id}/reingest
GET    /api/v1/workspaces/{workspace_id}/brand/jobs/{job_id}      # queued|processing|succeeded|failed + error.code

GET    /api/v1/workspaces/{workspace_id}/brand/assets
POST   /api/v1/workspaces/{workspace_id}/brand/asset-uploads      # 202 + job; bytes OR source_url
GET    /api/v1/workspaces/{workspace_id}/brand/asset-uploads/{job_id}
GET|PATCH|DELETE /api/v1/workspaces/{workspace_id}/brand/assets/{asset_id}
```

Design decisions and their precedent:
- **`202 Accepted` + `Location` + status resource** for every ingestion and upload; statuses `queued / processing / succeeded / failed` plus a typed `error.code` (`file_too_big`, `unsupported_format`, `fetch_failed`, `crawl_blocked`). *(Canva, Writer.)*
- **Accept bytes *or* `source_url`** on asset upload — agents almost always hold a URL, and ContentStudio's existing `MediaUploadRequest` already does exactly this.
- **Brand derivable from sources**, not just structured fields — the feature that differentiates the surface. *(Jasper.)*
- **Usage rules live on the asset**, not a separate doc. *(Frontify.)*
- **Per-operation rate limits** rather than the current flat 100/min: reads 100/min, writes 20/min, ingestion 10-30/min. *(Canva's shape.)*
- **Webhooks** for `brand.updated` and `brand.source_material.ingested`.

### 8.2 What belongs on which surface

| Surface | Scope | Why |
|---|---|---|
| **REST** | **Everything, full CRUD** | Where agencies provision workspaces and where brand-sync integrations live |
| **MCP** | **Read-only, ~6 dedicated tools, no long-tail** | Strong consensus — Jasper, Canva, zeroheight all keep brand read-only on MCP |
| **CLI** | **Full CRUD plus `--wait` job polling and bulk import** | The CLI is where job polling stops being the caller's problem |
| **ai-agents toolkit** | **Read-only in v1**, gated writes later | Inherits the same risk logic; writes must go through the Agno confirmation gate |

Proposed MCP / toolkit tools:

| Tool | Notes |
|---|---|
| `brand_get_profile` | Compact — profile + style summary in one call |
| `brand_get_voice` | The voice body |
| `brand_get_style` | Colors, fonts, logo usage rules; compact format for token lists |
| `brand_search_knowledge` | **Search, not list**, over source materials. Default `max_results: 5`, snippets not full docs |
| `brand_list_assets` | Filterable by type; IDs + thumbnail URLs + usage rules, never bytes |

Rules to enforce: prefix every tool `brand_`; cap every response server-side; cursor pagination with `has_more`; support `fields` / `format: compact`; **do not** route brand through the existing `discover` / `execute_read` long-tail — brand context is high-frequency and belongs alongside the dedicated tools.

### 8.3 In-product AI chat

Different rules from external MCP, because one brand per workspace removes ambiguity:
- **Auto-load** brand profile + voice + style at session start. Every vendor auto-applies brand context inside its own product even though none does it over MCP.
- **Do not auto-load source materials.** Retrieve on demand via the same bounded search — otherwise the context window fills with guidelines and leaves no room for the content.
- **Surface which brand knowledge was used**, so users can trust and correct it.

### 8.4 The strategic point

**No social media management platform exposes brand voice or brand kit over any developer surface.** Buffer, Hootsuite, Sprout, SocialBee and Publer all have brand-voice-flavoured AI features and all keep that layer in the dashboard. The companies doing this well — Jasper, Frontify, Canva, Writer — are not competing for ContentStudio's customer.

ContentStudio already has the three things that take longest to build: a live public REST API, a shipped CLI, and an MCP server with 21 dedicated tools. Adding brand knowledge to those surfaces would make ContentStudio, as far as this research can determine, **the first social media management platform with programmable brand knowledge.** The read-only MCP tools are the wedge: they cost the least and carry no write risk.

---
# Part B — Codebase Analysis

> Grounded in `contentstudio-backend/`, `contentstudio-ai-agents/`, `contentstudio-frontend/`, `contentstudio-flutter/` as of 2026-09-20.

## 1. Headline finding: a locked decision stands against this epic

`contentstudio-backend/app/Data/Ai/PublicApi/BrandStatusResponseData.php` documents an explicitly locked product decision (**D2**) governing the one public brand endpoint that exists today:

- Route: `contentstudio-backend/routes/api/v1.php:898` — `GET workspaces/{workspace_id}/ai/brand` → `AiImageController::brand`.
- It returns **two booleans and nothing else**: `{configured, enabled}`.
- The docblock states the reasoning verbatim: brand arriving in a request body could be forged by any API key, so brand is always resolved **server-side**; and the wire contract (business_name / style / profile / voice / media_references) "would become a public contract a future brand revamp has to maintain."

**This epic directly reverses D2.** That reversal is a product decision the PO must take explicitly — it is not an implementation detail. The mitigations available are: version the public brand schema separately from the internal one, keep *writes* out of the generation path (brand is still resolved server-side at generation time, never accepted inline in a generate request), and gate the surface.

## 2. Brand Knowledge data model — one embedded document, no Brand entity

There is **no** `Brand`, `BrandVoice`, `BrandStyle`, `BrandProfile`, `BrandAsset` or `SourceMaterial` model. Everything is one embedded schema on a single model.

**`contentstudio-backend/app/Models/Ai/AiContentLibrary/AiContentLibraryProfile.php`** (182 lines, `MongoDB\Laravel\Eloquent\Model`, `SoftDeletes`) → collection **`ai_content_library_profiles`**. No migration creates it; there are no indexes defined.

Field groups (from `$attributes` defaults, the authoritative shape):

| Group | Fields |
|---|---|
| `brand_style` | `id`, `logo` (single sanitized raster URL), `colors[]` (`{hex, role}` where role ∈ brand/background/text/accent), `title_font`, `body_font`, `visual_identity_description` |
| `brand_profile` | `business_name`, `core_identity`, `market_positioning`, `competitors[]`, `competitive_advantages[]`, `primary_customer_segments[]`, `primary_value_drivers[]` |
| `brand_voice` | `id`, `purpose`, `audience`, `tone[]`, `emotion[]`, `character[]`, `language[]`, `voice_description` |
| `brand_topics` | `[{name, description}]` |
| `source_materials` | `[{id, name, type, url, content, content_url, social_accounts[], status, last_synced_at, added_at, use_for_auto_replies, auto_reply_index_status, auto_reply_vector_count}]` |
| `post_generation_settings` | `social_platform`, `language`, `no_of_posts`, `post_type`, `caption_length`, `emoji_usage`, `hashtag_usage`, `aspect_ratio`, `image_style` (+ vestigial `style`/`brand_voice`/`theme` pointers) |
| Flags | `brand_enabled`, `is_hosting_brand_assets`, `brand_assets_hosting_started_at`, `is_setup_complete`, `brand_migrated_at` |
| Legacy (CLN-1 removal) | `setup_details`, `business_name`, `company_name`, `styles[]`, `brand_voices[]` |

**Brand assets are media rows, not their own collection.** `App\Models\Utilities\Media` carries a boolean `brand_asset` flag plus a profile link. `contentstudio-backend/app/Repository/Utilities/MediaRepository.php`:
- `hostBrandImageFromUrl(...)` (line 614) stores to GCS at `ai_content_library/brand_assets/{workspaceId}/{profileId}/{sha1(sourceUrl)}.{ext}`, sets `brand_asset => true` (line 760), and is idempotent by source URL.
- `findMediaAssets()` supports a `brand_asset` filter (line 908) but **excludes brand assets from the normal media grid by default** (line 976).

> Note: this conflicts with brand-knowledge-revamp **BR-5**, which specifies Brand Assets as a per-workspace "Brand Assets" *folder* in the Media Library. Today it is a flag, not a folder. Whichever wins must be settled before the public asset contract is fixed.

### Caps — real, but enforced deep and silently

| Cap | Value | Where | Behavior when hit |
|---|---|---|---|
| Source materials | 50 | `AiContentLibraryProfileRepo::MAX_SOURCE_MATERIALS` | returns `null` → caller emits **404 "profile not found"** (wrong code) |
| Source content length | 100,000 chars | same repo | silent truncation |
| Brand assets stored | 100 | `HostBrandAssetsJob::TARGET_HOSTED` | harvest stops |
| Website sources actually scraped | 5 | `ContentLibraryHelper::MAX_WEBSITE_SOURCES` | **silent truncation** |
| Brand colors | 6 | `ContentLibraryHelper::MAX_BRAND_COLORS` | — |
| Tag/list item length | 200 chars | `ContentLibraryHelper::MAX_TAG_LENGTH` | truncation |
| Document upload | 50 MB, `pdf,txt,md,docx` | `UploadFileRequest` | 422 |
| Logo upload | 5 MB, `jpeg,jpg,png,gif` | `UploadLogoRequest` | 422 |

**No enum or allow-list exists** for `tone`, `emotion`, `character`, `language`, or `colors[].role`. They are free-form string arrays. `brand|background|text|accent` exists only as a code comment. A public contract has to decide: publish an enum, or publish free-form and document it.

## 3. Brand Knowledge has no public API surface at all

Every brand endpoint is in the **`web` middleware group** — session/JWT auth, not API-key auth.

`contentstudio-backend/routes/web/ai.php:96-140`, prefix `aiContentLibrary`, middleware `['auth','set.locale']` + `PermissionMiddleware`:

| Path | Controller method | Purpose |
|---|---|---|
| `GET profile/get` | `getProfile` | read whole profile |
| `DELETE profile/delete` | `deleteProfile` | delete brand |
| `POST profile/updateBrandSection` | `updateBrandSection` | partial-merge one of style/profile/voice |
| `POST profile/setBrandEnabled` | `setBrandEnabled` | on/off |
| `POST profile/sources/add` | `addSourceMaterial` | add source |
| `POST profile/sources/update` | `updateSourceMaterial` | rename/edit source |
| `DELETE profile/sources/delete` | `deleteSourceMaterial` | delete source |
| `POST profile/sources/sync` | `syncSources` | **destructive** re-derive from all sources |
| `POST profile/sources/enrich` | `enrichSources` | **non-destructive** blend |
| `POST profile/sources/autoReplyToggle` | `toggleSourceAutoReply` | index a doc for inbox auto-reply |
| `POST profile/sources/affectedRules` | `sourceAffectedRules` | which auto-reply rules use this source |
| `POST uploadfile` / `uploadlogo` | — | direct GCS upload, bypasses Media Library |
| `POST analyzeBrand` / `analyzeBrandStream` | — | build brand from sources (sync / SSE) |
| `POST profile/savePostGenerationSettings` | — | post-gen defaults |

Controller: `contentstudio-backend/app/Http/Controllers/AI/AiContentLibrary/AiContentLibraryProfileController.php` (854 lines).

Convention mismatch that makes this net-new work rather than a re-route:

| | Internal brand routes | Public API v1 |
|---|---|---|
| Auth | session (`auth`) | `X-API-Key` (`api.key`) |
| Path style | `aiContentLibrary/profile/updateBrandSection` (camelCase verbs) | `workspaces/{workspace_id}/<resource>` (REST nouns) |
| Workspace | in the **body** | in the **path** (authorized from the path) |
| Validation | classic FormRequests | spatie `laravel-data` DTOs (mandated) |
| Response | hand-rolled `{profile, sections, status}` | typed envelope |
| OpenAPI | none (scanner only reads `app/Http/Controllers/Api/V1`) | `@OA\*` annotations required |

## 4. Ingestion is synchronous today — and that is the hard blocker

- `analyzeBrand` / `analyzeBrandStream` — run **synchronously / SSE-streamed on the web request thread**.
- `syncSources` and `enrichSources` — **synchronous, inside the HTTP request**, then stamp sources `synced` or `unreachable`.
- Only image hosting is queued: `contentstudio-backend/app/Jobs/AI/HostBrandAssetsJob.php` (queue `media_library:processing_data`, `timeout=300`, `tries=1`, chunked self-dispatch of 25, ceiling 100 assets).
- `contentstudio-backend/app/Jobs/AI/BrandKnowledgeGenerationJob.php` (`timeout=600`, `tries=2`) is the onboarding path and already models progress tracking via `OnboardingBrandJob` (`status`, `progress`, `progress_message`, `retry_count`).

**The AI-agents call budget is 600 seconds.** `contentstudio-backend/config/ai_agents.php` sets `timeout` to 600 with a comment explaining the worst case for a single website source: scrape ≤60s + map + batch scrape ≤90s + up to 6 curate/vision rounds at ~45s ≈ 390s.

The public API throttles at 100 req/min and meters one API credit per workspace-scoped request. **No public HTTP request can hold open for 390-600s.** Async job + a status resource is mandatory work for this epic, not a nice-to-have. `OnboardingBrandJob` is the existing template for the job-status shape.

Kafka path for document embedding already exists: `contentstudio-backend/app/Services/Inbox/DocIndexEventPublisher.php` has `SOURCE_BRAND_KNOWLEDGE = 'brand_knowledge'`, publishes to the `doc_index` topic, consumed by VectorBridge → Pinecone, with a callback `PATCH internal/brand-knowledge-docs/{id}/indexed` (`routes/web/inbox.php:49-62`, `InternalApiMiddleware`).

## 5. Public API v1 conventions to copy

**Registration** — `routes/api.php:322` → `routes/api/v1.php`, prefix `/api/v1`. Main group (`v1.php:72`):

```php
Route::middleware(['force.json', 'api.key', 'api.request.log', 'throttle:api-v1'])->group(function () {
    Route::middleware(PermissionMiddleware::class)->group(function () { /* workspace-scoped */ });
});
```

AI routes use a stricter bucket and a nested prefix (`v1.php:864`) — the closest template for a brand resource:

```php
Route::middleware(['set.locale', PermissionMiddleware::class, 'throttle:ai-tools'])
    ->prefix('workspaces/{workspace_id}/ai')
    ->name('api.v1.workspace.ai.')
    ->group(function () { ... });
```

**Auth + metering** — `ApiKeyMiddleware`: key via `X-API-Key` / `Bearer` / `?api_key=`, SHA-256 hashed with `cs_` prefix. On workspace-scoped routes it also checks subscription, workspace lock, then **charges one API credit per request** (`API_ACCESS_NOT_ALLOWED` / `API_CREDIT_LIMIT_EXCEEDED`, both 403). Counter `used_api_credit` on the Workspace doc.

**Rate limits** — `config/api_rate_limits.php`: `default` 100/min, `ai_tools` 30/min (keyed on route `workspace_id`), `temp_uploads` 20/min. Error `RATE_LIMIT_EXCEEDED` (429) with `retry_after`.

**Envelope** — `app/Data/Concerns/HasTypedResponse.php`:
- success `{status: true, message, data: {...}}`
- error `{status: false, message, error_code, errors?}`
- validation → 422 with `error_code: VALIDATION_ERROR`
- **lists** follow `BaseController::transformPaginatedData()` (flat): `{status, message, current_page, per_page, total, last_page, from, to, data: []}`

**Workspace scoping** — `app/Support/RequestWorkspaceId.php`: the **path segment wins** and is what gets authorized; a contradicting query/header value is refused. `PermissionMiddleware` resolves via `WorkspaceAccessResolver`.

**Exemplar full CRUD** — Content Categories (`routes/api/v1.php:98-113`, `app/Http/Controllers/Api/V1/ContentCategoryController.php`, `app/Data/Settings/Requests/ContentCategoryStoreRequest.php`, `app/Http/Resources/Api/V1/ContentCategoryResource.php`, `app/Http/Controllers/Api/V1/ContentCategoryApiSchemas.php`). Resources emit both `id` and a deprecated `_id` alias.

**Media upload** — the model for brand assets. `MediaUploadRequest`: `url` **or** multipart `file`, mutually exclusive, plus optional `folder_id`; no size/mime rule at the request layer (enforced downstream against `SubscriptionLimits::availableMediaLimits`). `MediaController::upload` is a thin delegation to the internal SPA controller. Response `MediaResource` at HTTP 201.

**OpenAPI** — `darkaonline/l5-swagger` v8 + `scalar/laravel`. Scanner reads **only** `app/Http/Controllers/Api/V1`. Artifact `storage/api-docs/api-docs.json`, served at `/api-docs.json`, UI at `/api-docs` and `/scalar`. Per the backend's AGENTS.md §22 a new endpoint must update both the annotation and a hand-written `docs/api/<resource>-endpoint.md`.

## 6. No gating exists

**Not found anywhere:** a plan gate, entitlement check, or feature flag for Brand Knowledge. Access is purely workspace membership. AI credits are charged on post *generation*, not brand editing; storage quota applies on media upload. The generic flag infrastructure does exist (`app/Traits/Common/HasFeatureFlags.php`, `FeatureFlagMiddleware` aliased `feature.flag`, e.g. `User::FLAG_LINKEDIN_PROFILE_ANALYTICS`) — so a per-user rollout flag is available if wanted.

## 7. `contentstudio-ai-agents/` — the toolkit layer

MCP is gone. Per `tasks/agentic-architecture/INDEX.md` §0, `src/orchestration/mcp/` and `src/orchestration/workflows/` are empty and "the live surface is the toolkits in `src/integrations/contentstudio/toolkits/`".

**The envelope** — `src/integrations/contentstudio/envelope.py`, a function returning a JSON string:

```python
Status = Literal["ok", "partial", "no_data", "error"]
MAX_RESULT_BYTES = 8_000       # model-facing ceiling per tool result
MAX_UI_BYTES     = 64_000      # UI channel ceiling, skips model context
MAX_PERSISTED_BYTES = 32_000
LIST_PER_PAGE = 15
```

Body keys in order: `status`, `summary`, `data?`, `missing?`, `not_found?`, `needs_reconnect?`, `other_platform?`, `unconfirmed?`, `warnings?`, `pagination?`, `render?`, `source`, plus `truncated: true`. Any incompleteness list being non-empty **forces `status="partial"`**. Truncation is structural (halve lists, cap fields at 500 chars), never string slicing. Errors carry stable codes (`CS_UNAUTHORIZED`, `CS_NOT_FOUND`, `CS_INVALID_REQUEST`, `CS_RATE_LIMITED`, `CS_TIMEOUT`, …) from `errors.py`.

> **The 8 KB model-facing ceiling is a real design constraint for brand.** A fully populated brand profile plus voice plus 50 source materials plus 100 assets will not fit. A brand read tool must return a summary by default and detail on request.

**Existing toolkits** — `src/integrations/contentstudio/toolkits/`:

| Toolkit | Mode | Tools |
|---|---|---|
| `WorkspaceToolkit` (`cs_workspace`) | read | 6 — workspaces, accounts, groupings, team members, approval workflows, connectable platforms |
| `PublishingToolkit` (`cs_publishing`) | read | 3 — media list, posts list, internal notes |
| `PublishingWriteToolkit` | write, all gated | 6 — create/update/delete/approve posts, create from plan, add note |
| `OrganisationWriteToolkit` | write, 10 of 11 gated | 11 — groupings CRUD, workspace CRUD, team invite/remove, connect account, media import |
| `AnalyticsToolkit` | read | 7 |
| `InboxToolkit` | read + write | 9, 2 gated |

**Registration is three files:** a new `toolkits/<name>.py` `Toolkit` subclass → export in `toolkits/__init__.py` → attach in `src/orchestration/agents.py::create_member_agents()` (line 1817).

**Hard constraint D-07 caps a member at ~13-14 model-visible tools.** Workspace Data sits at 13, Operations at 12-13 — which is exactly why Analytics and Inbox each became their own member. A brand toolkit with more than one or two tools needs **its own team member**, not an addition to an existing one.

**Write gating** is Agno's native confirmation gate — `requires_confirmation_tools` on the Toolkit sets `Function.requires_confirmation`, Agno pauses *before* the body runs and hands back the exact arguments, so what the user approves is what executes. `hitl.py` (1101 lines) persists the paused run in the Neon session row (`PENDING_KEY`, TTL 1800s) and batches several pending writes into one confirmation card.

**Auth** — `src/integrations/contentstudio/client.py` holds **no credential**; every call takes a `WorkspaceScope` (frozen dataclass: api_key, workspace_id, timezone, locale, actor_user_id). `build_scope(metadata)` reads the caller's key and **deliberately refuses to fall back to the service key** (`key_source()` returns `service_key_withheld`) because `workspace_id` is caller-asserted. Scope never reaches the model. Base URL `https://api.contentstudio.io/api/v1`, timeout 45s.

> **Consequence:** the ai-agents toolkit talks to the *public* API with the *user's* API key. So a brand toolkit is strictly downstream of the public REST endpoints — it cannot ship first, and it inherits the 45s client timeout, which again forces async for ingestion.

**Brand today in ai-agents:**
- **Produced here.** `src/api/routers/business_info/business_info_router.py` (~1900 LOC, mounted at `/api/v1/business-info`) scrapes a site (Firecrawl) + documents and runs six agents from `src/agents/tools/business_info_agent.py` (identity, voice, style, topics, competitors, source language), harvests and curates brand assets, and detects the brand's own language.
- **Consumed as a prompt block, never as a tool.** `src/models/brand_voice.py` (~780 LOC) parses `metadata.brand_guidelines` pushed in by the PHP backend and renders a `=== BRAND VOICE GUIDELINES ===` block. Injection at `src/api/routers/streaming_router.py:1576-1627`, then into `team_session_state["brand_voice_block"]` which substitutes into `{brand_voice_block}` placeholders in member instructions. `src/agents/brand/induction.py` is the shared "brand-induction seam".
- `_BRAND_GROUNDING` (`agents.py:1703`) is the only brand-aware prompt fragment, shared across creative members only.
- **Not found:** any brand tool, toolkit, client method, or brand path in the 148-path `api-docs.json` snapshot; any brand entry in the architecture decision register (D-01…D-26).

**Net:** an agent today **cannot read brand knowledge**. Brand arrives as an opaque blob for creative members. Workspace Data, Analytics, Inbox and Operations members never see it.

## 8. `contentstudio-frontend/` — AI chat and the Brand Knowledge UI

**Brand Knowledge settings UI** — `src/modules/publisher/ai-content-library/components/brand-knowledge/`: `BrandKnowledgeEditor.vue` plus `BrandStyleTab.vue`, `BrandProfileTab.vue`, `BrandVoiceTab.vue`, `SourceMaterialsTab.vue`, `MediaAssetsTab.vue`. Route `brand-settings` (`src/modules/setting/config/routes/setting.ts:213`). Canonical TS shape in `src/modules/publisher/ai-content-library/types/brand-knowledge.ts`. Queries in `queries/useBrandKnowledgeQueries.ts`; API functions in `src/api/ai-content-library.ts`; URLs in **`src/modules/publisher/config/api-utils.ts`** (not the root `src/config/api-utils.ts`).

**AI chat module** — `src/modules/AI-tools/` (there is no `ai-chat` directory). Store `src/stores/core/useAIChatStore.ts` (Pinia, `defineStore('aiChat')`). Engine `composables/useAiChatEngine.ts`, transport `composables/useAIChatStream.ts` (SSE via `fetchEventSource`).

**Stream protocol v2** — `src/modules/AI-tools/types/stream-frames.ts`: 8 closed wire types (`run.started/finished/error/cancelled`, `block.start/delta/end`, `ui.suggestions`), `BlockStatus`, `BlockVisibility`, and the render switch:

```ts
export type RenderComponent =
  | 'prose' | 'step' | 'table' | 'chart' | 'metrics'
  | 'asset' | 'preview' | 'sources' | 'question'
```

Registry in `components/timeline/AiBlock.vue` has **ten** arms (the nine above plus `skill_draft`), with two fallbacks: no `render` → `AiStepRow`, unrecognised component → `AiProse`. `preview` and `question` emit `confirm`; `table` emits `rowAction` and `nextPage`.

> **A brand card can ride on the existing `table` / `metrics` / `preview` arms with no client change**, because table columns are derived from the rows present. A dedicated `brand` arm would require a coordinated change in `render.py`'s `COMPONENTS`, the TS union, the Vue registry **and** the Flutter switch — worth avoiding in v1.

**What the chat knows about brand today: an on/off switch and two opaque ids.** `components/BrandVoiceSelector.vue` is a single toggle. `utils/buildChatStreamPayload.ts:145-150` sends the entirety of it:

```ts
...(ctx.brandVoice?.enabled && { brand_voice_id: ctx.brandVoice.id ?? null }),
...(ctx.style?.enabled && { style_id: ctx.style.id ?? null }),
```

The PHP backend resolves those ids into `metadata.brand_guidelines`. **Not found in the chat:** any brand fetch, any brand render arm, any "edit my brand" affordance, any display of brand profile / source materials / brand assets in the conversation.

## 9. `contentstudio-flutter/` — full timeline parity, zero brand awareness

`lib/features/ai_assistant/` (69 files, clean architecture, Riverpod). Endpoints `data/ai_assistant_endpoints.dart` declare `streamProtocolVersion = 2`, explicitly aligned with the web app's `buildChatStreamPayload.ts`.

Render registry `presentation/widgets/timeline/ai_block_view.dart` switches on the same closed set of nine components with the same two fallbacks (`skill_draft` is absent and degrades to prose). Gates are answered through `answerGate(GateReply)` and `runRowAction(RowActionReply)` on `AiAssistantController`.

**Brand awareness: none.** `grep -rni brand lib/features/ai_assistant/` returns only incidental hits ("a brand-new chat", theme colour) plus one display-label map in `ai_preview_card.dart:37-38` (`'brand_asset' → 'source_brand_asset'`) that labels a preview when the image pipeline used a brand asset. It is a localisation lookup, not brand knowledge.

> Because the Flutter timeline already renders the same nine components, **any brand result that rides existing arms appears on mobile for free.** The brand-knowledge-revamp PRD's **BR-8 ("AI brand features are web-only")** therefore needs an explicit ruling for this epic: if a brand read tool lands on the Workspace Data member, a mobile user asking "what's my brand voice?" will get an answer whether or not we intended it.

## 10. Integration points summary

| Surface | Where it plugs in | New or extend |
|---|---|---|
| Public REST | `contentstudio-backend/routes/api/v1.php`, new `app/Http/Controllers/Api/V1/Brand*Controller.php`, DTOs under `app/Data/`, resources under `app/Http/Resources/Api/V1/`, `@OA` schema class | **New** |
| Async ingestion | new job modeled on `BrandKnowledgeGenerationJob` + a status resource modeled on `OnboardingBrandJob` | **New** |
| Brand assets | `MediaRepository` `brand_asset` flag + `MediaController::upload` pattern | Extend |
| OpenAPI | `storage/api-docs/api-docs.json` via l5-swagger + `docs/api/<resource>-endpoint.md` | Extend |
| MCP | `github.com/d4interactive/contentstudio-mcp` — **not mounted in this repo**; adds tools over the public API | **New, external repo** |
| CLI | `npm contentstudio-cli`, repo `github.com/contentstudioio/contentstudio-agent` — **not mounted**; colon-syntax commands | **New, external repo** |
| ai-agents toolkit | new `src/integrations/contentstudio/toolkits/brand.py` + export + attach in `agents.py::create_member_agents()` | **New** |
| FE AI chat | `src/modules/AI-tools/` — ride existing `table`/`metrics`/`preview` arms; no new render component in v1 | Extend |
| Flutter | inherits automatically if the result rides existing arms | None (policy decision only) |

## 11. Technical considerations and risks from the code

1. **D2 reversal** — the locked "status only, never brand content" decision must be explicitly overturned by the PO.
2. **Async ingestion is mandatory**, not optional: 600s backend budget vs 45s ai-agents client timeout vs a 100/min public API.
3. **8 KB envelope ceiling** in ai-agents forces a summary/detail split on any brand read tool.
4. **D-07 tool ceiling** means a brand toolkit likely needs its own team member.
5. **Two external repos** (MCP, CLI) are not mounted here and cannot be analyzed or implemented from this workspace.
6. **The parity contract epic** (`docs/stories/public-api-parity-contract/`, P0) says capabilities should be defined once and generated onto every surface. Building brand by hand on REST + MCP + CLI is exactly the six-tickets-per-change pattern that epic exists to kill. Sequencing decision required.
7. **Brand assets: flag vs folder** — code uses a `brand_asset` boolean; brand-knowledge-revamp BR-5 specifies a Media Library folder. Unresolved.
8. **No enums** for tone/emotion/character/language/color role — a public contract must decide free-form vs enumerated.
9. **Silent caps** (50 sources, 5 websites scraped, 100 assets) become API lies unless surfaced as explicit errors. The 50-source cap currently returns a misleading 404.
10. **Credential model** — ai-agents refuses to use a service key against a caller-asserted workspace. Brand tools inherit this: the user's own API key, or nothing.
11. **Write safety** — brand writes through an agent must go through the Agno confirmation gate, same as every other write toolkit.
12. **No indexes** on `ai_content_library_profiles`; a list-brands-across-workspaces public read would have no index support.

---

# Part C — Synthesis and Decisions Needed

## 1. Where market and codebase agree

The market's recommended shape and ContentStudio's actual constraints point the same way on four things, which is unusually convenient:

| Market says | Codebase says | Conclusion |
|---|---|---|
| Ingestion must be `202` + job + poll | Ingestion is synchronous today against a 600s budget, behind a 100/min API and a 45s agent client timeout | **Async job + status resource is mandatory.** Not a v2 nicety |
| Binary upload is always a separate async flow, accepting bytes *or* a URL | `MediaUploadRequest` already accepts `url` **or** multipart `file`, mutually exclusive | **Copy the existing media pattern.** Little new design needed |
| Read-only on MCP, full CRUD on REST | ai-agents write tools all sit behind Agno's confirmation gate; adding gated brand writes is real work | **Read-only agent tools in v1** is both the market norm and the cheaper build |
| Bound every agent response; search, don't list | The ai-agents envelope hard-caps model-facing results at **8 KB** and truncates structurally | **A summary/detail split is forced by the platform**, not just advisable |

## 2. Where they conflict — decisions the PO must take

### D1. Reverse the locked "status only, never brand content" decision (**blocking**)

`BrandStatusResponseData.php` records decision **D2**: the public API deliberately exposes only `{configured, enabled}` because (a) brand in a request body could be forged by any API key, and (b) publishing the wire contract creates a public contract that future revamps must maintain.

Both concerns are real and both are answerable, but the reversal is a product call:

- **(a) Forgery** is answerable by keeping brand strictly *resolved server-side at generation time*. Nothing in this epic requires accepting inline brand in a generate request. CRUD on the brand resource is a separate, authenticated, workspace-scoped write — the same risk profile as editing it in the dashboard.
- **(b) Contract lock-in** is answerable by versioning the public brand schema independently of the internal document, and by shipping read before write. It is still a genuine cost: the brand-knowledge-revamp is itself mid-flight, so we would be publishing a shape that is actively changing.

**Sequencing consequence:** the revamp's data model (one brand, five tabs) should be *landed* before its public contract is frozen. Publishing the API against the pre-revamp shape and then migrating it is the worst of both worlds.

### D2. Build by hand, or through the parity contract? (**blocking**)

`docs/stories/public-api-parity-contract/` is a **P0** epic whose entire premise is that adding labels and campaigns to the API took ten stories because Zapier, Make, n8n, the GPT action schema, the Claude extension and the MCP tools each needed their own ticket. It exists to replace hand-copying with one definition source that generates every surface.

Brand Knowledge across REST + MCP + CLI + agent toolkit + chat is exactly the shape that epic was written to absorb. Three options:

| Option | Cost | Risk |
|---|---|---|
| **A. Wait for the parity contract**, then define brand once | Slowest start | Brand ships late; parity epic has no delivery date |
| **B. Build brand by hand now**, in the current per-surface way | Fastest to first value | Re-pays the ten-ticket tax; adds one more surface set to migrate later |
| **C. Build REST by hand now; make brand the first capability defined through the contract** when it lands | Middle | Requires coordinating two epics |

Recommendation: **C**, with the REST resource explicitly designed to be the definition source's first input.

### D3. Brand Assets — flag or folder?

The code stores brand assets as `media` rows with a `brand_asset: true` boolean, excluded from the normal media grid. brand-knowledge-revamp **BR-5** specifies a per-workspace "Brand Assets" **folder** in the Media Library as the single source of truth. These are different models and the public asset contract cannot be written until one wins.

### D4. Does this reach mobile? (BR-8 conflict)

brand-knowledge-revamp **BR-8** says AI brand features are web-only. But `contentstudio-flutter/lib/features/ai_assistant/` already renders the same nine stream-protocol-v2 components as the web timeline. **If a brand read tool is attached to a team member, a Flutter user asking "what's my brand voice?" gets an answer whether or not we intended it.** Options: accept it (zero extra work, and it is a read), suppress brand tools for mobile sessions (real work), or write an explicit `[Flutter]` story to render it deliberately.

### D5. Enums or free-form?

`tone`, `emotion`, `character`, `language` and `colors[].role` are free-form strings today, capped at 200 chars, with `brand|background|text|accent` existing only as a code comment. A public contract must either publish an enum (better for agents, breaking to change) or publish free-form and document it (honest, harder for an agent to fill correctly).

### D6. Gating and metering

There is no plan gate or feature flag on Brand Knowledge at all today. The market norm is aggressive gating on REST and free MCP as an acquisition wedge. Decisions needed: which plans get brand API access; whether brand reads cost an API credit (every workspace-scoped request currently charges one); whether ingestion — the expensive operation — costs AI credits.

### D7. Where do MCP and CLI actually get built?

`contentstudio-mcp` (`github.com/d4interactive/contentstudio-mcp`) and `contentstudio-cli` / `contentstudio-agent` (`github.com/contentstudioio/contentstudio-agent`) are **separate repos, not mounted in this workspace.** They cannot be analyzed or implemented from here. Stories targeting them are written blind and need their own grounding pass, or the epic scopes them as a follow-on.

## 3. Open questions for the review gate

| # | Question | Why it matters |
|---|---|---|
| 1 | Do we overturn D2 and publish brand content on the public API? | Blocking. The whole epic rests on it |
| 2 | Does this wait for the brand-knowledge-revamp data model to land? | Publishing a contract against a shape that is mid-revamp is expensive |
| 3 | Parity contract: A, B or C above? | Determines whether this is ~1 epic or ~3 |
| 4 | Brand assets: `brand_asset` flag or Media Library folder? | Blocks the asset contract |
| 5 | Read-only agent tools in v1, or gated writes too? | Market says read-only; user's ask says "the whole thing" |
| 6 | Does brand reach the Flutter AI assistant, given BR-8? | It arrives for free unless suppressed |
| 7 | Enums for tone/emotion/character/language, or free-form? | Breaking to change later |
| 8 | Which plans, and do brand reads consume an API credit? | No gating exists today |
| 9 | Are MCP and CLI in this epic, or a follow-on? | Their repos are not mounted here |
| 10 | Do the silent caps (50 sources, 5 websites scraped, 100 assets) become documented API limits with real error codes? | They currently produce a misleading 404 |

## 4. Recommended scope shape (for Step 2 to refine)

**v1 — the wedge**
1. Public REST **read** for all five families, plus the brand status endpoint kept as-is for compatibility.
2. Async **ingestion job + status resource** — the unblocking piece for everything else.
3. Public REST **write** for profile, style, voice, and source-material add/delete/reingest.
4. **Brand asset upload** reusing the media upload pattern, with a pollable job.
5. **Read-only brand toolkit** in `contentstudio-ai-agents`, likely on its own team member given the D-07 tool ceiling.
6. **AI chat** auto-loads brand summary; source materials retrieved on demand via bounded search; the answer states which brand knowledge it used.
7. OpenAPI annotations + a hand-written `docs/api/brand-endpoint.md`.

**v2 — follow-on**
- MCP server tools and CLI commands (separate repos).
- Gated brand writes through the agent confirmation gate.
- Webhooks on brand change and ingestion completion.
- Per-operation rate limits.
- DTCG export of brand style.
- Draft/publish gate on brand knowledge.

**Explicitly out of scope**
- Multiple brands per workspace — the revamp removes it by design.
- Accepting inline brand in generation requests — brand stays resolved server-side (this is what keeps D2's forgery concern answered).
- A new `brand` render component in the chat timeline — ride the existing `table` / `metrics` / `preview` arms, which also avoids a coordinated four-repo change.
- Adobe-style brand fine-tuning.
