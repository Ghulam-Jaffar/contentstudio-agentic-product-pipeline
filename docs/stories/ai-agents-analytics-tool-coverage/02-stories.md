# Stories: Competitor analytics and full analytics coverage for the assistant

---

## [BE] Give the assistant competitor analytics and close the analytics tool coverage gap

### Description:

As a customer asking ContentStudio's AI assistant about performance, I want it to reach everything
the analytics product actually measures, so that it answers competitor, ads, Threads, Bluesky,
demographics and campaign questions instead of claiming ContentStudio cannot do things it has done
for months.

The assistant's analytics knowledge was built against a snapshot of the analytics API taken at the
end of July. Since then competitor analytics, Meta Ads, Google Ads, Threads and Bluesky have all
shipped, and none of them exist as far as the assistant is concerned. Counting against what the
analytics API serves today, the assistant can reach roughly a quarter of it. The rest is not
missing data, it is data the assistant has no way to ask for.

The cost is worse than silence. When the assistant has no way to make a request, it does not say
so. It tells the user the product cannot do it. That has already happened once in production, where
the assistant told users X had no top-performing-posts view while the product was showing exactly
that screen. Every gap below is another chance to repeat it, and competitor analytics is the most
expensive one to get wrong because it is a paid, differentiating part of the product that customers
actively ask about.

This story brings the assistant's analytics reach up to what the product measures, with competitor
analytics as the headline capability.

---

### Workflow:

**Competitor questions**

1. A user with competitor tracking set up asks the assistant "how do we compare to our competitors
   this month?"
2. The assistant finds the competitor sets the workspace has saved, and if there is more than one,
   asks which to use rather than guessing.
3. It reads the comparison for that set and answers with the customer's numbers next to each
   competitor's, naming the period and the accounts included.
4. The user follows up with "who posts more?", "what type of content do they post?", "which
   hashtags are they using?" or "show me their best posts" and gets each answered from the same
   competitor set.
5. A user whose competitor set has no data yet is told the set exists but has not been filled in
   yet, never that their competitors scored zero.
6. A user asks about competitors on LinkedIn or TikTok and is told competitor tracking covers
   Facebook and Instagram, rather than being handed an empty comparison.

**Everything else the assistant could not reach**

1. A user asks "who follows us, what age and what country?" and gets the audience breakdown for
   the networks that report it, instead of being told ContentStudio does not track that.
2. A user asks "how did the summer campaign do?" and gets performance for that campaign or label,
   which the assistant can already see by name but could not read numbers against.
3. A user asks "what were our worst posts?" or "show me top posts by comments rather than
   engagement" and gets the ranking they asked for.
4. A user asks about Threads or Bluesky and gets real numbers, the same way they do for the
   networks that were connected first.
5. A user asks "what did we spend on ads and what did we get back?" and gets campaign, ad set and
   ad performance for their connected ad accounts.
6. A Google Business Profile user asks "what are people searching to find us?" or "how are our
   reviews doing?" and gets an answer.
7. A user asks about posting cadence, format mix, hashtags or the formats specific to a network,
   such as Reels, Stories or pins, and gets an answer rather than a deflection.
8. When something genuinely is not measured for a network, the assistant says that specific thing
   is not available for that network, and never generalises it into a product limitation.

---

### Acceptance criteria:

**Competitor analytics**

- [ ] The assistant can list the competitor sets saved in the workspace, and when more than one exists it asks which to use rather than picking one
- [ ] The assistant can answer a head-to-head comparison between the customer's account and the competitors in a chosen set, for Facebook and for Instagram
- [ ] The assistant can answer how often competitors post, what content types they post, and how that compares to the customer's own posting
- [ ] The assistant can answer which hashtags competitors use and how a named hashtag performs for them
- [ ] The assistant can return a competitor's best and worst performing posts
- [ ] The assistant can answer how competitor follower counts have moved over the period, and how engagement on their posts has moved over time
- [ ] The assistant can read a competitor's profile description for context when the user asks who a competitor is
- [ ] A competitor set that has no data yet produces an explanation that the data has not been collected yet, never a zero or an empty comparison presented as a result
- [ ] Asking for competitors on any network other than Facebook or Instagram returns a clear statement that competitor tracking covers those two networks, rather than an empty result
- [ ] A competitor set belonging to another workspace is never readable, and asking for one returns a not-found explanation rather than an error the user sees
- [ ] Competitor answers state the period they cover and which competitors are included, the same way existing analytics answers state their coverage

**Coverage of the rest of the analytics surface**

- [ ] The assistant's analytics knowledge is rebuilt from the current analytics API rather than the July snapshot, and a check fails the build if the two drift apart again
- [ ] The assistant can answer audience composition questions — age, gender, country and city — for every network that reports them
- [ ] The assistant can answer performance questions about a named campaign or label, which it can already see by name
- [ ] The assistant can return worst-performing content, and can rank content by a metric the user names rather than only the default ranking
- [ ] The assistant can answer questions about Threads and about Bluesky, including their own trends, top content and audience figures
- [ ] The assistant can answer ads questions for connected Meta Ads and Google Ads accounts, covering spend, results, and performance by campaign, ad set or group, and ad
- [ ] The assistant can answer Google Business Profile questions about reviews, the searches that surfaced the listing, customer actions and media activity
- [ ] The assistant can answer posting cadence and format-mix questions for every network that reports them
- [ ] The assistant can answer hashtag performance questions for every network that reports them
- [ ] The assistant can answer questions about network-specific formats, including Reels, Stories, video performance and pin performance
- [ ] The assistant can answer how many X credits have been used
- [ ] Where two versions of the same trend exist, one cumulative and one daily change, the assistant answers with one of them consistently and the decision not to expose the other is recorded rather than left as a silent gap

**Honesty and safety**

- [ ] When a network genuinely does not report something, the answer names that network and that metric, and never states or implies that ContentStudio does not offer the capability at all
- [ ] No capability claim in the assistant's own tool descriptions asserts that a network lacks something the analytics API provides, and a check fails the build if one does
- [ ] Every new answer states which accounts it covers and which period, and reports any account whose data failed to load, matching how existing analytics answers already behave
- [ ] Adding these capabilities does not push any single assistant capability past its tool limit
- [ ] Every new read is restricted to the workspace the user is asking from
- [ ] Answers remain available in every language the assistant already supports

---

### Mock-ups:

N/A, backend only. The assistant's existing answer formats are reused.

---

### Impact on existing data:

None. Every capability in this story is read-only. No new records are created and no existing
record changes shape.

---

### Impact on other products:

- **Mobile apps:** the mobile AI assistant asks the same assistant, so these answers become
  available there with no mobile work in this story.
- **Web:** AI chat gains the new answers with no frontend change, unless a new answer shape needs a
  matching display, which is handled by the assistant's existing rendering.
- **Chrome extension:** not affected.
- **Public API:** not affected. This story consumes analytics endpoints that already exist and adds
  none.

---

### Dependencies:

- Competitor reads and competitor set management are already available on the public API, delivered
  by **[BE] Add competitor search and competitor report management to the public API** and
  **[BE] Add competitor comparison reads to the public API**. No further API work is required for
  the competitor half of this story.
- Single post lookups and single post performance are deliberately excluded here because they are
  covered by **[BE] Add post detail and post performance reads for the assistant**. If that story
  has not landed, this one does not add them.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
