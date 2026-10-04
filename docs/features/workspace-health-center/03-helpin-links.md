# Helpin links: Workspace Health Center

First pushed 2026-09-30 as 3 skeleton stories, then rewritten the same day from the CTO prototype. Narrowed on 2026-10-01 after the PO's discussion with the team: 9 stories kept (several renamed and rewritten), 3 moved to Abandoned. Engineering team, sprint **Oct 05 - Oct 18 - 2026**, state **Ready**, priority **medium**.

**Epic:** [Workspace Health Center](https://app.helpin.ai/w/contentstudio/pm/epics/0187c04c-42ad-4ccd-b1db-b7a5322ec2d7) (`0187c04c-42ad-4ccd-b1db-b7a5322ec2d7`), description updated 2026-10-01 and 2026-10-02.

**Design canvas:** https://claude.ai/artifact/VDtqPGHn2VUJm7dComQU26. Linked from the epic and every story since 2026-10-02 (replacing the CTO prototype link). The design story was renamed to "...from the design canvas" the same day.

| Key | Story | Owner | Labels |
|---|---|---|---|
| [CONT-4185](https://app.helpin.ai/w/contentstudio/pm/epics/0187c04c-42ad-4ccd-b1db-b7a5322ec2d7?task=CONT-4185) | [Design] Finalise the Health Center tabs and heart rail icon from the design canvas | Fasih Shaukat | design |
| [CONT-4186](https://app.helpin.ai/w/contentstudio/pm/epics/0187c04c-42ad-4ccd-b1db-b7a5322ec2d7?task=CONT-4186) | [BE] Track account health: reconnect status, consecutive failed posts, access expiry and error log | | backend |
| [CONT-4221](https://app.helpin.ai/w/contentstudio/pm/epics/0187c04c-42ad-4ccd-b1db-b7a5322ec2d7?task=CONT-4221) | [BE] Track per-account post failures and delivery-rate metrics for the Health Center | | backend |
| [CONT-4222](https://app.helpin.ai/w/contentstudio/pm/epics/0187c04c-42ad-4ccd-b1db-b7a5322ec2d7?task=CONT-4222) | [Research] Decide how ContentStudio should protect accounts automatically when platforms limit or break posting | Bilal Tariq | backend |
| [CONT-4187](https://app.helpin.ai/w/contentstudio/pm/epics/0187c04c-42ad-4ccd-b1db-b7a5322ec2d7?task=CONT-4187) | [FE] Build the Health Center Overview tab | | frontend |
| [CONT-4224](https://app.helpin.ai/w/contentstudio/pm/epics/0187c04c-42ad-4ccd-b1db-b7a5322ec2d7?task=CONT-4224) | [FE] Build the Health Center Accounts tab | | frontend |
| [CONT-4225](https://app.helpin.ai/w/contentstudio/pm/epics/0187c04c-42ad-4ccd-b1db-b7a5322ec2d7?task=CONT-4225) | [FE] Build the Health Center Post delivery tab with per-account error logs | | frontend |
| [CONT-4228](https://app.helpin.ai/w/contentstudio/pm/epics/0187c04c-42ad-4ccd-b1db-b7a5322ec2d7?task=CONT-4228) | [FE] Add the Health item to the rail with a heart that fills with workspace health | | frontend |
| [CONT-4229](https://app.helpin.ai/w/contentstudio/pm/epics/0187c04c-42ad-4ccd-b1db-b7a5322ec2d7?task=CONT-4229) | [Flutter] Bring the Health Center to the mobile app | | frontend (auto) |

**Abandoned (2026-10-01, out of scope):** archiving them is a manual UI job for the PO if wanted.

| Key | Story | Why |
|---|---|---|
| [CONT-4223](https://app.helpin.ai/w/contentstudio/pm/epics/0187c04c-42ad-4ccd-b1db-b7a5322ec2d7?task=CONT-4223) | [BE] Send health alerts by the workspace's alert settings | Notifications already exist |
| [CONT-4226](https://app.helpin.ai/w/contentstudio/pm/epics/0187c04c-42ad-4ccd-b1db-b7a5322ec2d7?task=CONT-4226) | [FE] Build Health Center alert settings | Alert settings dropped |
| [CONT-4227](https://app.helpin.ai/w/contentstudio/pm/epics/0187c04c-42ad-4ccd-b1db-b7a5322ec2d7?task=CONT-4227) | [FE] Add category tabs to the notification panel and a health banner on Home | Existing bottom reconnect banner covers it |

**Doc:** "Research: Workspace Health Center" (`8e485e51-a795-42f5-818d-d2964f92fa8d`), Research collection, attached to the epic. On 2026-10-02 the Design canvas, Team decisions and CTO prototype (original) sections were appended, so it now matches the local 01-research.md.

**PO decisions:** no Slack; heart thresholds as in the epic; automatic protection is a research story for Bilal Tariq (2026-09-30). Team narrowing, no alerts, no Home banner, no CSV export, one Flutter story (2026-10-01).
