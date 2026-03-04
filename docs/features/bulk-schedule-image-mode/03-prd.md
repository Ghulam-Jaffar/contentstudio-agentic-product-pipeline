# PRD: Bulk Schedule — Image Upload Mode

**Author:** Casper (Product Owner)
**Last Updated:** 2026-03-03
**Status:** Approved
**Target Release:** Q1 2026

---

## 1. Overview

Bulk Schedule today requires a CSV/XLSX spreadsheet — users must write captions, format dates, and structure a file before they can upload anything. For social media managers who already have their visual assets ready, this is unnecessary friction.

This feature adds a second mode to Bulk Schedule: **Schedule via Images**. Users drag and drop up to 100 images, set their accounts and schedule, and on the final step generate AI captions for all posts at once or one at a time. The entire workflow — from raw images to fully scheduled posts — happens inside the existing Bulk Schedule wizard without leaving ContentStudio.

The CSV/XLSX mode is unchanged. Both modes share the same wizard shell and reuse Steps 2 (Name & Accounts) and 3 (Scheduling). They diverge at Step 1 (CSV upload vs. image upload) and Step 4 (static review table vs. AI caption generation).

---

## 2. Problem Statement

**What problem are we solving?**

Users with visual content ready to post face a mandatory spreadsheet step before they can use Bulk Schedule. Writing captions, formatting CSV columns, and managing a local file is slow, error-prone, and disconnected from their actual workflow. Many users skip Bulk Schedule entirely because of this overhead.

**Who has this problem?**

- Social media managers with a batch of image assets ready for the week or a campaign
- Agency account managers who receive client images and need to get them posted quickly
- Content creators who shoot content in batches and want to schedule a week's worth of posts at once

**What happens if we don't solve it?**

- Bulk Schedule remains underused — users default to scheduling posts one at a time in the Composer
- Competitors offering drag-and-drop image scheduling with AI captions take users who value speed
- Automation module engagement stays low; users don't associate ContentStudio with "fast batch publishing"

---

## 3. Goals & Success Metrics

| Goal | Metric | Target | How We'll Measure |
|---|---|---|---|
| Increase Bulk Schedule usage | Weekly active Bulk Schedule sessions | +40% within 60 days of launch | Product analytics |
| Drive AI caption credit usage | AI text credits consumed via Bulk Schedule | 10% of total AI text credit usage | Credit deduction logs |
| Reduce time-to-scheduled for batch posts | Session duration: image upload → scheduled | Under 3 minutes for 10 posts | Session timing |

---

## 4. Target Users

**Primary Persona:**
Social media manager at a brand or agency — has a folder of images, needs captions fast, wants to schedule a week's worth of posts in one sitting. Uses ContentStudio daily, has an active plan with AI credits.

**Secondary Persona:**
Content creator or freelancer — shoots content in batches. Wants to upload all at once, let AI write the captions, and be done.

**Non-Users (out of scope):**
Free plan users without AI text credits (can still upload images and schedule without captions). Mobile app users (web-only for v1).

---

## 5. User Stories / Jobs to Be Done

| ID | As a... | I want to... | So that... | Priority |
|---|---|---|---|---|
| US-1 | Social media manager | upload a batch of images directly to Bulk Schedule | I don't have to write a spreadsheet first | Must Have |
| US-2 | User | generate AI captions for all my uploaded images at once | I can schedule a full week of posts in minutes | Must Have |
| US-3 | User | generate a caption for a single post without waiting for all | I can fix or try one image without restarting | Must Have |
| US-4 | User | select multiple posts and generate captions for just those | I can selectively fill gaps without running the full batch | Should Have |
| US-5 | User | see which posts still need captions at a glance | I know what's left before I schedule | Must Have |
| US-6 | User | remove individual images before finalizing | I can fix mistakes without restarting the upload | Must Have |
| US-7 | User | schedule posts without captions (image-only) | I'm not blocked if I prefer to caption manually later | Should Have |

---

## 6. Requirements

### 6.1 Must Have (P0)

- **Mode selector landing page:** Replace the existing single-CTA landing page with a two-card layout: "Schedule via CSV / XLSX" (existing) and "Schedule via Images" (new)
- **Image upload step (Step 1 — Image mode):** Drag-and-drop zone + file picker; accepts JPG, PNG, GIF, WebP; max 100 images; max 10 MB per file; thumbnail preview grid with individual remove; counter; Remove all
- **Wizard routing:** Image mode follows the same 4-step structure; Steps 2 and 3 are shared; Step 3 removes the "Use times from CSV" tab for image mode
- **Backend: image draft save:** Accept image files (JPG, PNG, GIF, WebP) in the draft save endpoint; validate file type and size (max 10 MB); upload to GCS; store file references in `csv_processing` document
- **Backend: process images into posts:** Each uploaded image becomes one post; posts passed into the same finalizing pipeline as CSV posts
- **Backend: AI caption generation endpoint:** New endpoint that accepts a post reference and returns an AI-generated caption for that image; consumed by all three frontend generation triggers (Generate All, single-post, bulk-selected); uses workspace Brand Settings for tone/style; deducts AI text credits per caption generated
- **Finalizing Posts — image mode:** Posts table shows thumbnail, caption state, and per-row status badge (Pending / Generating / Ready)
- **Generate All Captions button:** In Step 4 header (image mode only); triggers AI caption generation for all posts with no caption; disappears when all captions are generated
- **Single-post generate:** Inline "Generate caption" link in the post content area for posts with no caption
- **Generation loading state:** Per-post thumbnail shows pulsing `border-primary-cs-200` ring animation; content area shows animated dots + "Generating caption…"
- **Caption applied state:** Caption text appears in the post content area; Status badge updates to "Ready"; generate link disappears
- **Pending status badge:** Yellow badge on posts with no caption yet
- **Ready status badge:** Green badge on posts with a caption
- **Generating status badge:** Blue badge while AI is working on that post
- **Zero credits — toast:** If user triggers any generation action with 0 AI text credits, show a toast: "You've run out of AI text credits. [Upgrade your plan] to get more." Generation buttons stay enabled.
- **Generation failure — retry:** If AI generation fails for a post, show inline error in the content area: "Couldn't generate caption. [Try again]" — clicking "Try again" re-triggers single-post generation
- **Close-while-generating guard:** If × is clicked while generation is running, show a blocking confirmation modal (see §8.1)
- **Info banner:** Show a dismissible banner in Step 4 before any captions are generated; auto-hides once generation starts or any caption exists

### 6.2 Should Have (P1)

- **Generate selected (bulk):** Bulk Actions floating bar → Actions dropdown → "Generate Captions" — generates for selected posts that have no caption only; hidden if all selected posts already have captions
- **File > 10 MB — inline toast:** Show toast: "'{filename}' is too large. Max file size is 10 MB." Skip the oversized file and continue with valid ones
- **> 100 images — silent truncate:** Silently keep only the first 100; counter shows "100 / 100 images selected"
- **Row deletion mid-generation:** Deleting a post whose generation is in progress cancels that post's generation job; row removed cleanly from table
- **Schedule without captions:** Posts with no caption are allowed to proceed to scheduling as image-only posts — no blocking warning

### 6.3 Nice to Have (P2)

- **Parallel generation:** Process multiple posts simultaneously where AI pipeline allows; fallback to sequential if capacity is limited
- **Drag-to-reorder images:** Allow reordering in the Step 1 grid before proceeding
- **Duplicate detection:** Warn if the same image file is uploaded twice in one batch

### 6.4 Explicitly Out of Scope (v1)

- Mobile app (iOS/Android) support
- Inline caption text editor — clicking Edit (✏️) opens the existing full Composer
- AI prompt customization — captions generated from workspace Brand Settings automatically
- Caption regeneration — no regenerate button once caption exists; user edits manually
- Per-image scheduling times — image mode does not support "Use times from CSV"
- Video uploads in image mode — images only
- Progress percentage indicator — animated dots only

---

## 7. User Flows

### 7.1 Happy Path — Image Upload + Generate All + Schedule

1. User navigates to Publisher sidebar → Automations → Bulk Schedule
2. Landing page shows two cards: "Schedule via CSV / XLSX" and "Schedule via Images"
3. User clicks "Upload Images" on the Image card
4. Wizard launches inline — Step 1: Upload Images
5. User drags 15 JPG images onto the drop zone — thumbnail grid appears with counter "15 / 100 images selected"
6. User clicks "Next" → Step 2: Name & Accounts
7. User enters a campaign title and selects social accounts → clicks "Next"
8. Step 3: Scheduling — "Use times from CSV" tab not shown; user picks Regular Intervals → clicks "Next"
9. Warning banner: "Review your settings before proceeding…" — user confirms
10. Step 4: Finalizing Posts — 15 rows in table, all showing "Pending" (yellow badge); info banner visible; "Generate All Captions" button in header
11. User clicks "Generate All Captions"
12. Posts generate one at a time: each shows pulsing ring + blue badge while generating → caption appears + green badge when done
13. "Generate All Captions" button disappears once last caption is generated
14. User selects all rows → clicks "Schedule Selected" in bulk actions bar → posts scheduled

### 7.2 Single-Post Generation

1. User is on Step 4 with some posts still showing "Pending"
2. User clicks "Generate caption" inline under a specific post
3. That post's thumbnail shows pulsing ring; content area shows "Generating caption…"; badge turns blue
4. Caption appears; badge turns green; generate link disappears

### 7.3 Zero AI Text Credits

1. User clicks "Generate All Captions" or any generate trigger
2. Toast shown: "You've run out of AI text credits. [Upgrade your plan] to get more." — orange/error toast
3. No generation starts; page state unchanged

### 7.4 Partial Credits (Runs Out Mid-Generation)

1. "Generate All Captions" starts and processes several posts
2. Credits are exhausted before all posts are complete
3. Generation stops; completed posts show "Ready" (green); remaining posts stay "Pending" (yellow)
4. Toast: "Some captions couldn't be generated — you ran out of AI text credits. [Upgrade your plan] to get more."
5. User can generate individually for remaining posts once credits are topped up

### 7.5 Generation Failure (Single Post)

1. AI generation fails for one post (network error, AI service failure)
2. Post content area shows: "Couldn't generate caption. [Try again]"
3. Status badge reverts to "Pending" (yellow)
4. Clicking "Try again" re-triggers generation for that post only

### 7.6 Closing During Active Generation

1. User clicks × on the wizard header while "Generate All" is running
2. Blocking modal: "Caption generation in progress" (see §8.1)
3. User clicks "Stay & Wait" → modal closes, generation continues
4. OR user clicks "Leave Anyway" → generation cancelled, user returns to landing page

---

## 8. Functional Specification

### 8.1 Entry Points — Two States

The Bulk Schedule section has two distinct states that require different entry-point patterns.

---

#### 8.1a Empty State (No existing bulk schedules)

When a user visits Bulk Schedule for the first time or has no existing schedules, the page shows a full-page empty state with a two-card mode selector.

**Page heading:** "Get started with Bulk Scheduling"
**Sub-heading:** "Choose how you'd like to create and schedule your posts"

**CSV / XLSX card:**
- Icon: FileSpreadsheet (Lucide)
- Title: "Schedule via CSV / XLSX"
- Description: "Schedule up to 500 posts by uploading a spreadsheet with dates, captions, and media links."
- Primary CTA: "Upload CSV File"
- Secondary CTA: "Download Sample CSV"

**Schedule via Images card (new):**
- Icon: Image (Lucide)
- Title: "Schedule via Images"
- Badge: "NEW · AI-Powered" — `bg-primary-cs-500` gradient, white text
- Description: "Upload up to 100 images and let AI generate captions instantly — then schedule in bulk."
- Primary CTA: "Upload Images"
- No secondary CTA

**Below both cards:** Retain existing "Learn more about Bulk Scheduling" text link.

---

#### 8.1b Listing State (Has existing bulk schedules)

When the user has existing bulk schedules, the page shows the listing table (existing `CsvProcessListing` component). The current single "Upload CSV File" button in the top-right becomes a split/dropdown button labeled **"+ Bulk Schedule"** that opens a small dropdown menu with two options.

**Button label:** "+ Bulk Schedule"
**Dropdown options:**

| Option | Icon | Main text | Sub-text |
|---|---|---|---|
| Option 1 | FileSpreadsheet (Lucide) | "Schedule via CSV / XLSX" | "Schedule up to 500 posts from a spreadsheet" |
| Option 2 | Image (Lucide) | "Schedule via Images" | "Upload images and generate captions with AI — up to 100 posts" |

Option 2 has a "NEW" badge (small, `bg-primary-cs-500`, white text) next to its main text label.

Clicking either option launches the respective wizard inline, replacing the listing table view. Clicking × in the wizard returns to the listing table.

---

Both entry points launch the same wizard. The wizard renders as a centered card, max-width 720px (Steps 1–3) and 1000px (Step 4).

### 8.2 Step 1 — Upload Images

Wizard header: "Bulk Schedule via Images" + "AI-Powered" badge (same style as card badge)

| Element | Spec |
|---|---|
| Instruction line 1 | "Upload up to 100 images. Each image becomes a separate post. AI captions will be generated in the last step." |
| Instruction line 2 | "Supported: JPG, PNG, GIF, WebP · Max file size: 10 MB" — grey, smaller text |
| Drop zone | Dashed `border-primary-cs-200` border, centered image icon (Lucide Image); label: "Drag & drop images here, or Upload" — "Upload" is a `text-primary-cs-500` hyperlink-styled trigger |
| Drag-over state | Border turns `border-primary-cs-500`, background turns `bg-primary-cs-50` |
| File picker | `accept="image/jpeg,image/png,image/gif,image/webp"`, `multiple` |
| Image limit | Max 100. If user selects more, keep first 100 silently; counter shows "100 / 100 images selected" |
| File size limit | Max 10 MB per file. Oversized files skipped; toast: "'{filename}' is too large. Max file size is 10 MB." |
| Preview grid | Auto-fill grid, min cell size 72px, `object-fit: cover`. Max height ~200px with vertical scroll. Each thumbnail: `border-radius: 8px`, `1px border-gray-200`, × remove button top-right |
| Counter | "{N} / 100 images selected" — bold count, grey label. Shown above grid. |
| Remove all link | "Remove all" — `text-red-500`, top-right of grid; clears all images |
| Next button | Disabled (grey) until ≥1 image uploaded; enabled once images are present |
| × wizard button | Returns to landing page — no confirmation (unless generation is in progress on Step 4) |

### 8.3 Step 2 — Name & Accounts

Identical in both modes — no changes.

### 8.4 Step 3 — Scheduling

Shared with CSV mode with one difference: "Use times from CSV" tab is not available in image mode.

| Element | Image mode | CSV mode |
|---|---|---|
| Scheduling tabs | Regular Intervals / Content Category / Add to Queue | Use times from CSV / Regular Intervals / Content Category / Add to Queue |
| Default tab | Regular Intervals | Use times from CSV |

All other scheduling options (Regular Intervals, Posting Status, Content Category, Add to Queue) are identical.

**Warning banner (always shown at bottom of Step 3):**
> "Note: Review your settings before proceeding. You won't be able to go back to previous steps after this."
> Style: amber/yellow alert (`bg-yellow-50`, `border-yellow-200`)

### 8.5 Step 4 — Finalizing Posts

Wizard card expands to max-width 1000px at this step.

**Header row:**

| Element | Spec |
|---|---|
| Post count | "{N} post(s)" — bold number, `text-gray-500` label |
| Generate All button (image mode only) | "Generate All Captions" — `bg-primary-cs-500` gradient button. Tooltip: "Uses your AI text credits." Disappears once all captions are generated (pendingCount === 0) |
| Generate All — active state | Animated dots + "Generating captions…" — `bg-primary-cs-50` background, `border-primary-cs-200` border, non-clickable cursor |
| Generate All — zero credits | Clicking shows toast: "You've run out of AI text credits. [Upgrade your plan] to get more." (orange toast, 5s). No modal. Button stays enabled. |

**Info banner (image mode, Step 4 only):**
> "Your images are ready. Click 'Generate All Captions' to auto-generate captions for all posts at once, or generate individually per post."
> Style: `bg-primary-cs-50`, `border-primary-cs-200`, left-aligned icon
> Shown only before any captions exist; auto-dismissed when generation starts or any caption is populated.

**Posts table columns:**

| Column | Content |
|---|---|
| Checkbox | Row-level checkbox; header select-all checkbox |
| Post | Thumbnail (56×56px) + content area (see §8.6) |
| Accounts | `SocialChannelsTooltip` with account avatars |
| Status | Pending / Generating / Ready badge (see §8.7) |
| Label | Tag chips |
| Campaign | Campaign badge |
| Actions | Schedule link, Edit (pencil icon), Delete (trash icon) |

### 8.6 Post Column — Content States

**Thumbnail:**
- Image post: `object-fit: cover`, `border-radius: 8px`, `1px border-gray-200`
- Broken/failed image: generic placeholder icon (Lucide ImageOff, rotates by post index)
- Generating state: pulsing `border-primary-cs-200` ring animation around the thumbnail

**Content area:**

| State | Line 1 | Line 2 |
|---|---|---|
| Has caption | Caption text (max 2 lines, `text-ellipsis`) | — |
| No caption — image mode | "No caption yet." — `text-gray-400`, italic | "Generate caption" — dotted underline, `text-primary-cs-500`; hover: `text-primary-cs-700`. Tooltip: "Uses your AI text credits." |
| No caption — CSV mode | "No caption yet." — `text-gray-400`, italic | "Edit manually →" — `text-gray-400`, italic |
| Generating | — | Animated three-dot indicator + "Generating caption…" — shown at fixed 52px height to keep row height stable |
| Generation failed | — | "Couldn't generate caption." + "[Try again]" link — `text-primary-cs-500`, triggers single-post generation |

### 8.7 Status Badge States

| State | Copy | Background | Use |
|---|---|---|---|
| Pending | "Pending" | `bg-yellow-100 text-yellow-700` | Post has no caption |
| Generating | "Generating" | `bg-primary-cs-50 text-primary-cs-600` | AI is currently generating for this post |
| Ready | "Ready" | `bg-green-100 text-green-700` | Post has a caption and is ready to schedule |

CSV mode does not use these status badges — it uses existing draft/scheduled/error/warning states.

### 8.8 AI Caption Generation

**Trigger 1 — Generate All (header button):**
- Generates captions for ALL posts that do not yet have a caption
- Posts processed one at a time (sequential), each with ~500ms stagger
- Disabled while any generation is in progress
- Disappears when pendingCount === 0
- No success toast shown after completion — visual state change is sufficient

**Trigger 2 — Single post (inline link):**
- Visible only on posts with no caption and not currently generating
- Clicking triggers generation for that post only
- Can be triggered while "Generate All" is running (post queued behind current)

**Trigger 3 — Bulk generate selected (Actions menu):**
- Visible in Actions dropdown only when ≥1 selected post has no caption
- Hidden if all selected posts already have captions
- Triggers generation for selected pending posts only
- Label: "Generate Captions"; sub-label: "For {N} selected post(s) · uses AI credits"
- Tooltip: "Uses your AI text credits."

**Credit handling:**
- Zero credits: toast shown, no generation triggered
- Credits run out mid-generation: completed posts stay "Ready"; remaining posts revert to "Pending"; toast: "Some captions couldn't be generated — you ran out of AI text credits. [Upgrade your plan] to get more."

### 8.9 Bulk Selection & Actions Bar

Floating bar at bottom-center of screen when ≥1 row is selected. Disappears when selection cleared.

| Element | Spec |
|---|---|
| Selected count | "{N} Selected" — white text on `bg-gray-800` |
| Unselect all | "( Unselect all )" — `text-primary-cs-300` link, clears selection |
| Actions button | "Actions ▾" — `bg-gray-700`, opens dropdown above the bar |

**Actions dropdown:**

| Action | Visibility | Spec |
|---|---|---|
| Generate Captions | Image mode only; only when ≥1 selected post has no caption | Sub-label: "For {N} selected post(s) · uses AI credits"; tooltip: "Uses your AI text credits." |
| Schedule Selected | Both modes | Sub-label: "Schedule {N} post(s)" |
| Delete Selected | Both modes | Red text; sub-label: "Remove {N} post(s) from this batch"; clears selection after |

### 8.10 Close-While-Generating Guard Modal

Shown when × is clicked and caption generation is actively running.

| Element | Copy |
|---|---|
| Modal title | "Caption generation in progress" |
| Body line 1 | "AI is still generating captions for your posts. If you leave now, generation will stop and any captions not yet completed will be lost." |
| Body line 2 (bold) | "AI credits already used for completed captions will not be refunded." |
| Primary CTA | "Stay & Wait" — primary button, dismisses modal, generation continues |
| Destructive CTA | "Leave Anyway" — white button, `text-red-600 border-red-300`, cancels generation and returns to landing page |

### 8.11 Navigation & Wizard Controls

- Wizard × button: returns to landing page. No tooltip needed.
- "Previous" button: navigates to prior step. Not available after Step 3 → Step 4 transition (processing begins on Step 3 → Next).
- Modal close (×, Cancel, click-outside): no generation triggered.

---

## 9. Credit System

| Credit type | When consumed | Rate |
|---|---|---|
| AI text credits | Per caption generated | Per generation call; exact cost defined by AI team |

- Credits shown only in tooltips ("Uses your AI text credits.") — not as a visible counter in the wizard
- No credit count displayed in the UI for this feature
- Zero credits → toast on attempt, no blocking
- Partial credits → generates until exhausted; toast on failure

---

## 10. Toast Notifications — Full Reference

| Trigger | Message | Style | Duration |
|---|---|---|---|
| File > 10 MB | "'{filename}' is too large. Max file size is 10 MB." | warn | 5s |
| Zero credits — any generation trigger | "You've run out of AI text credits. [Upgrade your plan] to get more." | error | 5s |
| Credits run out mid-generation | "Some captions couldn't be generated — you ran out of AI text credits. [Upgrade your plan] to get more." | error | 7s |

No success toast for caption generation — the UI state change (badge + caption appearing) is sufficient feedback.

---

## 11. Business Rules & Constraints

| Rule ID | Rule | Rationale |
|---|---|---|
| BR-1 | Max 100 images per batch | Server performance and UX manageability |
| BR-2 | Max 10 MB per image file | GCS upload limits and page performance |
| BR-3 | Accepted image formats: JPG, PNG, GIF, WebP | ContentStudio's supported image types |
| BR-4 | Posts with no caption can proceed to scheduling as image-only posts | Users shouldn't be blocked if they prefer manual captioning |
| BR-5 | Generation is sequential (one at a time) for v1 | Reduces AI pipeline load; parallel as P2 upgrade |
| BR-6 | AI features are web-only | Aligned with ContentStudio's AI product strategy |
| BR-7 | Zero credits → toast, not disabled buttons | ContentStudio shows toasts for credit exhaustion, not button-level blocking |
| BR-8 | Generation failure → inline retry | Users should be able to retry without restarting the whole flow |
| BR-9 | "Use times from CSV" tab hidden in image mode | Image mode has no CSV file to extract times from |
| BR-10 | Captions generated from workspace Brand Settings | No per-wizard prompt customization in v1 |

---

## 12. Open Questions

| Question | Options | Owner | Decision |
|---|---|---|---|
| Exact AI credit cost per caption call? | Defined per word / per call | AI team | To decide |
| Can generation run in parallel in v1? | Sequential (safe) / Parallel (faster) | Engineering | Sequential for v1, parallel as P2 |
| Delete row mid-generation — cancel the job? | Yes (cancel) / No (let complete, discard result) | Engineering | Leaning yes (cancel) |
| Mobile app: does Bulk Schedule exist on mobile? | Yes / No | Product | To confirm — mobile out of scope for v1 |
| What Brand Settings fields does AI caption use? | Tone, industry, keywords, brand voice | AI team | To confirm |

---

## 13. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| AI generates poor quality captions for niche industries | Medium | Medium | User can edit manually or regenerate individually; Brand Settings used automatically |
| Users upload 100 large images — slow page performance | Low | Medium | 10 MB file limit + preview grid max-height + lazy load thumbnails |
| Credits exhausted mid-generation confuses users | Medium | Low | Toast explains what happened + upgrade link; remaining posts clearly still "Pending" |
| Generation race condition (Generate All + single post click simultaneously) | Low | Low | Single-post clicks queue behind current generation; sequential processing prevents conflicts |

---

## 14. Dependencies

- **Existing:** `BulkUploadAutomationSave.vue` — wizard shell, step indicator, steps 2–3 components
- **Existing:** `AutomationScheduleOptions` — step 3, reused with tab filtering
- **Existing:** `TraitBulkImageUploader` (backend) — GCS image upload, already handles media library storage
- **Existing:** `CSVController` / `CsvProcessing` model — extended, not replaced
- **Existing:** `DataTable` — reused for Step 4 posts table
- **Existing:** AI caption generation service — same service used in Composer
- **Existing:** AI credit system (`getCreditUsedLimit` / `getCreditSubscribeLimit`)
- **New:** API endpoint for image-based draft save + processing pipeline

---

## 15. Appendix

- **User-provided spec:** Bulk Schedule Image Upload Mode v1.0 (March 2026) — attached to Shortcut epic
- **Research doc:** `docs/features/bulk-schedule-image-mode/01-research.md`
- **Existing wizard:** `contentstudio-frontend/src/modules/automation/components/csv/BulkUploadAutomationSave.vue`
- **Backend controller:** `contentstudio-backend/app/Http/Controllers/Automation/CSVController.php`

---

## Changelog

| Date | Author | Changes |
|---|---|---|
| 2026-03-03 | Product (Claude) | Initial PRD from user spec + codebase research |
| 2026-03-03 | Product | Fixed tooltip copy (removed "from your workspace"), zero credits → toast not disable, resolved all hardcoded colors, defined AI failure retry, finalized close-during-generation modal copy |
| 2026-03-03 | Casper | Added listing-state entry point (dropdown button); clarified BE story covers caption generation endpoint; confirmed image wizard Step 3 has no "Use times from CSV" tab |
