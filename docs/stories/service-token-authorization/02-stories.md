# Epic + stories — Service token authorization for browser-reachable services

**Scope of this doc:** the epic and **one** story, a research ticket. No implementation stories yet. Nothing is pushed to any tracker. The PO creates this by hand.

---

## Epic: Service token authorization for browser-reachable services

ContentStudio's web app does not talk to a single backend. As well as the main API, the browser calls the analytics service and the social inbox service directly. Those two services are reachable from the internet, and each one checks the caller's token on its own.

The check they do is a signature check: it confirms the token is one we issued and identifies who it belongs to. What it does not confirm is that the person holding the token is allowed to see the workspace they asked for. The workspace is named in the request itself, so a caller can change it. Anyone holding a valid token can therefore ask those services for another customer's data, and today they will get an answer. The main API does not have this problem, because it checks workspace membership and the user's role on every request. That model simply was never extended to the two services the browser also calls.

The social inbox service has already started closing this: it has a workspace membership check that reuses the same team data the main API uses. It sits behind a rollout flag, it only applies to requests that name a workspace, and it checks membership rather than role. The analytics service has no equivalent at all.

This epic makes authorization consistent across every service the browser can reach. The goal is that a request is only served when the caller is proven to be the user the token was issued to, and that user is proven to have access to the specific workspace and accounts the request touches, with the same rules the main API already applies. It also has to keep working for the paths that are deliberately not a logged-in user: publicly shared analytics report links, and API keys used by integrations.

**This epic starts with research only.** The design decision here affects every service and every integration, so the first ticket is a proper options assessment with a recommendation. Implementation stories are written after that is reviewed.

### Out of scope for now

- Any implementation or rollout. The approach has to be chosen first.
- Changing how users log in, or the login token's lifetime, beyond what the research recommends.
- The internal-only services that the browser never calls. Those are covered by the separate internal service networking epic.

### Stories

1. `[Research]` Assess how to prove workspace access, not just a valid token, on every service the browser can reach

---

## [Research] Assess how to prove workspace access, not just a valid token, on every service the browser can reach

### Description

As the engineering team, we want a documented assessment of how every browser-reachable ContentStudio service should verify that the caller genuinely owns the token they present and genuinely has access to the workspace and accounts they are asking about, so that we can close the gap where a valid token is currently enough to read another customer's data, and do it once, consistently, instead of per service.

### Workflow

1. The team lists every service the browser can call directly, and for each one records what it checks today before it serves a request.
2. For each service, the team records what identifies the caller, where the requested workspace comes from, and whether anything confirms the caller may access that workspace.
3. The team lists every other way those services can be called: publicly shared report links, integration API keys, staff acting on behalf of a customer, and internal service-to-service calls. Each of these has to keep working, so each is documented as a case the design must handle.
4. The team documents the authorization rules the main API already applies, so the target model matches them rather than inventing a second set of rules.
5. The team writes down the concrete risk in plain terms: what an actor holding a valid token can reach today on each service, including anything that affects billing or credit consumption, so the priority is based on real exposure and not a guess.
6. The team compares the realistic options for fixing it, with the trade-offs of each: cost to build, added latency, how much duplicated logic it creates, and how it behaves if one service is unavailable.
7. The team recommends one option and explains why.
8. The team defines the rollout: what is checked first, how it is verified in a live environment without breaking existing customers, how it is switched on safely, and how the team confirms nothing legitimate was blocked.
9. The output is reviewed and approved. Implementation stories are written from it afterwards.

### Acceptance criteria

- [ ] The document lists every service the browser can call directly and, for each, what it verifies before serving a request today.
- [ ] For each service, the document states where the requested workspace comes from in the request and whether the caller's access to it is verified.
- [ ] The document states, per service, what an actor holding someone else's valid token can currently read or do, including any action that consumes credits or affects billing.
- [ ] The document records the partial protection that already exists on the inbox service, including that it is behind a rollout flag, and states its current on or off state per environment.
- [ ] The document identifies the request paths that carry no workspace at all and are therefore unprotected by a workspace check, and says how the target model handles them.
- [ ] The document lists every non-browser caller that must keep working: publicly shared analytics report links, integration API keys, staff acting on behalf of a customer, and internal service-to-service calls.
- [ ] The authorization rules the main API already applies are documented, including workspace membership, the user's role in that workspace, per-account access within a workspace, and the sample workspace exception, so the target model can match them.
- [ ] The document confirms whether role and per-account access, not just workspace membership, must be enforced on the other services, with examples of where the answer changes what a user can see.
- [ ] The document covers the token itself: how many different kinds of token the services accept, whether a token issued for one purpose can be used against another service, and whether tokens can be revoked when a user signs out or loses access.
- [ ] At least three realistic approaches are compared, each with build cost, latency impact, duplicated-logic risk, and failure behavior if a dependency is unavailable.
- [ ] One approach is recommended, with the reasoning stated.
- [ ] A rollout plan is documented: the order services are protected in, how enforcement is verified against real traffic before being enforced, how it is switched on, how it is rolled back, and how the team confirms no legitimate request was blocked.
- [ ] The plan states how the team will detect and alert on blocked cross-workspace attempts after rollout.
- [ ] Open questions and recommendations are captured for the review gate.
- [ ] The document is reviewed and approved before any implementation story is written.

### Mock-ups

None. This is a research ticket with no UI.

### Impact on existing data

None. Nothing is changed, migrated or deleted by this ticket. The document does record which shared data the services read to answer "may this user access this workspace", so the implementation phase knows what it depends on.

### Impact on other products

The mobile app and the Chrome extension also call ContentStudio services with a token, and integrations call them with API keys. The research must confirm what each of those sends today so the chosen approach does not lock them out. No product changes yet.

### Dependencies

- Sequences alongside the internal service networking epic, since one of the candidate approaches puts a single authorization point in front of these services.
- Sequences alongside the existing public analytics API rollout work, so integration API keys are covered by the same model.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, research ticket with no UI
- [ ] Multilingual support — N/A, research ticket with no user-facing copy
- [ ] UI theming support — N/A, research ticket with no UI
- [ ] White-label domains impact review (white-label domains call the same services, and shared report links are served under them)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
