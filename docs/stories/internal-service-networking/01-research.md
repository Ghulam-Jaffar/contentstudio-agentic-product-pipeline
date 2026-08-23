# 01 Research — Internal service networking via cluster-internal DNS

Date: 2026-08-23
Scope: which ContentStudio services call each other over public hostnames today, which of those callees are internal-only, and what has to change to route that traffic inside the cluster.

This doc is the local grounding for the epic. Nothing here goes into the story body.

---

## 1. Service-to-service calls that go out over the public internet

All of these are configured as public `https://` hostnames in `contentstudio-backend`, so backend-to-service traffic leaves the cluster, crosses the public internet and comes back in through a public ingress.

| Callee | Config | Reachable from the browser? |
|---|---|---|
| AI agents platform | `AI_AGENTS_BASE_URL`, `config/ai_agents.php` — default `https://ai-agents.contentstudio.io/api/v1/` | No, internal only |
| Micro-links service | `env('MICROLINKS_API')`, read directly with `env()` at 6+ call sites | No, internal only |
| Optimal times service | `OPTIMAL_TIMES_API_URL`, `app/Services/Scheduling/OptimalTimesServiceClient.php` | No, internal only |
| Go analytics / reports | `ANALYTICS_API_URL`, `config/services.php` `go_reports.base_url`, `app/Services/Analytics/GoReportsClient.php` | **Yes** — the browser calls it directly via `VITE_ANALYTICS_GO_URL` |
| Social listening Go service | `LISTENING_GO_SERVICE_URL` | Served by the same Go API surface |
| Inbox service | `INBOX_V2_API_URL` — default `https://app-inbox-api.contentstudio.io`, `app/Services/Inbox/InboxServiceClient.php` | **Yes** — the browser calls it directly via `VITE_INBOX_API_URL` |
| Gotenberg (PDF) | `GOTENBERG_BASE_URL` — already `http://gotenberg:3000` | No, already internal |

So the split is:

- **Internal only, safe to take fully private:** AI agents, micro-links, optimal times, Gotenberg (already done).
- **Internal and browser-facing:** analytics/listening Go API, inbox service. These must keep a public ingress for the browser, but backend-to-service traffic should still resolve internally.

Gotenberg is the existing proof the pattern works. It is already addressed by an in-cluster hostname.

## 2. Why this matters beyond tidiness

1. **Attack surface.** `ai-agents.contentstudio.io` and the micro-links host answer requests from anywhere. Nothing about their function requires that. Their auth is a shared API key (`AI_AGENTS_API_KEY`) rather than a per-user identity.
2. **Latency and cost.** Every internal call pays DNS resolution, TLS handshake, public ingress and egress. The AI agents timeout is 300 seconds, so these are not trivial calls.
3. **Environment bleed.** `.env.example` carries an explicit warning on this: `config/ai_agents.php` defaults `base_url` to a production host while the example file sets a staging host, so an environment that forgets the variable silently proxies its traffic somewhere it should not. Internal DNS names are per-cluster, which makes this class of mistake structurally impossible.
4. **Reachability of the private hosts is currently the only thing protecting some of them.** Anything that relies on "nobody knows the URL" is not protected.

## 3. What is not just a hostname swap

- **Callbacks run the other way too.** `GoReportsClient` passes `services.go_reports.callback_url` plus `webhook_secret`, so the Go service calls back into Laravel. Both directions need an internal name, and the callback URL has to be one the callee can actually resolve.
- **`env()` is called at runtime in application code** for micro-links (`env('MICROLINKS_API')` inside `app/Libraries/Discovery/Discovery/Video/Video.php`, `app/Libraries/Inbox/HelperClasses/LinkedinHelper.php`, `app/Libraries/Publish/Helper/SocialHelper.php`, `app/Libraries/Discovery/Composer/ImagesAPI.php`, `app/Http/Controllers/Planner/HelperController.php`, `app/Http/Controllers/Discovery/Video/VideoController.php`). With config caching enabled, `env()` outside config files returns null. These need to move to config keys as part of the change.
- **One hardcoded public URL bypasses config entirely:** `app/Http/Controllers/Planner/HelperController.php` calls `https://pro.microlink.io/?url=...` directly. That is a third-party SaaS, not our micro-links service, so it stays external. It should not be confused with `MICROLINKS_API` during the inventory.
- **TLS.** Internal traffic over plain HTTP inside the cluster versus internal TLS is a decision, not a default.
- **Webhooks from third parties** (Paddle, Meta, Google) must keep hitting public ingress. Only service-to-service traffic moves.
- **Local development** must keep working when there is no cluster. The config needs to degrade to a reachable value outside the cluster.

## 4. Cluster context

`social-inbox-manager` ships `k8s/` and `argo/` directories, and `contentstudio-social-analytics-go` has a `sidecar/` and `scripts/`, so the cluster and its deployment tooling already exist. The work is configuration and routing, not new infrastructure.

## 5. Open questions for the ticket

1. Complete list of caller-to-callee pairs, including service-to-service calls that do not involve `contentstudio-backend` at all, for example ai-agents to the public API, or the inbox service to the backend.
2. Naming convention for in-cluster service names, and whether it is namespaced per environment.
3. Internal TLS: required or not.
4. Do the browser-facing services keep a single ingress with authorization in front of it (this overlaps the service authorization epic), or a public ingress plus a separate internal address.
5. Once internal, should the private services drop their public ingress entirely, and what is the verification that they are actually unreachable from outside.
