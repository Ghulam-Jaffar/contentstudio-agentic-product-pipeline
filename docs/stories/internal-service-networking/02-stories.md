# Epic + stories — Internal service networking via cluster-internal DNS

**Scope of this doc:** the epic and **one** story. Deliberately kept at an overall level, not a technical design. Nothing is pushed to any tracker. The PO creates this by hand.

---

## Epic: Internal service networking via cluster-internal DNS

ContentStudio is made of several services that call each other: the main backend, the AI agents platform, the micro-links service, the analytics service, the optimal-times service, the social inbox service. Some of those are things a customer's browser talks to. Most of them are not: they exist only to serve other ContentStudio services.

Today almost all of that traffic goes out over the public internet and comes back in. A call from the backend to the AI agents platform leaves our network, resolves a public address, and re-enters through a public entry point, even though both sides are running in the same cluster. That means services that no customer should ever be able to reach are answering requests from anywhere on the internet, and every internal call pays the cost of a round trip it does not need.

We already run a cluster, and one service is already addressed internally, so the pattern is proven. This epic applies it everywhere it should apply: services that are internal-only stop being reachable from outside, and internal traffic between all our services resolves inside the cluster. The services a browser genuinely needs, such as analytics and the social inbox, keep their public entry point for the browser, but calls between our own services still go the internal route.

There is a second benefit worth naming. Because internal addresses are specific to the cluster they run in, an environment cannot accidentally point at another environment's service. That has been a real risk, and configuration alone has not prevented it.

### Out of scope

- Third-party webhooks. Payment, social platform and other inbound callbacks must keep arriving on public entry points and are untouched.
- Calls to genuinely external services. Those stay external.
- How the browser-facing services authorize requests. That is the separate service authorization epic. This epic only changes how our services reach each other.
- Any change to what the services do or what they return.

### Stories

1. `[BE]` Route service-to-service traffic through cluster-internal addresses and close public access to internal-only services

---

## [BE] Route service-to-service traffic through cluster-internal addresses and close public access to internal-only services

### Description

As the engineering team, we want every call between ContentStudio's own services to travel inside the cluster, and services that exist only to serve other services to stop being reachable from the public internet, so that we reduce our exposed surface, cut needless latency on internal calls, and make it impossible for one environment to accidentally call another environment's service.

### Workflow

1. The team lists every place one ContentStudio service calls another, in both directions, including callbacks a service makes back into the main backend.
2. For each callee, the team decides whether it is internal-only or whether a customer's browser or a third party genuinely needs to reach it.
3. Services judged internal-only are given a cluster-internal address, and every caller is switched over to it.
4. Services that must stay reachable from the browser keep their public entry point for the browser, but calls from our own services are switched to the internal address as well.
5. The team verifies, from outside the network, that the internal-only services no longer answer.
6. The team verifies, in the product, that the features depending on those services still behave exactly as before: AI generation, link previews and metadata, best time to post, analytics, social listening, the social inbox, and PDF report generation.
7. The team confirms local development and every non-production environment still work, and that no environment is configured to reach another environment's service.

### Acceptance criteria

- [ ] A documented inventory lists every service-to-service call in the product, naming the caller, the callee, the direction, and whether the callee is internal-only or must stay publicly reachable.
- [ ] The inventory distinguishes our own services from genuinely third-party services, so third-party calls are not caught up in the change.
- [ ] Every call between our own services resolves through a cluster-internal address, including callbacks a service makes back into the main backend.
- [ ] Services identified as internal-only return no response to a request originating from outside the network, verified from an external network and recorded as evidence.
- [ ] Services that must remain reachable from the browser still are, and continue to work for logged-in users and for publicly shared report links.
- [ ] Third-party inbound webhooks, including payment and social platform callbacks, continue to arrive and be processed.
- [ ] Service addresses come from configuration in every case. No service address is hardcoded in application code, and none is read in a way that breaks when configuration is cached.
- [ ] No environment can resolve another environment's service. This is verified per environment, not assumed.
- [ ] Local development works without a cluster, with a documented way to point at the services a developer needs.
- [ ] All features that depend on the affected services are verified working after the change: AI content generation, link previews and metadata, best time to post, analytics dashboards and reports, social listening, the social inbox, and PDF report generation.
- [ ] Report generation that involves a callback from the analytics service back into the backend completes end to end.
- [ ] The change can be rolled back per service without a code deploy, and the rollback path is documented.
- [ ] Internal call failures are visible to the team, so a misrouted internal address surfaces as an alert rather than as a silently broken feature.
- [ ] The final routing and the internal naming convention are documented for the next service the team adds.

### Mock-ups

None. No user-facing change.

### Impact on existing data

None. No data is created, changed, migrated or deleted. Only how services address each other changes.

### Impact on other products

None directly. The mobile app, the Chrome extension and integrations keep calling the same public entry points they call today. The change is limited to how our own services reach each other, so the verification must confirm those clients are unaffected.

### Dependencies

- Coordinates with the service authorization epic, since one of the approaches under assessment there puts a single authorization point in front of the browser-facing services.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, no UI in this story
- [ ] Multilingual support — N/A, no user-facing copy in this story
- [ ] UI theming support — N/A, no UI in this story
- [ ] White-label domains impact review (white-label domains resolve to the same services and must be verified after the change)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
