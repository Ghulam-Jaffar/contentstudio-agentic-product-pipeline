# Prototype UI prompt — Reports tab (multi-workspace analytics report) — paste into Lovable / v0 / Claude

> Builds the UI for the new "Reports" tab on the All Workspaces page: the tab, the empty state, the reports table, and the "New report" dialog. It reuses ContentStudio's existing patterns: the analytics "Download Reports" table and the "Export Report" modal. Copy is final. Web only, light theme. (The generated PDF's internal layout, cover page plus the four aggregated widgets, is a separate design and is out of scope for this UI mockup.)

---

Build a high-fidelity, clickable web UI prototype of ContentStudio's "All Workspaces" page with a new "Reports" tab. Deliver it as a single self-contained React/JSX component, icons inline as SVG, no external deps, seeded mock data, working interactions.

VISUAL STYLE (match the existing ContentStudio app in the reference screenshots): light theme only, no dark mode, no RTL. Light-gray page background, white cards with subtle gray borders and rounded corners, blue primary #157FFF, gray-900 headings, gray-500/600 secondary text. Inter / system font. Green "Download" action buttons (soft green background, green text + download icon) exactly like the existing Download Reports table. Desktop-first.

## 1) All Workspaces page tab bar
A top tab bar with four tabs: "Manage Workspaces", "Manage Team & Roles", "Manage Limits", and a NEW fourth tab "Reports". A "Home" link sits on the far right. Stub the first three tabs (they already exist); open the prototype on the "Reports" tab (active).

## 2) Reports tab — EMPTY STATE (no reports yet)
A centered empty state inside the tab:
- A simple report/document icon.
- Heading: "Workspace analytics reports"
- Subtext: "Generate one report that combines the overview analytics of the workspaces you choose. Pick your workspaces, a date range, and a language, and we compile it into a single downloadable PDF."
- A primary button: "Download new report"
- A "Learn more" link below the button.

## 3) Reports tab — POPULATED STATE (has reports)
Reuse the look of the existing "Download Reports" table:
- A section header "Reports" on the left and a "Download new report" primary button on the right. A search box (like the existing table's Search).
- A table with columns: **Title | Workspaces | Period | Creation Date | Language | Status | Action**.
  - **Title**: the user-given report name (e.g. "Q2 performance across brands").
  - **Workspaces**: a row of small stacked workspace avatars/initials with a "+N" overflow bubble (like the "Accounts" column in the existing table). Use workspace initials/names, NOT social accounts.
  - **Period**: the date range, e.g. "07/01/24 - 06/30/26".
  - **Creation Date**: e.g. "07/13/26".
  - **Language**: e.g. "English", "French".
  - **Status**: for a generating report, show "Generating report... 25%" with a thin blue progress bar underneath (exactly like the existing table). For a ready report, show a green "Download" button with a download icon. For a failed report, show a red "Failed" label with a "Try again" link.
  - **Action**: a red trash/delete icon.
- Seed ~4 rows: one Generating (25%, blue bar, multiple workspace avatars + "+3"), two Ready (green Download), one Failed (Failed + Try again).

## 4) NEW REPORT dialog (opens from "Download new report")
Model it on the existing "Export Report" modal layout, adapted to this feature:
- Title: "New workspace analytics report", with an X to close.
- **Title** field (text input) — placeholder "For example, Q2 performance across brands"; helper text below: "A name to recognize this report in the list."
- **Report Language** field (dropdown with a small flag, default "English"), same style as the existing Export Report modal.
- **Date range** field (a date-range picker with presets like Last 30 days, This month, etc., matching the analytics date pickers).
- **Workspaces** field (a multi-select showing small workspace avatars/initials + names with a "+N" overflow, styled like the "Social Accounts" multi-select in the existing modal, but for WORKSPACES). Helper text: "Only workspaces you can access with analytics are shown."
- Bottom-left: a "Learn more" link. Bottom-right: a secondary "Cancel" button and a primary "Generate report" button.
- Validation: "Generate report" is disabled until a title is entered, a date range is set, and at least one workspace is selected. Inline messages when the user tries to submit early: "Enter a title", "Select a date range", "Select at least one workspace".
- On submit: the dialog closes, a new row appears at the top of the table in the "Generating report... 0%" state with a progress bar, and a toast appears: "Your report is being generated. It will appear in the list when it is ready."

Seed the workspace list (for the multi-select and the avatars) with realistic names: Ellis&Co, Bloomville, Casper, demo, Instagram, ISTG VIEW, Quality Team, Testing, Tiktok testing, Working nest (initials/avatars). Only these appear (they represent the workspaces the user can access with analytics).

## 5) States
- Loading: a skeleton/spinner for the table while it loads.
- Error: "We could not load your reports. Please try again." with a "Retry" action.

## 6) Prototype controls (for review)
- A toggle to switch the tab between the empty state and the populated table.
- A control to simulate a generating report advancing to "Ready" (so the reviewer can see the progress bar fill and turn into a green Download button).

DO NOT include: dark mode, RTL, a "Report Type" dropdown (this tab only produces the combined workspace report), social-account selection (it is workspaces here, not social accounts), or em dashes in any copy. Keep it desktop web, matching the existing screenshots' look.
