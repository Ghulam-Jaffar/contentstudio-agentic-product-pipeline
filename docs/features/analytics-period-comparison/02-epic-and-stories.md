# EPIC: Analytics period comparison: choose what your numbers are compared against

Let users see and choose what their analytics are compared against, everywhere a change is shown.

Today every card, rollup and report compares the selected dates with "the same number of days just before". Users can't see those dates or the previous value, and they can't choose a different comparison, such as last month or the same month last year. When the comparison period has no data, cards show 0%. Customers have repeatedly reported this as wrong data. In one report, engagement dropped month over month but the report showed +27.55%. In another, a Q3 report showed reach up when Q3 versus Q2 was down. Another customer asked outright for a custom comparison range.

This epic adds a "Compare to" choice next to the date range: Previous period, Same period last year, Custom range or No comparison. The choice is shown with its dates, so every change can be checked. It carries into exported, emailed and scheduled reports, where a schedule moves the comparison forward with each send, and the PDF prints both ranges. Cards show "No data" instead of a misleading 0%.

The epic starts with a research story, because the comparison is calculated in several places today and the platforms keep different amounts of history. The build stories will be written from its findings.

Prototype: https://claude.ai/artifact/GkPcaAzYtjeL5jaedTAeyp

Related support tickets: CONT-3152, CONT-2392, CONT-4294, CONT-4343.

---

| Story | Covers |
|---|---|
| [Research] Work out how analytics comparisons should work across dashboards, reports and schedules | Current behaviour, the comparison rules, data limits, and the follow-up stories for this epic |

---

## [Research] Work out how analytics comparisons should work across dashboards, reports and schedules

### Description:

As the **analytics team**, we want to understand exactly how comparisons work today and what it takes to let users choose them so that we can plan the build stories for this epic with confidence.

Every change percentage in analytics compares the selected dates with a period users can't see or choose. Support tickets show the cost:

- **CONT-3152:** a customer asked what "previous time period" means and for a custom comparison range.
- **CONT-2392:** engagement fell month over month, but the report showed +27.55%.
- **CONT-4294:** a Q3 report showed reach up when Q3 versus Q2 was down.
- **CONT-4343:** an Instagram Q3 versus Q2 engagement rate didn't add up.

The direction is set on the prototype canvas:

- A "Compare to" choice: Previous period, Same period last year, Custom range or No comparison.
- The comparison is saved as a rule, so scheduled reports move forward with each send.
- Each kind of widget gets a clear treatment.
- "No data" replaces a misleading 0%.

This story answers what has to be true for that to work and be trustworthy, and turns the answer into the build stories.

Prototype: https://claude.ai/artifact/GkPcaAzYtjeL5jaedTAeyp

### Workflow:

1. Map every place a comparison is calculated today, across the dashboards, the report engines, scheduled reports and the public API. Note where they give different windows for the same dates.
2. Check how much history each platform keeps, and what that means for "Same period last year" and for long ranges.
3. Work through the open decisions below with the PO, using the prototype canvas, and record each answer.
4. Propose the build stories for the epic ([BE], [FE], [Design], and any public API, CLI or MCP work) and review them with the PO.

### Acceptance criteria:

**Today's behaviour**

- [ ] A written summary lists every place that calculates a comparison window today. For each, it gives the window it produces for the same sample dates, and says where results differ (including the window that comes out one day shorter)
- [ ] The summary explains the CONT-2392 and CONT-4294 results in plain language: which window was used, and what the customer expected

**Data limits**

- [ ] For each platform (Facebook, Instagram, LinkedIn, X, TikTok, YouTube, Pinterest, Threads, Bluesky, Google Business Profile, Meta Ads, Google Ads), the summary states how far back data is available. It says whether "Same period last year" can work, and what users see when it can't
- [ ] A recommendation on partial comparison data: how it is labelled, and whether the change is hidden below a threshold (for example, under half the days)

**Open decisions**

Each has a recommended answer, agreed with the PO:

- [ ] Monthly reports, "Previous period": the previous calendar month, or the same number of days?
- [ ] Weekly reports, "Same period last year": the same dates, or the same weekdays 52 weeks back?
- [ ] Custom range on scheduled reports: not allowed, or allowed as a fixed baseline that every send compares against?
- [ ] Ranges of different lengths: compare totals with a warning, or compare per-day averages?
- [ ] Grouped reports: when accounts are added to or removed from a group, do both periods use the same set of accounts?
- [ ] Is custom comparison available on every plan, or only on some?

**Coverage and output**

- [ ] A list of every widget in Overview, each social platform, Meta Ads, Google Ads, Campaign & Label and competitor analytics, with its comparison treatment: change and previous value, change in the header, overlay line, change column, or not compared
- [ ] A recommendation on how a report stores the comparison rule, so exports use the dates at the moment they run and schedules move forward with each send
- [ ] Draft build stories for this epic, reviewed with the PO, covering:
    - dashboards
    - reports and schedules
    - the "No data" and "Partial" states
    - design
    - the public API, CLI and MCP
- [ ] Findings are posted as a comment on this ticket

### Mock-ups:

Prototype canvas: https://claude.ai/artifact/GkPcaAzYtjeL5jaedTAeyp

### Impact on existing data:

None. Research only.

### Impact on other products:

None from this story. The findings will cover the web app dashboards, PDF reports, scheduled reports, AI Insights and the public developer surfaces.

### Dependencies:

None.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness: N/A, research only
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support: N/A, research only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (any new or changed API is reflected in the public API, CLI, MCP server and automation apps, N/A when nothing API-facing changes)
