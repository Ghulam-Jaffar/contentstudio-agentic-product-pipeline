# Research: Adding Swedish and Czech

**Date:** 2026-09-13
**Deliverable:** `02-stories.md`

This doc stays local. Codebase paths, entry points, gotchas and open questions live here and never in a story body.

---

## 1. The ask

Add **Swedish** and **Czech** as supported languages, in two stories: one for the web app, one for the Flutter app covering both iOS and Android.

---

## 2. Backlog dedupe

Nothing in `docs/stories/` or `docs/features/` covers adding a language. The word appears across 190 story files only as the standard "Multilingual support" line in the global quality checklist, which is the reverse concern: those stories ask whether a feature is translatable, not whether a new locale exists.

---

## 3. Web: eight locales today, living in three separate places

This is the finding that matters most for scoping. A language is not one list.

| Layer | Where | What it covers |
|---|---|---|
| **Frontend** | `contentstudio-frontend/src/locales/<locale>/`, 20 namespace JSON files per locale | Every string in the app UI |
| **Laravel backend** | `contentstudio-backend/lang/<locale>/` plus a `<locale>.json` per locale | Emails and server-side strings |
| **Go report engine** | `contentstudio-social-analytics-go/src/services/reports/localization/` | Text inside generated PDF reports |

All three currently list the same eight: `en`, `fr`, `de`, `it`, `es`, `el`, `zh`, `pl`.

**A language added to the frontend alone would look supported and behave inconsistently:** the UI would be Swedish while emails and generated PDFs stayed English.

### Frontend, what has to change

- `src/locales/sv/` and `src/locales/cs/`, **20 JSON files each**. Current volume is roughly **25,000 lines per locale**, so this is the bulk of the work by a wide margin and it is translation effort rather than engineering.
- `SUPPORTED_LOCALES` in `src/i18n/index.ts`, currently a flat array of the eight codes. Used for the stored-locale check and for browser-language detection.
- `getSupportedLanguages()` in the same file, a hardcoded array of objects carrying `code`, `name`, `nativeName`, a flag emoji and a `countryCode`.
- `common.languages.<code>` in **every** locale's `common.json`, holding `name` and `native_name`. So adding two languages adds two entries to that block in all ten locale files, not just the two new ones.
- Namespaces are declared once in `LOCALE_NAMESPACES` and the loader maps over them, so no per-namespace registration is needed. English is eager, every other locale lazy-loads.

### A gotcha worth catching in review: language code is not country code

The flag comes from a `countryCode` field, and the existing data already proves they diverge: `en` uses `GB`, `el` uses `GR`, `zh` uses `CN`. For the two new ones:

- **Swedish** is language `sv`, country **`SE`**
- **Czech** is language `cs`, country **`CZ`**

Using `SV` or `CS` would render the wrong flag or none at all. `SV` is El Salvador.

### The real engineering gotcha: Czech pluralization

The i18n instance already registers a **custom plural rule for Polish**:

```ts
pluralRules: { pl: polishPluralRule },
```

Czech is in the same Slavic plural family and needs the same treatment, with distinct forms for one, few, many and other. **Swedish needs nothing**, since it uses the same two-form rule as English.

Without a Czech rule, any pluralized string silently resolves to the wrong form. Note the codebase barely uses pluralization today, so this will not be caught by casual testing.

### Consumers of the language list, beyond the switcher

`getSupportedLanguages()` feeds three surfaces:

1. The header language switcher.
2. The profile settings language preference.
3. **The analytics report language dropdown.**

That third one is a trap. Adding the two languages to the shared list immediately offers **Swedish and Czech PDF reports**, and the Go engine has its own separate list.

### Go report engine

```go
// supported lists the app's report languages. Anything else falls back to English.
var supported = map[string]bool{
	"en": true, "de": true, "es": true, "fr": true,
	"it": true, "pl": true, "el": true, "zh": true,
}
```

Good news: **the failure mode is graceful.** An unknown language falls back to English rather than erroring. So a report requested in Swedish before the Go side is updated produces an English PDF, not a broken one.

Two pieces to add there:

- **The message catalogs are generated from the frontend's own locale files.** There is a one-off generator with a documented invocation that reads the frontend locales directory and writes one flat file per language. So once the frontend translations exist, the Go catalogs are produced by running it rather than translated again. Its own `langs` list needs the two codes added.
- **`dates.go` holds hand-written localized month-name arrays per language.** These are not covered by the generator and need Swedish and Czech entries added by hand.

### Laravel

`lang/` holds a directory plus a JSON file per locale. These carry email copy, including the analytics report emails. Same two additions.

### One more thing worth knowing

Locale files **drift**. Polish is currently around 350 lines behind English across its namespaces, so the existing locales are not all complete. The two new locales should be complete at the point they are added, and it is worth deciding whether a completeness check belongs in CI, since nothing enforces it today.

---

## 4. Web: open questions

1. **Is the Go report engine and the Laravel email layer in scope for this story, or is a UI-only first pass acceptable?** Both fall back to English gracefully, so a phased release is viable. It just needs to be a decision rather than an oversight.
2. **Where do the translations come from?** Roughly 25,000 lines per locale is a translation-vendor question, not an engineering one, and it dominates the timeline.
3. **Should a locale-completeness check be added to CI**, given the existing drift?

---

## 5. Flutter: a much smaller change surface

The mobile app uses a **hand-rolled localizations delegate over bundled JSON assets**. Not ARB files, not a generated-code package, and **no codegen step at all**, so the story needs no generation task.

- Translations are `assets/i18n/<code>.json`, **one file per language**, named by bare language code.
- They are registered in the manifest as a **directory** asset, so dropping two new files in requires **no manifest edit**.
- Each file is roughly 80 to 124 KB and holds **1,557 leaf keys** across 23 modules. Every existing locale has exactly that count.

### One list drives everything

```dart
const List<CSLanguage> kSupportedLanguages = [
  CSLanguage('en', 'English', 'English', '🇬🇧'),
  ...
  CSLanguage('pl', 'Polish', 'Polski', '🇵🇱'),
];
```

`CSLanguage` carries code, English name, native name and a flag emoji. Every downstream surface derives from it: the app's supported locales, the settings language screen, the intro carousel's language sheet, the menu row that shows the current language, and the parity test. **There is no second hardcoded list to find.**

New entries would be `CSLanguage('sv', 'Swedish', 'Svenska', '🇸🇪')` and `CSLanguage('cs', 'Czech', 'Čeština', '🇨🇿')`. Display names are Dart constants rather than translated keys, so they do not go in the JSON.

### The complete edit list

1. The supported-languages list, two lines.
2. Two new JSON files, 1,557 keys each.
3. **The iOS `Info.plist` localizations array.** Its own comment explains why: the App Store "Languages" field is derived from the bundle's declared localizations, not from the JSON assets. This is the one platform file that must change.
4. Documentation that says "eight languages" in a few places.

**Android needs nothing.** There are no per-locale resource directories, no locale config, and no localized strings in Android resources.

**No store metadata lives in the repo.** Listing copy is managed by hand in the stores. One relevant note from the repo's own distribution doc: Play blocks a release if a listing language has an empty field, and there is a copy-to-all-languages control for when translations are not ready.

### Fallback behaviour is good

Each locale's JSON is **deep-merged over English**, so a missing key falls back to English per key rather than showing a blank or crashing. A key missing everywhere returns its own dot-path. The persisted locale falls back to English if it is not in the supported list.

### The picked language is also an API header

The persisted locale is sent as an `X-Locale` header on API requests. So choosing Swedish in the app flips the API locale too, and the backend needs to accept the two new codes. That is a genuine cross-repo link between the two stories.

---

## 6. Flutter: three things that will bite

### A test will fail on Czech, by design

There is a **layout-overflow guard** that iterates every supported locale and asserts that no locale's approval-workflow "save draft" label is longer than the English one. English is "Save as draft" at 13 characters. The natural Czech translation is about 19. **This test will fail** unless the Czech string is kept to 13 characters or the widget is reworked. Easy to miss, and it fails as an assertion rather than as anything that looks like a translation problem.

### The parity test is a hard gate, which is a good thing

A parity test loops the supported languages, flattens each file to dot-paths, and asserts the key set **exactly equals** English, with both missing and extra keys failing, plus that every placeholder in the English string survives in the translation.

So **partial translations cannot land**. Adding the two languages to the list with incomplete files fails the test suite immediately. This is worth stating as a sequencing constraint: the list entry and the complete file have to arrive together.

### Czech plurals will be approximate

**There is no plural engine.** String lookup does dot-path resolution and placeholder substitution, nothing more. Plurals are handled by hand-picking between a `_one` and an `_other` key at just three call sites. Beyond those, around **50 keys embed a raw count with a hardcoded English plural noun**, for example "{count} minutes ago" and "{count} posts".

Czech needs four forms. The current model can express two, and only where someone deliberately added the pair.

Two options:

- **Accept an approximate Czech translation** for those strings. Cheapest, and there is precedent: **Polish already ships with exactly this limitation** and has the same Slavic plural rules.
- **Add a real plural helper** and widen the key convention. Correct, but it touches all eight existing locale files and the parity test.

Recommend the first, matching the Polish precedent, and record it as a known limitation rather than discovering it in QA.

### A pre-existing gap worth noting, not fixing here

Date formatting is constructed **without a locale** everywhere, with English-centric patterns, and date formatting is never initialised for other locales. So dates render in US English regardless of the chosen language. Swedish and Czech users will see English month abbreviations and twelve-hour time. **This is pre-existing and affects all eight current locales**, so it is not caused by this work, but it will be more noticeable and is worth a follow-up.

### A minor naming footgun

Czech's language code is `cs`, which collides with the codebase's pervasive ContentStudio prefix. No functional conflict, since the code is only ever a data string, but searches for that prefix will be noisier afterwards.

---

## 7. Flutter: CI does not run the tests

The only workflow is an opt-in distribution job triggered by commit-message markers, and it runs the store lanes only. **It never runs the test suite or the analyzer.** So the parity test and the overflow guard, both of which directly protect this work, are enforced **only when someone runs them locally**.

Worth an explicit decision: add a job that runs the tests, or accept that the gate is manual.

---

## 8. Story shape

Two stories, as asked.

- **Web**, spanning the frontend locales and language list, the Laravel language files, and the Go report engine. Flagged so the last two can be phased if wanted, since both fall back to English gracefully.
- **Flutter**, covering both platforms from one codebase, with the iOS bundle declaration as the only platform-specific edit.

**No design story.** This adds no new interface, only two more rows to lists that already render exactly this shape on both platforms. Recorded as a deliberate decision rather than an oversight.
