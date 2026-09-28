# Research: Separate GCS buckets per environment (QA / Staging / UAT / Production)

## Current State

**Every environment writes into the same three production GCS buckets.** There is no environment dimension in storage at all — not in the bucket name, not in a path prefix, not in the GCP project.

### The buckets

`contentstudio-backend/config/filesystems.php` defines three GCS disks plus a legacy S3 disk:

| Disk | Bucket (default) | Env var | Used for |
|---|---|---|---|
| `gcs` | `lumotive-web-storage` | `GOOGLE_CLOUD_STORAGE_BUCKET` | Legacy: publish media, settings logos, analytics company logos, discovery topic art, social profile images, inbox attachments |
| `gcs-contentstudio-media-library-nearline` | `contentstudio-media-library-nearline` | `GCS_ML_NEARLINE_STORAGE_BUCKET` | **Primary.** Media Library uploads, thumbnails, CSV exports, AI-generated images, signed-upload targets |
| `gcs-contentstudio-media-library-standard` | `contentstudio-media-library-standard` | `GCS_ML_STANDARD_STORAGE_BUCKET` | Secondary media library tier |
| `s3` | `contentstudio-assets` | `AWS_BUCKET` | Legacy, effectively dormant |

All three GCS disks share one project id (`GOOGLE_CLOUD_PROJECT_ID`, default `contentstudio-156007`) and **one service-account credential blob** (`GCS_CREDENTIALS_JSON`), so a non-production deploy holds full read/write/delete rights on production customer media.

### Why setting the env var is not enough

The bucket names *are* env-driven at the disk level, but the literal bucket strings are also **hardcoded in 185 places across 21 backend PHP files** and in a second service. They are used for two things that break the moment a non-prod bucket is introduced:

1. **URL → storage key derivation** — `str_replace('https://storage.googleapis.com/contentstudio-media-library-nearline', '', $url)`. With a different bucket the replace is a no-op, the stored `key` keeps the full URL, and every later operation keyed on it (delete, signed URL, thumbnail, CSV re-use) silently misses.
2. **Disk routing** — `str_contains($link, '/contentstudio-media-library-nearline/')` chooses which disk to act on. In a non-prod env nothing matches, so the code falls through to the legacy branch or does nothing.

Highest-traffic offenders:

- `app/Http/Controllers/Storage/MediaLibrary/MediaLibraryAssetsController.php` — key derivation at L162, L298, L451, L1360, L2000, L2219; disk routing at L920-924; hardcoded prefix list at L1749
- `app/Libraries/Media.php` L548-555 — the three-way bucket router
- `app/Libraries/Storage/MediaLibrary.php` L53, L72, L119, L216, L235 — `Storage::disk('gcs-contentstudio-media-library-nearline')` (disk *name*, so env-safe, but hardcoded to nearline)
- `app/Libraries/MediaLibrary/GcsSignedUrlGenerator.php` L68 — signed upload URLs
- `app/Repository/Utilities/MediaRepository.php` L1331, `app/Http/Controllers/Storage/MediaController.php` L315
- `app/Builders/EvergreenBulkBuilder.php` L145

### Two config arrays that are not env-driven at all

- **`config/filesystems.php` L103 `cs_gcs_buckets`** — a hardcoded literal array of the three production bucket names. Read by `Helper::isMediaFromGcsStorage()` (`app/Libraries/Helper.php` L953), which decides "is this URL already ours, or does it need re-hosting?". In a non-prod env with its own buckets this returns `false` for the env's own media, causing needless re-download and re-upload loops.
- **`config/app.php` L457 `media_storage_url_prefix`** — hardcoded to `https://storage.googleapis.com/contentstudio-media-library-nearline`. Used by `app/Builders/CSVBuilder.php` L519, L1126, L1178 and `app/Http/Controllers/Api/V1/PostController.php` L2549, L2766.

### Other services that write to the same buckets

- **`social-inbox-manager/`** hardcodes the bucket as a Python string literal, not config: `app/utils/instagram_helper.py` L1357, L1503; `app/utils/facebook_helper.py` L1153, L1289; `app/utils/linkedin_helper.py` L907 — each followed by `upload_content_to_gcs(file_data, bucket, blob, content_type)`. So QA inbox sync writes attachments into the production legacy bucket.
- **`contentstudio-ai-agents/`** does not hold GCS credentials; it hands assets to the backend to re-host (`src/orchestration/carousel/delivery.py`, `src/api/routers/workflows/workflow_router.py` L292 builds `media_library/{workspace_id}/uncategorized/original`). It inherits whatever the backend is pointed at, so no change needed there beyond the backend fix.
- **`contentstudio-social-analytics-go-main/`** — no GCS bucket references. Not affected.

### Static assets are a separate class and should stay shared

A large share of the hardcoded URLs are **read-only branding and seed assets that are identical in every environment** and should deliberately *not* be duplicated per env:

- `config/topics.php` — ~200 discovery topic card images under `lumotive-web-storage/discovery/quick-topics/`
- Default avatars and placeholders: `lumotive-web-storage/default/profile_default.svg`, `default/google-business.png`, `no-image-available-small.png`
- Frontend references the same set in ~15 files (`src/utils/platformVisuals.ts`, `src/components/layout/TopHeaderBar.vue`, the social-preview components, etc.) and `contentstudio-frontend` holds no upload-side bucket logic
- `SocialRepo.php` L1294 Facebook group default cover, `PlannerViewService.php` L533-534

These are fine on a single shared public-assets bucket. The story should scope env separation to **customer-written data only**, or the change balloons into re-hosting the entire static asset set per environment.

### Environment config today

`.env.example` documents `AWS_*` but **contains no `GCS_*` or `GOOGLE_CLOUD_*` entries at all**. Combined with the production bucket names being the config defaults, a freshly provisioned environment that forgets these vars silently points at production. That is the failure mode to close.

## What Needs to Change

1. **Provision env-scoped buckets** for QA, Staging, UAT alongside the existing production ones — nearline and standard tiers per env, plus an env-scoped equivalent of the legacy `lumotive-web-storage` write paths.
2. **Separate credentials per environment** — a service account per env whose IAM grants reach only that env's buckets, so a misconfigured non-prod deploy cannot read or delete production objects even if a bucket name leaks through.
3. **Make `cs_gcs_buckets` and `media_storage_url_prefix` derive from the configured disks** rather than being literal arrays, so the "is this ours?" check and CSV key derivation follow the environment.
4. **Replace the 185 hardcoded bucket literals** with config-derived values. The two patterns to build: a key-from-URL helper that strips whichever configured bucket base matches, and a disk-from-URL resolver that maps a URL to its disk by comparing against configured bucket URLs instead of literal names.
5. **Fail fast on missing config** — a non-production environment booting without explicit GCS bucket vars should refuse to start rather than default to production buckets. Remove the production bucket names as config defaults.
6. **Fix `social-inbox-manager`** — read the bucket from settings/env instead of the four hardcoded `bucket = "lumotive-web-storage"` literals.
7. **Document the vars** in `.env.example` (and `.env.testing.example`) so new environments get them.
8. **Decide the static-asset policy explicitly** — recommend keeping branding/seed assets on one shared read-only bucket, with non-prod holding read-only IAM on it.

## Files Involved

**Backend (`contentstudio-backend/`)**
- `config/filesystems.php` — disks, `cs_gcs_buckets`
- `config/app.php` L457 — `media_storage_url_prefix`
- `app/Libraries/Helper.php` L953 — `isMediaFromGcsStorage`
- `app/Libraries/Media.php` L503, L548-555
- `app/Libraries/Storage/MediaLibrary.php` L53, L72, L119, L216, L235
- `app/Libraries/MediaLibrary/GcsSignedUrlGenerator.php` L68
- `app/Http/Controllers/Storage/MediaLibrary/MediaLibraryAssetsController.php` (12 sites)
- `app/Http/Controllers/Storage/MediaController.php` L315, L422
- `app/Http/Controllers/Api/V1/PostController.php` L2549, L2766
- `app/Builders/CSVBuilder.php` L519, L1126, L1178
- `app/Builders/EvergreenBulkBuilder.php` L145
- `app/Repository/Utilities/MediaRepository.php` L1331
- `app/Repository/Integrations/Platforms/SocialRepo.php` L1294
- `app/Jobs/Storage/ImportDriveMediaJob.php`, `app/Jobs/Integrations/PlatformChangesJob.php`, `app/Jobs/Integrations/Platform/PlatformObserversJob.php`, `app/Jobs/Integrations/Platform/PlatformSavedJob.php`
- `app/Services/PlannerViewService.php` L533-534
- `app/Console/Commands/Integrations/AccountImageCommand.php`, `app/Console/Commands/WorkspaceSeedSampleFromXlsxCommand.php`
- `.env.example`, `.env.testing.example`

**Social inbox (`social-inbox-manager/`)**
- `app/utils/instagram_helper.py` L1357, L1503
- `app/utils/facebook_helper.py` L1153, L1289
- `app/utils/linkedin_helper.py` L907

**Not affected:** `contentstudio-frontend/` (static asset URLs only, no upload-side bucket logic), `contentstudio-flutter/` (test fixtures only), `contentstudio-social-analytics-go-main/`, `contentstudio-ai-agents/` (inherits backend).
