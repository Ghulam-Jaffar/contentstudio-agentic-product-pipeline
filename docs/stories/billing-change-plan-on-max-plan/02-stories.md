# Story: Show a plan action on the highest plan

## [FE] Show a "Change plan" action for users already on the highest plan

### Description

In Billing, users on lower plans see an "Upgrade plan" call to action, but users already on the highest available plan see no way to manage their plan at all. Those users may still want to change their billing cycle (for example monthly to annual) or downgrade. Show a plan action for them too, labelled "Change plan" (instead of "Upgrade plan") when they are on the top plan.

Highest available plans: Agency Unlimited on the new plans; Business and Agency on the old plans.

### Workflow

1. A user already on the highest available plan opens Billing.
2. Instead of no action, the user sees a "Change plan" button.
3. Clicking it opens the same plan selection and management flow used for upgrades, where the user can change billing cycle or move to a different plan, including a downgrade.

### Acceptance criteria

- [ ] Users on the highest available plan (Agency Unlimited for new plans, Business and Agency for old plans) see a "Change plan" action in Billing.
- [ ] Users on lower plans continue to see the existing "Upgrade plan" call to action, unchanged.
- [ ] The action label reads "Change plan" for highest-plan users and "Upgrade plan" for everyone else.
- [ ] Clicking "Change plan" opens the existing plan selection flow and allows changing billing cycle and downgrading.
- [ ] Existing billing-access rules still apply, so users who cannot manage billing do not see the action.

### UI copy

- Button on the highest plan: "Change plan"
- Button on lower plans (unchanged): "Upgrade plan"

### Impact on existing data

None.

### Impact on other products

Billing is web only. Mobile and Chrome extension: N/A.

### Dependencies

None.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only)
- [ ] Multilingual support verified (translations available or fallback handled)
- [ ] UI theming supported (default and white-label) — no dark mode
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

- Plans and CTA surface: `contentstudio-frontend/src/modules/billing/components/SubscriptionPlansMain.vue` (uses `modalContext = 'upgrade-plan'`; popular plan id `agency-unlimited`).
- Find where the upgrade CTA is currently hidden for highest-plan users and, instead of hiding, render the "Change plan" action that opens the same plan selection modal flow.
- Confirm the highest-plan detection: new plan id `agency-unlimited`; old plan ids `business` and `agency`.
