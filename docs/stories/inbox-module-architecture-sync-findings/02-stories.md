# Stories: Inbox Module Architecture and Sync Findings

---

## Story 1: [BE] Stabilize inbox architecture and sync reliability based on findings doc

### Description:

As the ContentStudio engineering team, we want to address the core inbox architecture and sync reliability issues identified in the findings document so that inbox behavior becomes consistent, debuggable, and trustworthy across list views, conversations, counts, send/reply flows, and realtime updates.

This story is the umbrella technical tracking story for the inbox findings captured here:
- Shortcut doc: https://app.shortcut.com/contentstudio-team/write/IkRvYyI6I3V1aWQgIjY5YzI2NGI0LTk5ZWMtNGFjMi1iZWIzLTFjNjU4OTVmYTcxMyI=
- Local reference: `docs/technical/inbox-module-architecture-and-sync-findings-2026-03-24.md`

The findings show that the inbox issues are not isolated bugs. The main problems are architectural:
- no single canonical read model for the inbox list
- fragmented persistence across multiple collections and services
- realtime implemented as notify-then-refetch instead of authoritative state updates
- stale or non-canonical mutation responses
- inconsistent status models and count derivation
- API contract issues that surface as `undefined` payloads and delayed UI reconciliation
- concrete sync worker bugs on specific platform paths

This story tracks the work needed to turn the findings into an implementation plan and begin correcting the core platform reliability issues rather than shipping more surface-only fixes.

---

### Workflow:

1. Engineering reviews the linked inbox findings document
2. Team identifies the highest-leverage architectural fixes first, especially canonical read model and sync/realtime correctness
3. Backend and service-layer changes are planned so inbox rows, counts, messages, comments, and state transitions come from a more reliable and consistent source of truth
4. Frontend reconciliation complexity is reduced as non-canonical and timing-sensitive backend behavior is removed
5. Inbox behavior becomes more predictable across refreshes, filters, sends, replies, and live updates

---

### Acceptance criteria:

- [ ] The findings document is treated as the source context for this story
- [ ] A concrete implementation plan is defined for the highest-priority inbox reliability fixes identified in the document
- [ ] The plan explicitly addresses the lack of a canonical inbox read model
- [ ] The plan explicitly addresses sync/realtime timing issues and notify-then-refetch drift
- [ ] The plan explicitly addresses stale mutation responses and inconsistent response contracts
- [ ] The plan explicitly addresses sidebar count and inbox list divergence
- [ ] The plan explicitly addresses known deterministic sync-worker bugs called out in the findings
- [ ] The first implementation slice from this plan is ready to split into follow-up engineering stories without duplicating or contradicting the findings doc

---

### Mock-ups:

N/A - architecture and platform reliability story.

---

### Impact on existing data:

No immediate schema migration is required for this tracking story, but follow-up implementation may affect inbox read models, sync persistence paths, or response contracts across `social-inbox-manager`, `contentstudio-backend`, and `contentstudio-frontend`.

---

### Impact on other products:

- Web App: Direct impact on inbox reliability, list behavior, counts, sends/replies, and realtime updates
- Mobile apps: Potential indirect impact if they rely on the same inbox backend behavior or contracts
- Chrome extension: No known direct impact
- White-label: Inbox behavior should remain consistent across white-label domains if backend response and realtime paths are unified

---

### Dependencies:

Depends on the linked inbox architecture and sync findings document.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness - N/A, architecture/platform story
- [ ] Multilingual support - N/A, no new UI copy in scope
- [ ] UI theming support - N/A, no UI changes in scope
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
