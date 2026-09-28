# Stories: Sentry full transaction tracing and observability for analytics

---

# [BE] Extend Sentry across the analytics service for full transaction tracing and observability

### Description

As a ContentStudio engineer or support agent, I want every part of the analytics service to report its activity to Sentry, not just its errors. Then I can see how analytics is performing, spot slowdowns and stalled syncs before customers notice them, and explain why a customer's analytics are late or missing.

Today Sentry only sees analytics errors, plus a sample of API requests. Most of the analytics work happens behind the scenes: scheduling syncs, fetching data from each social platform, processing it and saving it. None of that is visible. When a customer reports stale analytics, nothing shows where their sync got stuck or how long each step took.

---

### Workflow

1. A customer reports that their Instagram analytics haven't updated since yesterday.
2. A support agent or engineer opens Sentry and searches by the customer's workspace or social account.
3. They see every recent sync for that account as one timeline: scheduled, fetched from Instagram, processed and saved. They can see how long each step took and where it stopped or failed.
4. They see the cause, for example that Instagram's API took 40 seconds to respond and then rejected the request, and can tell the customer what happened.
5. Separately, when an analytics job stops running, slows down sharply or starts failing, the team gets an alert in its channel without anyone watching Sentry.
6. On any day, the team opens a shared Sentry dashboard to check the health of analytics for every platform at a glance.

---

### Acceptance criteria

**Coverage**
- [ ] Every part of the analytics service reports activity to Sentry: API requests, scheduling, fetching from each social platform, processing, saving, report generation, competitor analytics and scheduled jobs
- [ ] This covers every supported platform: Facebook, Instagram, LinkedIn, X, TikTok, YouTube, Pinterest, Google Business Profile, Meta Ads, Google Ads and Social Listening
- [ ] All transactions are captured, not just errors. Successful work shows up in Sentry with how long it took

**Following one sync end to end**
- [ ] One account's sync appears in Sentry as a single connected timeline from scheduling to saving, even though several services handle it
- [ ] The timeline shows how long each step took, including time spent waiting on the social platform, the databases and the cache
- [ ] Sentry can search activity by workspace, social account, platform and step
- [ ] A support agent can find a specific customer's recent syncs in Sentry in under a minute using only their workspace or account

**Scheduled jobs**
- [ ] Every scheduled analytics job is monitored in Sentry
- [ ] The team is alerted when a scheduled job does not run on time, runs far longer than usual, or fails

**Alerts**
- [ ] The team gets an alert in its channel for: a new kind of error, a sudden rise in errors, a platform's syncs slowing down noticeably, and a missed or failed scheduled job
- [ ] Alerts say which platform and which step are affected

**Dashboard**
- [ ] A shared Sentry dashboard shows, for each platform: how many syncs ran, how many failed, and how long they took
- [ ] The dashboard shows the slowest steps and the most common failures

**Releases and environments**
- [ ] Every event is labelled with its environment (production or staging) and the release it came from
- [ ] The team can tell whether a spike in errors or slowness started with a specific deploy

**Control and cost**
- [ ] The share of transactions sent to Sentry can be set per environment and per service without a code change. Errors are always sent in full
- [ ] Staging can send all transactions while production sends a sample. The production rate is agreed with the team, keeping the Sentry bill within budget
- [ ] Sentry reporting can be turned off entirely through configuration without breaking analytics

**Safety**
- [ ] No access tokens, passwords, API keys or customer post content reach Sentry
- [ ] Error reporting that already works keeps working, and no error is reported twice
- [ ] Reporting to Sentry does not noticeably slow down analytics syncs. If Sentry is unreachable, analytics keeps working

---

### Mock-ups

N/A

---

### Impact on existing data

None. No customer data is changed. Only the activity information sent to Sentry is new.

---

### Impact on other products

None for customers. There is no change to the web app, mobile app, Chrome extension or white-label domains. The Sentry bill will grow with the number of transactions sent, which is why the sampling control above is part of this ticket.

---

### Dependencies

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (N/A, backend-only story)
- [ ] Multilingual support (N/A, no user-facing text)
- [ ] UI theming support (N/A, no UI)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, no new or changed API)
