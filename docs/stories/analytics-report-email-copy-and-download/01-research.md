# Research: "Send PDF" analytics report email copy and download CTA

## Flow

1. Analytics, Export, "Send PDF" opens the send-by-email modal, `contentstudio-frontend/src/modules/analytics/components/reports/modals/SendReportByEmailModal.vue` (opened from `views/common/ExportButton.vue` and `reports/DownloadPdfButton.vue`). The sender enters recipient email addresses, picks a language, and sends. The recipient is entered as a raw email, so we do not have the recipient's name.
2. Backend generates the PDF and sends the email via `App\Jobs\Analyze\ExportReportJob` (helper `App\Libraries\Analytics\ReportsHelper`), using the blades:
   - `contentstudio-backend/resources/views/emails/notifications/analytics/single_reports_email.blade.php`
   - `contentstudio-backend/resources/views/emails/notifications/analytics/multiple_reports_email.blade.php`
3. Copy comes from localized strings via `\App\Helpers\LocalizationHelper::emailResponses('analytics_reports.*')`: `generated_body`, `generated_cta`, `view_reports_button`. Because the modal has a language selector, the email is localized, so any copy change must be made across all supported locales.

## What is wrong today

The email reads as if the recipient generated the report, and the CTA points to the reports section rather than downloading the PDF:

```
Hi Ghulam,

Your report has been generated successfully in Casper.

You can download it from the reports section.

[ View reports ]

Best regards,
The ContentStudio Team
```

Problems:
- Greeting uses a name, but we only captured the recipient's email, so there is no reliable name.
- Body says "Your report has been generated", implying the recipient made it. The recipient is a third party who was sent the report by a ContentStudio user.
- No mention of who sent it, or what report it is.
- CTA is "View reports" and links to the app reports section. The recipient may not be a ContentStudio user, so that is useless to them. It should download the actual report PDF.

## Data available for the new copy

- **Sender name:** the ContentStudio user who sent the report (available in the job/workspace context).
- **Report platform / type:** the report being exported has a platform/type (Facebook, Instagram, LinkedIn, X, Pinterest, YouTube, TikTok, GBP, Threads, Bluesky, or a non social type like Overview, Competitor, Campaign & Label, Workspace). Used to say "Facebook analytics report" vs a generic "analytics report".
- **PDF URL:** the generated report file URL is already passed to the blade (`$urls`), so the download CTA can link straight to the PDF.

## Decision

One `[BE]` ticket: rewrite the send-by-email report copy (localized) so it is addressed to a third-party recipient, names the sender, names the platform dynamically, and changes the CTA to a direct "Download report" that downloads the PDF. White-label aware for the brand/team signature.

## Failure email (report could not be generated)

The recipient can also receive a failure email when generation fails. It has the same defects plus worse ones: it greets by a name we do not have ("Hi Quality"), says "We were unable to generate your report in {workspace}", and offers "Try Again" and "contact support", none of which a third-party recipient can act on.

Findings while tracing this:
- `App\Jobs\Analyze\ExportReportJob::handleFailure` only writes an in-app notification and a Pusher event to the report owner. It does not send an email. So the failure email in the screenshot is rendered from a different path (a Notification class, a scheduled-report path such as `EmailReportJob`, or another template). The exact source must be located during build.
- The `analytics_reports` email string block (`lang/en/emails.php` around lines 789 to 813) has success and instant keys but no failure keys. New failure keys must be added there across all 8 locales.

So the failure email needs the same recipient-facing rewrite: no name greeting, name the sender and platform, say it could not be generated, and tell the recipient to contact the sender to resend (no Try Again, no contact support for the recipient). Tracked as a second `[BE]` story in this doc.
