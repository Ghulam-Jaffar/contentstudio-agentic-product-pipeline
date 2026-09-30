# Stories: LinkedIn Profile Analytics access at the super admin level

## [BE] Grant LinkedIn Profile Analytics access at the super admin level

### Description:

As the ContentStudio support team, we want to turn on LinkedIn Profile Analytics once for a customer's super admin and have it apply to all of their workspaces and team members, so that we stop collecting and enabling individual emails one by one, and so that access works the same way the rest of a customer's plan and add-on access does.

Today access is granted per person. A customer asks, sends us the email addresses of the people who want it, and we enable exactly those users. Anyone on the same team who was not on the list cannot see the feature, which leads to repeat requests and confusion when teammates see different things in the same workspace.

---

### Workflow:

1. A customer asks support for LinkedIn Profile Analytics.
2. Support enables it for the customer's super admin (the account owner). No list of team member emails is needed.
3. The super admin opens Analytics in any of their workspaces and sees LinkedIn personal profiles alongside Company Pages.
4. A team member signs in, opens a workspace owned by that super admin, and sees the same LinkedIn personal profiles in Analytics, including the prompt to reconnect a profile for analytics.
5. A team member who also belongs to a workspace owned by a different super admin, who does not have access, does not see LinkedIn personal profiles in Analytics while in that other workspace.
6. When a new team member is invited to one of the super admin's workspaces, they get the feature straight away, with no extra request to support.

---

### Acceptance criteria:

- [ ] When the super admin of a workspace has LinkedIn Profile Analytics enabled, every team member of that workspace sees LinkedIn personal profiles in Analytics, whatever their role
- [ ] This applies across every workspace owned by that super admin, with no per-workspace or per-user setup
- [ ] A newly invited team member gets access as soon as they join one of the super admin's workspaces
- [ ] When the super admin does not have access, team members of their workspaces do not see LinkedIn personal profiles in Analytics, even if the team member has access in a different super admin's workspace
- [ ] Switching workspaces updates access to match the super admin of the workspace now in use, without signing out
- [ ] A team member of an enabled super admin can complete the reconnect-for-analytics flow for a LinkedIn personal profile, and the profile's analytics load afterwards
- [ ] The analytics data for an enabled super admin's LinkedIn personal profiles is fetched and kept up to date regardless of which team member connected the profile
- [ ] Turning access off for the super admin removes LinkedIn personal profiles from Analytics for the super admin and all of their team members
- [ ] Users who were enabled individually before this change keep access after release, so no current customer loses the feature
- [ ] ContentStudio internal staff keep access in every workspace, as today
- [ ] The web app and the analytics data pipeline agree on who has access in every case above (nobody sees profiles in the picker whose data is never fetched, or the reverse)

---

### Mock-ups:

N/A. No new screens or copy. The existing LinkedIn Profile Analytics screens are unchanged.

---

### Impact on existing data:

- No schema change for workspaces or teams.
- Existing individual grants stay in place. Support should move grants that were given to team members over to the matching super admin, so access becomes consistent for the whole team.

---

### Impact on other products:

- Web app: LinkedIn personal profiles in Analytics follow the super admin rule automatically. No visible change for users who already have access.
- Mobile app (Flutter): N/A, LinkedIn Profile Analytics is not in the mobile app.
- Chrome extension: N/A.
- White-label: a white-label customer's super admin is enabled the same way, and their team members inherit access on white-label domains too.

---

### Dependencies:

- Builds on **[BE] Add LinkedIn profile analytics authorization and a request-access feature flag**.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (N/A, backend-only story)
- [ ] Multilingual support (N/A, no new copy)
- [ ] UI theming support (N/A, no UI change)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
- [ ] Developer surfaces coverage (N/A, no new or changed API)
