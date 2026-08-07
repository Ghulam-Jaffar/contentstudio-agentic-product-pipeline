# Platform content overrides: documentation and surface parity

## Problem

`POST /api/v1/workspaces/{workspace_id}/posts` accepts an `overrides` block that lets a caller set a different caption and different media per platform. It is live in production and it works. It is also effectively invisible.

The field shipped without its OpenAPI annotation, so it does not appear in the generated `api-docs.json`, the Swagger UI, or the interactive reference at `api.contentstudio.io/guide`, even though the field is validated and handled in production. A developer checking the published reference concludes the endpoint supports only one caption for all platforms. That already happened internally: our own team reported per-platform captions as unavailable while the feature was running in production.

Every surface that wraps post creation inherits the same blind spot. The CLI, the agent skill, and the Zapier, Make and n8n apps all expose a single caption only.

This matters beyond documentation hygiene. Customising a caption per platform is one of the highest-value things the API does, and it is the difference between one identical post blasted everywhere and a genuinely useful integration. Right now nobody can find it.

## Verified current state

| Surface | Supports `overrides` |
|---|---|
| Production API, `origin/master` | Yes, working |
| `docs/api/posts-endpoint.md` | Yes, with example and validation rules |
| `@OA` annotations on `PostController::store` | **No** |
| Generated `api-docs.json`, Swagger UI, `api.contentstudio.io/guide` | **No** |
| `contentstudio-cli` post creation | Needs confirming, repo not available locally |
| `contentstudio-agent` skill | Needs confirming, repo not available locally |
| Zapier app | Needs confirming |
| Make app | Needs confirming |
| n8n app | Needs confirming |
| Help docs, article 1163 | Needs confirming |

## How overrides work today

```json
{
  "content": { "text": "Default caption used for every account" },
  "accounts": ["acc_1", "acc_2", "acc_3"],
  "overrides": {
    "linkedin":  { "content": { "text": "Longer professional version" } },
    "twitter":   { "content": { "text": "Short punchy version" } },
    "instagram": { "content": { "text": "Visual caption with hashtags",
                                "media": { "images": ["https://..."] } } }
  }
}
```

- `content.text` is the default, applied to any platform without an override.
- Supported override keys: facebook, instagram, twitter, linkedin, pinterest, youtube, tiktok, gmb, tumblr, threads, bluesky, telegram.
- An override may carry `content.text`, `content.media.images`, `content.media.video`, or a combination.
- Override text is capped at 5000 characters. Images cap at 10. Images and video cannot both be set in the same override.
- Presence of any override sets `common_box_status` to false internally, which is the same mechanism the in-app composer's per-platform toggle uses. The API and the composer share one data model.
- **Overrides are keyed by platform, not by account.** Three Facebook pages all receive the same Facebook override.

## Out of scope

- Per-account caption overrides. Only platform-level exists and changing that is a separate data-model conversation.
- Any change to override behaviour itself. This work documents and propagates what already ships.
- **MCP servers.** Both the hosted `/api/mcp` server and the npm `contentstudio-mcp` package are also missing override support. Confirmed missing, deliberately excluded here, and worth tracking separately so it is not lost.
- The AI chat scheduling flow, which also collapses per-platform captions into one before creating the post. Same reason.

---

## [BE] Add platform content overrides to the API reference and integration surfaces

### Description

The `overrides` request field on the create-post endpoint has no OpenAPI annotation, so the published API reference documents an endpoint that appears to support only one caption for every platform. The CLI, the agent skill, and the Zapier, Make and n8n apps all expose a single caption for the same reason.

This story annotates the field so the reference matches the implementation, then brings each integration surface to parity. No change to override behaviour itself.

Annotate first. The rest of the work should be built against a correct documented contract, and the annotation is what is actively misleading people today.

### Workflow

**API reference.** A developer opens the ContentStudio API reference and looks at the create-post endpoint. Alongside `content`, `accounts` and `scheduling`, they see an `overrides` object. Expanding it shows one entry per supported platform, each accepting its own `content` with `text` and `media`. The description explains that `content.text` is the default and an override replaces it for that platform only. The example request includes a populated `overrides` block they can copy and send. They try it in the interactive console and get back a created post with different captions per platform.

**CLI.** A user creates a post with per-platform captions, supplied from a JSON payload file rather than twelve flags. `--dry-run` shows the resolved caption for each platform before anything is sent, so they can check the LinkedIn version is the long one. `--json` output reports which platforms received an override.

**Agent skill.** A user asks their agent to publish an update with a different caption on LinkedIn. The skill tells the agent that per-platform captions exist, so the agent uses them instead of sending one caption everywhere.

**Zapier, Make and n8n.** A user building a post-creation step sees optional per-platform caption fields alongside the main caption. They leave them blank for a single caption everywhere, or fill in the platforms they want to differ. The field labels make clear that a blank platform field falls back to the main caption, and that a platform override applies to every account on that platform.

### Acceptance criteria

**API reference**

- [ ] The `overrides` field appears in the create-post request schema in `api-docs.json` after regenerating the spec
- [ ] All 12 supported platform keys are documented: facebook, instagram, twitter, linkedin, pinterest, youtube, tiktok, gmb, tumblr, threads, bluesky, telegram
- [ ] Each platform override documents `content.text`, `content.media.images` and `content.media.video`
- [ ] The field description states that top-level `content.text` is the default and an override replaces it for that platform only
- [ ] Documented constraints match the validation in force: text max 5000 characters, images max 10, images and video mutually exclusive within one override
- [ ] The description states that overrides apply per platform, not per account, so multiple accounts on the same platform share one override
- [ ] The endpoint example request includes a populated `overrides` block with at least two platforms
- [ ] The field is shown as optional
- [ ] Swagger UI and `api.contentstudio.io/guide` both render the field after the spec is regenerated
- [ ] Sending the documented example request succeeds and produces per-platform captions on the created post
- [ ] `docs/api/posts-endpoint.md` and `docs/api/posts-quick-reference.md` are reviewed and corrected if they disagree with the annotation

**CLI and agent skill**

- [ ] Current override support in `contentstudio-cli` and `contentstudio-agent` is confirmed and recorded before changes begin
- [ ] CLI post creation accepts per-platform captions and media
- [ ] Overrides can be supplied via a JSON payload file, not flags alone
- [ ] `--dry-run` shows the resolved caption per platform before sending
- [ ] `--json` output follows the existing `{ok, data}` envelope and reports which platforms received an override
- [ ] CLI help text documents the option with a working example
- [ ] The agent skill describes per-platform captions so agents use the capability
- [ ] An invalid platform key produces an error naming the supported platforms, not a generic failure

**Zapier, Make and n8n apps**

- [ ] Current override support in each of the three apps is confirmed and recorded before changes begin
- [ ] The post-creation action in each app exposes optional per-platform caption fields
- [ ] Per-platform media is supported where the app's field model allows it; where it does not, the limitation is recorded
- [ ] Field labels state that leaving a platform field blank falls back to the main caption
- [ ] Field labels or help text state that a platform override applies to all accounts on that platform
- [ ] All 12 platforms are available, or the supported subset is documented with a reason
- [ ] An existing published scenario or zap that sends no overrides continues to work unchanged

**Across all surfaces**

- [ ] Calls that send no `overrides` behave exactly as before on every surface
- [ ] Help centre articles covering the API, CLI and skill are updated where they describe post creation
- [ ] The override contract is identical across surfaces, so a user gets the same capability regardless of which one they use

### Impact on existing data

None. `overrides` is additive and optional everywhere. No schema change, no behaviour change to existing calls.

### Impact on other products

The in-app composer already writes the same `common_box_status` and per-platform sharing details, so there is no divergence to reconcile there.

The MCP servers and the AI chat scheduling flow remain without override support after this work. Both are listed as out of scope above and should be tracked separately.

### Dependencies

None blocking. The annotation work should complete before the CLI, skill and no-code app work begins, so those are built against a correct contract.

### Global quality checklist

- [ ] Mobile responsiveness — N/A, backend, CLI and third-party app configuration
- [ ] Multilingual support (error messages localised where the surface supports it; API reference and CLI output are English only)
- [ ] UI theming support — N/A, ContentStudio has no dark mode
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references

- Validation rules: `contentstudio-backend/app/Http/Requests/Api/V1/PostStoreRequest.php`, the `overrides` and `overrides.*` rules
- Handling: `contentstudio-backend/app/Http/Controllers/Api/V1/PostController.php`, where `common_box_status` is derived from override presence and platform sharing details are populated
- Regenerate the spec with `vendor/bin/sail artisan l5-swagger:generate`
- External repos: `contentstudio-cli` on npm, `github.com/contentstudioio/contentstudio-agent`
- Help centre article 1163 covers the public API; confirm which articles cover the CLI and the no-code apps
- The API v1.9 rollout carried Zapier, Make and n8n as a follow-on surface; reuse that pattern
