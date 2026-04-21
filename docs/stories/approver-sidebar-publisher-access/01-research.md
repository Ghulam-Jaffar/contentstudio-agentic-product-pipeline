# Research — Enable Publisher Sidebar for Approver Role

## Current state

The publisher module's left sidebar is conditionally rendered in [PublisherMain.vue:3](contentstudio-frontend/src/modules/publisher/views/PublisherMain.vue#L3):

```vue
<SidebarMain v-if="canAccessSidebar" ... />
```

Where `canAccessSidebar` is:

```ts
const canAccessSidebar = computed(() => {
  return (
    hasPermission('can_access_top_header') || hasPermission('can_create_post')
  )
})
```

And in [usePermission.ts:163-175](contentstudio-frontend/src/composables/usePermission.ts#L163-L175), the approver role only returns `true` for these permissions:

```ts
if (user.role === 'approver') {
  switch (action) {
    case 'can_schedule_plan':     return true
    case 'can_add_notes':          return user?.permissions?.approverCanAddNotes ?? false
    case 'can_create_post':        return user?.permissions?.approverCanCreatePost ?? false
    default:                       return false   // everything else, including can_access_top_header
  }
}
```

So for approvers:

| Approver sub-permission | `canAccessSidebar` | Sidebar shown? |
|---|---|---|
| `approverCanCreatePost = true` | true (via `can_create_post`) | Yes — limited version |
| `approverCanCreatePost = false` | false | **No — entire sidebar hidden** |

This is the gap the PO flagged: approvers without create-post permission see no sidebar at all in the publisher. They have no way to browse their planner custom views.

## Inside the sidebar — what's already correctly gated

The permission guards inside [SidebarMain.vue](contentstudio-frontend/src/modules/publisher/components/SidebarMain.vue) already produce the exact layout the PO wants for approvers — we just need to stop hiding the whole sidebar.

- **Composer button** (expanded [L12](contentstudio-frontend/src/modules/publisher/components/SidebarMain.vue#L12) + shrunk [L519](contentstudio-frontend/src/modules/publisher/components/SidebarMain.vue#L519)) — gated by `canAccessTopHeader = can_access_top_header || can_create_post`. For an approver this is `approverCanCreatePost`. Template option is on by default via `:show-template-attachment="true"` in the ComposeActionsDropdown ([L25](contentstudio-frontend/src/modules/publisher/components/SidebarMain.vue#L25)). Blog composer action is hidden because `:can-access-blog-composer="hasFullAccess"` is false for approvers.
- **Planner with Custom Views collapsible** ([L100-L350](contentstudio-frontend/src/modules/publisher/components/SidebarMain.vue#L100)) — unconditionally rendered. Shows for every user who gets the sidebar.
- **AI Studio, Automations, Planner Settings sections** ([L352, L409, L466](contentstudio-frontend/src/modules/publisher/components/SidebarMain.vue#L352)) — all gated by `hasFullAccess = hasPermission('can_access_top_header')`. For approvers this is always false, so these are hidden — matches the PO requirement of "nothing below Planner, no AI, no automation, no settings."
- **Shrunk body** ([L598](contentstudio-frontend/src/modules/publisher/components/SidebarMain.vue#L598)) — `getFilteredItems()` returns only `[plannerItem]` for non-full-access users ([L1311-L1323](contentstudio-frontend/src/modules/publisher/components/SidebarMain.vue#L1311)), and then filters planner out of the "other items" loop. Shrunk view for approver shows only: compose icon (if `can_create_post`) + planner dropdown. Clean.

## What needs to change

**One file, one condition.** Update `canAccessSidebar` in [PublisherMain.vue:38-42](contentstudio-frontend/src/modules/publisher/views/PublisherMain.vue#L38-L42) to also return true for approvers regardless of `approverCanCreatePost`:

```ts
const canAccessSidebar = computed(() => {
  return (
    hasPermission('can_access_top_header') ||
    hasPermission('can_create_post') ||
    /* approver, even without approverCanCreatePost */
    isApproverRole.value
  )
})
```

`isApproverRole` check can come from either:
- a new helper in `usePermissions` (e.g. `isApprover()`) — preferred, keeps role checks centralized
- or reading the user role directly from the profile store

All other guards (`canAccessTopHeader` for Composer visibility, `hasFullAccess` for AI/Automation/Settings sections) stay as-is — they already produce the right layout.

## UX reference

Not applicable — purely an internal permission fix; no external UX pattern needed.

## Files involved

- [contentstudio-frontend/src/modules/publisher/views/PublisherMain.vue](contentstudio-frontend/src/modules/publisher/views/PublisherMain.vue) — `canAccessSidebar` computed
- Optionally [contentstudio-frontend/src/composables/usePermission.ts](contentstudio-frontend/src/composables/usePermission.ts) — add an `isApprover()` helper if we go that route

No backend changes. No other frontend files touched. No mobile impact — this is a web-only publisher route.

## Notes for implementation

- The v1/v2 legacy `composables/useApproval.ts` already references `hasPermission?.('can_create_post')` — this story doesn't change that flow.
- `PublisherMain.vue` is already a `<script setup>` file; no Options API conversion needed.
- No i18n strings added or changed.
- No new components; no new routes.
