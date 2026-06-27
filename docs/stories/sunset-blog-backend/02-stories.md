# Stories — Sunset blog publishing (backend)

Three standalone stories for the **Q2 - 2026: Miscellaneous** epic. Each carries its own Shortcut fields block. Nothing is pushed to Shortcut — the Product Owner creates these manually using the **New Feature Template**.

> **Sequencing:** (1) **[BE] Define and finalize the strategy for handling existing blog data** first (it confirms no blog posts are still scheduled). (2) **[BE] Remove blog publishing code from the backend**. (3) **[BE] Execute the agreed blog data cleanup** — runs once the strategy is signed off; its details come from story 1.

---

## 1. [BE] Define and finalize the strategy for handling existing blog data

### Description:
As ContentStudio, we want a clear, agreed plan for what to do with the blog data we still store — published blog posts and all their associated records — now that blog publishing has been sunset. We need to decide whether to keep, export/archive, soft-delete, or hard-delete this data, document the plan, and confirm that **no blog posts are still scheduled** before any cleanup or code removal goes ahead.

The deliverable is a **finalized, signed-off strategy ready to execute** — not the deletion itself. We can't safely remove the blog code or storage without knowing what happens to existing blog data and confirming nothing is mid-flight.

---

### Workflow:
1. Inventory all existing blog data: count blog posts by status (published, draft, failed, etc.) and the associated records (blog content records, blog platform credentials, platform-specific blog records).
2. Check whether any blog posts are still in a scheduled (or queued-to-publish) state.
3. If any scheduled blogs exist, document how they're handled (let them publish out, or cancel them) so that none remain scheduled.
4. Evaluate the options for the remaining data — keep as-is, export/archive, soft-delete, or hard-delete — weighing storage, user impact, and data-retention/compliance.
5. Finalize one recommended approach and get sign-off from the team lead / PO.
6. Document the agreed plan (scope, method, ordering vs. code removal, backup/rollback, how it'll be executed and verified) so the execution can be scheduled as the next step.

---

### Acceptance criteria:

- [ ] A complete inventory of existing blog data is produced: counts of blog posts by status, plus the associated records (blog content records, blog platform credentials, platform-specific blog records).
- [ ] It is confirmed whether any blog posts are currently scheduled or queued to publish.
- [ ] If any scheduled/queued blog posts exist, a documented plan handles them so that **none remain scheduled** (this is the precondition that unblocks the code-removal story).
- [ ] The options for the remaining blog data are documented — keep as-is, export/archive, soft-delete (flag + hide), or hard-delete — each with pros/cons, storage impact, and data-retention/compliance considerations.
- [ ] A single recommended approach is finalized and **signed off by the team lead / PO**, covering all blog data (blog posts, blog content records, blog platform credentials, platform-specific blog records).
- [ ] The finalized strategy specifies: scope (which records), method (hard/soft/archive/keep), ordering relative to the code-removal story, a backup/rollback plan, and how the cleanup will be executed and verified.
- [ ] The strategy explicitly confirms **social data is left untouched** — e.g. Tumblr social accounts and the shared records used by social posts.
- [ ] Sign-off is captured so the execution work can be scheduled as the agreed next step.

---

### Open decision (this story exists to resolve it):
**Hard delete vs. soft delete vs. archive/export vs. keep** for existing blog data — to be decided and signed off with the team lead / PO as the core outcome of this story.

---

### Mock-ups:
N/A — research/strategy story.

---

### Impact on existing data:
This story makes **no** data changes itself. It produces the agreed plan for how existing blog data will be handled in a follow-on execution. (Any scheduled blogs found are dispositioned per step 3 so none remain scheduled.)

---

### Impact on other products:
- The chosen approach determines whether users retain access to historical blog records.
- **Mobile / Chrome extension:** N/A (blog publishing was web-only).
- **White-label:** applies uniformly — blog data is partitioned per workspace.

---

### Dependencies:
- **Blocks: [BE] Remove blog publishing code from the backend** — that story should proceed only after this story confirms no blog posts remain scheduled.
- Executing the chosen cleanup is the agreed next step once this strategy is signed off.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend research/strategy
- [ ] Multilingual support — N/A, no user-facing copy
- [ ] UI theming support — N/A, no UI
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Codebase:** `contentstudio-backend/` (Laravel 10, MongoDB).

**Where blog data lives:**
- `plans` collection — blog posts are the documents where `blog_selection` is not null (social posts have no `blog_selection`). Same `status` field as social.
- `blog_posts` — blog content/metadata (referenced from `plans.blog_reference`).
- `blog_integrations` — blog platform credentials (`type = 'Blog'`).
- `wordpress_blogs`, `medium_blogs` — platform-specific blog records.
- `tumblr_accounts` with `type = 'Blog'` — Tumblr **blog** accounts (Tumblr **social** accounts are separate and must be preserved).

**Queries to ground the inventory / precondition:**
- All blog plans: `plans` where `blog_selection` not null.
- Published blogs: `plans` where `blog_selection` not null AND `status = 'published'`.
- Scheduled/queued blogs (the precondition check — expect zero): `plans` where `blog_selection` not null AND `status ∈ {scheduled, queued}`.

**Gotchas to fold into the strategy:**
- Dependent/secondary posts can reference a primary blog plan (`parent_reference` / `parent_posted`).
- `plans.blog_reference` → `blog_posts` can leave orphans if not cleaned together.
- Tumblr is dual-use — do not touch Tumblr social accounts/data.
- All blog data is per-`workspace_id`; cleanup is uniform across tenants/white-label.

---

### Shortcut fields
- **Template:** New Feature Template
- **Story type:** Chore
- **Project:** Web App
- **Group:** Backend
- **Epic:** Q2 - 2026: Miscellaneous
- **Priority:** Medium
- **Product area:** Blog Post Composer
- **Skill set:** Backend
- **Estimate:** _(empty — devs estimate at sprint planning)_
- **Labels:** none
- **Iteration:** assigned by PO at creation

---

## 2. [BE] Remove blog publishing code from the backend

### Description:
As ContentStudio, we want to remove the now-sunset blog-publishing functionality from the backend codebase, so we stop maintaining dead code, shrink the surface area, and free queue/worker capacity reserved for a feature we no longer offer. This removes the code that created, scheduled, and published blog posts to blog platforms (WordPress, Medium, Tumblr blog, Shopify, Webflow) — while keeping **social publishing fully intact**, including Tumblr as a social destination.

This story removes **code only**. Existing blog data is left in place and handled by the separate strategy story.

---

### Workflow:
This is a backend cleanup/refactor — there is no user-facing flow. After this change:
- Blog publishing endpoints and the blog publishing job no longer exist.
- The shared publishing pipeline keeps working for social posts exactly as before.
- No queue capacity is reserved for blog publishing.

---

### Acceptance criteria:

- [ ] All blog-publishing code is removed: blog post create/edit/process endpoints, blog platform connection endpoints (WordPress, Medium, Tumblr blog, Shopify, Webflow), the blog publishing job, blog posting builders, blog repositories/services/helpers, and blog-specific routes.
- [ ] The blog-specific background queue/supervisor is removed from the queue configuration.
- [ ] Shared publishing code is **surgically updated, not deleted** — after the blog branch is removed, social posting (scheduling, publishing, retries, finalization, and dependent/secondary social-post handling) works exactly as before.
- [ ] **Tumblr as a social destination remains fully functional** — only the Tumblr *blog* posting path is removed.
- [ ] Blog-related permission and subscription-limit checks are removed without breaking the permission/subscription framework used by other features.
- [ ] Blog-related AI helper endpoints (e.g. blog title suggestions) are removed.
- [ ] Existing blog **data is not deleted or altered** by this story — removing code touches no blog records (data handling is covered by the separate strategy story).
- [ ] No dead references remain — no broken imports, no calls to removed blog classes or routes.
- [ ] End-to-end regression check: social posts schedule, publish, retry, and finalize correctly after the removal.

---

### Mock-ups:
N/A — backend cleanup.

---

### Impact on existing data:
None. Blog records (the blog documents in the shared posts collection and the blog-specific collections) are intentionally left in place; orphaned blog fields/records remain until the separate data-strategy story decides their fate.

---

### Impact on other products:
- **Frontend:** the blog composer/UI and any blog API consumers must be removed/disabled in tandem so users aren't left calling removed endpoints — tracked as a separate frontend effort.
- **WordPress app / plugin:** part of the same blog sunset — coordinate its removal/deprecation.
- **Mobile / Chrome extension:** N/A — blog publishing was web-only.
- **White-label:** removal is uniform across tenants.

---

### Dependencies:
- **Depends on: [BE] Define and finalize the strategy for handling existing blog data** — proceed only after that story confirms no blog posts remain scheduled.
- Coordinate with the frontend blog-UI removal so the two land together.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only
- [ ] Multilingual support — N/A, no user-facing copy
- [ ] UI theming support — N/A, no UI
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Codebase:** `contentstudio-backend/` (Laravel 10, MongoDB, Redis, Horizon).

**Delete outright:**
- `app/Jobs/BlogUpdateJob.php`
- `app/Http/Controllers/Integrations/Platforms/Blog/` (Blog, Wordpress, Medium, Webflow controllers)
- `app/Libraries/Publish/Posting/BlogPosting.php`
- `app/Builders/Publish/Posting/{Wordpress,Tumblr,Medium,Shopify,Webflow}Posting.php`
- `app/Repository/Integrations/Platforms/BlogRepo.php`, `app/Repository/Integrations/Platforms/Blogs/*`, `app/Repository/Publish/Planner/BlogPostRepository.php`
- `app/Libraries/Publish/Helper/BlogHelper.php`, `app/Libraries/Integrations/Platforms/Blog/Wordpress/*` (5 connection strategies)
- Models `BlogPosts`, `BlogIntegrations`, `WordpressBlogs`, `MediumBlogs` (deletion of model classes; the underlying collections are handled by the data-strategy story)
- Blog routes in `routes/web/integrations.php`, `routes/web/planner.php`, and `/suggestBlogTitle` in `routes/web/ai.php`
- The `BlogUpdateJob` queue/supervisor (`plan-update`) in `config/horizon.php`

**Surgical edits (keep the file, remove only the blog branch):**
- `app/Jobs/PlanPostingJob.php` — dispatches `BlogUpdateJob` when `blog_selection` is set; remove the blog dispatch, keep the social dispatch.
- `app/Libraries/Publish/Posting/Posting.php` — routes to `BlogPosting` vs `SocialPosting`; remove the blog branch.
- `app/Jobs/PlanFinalizerJob.php` — finalizes blog **and** social; remove blog-specific finalization but **preserve** the social path and dependent/secondary social-post handling.
- `app/Libraries/Permission/PermissionHelper.php` (`save_blog`), `app/Libraries/Settings/SubscriptionLimits.php` (`blogs` limit) — remove blog checks without breaking the framework.

**Gotchas:**
- **Tumblr dual-use** — preserve `TumblrSocialPosting` and Tumblr social accounts; only remove the Tumblr blog variant.
- The blog vs social distinction everywhere is `isset($plan['blog_selection'])` — use that as the seam, and verify the social path after each edit.
- Frontend still calls blog endpoints — removal must be coordinated with the FE blog-UI removal.

---

### Shortcut fields
- **Template:** New Feature Template
- **Story type:** Chore
- **Project:** Web App
- **Group:** Backend
- **Epic:** Q2 - 2026: Miscellaneous
- **Priority:** Medium
- **Product area:** Blog Post Composer
- **Skill set:** Backend
- **Estimate:** _(empty — devs estimate at sprint planning)_
- **Labels:** none
- **Iteration:** assigned by PO at creation

---

## 3. [BE] Execute the agreed blog data cleanup

> **Placeholder by design.** The specifics (scope, method, steps, verification) come from **[BE] Define and finalize the strategy for handling existing blog data**. Fill this story out once that strategy is signed off.

### Description:
As ContentStudio, we want to carry out the blog-data cleanup exactly as decided in the strategy story — applying the agreed method (keep / archive / soft-delete / hard-delete) to all existing blog data — so the blog-publishing sunset is fully complete on the data side.

---

### Workflow:
Execute the cleanup plan finalized in the strategy story, in the agreed order, with the agreed backup/rollback in place, then verify the end state.

---

### Acceptance criteria:

- [ ] The blog data cleanup is executed per the approach finalized in **[BE] Define and finalize the strategy for handling existing blog data** (scope, method, ordering, backup/rollback).
- [ ] Only blog data is affected — social data (including Tumblr social accounts and shared records used by social posts) is left untouched.
- [ ] A backup / rollback path is in place before any destructive step.
- [ ] Execution is verified — blog data is in the expected end state and the counts match the pre-cleanup inventory.
- [ ] _(Detailed acceptance criteria to be added from the finalized strategy.)_

---

### Mock-ups:
N/A — backend data operation.

---

### Impact on existing data:
This is the story that actually changes/removes existing blog data, per the method agreed in the strategy story.

---

### Impact on other products:
- Determines what historical blog data remains accessible.
- **Mobile / Chrome extension:** N/A.
- **White-label:** applies uniformly (blog data is per workspace).

---

### Dependencies:
- **Depends on: [BE] Define and finalize the strategy for handling existing blog data** (must be signed off first — this story executes its outcome).
- Typically runs after **[BE] Remove blog publishing code from the backend**, per the finalized strategy's ordering.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend data operation
- [ ] Multilingual support — N/A, no user-facing copy
- [ ] UI theming support — N/A, no UI
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Shortcut fields
- **Template:** New Feature Template
- **Story type:** Chore
- **Project:** Web App
- **Group:** Backend
- **Epic:** Q2 - 2026: Miscellaneous
- **Priority:** Medium
- **Product area:** Blog Post Composer
- **Skill set:** Backend
- **Estimate:** _(empty — devs estimate at sprint planning)_
- **Labels:** none
- **Iteration:** assigned by PO at creation
