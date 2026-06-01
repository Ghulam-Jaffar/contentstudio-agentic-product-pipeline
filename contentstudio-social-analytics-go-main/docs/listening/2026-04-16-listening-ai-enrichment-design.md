# Listening AI Enrichment Pipeline — Design Spec

**Date:** 2026-04-16
**Status:** Approved
**Scope:** Two features across two projects (Go pipeline + Python AI agents)

---

## 1. Problem Statement

The social listening pipeline scrapes posts via Data365 API across 6 platforms. Posts need:
1. **AI tagging** — classify each post into one or more tags from a closed set of 10
2. **Sentiment analysis** — label (positive/negative/neutral) + confidence score

The current pipeline has a sentiment stage that calls the AI agents project **one post at a time**, which is too expensive at scale (potentially 20,000 posts per Twitter work order). Additionally, tags like "Brand Mention" and "Competitor Mention" require brand/competitor context that doesn't exist yet.

## 2. Features

### Feature 1: Brand Discovery Endpoint

A standalone endpoint where the frontend sends a website URL and receives AI-detected brand information, competitors, industry category, and suggested listening topics. This is called once during topic creation — results are stored on the topic for reuse by the enrichment pipeline.

### Feature 2: Batch Sentiment + AI Tagging

Replace the existing one-by-one sentiment stage with an async batch enrichment service that performs both sentiment analysis and AI tagging in a single LLM call per batch of up to 50 posts, using brand/competitor context from the topic.

---

## 3. Feature 1: Brand Discovery Endpoint

### 3.1 API Contract

**Project:** contentstudio-ai-agents (Python/FastAPI)
**Endpoint:** `POST /api/v1/listening/discover`
**Caller:** Frontend (Laravel) directly — the Go pipeline does not call this.

#### Request

```python
class BrandDiscoveryRequest(BaseModel):
    url: str = Field(description="Website URL to analyze")
```

#### Response

```python
class BrandInfo(BaseModel):
    name: str                  # e.g. "ContentStudio"
    description: str           # e.g. "Social media management platform"
    keywords: list[str]        # e.g. ["ContentStudio", "contentstudio.io"]

class CompetitorInfo(BaseModel):
    name: str                  # e.g. "Hootsuite"
    keywords: list[str]        # e.g. ["Hootsuite", "hootsuite.com"]

class SuggestedTopic(BaseModel):
    name: str                  # e.g. "Brand Reputation"
    keywords: list[str]        # e.g. ["ContentStudio", "contentstudio"]
    description: str           # e.g. "Monitor mentions of your brand"

class BrandDiscoveryResponse(BaseModel):
    success: bool
    brand: BrandInfo
    industry: str              # e.g. "SaaS / Social Media Marketing"
    competitors: list[CompetitorInfo]  # always 5
    suggested_topics: list[SuggestedTopic]
    processing_time: float | None = None
    model_used: str | None = None
```

### 3.2 Architecture

**Agent:** `BrandDiscoveryAgent` extending `BaseAgent`
**Model:** Groq (primary), GPT (fallback)
**Approach:** Two-phase, single LLM call

#### Data Flow

```
POST /api/v1/listening/discover { url: "https://example.com" }
    |
    v
Step 1: Firecrawl scrape homepage
    -> markdown content, summary, branding.name, metadata.title
    |
    v
Step 2: Extract brand name
    -> branding.name || metadata.title || domain parse fallback
    |
    v
Step 3: Exa web search "{brand_name} competitors alternatives"
    -> top search results
    |
    v
Step 4: Single Groq LLM call
    prompt = homepage markdown + web search results
    output_schema = BrandDiscoveryOutput
    -> brand, industry, 5 competitors, suggested topics
    |
    v
BrandDiscoveryResponse -> Frontend
    |
    v
User reviews/edits -> Saves as ListeningTopic in MongoDB
    (brand discovery results stored in ai_context_hint field)
```

#### Degradation

- Firecrawl fails (invalid URL, timeout) -> return HTTP error immediately, do not call LLM
- Exa search fails -> proceed with homepage content only, LLM infers competitors from training knowledge

### 3.3 File Structure

```
contentstudio-ai-agents/src/
    agents/
        listening/
            __init__.py
            brand_discovery.py       # BrandDiscoveryAgent + factory function
    api/
        routers/
            listening/
                __init__.py
                listening_router.py  # POST /listening/discover endpoint
    models/
        listening.py                 # Request/Response Pydantic models
```

Router registration in `src/api/main.py`:
```python
from src.api.routers.listening.listening_router import router as listening_router
app.include_router(listening_router, prefix="/api/v1/listening", tags=["listening"])
```

### 3.4 Key Implementation Details

- Firecrawl: reuse existing pattern from `business_info_router.py` — `formats=["markdown", "summary", "branding"]`, `only_main_content=True`
- Exa: use `exa-py` SDK directly (programmatic search, not an Agno tool)
- Brand name extraction: `branding.name` -> `metadata.title` -> domain parse (fallback chain)
- Competitors: always return exactly 5
- Observability: `@trace_api(name="listening.discover")` for Langfuse tracing
- Singleton agent pattern: global `BrandDiscoveryAgent` instance, lazy-initialized

---

## 4. Feature 2: Batch Sentiment + AI Tagging

### 4.1 Pipeline Restructure (Go Side)

#### Current Flow (being replaced)

```
Parser -> listening-parsed -> Sentiment (1-by-1) -> listening-enriched -> Sink -> ClickHouse
```

#### New Flow

```
Parser -> listening-parsed -+-> Sink (consumer group: listening-sink-group)
                            |       -> ClickHouse (immediate, no tags/sentiment)
                            |
                            +-> Enrichment (consumer group: listening-enrichment-group)
                                    -> buffers by topic_id
                                    -> flushes at 50 posts or 30s timeout
                                    -> reads ai_context_hint from MongoDB (cached 5min)
                                    -> calls POST /api/v1/listening/analyze-batch
                                    -> writes enriched rows to ClickHouse
                                       (ReplacingMergeTree upsert with newer updated_at)
```

#### Changes

1. **Sink rewired** — consumes from `listening-parsed` instead of `listening-enriched`
2. **Sentiment stage removed** — consumer, service, and stage registration deleted
3. **New enrichment service** — separate Kafka consumer group on `listening-parsed`
4. **`listening-enriched` topic** — no longer produced to or consumed from. Keep the topic definition in code (no breaking change) but remove all producers/consumers. Can be fully retired in a future cleanup pass

### 4.2 AI Agents Endpoint

**Project:** contentstudio-ai-agents (Python/FastAPI)
**Endpoint:** `POST /api/v1/listening/analyze-batch`

#### Request

```python
class MentionForAnalysis(BaseModel):
    mention_id: str
    text: str

class AIContext(BaseModel):
    brand_name: str
    brand_keywords: list[str]
    competitors: list[CompetitorInfo]  # name + keywords
    industry: str

class AnalyzeBatchRequest(BaseModel):
    mentions: list[MentionForAnalysis]  # up to 50 per call
    context: AIContext
```

#### Response

```python
class MentionAnalysisResult(BaseModel):
    mention_id: str
    sentiment_label: str        # "positive", "negative", "neutral"
    sentiment_score: float      # 0.0 - 1.0
    ai_tags: list[str]          # subset of closed set

class AnalyzeBatchResponse(BaseModel):
    success: bool
    results: list[MentionAnalysisResult]
    processing_time: float | None = None
    model_used: str | None = None
```

#### Closed Tag Set

```
Brand Mention, Competitor Mention, Industry Insight, Buy Intent,
Bug Report, User Feedback, Promotional Post, Product Question, Event, Hiring
```

Posts can have multiple tags (multi-select). Tags are from this closed set only.

#### Agent

- `MentionAnalyzerAgent` extending `BaseAgent`, Groq primary, GPT fallback
- Single LLM call per batch — prompt includes all post texts + brand/competitor context
- System prompt instructs: for each post, return sentiment (label + score) and applicable tags from the closed set
- Structured output via Pydantic schema

### 4.3 Enrichment Service (Go Side)

**Location:** `src/services/listening/enrichment/service.go`

#### Buffering Strategy

```
EnrichmentService
    topicBuffers: map[topicID][]ListeningMention

    Flush triggers:
      1. Buffer hits 50 mentions for a given topic
      2. 30-second timer fires (flushes all topics)
      (whichever comes first)

    On flush(topicID):
      1. Read ai_context_hint from MongoDB (in-memory cache, 5min TTL)
      2. POST /api/v1/listening/analyze-batch with {mentions, context}
      3. Merge AI results into full mention rows (carry complete data from Kafka)
      4. Batch INSERT to ClickHouse (newer updated_at -> RMT keeps latest)
      5. Clear buffer for that topic
```

#### Graceful Degradation

- AI agents endpoint down -> mentions stay in buffer, retry on next flush cycle
- Batch call partially fails -> store successful results, re-buffer failures
- `ai_context_hint` empty (topic created without brand discovery) -> call AI agents without brand/competitor context, LLM infers what it can
- Max buffer age: 5 minutes -> if AI is down for 5min, send to DLQ to prevent unbounded memory growth

#### Integration in listening-consumer/main.go

- New consumer: `listening-enrichment-group` on `listening-parsed`
- New stage: `enrichment` started alongside fetcher, parser, sink
- Removed: sentiment consumer, sentiment service, sentiment stage registration

### 4.4 MongoDB: ai_context_hint Field

**Added to `ListeningTopic` model:**

```go
AIContextHint string `bson:"ai_context_hint" json:"ai_context_hint"`
```

**Contents:** JSON string storing brand discovery results:

```json
{
    "brand_name": "ContentStudio",
    "brand_keywords": ["ContentStudio", "contentstudio.io"],
    "competitors": [
        {"name": "Hootsuite", "keywords": ["Hootsuite", "hootsuite.com"]},
        {"name": "Buffer", "keywords": ["Buffer", "buffer.com"]}
    ],
    "industry": "SaaS / Social Media Marketing"
}
```

**Lifecycle:**
1. Frontend calls brand discovery endpoint -> gets suggestions
2. User reviews/edits -> frontend saves topic to MongoDB with `ai_context_hint` populated
3. Go enrichment service reads `ai_context_hint` when processing mentions for that topic
4. Cached in-memory for 5 minutes to avoid repeated MongoDB reads

### 4.5 File Structure (Go Side Changes)

```
src/services/listening/
    enrichment/
        service.go              # EnrichmentService with buffering + flush logic
    sentiment/
        service.go              # DELETED (replaced by enrichment)
    sink/
        service.go              # MODIFIED: consume listening-parsed instead of listening-enriched
    listening-consumer/
        main.go                 # MODIFIED: remove sentiment stage, add enrichment stage,
                                #           rewire sink to listening-parsed

src/models/db/mongo/
    listening_topic.go          # MODIFIED: add AIContextHint field
```

### 4.6 File Structure (AI Agents Side)

```
contentstudio-ai-agents/src/
    agents/
        listening/
            __init__.py
            brand_discovery.py       # (Feature 1)
            mention_analyzer.py      # MentionAnalyzerAgent + factory
    api/
        routers/
            listening/
                __init__.py
                listening_router.py  # Both endpoints registered here
    models/
        listening.py                 # All request/response models for both features
```

---

## 5. Cost Analysis

### Brand Discovery (per call)
- 1 Firecrawl scrape: ~$0.01
- 1 Exa search: ~$0.01
- 1 Groq LLM call (~3K tokens): ~$0.001
- **Total: ~$0.02 per topic creation** (one-time cost)

### Batch Enrichment (per 50 posts)
- 1 Groq LLM call (~5-8K tokens for 50 short posts + context): ~$0.005
- **Per-post cost: ~$0.0001** (vs ~$0.001 for individual calls = 10x savings)
- **20,000 Twitter posts: ~$2** (400 batches x $0.005)

### Comparison to Current (one-by-one sentiment)
- Current: 20,000 posts x individual call = 20,000 LLM calls
- New: 20,000 posts / 50 per batch = 400 LLM calls (50x fewer calls)
- Tags included at zero marginal cost (same call does both)

---

## 6. Observability

- Brand discovery: `@trace_api(name="listening.discover")` — Langfuse tracing
- Batch analysis: `@trace_api(name="listening.analyze_batch")` — Langfuse tracing
- Go enrichment service: structured zerolog with buffer utilization, flush counts, AI latency
- ClickHouse: enriched row count trackable via `sentiment_label != ''` filter

---

## 7. Decisions Log

| Decision | Chosen | Alternatives Considered |
|---|---|---|
| Architecture | Two-phase single LLM call | Agent with Agno tools, Two cheap model calls |
| Scraping | Firecrawl homepage only | Multi-page crawl, simple HTTP fetch |
| Competitor discovery | LLM + Exa web search | LLM inference only |
| Competitor count | Fixed top 5 | Variable with cap, user-configurable |
| Endpoint location | `/api/v1/listening/discover` | `/api/v1/brand/discover` |
| Pipeline change | Replace sentiment stage | Add separate tagging stage alongside |
| Enrichment timing | Async (posts land without tags first) | Synchronous (tags before storage) |
| Tag context | `ai_context_hint` from MongoDB | Re-discover per batch, pass from frontend |
| Tag set | Closed set of 10 | Open/extensible |
| Model | Groq everywhere | Claude Sonnet, Haiku, GPT-4o-mini |
| Batch size | 50 posts per LLM call | 20, 100 |
| Flush timeout | 30 seconds | 10s, 60s |
| Max buffer age | 5 minutes then DLQ | Unbounded retry |
| Exclude keywords | Not included (tight scope) | AI-suggested exclusions |
