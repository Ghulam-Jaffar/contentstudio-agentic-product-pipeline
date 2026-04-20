# Research — Social Platform Token Visibility & Expiry Gaps

## Background

Team lead flagged a concern about "platform observer" and "GBP token visibility." Investigation confirms the concern is real and broader than GBP alone — there are four distinct gaps in ContentStudio's token-lifecycle management across social platforms.

## Codebase findings

### 1. Platform observer coverage

The long-running daemon `php artisan social:observers` pops events off a Redis queue named `social_observers` and dispatches `PlatformSavedJob` to refresh account metadata, validate tokens, sync inbox, etc. Account models push to that queue from their `boot()` lifecycle hooks on `created` / `updated`.

| Platform | Pushes to `social_observers`? | Source |
|---|---|---|
| Facebook | Yes | `contentstudio-backend/app/Models/Integrations/Platforms/Social/FacebookAccounts.php:99` |
| Instagram | Yes | `InstagramAccounts.php:97` |
| LinkedIn | Yes | `LinkedinAccounts.php:99` |
| Pinterest | Yes | `PinterestAccounts.php:102` |
| Twitter | Yes | `TwitterAccounts.php:113` |
| **GMB / GBP** | **No** | `GoogleMyBusinessAccounts.php:83-94` — observer import commented out, boot only sets `is_migrated = false` |
| **YouTube** | **No** | `YoutubeAccounts.php:79` — push line is present but commented out |
| **TikTok** | **No** | No `rpush` in the model |
| **Threads** | **No** | No `rpush` in the model |
| **Bluesky** | **No** | No `rpush` in the model |
| **Tumblr** | **No** | No `rpush` in the model |

Relevant files:
- Daemon: `contentstudio-backend/app/Console/Commands/Integrations/SocialObserversCommand.php`
- Observer: `contentstudio-backend/app/Observers/Integrations/PlatformObserver.php`

### 2. Expiry email coverage

`AccountExpiryJob::checkSocialExpiry()` is the daily email flow. Its entire body:

```php
private function checkSocialExpiry(){
    $this->checkFacebookExpiry();
    $this->checkLinkedinExpiry();
}
```

So only Facebook and LinkedIn send "your account expired, please reconnect" emails (via the Lumotive Emails API → `social_expire` notification type). The other 9 platforms silently stop publishing.

Source: `contentstudio-backend/app/Jobs/Integrations/AccountExpiryJob.php:49-52`.

### 3. Token visibility in API

Every social account model has these token-tracking fields in its `$fillable` and default `$attributes`:

- `validity` (valid / expired / expiring_soon / invalid)
- `validity_error` (error message string)
- `validity_status` (HTTP status code)
- `invalid_tries` (int)
- `sent_invalid_email` (flag)
- `token_expires_at`
- `token_issued_at`
- `limit_exceed_tries`

But the account API response only includes `validity` and `token_expires_at`. The other fields are hidden from the frontend.

Relevant files:
- `contentstudio-backend/app/Repository/Integrations/Platforms/SocialRepo.php:19-41` — default fields list
- `contentstudio-backend/app/Http/Controllers/Integrations/Platforms/PlatformController.php` — `getSocialAccountsList()` required-fields list

### 4. Periodic token validation is paused

`SocialAccountsJob` is the per-account token validator — it calls each platform's token-debug / profile endpoint to verify validity and refresh `expires_at`. It handles notifications at 1 / 3 / 7 days before expiry.

Current state:

```php
public function handle()
{
    //job paused
    return false;   // ← kills the entire job
    // ... code below never runs ...
}
```

Source: `contentstudio-backend/app/Jobs/Integrations/Validate/SocialAccountsJob.php:48-51`.

Even before it was paused, the job only covered Facebook and LinkedIn. No other platforms had validation code.

**Why paused:** No commit history or comment explains the reason. Could be rate limits, an API change, cost, or abandoned mid-refactor. Worth asking the team (Bilal) before touching.

## Summary of gaps

| Platform | Pushes to `social_observers` | Expiry email | Token visibility in API | Periodic validation |
|---|---|---|---|---|
| Facebook | Yes | Yes | Partial | Yes (code exists, job paused) |
| LinkedIn | Yes | Yes | Partial | Yes (code exists, job paused) |
| Instagram | Yes | No | Partial | No |
| Twitter | Yes | No | Partial | No |
| Pinterest | Yes | No | Partial | No |
| GBP | No | No | Partial | No |
| YouTube | No | No | Partial | No |
| TikTok | No | No | Partial | No |
| Threads | No | No | Partial | No |
| Bluesky | No | No | Partial | No |
| Tumblr | No | No | Partial | No |

Only Facebook and LinkedIn are fully covered end-to-end (minus the paused validator). GBP is the worst offender (zero coverage across all four gaps).
