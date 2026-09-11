# Research: time zone mismatch prompt

**Date:** 2026-09-07
**Trigger:** CEO shared a screenshot from another product showing a modal that detects a mismatch between the account time zone and the device time zone and offers to update it. Ask is to add the same to ContentStudio, with edge cases covered.

---

## 1. Backlog check

No existing timezone-mismatch story in `docs/stories/` or `docs/features/`. Timezone appears in adjacent work only: `analytics-report-date-time-preferences`, `public-api-workspace-settings`, `q2-2026-ios-improvements`. Nothing overlaps. This is new.

---

## 2. The headline finding: ContentStudio has no user time zone

The screenshot says "**your account** time zone". ContentStudio does not have one. Time zone lives on the **workspace**.

- `contentstudio-backend/app/Models/Settings/Workspace.php:60-68` sets `timezone` in the default workspace payload.
- `contentstudio-backend/app/Libraries/Account/Account.php:104-114` is the single read point: `Account::fetchWorkspaceTimezone($workspace_id)`, defaulting to `UTC`.
- The signup form does collect a `timezone` (`app/Http/Requests/Authentication/RegisterProfileRequest.php:45`), but `app/Repository/Account/UsersRepository.php:409-412` uses it only to seed the **first workspace's** timezone. It is never persisted on the user document.

**Why this matters more than it sounds.** In the source product, updating your time zone changes what you personally see. In ContentStudio, changing the workspace time zone changes **when scheduled content publishes for the entire team**. The same modal, ported literally, is a shared-state mutation dressed up as a personal preference. The copy, the permission model and the warnings all have to be different.

### Latent bug found along the way

`contentstudio-backend/app/Repository/Publish/Automation/EvergreenAutomationRepository.php:371-374` reads `$user->timezone` via `UsersRepository::getUserDetails(..., ['timezone'])` and falls back to `'UTC'`. Since no user document ever gets a `timezone` field written, that lookup always misses and evergreen automation last-execution time is computed in UTC rather than the workspace time zone. Not in scope for this story, but worth a separate ticket.

---

## 3. What already exists

### Device time zone detection is everywhere already

`Intl.DateTimeFormat().resolvedOptions().timeZone` is used in seven places in `contentstudio-frontend/src/`:

| Location | Purpose |
|---|---|
| `src/modules/account/views/SignUp.vue:288` | Seeds the first workspace time zone |
| `src/modules/account/components/CreateWorkspaceForm.vue:200` | Pre-fills time zone on new workspace |
| `src/modules/setting/components/workspace/WorkspaceFields.vue:351` | Pre-fills the workspace settings field |
| `src/modules/analytics/composables/usePDFReports.ts:661` | Report rendering |
| `src/modules/analytics/composables/competitor/useCompetitorReport.ts:104` | Report rendering |
| `src/api/auth.ts:256` | Device fingerprint for login security |
| `src/modules/setting/services.ts:40` | Device fingerprint for 2FA |

Plus `src/composables/useDateTime.ts:146` exposes `guessTimezone()` wrapping `dayjs.tz.guess()`.

**Careful:** the `tz` in `src/api/auth.ts:256` and `src/modules/setting/services.ts:40` feeds `app/Services/Auth/DeviceFingerprintService.php`, which is a **security fingerprint**, not a preference. Do not reuse or repurpose that channel for this feature. Coincidentally its docblock example at line 28 is `'tz' => 'Europe/Istanbul'`.

The detection primitive is therefore already solved. What is missing is the ongoing comparison and the prompt.

### The update endpoint already exists

`contentstudio-backend/app/Http/Controllers/Settings/WorkspaceController.php:85-128` implements `updateWorkspaceTimezone`. It validates, canonicalises via `Helper::mapToCanonicalTimezone`, updates, and returns the canonical value. The frontend already calls it from `src/api/onboarding.ts:99` via `src/modules/account/composables/useUserOnboarding.ts:300`.

`Helper::mapToCanonicalTimezone` (`app/Libraries/Helper.php:1038+`) maps legacy aliases such as `US/Eastern` to `America/New_York`. Useful, because browsers on older systems can still report legacy identifiers.

### Permission gap on that endpoint

`contentstudio-backend/routes/web/settings.php:95` registers `/updateWorkspaceTimezone` **outside** any `PermissionMiddleware` group. Compare `removeWorkspace` two lines below, which is wrapped in `PermissionMiddleware::class.':remove_workspace'`. As it stands, any authenticated member of a workspace can change the time zone that governs the whole team's publishing schedule. That is a real gap independent of this feature, and this feature would make it much easier to trip over.

### Time zone conversion layer

`contentstudio-frontend/src/utils/dayjs.ts` provides `getWorkspaceTimezone()` and a `dayjs.prototype.inWorkspaceTimezone()`. 199 frontend files reference timezone. `src/components/common/TimezoneSelect.vue` and `src/modules/setting/config/timezone.ts` (428 entries, `{ name, value }`) provide the picker.

### Dismissal persistence needs no backend work

`app/Repository/Account/UsersRepository.php:896-922` exposes `setPreferences` and `setPreferencesBulk`, both of which write arbitrary keys under `preferences.` with no whitelist. Route: `/preferences/setPreferences`. Existing precedents for one-time dismissals in `src/stores/core/useProfileStore.ts`: `black_friday_banner_status`, `dashboard_banner_view_status`, `show_inbox_sync_button_alert`, `feature_badges.seen`.

So the frontend can persist a dismissal record without any backend change.

---

## 4. What actually happens when the workspace time zone changes

This is the part the story has to get right, and the part the source screenshot's product does not have to worry about.

### Individually scheduled posts: safe, but relabelled

`execution_time` is stored in **UTC**. `app/Models/Publish/Planner/Plans.php:301` and `app/Models/Publish/Planner/V2/Plans.php:113` compare against `Carbon::now('UTC')`, and `app/Libraries/Publish/Planning.php:377` reads it as `new Carbon($execution_time['execution_time']['date'], 'UTC')`.

Consequence: a post scheduled for 3:00 PM Istanbul fires at the **same absolute instant** after the workspace moves to Madrid. But the planner will now display it as 2:00 PM. Nothing moved. It only looks like it did. Users will absolutely read that as "the app changed my schedule", so the copy has to pre-empt it.

### Queue slots: these really do move

`app/Libraries/Publish/QueueSlotsHelper.php` stores per-account `QueueSlots` as weekday plus wall-clock `times` arrays (`generateRandomQueueSlots` at line 109, `sortQueueSlots` at line 94). Slot resolution calls `Account::fetchWorkspaceTimezone($workspace_id)` at line 242 and evaluates with `Carbon::now($timezone)` at lines 322-366.

Consequence: a "9:00 AM" slot means 9:00 AM in whatever the workspace time zone currently is. Change the zone and every future queued post shifts by the offset in absolute terms. **This is a genuine behaviour change, not a display change.**

### Same wall-clock problem applies to

- Content category slots (`app/Models/Settings/ContentCategoriesSlots.php`, `app/Libraries/Publish/ContentCategorySlotsHelper.php`)
- Evergreen and RSS automation schedules (`app/Repository/Publish/Automation/`)
- Optimal posting time recommendations (`app/Http/Controllers/Api/V1/Scheduling/OptimalTimesController.php:156` resolves workspace timezone)

### Not affected

- Scheduled analytics reports carry their own `timezone` field (`app/Models/Analytics/ScheduleReportsModel.php:57`), so they are independent.
- Analytics data itself is stored against absolute timestamps and re-bucketed on read.

---

## 5. Edge cases the story must handle

Grouped by what breaks if they are missed.

### Comparison correctness

1. **Different name, same offset.** `Europe/Madrid` and `Europe/Paris` are both GMT+2. Prompting here is noise. Compare current UTC offsets, not IANA strings.
2. **DST drift.** A workspace on `Europe/Istanbul` (no DST, permanently GMT+3) against a device on `Europe/Madrid` (GMT+1 in winter, GMT+2 in summer) has a gap that changes twice a year. Comparing offsets at a single instant is right, but the prompt must not re-fire every March and October for someone who already declined.
3. **Legacy and alias identifiers.** A browser can report `Asia/Calcutta` or `US/Eastern`. Canonicalise both sides before comparing, otherwise identical zones look different.
4. **Device zone not in the 428-entry picker list.** Newer or obscure IANA zones may be missing. Decide the fallback rather than offering a value the picker cannot represent.
5. **Never use IP geolocation.** `Helper::getTimezoneFromIp` exists and is used at signup, but a VPN changes the IP and not the OS time zone. Device time zone only.

### Who should even see this

6. **Non-admins must not be prompted.** A viewer, approver or client guest cannot and should not change the whole team's publishing schedule. Given the endpoint currently has no permission gate, the frontend gating alone is not enough.
7. **The offshore team member case.** A US workspace with a VA in Asia. That VA's device is legitimately 10 hours off, permanently. Prompting them to move the workspace to their zone is actively harmful. This is the single strongest argument for gating on permission and for a durable per-workspace dismissal.
8. **The agency case.** One user, several workspaces, each deliberately set to a client's region. Prompting on every workspace switch would be intolerable. Dismissal must be per workspace, not global.

### Not nagging

9. **Travel.** Someone in Madrid for a week should be able to decline once and not see it again for that trip. Storing only a boolean is wrong. Store which device zone was declined, so a later move to a genuinely new zone can still prompt.
10. **Multiple workspaces, multiple dismissals.** Keyed by workspace plus declined device zone.
11. **An explicit opt out.** A "do not ask again for this workspace" control, so the agency and offshore cases can silence it permanently.

### Where and when not to show it

12. Onboarding, since the workspace zone was just set from this same device.
13. A newly created workspace, same reason.
14. Shared analytics links and public planner links, where the viewer has no account.
15. The sample or demo workspace.
16. Mid-flow in composer, publishing, or any modal that is already open.

### Consequences of saying yes

17. Currently queued posts shift by the offset. The user has to be told before they confirm, not after.
18. Displayed times for already-scheduled posts change without the posts moving. Also has to be said before they confirm.
19. A post due to publish within the next few minutes while the change is being applied.
20. Workspace is in paused-posting state.
21. Other browser tabs still showing the old zone. Everything on screen is now wrong until refresh.
22. Two admins changing it at the same time.
23. Whether the UI can re-render all times without a full reload. `dayjs.ts` reads the workspace zone through `getWorkspaceTimezone()`, so a store update should propagate, but every cached query result holding formatted strings needs invalidating.

### Presentation

24. White-label domains need correct theming.
25. Small viewports.
26. The offset label in the copy has to be computed at render time, since offsets change with DST.

---

## 6. Recommended shape

- Compare canonicalised device zone against canonicalised workspace zone by current UTC offset, once per session, after the app has booted and the active workspace is known.
- Show only to users who can change workspace settings.
- Frame the copy around the **workspace**, name it explicitly, and state the effect on queue slots and on displayed times before the user confirms.
- Two explicit buttons naming both zones, plus a "do not ask again for this workspace" checkbox.
- Persist dismissal as workspace plus declined device zone, under user preferences.
- Close the permission gap on the update endpoint as a paired backend story.

---

## 7. Open questions for the PO

1. Should this ever be offered to non-admins as a read-only heads-up ("this workspace runs on Europe/Istanbul, your device is on Europe/Madrid") with no action? Recommendation is no for v1, since it adds noise and the user can do nothing about it.
2. Should we also introduce a genuine per-user display time zone, so a traveller can see times in their local zone without changing the team's publishing schedule? That is the correct long-term answer to this problem and is a much larger piece of work. Worth logging separately.
3. Mobile: the Flutter app has the strongest case for this feature, since phones travel and change zone automatically. Not included in this batch. Recommend a follow-up once the web behaviour and copy have settled.
4. Should the prompt appear once per session, or once and then never until the device zone changes again? Recommendation is the latter.
