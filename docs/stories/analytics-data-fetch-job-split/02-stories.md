# Stories: Analytics Data Fetch Job Split (Daily + Bi-weekly)

---

## [BE] Split Analytics Data Fetching Into Daily and Bi-Weekly Jobs

### Description:
As the ContentStudio analytics system, I need two separate data-fetching schedules — a lightweight daily job and a comprehensive bi-weekly job — so that analytics data stays up to date without unnecessarily re-fetching historical data every single day.

Currently, the analytics pipeline runs a data fetch job daily for all connected social platforms (Facebook, Instagram, LinkedIn, TikTok), and each run pulls the last 2 weeks of data. This means the system is re-fetching the same historical data every day, which is wasteful.

The new approach splits this into two jobs:
1. **Daily job** — runs every day and fetches only the current day's data for all platforms
2. **Bi-weekly job** — runs every two weeks and fetches the full 2-week window of data for all platforms, ensuring any missed or updated historical data is captured

---

### Workflow:

This is a background system job — there is no direct user interaction. The workflow describes the system behavior:

1. Every day, the daily analytics job runs for all platforms (Facebook, Instagram, LinkedIn, TikTok) and collects only that day's data for all connected accounts
2. Every two weeks, the bi-weekly sync job runs for all platforms and collects the full past 2 weeks of data for all connected accounts
3. Analytics data visible to users in the Analytics section reflects both the daily updates and the periodic full syncs

---

### Acceptance criteria:

- [ ] A daily job runs for all supported platforms (Facebook, Instagram, LinkedIn, TikTok) and fetches only the current day's data
- [ ] A bi-weekly job runs for all supported platforms and fetches the full 2-week window of data
- [ ] Both jobs run independently and do not interfere with each other
- [ ] If the bi-weekly job runs on the same day as the daily job, both complete successfully without conflict
- [ ] Analytics data in the product reflects data from both jobs correctly (no gaps, no duplicates)
- [ ] All four platforms are covered by both jobs

---

### Mock-ups:
N/A — no UI changes.

---

### Impact on existing data:
The data collected remains the same. The change is only in how frequently full historical syncs happen. Daily runs will now store less data per execution (current day only), while the bi-weekly run handles the full historical window. No existing analytics data is deleted or modified.

---

### Impact on other products:
- **Web app (Analytics):** No visible change expected — data should continue to appear correctly
- **Mobile apps:** No impact
- **Chrome extension:** No impact
- **White-label:** No impact

---

### Dependencies:
None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support — N/A, no user-facing strings
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)
