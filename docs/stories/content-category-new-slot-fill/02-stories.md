# Content Categories — Smarter Slot Filling · Epic & Stories

**Platform:** Web. **Scope:** 1 × `[BE]` (scheduling fix) + 1 × `[FE]` (modal note).

## Epic: Content Categories — Smarter Slot Filling

When a user adds a new posting slot to a content category — or moves an existing slot to an earlier time/day — that slot should start getting used right away. Today, content category posts are always scheduled after the *last occupied slot*, so a user with ~2 months of queued content who adds a Tuesday slot sees it sit empty for months while new posts pile up at the end of the queue. This epic makes newly-added and re-timed slots behave like freed-up slots already do in the **post-deletion** flow: their nearest upcoming, unfilled occurrences are filled first, then scheduling continues forward. It also adds an in-modal heads-up so users understand the behavior when they add/edit a slot in an already-scheduled category.

**Status:** Locked in (approved).

| # | Story | Priority |
|---|---|---|
| S-1 | [BE] Fill newly-added or re-timed content category slots before scheduling after existing posts | High |
| S-2 | [FE] Show a "new posts fill this slot first" note when adding or editing a slot in a category that already has a schedule | Medium |

---

## S-1 · [BE] Fill newly-added or re-timed content category slots before scheduling after existing posts
**Project:** Web App · **Group:** Backend · **Skill:** Backend · **Product area:** Publishing · **Priority:** High · **Type:** Feature

### Description
As a user who adds a new posting slot — or moves an existing slot to an earlier time/day — in a content category that's already scheduled out, I want my next posts to drop into that slot's upcoming dates first — instead of being pushed to the very end of my months-long queue — so that the slot actually gets used right away.

Today, new content category posts are scheduled after the last occupied slot, so a newly-added (or newly re-timed) slot stays empty until the whole existing queue is exhausted. This should behave like the existing post-deletion flow, where a freed-up slot earlier in the timeline is filled before scheduling continues forward.

### Workflow
1. A user has a content category with content scheduled out for, say, the next 2 months.
2. The user adds a new slot (e.g., Tuesday 9:00 AM).
3. The user adds new posts to that category.
4. The new posts fill the new slot's **next upcoming, unfilled dates first** (e.g., next Tuesday), then continue scheduling forward after the existing queue once the new slot's nearer dates are taken.

### Acceptance criteria
- [ ] When a new slot is added to a content category that already has scheduled posts, newly added posts fill that slot's earliest upcoming unfilled occurrences **before** being scheduled after the last occupied slot.
- [ ] The same fill-first behavior applies when an **existing slot is edited to an earlier time/day** — its nearer, unfilled occurrences are filled first. (Editing a slot today only updates the slot's time/category and does **not** reschedule existing posts, so this governs where the next new posts land.)
- [ ] This reuses/extends the existing "fill the gap first" behavior used when a post is deleted (the removed-slot gap-fill), so newly-added slots, re-timed slots, and freed slots are all treated as fillable gaps.
- [ ] Scheduling never places a post in the **past** — only upcoming occurrences of the new slot are used.
- [ ] Once the new slot's nearer dates are filled, scheduling continues correctly after the existing queue (no duplicate assignment to the same slot/time, no skipped dates).
- [ ] Existing scheduled posts are **not** moved or rescheduled by adding a new slot — only newly added posts use the new slot.
- [ ] Behavior is correct when multiple new slots are added (all are treated as fillable upcoming gaps).
- [ ] Works for both workspace-level and global (shared) content categories; for global categories, the fill-first behavior applies per workspace using the category without unintended cross-workspace effects.
- [ ] Slot times are evaluated in the workspace timezone (consistent with current slot scheduling).
- [ ] No regression to the existing post-deletion gap-fill behavior.

### Mock-ups
N/A — backend scheduling behavior.

### Impact on existing data
No schema change expected. Affects how the next slot/time is computed for new content-category posts; existing scheduled posts are untouched.

### Impact on other products
Content category auto-scheduling (publishing). The FE Add Slot note (paired story) communicates the behavior. Mobile/Chrome unaffected (scheduling is server-side).

### Dependencies
- Paired with **[FE] Show a "new posts fill this slot first" note when adding or editing a slot in a category that already has a schedule**.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*
- `contentstudio-backend/app/Libraries/Publish/ContentCategorySlotsHelper.php` — computes the next slot/time. `getNextSlot()` → `pickTimeFromSlot()`; the constructor anchors on `PlansRepository::getLastOccupiedContentCategorySlot(...)`, which is why new posts land at the end. `hasRemovedSlots()` already fills earlier *removed* slots first — extend the same "earlier unfilled gap" logic to **newly-added** slots.
- `contentstudio-backend/app/Repository/Settings/ContentCategoriesSlotsRepository.php` — slot records; identify when a slot was newly added vs. already filled.
- `contentstudio-backend/app/Libraries/Publish/QueueSlotsHelper.php` — related queue/slot helper.
- Mirror the deletion flow: the existing removed-slot handling is the proven pattern for "fill the empty earlier slot before moving forward."

---

## S-2 · [FE] Show a "new posts fill this slot first" note when adding or editing a slot in a category that already has a schedule
**Project:** Web App · **Group:** Frontend · **Skill:** Frontend · **Product area:** Publishing · **Priority:** Medium · **Type:** Feature

### Description
As a user adding a new slot to a content category that's already scheduled out, I want a heads-up in the Add Slot modal that new posts will fill this new slot first, so that I understand why my next posts will appear in the new slot's upcoming dates rather than at the end of my queue.

### Workflow
1. The user opens the Add/Edit Slot modal for a content category — either adding a new slot or changing an existing slot's time/day.
2. If that category **already has content scheduled** (queued posts), an info note appears at the bottom of the modal.
3. The note explains that new posts will fill this slot first, then continue after existing posts.
4. The user picks the time/days and clicks **Create** (or saves the edit).

### Acceptance criteria
- [ ] When the category **already has scheduled (queued) content**, an informational note shows at the bottom of the modal — both when **adding a new slot** and when **editing an existing slot's time/day** — using the `Alert` component (matching how other modals show notes, e.g. Add Category).
- [ ] Note copy: **"New posts will fill this slot first. Since this category already has content scheduled, your next posts will go into this slot's upcoming dates before continuing after your existing posts."**
- [ ] The note does **not** show when the category has **no scheduled content** — even if it already has slots configured — because there's nothing to "fill first." (e.g., a category with slots but an empty queue shows no note.)
- [ ] The note does **not** show for a brand-new category (no slots and no scheduled content).
- [ ] The note reflects the same fill-first behavior whether the user is adding a slot or moving an existing slot earlier (both can leave nearer, unfilled occurrences).
- [ ] The note is purely informational (no extra action/CTA) and uses theme-aware styling (no hardcoded colors).
- [ ] The note copy is added to the relevant settings namespace across every locale directory under `src/locales/`, English first.

### Mock-ups
Reference image provided (Add Slot modal). The note sits below the day selector / above or beside the Create button, as a full-width `Alert`. (Aligns with the BE behavior in the paired story.)

### Impact on existing data
None — UI copy/affordance only.

### Impact on other products
Web settings (Content Categories → Add Slot). No mobile/Chrome impact. White-label safe.

### Dependencies
- Pairs with **[BE] Fill newly-added or re-timed content category slots before scheduling after existing posts** (the note describes that behavior; ship together so the message is accurate).

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*
- `contentstudio-frontend/src/modules/setting/components/content-categories/dialogs/AddSlot.vue` — the Add Slot modal; title already toggles add vs edit via `getContentCategorySlotAdd._id`. Add the conditional `Alert` note here.
- `contentstudio-frontend/src/modules/setting/components/content-categories/dialogs/AddCategory.vue` — precedent for an `Alert` note inside a content-categories modal; follow the same pattern/placement.
- `contentstudio-frontend/src/modules/setting/composables/useAddSlot.ts` / `useContentCategories.ts` — gate the note on whether the selected category **has scheduled (queued) content**, not merely whether slots exist. If the queued-count isn't already available client-side, source it from the category data / a lightweight check rather than assuming "has slots = has content."
- Use `@contentstudio/ui` `Alert` (info variant); theme tokens only.
