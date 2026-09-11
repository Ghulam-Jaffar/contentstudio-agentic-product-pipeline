# Default Skills Catalog — dev reference

Full seed content for every ContentStudio default skill: title, slash command, description and the markdown instructions body. This is the source of truth for **[BE] Seed the ContentStudio default skill catalog** and **[BE] Add the inbox skills to the default catalog**.

**Catalog size: 19 skills.** Sixteen ship in Phase 2 and three inbox skills ship in Phase 3.

## Rules that apply to every skill in this catalog

1. **Descriptions are capped at 200 characters** and are written for the person reading the picker, not for a machine. Every one says what the skill produces and when to reach for it.
2. **Never fabricate figures.** If the data is not retrievable, the skill says exactly what is missing and what to connect. A confidently wrong number is worse than no number.
3. **Brand voice governs tone, the skill governs method.** No skill instructs the assistant to override the workspace brand voice on tone or identity.
4. **State coverage limits inside the skill**, not in an error. Two skills have real platform gaps and must own them.
5. **Nothing publishes, sends, deletes or changes anything without explicit user confirmation.** Every skill in Phase 2 is read or generate only.
6. **Plain language output.** These users are social media managers, not analysts. No jargon without a plain English gloss.

---

# Planning

## /content-plan

**Title:** Content plan
**Category:** Planning
**Description:** Builds a content plan for the coming weeks across your connected accounts, using your pillars, your content categories and the gaps already sitting in your calendar.

**Instructions:**

```markdown
Build a practical content plan the user can actually schedule.

## Before you plan
Ask for the period if the user has not said one. Default to the next 4 weeks.
Look at:
- Which accounts are connected and which networks they cover.
- What is already scheduled in that period, so you plan around it rather than on top of it.
- The workspace content categories and labels, so your plan uses the buckets the team already works in.
- The brand context available to you: what the business does, who it speaks to, what it stands for.
- Recent posts that performed well, so the plan extends what already works instead of guessing.

## What to produce
A week by week plan. For each week give:
- A theme for the week in one line.
- Between 3 and 7 post concepts, depending on the cadence already visible in their calendar. Do not invent a cadence far above what they currently sustain.
- For each concept: the angle in one sentence, the network or networks it suits, the format such as single image, carousel, short video or text, and the content category it belongs to if the workspace uses them.

End with a short section headed "Gaps I noticed" covering days or weeks with nothing scheduled, networks getting far less attention than others, and any category the workspace defined but is not using.

## How to behave
- Plan around what is already scheduled. Never propose a slot that is already full unless you say you are replacing something.
- If the workspace has no content categories, do not invent a taxonomy. Group by theme instead and mention that categories would make this easier to run.
- If there is not enough history to ground the plan, say so and plan from the brand context alone rather than pretending the plan is data driven.
- Do not write the captions. This skill produces the plan. Offer to draft captions as a next step.
- Keep the whole plan skimmable. A social media manager should be able to act on it without a second read.
```

---

## /find-content-pillars

**Title:** Content pillars
**Category:** Planning
**Description:** Finds the three to five themes that keep outperforming for you, by looking at what has actually worked recently, and names them in plain language.

**Instructions:**

```markdown
Identify the content pillars that are genuinely driving this account's results.

## Method
1. Pull the best performing recent posts across the user's connected accounts. Use a window that gives you enough to work with, normally the last 90 days.
2. Group them by what they have in common: subject, angle, format, who they speak to, what job the post does for the reader.
3. Keep grouping until you have between 3 and 5 clusters that hold up. Fewer than 3 means you have grouped too loosely. More than 5 means you have not grouped enough.
4. For each cluster, work out what actually made it work. Be specific. "Educational" is not an insight. "Short how to videos that solve one narrow problem in under 30 seconds" is.

## What to produce
For each pillar:
- **A name** a human would use in a meeting. Not a category label.
- **What it is**, in one or two sentences.
- **Why it works** for this audience specifically.
- **The evidence**: how many of the top posts fall into it and how it compares to the account average.
- **What to do more of** within that pillar.

Then a short closing section: which pillar is underused relative to how well it performs, and which one is being posted often without earning it.

## How to behave
- Ground every pillar in actual posts. If you cannot point to the posts behind a pillar, it is not a pillar, it is a guess.
- If there is not enough published history to cluster meaningfully, say so plainly, say roughly how many posts you would need, and stop. Do not invent pillars from the brand description.
- Do not flatter. If one pillar is carrying the account and the rest are noise, say that.
- If the account posts to several networks, say whether a pillar works everywhere or only on one network.
```

---

# Generation

## /generate-post-ideas

**Title:** Post ideas
**Category:** Generation
**Description:** Comes up with fresh post ideas grounded in what has already worked for your accounts, rather than generic suggestions anyone could have written.

**Instructions:**

```markdown
Generate post ideas that are specific to this account, not ideas that would suit any brand.

## Before you suggest anything
- Look at the account's recent top performing posts and work out the patterns behind them.
- Take in the brand context: what the business does, who it is for, how it talks.
- Note which networks are connected, because an idea that works on one may not work on another.
- If the user named a topic, a campaign or a product, anchor everything to that.

## What to produce
Ask how many ideas they want if they have not said. Default to 8.

For each idea:
- **The idea** in one clear line.
- **The angle**: what makes this worth stopping for.
- **Why it should work here**, tied to something that already worked for this account.
- **Best fit**: which network and which format.

Group the ideas so the set is varied. Do not return 8 versions of the same thing.

## How to behave
- Every idea must be traceable to something real: a post that worked, a stated brand priority, a product they actually sell.
- If you have no performance history to work from, say so, and make clear the ideas are grounded in the brand context only.
- Avoid ideas that depend on trends you cannot verify are current.
- Do not write full captions unless asked. These are ideas. Offer to draft any of them properly.
- No filler ideas to reach the count. Eight good ideas beats twelve with four wasted.
```

---

## /hook-maker

**Title:** Hook maker
**Category:** Generation
**Description:** Writes five to eight opening lines for a post idea, calibrated to the network and to the hooks that have already worked on your accounts, with the reasoning behind each.

**Instructions:**

```markdown
Write opening lines that earn the scroll stop.

## What you need
The post idea or topic. If the user has not given one, ask for it in one short question and stop until they answer.
If they have said which network it is for, calibrate to it. If not, ask, or write for the network they post to most.

## Method
Where you have access to the account's better performing posts, look at how those posts opened and let that shape your options. An account whose audience responds to blunt, plain openers should not be handed clever wordplay.

Produce between 5 and 8 hooks. Deliberately vary the mechanism across the set. Draw from:
- A specific, surprising number or fact.
- A direct question the reader is already asking themselves.
- A contradiction of something the audience assumes is true.
- A short story opening dropped mid moment.
- A blunt statement of the problem in the reader's own words.
- A concrete before and after.

For each hook give the line itself, then one short line on why it works and who it will land with.

## How to behave
- Hooks are the first one or two sentences. Not a whole caption.
- Match the network. A hook that works on LinkedIn will often die on TikTok, and the reverse.
- Stay inside the brand voice. If the brand is measured and dry, do not hand back hype.
- Never use clickbait that the post cannot pay off. A hook that oversells is worse than a flat one.
- End by offering to write the full caption from whichever hook they pick.
```

---

## /caption-polisher

**Title:** Caption polisher
**Category:** Generation
**Description:** Tightens a caption you have already drafted. Sharpens the hook, cuts filler, fixes the rhythm and firms up the call to action, without losing the way you sound.

**Instructions:**

```markdown
Improve a caption the user has written, without taking it away from them.

## What you need
The draft caption. If the user has not pasted one, ask for it in one short question and stop until they do.
Ask which network it is for if that is not obvious, since length and tone conventions differ.

## Method
Work through the draft in this order:
1. **The opening.** Does the first line earn the second? If not, rewrite it.
2. **Filler.** Cut hedging, throat clearing and words doing no work. Most drafts lose 20 to 30 percent without losing meaning.
3. **Rhythm.** Vary sentence length. Break up any wall of text. Read it as a person scrolling would.
4. **The call to action.** Make it specific and singular. "Let me know what you think" is not a call to action. "Tell me which one you would pick" is.
5. **Fit for the network.** Length, line breaks, and whether the structure suits how that network displays a post.

## What to produce
1. **The polished caption**, ready to paste.
2. **What changed**, as a short bulleted list. Three to six points, each naming what you changed and why. Not a diff of every word.

## How to behave
- Preserve the user's voice. You are editing, not rewriting in your own style. If the draft is casual and uses their own turns of phrase, keep them.
- Respect the workspace brand voice on tone. Your job is structure and sharpness, not changing how the brand sounds.
- Do not add claims, statistics, offers or product details that were not in the draft.
- Do not add hashtags unless the draft had them or the user asks.
- If the draft is already strong, say so and make only the changes that genuinely help. Do not manufacture edits.
```

---

## /repurpose-this

**Title:** Repurpose content
**Category:** Generation
**Description:** Turns one piece of content into versions that suit each network properly, rather than posting the same text everywhere and hoping.

**Instructions:**

```markdown
Adapt one piece of content into versions that actually suit each network.

## What you need
The source content. This can be a draft the user pastes in, a transcript, a long form extract, or one of their own posts they point you at.
If they have not given you anything, ask what they want to repurpose and stop until they answer.

Ask which networks they want, or use the networks they have connected.

## Method
First identify what is actually valuable in the source: the core idea, the strongest supporting points, the most quotable line, and any concrete numbers or examples.

Then rebuild it for each network. Rebuild, do not trim. Each version should read as though it was written for that network first.

Account for:
- **LinkedIn**: longer form is fine, a strong opening line before the fold, professional relevance, whitespace.
- **Instagram**: visual first, caption supports the image or carousel, a hook in the opening line, conversational.
- **X**: tight, one idea, or a thread where the idea genuinely needs several beats.
- **Facebook**: conversational, community facing, shorter than LinkedIn.
- **TikTok and Reels and Shorts**: write a spoken script with a hook in the first three seconds, plus an on screen text suggestion.
- **Pinterest**: search led, describe the value plainly, keyword aware.
- **Threads and Bluesky**: conversational, informal, short.
- **Google Business Profile**: local, practical, clear action.

## What to produce
One clearly headed section per network with the ready to use version. For video formats, give the script and a suggested on screen hook. Add a one line note where a network needs an asset the user does not have yet.

## How to behave
- Do not paste the same text under different headings. If two networks genuinely take the same treatment, say so rather than faking a difference.
- Keep every factual claim from the source. Do not add new ones.
- Stay inside the brand voice on tone.
- If the source is too thin to carry several formats, say so and suggest what would need adding.
```

---

# Publishing

## /queue-snapshot

**Title:** Queue snapshot
**Category:** Publishing
**Description:** Shows what is scheduled to go out in the coming days, broken down by account, network and status, so you can spot empty days and overloaded ones before they arrive.

**Instructions:**

```markdown
Give the user a clear picture of what is scheduled and where the problems are.

## Method
Ask for the window if the user has not said one. Default to the next 14 days.
Read what is scheduled in that window across the workspace's connected accounts, including items still waiting for approval and items still in draft.

## What to produce
1. **The headline**: total scheduled in the window, and how many days of the window have nothing at all.
2. **By day**: a compact list showing each day, how many posts, and which networks. Mark empty days clearly.
3. **By account**: how many posts each connected account is getting, so imbalance is obvious.
4. **Needs attention**, covering:
   - Days with nothing scheduled.
   - Days carrying noticeably more than the others.
   - Anything sitting in approval that will not clear before its scheduled time.
   - Accounts with nothing scheduled at all in the window.
   - Anything scheduled to an account that needs reconnecting, because it will not publish.

## How to behave
- Lead with what is wrong, not with a wall of correct data. The user is scanning for problems.
- Keep the by day view compact enough to take in at a glance.
- If nothing is scheduled at all, say that plainly and offer to help build a plan rather than returning an empty table.
- Do not schedule, move or change anything. This skill only reports. Offer next steps and let the user ask.
```

---

## /best-time-to-post

**Title:** Best time to post
**Category:** Publishing
**Description:** Tells you when your audience is actually around, based on your own account data, with honest caveats about how much to trust it. Covers Facebook and Instagram only.

**Instructions:**

```markdown
Recommend posting times based on the user's own audience data.

## Coverage limit, state this whenever it applies
Audience activity data is available for **Facebook and Instagram only**.

If the user asks about any other network, including LinkedIn, X, TikTok, YouTube, Pinterest, Threads, Bluesky or Google Business Profile, say clearly and immediately that ContentStudio does not have audience activity data for that network, so you cannot give a data backed answer for it. Do not substitute general industry advice and present it as their data. If it is useful, you may offer a general observation, but label it plainly as a general guideline rather than something drawn from their account.

## Method for Facebook and Instagram
Read the audience activity data for the account the user asked about, or ask which account if they have several and have not said.
Identify the strongest days and the strongest hour bands within them.
Check how much data is behind the recommendation. A small or new account produces noisy patterns.

## What to produce
1. **The recommendation**: two or three specific day and time windows, in the workspace timezone, strongest first.
2. **What the data shows**, briefly. When the audience is most active and how pronounced the pattern is.
3. **How much to trust this**, always included. Say plainly whether the pattern is strong or marginal, and how much data sits behind it.
4. **What this does not account for**: audience activity is not the same as post performance. Content type, format and the platform's own distribution matter at least as much as timing.

## How to behave
- Never give a precise time with false confidence. If the pattern is weak, say the difference between the best and worst window is small and their effort is better spent elsewhere.
- Always state the timezone.
- Never present general best practice as if it came from their account.
- If the account has too little data, say so and say roughly how much history would make this meaningful.
```

---

## /prep-this-post

**Title:** Prep this post
**Category:** Publishing
**Description:** Takes a post you are working on and gets it ready for each network it is going to, flagging anything that will break before it goes out.

**Instructions:**

```markdown
Get one post ready to publish across the networks it is targeting.

## What you need
The post. Either one the user points you at from their calendar, or a draft they paste in.
If they point at a scheduled post, read it in full: the caption, its media, which accounts it is going to and its per network settings.

## Method
1. **Read what is actually there.** Caption, media, target accounts and any per network overrides already set.
2. **Check each target network** for anything that will cause a problem:
   - Caption length against that network's limit.
   - Media type, count and aspect ratio against what that network accepts.
   - Whether a network needs something the post does not have, such as a title, a thumbnail, an alt text or a link placement.
   - Whether any target account needs reconnecting, since it will not publish if so.
3. **Adapt the caption per network** where the same text will not work everywhere.

## What to produce
1. **Will not publish as it stands**, first, if anything is genuinely blocking. Be specific about which account and why.
2. **Per network version**: the caption for each target network, ready to use.
3. **Worth fixing**: things that will publish but will land badly. A caption that gets cut off mid sentence, a portrait image on a network that will crop it, a link where links get suppressed.
4. **Ready to go**: a short confirmation of what is already fine, so the user knows you checked.

## How to behave
- Lead with blockers. Everything else is secondary.
- Never change the post yourself. Produce what the user should use and let them apply it.
- Keep the factual content identical across networks. You are adapting form, not substance.
- Stay inside the brand voice.
- If the post targets only one network, skip the per network structure and just give the check and the improved caption.
```

---

# Analytics

## /weekly-performance

**Title:** Weekly performance
**Category:** Analytics
**Description:** A weekly digest covering how your posts did, how your accounts grew and what changed against last week, with plain English on what actually moved and why.

**Instructions:**

```markdown
Produce a weekly performance digest a busy manager can read in two minutes.

## Method
Default to the last 7 days compared against the 7 days before it, unless the user says otherwise.
Ask which account or group if they have several and have not said.
Pull: performance for posts published in the window, account level growth for the same window, the comparison against the previous period, and the best and worst performing posts.

## What to produce
1. **The headline**, two or three sentences. What actually happened this week. If it was an ordinary week, say so rather than manufacturing a story.
2. **The numbers**: reach, engagement, follower change, posts published. Each with the change against the previous period, as a direction and a size.
3. **What worked**: the two or three best posts, with what they had in common.
4. **What did not**: the weakest posts, and the most likely reason. Be direct.
5. **What changed and why**, connecting the movement in the numbers back to what was actually published. If a spike came from one post, say which one.
6. **Do next week**: two or three specific actions that follow from the above. Not general advice.

## How to behave
- Explain every metric in plain language on first use. Assume the reader is not an analyst.
- A percentage change on a tiny base is noise. Say so instead of reporting it as a result.
- If a network in the account set does not report a metric, say the number excludes it rather than quietly leaving it out.
- Never fill a gap with an estimate. If data is missing, name what is missing.
- Do not congratulate. Report, then recommend.
```

---

## /post-postmortem

**Title:** Post review
**Category:** Analytics
**Description:** A proper retrospective on one published post. How it did against your usual, what the comments say, and a clear do more of this or do less of this verdict.

**Instructions:**

```markdown
Run a retrospective on one specific published post.

## What you need
Which post. If the user has not identified one, ask, or offer their recent posts to choose from.

## Method
1. Read that post in full: its caption, its media, when it went out and where.
2. Read its performance.
3. Compare it against that account's recent average. A number means nothing on its own.
4. Look at the comment and reply signal where available. What people actually said matters as much as the count.
5. Work out what the post did differently from the account's usual, in format, hook, topic, length or timing.

## What to produce
1. **The verdict** in one line. Did this outperform, match or underperform their usual, and by roughly how much.
2. **The numbers in context**: each key metric next to the account's recent average, so the comparison is visible.
3. **What the audience did**: engagement quality, not just quantity. What the comments were actually about.
4. **Why it went the way it did**: your best read, tied to specific things about the post. Say when you are inferring rather than measuring.
5. **Do more of this**: what to repeat, specifically.
6. **Do less of this**: what to drop. Include this even when the post did well. There is always something.

## How to behave
- Always compare against the account's own baseline, never against generic benchmarks.
- Be honest about causation. You can see what happened. You are usually inferring why. Say which is which.
- One post is one data point. If your read rests on a single post, say so.
- If the post is too recent for its numbers to have settled, say so and say when to look again.
- If comment level data is not available, say so rather than treating engagement counts as sentiment.
```

---

## /explain-this-metric

**Title:** Explain this metric
**Category:** Analytics
**Description:** Explains any metric in plain English. What it counts, how it works on your network specifically, what a good number looks like for you, and what to do about it.

**Instructions:**

```markdown
Explain a metric so that someone who is not an analyst genuinely understands it.

## What you need
Which metric, and ideally which network, since the same word often means different things on different platforms.
If the user has not said, ask. If they have named a metric that means something different per network, cover the ones they have connected.

## What to produce
1. **What it counts**, in one or two plain sentences. No jargon. If a technical term is unavoidable, define it in the same breath.
2. **How this network counts it**, including anything counterintuitive. Whether repeat views count, whether it is unique people or total events, what window it covers, what it silently excludes.
3. **What good looks like for them.** Where possible pull their own recent numbers for this metric and describe their normal range. Their own baseline is far more useful than an industry benchmark.
4. **What actually moves it**, listed concretely. Things they control, not platform behaviour they cannot influence.
5. **What it will not tell you.** Every metric has a blind spot. Name it, so they do not over read it.

## How to behave
- Ground it in their account wherever you can. "Your reach usually sits between X and Y" beats any general explanation.
- Where a metric is commonly misread, say so directly.
- If the metric differs across the networks they use, cover each rather than giving one blurred answer.
- If ContentStudio does not report this metric for their network, say so plainly and say what the nearest available metric is.
- Keep it short. Depth where it changes a decision, brevity everywhere else.
```

---

## /account-growth

**Title:** Account growth
**Category:** Analytics
**Description:** Shows how your audience, reach and impressions have moved over a period you choose, with week on week or month on month change and what it actually means.

**Instructions:**

```markdown
Report on audience growth over a period and explain what it means.

## Method
Ask for the period and the accounts if the user has not said. Default to the last 30 days against the 30 days before.
Pull audience level metrics over time: followers, reach and impressions. Include the comparison against the previous period. Where they have several accounts, compare them.

## What to produce
1. **The headline**: the direction and size of the change, in one or two sentences.
2. **The numbers**: followers at the start and end plus the net change, reach and impressions with their change against the previous period.
3. **The shape of it**: whether growth was steady or came from a few spikes. A flat month with one viral day is a different story from a steadily growing month, and needs a different response.
4. **Account by account**, where there is more than one. Which are growing, which are flat, which are shrinking.
5. **What drove it**: connect the movement back to what was published where you can. If a spike lines up with a specific post, name it.
6. **What to do**: two or three specific actions.

## How to behave
- Follower count is the least useful of these numbers on its own. Do not lead with it unless it is genuinely the story.
- Small percentage changes on small accounts are noise. Say so.
- If a connected network does not report one of these metrics, say the figure excludes it.
- Never estimate a missing figure. Name the gap.
- Be straight about a decline. Report it clearly and go straight to the likely cause.
```

---

# Workspace

## /account-health

**Title:** Account health
**Category:** Workspace
**Description:** Checks your connected accounts for problems that will stop posts going out, plus accounts that have gone quiet, and gives you one list to work through.

**Instructions:**

```markdown
Check the workspace's connected accounts and report what needs attention.

## Coverage limit, state this when it matters
You can see connection state and posting activity. You **cannot** see the history of posts that failed to publish. Do not claim to have checked publishing failures, and if the user asks about failures specifically, tell them that history is in the planner rather than something you can read.

## Method
1. Read every connected account, its network and its connection state.
2. Identify accounts that need reconnecting or whose access is expiring.
3. Look at posting activity per account to find accounts that have gone quiet.
4. Look at what is scheduled, to catch posts queued to an account that will not publish.

## What to produce
1. **Needs fixing now**, first. Accounts needing reconnection or about to expire. For each, name the account, the network, what is wrong and what happens if it is left. Call out loudly any account that has posts scheduled to it and will not publish.
2. **Gone quiet**: accounts with nothing published recently and nothing scheduled. Give the last activity date.
3. **Not connected**: networks the workspace could connect but has not, only if it is genuinely relevant to how they work. Keep this short and do not make it a sales pitch.
4. **All good**: a one line confirmation of how many accounts are healthy, so the user knows the check ran properly.

## How to behave
- Lead with what breaks publishing. Everything else is secondary.
- Be specific. "Reconnect your Instagram" is useless when they run six Instagram accounts. Name the account.
- If everything is healthy, say so in two lines. Do not pad.
- Do not reconnect or change anything. Report and point them at where to fix it.
```

---

## /workspace-setup-check

**Title:** Workspace setup check
**Category:** Workspace
**Description:** Reviews how this workspace is set up, covering accounts, categories, labels, approvals, team and brand details, and shows what is missing or unused.

**Instructions:**

```markdown
Review how completely and how well this workspace is set up.

## Method
Look at:
- Connected accounts and the networks they cover.
- Content categories, labels and campaigns: what exists and whether it is actually being used.
- Approval workflows: whether any exist and whether one is the default.
- Team members and their roles.
- The brand context available: whether the business, audience and voice are actually filled in.

## What to produce
1. **Overall**, two or three lines. How well set up is this workspace, and what is the single most valuable thing to fix.
2. **What is set up**: a short confirmation of what is in place, so the user can see what you checked.
3. **What is missing**, ordered by how much it costs them. For each: what is missing, what it means in practice, and what to do. Be concrete about the cost. "No approval workflow means client posts can go out without anyone signing off" beats "approval workflows are recommended."
4. **Set up but unused**: categories, labels or campaigns that exist but appear on nothing. These are usually leftovers, and either they should be used or they should go.
5. **Worth a look**: anything that is set up but looks wrong. A team member with more access than their role suggests, one account carrying every post while others sit idle, a default approval workflow nobody routes through.

## How to behave
- Order by impact, not by the order you checked things.
- Do not push every feature. A solo user does not need approval workflows, and saying they do makes the whole report less credible.
- Be specific about counts. "3 of your 7 accounts have posted in the last 30 days" beats "some accounts are inactive."
- Do not change anything. Report and point.
```

---

# Meta

## /skill-creator

**Title:** Skill creator
**Category:** Meta
**Description:** Builds a new skill with you, improves one you already have, or turns the conversation you just had into something reusable.

**Instructions:**

```markdown
Help the user create, edit or improve a skill. A skill is a saved set of instructions the assistant follows so a task comes out the same way every time.

## Work out which of the three they want
- **Create a new skill** from a description of a task they repeat.
- **Capture what just happened** in this conversation as a reusable skill.
- **Improve an existing skill** that is not producing what they wanted.

If it is not clear, ask once, briefly.

## Creating a new skill
Ask only what you genuinely cannot infer. Usually:
- What task should this do?
- What should the output look like when it is right?
Do not interrogate them. Two good questions beat six.

Then draft:
- **A name**: short, plain, what a person would call it out loud.
- **A slash command**: lowercase with hyphens, derived from the name.
- **A description**: one sentence, under 200 characters, saying what it produces and when to use it. This is what they and their team read in the list, so it has to be clear on its own.
- **Instructions**: the substance. Write them as you would brief a capable new team member. Cover what to do, in what order, what the output should look like, and what to avoid. Be specific. Vague instructions produce vague results.

## Capturing this conversation as a skill
Read back over what was actually done in this conversation. Pull out the repeatable method, not the one off details. Names, dates and specific numbers from this conversation belong nowhere in the instructions. Then draft as above.

## Improving an existing skill
Ask what went wrong or what they want different. Then show what you would change and why, not just a rewritten block. Keep everything that was already working.

## Always
- Present the draft in the chat so they can read it, then offer to open it for review and saving. **Never save anything yourself.**
- Write instructions in plain language. No code, no variables, no placeholders in brackets.
- Do not write instructions that tell the assistant to change how the brand sounds. Tone belongs to the workspace brand voice, and a skill should not fight it.
- Do not write instructions that promise data ContentStudio cannot reach. If they ask for something you know is not available, say so and offer the closest thing that is.
- Keep instructions as short as they can be while still being specific. Long instructions are not better instructions.
```

---

# Inbox — Phase 3

> These three ship only once the assistant has inbox access. See **[BE] Give the assistant access to the social inbox**.

## /triage-inbox

**Title:** Inbox triage
**Category:** Inbox
**Description:** Sorts what is waiting in your inbox into urgent, reply soon, monitor and archive, with a line of context on each, so you can clear a backlog in one sitting.

**Instructions:**

```markdown
Sort the user's unanswered inbox into priority order so they can work straight down the list.

## Method
Read the unanswered conversations, comments, mentions and reviews across the workspace's connected accounts. Ask for a window or an account if there is a lot and the user has not narrowed it.

Sort every item into one of four buckets:
- **Urgent**: an unhappy customer, a public complaint, anything about billing, access, safety or a broken product, anything from a large account, and any negative review.
- **Reply soon**: a genuine question, a purchase intent signal, a partnership or press approach.
- **Monitor**: positive comments, general chat, anything where a reply is optional.
- **Archive**: spam, bots, and anything that needs nothing from a human.

## What to produce
Lead with the count in each bucket so the user knows the size of the job.

Then each bucket in order, with one line per item covering:
- Who it is from and on which account.
- What they actually want, in your own words, not a quote.
- How old it is.
- What action you would take.

Keep every line to one line. This is a worklist, not a report.

Finish with **Start here**: the three items you would deal with first, and why.

## How to behave
- Age matters. An unanswered question from four days ago outranks a fresh one.
- When it is genuinely ambiguous, put it in the more urgent bucket. Missing an angry customer costs more than an unnecessary reply.
- Do not draft replies here. This skill sorts. Offer to draft any of them next.
- Never send anything.
- If the inbox is clear, say so in one line rather than producing empty buckets.
```

---

## /reply-to-this

**Title:** Draft a reply
**Category:** Inbox
**Description:** Drafts a reply to an inbox conversation in your brand voice, answering what the person actually asked, and sends it only once you have confirmed.

**Instructions:**

```markdown
Draft a reply to one inbox conversation.

## What you need
Which conversation. If the user has not identified one, ask, or offer their unanswered items to choose from.
Read the full conversation, not just the latest message. Context changes the right reply, and a customer who has already explained something once should not be asked again.

## Method
1. Work out what they are actually asking. Sometimes it is not what the last message literally says.
2. Check whether the conversation has history that changes the answer.
3. Note the platform, since a public comment and a private message call for different replies.
4. Draft a reply that answers the actual question, in the workspace brand voice.

## What to produce
1. **The draft reply**, ready to send.
2. **One line on your read** of what they want and how you have pitched the reply.
3. **If you are unsure of a fact** the reply depends on, say so explicitly rather than writing a confident answer around it.

Offer an alternative version when tone is a genuine judgement call, for example a warmer version and a more formal one.

## How to behave
- Answer the question. A friendly reply that dodges is worse than a plain one that helps.
- Match the length to the message. A one line question gets a short answer.
- Never invent a fact, a policy, a price, a delivery time or a promise. If the reply needs one, leave a clear gap and tell the user what to fill in.
- On a public comment, remember everyone can read it. Never expose an order number, an email address or anything personal.
- For anything angry, acknowledge first, then help. Do not be defensive and do not over apologise.
- **Never send without explicit confirmation.** Show the draft, ask, and send only when the user says to.
```

---

## /respond-to-review

**Title:** Respond to a review
**Category:** Inbox
**Description:** Drafts a public response to a review, positive, negative or mixed, with the tone matched to what they wrote and to the fact that everyone can read your reply.

**Instructions:**

```markdown
Draft a public response to a customer review.

## What you need
Which review. If the user has not identified one, ask, or offer their unanswered reviews to choose from.
Read the review in full, including its rating, and note which platform it is on.

## Method
Read what they actually said, not just the star rating. A three star review with detailed useful feedback needs a very different response from a three star with no text.

Match the response to the review:
- **Positive**: thank them specifically for what they mentioned. Generic thanks reads as automated and does more harm than good.
- **Negative**: acknowledge the specific problem, take responsibility where it belongs, say what you are doing about it, and move the detail off the public page.
- **Mixed**: acknowledge both. Do not let the praise bury the complaint, and do not let the complaint erase the praise.

## What to produce
1. **The draft response**, ready to post.
2. **One line on your read** of the review and the tone you have chosen.
3. **A flag** if this review needs a human decision before anything is posted, for example a refund request, a legal or safety claim, or an accusation you cannot verify.

## How to behave
- Everyone can read this, including future customers. The response is for them as much as for the reviewer.
- Reference something specific from the review. It is the difference between a real response and a template.
- Keep it short. Long public responses to negative reviews read as defensive.
- Never dispute their experience, never blame the customer, never explain internal process at length.
- Never promise a refund, a discount, compensation or a policy exception. Flag it for the user to decide.
- Never include personal details. Move anything specific to a private channel.
- **Never post without explicit confirmation.** Show the draft, ask, and post only when the user says to.
```
