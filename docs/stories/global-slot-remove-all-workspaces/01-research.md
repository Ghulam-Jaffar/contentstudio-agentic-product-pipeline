# Research: remove a global category slot from all workspaces

**Date:** 2026-09-08
**Trigger:** The Add Slot modal offers "Add slots within this workspace" / "Add slots in all workspaces" for global content categories. The remove-slot flow has no equivalent choice. Ask is to add it.

---

## 1. Backlog check

No existing story for this in `docs/stories/` or `docs/features/`. Adjacent content-category work exists but none of it touches the slot removal scope. New.

---

## 2. The asymmetry, confirmed

Three operations on global content category slots, three different behaviours:

| Operation | Endpoint | Global behaviour | User choice? |
|---|---|---|---|
| Add one slot | `POST global/categories/slots/store` → `globalSlotsStore` | Iterates every workspace copy of the global category and creates the slot in each | **Yes**, explicit `local` / `global` radio |
| Remove **all** slots of a category | `POST categories/slots/delete_all` → `deleteAll` | Always removes from every workspace copy, no choice offered | **No**, automatic |
| Remove **one** slot | `POST categories/slots/delete` → `remove` | Current workspace only | **No option at all** |

The single-slot removal is the odd one out. A user can add a slot to twelve workspaces in one click, then has to visit all twelve to take it back.

---

## 3. How the add-all-workspaces path works

`contentstudio-backend/app/Http/Controllers/Settings/ContentCategories/ContentCategoriesSlotsController.php:236-311` (`globalSlotsStore`)

1. Validates `category_id`, `workspace_id`, `minute`, `hour`, `period`, `weekdays`.
2. Checks access via `ContentCategoryAccessService::userHasAccessToCategory`.
3. Normalises hour 12 to 0.
4. Loads the category, confirms `category_state === 'global'`.
5. Resolves sibling copies with `ContentCategoriesRepository::getByGlobalCategoryId($category->global_content_category_id)`.
6. For each copy, for each weekday: `ContentCategoriesSlotsRepository::createOrUpdate` then `ContentCategorySlotBackfillService(workspace_id, category_id)->backfill([day, hour, minute, period])`.

Note the comment at line 277: "each copy has its own last-occupied post, timezone and removed_slots". So a per-workspace backfill service instance is required, not one shared instance.

### Frontend side

`contentstudio-frontend/src/modules/setting/components/content-categories/dialogs/AddSlot.vue:217-253`

Two `Radio` components bound to `getContentCategorySlotAdd.slots_state`, values `local` and `global`, each with a `CircleHelp` `Icon` carrying a tooltip. The block renders only when **all** of these hold:

- `!getContentCategorySlotAdd._id` (create, not edit)
- `getContentCategorySlotAdd.category_id` is set
- `selected_category.category_state === 'global'`

Copy keys, `contentstudio-frontend/src/locales/en/settings.json:3200-3207`:

```
settings.add_slot.slot_options.workspace_local  = "Add slots within this workspace"
settings.add_slot.slot_options.workspace_global = "Add slots in all workspaces"
settings.add_slot.tooltips.workspace_local      = "Your category slots will be created only within this workspace."
settings.add_slot.tooltips.workspace_global     = "Your category slots will be created in all workspaces where your global category is present."
```

---

## 4. How the single-slot removal works today

### Backend

`ContentCategoriesSlotsController.php:127-157` (`remove`)

1. Validates `slot_id` only. No `workspace_id`, no `category_id`, and **no access check** (unlike `globalSlotsStore` and `deleteAll`, both of which call `ContentCategoryAccessService::userHasAccessToCategory`).
2. `ContentCategoriesSlotsRepository::getById($slot_id)` then `removeById($slot_id)`.
3. On success, purges the slot's back-filled occurrences: `new ContentCategorySlotBackfillService($slot['workspace_id'], $slot['category_id'])->purge([day, hour, minute, period])`.

### Frontend

`contentstudio-frontend/src/modules/setting/composables/useContentCategories.ts:292-311` (`removeSlotModal`)

A plain `$cstuModal.msgBoxConfirm` with `settings.content_categories.popups.remove_slot.title` and `.description`, then `contentCategoryStore.deleteCategorySlot(slotPayload)`.

`contentstudio-frontend/src/stores/setting/useContentCategoryStore.ts:627-634` calls `deleteCategorySlotApi(payload.slotId)` with the slot id alone, then splices the slot out of local state by `_id`.

Entry points: `calendar/CalendarSlot.vue:20` and `calendar/TableData.vue:99-100`.

**Implication for the build:** the confirmation is currently a message box, which cannot host radio inputs. Adding the scope choice means turning it into a small custom modal. That is the main non-trivial part of the frontend work.

---

## 5. The join key for matching a slot across workspaces

Slots in sibling workspaces are separate documents with **different `_id`s**, so the removal cannot work from a slot id alone.

`globalSlotsStore` creates them from the tuple `(day, hour, minute, period)` against each workspace's category copy. That same tuple is therefore the correct match key for removal:

1. From the clicked slot, read `category_id` and `workspace_id`.
2. Load that category, read `global_content_category_id`.
3. `getByGlobalCategoryId` to get every sibling copy.
4. In each copy, find slots matching `day`, `hour`, `minute`, `period` and remove them.
5. Purge backfill per workspace, using a fresh `ContentCategorySlotBackfillService` per copy.

`ContentCategoriesSlotsRepository` already has `removeById`, `removeByCategoryId` and `removeByCategoryIds`. A method that removes by category ids plus the time tuple does not exist yet and is the new repository work.

---

## 6. Things found along the way

1. **`remove` has no access check.** `globalSlotsStore` and `deleteAll` both call `ContentCategoryAccessService::userHasAccessToCategory`. The single-slot `remove` validates only that `slot_id` is present. Worth adding, and it becomes more important once one call can affect many workspaces.

2. **`deleteAll` does not purge backfill.** Its global branch calls `ContentCategoriesSlotsRepository::removeByCategoryIds($categoriesIds)` and never touches `ContentCategorySlotBackfillService`, while the single-slot `remove` does purge. That likely leaves stale `removed_slots` entries behind after a bulk delete. Separate ticket, not this one, but the new global-remove path must not repeat the omission.

3. **Hour 12 normalisation.** `globalSlotsStore` maps hour `12` to `0` before storing. Any matching logic on the removal side has to apply the same normalisation or a 12 o'clock slot will not match.

4. **The Add Slot radio is hidden on edit.** Only shown on create. Worth deciding whether the removal choice should likewise be hidden in any context, though there is no equivalent edit case for removal.

---

## 7. Recommended shape

- Mirror the Add Slot pattern exactly: two radios, same tooltip style, shown only when the category is global, defaulting to **this workspace only** as the safer option.
- For local categories, keep today's plain confirmation with no radios.
- New backend endpoint or a scope parameter on the existing one. A parameter on the existing `delete` route keeps the surface smaller, but a sibling of `global/categories/slots/store` is more consistent with how the add path is organised. Either is fine, dev's call.
- Add the missing access check while touching this code.

---

## 8. Open questions for the PO

1. Should this also be offered on "Remove all slots", which currently removes from every workspace with no choice for global categories? Today's behaviour is inconsistent with Add, and arguably that dialog needs the same radio. Out of scope unless you want it.
2. What should happen to posts already scheduled into the removed slot in the other workspaces? Recommendation is to match today's single-workspace behaviour exactly, so no new rules are introduced by this story.
3. No `[Design]` story is proposed, because the layout copies the Add Slot modal one for one. If you want a mock of the new confirmation modal before build, add one.
