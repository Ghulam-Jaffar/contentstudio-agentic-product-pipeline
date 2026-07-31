# Stories: Fix the "Send PDF" analytics report emails (success + failure)

Two related `[BE]` tickets on the same flow: the success email (report delivered) and the failure email (report could not be generated). Both must be addressed to a third-party recipient, not to the person who generated the report.

---

## [BE] Rewrite the send-by-email analytics report so it addresses the recipient, names the sender and platform, and downloads the PDF

### Description

When a user exports an analytics report with Analytics, Export, Send PDF, they enter a recipient email and the recipient receives an email. Today that email is wrong: it is written as if the recipient generated the report, it greets them by a name we do not have, and its CTA sends them to the app reports section instead of downloading the report.

Fix the email so it reads correctly for a third-party recipient:
- Do not greet by name. We only captured the recipient's email address, not their name.
- Say who sent it and what it is: the sending user's name, and the report platform when it is a specific social platform, otherwise a generic "analytics report".
- Change the CTA to "Download report" and make it download the report PDF directly (the recipient may not be a ContentStudio user, so it must not depend on being logged in).

The copy is localized (the send modal has a language selector), so update the strings in every supported locale. Keep the signature white-label aware.

### Current copy (for reference, this is what is wrong)

```
Hi Ghulam,

Your report has been generated successfully in Casper.

You can download it from the reports section.

[ View reports ]

Best regards,
The ContentStudio Team
```

### New copy

**Subject**
- Social platform report: `{Platform} analytics report from {Sender}`
  - e.g. `Facebook analytics report from Ghulam Jaffar`
- Non social report (Overview, Competitor, Campaign & Label, Workspace, etc.): `Analytics report from {Sender}`

**Body**

```
Hi,

{Sender} sent you {a|an} {Platform} analytics report.

You can download the full report as a PDF below.

[ Download report ]

Best regards,
The {Brand} Team
```

Rules for the variables:
- **Greeting:** always the plain "Hi," with no name.
- **{Sender}:** full name of the ContentStudio user who sent the report.
- **Platform sentence, social report:** `{Sender} sent you a Facebook analytics report.` The article is dynamic: use "an" before a vowel sound (for example "an Instagram analytics report", "an X analytics report"), otherwise "a". If localizing makes the inline article awkward, the acceptable equivalent phrasing is `{Sender} sent you an analytics report for {Platform}.` which avoids the article problem entirely and is preferred for non-English locales.
- **Platform sentence, non social report:** drop the platform, so `{Sender} sent you an analytics report.`
- **{Brand} Team:** white-label aware. On the default app this is "The ContentStudio Team"; on a white-label domain use the white-label brand name. Do not hardcode "ContentStudio", and do not include the workspace name (remove the old "in Casper").
- **Multiple reports** (the `multiple_reports_email` template): use the plural, `{Sender} sent you analytics reports.` (or list them), and keep a single "Download report(s)" CTA.

**CTA**
- Label: `Download report` (was "View reports").
- Action: links directly to the report PDF (`$urls` is already passed to the blade). Downloads or opens the PDF without requiring the recipient to log in.

### Acceptance criteria

- [ ] The recipient email greeting is "Hi," with no name.
- [ ] The body names the sender and, for a social platform report, the platform: "{Sender} sent you a {Platform} analytics report." For non social reports (Overview, Competitor, Campaign & Label, Workspace, and similar) it reads "{Sender} sent you an analytics report." with no platform.
- [ ] The article before the platform is grammatically correct (a vs an), or the "an analytics report for {Platform}" phrasing is used.
- [ ] The old "Your report has been generated successfully in {Workspace}" wording is gone. The workspace name is not shown.
- [ ] The CTA reads "Download report" and downloads or opens the report PDF directly. It works for a recipient who is not logged in and is not a ContentStudio user.
- [ ] The signature is white-label aware (default brand or white-label brand), not hardcoded to ContentStudio.
- [ ] All of the above are applied to both the single and multiple report templates.
- [ ] The new strings are added and translated in every supported locale (the send modal language selector controls which is used).
- [ ] No regression to the scheduled report emails if they share these strings. If they must keep the old wording, use separate strings so only the send-by-email flow changes (confirm during build).

### Implementation references

*Pointers from research, not a contract.*

- Templates: `contentstudio-backend/resources/views/emails/notifications/analytics/single_reports_email.blade.php` and `multiple_reports_email.blade.php`.
- Send flow: `App\Jobs\Analyze\ExportReportJob` and `App\Libraries\Analytics\ReportsHelper` (choose template, build payload, pass `$urls`). Pass the sender name and the report platform/type into the payload.
- Copy source: `\App\Helpers\LocalizationHelper::emailResponses('analytics_reports.*')` (`generated_body`, `generated_cta`, `view_reports_button`). Add or update the strings here across locales. Prefer new keys for the send-by-email wording so scheduled report emails are not disturbed.
- Frontend context (no change expected): `contentstudio-frontend/src/modules/analytics/components/reports/modals/SendReportByEmailModal.vue` collects recipient emails and the language; `views/common/ExportButton.vue` and `reports/DownloadPdfButton.vue` open it.

### Open questions for the discussion

- Is the PDF delivered as a **direct download link** (signed URL) or **attached** to the email? The CTA must work for a non-CS recipient either way; confirm which.
- Do we want the report period or date range shown in the body (for example "for 1 to 30 June")? Not required, but easy to add if wanted.

---

## [BE] Fix the report-generation failure email so it is addressed to the recipient and points them to the sender

### Description

When a report sent through "Send PDF" fails to generate, the recipient can receive a failure email that reads as if they generated it. It greets them by a name we do not have, says their report could not be generated in a workspace they may not belong to, and offers "Try Again" and "contact support". A third-party recipient cannot retry or contact support about someone else's report.

Rewrite the failure email so it tells the recipient that a sender tried to send them a report that could not be generated, and to contact the sender to resend. Remove the "Try Again" and "contact support" actions for the recipient. Localized (all supported locales) and white-label aware. This mirrors the success-email fix above.

Note during scoping: the export job's failure handler (`ExportReportJob::handleFailure`) currently only creates an in-app and Pusher notification, not an email, and the `analytics_reports` email strings have no failure keys. So part of this ticket is locating where the failure email is actually sent and who receives it (see open questions), then applying the copy below.

### Current copy (for reference, this is what is wrong)

```
Subject: Your report could not be generated - BloomVille

Hi Quality,

We were unable to generate your report in BloomVille.

Please try generating it again. If the problem persists, contact support.

[ Try Again ]

Best regards,
The ContentStudio Team
```

### New copy

**Subject**
- `There was a problem with the analytics report {Sender} sent you`
  - or, if you prefer to keep the platform: `{Platform} analytics report from {Sender} could not be generated`

**Body**

```
Hi,

{Sender} tried to send you a {Platform} analytics report, but it could not be generated.

Please contact {Sender} to resend it.

Best regards,
The {Brand} Team
```

Rules for the variables (same as the success email):
- **Greeting:** plain "Hi," with no name. We only have the recipient's email.
- **{Sender}:** full name of the ContentStudio user who tried to send the report.
- **{Platform}, social report:** name the platform, so "a Facebook analytics report". Handle the article (a vs an), or use "an analytics report for {Platform}".
- **Non social report:** drop the platform, so "{Sender} tried to send you an analytics report, but it could not be generated."
- **{Brand} Team:** white-label aware. Do not hardcode ContentStudio. Do not include the workspace name.
- **No retry or support action for the recipient.** Remove the "Try Again" button and the "contact support" line. The only action is to contact the sender.
- Optional: if the sender's email is available, a "Contact {Sender}" button can `mailto:` the sender. Confirm during build; otherwise text only.

### Acceptance criteria

- [ ] The failure email greeting is "Hi," with no name.
- [ ] The body names the sender and, for a social platform report, the platform: "{Sender} tried to send you a {Platform} analytics report, but it could not be generated." For non social reports it omits the platform.
- [ ] The email tells the recipient to contact the sender to resend. It no longer says "try generating it again" or "contact support".
- [ ] The "Try Again" button is removed (or replaced by an optional "Contact {Sender}" mailto if the sender email is available and product agrees).
- [ ] The workspace name is not shown, and the signature is white-label aware.
- [ ] The article before the platform is correct (a vs an), or the "an analytics report for {Platform}" phrasing is used.
- [ ] New failure strings are added and translated in all supported locales.
- [ ] If the report owner (sender) also receives a failure notification, their version may keep retry-oriented wording; only the recipient version uses the "contact the sender" copy.

### Implementation references

*Pointers from research, not a contract.*

- Copy source: `\App\Helpers\LocalizationHelper::emailResponses('analytics_reports.*')` in `contentstudio-backend/lang/en/emails.php` (the `analytics_reports` block, around lines 789 to 813). It has no failure keys today; add them here and mirror to all locales.
- Failure path in the export flow: `App\Jobs\Analyze\ExportReportJob::handleFailure` (currently in-app and Pusher only). Confirm where the recipient failure email originates (a Notification class, `App\Jobs\Analytics\EmailReportJob`, or another template) and apply the new copy there.
- Reuse the sender and platform values from the same payload used by the success email.

### Open questions for the discussion

- **Where is this failure email sent, and to whom?** The export failure handler does not send an email, so confirm the exact template and whether it currently goes to the send-by-email recipients, the report owner, or both.
- Should the recipient receive a failure email at all, or only the sender? The request is that the recipient is told to contact the sender, so this assumes the recipient does receive it.
