# Epic + stories — Developer surface parity contract

**Priority: P0.** Every other epic in this program multiplies through it. **Reconcile with the existing AI surfaces architecture epic before creating this** — they are the same problem, and this should be folded into it rather than run beside it. Nothing is pushed to any tracker.

---

## Epic: Developer surface parity contract

ContentStudio exposes its capabilities to developers through a growing set of surfaces: the REST API, the CLI, the agent skill, the MCP server, Zapier, Make, n8n, and the ChatGPT and Claude connectors that consume the MCP server. A customer picks whichever fits how they work, and reasonably expects them to be able to do the same things.

Today each surface is updated by hand, one ticket at a time. The record shows what that costs: adding labels and campaigns to the API, a small change, took ten stories, because the Zapier app, the Make module, the n8n node, the GPT action schema, the Claude extension and the MCP tools each needed their own ticket to make the same change. Multiply that by a programme of eight capability epics and it is roughly eighty tickets of mechanical copying, each one an opportunity for a surface to be forgotten.

The result is drift, and drift is worse than absence. A customer who finds an action in the CLI and not in Zapier does not conclude that Zapier is behind; they conclude the product is unreliable. And because nothing checks, the only way we currently learn that a surface was missed is a support ticket.

This epic replaces hand-copying with a contract. One place a capability is defined, every surface generated or driven from it, and a check that fails when a surface falls behind. After it lands, shipping a new capability everywhere is one story instead of six, and forgetting a surface becomes impossible rather than merely unlikely.

### Out of scope

- Building any new capability. This epic is the plumbing the capability epics run through.
- Changing what any surface currently does. Existing behavior is preserved, and proving that is part of the work.
- The MCP authorization model and the connector directory submissions. Those are the separate MCP epic, which this depends on for how connectors authenticate.

### Stories

1. `[Research] Map every developer surface and define the parity contract`
2. `[BE] Build the single capability definition source and generate the surfaces from it`
3. `[BE] Add the parity guard that fails when a surface falls behind`

---

## [Research] Map every developer surface and define the parity contract

### Description

As the engineering team, we want every developer surface mapped, with each capability it exposes and the shape it exposes it in, plus an agreed contract for how a new capability reaches all of them at once, so that we stop paying six tickets for one change and stop discovering drift from support tickets.

### Workflow

1. The team lists every surface a developer can drive ContentStudio through, and for each one records where its code lives and who owns it.
2. For each surface, the team lists the capabilities it exposes today.
3. The team produces the parity matrix: capability against surface, showing where each is present, absent, or present but shaped differently.
4. The team records, for each surface, what shape a capability has to take to be added to it, so the contract accounts for real differences rather than assuming they are all the same.
5. The team defines the single definition source: what a capability definition contains, and which surfaces are generated from it versus driven by it at runtime.
6. The team defines the propagation checklist that every new capability must satisfy before it is considered shipped.
7. The team defines how the checklist is enforced, so a missed surface fails a check rather than reaching a customer.
8. The team defines the migration path for the surfaces that exist today, and which are converted in what order.
9. The output is reviewed and approved.

### Acceptance criteria

- [ ] Every developer surface is listed with its code location and owner: the REST API, the CLI, the agent skill, the MCP server, Zapier, Make, n8n, and the ChatGPT and Claude connectors.
- [ ] For each surface, the capabilities it exposes today are listed.
- [ ] A parity matrix shows capability against surface, marking each cell present, absent, or present-but-different, with the difference described where it exists.
- [ ] The matrix identifies every current drift: a capability on one surface and not another, or behaving differently between two.
- [ ] For each surface, the document records what form a capability must take there, so genuine differences between a REST endpoint, an MCP tool, a Zapier action and a CLI command are accounted for rather than flattened.
- [ ] The document defines what a single capability definition contains: the action, its inputs and their validation, its outputs, its authorization requirement, its plan gating, and its human-readable description.
- [ ] The document states, per surface, whether it is generated from the definition or driven by it at runtime, with the reasoning.
- [ ] The document defines the propagation checklist a new capability must satisfy to count as shipped, including the public documentation and the OpenAPI document.
- [ ] The document defines how the checklist is enforced automatically.
- [ ] A migration path is documented for the existing surfaces, with an order and the risks of each step.
- [ ] The document states how the ChatGPT and Claude connectors inherit capabilities through the MCP server, and confirms that assumption against the MCP epic rather than assuming it.
- [ ] The document is reconciled against the existing AI surfaces architecture epic, and states plainly whether this work is a part of that epic or distinct from it.
- [ ] Open questions and recommendations are captured for the review gate.
- [ ] The document is reviewed and approved before implementation begins.

### Mock-ups

None. Research ticket with no UI.

### Impact on existing data

None.

### Impact on other products

This research covers every surface a developer can use, so its conclusions govern how future capabilities reach all of them. No product changes yet.

### Dependencies

- Must be reconciled with the existing AI surfaces architecture epic, which is chartered on the same problem for AI tool definitions.
- Takes from the MCP epic how the ChatGPT and Claude connectors authenticate and inherit capabilities.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, research ticket with no UI
- [ ] Multilingual support (surface descriptions and error messages are customer-facing, so the contract must state how they are translated)
- [ ] UI theming support — N/A, research ticket with no UI
- [ ] White-label domains impact review (the contract must state how a white-label customer's surfaces are addressed)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Build the single capability definition source and generate the surfaces from it

### Description

As a developer integrating with ContentStudio, I want every surface to offer the same capabilities with the same behavior, so that I can pick the surface that suits me without finding out later that it is missing an action the others have.

### Workflow

1. A capability is defined once, in one place, with its inputs, outputs, authorization and plan gating.
2. Every surface that can be generated from that definition is generated: the OpenAPI document, the MCP tool list, the CLI commands, the agent skill, and the connector action schemas.
3. Surfaces that cannot be generated are driven by the definition at runtime, so they cannot describe a capability the definition does not have.
4. The existing capabilities are migrated onto the definition source, in the agreed order, with no change in behavior on any surface.
5. Adding a capability afterwards means editing the definition, and every surface picks it up.

### Acceptance criteria

- [ ] A capability is defined in exactly one place, containing its action, inputs and validation, outputs, authorization requirement, plan gating, and description.
- [ ] The OpenAPI document is generated from the definition source and is never hand-edited.
- [ ] The MCP tool list is generated or driven from the definition source, so the ChatGPT and Claude connectors inherit new capabilities without a separate change.
- [ ] The CLI commands and the agent skill are generated or driven from the definition source.
- [ ] The Zapier, Make and n8n surfaces are generated or driven from the definition source, or where a marketplace requires a manual submission, the submission payload is generated and the manual step is the only manual step.
- [ ] Every capability that exists today is migrated onto the definition source, and every surface behaves exactly as it did before, verified surface by surface.
- [ ] The migration causes no breaking change for any existing integration. Existing request and response shapes are unchanged, verified against the current live behavior.
- [ ] Adding one new capability to the definition source makes it appear on every generated surface with no further code change, demonstrated end to end with a real capability.
- [ ] Authorization and plan gating come from the definition, so a capability cannot be exposed on one surface with weaker checks than on another.
- [ ] Where a surface genuinely cannot support a capability, that is recorded as an explicit, reasoned exclusion in the definition rather than a silent absence.
- [ ] Customer-facing descriptions and error messages come from the definition and are translated.

### Mock-ups

None. No user-facing UI in this story.

### Impact on existing data

None. Definitions describe capabilities; no customer data is created, changed or deleted. The migration must not alter any existing request or response shape.

### Impact on other products

Every developer surface is touched. The mobile app and the web app are not, because they use internal endpoints rather than the public surfaces, and the story must verify that holds.

### Dependencies

- Depends on `[Research] Map every developer surface and define the parity contract`.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, no UI in this story
- [ ] Multilingual support (descriptions and error messages from the definition must be translated)
- [ ] UI theming support — N/A, no UI in this story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Add the parity guard that fails when a surface falls behind

### Description

As the engineering team, we want an automatic check that fails when a capability is missing from a surface, so that drift is caught before release instead of being reported by a customer.

### Workflow

1. A developer adds or changes a capability.
2. The check compares every surface against the definition source.
3. If a surface is missing a capability, or exposes one the definition does not have, or exposes it with a different shape, the check fails and names the surface and the capability.
4. A deliberate exclusion, recorded in the definition with a reason, passes.
5. The current parity state is visible to the team, so gaps are known rather than discovered.

### Acceptance criteria

- [ ] An automatic check compares every developer surface against the definition source and fails on any mismatch.
- [ ] The failure message names the surface, the capability and the nature of the mismatch, so it is actionable without investigation.
- [ ] A capability marked as a deliberate exclusion for a named surface, with a reason, passes the check.
- [ ] The check catches a capability present in the definition but missing from a surface, a capability on a surface but not in the definition, and a capability whose inputs or outputs differ between the definition and a surface.
- [ ] The check covers the public documentation and the OpenAPI document, so a documented-but-unbuilt or built-but-undocumented capability fails.
- [ ] The check runs automatically on every change, not on request.
- [ ] The current parity state is reported somewhere the team can see it, including the deliberate exclusions and their reasons.
- [ ] The check is verified by deliberately removing a capability from one surface and confirming it fails, then restoring it and confirming it passes.
- [ ] Every drift identified in the research is either fixed or recorded as a deliberate exclusion. The check does not go live with a list of pre-existing failures suppressed wholesale.

### Mock-ups

None.

### Impact on existing data

None.

### Impact on other products

None directly. The check governs how future changes reach the developer surfaces.

### Dependencies

- Depends on `[BE] Build the single capability definition source and generate the surfaces from it`.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, no UI in this story
- [ ] Multilingual support — N/A, no customer-facing copy in this story
- [ ] UI theming support — N/A, no UI in this story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
