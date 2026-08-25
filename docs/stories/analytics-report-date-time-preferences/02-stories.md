# Analytics Reports: Respect Date, Time and Timezone Preferences · Story

---

## [BE] Show dates and times in analytics reports using the user's own format and timezone

### Description

As a ContentStudio user, I want every date and time in my analytics reports to appear in the date format, clock format and timezone I have already chosen in my settings, so that a downloaded report reads the same way as the rest of the product and I do not have to translate dates in my head before sharing it with a client.

Today the product honours these preferences everywhere on screen, but analytics report PDFs ignore them. A user who has chosen a day-first date and a 24-hour clock still receives a report showing an American month-first date and a 12-hour time. The "generated on" stamp on the report cover is worse: it uses the server's clock rather than the workspace timezone, so for users far from the server it can show the wrong day for reports produced late in the evening or early in the morning.

This is most visible for clients outside the United States, who receive reports that look foreign next to the rest of their account, and for agencies who send these PDFs to their own customers.

### Workflow

1. A user sets their preferred date format and clock format in their profile settings, and their workspace has a timezone set.
2. The user generates an analytics report, either by downloading one directly or by receiving a scheduled one by email.
3. Every date on the report cover, on chart labels and in tables appears in the format the user chose.
4. Every time on the report appears in the clock format the user chose.
5. The "generated on" stamp on the cover shows the moment the report was produced, expressed in the workspace timezone.
6. If the report is in a language other than English, month names still appear in that language, in the user's chosen ordering.

### Acceptance criteria

- [ ] The date range shown on the report cover uses the user's saved date format
- [ ] The "generated on" stamp on the report cover uses the user's saved date format
- [ ] The "generated on" stamp uses the user's saved clock format, showing a 24-hour time for users who have chosen a 24-hour clock
- [ ] The "generated on" stamp is expressed in the workspace timezone, not the server's timezone
- [ ] A report generated near midnight shows the correct calendar day for the workspace timezone, not the server's day
- [ ] Dates shown on chart labels inside the report use the user's saved date format
- [ ] Dates shown in tables inside the report use the user's saved date format
- [ ] All four supported date formats render correctly on the report: month-first short, month-first long, year-first, and day-first
- [ ] Both clock formats render correctly on the report: 12-hour and 24-hour
- [ ] When a user has no date format saved, reports fall back to the same default the rest of the product uses, so the report and the screen agree
- [ ] When a user has no clock format saved, reports fall back to a 12-hour clock
- [ ] When a workspace has no timezone set, the report falls back to the same timezone the rest of the product uses for that workspace
- [ ] For a report generated in a language other than English, month names still appear translated, in the user's chosen date ordering
- [ ] Reports downloaded on demand and reports delivered by scheduled email both follow the user's preferences
- [ ] A report opened through a shared link shows the same dates and times as the downloaded version of that report
- [ ] The data covered by the report is unchanged. Only how dates and times are displayed changes, and the reporting period itself still covers exactly the same days it does today
- [ ] Hourly breakdown charts continue to show hour buckets in the workspace timezone as they do today, unchanged

### Mock-ups

N/A. No layout changes, only the way existing dates and times are written.

### Impact on existing data

None. No stored data changes. Reports already generated keep whatever formatting they were produced with, since the PDFs are fixed files. Only newly generated reports pick up the preferences.

### Impact on other products

- **Mobile app:** no impact. Reports are produced server-side and are not composed in the app.
- **Chrome extension:** no impact.
- **White-label:** reports sent under a white-label brand follow the same preferences. No branding change.

### Dependencies

None.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
