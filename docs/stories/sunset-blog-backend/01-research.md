# Research — Sunset blog publishing: backend code removal + data strategy

Two stories for the **Q2 - 2026: Miscellaneous** epic (effectively a small epic on its own):

1. `[BE]` Remove blog-publishing code from the backend.
2. `[BE]` Define and finalize the strategy for handling existing blog data (and confirm none are scheduled).

Context: blog publishing has been **sunset** — we no longer offer it. This cleans up the dead backend code and decides what happens to the data that remains.

---

## Current State

**How a blog post is identified:** blog posts live in the shared `plans` collection (MongoDB), marked by a `blog_selection` field (a nested object keyed per platform: `wordpress`, `tumblr`, `medium`, `shopify`, `webflow`). Social posts have no `blog_selection`. The same plan lifecycle/status field (`draft`, `scheduled`, `queued`, `published`, `failed`, …) applies to both.

**The publishing flow is shared with social** (the main removal risk):
- `app/Jobs/PlanPostingJob.php` dispatches `BlogUpdateJob` when `blog_selection` is set, else `SocialPlanPostingJob`.
- `app/Libraries/Publish/Posting/Posting.php` routes to `BlogPosting` or `SocialPosting`.
- `app/Jobs/PlanFinalizerJob.php` finalizes **both** blog and social, including dependent/secondary post chains.

### Blog code footprint (Story 1 removes)
- **Job:** `app/Jobs/BlogUpdateJob.php` (+ its Horizon supervisor `plan-update` / `BlogUpdateJob` queue in `config/horizon.php`).
- **Controllers:** `app/Http/Controllers/Integrations/Platforms/Blog/` (Blog, Wordpress, Medium, Webflow controllers).
- **Posting logic:** `app/Libraries/Publish/Posting/BlogPosting.php`; builders `app/Builders/Publish/Posting/{Wordpress,Tumblr,Medium,Shopify,Webflow}Posting.php`.
- **Repositories/services/helpers:** `app/Repository/Integrations/Platforms/BlogRepo.php`, `app/Repository/Integrations/Platforms/Blogs/*`, `app/Repository/Publish/Planner/BlogPostRepository.php`, `app/Libraries/Publish/Helper/BlogHelper.php`, `app/Libraries/Integrations/Platforms/Blog/Wordpress/*` (5 connection strategies).
- **Routes:** blog routes in `routes/web/integrations.php` and `routes/web/planner.php` (e.g. `/saveBlog`, `/fetchBlogs`, `/processBlogPost`, `/editWordpressBlogPost`, `/editMediumBlog`, `/editTumblrBlog`, …); blog AI route `/suggestBlogTitle` in `routes/web/ai.php`.
- **Models:** `BlogPosts`, `BlogIntegrations`, `WordpressBlogs`, `MediumBlogs`.
- **Shared (surgical edits, NOT delete):** `Posting.php` (remove blog branch), `PlanPostingJob.php` (remove blog dispatch), `PlanFinalizerJob.php` (remove blog finalization, keep social + dependent-social handling), permission `save_blog` (`PermissionHelper`), subscription limit `blogs` (`SubscriptionLimits`).

### Blog data footprint (Story 2 plans)
| Collection | Holds | Identify by |
|---|---|---|
| `plans` | blog posts (alongside social) | `blog_selection` not null |
| `blog_posts` | blog content/metadata | referenced by `plans.blog_reference` |
| `blog_integrations` | blog platform credentials | `type = 'Blog'` |
| `wordpress_blogs` | WordPress site metadata | — |
| `medium_blogs` | Medium publication metadata | — |
| `tumblr_accounts` | Tumblr blog accounts | `type = 'Blog'` (Tumblr social accounts are separate) |

- **Published blogs:** `plans` where `blog_selection` not null AND `status = 'published'`.
- **All blog plans:** `plans` where `blog_selection` not null.

### Scheduling precondition (Story 2 verifies)
- **Scheduled blogs:** `plans` where `blog_selection` not null AND `status ∈ {scheduled, queued}`. Story 2 must confirm this count is **zero** (or handle any that exist) before code removal proceeds.

---

## Gotchas
- **Tumblr is dual-use** (social + blog). Preserve Tumblr social posting and Tumblr social accounts; only remove the Tumblr *blog* variant.
- **Shared publishing pipeline** — removing the blog branch from `PlanPostingJob` / `Posting` / `PlanFinalizerJob` must leave the social path (and dependent-social-post chains tied to a primary post) fully intact.
- **Frontend still has blog UI** — removing backend endpoints must be coordinated with FE blog-UI removal (separate effort) so users aren't left calling dead endpoints. WordPress app/plugin integration is part of the same sunset.
- **Dependent plans** — secondary posts can reference a primary blog plan (`parent_reference` / `parent_posted`); the data strategy must account for these.
- **Orphaned records** — `plans.blog_reference` points to `blog_posts`; cleanup must handle orphans.
- **Multi-tenant/white-label** — all blog data is partitioned by `workspace_id`; cleanup is uniform across tenants.

## Files Involved
See the footprint tables above. Anchors: `app/Jobs/{BlogUpdateJob,PlanPostingJob,PlanFinalizerJob}.php`, `app/Libraries/Publish/Posting/{Posting,BlogPosting}.php`, `app/Models/Publish/Planner/Plans.php` (+ `BlogPosts`, `BlogIntegrations`, `WordpressBlogs`, `MediumBlogs`), `config/horizon.php`, `routes/web/{integrations,planner,ai}.php`.
