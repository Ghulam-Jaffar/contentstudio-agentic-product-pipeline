# Stories: Add Swedish and Czech

**Date:** 2026-09-13
**Stories:** 2
**Research:** `01-research.md`

---

## Overview

ContentStudio supports eight languages today: English, French, German, Italian, Spanish, Greek, Chinese and Polish. This adds **Swedish** and **Czech**, taking it to ten, on the web app and in the mobile app.

Two stories, one per platform. They are independent and can ship in either order, with one link between them: the mobile app sends the chosen language to the API, so the backend needs to accept the two new codes for the mobile side to behave correctly.

### No design story, deliberately

This introduces no new interface. Both platforms already render a language list with a flag, a native name and a check mark, and this adds two rows to each. Recorded here as a decision rather than an omission.

### The real cost is translation, not engineering

On both platforms the code change is small and well contained. What dominates is volume:

- **Web:** roughly 25,000 lines per locale across 20 files.
- **Mobile:** 1,557 keys per locale in a single file.

Where those translations come from, and who signs off on their quality, is the question that actually sets the timeline.

### Two things to settle first

1. **How far does "web" go?** A language added to the app interface alone leaves emails and generated PDF reports in English. Both fall back to English gracefully rather than breaking, so a phased release is viable, but it should be a decision. The web story is written to cover all three layers, with the split called out so it can be trimmed.
2. **Czech plurals will be approximate on mobile**, and to a lesser degree on web. Czech needs four plural forms and the mobile app has no plural engine at all. Polish already ships with exactly this limitation, so there is precedent for accepting it. See each story.

---

## Story 1

### Title

**[Full Stack] Add Swedish and Czech to the web app**

### Description

As a Swedish-speaking or Czech-speaking customer, I want ContentStudio in my own language, so that I can use the product without translating the interface in my head, and so that what it sends me reads the way the rest of it does.

The web app supports eight languages today. Adding two more is mostly a matter of supplying translations and registering the languages, because the loading, switching and storage machinery is already generic. What makes this more than a frontend change is that a language lives in three places: the app interface, the emails the backend sends, and the text inside generated PDF reports. Adding it to the interface alone produces a Swedish app that emails you in English.

---

### Workflow

1. User opens the language switcher in the header and sees Swedish and Czech alongside the existing eight, each with its flag and its own native name.
2. User picks Swedish. The interface switches without a reload, and the choice is remembered next time they sign in.
3. User can also set the same preference from their profile settings.
4. A user whose browser is set to Swedish and who has never chosen a language gets Swedish automatically on first visit.
5. Emails that ContentStudio sends that user arrive in Swedish.
6. When that user exports an analytics report and picks Swedish as the report language, the generated PDF is in Swedish.
7. Anything not yet translated falls back to English rather than showing a blank or a key name.

---

### Acceptance criteria

**The app interface**

- [ ] Swedish and Czech appear in the header language switcher, in the profile settings language preference, and anywhere else the supported-language list is used
- [ ] Each shows its native name, being "Svenska" and "Čeština", consistent with how the existing languages are shown
- [ ] **The correct flag is shown for each**, noting that the flag comes from a country code that differs from the language code: Swedish is language `sv` and country `SE`, Czech is language `cs` and country `CZ`
- [ ] Selecting either switches the interface immediately, without a page reload
- [ ] The choice persists across sessions, in the same way the existing languages do
- [ ] A browser set to Swedish or Czech, for a user who has not chosen a language, gets that language on first visit
- [ ] Both locales are complete, with every namespace translated, so nothing falls back to English at launch
- [ ] The language name entries for the two new languages are added to **every** locale, not only the two new ones, so a French user sees the Swedish option named in French
- [ ] Any string that is missing for any reason falls back to English rather than rendering blank or showing a key path

**Czech pluralization**

- [ ] Czech uses a plural rule with its distinct one, few, many and other forms, in the same way a custom rule is already registered for Polish
- [ ] Swedish uses the default two-form rule, matching English, and needs no special rule
- [ ] A pluralized string renders the correct Czech form for counts of one, two, five and a fractional value

**Emails**

- [ ] Emails sent to a Swedish or Czech user arrive in that language
- [ ] An email string with no translation falls back to English rather than failing to send

**Generated PDF reports**

- [ ] Swedish and Czech are selectable as the report language wherever a report language can be chosen
- [ ] A report generated in Swedish or Czech has its report text in that language
- [ ] Month names inside generated reports render in the chosen language, which requires adding entries to the hand-maintained month-name lists rather than relying on the generated catalogs
- [ ] A report language with no catalog still generates in English rather than failing, preserving the existing fallback

**Consistency**

- [ ] The language switcher, the profile preference and the report language dropdown all offer the same ten languages, with no surface left behind
- [ ] Switching language fires the existing language-change analytics event for the two new languages, exactly as it does for the current eight

---

### Mock-ups

N/A. No new interface. Both new languages appear as additional rows in lists that already exist.

---

### Impact on existing data

None. No schema change, no migration. A user's language preference is already a stored string, and the two new codes are simply values it can now hold.

One thing worth knowing for QA rather than for the build: the existing locales are **not all complete**. Polish currently trails English by a few hundred lines across its files. So "every locale is complete" is not true today, and the two new locales should be complete when they land even though their neighbours are not.

---

### Impact on other products

- **Mobile app:** the app sends the user's chosen language to the API, so once it offers Swedish and Czech the backend will start receiving those codes. The backend should accept them rather than rejecting or ignoring them. This is the one dependency between the two stories.
- **Chrome extension:** no impact.
- **Public API:** no impact.
- **White label:** customers on white-label domains get the same language list. Worth confirming that is intended rather than something a reseller should control.

---

### Dependencies

None on other stories. The real dependency is on the translations themselves being available.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Story 2

### Title

**[Flutter] Add Swedish and Czech to the mobile app**

### Description

As a Swedish-speaking or Czech-speaking customer using ContentStudio on my phone, I want the app in my own language, so that it matches the language I use everywhere else in the product.

The app supports eight languages today and the mechanism is already generic: one list of languages drives the app's supported locales, both language pickers, the menu row that shows the current language, and the test that guards translation completeness. Adding two languages is that list plus two translation files, with one platform file to update so the App Store lists the languages correctly. One change covers both iOS and Android.

---

### Workflow

1. User opens the menu and taps Language.
2. Swedish and Czech appear alongside the existing eight, each with its flag, its native name and its English name.
3. User taps Swedish. The app switches immediately and the choice is remembered.
4. A first-time user on the intro carousel can pick either language from the globe control before signing in.
5. The menu shows the chosen language's native name as the Language row's subtitle.
6. Anything not yet translated shows the English text rather than a blank or a key name.
7. The App Store lists Swedish and Czech among the app's languages.

---

### Acceptance criteria

**The languages**

- [ ] Swedish and Czech appear in the Language screen reached from the menu
- [ ] Both also appear in the language sheet on the intro carousel, which a user sees before signing in
- [ ] Each shows its flag, its native name as the primary label, and its English name as the secondary label, matching the existing rows
- [ ] Selecting either switches the app immediately and persists the choice
- [ ] The menu's Language row shows the selected language's native name
- [ ] The intro carousel's language control shows the two-letter code for the selection, which will read "SV" and "CS"
- [ ] Both translation files are **complete**, with exactly the same key set as English
- [ ] A string missing for any reason falls back to its English text, which the existing per-key fallback already provides

**Tests that must pass**

- [ ] The translation parity test passes for both new languages, meaning identical key sets to English with no missing and no extra keys, and every placeholder preserved
- [ ] **The approval-workflow layout guard passes.** That test asserts no language's "save as draft" label is longer than the English one, which is 13 characters. The natural Czech translation is longer and **will fail it**, so either the Czech string is kept within the limit or the widget is changed to accommodate a longer label. Decide which, rather than discovering it as a red test
- [ ] Adding a language with an incomplete translation file fails the test suite, which is existing behaviour and should stay that way

**Platform**

- [ ] The iOS bundle declares both new languages, so the App Store "Languages" field lists them. This is derived from the bundle declaration, not from the translation files
- [ ] No Android resource changes are needed, and none are made
- [ ] Both languages are verified on a physical or simulated device on **iOS and Android**, since one codebase serves both

**Known limitations, to be recorded rather than fixed**

- [ ] Czech plural forms are approximate. The app has no plural engine, and around fifty strings embed a count with a fixed plural noun. Czech needs four forms and at most two can be expressed. **Polish already ships with the same limitation**, so this matches existing behaviour rather than introducing a new defect. Record it rather than leaving it to be found in QA
- [ ] Swedish plurals are unaffected, since it uses the same two-form pattern as English
- [ ] Dates and times continue to render in US English formatting for these languages, as they already do for all eight existing languages, because date formatting is not localized anywhere in the app. Not introduced by this work and not fixed by it

**Backend**

- [ ] The app sends the selected language to the API, so the backend accepts the two new codes rather than rejecting or ignoring them

---

### UI copy

No new copy is authored for this story beyond the language names themselves, which are data rather than translated strings:

> **Swedish:** native name "Svenska", English name "Swedish", flag 🇸🇪
> **Czech:** native name "Čeština", English name "Czech", flag 🇨🇿

Everything else is the translation of existing strings.

---

### Mock-ups

N/A. No new interface. Both languages appear as additional rows in the two existing language lists.

---

### Impact on existing data

None. No schema change and no migration. The stored language preference is already a string, and an unrecognised value already falls back to English.

---

### Impact on other products

- **Web:** no impact, but the two should offer the same languages so a user does not find their language on one and not the other. Worth coordinating the release of the two stories even though they are technically independent.
- **Backend:** the app sends the chosen language on API requests, so the two new codes will start arriving.
- **App Store and Play Console:** the store listing languages are managed by hand outside the repository. Note that Play blocks a release when a listing language has an empty field, and offers a copy-to-all-languages control for when translated listing copy is not ready.

---

### Dependencies

None on other stories. The dependency is on the translations being available.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
