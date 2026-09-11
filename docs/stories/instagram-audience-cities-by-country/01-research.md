# Research: Instagram audience top cities, broken down by country

**Date:** 2026-09-11
**Deliverable:** `02-stories.md`

This doc stays local. Codebase paths, entry points, gotchas and open questions live here and never in a story body.

---

## 1. The ask

Instagram analytics has two audience demographics widgets: **Audience Top Countries** and **Audience Top Cities**. Both stay as they are. What changes is that the Cities widget gains a dropdown:

> **Audience top cities by [country]**

- The dropdown **defaults to the top audience country**, meaning the country ranked first in the Countries widget.
- Its options are **Overall**, plus the countries available in the Countries widget.
- **Only countries that have data appear in the list.**
- The user views **one country at a time**, or Overall.
- **The generated report shows whichever country is selected** in the dropdown.

The point is to let a user drill into a single country's cities rather than reading a single global list where one large market swamps everything else.

---

## 2. Backlog dedupe

Checked `docs/stories/` and `docs/features/`. **Nothing covers Instagram audience city or country demographics.** The near misses:

| Existing deliverable | Relationship |
|---|---|
| `docs/stories/youtube-demographics-analytics/` | Demographics for a different platform. Worth reading for the widget pattern it established, but no shared scope. |
| `docs/stories/analytics-report-sections-and-overview-cards/` | Lets users pick which sections appear in a report. Relevant because this feature also needs a per-widget choice to survive into a report, and that story is the closest precedent for report configuration. |
| `docs/stories/analytics-empty-state-screens/` | Covers analytics empty states, which matters here because Instagram withholds demographics for accounts below a follower threshold, so these widgets are empty more often than most. |
| `docs/stories/analytics-account-support-tooltips/` | The pattern for explaining why a widget has no data for a given account type. |
| `docs/stories/analytics-report-date-time-preferences/` | Unrelated in scope, but the same report-generation path this change has to reach. |

---

## 3. The constraint that shapes the whole feature

Instagram returns a **capped list of top cities and top countries**, on the order of the top 45 of each, computed globally across the account's followers. It does **not** return cities grouped by country, and it does not accept a country filter on a city breakdown.

That has three consequences worth settling before design:

1. **A country can appear in the Countries widget and still have no cities in our data.** If a country ranks eighth by followers but none of its cities make the global top 45, selecting it would show an empty widget. So "only countries that have data" has to mean **countries that have at least one city in the city list**, not simply countries present in the Countries widget. Those two sets are different, and the ask names the second one. This is the single most important thing to confirm.
2. **Whether we can attribute a city to a country at all depends on what the city records carry.** If a city record carries a country code, the grouping is free. If a city is only a label such as "Lahore, Punjab", then grouping by country needs a lookup we do not have today, and that changes the size of the work completely.
3. **Percentages need a defined base.** When a country is selected, is a city's share calculated against that country's followers, or against all followers? The first answers "how is my audience distributed within India", the second answers "how much of my total audience is in Mumbai". Both are defensible and they produce different numbers from the same data.

Findings on points 1 and 2 follow once the code sweeps report.

Sources consulted on the API's shape: [Instagram follower demographics overview](https://www.getphyllo.com/post/instagram-audience-demographics-for-influencer-marketing-platforms), [Graph API audience insights](https://www.getphyllo.com/post/how-to-use-the-instagram-graph-api-for-audience-insight-iv). The authoritative answer for ContentStudio is what our own ingestion stores, not what the API can theoretically return, so the code sweep decides it.

---

## 4. The blocking finding: Instagram does not tell us which country a city is in

This is the load-bearing fact for the whole feature, so it is worth stating precisely.

### What Instagram is asked for, and what comes back

ContentStudio uses the current `follower_demographics` metric, not the deprecated audience insights. The fetcher makes **fifteen separate calls**, three metrics crossed with five breakdowns:

```go
demographicMetrics := []struct {
    metrics    []string
    breakdowns []string
}{
    {metrics: []string{"follower_demographics"},          breakdowns: []string{"city", "country", "age", "gender", "age,gender"}},
    {metrics: []string{"engaged_audience_demographics"},  breakdowns: []string{"city", "country", "age", "gender", "age,gender"}},
    {metrics: []string{"reached_audience_demographics"},  breakdowns: []string{"city", "country", "age", "gender", "age,gender"}},
}
```

**`city` and `country` are two independent requests.** Nothing correlates them. The only combined breakdown Meta supports on this metric is `age,gender`, which is exactly why that is the only comma-joined value in the list. There is no `city,country` combination and no country filter parameter on the call.

For a city breakdown, each result carries a **one-element** `dimension_values` holding an opaque city label. The parser stores it verbatim:

```go
if len(dims) == 1 {
    // Single dimension (city, country, age, gender)
    dim := dims[0].(string)
    demographics = append(demographics, fmt.Sprintf("%s:%d", dim, value))
}
```

Meta's city labels for this metric are of the form **"City, Region"**, for example "Lahore, Punjab". The second part is a subdivision, **not a country**.

### What is stored

One wide table, `instagram_insights`, with each breakdown as a flat array of `"label:count"` strings:

```sql
`audience_city` Array(String),
`audience_country` Array(String),
```

There is **no country column on city data, no per-city row, and no city-to-country join anywhere.** City and country are two unrelated arrays on the same row. The API response mirrors that: both use the same `LocationCount {Name, Value}` struct, where a country `Name` is an ISO-2 code and a city `Name` is free text.

The read path compounds it. `GetLocations` reads **only the single latest row and ignores the requested date range entirely**, because demographics are a lifetime snapshot copied onto every day's row. So a country filter cannot be pushed down into SQL. Any filtering happens in memory after parsing, against whatever country attribution we invent.

### The conclusion

**The feature as described cannot be served from the data ContentStudio currently fetches or stores.** Selecting a country and seeing that country's cities requires knowing which country each city is in, and nothing in the pipeline knows that.

### Two pointed contrasts that shape the options

- **Facebook already has what Instagram lacks.** Meta's `page_fans_city` keys are `"City, Region, Country"`, with the country embedded in the label. So this exact feature is nearly free on Facebook and genuinely hard on Instagram. Worth knowing if the ask is really about audience geography rather than about Instagram specifically.
- **Meta Ads already implements this exact UX**, and is the pattern to copy for the API surface: the request takes a `country` parameter and the response carries an `AvailableCountries []string` list explicitly described as being for the region filter dropdown. It works because its ClickHouse table has a real `country LowCardinality(String)` column per row. Its Ads Insights API supports `breakdowns=region,country` together, which the organic Graph API does not.

---

## 5. Options for getting a country onto each city

### Option A, recommended: resolve the city label to a country at read time, scoped to the account's own countries

Take the city label, split it into city and region, and match it against a geographic reference table to get an ISO country code. Crucially, **restrict candidate matches to the countries that already appear in that account's Countries array.** That single constraint removes most of the ambiguity that makes city-name matching unreliable in general: "Punjab" is ambiguous between Pakistan and India in the abstract, but not for an account whose audience is in one and not the other.

Why read time rather than ingest time:

- The data volume is tiny. Instagram returns on the order of 45 cities, and the read path already loads a single row.
- No migration, no backfill, and no reprocessing of historical rows.
- A correction ships as a new reference dataset rather than a data repair.

Trade-offs to state plainly in the story: the attribution is **derived and approximate, not reported by Instagram**. Some labels will not resolve, so there must be a defined behaviour for them. Accuracy depends on a reference dataset somebody owns.

### Option B: resolve at ingest and store a country alongside each city

Add a parallel array or a combined encoding, and resolve during parsing. Cleaner reads, but it requires a schema change, a backfill for existing rows, and a reprocess whenever the reference data improves. For 45 values read from one row, this buys very little.

### Option C: narrow the feature to what the data supports

Keep the two widgets as they are and drop the country breakdown for Instagram, or build it for Facebook where the country is already in the label. Honest, cheap, and does not deliver what was asked.

### Option D: ask Meta for a combined breakdown

Not available. `city,country` is not a supported breakdown and there is no country filter on the metric.

---

## 6. The report cannot carry a per-widget choice today

The report definition lists widgets as a flat array of ids:

```go
// Widgets is the ordered list of widget ids to include, resolved against the
// widget catalog. Order is the render order.
Widgets []string `json:"widgets"`
```

The widget catalog entry has no options or parameters field either. **There is no per-widget option mechanism anywhere in report generation.**

Where a widget does have a user-facing choice today, the catalog **expands it into separate widget ids** instead. The Overview's "Top posts by ..." dropdown becomes three ids, one per metric, and Meta Ads does the same for its metric dropdown. That pattern does not work here: the option is a country, the set of countries differs per account, and it is unbounded, so it cannot be enumerated into fixed widget ids.

Also telling: **Meta Ads' own country filter does not reach its PDF report.** Its report adapter calls the demographics function with no country set, so the report is always unfiltered. So there is no precedent anywhere in the product for a per-widget option surviving into a report. This ask would establish one.

The cities widget in the catalog today:

```go
register(reports.WidgetCatalogEntry{
    ID: "ig_demographics_city", Title: "Audience Top Cities",
    Type: reports.WidgetChart, Chart: reports.ChartHorizontalBar,
    Platforms: []string{ig}, DataDeps: []string{"country_city"},
})
```

### Recommended shape

Add an **additive, report-level options map** keyed by widget id, for example a `widget_options` field carrying `{"ig_demographics_city": {"country": "PK"}}`.

Preferred over changing `Widgets` from a list of strings to a list of objects, because that field is already persisted in two MongoDB collections, `reports` and `schedule_reports`, and reshaping it is a breaking change to stored documents. An additive field is backward compatible in both directions, and the report generate handler uses a lenient JSON decoder, so an older service ignores an unknown field rather than failing.

Persistence is feasible: both collections are schemaless and the create path passes the whole request through. Note that `widgets` is deliberately written **outside** the model's fillable list, so a new option should follow the same direct-write convention rather than being added to fillable. The pieces to touch are the report generate request DTO, the report controller's store path, the definition builder in the reports client, and the schedule carry-through helper so a scheduled report keeps the selection on every run.

---

## 7. Existing-data risk worth knowing

Three different encodings for these arrays exist historically. The current Go writer produces `"label:count"`, the parser also accepts a JSON object form, and the legacy Laravel builder produced a `$`-delimited form that the Go parser does **not** handle. Rows still carrying the `$` encoding drop out silently rather than erroring. Any work in this area should confirm whether such rows are still being read for live accounts.

Separately, the parser does not know which breakdown it is parsing and **infers it from the string shape**: two characters means a country, a short string containing a hyphen means an age range, `M`, `F` or `U` means gender, and anything else falls through to city. A two-character city name would be misfiled as a country. This is pre-existing, but it is directly adjacent to the work and worth a look while someone is in there.

Finally, the public API proxy whitelists query parameters through a request DTO, so a new `country` parameter must be added there or it will be silently dropped. That DTO already carries a `breakdown` string used by Threads, which is a usable precedent.

---

## 8. Frontend: current state

### Both widgets are one component

`views/instagram/components/graphs/AudienceLocationChart.vue` renders **both** widgets, instantiated twice with a `type` prop:

```vue
<div class="grid grid-cols-1 2xl:grid-cols-2 gap-4">
  <AudienceLocationChart type="country" />
  <AudienceLocationChart type="city" />
</div>
```

It is a horizontal ECharts bar chart showing the top ten, and one of the few charts that still hand-builds its chart options rather than using the shared bar chart component, because the country axis renders flag images. The card title is assembled from a shared prefix plus "Countries" or "Cities", which is convenient: "Audience Top Cities by India" is a title change in one place.

**It is Instagram-specific.** Facebook has a forked near-duplicate of the same file, differing only in imports, translation namespace and an array-versus-object guard. So any change here has an obvious Facebook twin that is **not** automatically covered, and that needs an explicit in-or-out decision.

Other platforms: LinkedIn has its own followers-demographics widget that **already has an in-card dropdown**, YouTube has no cities, Meta Ads and Google Ads have country and region tables with a country filter, and TikTok, X, Pinterest, Threads and Bluesky have no audience-location widget at all.

### The payload confirms the backend finding

```ts
export interface InstagramLocationCount {
  name: string
  value: number
}

export interface InstagramCountryCityResponse {
  audience_city: InstagramLocationCount[]
  audience_country: InstagramLocationCount[]
}
```

A city carries **only a name and a value**. Country names are ISO-2 codes, fed straight to a flag image URL and an ISO country name map. There is **no city-to-country mapping anywhere in the frontend**, and no country-list dependency. The frontend cannot derive this, confirming it must come from the API.

### Meta Ads already implements this exact interface

This is the closest precedent and it is worth copying almost literally. Its demographics tab holds a module-scope country filter ref, sends it as an extra payload parameter, and reads the option list **off the response itself**:

```ts
const regionCountryQuery = useMetaAdsOverviewSectionQuery({
  section: 'demographicsRegionCountry',
  extraPayload: computed(() => ({
    breakdown: demographicsBreakdown.value,
    country: demographicsBreakdown.value === 'region' ? demographicsCountryFilter.value : '',
  })),
})
const availableCountries = computed<string[]>(
  () => (regionCountryQuery.data?.value?.available_countries || []) as string[]
)
```

Its dropdown already has an "All countries" first item, which is exactly the requested "Overall" option, and it labels options through a shared `useCountryDisplay` composable that normalizes ISO-2 codes and full names into a display name. That composable already lives in the shared analytics folder with two consumers, so a third is expected rather than intrusive.

Two things Meta Ads does **not** do that this feature must: its filter does not reach its PDF report, and it is a table rather than a chart.

### There is one complete precedent for a widget option reaching the report

The top-posts limit makes the full round trip, and it is the chain to follow:

1. The widget's dropdown emits the value on the event bus.
2. The analytics filter bar catches it into a local ref.
3. The export button stamps it onto every report-creating payload.
4. Laravel stores it on the report document.
5. When the PDF renders, it is read back off the document and fed into the section fetch.

The report payload type is deliberately open-ended, described in the code as "backend reads the whole request", so adding a field costs nothing type-wise. It simply has to be added at all four places that build a report payload: the export modal, the schedule modal, the send-by-email modal, and the read-back in the PDF composable.

### There are two live PDF paths, and both need answering

- **Path A, Go-native**, flag-gated and primary. The frontend sends a list of widget ids and the Go service composes the PDF from its widget catalog. The Instagram audience section owns `ig_demographics_country` and `ig_demographics_city`. This path does not currently accept any per-widget option.
- **Path B, Gotenberg**, the fallback that is still shipped. Laravel renders the Vue report route headlessly. It **reuses the same Vue components** with a module-scope `isReportView` flag that components use to hide interactive chrome, exactly as the LinkedIn demographics dropdown and the top-posts dropdown already do. Data is fetched client-side by the PDF composable, which is where a country parameter would have to be threaded.

Notably, **Path B does not read the widget list at all**, so section selection only affects the Go renderer. A country selection therefore has to be plumbed through both paths independently.

### Why the existing "expand options into fixed widgets" trick cannot be used

Both the Go catalog and the LinkedIn report solve "which option" by rendering one widget per option: three top-posts widget ids for three metrics, and four LinkedIn demographic pages for four fixed types. That works because the option set is small, fixed and known at build time. **A country list is user data, differs per account and is unbounded**, so it cannot be enumerated into widget ids. This is precisely why the report needs a real per-widget option rather than more widget ids.

### Empty states, and a gap worth fixing while nearby

The widget already has a well-handled empty state, including a specific message when the account is under the follower threshold: *"You'll need 100 followers (who aren't your friends) to view this demographic information."* That threshold branch depends on the summary section having loaded, which is why the demographics section fires its own summary query.

Two things the new dropdown must handle:

- **When the country list is empty, which is the common case for small accounts, the dropdown has no options and no default.** The likely answer is to hide the dropdown entirely and keep the existing empty card, mirroring how the posts section gates its own dropdown on having loaded data.
- **The widget currently cannot say "this failed".** No error is passed to the chart frame, so a failed request renders as empty rather than as an error with a retry. The frame supports an error state; this widget just does not use it. Adding a dropdown that triggers refetches makes that gap more visible, since a failed country switch would silently look like "this country has no cities".

### Files a change would touch

The chart component, the Instagram analytics composable for the new selection ref and payload, the demographics section that owns the query, the response types, the filter bar and export button for the report payload, the three report modals, the PDF composable for the Gotenberg path, the shared country display composable for labels, translation files across eight locales, and a decision on the forked Facebook twin.
