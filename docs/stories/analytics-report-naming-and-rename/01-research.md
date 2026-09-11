# Research: Report names for exported and emailed analytics reports, and renaming from the lists

**Date:** 2026-09-11
**Deliverable:** `02-stories.md`

This doc stays local. Codebase paths, entry points, gotchas and open questions live here and never in a story body.

---

## 1. The ask

Analytics has three ways to produce a report, reached from the same export control: **Export PDF**, **Send by email**, and **Schedule**. Only the Schedule modal lets the user name the report, and only the Scheduled Reports list shows a name.

1. Add the **report name** field to the **Export** modal and the **Send by email** modal, working the same way it already does in the Schedule modal.
2. Show the name in the **Download Reports** table, as the Scheduled Reports table already does.
3. Let users **rename** from the list rows on **both** pages: a downloaded report and a scheduled report.

The reasoning given for rename matters, because it shapes the behaviour: a downloaded report can be downloaded again, and a scheduled report will go out again, so the name a user picked once may need changing later. Rename is therefore not cosmetic tidying, it is about what the next download or the next send will be called.

---

## 2. Backlog dedupe

Checked `docs/stories/` and `docs/features/` before writing. Nothing covers report naming or renaming. Two deliverables are adjacent:

| Existing deliverable | Relationship |
|---|---|
| `docs/stories/analytics-report-email-copy-and-download/` | Rewrites the send-by-email report emails, both success and failure, so they address a third-party recipient, name the sender and platform, and offer a direct download. **Distinct scope**, but it touches the same modal and the same email. Worth coordinating: once a report has a user-given name, that email is the obvious place to use it, and that story is already rewriting the copy. If both are in flight, the email templates are the shared surface. |
| `docs/stories/analytics-report-date-time-preferences/` | Applies the user's saved date and clock formats to report contents. Same report pipeline, unrelated field. Relevant only in that both change what the report generator receives. |

---

## 3. The surfaces involved

Five, and they are not all the same component:

- **Export modal**, reached from the analytics export control. Needs a name field added.
- **Send by email modal**, same control. Needs a name field added.
- **Schedule modal**, same control. Already has the field. It is the reference implementation.
- **Download Reports** list page. Needs a name column and inline rename.
- **Scheduled Reports** list page. Shows the name already. Needs inline rename.

---

## 4. Frontend: the structural surprise

**The Schedule modal already is the Export and Email modal for most platforms.** It is tri-modal. A mode variable is set by whichever event opened it, giving `schedule`, `export` or `email`, and the export control routes by platform:

- For **every platform except Campaign and Label**, both "Export PDF" and "Send by email" open the **Schedule modal** in the corresponding mode.
- The standalone Export modal serves **Campaign and Label plus competitor** exports only.
- The standalone Email modal serves **Campaign and Label** only.

And the tri-modal currently **hides the name field explicitly in those two modes**:

```vue
<div v-if="!(isModeExport || isModeEmail)" class="flex flex-col gap-1">
```

So the single most useful finding: **for most platforms this feature is that condition being removed.** The field, the state, the payload key and the validation already exist and are simply suppressed in export and email mode.

The consequence for scoping is that the field has to be added in **two places per flow**, not one: the tri-modal's export and email branches, and the two standalone modals, which are independent implementations sharing almost nothing beyond the section picker. Building a small shared name-input component or composable is the obvious move, since it is needed in three or four places.

### The existing name field, which is the reference implementation

- **Label** "Export Name", **placeholder** "Enter a name for export..."
- State is a **Pinia** draft field, not a local ref, defaulting to `null`. There is no auto-generated default name.
- It is a **raw `input` element, not the design system's text input**. No max length, no character counter, no inline error slot, no trim.
- **Validation is required-only and post-submit**: the mutation checks for a name and, if missing, fires a toast reading "Please enter a report name." The submit button is not disabled for an empty name, only for invalid sections. There is no max length and no uniqueness check.
- Payload key is `name`, sent to the schedule save endpoint.

Both of those are worth improving while the field is being touched: moving to the design system input, adding a max length, and turning the post-submit toast into an inline error.

### A naming collision to resolve before build

The email flow **already computes a `file_name`**, a string like "Facebook Report (Slug) - {date}.pdf", and the tri-modal's own email path does the same. That is a **filename**, not a user-visible report name, and it is a different key from the `name` the schedule endpoint takes.

So the story must say explicitly whether the new field writes `name`, replaces `file_name`, or feeds both. Recommendation in section 7.

### An already-shipped precedent for exactly this shape

Settings has a cross-workspace analytics reports feature whose modal's **first field is a name** and whose table's **first column is that name**. It is the closest thing to the end state being asked for, and its copy is better than the existing "Export Name":

> **Label:** Report title
> **Placeholder:** A name to recognize this report in the list

Two differences to note: it uses a `title` key where analytics reports use `name`, and it has **no rename**. Its validation is better too, disabling the submit button on an empty name rather than toasting after the fact.

---

## 5. Frontend: the two list pages

**They are completely separate components with zero shared code.** Anything built for rename has to be implemented twice, or newly extracted. They share only page shell classes, a sticky-header table idiom, the skeleton box and an empty-state pattern, all copy-pasted.

### Scheduled Reports

One 517-line monolith holding page, table and row, with no composable.

- **Columns:** Report Name, Accounts, Schedule, Report Language, Creation Date, Actions.
- The name cell truncates at 30 characters with the full name in a tooltip. That is where inline rename goes.
- **Row actions:** Edit, which deep-clones the row into the Pinia draft and opens the tri-modal in schedule mode, and Delete, whose confirmation already interpolates the report name.
- No checkboxes, no bulk actions, no pagination.

### Download Reports

A page component plus a row component, three cell components and a composable.

- **Columns:** checkbox, Report Type, Accounts, Period, Creation Date, Language, Status, Action.
- **It shows no name.** The second column is a **derived type label** such as "Facebook" or "Grouped overview report (single PDF)", not anything the user set. A name column is genuinely new.
- Its view-model type has **no name field**, although the underlying API response type already declares an optional `name`. That is a strong hint the backend may already carry one.
- **Row actions:** expand/collapse, select, Delete, Retry on failed rows, and Download.
- The list **polls every ten seconds** with background refetching, and pages client-side at twenty rows.
- Search covers the derived type name and account names, not a report name, because there is none.

### A hazard specific to rename here

The existing delete mutations splice the cached list **by index**, and the composable deliberately resolves that index against the full list because **the list re-sorts on every ten-second poll**. A rename must therefore key on the report id and never on an index, or a poll landing mid-edit will rename the wrong row.

### Missing error states, pre-existing

**Neither table has an error state.** Query error flags are never read, so a failed fetch renders as the empty state. The repo's own quality rule calls for loading, error and empty states every time. This is pre-existing debt sitting directly under the feature: adding rename makes it worse, because a failed rename on a table that cannot show errors reads as nothing happening. Either fix it here or scope it out explicitly.

Neither empty state distinguishes "no reports" from "no search matches" either, since both filter client-side into the same branch.

---

## 6. Frontend: rename patterns already in the codebase

Three real precedents. The first is the one to follow.

### Best fit: Discovery folder rename, edit in place

The name span is swapped for a design system text input in the same row, entered from a dropdown item labelled "Rename" with a pencil icon.

```vue
<TextInput
  v-if="isRenaming"
  ref="renameInputRef"
  v-model="draftName"
  size="sm" radius="md" full-width
  :error="renameError"
  :input-props="{ autofocus: true, maxlength: 60 }"
  @keydown.enter.prevent="saveRename"
  @keydown.esc.prevent="cancelRename"
  @blur="cancelRename"
  @click.stop
/>
```

Everything about its behaviour is worth copying:

- Trim on save. **Empty or unchanged name cancels silently with no request.**
- Max length enforced on the input itself.
- **A duplicate-name error is shown inline on the row, deliberately not toasted.** Other failures do toast.
- Optimistic update with a cached rollback on error, then invalidate.
- Enter saves, Escape cancels, blur cancels.
- The error clears as the user types.
- Actions are provided through an inject-based action bag so deep rows do not prop-drill, which is directly transferable if the download report row is to stay dumb.

### Second: Listening topic type rename, shared inline form panel

Edits in place but as one shared form panel below the list rather than inside the row, with a single editing-id and a single input value serving both add and rename. Save is disabled while empty or pending, errors render under the input rather than as a toast, and uniqueness is checked client-side excluding self. Worth considering if an input inside a dense table row is unattractive.

### Third: Media library folder rename, small modal

A dedicated centred modal with a text input, a **visible character counter** against a max of 40, a disabled save while empty or in flight, an inline spinner in the save button, and an unchanged-name check that closes without a request. Reference this if rename should be a modal instead, which would be consistent with how the scheduled report Edit action already opens a modal.

---

## 7. Frontend: the API shape problem

- **There is no partial update endpoint for analytics reports.** The schedule save endpoint is a create-or-update upsert keyed on the report id, and the frontend always sends the **whole item**, including accounts, email list, widgets and frequency.
- Worse for a rename: the existing mutation **hard-requires** a name, at least one account, and either an email recipient or copy-to-self. So a naive "send only the id and the new name" would be rejected in the browser before it left.
- **Renaming a scheduled report is possible today** by cloning the cached row, overwriting the name and sending the whole object. It even gives the optimistic row swap for free. The downside is real though: a name-only change re-submits the whole schedule, so any server-side side effect of a schedule save, such as recomputing the cron expression or the next run time, fires for what the user thinks is a rename.
- **Renaming a downloaded report has no endpoint at all.** The only write endpoints for that list are remove and retry.

So a dedicated rename endpoint is the right call for both, rather than reusing the upsert. The closest partial-update precedent in the repo is Listening's topic type update, which takes an id and a small body, and is a good shape to propose.

Reusable pieces: the existing optimistic delete mutations for both lists, the shared query option builders and keys, and a generic optimistic handler helper in the discovery module.

---

## 8. Frontend: copy and translations

- Report strings live in the `analytics` namespace, one JSON file per locale, under `analytics.download_reports.*`, `analytics.my_reports.*`, `analytics.common.schedule_report_modal.*`, `analytics.common.schedule_reports_validation.*` and `analytics.common.send_report_by_email_modal.*`.
- There are **eight locales**, so every new string is eight file edits, and repo policy requires them all in the same commit.
- Existing copy worth reusing rather than reinventing: the scheduled report delete confirmation already interpolates the report name, and there is a generic "something went wrong" key both tables already fall back to.
- Test id convention is `analytics-*`, and the schedule modal's name input already has one. New rename affordances should follow the same convention, and there are existing row test ids for both tables to hang them off.

---

## 9. Backend: the good news is the field already exists

**`name` is already a field on both collections.** Nothing needs adding to the data model.

- On **`schedule_reports`** it is user-supplied and already works.
- On **`reports`** it exists and is written by several paths, but **not by the browser's one-off export or email flows**, which never send it. It is set to the account display name for per-account child reports, to the user's title for cross-workspace reports, from the request on the public API, and snapshotted from the schedule on the newer scheduled path.

So the backend half of "add the name" is largely a matter of the browser sending it and the list returning it.

### One collection, discriminated by source

All generated outputs live in **`reports`**. `schedule_reports` holds only the recurrence definition and never a generated file.

| Kind | `source` | In the Download Reports list |
|---|---|---|
| One-off PDF export | unset | **yes** |
| Emailed one-off | unset | **yes** |
| Public API generate | `export` | no |
| Scheduled run | `scheduled` | no |
| Cross-workspace | unset, own type | no, own list |

The list filter is `source === null`, which means **one-off exports and emailed reports are the same rows and cannot be told apart today**. If the two flows should be distinguishable in the list now that both can be named, that needs a new discriminator, and it is a product question rather than an implementation detail.

The list is also **hard-capped at thirty rows with no pagination**. Worth flagging: a name column invites browsing, and thirty rows is a small window.

### Things about the save endpoint that matter for a rename

- `POST analytics/reports/save` takes **`$request->all()` with no validation whatsoever**, and the repository writes **attribute by attribute, bypassing the model's fillable list**. So every key the browser posts lands on the document, and a `name` key would persist today with zero backend changes.
- It is already technically an **upsert**, because the repository honours an `_id` in the payload.
- But it is the wrong vehicle for a rename: **no validation, no permission middleware, and its `_id` lookup is not scoped by workspace**, which makes it a cross-workspace write primitive. It also re-dispatches generation depending on an `action` key, so a bare rename would have to be careful not to trigger a regeneration.

### The competitor projection will silently drop the name

The list method's competitor branch runs an aggregation with an explicit projection that **omits `name`**. So even once one-off reports carry a name, competitor rows will come back with no name unless the projection is extended. Easy to miss, and it would look like a bug in the new column.

---

## 10. Backend: the file name question, which is the real decision

**The user-supplied name does not currently feed the generated file name.** The file name is built as `"{workspace} - {Platform} {Type} - {date range} - {timestamp}.pdf"`. The report's `name` does already reach the PDF, but as the **cover title**, not the file name.

And the file name **is embedded in the storage object key**. The download URL stored on the report is a direct public URL to that object, so **the browser derives the saved filename from the URL path**. There is no content-disposition header on that path.

The consequence, which the story has to state plainly:

> **Renaming a report after it has been generated does not change the file that was already produced.** The PDF is an immutable stored object whose key was fixed at generation time.

Three ways to handle it:

1. **Metadata only.** The list shows the new name, the PDF cover shows it on any future generation, and future emails use it. The already-generated file keeps its old filename when downloaded. Cheapest, no storage work, breaks no existing links.
2. **Copy or move the stored object and rewrite the URL.** Makes the downloaded filename follow the name, but breaks any link already shared, and a retry would regenerate under a fresh timestamped name anyway.
3. **Add a download route that streams the file with a content-disposition built from the current name.** The only approach where a rename retroactively changes the downloaded filename, without touching stored objects or breaking links.

**This matters because of the reason given for the feature.** "Scheduled reports will go out again" is fully served by option 1, since the emails read the schedule live and pick up a new name immediately. But "users can download reports again" is only **partly** served by option 1: the row would show the new name while the downloaded file still carried the old one. If the downloaded filename is the point, option 3 is the one that delivers it.

Also worth knowing: a **retry regenerates** the file with a fresh timestamp and overwrites the stored URL. So under any option, a rename followed by a retry would pick up the new name provided the file-name builder is taught to use it.

---

## 11. Backend: rename endpoints

- **For a report there is no update endpoint of any kind.** The only write paths are remove and retry. A rename needs a new endpoint, scoped to the workspace, with the same permission middleware and the same `reports:manage` gate that remove and retry already carry.
- **For a schedule an update endpoint exists, but it is a full replace.** Every field is required, so renaming through it means the client resends the whole schedule or it silently clears accounts, email list, time, timezone and callback. Worse, the repository **recomputes the next run time** whenever frequency, day, time or timezone appear in the payload, so a full-replace rename would move when the next report goes out.

So both need a partial, name-only path. Two in-repo precedents to follow: the schedule state sub-resource, which is an existing example of a single-field update endpoint, and the share links update, which is the documented precedent for "a PUT that only renames must not blank the other fields".

### Permissions as they stand

The gate is `reports:manage`, resolving to a `manage_reports` permission, with an owner short-circuit:

| Role | Can rename someone else's report |
|---|---|
| super admin | yes |
| admin | yes |
| collaborator | **no** |
| approver | **no** |

A collaborator or approver **can** rename their **own** report, via the owner short-circuit. That seems right and needs no change.

Two gaps worth naming: **schedules have no permission gate at all today**, neither internally nor on the public API, so any workspace member can already rename any schedule. And the save and list endpoints are missing permission middleware entirely. A new rename route should carry the gate even though its neighbours do not.

**Demo workspaces:** a mutating route is denied on sample workspaces by default, which is correct. The note matters only so nobody "fixes" that as a bug by adding the route to the read-safe list.

---

## 12. Backend: scheduled runs and what a rename means for them

There are **two live scheduled paths** with different behaviour, and this is the subtlest part of the feature.

- **The newer path snapshots the name onto each run** and also stores the schedule id on it. So runs already produced keep the **old** name, and the list would show old and new names side by side. Because the schedule id is there, propagating a rename to past runs is trivially expressible if that is wanted.
- **The legacy path carries neither the name nor the schedule id.** Its runs cannot be named at all and cannot be found from the schedule.
- **The emails read the schedule live**, so a rename changes them immediately. And because the email job is dispatched with a **ten-minute delay**, a rename inside that window changes the email for a run that was generated under the old name.

Recommended position, to state in the story: **runs are immutable snapshots and a rename applies to future runs only.** That matches the newer path's existing behaviour and is the least surprising. Whether the legacy path should be taught to write the name and schedule id is a separate call.

### A free win nearby

The "report is ready" notification carries **no report name**, just a generic "Your report has been generated." Once reports have names, that is the obvious place to use one, and it costs almost nothing.

Similarly, the one-off and emailed report email subjects contain no report name today, while the scheduled ones do.

---

## 13. Validation, uniqueness and indexes

- The scheduled name is validated on the **public API only**, as required with a maximum of 255 characters. The internal endpoint the browser uses validates **nothing** about the name.
- **There is no uniqueness anywhere**, no unique rule and no check in either repository.
- **Neither collection has any migration or index at all.** If the feature wants uniqueness per workspace, or wants the lists to search by name, that needs a new index migration on both collections.
- There is no plan limit on the number of reports.

---

## 14. Recommendations

1. **Pre-fill the name field rather than requiring it.** Default it to the name the system would generate anyway, being workspace, platform, type and date range, so the field is never empty, a user who does not care is never blocked, and the list never shows a blank. This is better than copying the Schedule modal's required-and-empty behaviour, which currently fails with a post-submit toast.
2. **For rows that predate the feature, fall back to the derived type label** the Download Reports table already shows. No backfill, and no empty cells.
3. **Names need not be unique.** Two reports can reasonably share a name, and adding a uniqueness constraint would mean an index migration plus an error path for very little benefit.
4. **Rename is metadata-only for the list, the cover and future emails.** Then explicitly decide the downloaded-filename question above, because the stated motivation touches it.
5. **Add the name to the report-ready notification** while in there.
6. **Use the design system text input with a maximum length**, and convert the existing post-submit toast into an inline error, rather than replicating a raw input with no limit into three more places.
7. **Follow the discovery folder rename pattern** for both lists: edit in place, enter to save, escape to cancel, trim, silent no-op on unchanged, inline error, optimistic with rollback.
8. **Key the optimistic cache update on the report id, never an index**, because the Download Reports list re-sorts on a ten-second poll.

---

## 15. Open questions

1. **Should the downloaded file name follow a rename?** Metadata-only does not do that. See section 10. This is the one question that changes the shape of the backend work.
2. **Should one-off exports and emailed reports be distinguishable in the Download Reports list?** They are the same rows with the same marker today.
3. **Should renaming a schedule propagate to runs already produced?** Recommended no. The newer path makes it easy if wanted; the legacy path cannot do it at all.
4. **Should the legacy scheduled path be taught to write the name and schedule id**, so its runs are nameable?
5. **Should the two lists get a proper error state** as part of this, or is that explicitly out of scope? Neither has one today, and rename makes the gap more visible.
6. **Should the lists become searchable by name?** That would want an index.
7. **Is the thirty-row cap on the Download Reports list acceptable** once rows are named and worth browsing?
