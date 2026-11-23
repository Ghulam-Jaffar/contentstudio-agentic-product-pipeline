# ChatGPT Business Hub: ContentStudio profile

ContentStudio's profile in ChatGPT's Business Hub (chatgpt.com/business-hub), created 2026-11-12 from the **D4 Interactive** ChatGPT workspace. Profile ID `bizprof_6d3aef2ac1b881909671239501c461be`.

## Status (2026-11-12)

| Item | State |
|---|---|
| Profile (name, website, category SaaS company, support email) | Done, approved, verified badge |
| Description | Done. Saved with the 30-day money-back line for first-time subscribers |
| Delivery / returns | "Delivery unavailable" / "Free returns within 30 days" |
| Logo | First upload (`logo-1024.png`, a modified mark) rejected: "may violate our policy on Software & CRM". Replaced with the official app icon (`logo-official-512.png`), saved and live |
| Cover photo | Requested from design (1200 x 400 px) |
| Products | **Blocked.** Every row rejected: "The domain in url was not verified for this profile when this row was checked." The file matches OpenAI's spec. The `openai-business-profile` TXT record is missing from contentstudio.io DNS, the likely cause. An empty feed "ContentStudio plans" (`bpfeed_38774034126b50e982c924896627f00c`) exists |
| Agent builder | Not available: "rolling out to different regions". `agent-guidelines.md` is ready for when it is |
| Insights (7 days to Nov 11) | 49,169 mentions (down 28.5%), 262 website clicks |

## To unblock products

Tried on our side, all still failing (2026-11-12): file checked against the spec, a one-row test with just `https://contentstudio.io`, retries with the `openai-business-profile` TXT record live (new value `lImu5JG1...`), and an upload through the existing feed's **Update products**.

**OpenAI support contacted 2026-11-12** (help.openai.com chat, D4 Interactive account) with the profile and feed IDs. When they confirm a fix, re-upload `products.csv` via Products → ContentStudio plans → ⋯ → Update products → Upload a new file.

Standard product uploads only reach US users for now (OpenAI spec: "Standard uploads use `["US"]`"), even though ContentStudio sells worldwide.

## Files in this folder

| File | Use |
|---|---|
| `business-info-review.md` | Every Business info field, as reviewed |
| `logo-official-512.png` | Official app icon, current logo upload |
| `logo-1024.png` | Modified logo. Rejected by OpenAI, do not use |
| `cover-1600x900.png`, `cover-source.html` | Rough cover draft for direction only. Real cover comes from design |
| `products.csv` | Product feed: 4 plans x 3 billing cycles (12 rows), per https://developers.openai.com/commerce/specs/file-upload/products |
| `agent-guidelines.md` | Guidelines and business context for the Agent builder |
| `slack-message.md` | Team update |

**`products.csv` needs an owner.** If prices change on the website and not here, ChatGPT shows stale prices.

## Website issue found while doing this

The website states the network count inconsistently. The app (`contentstudio-frontend/src/config/platformConfigs.ts`) and `public/pricing.md` have **12** publishing networks, including Telegram and Tumblr. These still say **11**, with different lists:

- `public/llms.txt` (lines 3 and 5, no Tumblr)
- `constants/reviews-page.ts`, `constants/compare-pages.ts`, `constants/upgrade-from.ts`, `constants/about-us/index.ts`
- `constants/hootsuite-alternative-ppc.ts` and `constants/hootsuite-alternative-ppc-evolution.ts` ("eleven networks")
- `constants/cs-api/index.ts` (the API page, no Tumblr). Check whether the public API supports Tumblr before changing this one
