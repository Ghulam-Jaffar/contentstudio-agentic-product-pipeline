# Story: White-label reseller banner

## [FE] Add a white-label reseller awareness banner (super admins and billing-access admins)

### Description

On the white-label page, add a banner that tells eligible users about the reseller feature, which lets them resell ContentStudio to their clients under their own brand. Show it only to super admins and to admins who have billing access.

### Workflow

1. A super admin, or an admin with billing access, opens the white-label page.
2. They see a banner introducing the reseller feature, with two actions: Contact support and Learn more.
3. Admins without billing access, and non-admins, do not see the banner.

### Acceptance criteria

- [ ] A banner appears on the white-label page for super admins and for admins with billing access.
- [ ] The banner is hidden for users without billing access and for non-admins.
- [ ] The banner shows the reseller headline, the description, and two actions: "Contact support" and "Learn more".
- [ ] "Contact support" opens the support contact flow.
- [ ] "Learn more" opens the reseller information link.
- [ ] The banner uses a design-library component and theme-aware styling.

### UI copy

- Headline: "White-label reseller"
- Description: "Resell ContentStudio to clients as a complete social media management platform under your brand."
- Primary action: "Contact support"
- Secondary action: "Learn more"

### Design

Refer to the product designer for the banner's visual design. Final visuals (layout, imagery, exact styling) to be provided by the product designer. Build against a design-library banner component in the meantime.

### Impact on existing data

None.

### Impact on other products

White-label is web only. Mobile and Chrome extension: N/A.

### Dependencies

Design input from the product designer for the banner visuals.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only)
- [ ] Multilingual support verified (translations available or fallback handled)
- [ ] UI theming supported (default and white-label) — no dark mode
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

- White-label page: `contentstudio-frontend/src/modules/setting/components/white-label/GeneralSettings.vue` (white-label section under Settings).
- Gating helpers already exist in the codebase: `isSuperAdmin` and `hasBillingAccess` / `billing_access`. Gate the banner on super admin OR admin-with-billing-access.
- Use a design-library banner or alert component (see `docs/ui-components.md`: `Alert`, or the full-width `CstBanner`) with theme-aware classes, rather than custom markup.
