# Automatic competitor migration: source connectivity evidence

Checked 2026-11-01 using public first-party documentation only. No customer API call, credentials, live source connection or actual export tested. **Documented** means the source publishes the capability; it does not prove that ContentStudio can obtain production access, use it for migration, or retain its output. **Unknown** is not a claim that an API is absent.

## Product conclusion

An automation-first migration product is credible as **a set of source-specific connectors**, followed by automatic inventory, extraction, mapping, held import and reconciliation. There is no evidence for a universal “connect any competitor and copy everything” API. AI cannot retrieve records that the authorized source interface does not expose or remove source/platform restrictions.

Preserve the intended experience: a customer authorizes the connector, selects client boundaries, the agent handles the repetitive transfer, and only unresolved mappings/permissions/cutover issues require attention. A file adapter can be an assisted fallback for restricted sources; it should not redefine the target product as a manual CSV uploader.

## Source comparison

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

## Hootsuite automation path: investigate both REST and operational MCP

The REST OAuth flow inherits the member’s dashboard permissions. Outbound-message reads use cursor pagination (max 100 rows/request), windows no longer than four weeks, and scheduled/sent/review/rejected/failed states. Profile responses distinguish the source ID from network-native `socialNetworkId`. App access is applied for; it was not obtained or tested. Ordinary REST eligibility was **not** verified as Enterprise-only.

Hootsuite now publishes operational MCP endpoints: `https://mcp.hootsuite.com/perch`, `/nest`, `/lumen`, `/parliament`. These differ from `developer.hootsuite.com/mcp`, which is a documentation interface. Perch advertises profile/workspace discovery, library access and analytics; Nest covers inbox reading/triage. Lumen needs a Talkwalker standalone licence in addition to MCP access. Product help states entitlement and role limits continue to apply, as do public-API social-data restrictions. The public pages do not establish complete bulk extraction, long-running unattended authorization or a complete tool schema. Treat MCP as an adapter candidate, not a full migration solution already available.

REST/API and MCP terms must be assessed separately. [API terms](https://www.hootsuite.com/legal/api-terms-of-service) include consent, deletion following deauthorization and agreement termination, and relevant business-use restrictions. [MCP terms](https://www.hootsuite.com/legal/mcp-server-terms) explicitly cover scripts and AI agents and restrict bulk extraction **for competitive tracking, external model training or fine-tuning**, along with misuse/rate-limit circumvention. Do not describe that clause as a blanket ban on all bulk migration; migration suitability remains unresolved. Customer authorization alone does not settle retention rights.

## Discovery should report capability, not promise “all data”

An agent should first produce an inventory by source object category with five possible findings: available to transfer, requires additional permission, representable only as reference/history, unavailable from this connector, or blocked by policy. Extraction must paginate deterministically, retain watermarks and source IDs, verify completeness, and keep credentials out of the LLM context. The AI can help map schemas; the source connector should control the actual data access.

Browser/computer-use automation is not a universal fallback for missing APIs. Only consider it after confirming vendor permission, export semantics, unattended access policy, credential handling and maintenance/security requirements. Never propose shared passwords, session-cookie collection, private endpoints or bypassing access controls.

## Recommended investigation sequence

1. **Hootsuite first, as the PO selected:** establish actual REST production access and migration rights; discover permitted MCP tool coverage in parallel. Verify org/team/member/social account inventory and representative post/media/history reads. Use the export as a verification fixture or gap fallback, not the product’s defining interaction.
2. **Prove the automatic engine through an accessible connector:** if Hootsuite onboarding blocks technical learning, Buffer OAuth or Publer scoped API access offers a documented route for unattended reads after customer setup. Retain Hootsuite as the priority target; do not silently change the product scope.
3. **Verify representative categories, not only a calendar parser:** identity, multi-client boundaries, members, drafts/schedules, media and historical references need explicit pass/fail coverage. Inbox, analytics and automation settings remain visible in the roadmap with source-specific capability limits.
4. **Keep source connection separate from social authorization:** the source connector reads competitor data. ContentStudio’s own platform connections are separately required to operate the resulting social accounts.

## Documentation conflicts to resolve during implementation

- Buffer developer/help pages use different scope spellings; use current live provider contract. The newer help explicitly excludes analytics permission from third-party app clients.
- Sprout’s marketing FAQ says no public API, but its dedicated developer reference and support article document one, including OAuth2. Prefer the product-specific reference.
- Later help and the reporting reference disagree on token duration; use returned JWT expiry and the authoritative live auth contract. Avoid using Influence reporting as evidence of Social scheduling API coverage.
