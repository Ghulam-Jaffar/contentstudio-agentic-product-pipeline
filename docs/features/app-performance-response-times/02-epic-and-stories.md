# Epic and Story: application response and query time

**Date:** 2026-09-07

One epic and one research story. The follow-up stories are deliberately not written yet. They come out of the research story's findings, so writing them now would be guessing.

---

## Epic

### Title

**Reduce application response and query times across the app**

### Goal

Response times and query times are slow across ContentStudio, and it is felt everywhere rather than on one screen. Today we cannot say which part of a slow page is the database, the API, a downstream service, or the browser, because server-side performance tracing is not switched on and there is no query-level instrumentation anywhere in the application. This epic makes the system measurable, finds where the time actually goes, and then removes it, one evidenced fix at a time.

### Approach

The epic runs in two phases.

**Phase 1, this batch: measure.** A single research story turns on the measurement we are missing, captures a baseline across the app's hot path, identifies the root cause behind the worst offenders, and produces a ranked list of follow-up stories with expected impact.

**Phase 2, created after Phase 1 reports: fix.** Individual stories for the specific problems the research finds. Each one references the baseline so its effect can be proven. Likely shapes based on what we already know: missing database indexes, N+1 query patterns in list views, an oversized application boot payload, uncached repeated reads, and frontend bundle size. None of these are committed to until the research confirms them.

### Scope

- Server-side and database performance measurement, switched on and kept on.
- A recorded baseline for the app's main surfaces: application boot, dashboard, planner, composer, analytics, inbox, approvals, discovery, settings.
- Root cause identification for the slowest and highest-traffic paths.
- A ranked, sized backlog of follow-up work.
- Later, the follow-up fixes themselves.

### Out of scope

- Rewrites. This epic removes latency from the system we have.
- Infrastructure resizing as a first move. We find out where time goes before we buy more capacity.
- Any change to product behaviour or UI. Users should notice speed, nothing else.
- Mobile app performance. The Flutter app has its own surfaces and its own measurement story, and is not covered here.

### Success measure

Agreed p95 targets per surface, met and held, measured against the baseline the research story records. The targets themselves are set as part of the research story rather than guessed now.

### Stories in this batch

1. **[BE] Measure and diagnose slow response and query times across the app**

Follow-up stories will be added to this epic once that story reports.

---

## Story 1

### Title

**[BE] Measure and diagnose slow response and query times across the app**

### Description

As an engineering team, we want server-side performance tracing switched on and a measured baseline of where request time is actually spent, so that we stop guessing at the cause of app-wide slowness and can plan performance work against evidence instead of intuition.

This is a research story. It produces measurement, findings and a ranked backlog. It does not fix the problems it finds, beyond changes needed to make the system measurable.

---

### Workflow

The reader of this story is an engineer, so the steps below are the investigation itself.

**Phase 1: make the system measurable**

1. Confirm what the production environment actually sets for the server-side trace sample rate and for queue job tracing. Both currently default to off in the codebase, so if they have never been set, no server-side timing data exists at all.
2. Raise the server-side trace sample rate to an agreed level and enable queue job tracing. Agree the level with whoever owns the self-hosted error and tracing infrastructure, since the ingest volume lands on hardware we run.
3. Verify that a single browser request now produces one connected trace showing browser time and server time together. The browser already sends trace headers to the API, so this should link up as soon as the server side is sampling.
4. Add per-request database instrumentation so that every traced request records how many queries it ran and how long they took in total. This is what makes repeated-query problems visible without reading code.
5. Confirm the instrumentation is visible in the tracing dashboard and costs an acceptable amount of overhead. Roll it back if it does not.

**Phase 2: capture a baseline**

6. Over an agreed observation window of at least one full week of normal traffic, record for each of the app's main surfaces: p50, p75, p95 and p99 server response time, request volume, query count per request, and total query time per request. Averages are not acceptable, since they hide exactly the problem we are chasing.
7. Surfaces to cover, at minimum: application boot, dashboard, planner calendar and list views, composer open and preview, analytics overview and per-platform reports, inbox conversation list and thread, approvals, discovery, and settings.
8. Pull the database profiler's slow query list and its index recommendations for the same window. This needs no code change and can start immediately, in parallel with Phase 1.
9. Pull the analytics data warehouse query log for the same window, so the analytics read path can be attributed across the API, the analytics service and the warehouse.
10. Capture a frontend baseline for the same surfaces: initial JavaScript payload size, Largest Contentful Paint, Interaction to Next Paint, and number of network requests per navigation.
11. For each surface, split the total wall-clock time a user waits into browser time and server time, so follow-up work is aimed at the right layer.
12. Check whether slowness correlates with workspace size, number of connected accounts, or selected date range. If it does, record the thresholds where things degrade.

**Phase 3: diagnose and report**

13. Take the ten worst offenders, ranked by p95 multiplied by request volume rather than by p95 alone, and identify a root cause for each. Categorise every one as a missing or unusable index, a repeated query pattern, an oversized payload, a missing cache, a slow downstream service call, or a frontend cost.
14. Write up the findings, including the baseline numbers, so future work can be measured against them.
15. Propose p95 targets per surface for the epic to be closed against, and get them agreed.
16. Produce a ranked list of follow-up stories, each with the surface it affects, the root cause category, the expected improvement, and a rough effort estimate.
17. Walk the findings through with the team so the follow-up stories can be written and prioritised.

---

### Acceptance criteria

**Measurement is in place**

- [ ] The production server-side trace sample rate and queue job tracing setting are confirmed and documented, including what they were before this story
- [ ] Server-side performance tracing is enabled in production at a sample rate agreed with whoever owns the tracing infrastructure
- [ ] Queue job tracing is enabled, so async work appears in traces alongside web requests
- [ ] A single user action produces one connected trace covering both browser time and server time
- [ ] Every traced request records the number of database queries it made and the total time spent in those queries
- [ ] The overhead added by the new instrumentation is measured and recorded, and is agreed to be acceptable before the story is closed
- [ ] Tracing stays enabled after this story closes, so later fixes can be verified against it

**Baseline is recorded**

- [ ] A baseline covering at least one full week of normal production traffic is recorded
- [ ] For every surface listed in the workflow, the baseline records p50, p75, p95, p99, request volume, query count per request and total query time per request
- [ ] No figure in the report is presented as an average
- [ ] The baseline records the frontend measures for the same surfaces: initial JavaScript payload size, Largest Contentful Paint, Interaction to Next Paint, and requests per navigation
- [ ] For each surface, the user's total wait is split into browser time and server time
- [ ] The database profiler's slow query list and index recommendations for the same window are captured
- [ ] The analytics warehouse query log for the same window is captured, and the analytics read path is attributed across the API, the analytics service and the warehouse
- [ ] Whether latency correlates with workspace size, connected account count or selected date range is answered either way, with the degradation thresholds recorded when it does

**Findings are actionable**

- [ ] The ten worst offenders are identified, ranked by p95 multiplied by request volume, not by p95 alone
- [ ] Every one of the ten has a named root cause, categorised as a missing or unusable index, a repeated query pattern, an oversized payload, a missing cache, a slow downstream service call, or a frontend cost
- [ ] A ranked list of follow-up stories is produced, each stating the surface affected, the root cause category, the expected improvement and a rough effort estimate
- [ ] Any quick win that is a configuration change with no product risk is identified and called out separately, so it can ship without waiting for the rest of the epic
- [ ] Proposed p95 targets per surface are stated and agreed, and become the exit criteria for the epic
- [ ] Findings are walked through with the team, and the follow-up stories are ready to be written from the report

**Boundaries**

- [ ] No product behaviour changes as a result of this story
- [ ] No user-facing interface changes as a result of this story
- [ ] No performance fix is implemented in this story, other than changes required to make the system measurable
- [ ] Any secret, token, message body, post content or personal data that would otherwise be captured by the new tracing is excluded from what gets sent

---

### Mock-ups

Not applicable. This story produces measurement and a written report, with no user-facing interface.

---

### Impact on existing data

None. No stored data is created, modified or deleted. The story adds observability output, which is written to the tracing system rather than to product data.

One thing to watch: the new tracing captures request paths, query shapes and timings. It must not capture query values, tokens, message bodies or post content, since that would put customer data into the tracing system. This is covered in the acceptance criteria and needs an explicit check before the sample rate is raised.

---

### Impact on other products

- **Web app:** no functional change. Users should eventually see the app get faster, which is the point of the epic, but nothing changes as a result of this story alone.
- **Mobile apps:** none. Mobile performance is not in scope for this epic. Mobile does share the same API, so anything the follow-up work fixes on the server side benefits the app for free.
- **Chrome extension:** none directly. It shares API endpoints, so the same free benefit applies later.
- **Public API:** none. Customers on the public API share the same backend, so server-side improvements reach them too.
- **Infrastructure and cost:** raising the trace sample rate increases ingest and storage on self-hosted tracing infrastructure. This is a real cost that needs sign-off before the rate goes up, and it should be reviewed once the observation window has run.

---

### Dependencies

- Sign-off from whoever owns the self-hosted error and tracing infrastructure on the trace sample rate and the resulting ingest volume.
- Read access to the production database profiler and its index recommendations.
- Read access to the analytics data warehouse query log.
- A full week of representative production traffic. Avoid running the observation window across a holiday period or an incident.
- Blocks every follow-up performance story in this epic. None of them should be written or estimated before this reports.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled) — N/A, no user-facing copy in this story
- [ ] UI theming support (default + white-label, design library components are being used) — N/A, no user-facing interface in this story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
