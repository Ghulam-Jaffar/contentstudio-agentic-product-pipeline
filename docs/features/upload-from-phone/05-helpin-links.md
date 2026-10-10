# Helpin Links: Upload from Phone (Scan QR)

Pushed 2026-09-28. Engineering team · sprint **Oct 05 - Oct 18 - 2026** · state **Ready** · priority **Medium** · unassigned.

Prototype: https://claude.ai/artifact/FbSZdddf9VP63dHSUsBJGJ

## Epic

[Upload from Phone (Scan QR)](https://app.helpin.ai/w/contentstudio/pm/epics/7284df75-f2c9-43fd-b09b-ef5135247c24) · `7284df75-f2c9-43fd-b09b-ef5135247c24`

## Stories

| Key | Story | Label | Docs attached |
|---|---|---|---|
| [CONT-4140](https://app.helpin.ai/w/contentstudio/pm/epics/7284df75-f2c9-43fd-b09b-ef5135247c24?task=CONT-4140) | [Design] Design the Upload from Phone modal tab, phone upload page and edge states | UIUX | PRD, Workflow |
| [CONT-4141](https://app.helpin.ai/w/contentstudio/pm/epics/7284df75-f2c9-43fd-b09b-ef5135247c24?task=CONT-4141) | [BE] Create upload-only phone sessions with QR link, expiry and live updates | backend | PRD, Research |
| [CONT-4142](https://app.helpin.ai/w/contentstudio/pm/epics/7284df75-f2c9-43fd-b09b-ef5135247c24?task=CONT-4142) | [BE] Accept phone uploads into the Content Library through the session link | backend | PRD, Research |
| [CONT-4143](https://app.helpin.ai/w/contentstudio/pm/epics/7284df75-f2c9-43fd-b09b-ef5135247c24?task=CONT-4143) | [BE] Convert HEIC and MOV, fix rotation and remove location data on uploaded media | backend | PRD, Research |
| [CONT-4144](https://app.helpin.ai/w/contentstudio/pm/epics/7284df75-f2c9-43fd-b09b-ef5135247c24?task=CONT-4144) | [FE] Add the From Phone tab and QR code to the upload modal | frontend | PRD, Workflow |
| [CONT-4145](https://app.helpin.ai/w/contentstudio/pm/epics/7284df75-f2c9-43fd-b09b-ef5135247c24?task=CONT-4145) | [FE] Show phone uploads live in the upload modal and use them where it was opened | frontend | PRD, Workflow |
| [CONT-4146](https://app.helpin.ai/w/contentstudio/pm/epics/7284df75-f2c9-43fd-b09b-ef5135247c24?task=CONT-4146) | [FE] Build the mobile upload page for Upload from Phone | frontend | PRD, Workflow |
| [CONT-4300](https://app.helpin.ai/w/contentstudio/pm/epics/7284df75-f2c9-43fd-b09b-ef5135247c24?task=CONT-4300) | [FE] Add an Uploaded from filter to the Content Library | frontend | PRD, Workflow |

## Docs (Engineering space, all attached to the epic)

| Doc | Collection | ID |
|---|---|---|
| Upload from Phone (Scan QR): Research | Research | `fb465e86-8a17-4da8-814f-7f0044e36564` |
| Upload from Phone (Scan QR): Workflow | PRDs & Feature Specs | `23f08c75-6d1c-4afa-b925-957169ea3865` |
| Upload from Phone (Scan QR): PRD | PRDs & Feature Specs | `832d7df6-07c1-4fb1-a986-b982167a5518` |

## Notes

- Returned labels matched the request on every story. Helpin added none on its own this time.
- The research and workflow docs had two stale lines and one local file link fixed before the push. The local copies match what's in Helpin.

## Update, 2026-10-05: Uploaded from filter

- Added **CONT-4300**, with the same fields as the batch: sprint Oct 05 - Oct 18, Ready, Medium, frontend, unassigned. The PRD and Workflow Docs are attached.
- **CONT-4142** now tags every phone upload `phone_upload` with its session ID, with no backfill. Its owner, Shaharyar Tariq, was kept.
- **CONT-4140** now covers the filter design. Its owner, Fasih Shaukat, was kept.
- In the epic description, the "Recent phone uploads" filter is no longer listed as out of scope. The planned start (2026-10-05) and deadline (2026-10-16), which someone set in Helpin, were left as they were.
- The PRD, Workflow and Research Docs were updated in place to match the local files.
