# Flutter Migration — Story Titles + Generic Template

ContentStudio is migrating the separate native **iOS** and **Android** apps to a **single Flutter app**. No more separate `[iOS]`/`[Android]` stories — these use a `[Flutter]` prefix and the **Mobile** project. Pick the generic body that fits (Feature/Screen vs DevOps/Infra) and copy-paste it as-is — nothing to fill in; the body refers to whatever the story title says.

> Suggested Shortcut fields for all: **Project:** Mobile · **Skill set:** Frontend · **Product area:** (use a mobile area) · **Type:** Feature (or Chore for infra). No estimates, no labels.

---

## Titles

### Screens / features (use the Feature template)
1. `[Flutter] Build Home with Planner`
2. `[Flutter] Build Post Preview`
3. `[Flutter] Build Calendar`
4. `[Flutter] Build Composer`
5. `[Flutter] Build AI Assistant`
6. `[Flutter] Build Inbox`
7. `[Flutter] Build Conversation view`
8. `[Flutter] Build Post Comments`
9. `[Flutter] Build Social Listening`
10. `[Flutter] Build Approval Workflow`
11. `[Flutter] Build Settings`
12. `[Flutter] Build Menu`
13. `[Flutter] Build Workspaces`
14. `[Flutter] Build Tab bar`
15. `[Flutter] Build Onboarding`
16. `[Flutter] Build Banner view`
17. `[Flutter] Build Web views`

### Cross-cutting / platform (use the Feature template)
18. `[Flutter] Implement Authentication`
19. `[Flutter] Implement SSO`
20. `[Flutter] Implement Session management`
21. `[Flutter] Implement App access handling`
22. `[Flutter] Implement Push notifications`
23. `[Flutter] Improve push notification workflow (iOS & Android)`
24. `[Flutter] Implement Payments`
25. `[Flutter] Implement Force update`
26. `[Flutter] Implement Localization`
27. `[Flutter] Implement Analytics events`

### DevOps / infra (use the Infra template)
28. `[Flutter] Set up build flavors`
29. `[Flutter] Set up Fastlane`
30. `[Flutter] Set up Android release pipeline`
31. `[Flutter] Set up iOS release pipeline`
32. `[Flutter] Automate the development process`
33. `[Flutter] Testing and bug fixing`

---

## Generic story body — Feature / Screen

> Copy-paste as-is. The body refers to "this screen/feature" — i.e. whatever the story title names.

### Description
As a ContentStudio mobile user, I want this screen/feature to work in the new unified ContentStudio Flutter app with the same capabilities I have today in the native iOS and Android apps, so that I get one consistent mobile experience on a single codebase after the migration.

### Workflow
1. User opens the ContentStudio mobile (Flutter) app.
2. User navigates to this screen/feature.
3. User performs the same actions available today, and sees the same results.
4. Behavior is identical on iOS and Android, served from the single Flutter codebase.

### Acceptance criteria
- [ ] This screen/feature is implemented in the Flutter app with feature parity to the current native iOS and Android apps.
- [ ] All user flows work on both iOS and Android from the single Flutter codebase.
- [ ] It uses the existing ContentStudio backend APIs — no new backend behavior, no regressions vs the native apps.
- [ ] UI matches the agreed design/parity and uses the app's shared components, theme, and navigation.
- [ ] All user-facing text is localized.
- [ ] Loading, empty, and error states are handled with clear messaging.
- [ ] Analytics events fire as they do in the native apps.
- [ ] Works across supported iOS and Android OS versions and common phone/tablet sizes.
- [ ] No crashes or blocking issues; verified on both platforms.

### Mock-ups
Parity with the current native iOS/Android screens. Link Figma/design here if a refreshed design exists.

### Impact on existing data
None — client rebuild against the same backend APIs. No schema or data changes.

### Impact on other products
Replaces the native iOS & Android apps for this screen/feature with the Flutter app. Web app and Chrome extension are unaffected. Backend APIs unchanged.

### Dependencies
Foundational migration stories (authentication, session management, app access handling, tab bar, menu, localization, analytics events) may need to be in place first.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Generic story body — DevOps / Infra

> Copy-paste as-is. The body refers to "this task" — i.e. whatever the story title names.

### Description
As the ContentStudio mobile team, we need the task in this story set up for the new Flutter app so that we can build, release, and maintain the unified app reliably across iOS and Android.

### Workflow
1. Engineer runs the documented command/process for this task.
2. The process completes successfully for both iOS and Android (where applicable).
3. Outcome is repeatable by any team member from the documented steps.

### Acceptance criteria
- [ ] This task is configured for the Flutter app and works for both iOS and Android where applicable.
- [ ] The setup is documented so any team member can run/repeat it.
- [ ] The process is reliable and repeatable (no manual one-off steps that aren't documented).
- [ ] Where relevant, it integrates with the existing release/CI tooling and credentials.
- [ ] Verified end-to-end at least once on both platforms (where applicable).

### Mock-ups
N/A — infrastructure/tooling.

### Impact on existing data
None.

### Impact on other products
Replaces the equivalent native iOS/Android tooling. No impact on web or backend.

### Dependencies
Foundational setup (build flavors, Fastlane) may need to be in place before the release-pipeline tasks.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, tooling/infra story
- [ ] Multilingual support — N/A, tooling/infra story
- [ ] UI theming support — N/A, tooling/infra story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Note
- **AI Assistant on mobile:** In scope. AI chat is already available in the native iOS app, so it must be rebuilt in Flutter too. (This is an exception to the older "AI is web-only" platform rule — AI chat/assistant exists on mobile.)
