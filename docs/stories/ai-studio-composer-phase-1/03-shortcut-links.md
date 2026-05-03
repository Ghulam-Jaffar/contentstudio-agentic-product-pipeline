# 03 — Shortcut Links: AI Studio Composer Phase 1

**Epic:** [AI Studio Composer: Presets & UI/UX Update](https://app.shortcut.com/contentstudio-team/epic/117353) (`sc-epic-117353`)
**Iteration:** [04 May - 15 May - 2026](https://app.shortcut.com/contentstudio-team/iteration/116060) (id `116060`, status `unstarted` at push time)
**Pushed:** 2026-05-03
**Group:** Frontend
**Project:** Web App
**Skill set:** Frontend
**Product area:** Composer
**Priority:** Medium (default — team can adjust at sprint planning)
**Workflow state:** Ready for Dev

All 10 stories use the New Feature Template, are linked to epic `117353`, and each has the 5 standard checklist tasks attached.

---

## Stories

| # | Title | Story | Tasks |
|---|---|---|---|
| 1 | [FE] Build collapsible AI Studio composer with Image/Video mode and redesigned model picker | [sc-117502](https://app.shortcut.com/contentstudio-team/story/117502) | 5/5 |
| 2 | [FE] Add reference slot system to AI Studio composer with contextualized Add Media modal | [sc-117508](https://app.shortcut.com/contentstudio-team/story/117508) | 5/5 |
| 3 | [FE] Add @mention attachment autocomplete to AI Studio composer | [sc-117514](https://app.shortcut.com/contentstudio-team/story/117514) | 5/5 |
| 4 | [FE] Add 16 preset workflows to AI Studio composer | [sc-117520](https://app.shortcut.com/contentstudio-team/story/117520) | 5/5 |
| 5 | [FE] Add multi-shot video generation to AI Studio composer | [sc-117526](https://app.shortcut.com/contentstudio-team/story/117526) | 5/5 |
| 6 | [FE] Restyle Brand selector as prominent chip in AI Studio composer | [sc-117532](https://app.shortcut.com/contentstudio-team/story/117532) | 5/5 |
| 7 | [FE] Apply three-tier chip styling system to AI Studio composer bottom row | [sc-117538](https://app.shortcut.com/contentstudio-team/story/117538) | 5/5 |
| 8 | [FE] Add Ratio, Duration, and Quality pills to AI Studio composer top control row | [sc-117544](https://app.shortcut.com/contentstudio-team/story/117544) | 5/5 |
| 9 | [FE] Add mobile-web responsive bottom sheets and compressed layout to AI Studio composer | [sc-117550](https://app.shortcut.com/contentstudio-team/story/117550) | 5/5 |
| 10 | [FE] Add validation banners, toasts, and confirm dialogs to AI Studio composer | [sc-117556](https://app.shortcut.com/contentstudio-team/story/117556) | 5/5 |

---

## Suggested implementation order

The stories aren't strictly dependent end-to-end, but for the smoothest sequence:

1. **sc-117502** — composer architecture + Mode + Model picker (foundation; blocks every other story)
2. **sc-117544** — top-row pills (Ratio, Duration, Quality) (needed before Multi-shot can place its switch)
3. **sc-117538** — bottom-row tier styling (needed before Brand chip restyle for shared tier-1 token)
4. **sc-117532** — Brand chip restyle (uses tier-1 token from sc-117538)
5. **sc-117508** — reference slots + Add Media modal contextualization
6. **sc-117520** — 16 presets (needs slots)
7. **sc-117526** — multi-shot (needs top-row from sc-117544)
8. **sc-117514** — @mention autocomplete
9. **sc-117556** — validation, toasts, confirm dialogs (needs the events the others emit)
10. **sc-117550** — mobile-web responsive layout (last — wraps everything in bottom sheets)

This is a recommendation, not a contract — the team may parallelize differently.
