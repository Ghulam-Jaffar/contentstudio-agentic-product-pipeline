# Research: Automated competitor migration agent for ContentStudio

**Date:** 2026-11-01  
**Status:** Initial feasibility research for PO discussion. No workflow, PRD, implementation, or tracker push approved.  
**Priority source:** Hootsuite, selected by the PO.  
**Scope:** Technical migration, security, authorization, and operational continuity for agencies and power users. Customer acquisition and sales are outside scope.

## 1. Finding

The product goal is an **automated migration agent**: the customer authorizes a source connection and client scope; the service discovers accessible data, proposes mappings, transfers records and media, verifies the destination, and presents only unresolved choices and the final handover for review. Customers should not re-enter every record or confirm every deterministic copy.

The preferred extraction route is an approved source API or operational MCP interface. Authorized export files are a fixture or assisted fallback for sources without a verified automatic route. They do not define the final product as a manual CSV importer.

The goal covers workspaces, members and permissions, social accounts, content, media, available history, approval context, listening and automations. **Complete portability is not established.** Each connector must report what is available, requires permission, is reference-only, is unavailable, or is policy-blocked. “All data” means all selected, accessible, permitted data that can be represented faithfully, with every omission disclosed.

Social accounts anchor many relationships, but organization/client/workspace boundaries govern authorization. Listening, media, members and planning objects are not all dependent on a social account.

The proposed implementation is a durable migration workflow plus a constrained AI mapping planner. AI helps interpret schemas, suggest relationships/rule translations and explain exceptions. Deterministic workers enforce identity, authorization, exact content copying, idempotency and completion checks. A custom-trained model is not needed to begin.

Potential value is reduced rebuilding and preserved operational context. The amount of work saved, source access, coverage, throughput and business value remain unmeasured. This research is not proof of a working agent.

Detailed evidence: [source connectivity](automation-source-research.md), [code-grounded agent architecture](agent-architecture-research.md), [readability research](readability-research.md).

## 2. Evidence and limits

This research combines local source inspection and public first-party documentation. No customer account, production system, competitor credential, or real customer export was accessed. Documentation confirms available routes; it does not prove which fields a particular agency can export or that the mappings work without loss.

**Observed:** Existing ContentStudio implementation and statements in cited official documentation.  
**Proposed:** The staging architecture, pilot scope, safeguards, and migration policy below.  
**Unverified:** Actual export schemas, customer entitlements, legal clearance for the proposed service, throughput, and end-to-end import fidelity.

Local code is a working checkout, not proof of production behavior. Runtime verification remains necessary before implementation commitments.

## 3. Hootsuite extraction routes

### Approved REST connector

Hootsuite documents OAuth2 and reads for organization/team/member permissions, social profiles and outbound messages/review history. Profile records distinguish source IDs from native network identifiers. Message reads have bounded windows and pagination. Complete ordinary-draft inventory and all configuration export remain unverified. Actual production app access was not obtained; do not assume ordinary REST is universally available or Enterprise-only. [REST reference](https://apidocs.hootsuite.com/docs/api/index.html).

### Operational MCP connector

Hootsuite publishes Perch, Nest, Lumen and Parliament MCP servers. Its help describes product-specific profile/workspace, library, analytics, inbox/triage and listening capabilities. A conversational tool’s availability does not prove exhaustive bulk enumeration, stable pagination or unattended migration access. Obtain the actual permitted tool catalog and test it; preserve source entitlements and network restrictions. The documentation MCP endpoint is distinct from these operational product connectors. [MCP capabilities and limitations](https://help.hootsuite.com/s/article/about-mcps), [published server interfaces](https://www.hootsuite.com/integrations/mcp).

### Exports as fixtures or explicit fallback

Calendar CSV is useful for fidelity checks and supported gaps, but omits ordinary unscheduled drafts and some platform fields; asset URLs expire. Inbox and analytics have separate export eligibility, network limits and formats. Files do not establish a seamless API connection or unrestricted retention rights. [Calendar](https://help.hootsuite.com/s/article/export-calendar), [Inbox](https://help.hootsuite.com/s/article/export-inbox-data), [analytics](https://help.hootsuite.com/s/article/export-analytics-report).

### Rights and operational limits

Assess the actual route’s source/customer/platform terms. REST API terms include consent, retention/deletion and relevant business-use conditions. MCP terms explicitly cover AI agents and restrict certain bulk-extraction purposes, including competitive tracking and external model training/fine-tuning; that is not a blanket statement that every migration is prohibited. Migration suitability and continued retention need legal clarification. Paid access does not remove those questions. [API terms](https://www.hootsuite.com/legal/api-terms-of-service), [MCP terms](https://www.hootsuite.com/legal/mcp-server-terms).

Do not collect source passwords or session cookies, invoke private endpoints, bypass rate/access controls or use browser automation as an assumed escape hatch. If no permitted automatic extraction route exists, disclose assisted migration or unavailable coverage.

## 4. Data mapping and feasibility

The following describes destination feasibility. Source availability remains conditional unless established above.

| Customer data | Recommended treatment | Main unresolved issue |
|---|---|---|
| Organizations, clients, teams, workspaces | Preview a mapping into destination workspaces; customer confirms boundaries | A Hootsuite team is not automatically a ContentStudio workspace; profiles can be shared across teams |
| Social accounts | Discover candidate matches, then connect through ContentStudio’s own platform authorization | Match verified platform identity, account type and workspace; competitor IDs, handles and names alone are insufficient |
| Members and roles | Propose existing-user matches and pending invitations; review effective access before sending | Owner/admin/approver/custom roles do not map exactly; agency admin status does not establish every client’s authority |
| Drafts and future posts | Stage content and media; preserve original schedule as metadata; activate only validated records | Missing platform options, assets, timezones, permissions, plan limits and source export omissions |
| Already published posts | Reference existing native posts or permitted historical records; never republish | Native post IDs may be absent; published content might also arrive through fresh network retrieval |
| Media library | Import permitted originals, folders and metadata where available | Calendar attachments are not the complete library; download links expire and asset rights can differ |
| Labels, campaigns and content categories | Suggest mappings; preserve source names/provenance | Similar names can hide different semantics; do not collapse distinct objects automatically |
| Approvals and internal comments | Keep source audit history attributed; require destination review for active approval rules | Past approval is not a ContentStudio approval; authors might not become destination members |
| Inbox conversations and contacts | Separate permitted historical reference from live platform conversations | Identity scopes, native thread IDs, API lookback, retention and reply windows limit continuity |
| Analytics | Begin native collection/retrieval after authorization; retain permitted source reports separately | Metric definitions, windows, attribution and historical snapshots differ; do not splice them into one native trend |
| Listening queries and results | Translate supported query intent for review; restart collection | Query syntax, licensed datasets, retention and platform coverage differ |
| Automations, queues and RSS rules | Recreate supported rules disabled, then review behavior | Trigger/action parity, timezones and duplicate execution across tools |
| Ideas, tasks and notes | Import supported editable records when extractable; retain attribution | No verified general Hootsuite export route for these objects yet |

**Account access is re-established, not transferred.** A source profile record can identify a candidate account, but it is not a ContentStudio connection. Do not import social-network secrets from another vendor. Matching should use a verified tuple of social platform, account type, native identifier and destination workspace, with explicit exceptions for app-scoped identifiers. An ambiguous match remains unresolved.

## 5. ContentStudio implementation findings

Paths below are relative to this repository. Line references describe the inspected checkout. Backend-specific instructions identify Laravel 11, despite the older Laravel 10 summary at repository root.

| Area | Observed code | Consequence for migration |
|---|---|---|
| Workspace setup | `contentstudio-backend/app/Http/Controllers/Api/V1/WorkspaceController.php:194` and `:296` validate entitlements/timezone and delegate creation | Reuse validation; explicitly map source clients into workspaces and check capacity |
| Team setup | `contentstudio-backend/app/Http/Controllers/Api/V1/TeamMemberController.php:258` delegates addition; `contentstudio-backend/app/Http/Controllers/Settings/Team/TeamController.php:306` sends the invitation | Stage proposed memberships before using normal create; validate effective account/category access, especially elevated roles |
| Account identity | `contentstudio-backend/app/Data/Integrations/SocialAccountData.php:25`; `contentstudio-backend/app/Models/Integrations/Platforms/SocialIntegrations.php:20` | Existing account contracts expose workspace, platform and native identifiers; they are mapping inputs, not proof of portable credentials |
| Connection lifecycle | `contentstudio-backend/app/Http/Controllers/Integrations/Platforms/Social/FacebookController.php:109` exchanges a code for the CS app; `contentstudio-backend/app/Jobs/Integrations/Platform/PostConnectPlatformProcessingJob.php:45` initializes webhooks/inbox/analytics | Connect normally; account for background side effects. An empty selected-step list means all steps (`:97`), not no steps |
| CSV staging | `contentstudio-backend/app/Repository/Publish/Automation/CsvSchedulingRepository.php:11` defaults omitted status to scheduled; `contentstudio-backend/app/Http/Controllers/Automation/CSVProcessController.php:300` atomically claims staging rows on approval | Reuse staging/validation precedents. Row claiming is not cross-batch source deduplication; explicitly hold imports |
| Post/approval writes | `contentstudio-backend/app/Http/Controllers/Api/V1/PostController.php:957` calls live post processing; `:981` attaches approval after creation, with failure returned as warning | Normal post creation is not a historical import API. Approval failures must block migration activation; do not replay historical approvals |
| Media ingestion | `contentstudio-backend/app/Services/Api/V1/PostMediaService.php:71` reuses library media; `:207` validates kind/quota and ingests; `:115` rolls back request-created items | Useful reusable service, but migration needs durable asset mapping, safe downloads and resumable cleanup beyond one request |
| Publishing pause | `contentstudio-backend/app/Http/Controllers/Api/V1/WorkspaceSettingsController.php:141` also locks RSS/Evergreen; resume handling at `:265` | Whole-workspace pause affects unrelated work. It cannot substitute for migration-specific holds in existing workspaces |
| Tasks | `contentstudio-backend/app/Http/Controllers/Planner/TasksController.php:48` requires a parent plan; `:76` emits collaboration updates | Map parent posts and assignees first; suppress historical notifications. General ideas/boards parity remains unverified |
| Inbox backfill | `social-inbox-manager/app/api/v1/routes/sync.py:148` defines per-platform capabilities; `:129` bounds fetch counts; `:721` excludes WhatsApp historical retrieval | Available network history varies. Fetch count limits do not establish historical retention; source-only notes/assignments cannot be recovered by reconnecting |
| Inbox action eligibility | `social-inbox-manager/app/services/auto_reply/message_age.py:9`; Facebook strategy calls it at `social-inbox-manager/app/social_sync/facebook_strategy.py:894` | Recent sync items can qualify for auto-reply. A backfill job identifier alone is insufficient; historical imports need explicit action suppression |
| Analytics backfill | `contentstudio-social-analytics-go/src/services/facebook/facebook-fetcher/main.go:489` uses bounded incremental/full-sync ranges and page caps | Facebook full sync currently selects 90 days. This code setting is not a network guarantee, nor evidence of equivalent history on other platforms |

Frontend entry points worth reusing during later design:

- Bulk upload/progress: `contentstudio-frontend/src/modules/automation/composables/useBulkUploadAutomationSave.ts:394` and `src/modules/automation/components/csv/composables/useCsvProcessingStatus.ts`.
- Account selection/connection: `contentstudio-frontend/src/modules/integration/components/dialogs/SaveSocialAccounts.vue` and `src/modules/integration/components/platforms/social_v2/composables/useSocialAccounts.ts`.
- Workspace/team setup: `contentstudio-frontend/src/modules/account/components/CreateWorkspaceForm.vue`, `src/api/team.ts:227`.
- Approval builder: `contentstudio-frontend/src/modules/approval-workflows/components/ApprovalWorkflowBuilder.vue`.
- Listening query setup: `contentstudio-frontend/src/modules/listening/components/topics/ListeningTopicFormModal.vue`; listening is not entirely channel-bound.

These are reusable entry points, not an instruction to bypass domain validation or tenant authorization. Any new migration API must also be planned across the public REST API, CLI, MCP server and applicable Zapier/Make/n8n surfaces per repository policy. Runtime scheduler eligibility, observer/webhook suppression, mapping-index uniqueness, crash consistency and billing effects remain engineering validation items.

No competitor-specific migration implementation was found in the backend/frontend searches performed for this investigation. This is a bounded search finding, not proof that no related work exists anywhere.

## 6. Proposed technical architecture

Use source-specific adapters around a shared migration engine. An adapter reads authorized exports or approved APIs and produces a normalized manifest; it does not directly publish, invite people, or send messages.

```mermaid
flowchart LR
    Source[Approved REST or MCP connector] --> Adapter[Source adapter]
    Adapter --> Stage[Isolated staging and provenance]
    Stage --> Validate[Validate assets and relationships]
    Validate --> Preview[Customer reviews mappings and gaps]
    Preview --> Commit[Commit held destination records]
    Commit --> Reconcile[Reconcile and verify source pause]
    Reconcile --> Activate[Activate selected future work]
```

This is an architecture hypothesis for research, not the approved user workflow.

1. **Manifest and provenance:** Record source vendor, organization, export snapshot/timezone, source object IDs, relationships, source timestamps, and schema version. Classify each item as importable, needs review, reference-only, excluded, or unavailable. Do not invent source IDs when absent; use clearly marked deterministic fingerprints with collision handling.
2. **Tenant-isolated staging:** Keep imported content outside active scheduling and reply pipelines. Recheck destination workspace permissions during preview and commit. Bind every job, mapping and media object to the authorized customer and workspace.
3. **Relationship resolution:** Resolve workspaces, accounts, users, labels/categories, media, posts and dependents in order. Keep source-ID-to-destination-ID mappings scoped to source organization and destination tenant. Do not match by display name alone or silently merge into an existing workspace.
4. **Validation:** Check timezone/DST, future/past schedule status, text and platform options, asset completeness, connected-account capabilities, approval rules, access rights, and plan/storage limits. Preserve unsupported fields in the issue report rather than silently dropping them.
5. **Safe, resumable writes:** Use durable per-object idempotency keys and checkpoints. A retry updates or reuses the intended held object; it does not duplicate a post, asset or invitation. Compare changed snapshots explicitly so reruns cannot overwrite the user’s destination edits.
6. **Separate commit and activation:** A committed draft is still held. Invitations, scheduled publishing, automation and downstream integrations require their own reviewed activation policy. A generic “save” call must not accidentally activate anything.
7. **Cutover:** Identify source work due around cutover, pause it in Hootsuite through the customer’s authorized process, reconcile any source posts published since the snapshot, then release only selected future posts in ContentStudio. Revalidate schedules that became past due. No immediate catch-up publishing. Cross-vendor exactly-once execution is not established; unresolved source state blocks affected activation.
8. **Reconciliation and recovery:** Report counts by workspace/account/type, missing relationships/assets, transformed fields, and object-level failures. Removal/reversion can cover only untouched migration-created destination records where safe. It cannot undo a post, email or reply already sent, or restore third-party state. Never promise universal rollback.

Existing account connection may start background sync before migration commit. The team must either isolate it or prove that imported history and fresh sync reconcile without duplicate records or unintended responses.

## 7. Security, privacy and legal design

There are three separate permissions to establish: authority to migrate this client’s data, permission to extract/use it through the selected source route, and permission to retain/process the underlying social-platform data. Authorization to see something in an agency dashboard does not by itself prove all three.

GDPR portability is conditional and concerns an individual’s qualifying personal data; it is not a general entitlement to extract a company’s entire workspace or third-party conversations. Rights of others remain relevant. Applicable jurisdictions, agency/client contracts and ContentStudio’s processing role need assessment. [Irish Data Protection Commission: portability](https://www.dataprotection.ie/en/individuals/know-your-rights/right-data-portability-article-20-gdpr).

Where ContentStudio acts as a processor, migration should follow documented controller instructions and the applicable processing agreement, including subprocessors and end-of-contract handling. Identify controller/processor roles per data flow rather than assuming one relationship for every agency and client. [ICO: controller–processor contracts](https://ico.org.uk/for-organisations/uk-gdpr-guidance-and-resources/accountability-and-governance/contracts-and-liabilities-between-controllers-and-processors-multi/).

Proposed controls:

- Collect only selected client workspaces and data categories. Keep temporary uploads and credentials for a defined, justified period, with actual deletion and revocation behavior. Durable destination content needs a separate retention policy. [European Commission: GDPR principles](https://commission.europa.eu/law/law-topic/data-protection/information-business-and-organisations/principles-gdpr_en).
- Encrypt content in transit and at rest; isolate tenant storage and jobs; restrict and audit staff access. Keep credentials, signed media URLs, message bodies and personal data out of routine logs. Do not route customer exports through general AI services for parsing by default.
- Where the source supports OAuth, use its supported authorization flow with minimal scopes and current OAuth protections; restrict the adapter to read operations. Do not collect source passwords or session cookies. [IETF OAuth security best practice](https://www.rfc-editor.org/rfc/rfc9700.html).
- Treat uploaded files and external asset URLs as untrusted input. Apply size/type/signature checks, private quarantine, scanning, safe filenames, safe rendering and bounded processing. Asset retrieval must block private/internal destinations and unsafe redirects/protocols, including revalidation after DNS/redirect resolution. [OWASP file uploads](https://cheatsheetseries.owasp.org/cheatsheets/File_Upload_Cheat_Sheet.html), [OWASP SSRF prevention](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html).
- Attach origin, licensing/retention policy and deletion handling to historical records. Do not equate upload by a customer with independent authorship or unrestricted rights. LinkedIn, for example, distinguishes member and organization data and applies different storage rules; classification matters. [LinkedIn data storage requirements](https://learn.microsoft.com/en-us/linkedin/marketing/data-storage-requirements).
- Confirm data residency, cross-border transfers and subprocessors for the selected customers. Assess whether a DPIA is required based on actual scale, data sensitivity and processing risks.

**Legal conclusion:** This research identifies questions and a conservative engineering route. It does not establish legal clearance. Before a real customer pilot, review the applicable source/customer/platform terms, data rights, processing arrangements and retention with counsel. Request migration-specific written clarification where API or retention rights remain ambiguous. Customer-generated exports are not a workaround for restrictions.

## 8. Automatic-agent proof and release evidence

Begin with Hootsuite, the PO’s chosen source, and prove the connected extraction path. Discover authorized categories and relationships across a controlled client scope. Use synthetic fixtures or appropriately de-identified samples before real customer data. Export files can help check fidelity, but a file parser alone does not prove the automatic source connection.

The first proof must return:

1. Actual source authentication, permissions, tool/endpoint coverage, pagination and missing categories.
2. Verified mappings for client/workspace, social identity, members, media and representative unpublished/history records.
3. Exact field and asset comparisons, supported transformations and a reason for every omission or hold.
4. Resumable extraction and writes, including crash-after-write-before-checkpoint and unknown response recovery, without duplicate objects or overwriting destination edits.
5. Zero unintended publishing, invitations, replies, automation or external activity during discovery, preview and held import.
6. A controlled handover with verified source state and final reconciliation; no automatic overdue catch-up publishing.
7. Measured user interventions, completion time, API/model cost and data-fidelity outcomes. Set thresholds from validated fixtures and risk, not invented model-confidence percentages.
8. A valid extraction/use/retention basis and an auditable deletion/cancellation process.

Keep wider entity coverage visible rather than redefining the product as only a future-calendar importer. Expand each data category only when both source reads and destination representation are proven. If Hootsuite app onboarding blocks technical learning, an accessible supported connector may validate the shared engine while Hootsuite remains the priority source.

Before adding infrastructure, evaluate the existing optional Temporal operations against Dramatiq plus a durable SQL ledger. Do not run a long migration inside a short-lived chat session.

## 9. Decisions to discuss with the PO and team

- Which data is essential to preserve the agency’s actual working setup, including history requirements?
- Can approved Hootsuite REST/MCP access cover that inventory and the intended migration/retention use?
- Which client boundaries and consequential permission changes need explicit confirmation?
- What bounded batch authorization can allow unattended deterministic copies without using chat-tool confirmations as a bypass?
- Which missing categories are acceptable as reference/history, assisted extraction or explicit exclusions?
- Who owns final source pause and activation when the source offers no supported automated pause operation?
- Which fidelity, intervention, cost and safety outcomes justify offering this as a premium service?

This remains research. Workflow, PRD, implementation stories and tracker writes require their respective review steps.

## 10. Competitor source connectivity

| Source | How an automated agent could connect | Documented extraction coverage | Eligibility / main limitation | Confidence |
|---|---|---|---|---|
| **Hootsuite** | REST OAuth2 with registered app; official operational MCP servers also use customer authorization | REST: org/team/member/permissions, social profile identity, outbound messages and review history. MCP: product-dependent workspaces/profiles, content library, analytics, inbox and listening operations | Production app access and migration use/retention need confirmation. A general complete draft inventory and all configuration export are unproven. MCP entitlement and tool/pagination coverage differ | High for published interfaces; unverified end-to-end |
| **Buffer** | Registered OAuth app client with PKCE; owner-created scoped personal key where appropriate | Organizations/channels, posts/queue, ideas and snippets; API also supports personal post analytics | All plans have API access subject to limits. **Third-party app clients do not receive `insightsRead`; personal keys can.** Full member/role/inbox/workflow export unproven | High |
| **Publer** | Customer supplies a scoped API key; requests identify workspace | Workspaces, accounts, paginated posts including published/scheduled/draft/recycling/recurring states; media library API | Enterprise and Business accounts in good standing. This is automated after key setup; no general OAuth app route established here. Full collaboration/inbox history unproven | High |
| **Sprout Social** | OAuth2 user-based or machine-to-machine client; API tokens also supported | Profiles, users/groups/teams/tags/topics/case queues; published posts, inbox/replies/care data, analytics and restricted listening | Advanced plan + API permission. Scheduled/draft publishing API retrieves a **known publishing ID**; a complete inventory endpoint was not established. Some network data excluded | High |
| **Agorapulse** | Per-user API key (`X-API-KEY`); respects member access | Open API: analytics/listening, inbox reads/triage, calendar notes and content-library operations. Separate publishing CSV includes drafts/scheduled/published state/content/assets | Open API is Custom-plan; complete publishing inventory via API unverified. Publishing CSV is current PPU plans, excludes inbox/analytics and Legacy plans | High |
| **Metricool** | Access token (`X-Mc-Auth`) plus account/brand IDs | Enumerate brands; list scheduled posts by date; read a scheduled post and approval configuration; reporting APIs | Advanced/Custom. Full member/permission history and original library completeness need endpoint validation | High |
| **Sendible** | Supported post/draft/queue file exports are established; current supported third-party API auth/eligibility not established in this check | Calendar/post/draft exports and queue CSV/HTML | Agent can parse supplied exports automatically. Do not present historical/third-party Sendible API references as proof of current supported access | Medium for files; API unknown |
| **Later** | Social product: analytics files. **Later Influence** has separate server-side client-credential Reporting API | Influence: instances/campaigns/creators/post performance and analytics. Social scheduling: reporting CSV | Influence access is issued via account team. This does **not** prove access to Later Social’s future calendar, drafts or media library. Do not label Later as “no API” globally | High for reporting; Social migration unknown |
| **Loomly** | Supported export files; help explicitly says no public API; Zapier has event triggers | Content PDF; analytics PDF/CSV. Triggers can support changes as they happen | A trigger stream is not proof of historical inventory/bulk export. PDF is a weaker structured source. Partnership may be necessary for true automatic extraction | High for stated limitation |
| **SocialBee** | No public API per help; support-assisted CSV export | Named content categories from specified workspaces can be exported by support | Initial extraction requires assistance. Zapier/Make availability does not establish complete historical read/export capability | High |

### Primary source links for the matrix

- Hootsuite: [REST reference](https://apidocs.hootsuite.com/docs/api/index.html), [MCP connectors](https://www.hootsuite.com/integrations/mcp), [MCP help](https://help.hootsuite.com/s/article/about-mcps), [developer access application](https://apps.hootsuite.com/).
- Buffer: [app-client OAuth guide](https://developers.buffer.com/guides/building-apps.html), [personal keys, owners, permissions and app-client analytics restriction](https://support.buffer.com/en-us/articles/how-to-create-your-buffer-api-key-ShIgYVwM6j), [API capabilities](https://support.buffer.com/articles/using-buffers-api-GtIYIQilz5).
- Publer: [auth](https://publer.com/docs/getting-started/authentication), [plan/access](https://publer.com/help/en/article/how-to-access-the-publer-api-1w08edo/), [post states/read API](https://publer.com/docs/api-reference/posts).
- Sprout: [reference, OAuth, metadata and publishing limitations](https://api.sproutsocial.com/docs), [plan and capabilities](https://support.sproutsocial.com/hc/en-us/articles/360045006152-Sprout-Public-API).
- Agorapulse: [Open API auth/eligibility](https://support.agorapulse.com/en/articles/12417183-how-to-connect-to-agorapulse-s-open-api), [publishing CSV](https://support.agorapulse.com/en/articles/15962987-how-to-export-posts-as-a-csv).
- Metricool: [auth, brand enumeration and GET scheduler endpoints](https://help.metricool.com/basic-guide-for-api-integration-r97af).
- Sendible: [post exports](https://support.sendible.com/hc/en-us/articles/6341725672989-Export-your-posts), [queue exports](https://support.sendible.com/hc/en-us/articles/7284538762525-Overview-of-Queues). Direct page content was partly unavailable this check; prior first-party export evidence supports the route, not the exact current schema.
- Later: [Social reporting](https://help.later.com/hc/en-us/articles/33109662792471-Later-s-Custom-Analytics), [Influence API overview](https://docs.reporting.api.later.com/), [credentials/access](https://docs.reporting.api.later.com/getting-started).
- Loomly: [no-public-API / Zapier](https://loomly.zendesk.com/hc/en-us/articles/38970621662619-How-to-Use-Zapier-to-Connect-Loomly-with-Other-Apps), [content exports](https://loomly.zendesk.com/hc/en-us/articles/39019394773531-Can-I-export-content).
- SocialBee: [no public API](https://help.socialbee.com/hc/en-us/articles/29979123668375-Do-you-have-a-public-API-or-white-labeling-options), [support-assisted export](https://help.socialbee.com/hc/en-us/articles/29979141511831).


## 11. Agent architecture and readability evidence

The [agent architecture research](agent-architecture-research.md) records actual reusable Agno tooling, server-injected workspace scope, background workers, optional workflow infrastructure and the new durable import/ledger contracts required. It also covers prompt injection, role mapping, crash recovery, metrics and an explicitly synthetic execution example.

The [readability research](readability-research.md) records peer-reviewed multimedia-learning and human–AI interaction papers, industry guidance and accessibility criteria. The artifact organizes the same substantive research around six reader questions with diagrams, consistent comparisons, nearby constraints and optional implementation detail. This is a design application of the literature, not a claim that user comprehension has been validated.
