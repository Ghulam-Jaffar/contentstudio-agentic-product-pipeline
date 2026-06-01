# Listening AI Enrichment Pipeline — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add brand discovery endpoint and batch sentiment + AI tagging to the social listening pipeline across both the Python AI agents project and the Go pipeline project.

**Architecture:** Two new endpoints in the AI agents project (brand discovery + batch analysis), a new async enrichment service in the Go pipeline that buffers mentions by topic and batch-calls the AI agents endpoint, and a pipeline rewire that removes the one-by-one sentiment stage in favor of direct sink writes with async enrichment.

**Tech Stack:** Python (FastAPI, Agno, Firecrawl, Exa, Groq), Go (Kafka consumers, ClickHouse batch writes, MongoDB)

**Spec:** `docs/superpowers/specs/2026-04-16-listening-ai-enrichment-design.md`

---

## Two-Project Coordination

This plan spans two repos. Tasks are ordered for dependency flow:

| Phase | Project | What |
|-------|---------|------|
| Tasks 1-5 | AI Agents (`contentstudio-ai-agents`) | Models, agents, endpoints |
| Tasks 6-9 | Go Pipeline (`contentstudio-social-analytics-go`) | MongoDB model, enrichment service, pipeline rewire |

---

## Task 1: Pydantic Models for Both Endpoints

**Project:** `contentstudio-ai-agents`

**Files:**
- Create: `src/models/listening.py`
- Create: `src/agents/listening/__init__.py`

- [ ] **Step 1: Create the listening models file**

```python
# src/models/listening.py
"""Pydantic models for social listening AI endpoints."""

from pydantic import BaseModel, Field


# --- Brand Discovery ---

class BrandDiscoveryRequest(BaseModel):
    """Request model for brand discovery endpoint."""
    url: str = Field(description="Website URL to analyze")


class BrandInfo(BaseModel):
    """Detected brand information."""
    name: str = Field(description="Brand name")
    description: str = Field(description="Short brand description")
    keywords: list[str] = Field(description="Search keywords for this brand")


class CompetitorInfo(BaseModel):
    """Detected competitor information."""
    name: str = Field(description="Competitor name")
    keywords: list[str] = Field(description="Search keywords for this competitor")


class SuggestedTopic(BaseModel):
    """AI-suggested listening topic."""
    name: str = Field(description="Topic name")
    keywords: list[str] = Field(description="Keywords to monitor")
    description: str = Field(description="What this topic monitors")


class BrandDiscoveryOutput(BaseModel):
    """Structured output schema for the brand discovery agent LLM call."""
    brand: BrandInfo
    industry: str = Field(description="Industry category, e.g. 'SaaS / Social Media Marketing'")
    competitors: list[CompetitorInfo] = Field(description="Top 5 competitors")
    suggested_topics: list[SuggestedTopic] = Field(description="Suggested listening topics")


class BrandDiscoveryResponse(BaseModel):
    """API response for brand discovery endpoint."""
    success: bool
    brand: BrandInfo
    industry: str
    competitors: list[CompetitorInfo]
    suggested_topics: list[SuggestedTopic]
    processing_time: float | None = None
    model_used: str | None = None


# --- Batch Mention Analysis ---

LISTENING_TAGS = [
    "Brand Mention",
    "Competitor Mention",
    "Industry Insight",
    "Buy Intent",
    "Bug Report",
    "User Feedback",
    "Promotional Post",
    "Product Question",
    "Event",
    "Hiring",
]


class AIContext(BaseModel):
    """Brand/competitor context for mention classification."""
    brand_name: str = Field(description="The brand being monitored")
    brand_keywords: list[str] = Field(description="Keywords identifying the brand")
    competitors: list[CompetitorInfo] = Field(description="Known competitors with keywords")
    industry: str = Field(description="Industry category")


class MentionForAnalysis(BaseModel):
    """A single mention to analyze."""
    mention_id: str = Field(description="Unique mention identifier")
    text: str = Field(description="Post text content")


class AnalyzeBatchRequest(BaseModel):
    """Request model for batch mention analysis."""
    mentions: list[MentionForAnalysis] = Field(description="Up to 50 mentions per call")
    context: AIContext


class MentionAnalysisResult(BaseModel):
    """Analysis result for a single mention."""
    mention_id: str
    sentiment_label: str = Field(description="positive, negative, or neutral")
    sentiment_score: float = Field(description="Confidence score 0.0 to 1.0")
    ai_tags: list[str] = Field(description="Tags from the closed set")


class BatchAnalysisOutput(BaseModel):
    """Structured output schema for the mention analyzer agent LLM call."""
    results: list[MentionAnalysisResult]


class AnalyzeBatchResponse(BaseModel):
    """API response for batch analysis endpoint."""
    success: bool
    results: list[MentionAnalysisResult]
    processing_time: float | None = None
    model_used: str | None = None
```

- [ ] **Step 2: Create the listening agents __init__.py**

```python
# src/agents/listening/__init__.py
```

Empty file — just makes it a package.

- [ ] **Step 3: Commit**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-ai-agents
git add src/models/listening.py src/agents/listening/__init__.py
git commit -m "feat(listening): add Pydantic models for brand discovery and batch analysis"
```

---

## Task 2: Brand Discovery Agent

**Project:** `contentstudio-ai-agents`

**Files:**
- Create: `src/agents/listening/brand_discovery.py`

- [ ] **Step 1: Create the brand discovery agent**

```python
# src/agents/listening/brand_discovery.py
"""Brand Discovery Agent — scrapes a website and identifies brand, competitors, industry, and topics."""

import asyncio
import os
from functools import partial
from typing import Any

from exa_py import Exa
from pydantic import BaseModel

from src.agents.base import BaseAgent
from src.models.listening import BrandDiscoveryOutput
from src.utils.config import get_config
from src.utils.logging import get_logger

logger = get_logger(__name__)

try:
    from firecrawl import Firecrawl

    FIRECRAWL_AVAILABLE = True
except ImportError:
    FIRECRAWL_AVAILABLE = False
    Firecrawl = None

SCRAPE_TIMEOUT_MS = 30000

_firecrawl_app = None


def _get_firecrawl_app():
    """Get or create Firecrawl app instance."""
    global _firecrawl_app
    if _firecrawl_app is None:
        if not FIRECRAWL_AVAILABLE:
            raise RuntimeError(
                "Firecrawl SDK not installed. Install with: uv pip install firecrawl-py"
            )
        api_key = os.getenv("FIRECRAWL_API_KEY")
        if not api_key:
            raise ValueError("FIRECRAWL_API_KEY environment variable not set")
        _firecrawl_app = Firecrawl(api_key=api_key)
        logger.info("Firecrawl app initialized")
    return _firecrawl_app


def _extract_scrape_payload(result: Any) -> dict[str, Any]:
    """Normalize Firecrawl SDK response into a plain dict."""
    if isinstance(result, dict):
        return result
    if hasattr(result, "model_dump"):
        return result.model_dump()
    if hasattr(result, "__dict__"):
        return result.__dict__
    return {}


class BrandDiscoveryAgent(BaseAgent):
    """Analyzes a website to discover brand, competitors, industry, and suggest topics."""

    def __init__(self):
        config = get_config()
        super().__init__(
            name="BrandDiscovery",
            description="Discovers brand identity, competitors, and suggests listening topics from a website URL",
            output_schema=BrandDiscoveryOutput,
            use_fallback=True,
            use_speed_model=False,
            primary_model_id=config.groq_model,
            fallback_model_id=config.gpt_model,
        )

    def _setup_agents(self) -> None:
        """Set up Groq as primary, GPT as fallback."""
        from agno.models.groq import Groq
        from agno.models.openai import OpenAIChat

        try:
            self.primary_agent = self._create_agent(
                model=Groq(
                    id=self.config.groq_model,
                    api_key=self.config.groq_api_key,
                ),
                name=f"{self.name}_primary",
            )
        except Exception as e:
            self.logger.log_error(
                e, {"context": "Failed to initialize Groq, falling back to GPT"}
            )
            self.primary_agent = self._create_agent(
                model=OpenAIChat(
                    id=self.config.gpt_model,
                    api_key=self.config.openai_api_key,
                ),
                name=f"{self.name}_primary",
            )

        if self.use_fallback:
            self.fallback_agent = self._create_agent(
                model=OpenAIChat(
                    id=self.config.gpt_model,
                    api_key=self.config.openai_api_key,
                ),
                name=f"{self.name}_fallback",
            )

    def get_instructions(self) -> list[str]:
        return [
            "You are a brand intelligence analyst.",
            "You will receive website content and web search results about a company.",
            "",
            "Your task:",
            "1. Identify the brand: name, short description, and search keywords (include brand name and domain).",
            "2. Classify the industry (e.g. 'SaaS / Social Media Marketing', 'E-commerce / Fashion').",
            "3. Identify exactly 5 top competitors with their names and search keywords.",
            "4. Suggest 3-5 listening topics with keywords and descriptions.",
            "",
            "Rules:",
            "- Competitors must be real, well-known companies in the same space.",
            "- Each competitor must have 2-3 keywords (brand name + domain).",
            "- Suggested topics should cover: brand reputation, competitor tracking, industry trends, and customer feedback.",
            "- Output must be valid JSON matching the schema exactly.",
        ]

    async def discover(self, url: str) -> BrandDiscoveryOutput:
        """Run the full brand discovery pipeline for a URL.

        Steps:
        1. Scrape homepage via Firecrawl
        2. Extract brand name from Firecrawl branding/metadata
        3. Search for competitors via Exa
        4. Single LLM call with all context
        """
        # Step 1: Scrape homepage
        website_data = await self._scrape_homepage(url)
        markdown = website_data.get("markdown", "")
        summary = website_data.get("summary", "")
        branding = website_data.get("branding") or {}
        metadata = website_data.get("metadata") or {}

        # Step 2: Extract brand name heuristically
        brand_name = (
            branding.get("name")
            or metadata.get("og:site_name")
            or metadata.get("title", "")
            or self._brand_from_url(url)
        )

        # Step 3: Search for competitors via Exa
        search_context = await self._search_competitors(brand_name)

        # Step 4: Build prompt and call LLM
        prompt = self._build_prompt(url, brand_name, markdown, summary, search_context)
        response = await self.run_async(prompt=prompt)

        if hasattr(response, "content"):
            return response.content
        return response

    async def _scrape_homepage(self, url: str) -> dict[str, Any]:
        app = _get_firecrawl_app()
        logger.info(f"Scraping homepage: {url}")

        scrape_kwargs = {
            "formats": ["markdown", "summary", "branding"],
            "only_main_content": True,
            "remove_base64_images": True,
            "block_ads": True,
            "timeout": SCRAPE_TIMEOUT_MS,
        }

        try:
            result = await asyncio.to_thread(partial(app.scrape, url, **scrape_kwargs))
        except TypeError:
            result = await asyncio.to_thread(
                partial(app.scrape, url, formats=["markdown", "branding"])
            )

        scrape_data = _extract_scrape_payload(result)
        content = scrape_data.get("markdown", "")
        summary = scrape_data.get("summary", "")

        if not content and not summary:
            raise ValueError("No content extracted from website")

        return {
            "markdown": content,
            "summary": summary,
            "branding": scrape_data.get("branding") or {},
            "metadata": scrape_data.get("metadata") or {},
        }

    async def _search_competitors(self, brand_name: str) -> str:
        """Search for competitors using Exa. Returns formatted search results or empty string on failure."""
        config = get_config()
        if not config.exa_enabled or not config.exa_api_key:
            logger.warning("Exa not configured, skipping competitor search")
            return ""

        try:
            exa = Exa(api_key=config.exa_api_key)
            query = f"{brand_name} competitors alternatives"
            results = await asyncio.to_thread(
                partial(exa.search_and_contents, query, text=True, num_results=5, type="auto")
            )

            lines = []
            for r in results.results:
                title = getattr(r, "title", "Untitled")
                url = getattr(r, "url", "")
                text = (getattr(r, "text", "") or "")[:500]
                lines.append(f"### {title}\nURL: {url}\n{text}\n")

            return "\n".join(lines)

        except Exception as e:
            logger.warning(f"Exa search failed, proceeding without: {e}")
            return ""

    def _build_prompt(
        self,
        url: str,
        brand_name: str,
        markdown: str,
        summary: str,
        search_context: str,
    ) -> str:
        # Truncate markdown to avoid token limits
        max_content_len = 4000
        if len(markdown) > max_content_len:
            markdown = markdown[:max_content_len] + "\n[...truncated]"

        parts = [
            f"## Website: {url}",
            f"## Detected Brand Name: {brand_name}",
            "",
            "## Homepage Summary",
            summary or "(no summary available)",
            "",
            "## Homepage Content",
            markdown or "(no content available)",
        ]

        if search_context:
            parts.extend([
                "",
                "## Web Search Results: Competitors & Alternatives",
                search_context,
            ])

        parts.extend([
            "",
            "Analyze the above and return the brand, industry, 5 competitors, and suggested topics as JSON.",
        ])

        return "\n".join(parts)

    @staticmethod
    def _brand_from_url(url: str) -> str:
        """Fallback: extract a brand-ish name from the domain."""
        from urllib.parse import urlparse

        hostname = urlparse(url).hostname or ""
        parts = hostname.replace("www.", "").split(".")
        return parts[0].capitalize() if parts else "Unknown"


def create_brand_discovery_agent() -> BrandDiscoveryAgent:
    """Factory function to create a BrandDiscoveryAgent instance."""
    return BrandDiscoveryAgent()
```

- [ ] **Step 2: Commit**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-ai-agents
git add src/agents/listening/brand_discovery.py
git commit -m "feat(listening): add BrandDiscoveryAgent with Firecrawl + Exa"
```

---

## Task 3: Brand Discovery Router + Registration

**Project:** `contentstudio-ai-agents`

**Files:**
- Create: `src/api/routers/listening/__init__.py`
- Create: `src/api/routers/listening/listening_router.py`
- Modify: `src/api/main.py`

- [ ] **Step 1: Create the listening router package**

```python
# src/api/routers/listening/__init__.py
```

- [ ] **Step 2: Create the listening router with the discover endpoint**

```python
# src/api/routers/listening/listening_router.py
"""Listening API endpoints — brand discovery and batch mention analysis."""

import time
import traceback

from fastapi import APIRouter, HTTPException

from src.agents.listening.brand_discovery import create_brand_discovery_agent
from src.models.listening import (
    BrandDiscoveryRequest,
    BrandDiscoveryResponse,
)
from src.utils.config import get_config
from src.utils.logging import get_logger
from src.utils.observability import trace_api

logger = get_logger(__name__)
router = APIRouter()

_brand_discovery_agent = None


def get_brand_discovery_agent():
    """Get or create brand discovery agent singleton."""
    global _brand_discovery_agent
    if _brand_discovery_agent is None:
        _brand_discovery_agent = create_brand_discovery_agent()
        logger.info("Brand discovery agent initialized")
    return _brand_discovery_agent


@router.post("/discover", response_model=BrandDiscoveryResponse)
@trace_api(name="listening.discover")
async def discover_brand(request: BrandDiscoveryRequest) -> BrandDiscoveryResponse:
    """Analyze a website URL to discover brand, competitors, industry, and suggest topics."""
    start_time = time.time()

    try:
        logger.info(f"Brand discovery for URL: {request.url}")
        agent = get_brand_discovery_agent()
        result = await agent.discover(request.url)

        processing_time = time.time() - start_time
        logger.info(f"Brand discovery completed in {processing_time:.2f}s")

        return BrandDiscoveryResponse(
            success=True,
            brand=result.brand,
            industry=result.industry,
            competitors=result.competitors,
            suggested_topics=result.suggested_topics,
            processing_time=processing_time,
            model_used=get_config().groq_model,
        )

    except HTTPException:
        raise
    except ValueError as e:
        logger.error(f"Brand discovery validation error: {e}")
        raise HTTPException(
            status_code=422,
            detail={"error": "Invalid input", "message": str(e)},
        )
    except Exception as e:
        logger.error(f"Brand discovery failed: {e}\n{traceback.format_exc()}")
        raise HTTPException(
            status_code=500,
            detail={"error": "Brand discovery failed", "message": str(e)},
        )
```

- [ ] **Step 3: Register the router in main.py**

Open `src/api/main.py`. Find the section where routers are imported and included. Add:

Import (add alongside existing router imports):
```python
from src.api.routers.listening import listening_router
```

Include (add alongside existing `app.include_router` calls):
```python
app.include_router(listening_router.router, prefix="/api/v1/listening", tags=["Listening"])
```

- [ ] **Step 4: Verify the server starts**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-ai-agents
python -c "from src.api.routers.listening.listening_router import router; print('Router imported OK')"
```

Expected: `Router imported OK`

- [ ] **Step 5: Commit**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-ai-agents
git add src/api/routers/listening/ src/api/main.py
git commit -m "feat(listening): add brand discovery endpoint POST /api/v1/listening/discover"
```

---

## Task 4: Mention Analyzer Agent

**Project:** `contentstudio-ai-agents`

**Files:**
- Create: `src/agents/listening/mention_analyzer.py`

- [ ] **Step 1: Create the mention analyzer agent**

```python
# src/agents/listening/mention_analyzer.py
"""Mention Analyzer Agent — batch sentiment + AI tagging for social listening posts."""

import json

from src.agents.base import BaseAgent
from src.models.listening import (
    AIContext,
    BatchAnalysisOutput,
    LISTENING_TAGS,
    MentionForAnalysis,
)
from src.utils.config import get_config
from src.utils.logging import get_logger

logger = get_logger(__name__)


class MentionAnalyzerAgent(BaseAgent):
    """Classifies batches of social media mentions with sentiment and tags."""

    def __init__(self):
        config = get_config()
        super().__init__(
            name="MentionAnalyzer",
            description="Batch sentiment analysis and AI tagging for social listening mentions",
            output_schema=BatchAnalysisOutput,
            use_fallback=True,
            use_speed_model=False,
            primary_model_id=config.groq_model,
            fallback_model_id=config.gpt_model,
        )

    def _setup_agents(self) -> None:
        """Set up Groq as primary, GPT as fallback."""
        from agno.models.groq import Groq
        from agno.models.openai import OpenAIChat

        try:
            self.primary_agent = self._create_agent(
                model=Groq(
                    id=self.config.groq_model,
                    api_key=self.config.groq_api_key,
                ),
                name=f"{self.name}_primary",
            )
        except Exception as e:
            self.logger.log_error(
                e, {"context": "Failed to initialize Groq, falling back to GPT"}
            )
            self.primary_agent = self._create_agent(
                model=OpenAIChat(
                    id=self.config.gpt_model,
                    api_key=self.config.openai_api_key,
                ),
                name=f"{self.name}_primary",
            )

        if self.use_fallback:
            self.fallback_agent = self._create_agent(
                model=OpenAIChat(
                    id=self.config.gpt_model,
                    api_key=self.config.openai_api_key,
                ),
                name=f"{self.name}_fallback",
            )

    def get_instructions(self) -> list[str]:
        tags_str = ", ".join(LISTENING_TAGS)
        return [
            "You are a social media content classifier.",
            "You will receive a batch of social media posts and brand/competitor context.",
            "",
            "For EACH post, you must determine:",
            "1. **Sentiment**: label ('positive', 'negative', or 'neutral') and score (0.0 to 1.0 confidence).",
            "2. **Tags**: one or more tags from this CLOSED set ONLY:",
            f"   [{tags_str}]",
            "",
            "Tag definitions:",
            "- Brand Mention: post mentions or discusses the monitored brand.",
            "- Competitor Mention: post mentions or discusses a known competitor.",
            "- Industry Insight: post discusses industry trends, news, or developments.",
            "- Buy Intent: post expresses intent to purchase or evaluates buying options.",
            "- Bug Report: post reports a bug, error, or technical issue.",
            "- User Feedback: post provides feedback, review, or opinion about a product/service.",
            "- Promotional Post: post is promotional, advertising, or marketing content.",
            "- Product Question: post asks a question about a product or service.",
            "- Event: post mentions an event, conference, webinar, or meetup.",
            "- Hiring: post is about job openings, hiring, or recruitment.",
            "",
            "Rules:",
            "- Use the brand context to distinguish Brand Mention from Competitor Mention.",
            "- A post can have MULTIPLE tags (e.g., both 'Brand Mention' and 'Bug Report').",
            "- Only use tags from the closed set above. Do NOT invent new tags.",
            "- Every post must get at least one tag.",
            "- Return results for ALL posts in the batch, in the same order.",
            "- Output must be valid JSON matching the schema exactly.",
        ]

    async def analyze_batch(
        self,
        mentions: list[MentionForAnalysis],
        context: AIContext,
    ) -> BatchAnalysisOutput:
        """Analyze a batch of mentions for sentiment and tags.

        Args:
            mentions: Up to 50 mentions with id and text.
            context: Brand/competitor context for classification.

        Returns:
            BatchAnalysisOutput with one result per mention.
        """
        prompt = self._build_prompt(mentions, context)
        response = await self.run_async(prompt=prompt)

        if hasattr(response, "content"):
            return response.content
        return response

    def _build_prompt(
        self,
        mentions: list[MentionForAnalysis],
        context: AIContext,
    ) -> str:
        context_block = (
            f"Brand: {context.brand_name}\n"
            f"Brand Keywords: {', '.join(context.brand_keywords)}\n"
            f"Industry: {context.industry}\n"
            f"Competitors:\n"
        )
        for comp in context.competitors:
            context_block += f"  - {comp.name} (keywords: {', '.join(comp.keywords)})\n"

        posts_block = ""
        for i, m in enumerate(mentions, 1):
            # Truncate very long posts to save tokens
            text = m.text[:500] if len(m.text) > 500 else m.text
            posts_block += f"### Post {i} (ID: {m.mention_id})\n{text}\n\n"

        return (
            f"## Brand Context\n{context_block}\n"
            f"## Posts to Analyze ({len(mentions)} total)\n{posts_block}\n"
            f"Analyze each post and return the results as JSON."
        )


def create_mention_analyzer_agent() -> MentionAnalyzerAgent:
    """Factory function to create a MentionAnalyzerAgent instance."""
    return MentionAnalyzerAgent()
```

- [ ] **Step 2: Commit**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-ai-agents
git add src/agents/listening/mention_analyzer.py
git commit -m "feat(listening): add MentionAnalyzerAgent for batch sentiment + tagging"
```

---

## Task 5: Batch Analysis Endpoint

**Project:** `contentstudio-ai-agents`

**Files:**
- Modify: `src/api/routers/listening/listening_router.py`

- [ ] **Step 1: Add the batch analysis endpoint to the listening router**

Add these imports at the top of `src/api/routers/listening/listening_router.py`:

```python
from src.agents.listening.mention_analyzer import create_mention_analyzer_agent
from src.models.listening import (
    AnalyzeBatchRequest,
    AnalyzeBatchResponse,
)
```

Add the singleton and endpoint after the existing `discover_brand` endpoint:

```python
_mention_analyzer_agent = None


def get_mention_analyzer_agent():
    """Get or create mention analyzer agent singleton."""
    global _mention_analyzer_agent
    if _mention_analyzer_agent is None:
        _mention_analyzer_agent = create_mention_analyzer_agent()
        logger.info("Mention analyzer agent initialized")
    return _mention_analyzer_agent


@router.post("/analyze-batch", response_model=AnalyzeBatchResponse)
@trace_api(name="listening.analyze_batch")
async def analyze_mentions_batch(request: AnalyzeBatchRequest) -> AnalyzeBatchResponse:
    """Analyze a batch of mentions for sentiment and AI tags."""
    start_time = time.time()

    try:
        if len(request.mentions) == 0:
            return AnalyzeBatchResponse(success=True, results=[], processing_time=0.0)

        if len(request.mentions) > 50:
            raise HTTPException(
                status_code=422,
                detail={"error": "Batch too large", "message": "Maximum 50 mentions per call"},
            )

        logger.info(f"Analyzing batch of {len(request.mentions)} mentions")
        agent = get_mention_analyzer_agent()
        result = await agent.analyze_batch(request.mentions, request.context)

        processing_time = time.time() - start_time
        logger.info(f"Batch analysis completed in {processing_time:.2f}s for {len(request.mentions)} mentions")

        return AnalyzeBatchResponse(
            success=True,
            results=result.results,
            processing_time=processing_time,
            model_used=get_config().groq_model,
        )

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Batch analysis failed: {e}\n{traceback.format_exc()}")
        raise HTTPException(
            status_code=500,
            detail={"error": "Batch analysis failed", "message": str(e)},
        )
```

- [ ] **Step 2: Verify imports resolve**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-ai-agents
python -c "from src.api.routers.listening.listening_router import router; print(f'Routes: {[r.path for r in router.routes]}')"
```

Expected: `Routes: ['/discover', '/analyze-batch']`

- [ ] **Step 3: Commit**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-ai-agents
git add src/api/routers/listening/listening_router.py
git commit -m "feat(listening): add batch analysis endpoint POST /api/v1/listening/analyze-batch"
```

---

## Task 6: Add ai_context_hint to MongoDB Model

**Project:** `contentstudio-social-analytics-go`

**Files:**
- Modify: `src/models/db/mongo/listening_topic.go`

- [ ] **Step 1: Add the AIContextHint field to ListeningTopic**

In `src/models/db/mongo/listening_topic.go`, add the new field to the `ListeningTopic` struct, after the `UpdatedAt` field and before the Laravel-written fields comment:

```go
	AIContextHint string `bson:"ai_context_hint" json:"ai_context_hint"`
```

The field goes between the existing `UpdatedAt` line and the `// Laravel-written fields` comment.

- [ ] **Step 2: Commit**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-social-analytics-go
git add src/models/db/mongo/listening_topic.go
git commit -m "feat(listening): add ai_context_hint field to ListeningTopic model"
```

---

## Task 7: Add GetAIContextHint to MongoDB Repository

**Project:** `contentstudio-social-analytics-go`

**Files:**
- Modify: `src/db/mongodb/listening.go`

- [ ] **Step 1: Add GetAIContextHint method**

Add this method to the `ListeningRepository` in `src/db/mongodb/listening.go`. This method fetches only the `ai_context_hint` field for a topic, used by the enrichment service:

```go
// GetAIContextHint retrieves the ai_context_hint for a topic.
// Returns empty string if the field is not set or the topic doesn't exist.
func (r *ListeningRepository) GetAIContextHint(ctx context.Context, topicID string) (string, error) {
	objID, err := primitive.ObjectIDFromHex(topicID)
	if err != nil {
		return "", fmt.Errorf("ListeningRepository.GetAIContextHint: invalid topic ID %q: %w", topicID, err)
	}

	var result struct {
		AIContextHint string `bson:"ai_context_hint"`
	}

	err = r.topics.FindOne(ctx, bson.M{"_id": objID}, options.FindOne().SetProjection(bson.M{"ai_context_hint": 1})).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", fmt.Errorf("ListeningRepository.GetAIContextHint: %w", err)
	}

	return result.AIContextHint, nil
}
```

Ensure these imports are present at the top of the file (they likely already are):
- `"go.mongodb.org/mongo-driver/bson/primitive"`
- `"go.mongodb.org/mongo-driver/mongo"`
- `"go.mongodb.org/mongo-driver/mongo/options"`

- [ ] **Step 2: Verify it compiles**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-social-analytics-go/src
go build ./db/mongodb/...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-social-analytics-go
git add src/db/mongodb/listening.go
git commit -m "feat(listening): add GetAIContextHint to ListeningRepository"
```

---

## Task 8: Create Enrichment Service

**Project:** `contentstudio-social-analytics-go`

**Files:**
- Create: `src/services/listening/enrichment/service.go`

This is the core new service. It buffers mentions by topic, flushes batches to the AI agents endpoint, and writes enriched rows back to ClickHouse.

- [ ] **Step 1: Create the enrichment service**

```go
// src/services/listening/enrichment/service.go
package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const (
	defaultBatchSize    = 50
	defaultFlushInterval = 30 * time.Second
	defaultMaxBufferAge  = 5 * time.Minute
	defaultCacheTTL      = 5 * time.Minute
)

// AIAnalyzer abstracts the HTTP call to the AI agents batch endpoint.
type AIAnalyzer interface {
	AnalyzeBatch(ctx context.Context, mentions []MentionPayload, aiContext string) ([]MentionResult, error)
}

// MentionWriter abstracts ClickHouse batch inserts for enriched mentions.
type MentionWriter interface {
	InsertMentions(ctx context.Context, mentions []chmodels.ListeningMentionRow) error
}

// ContextProvider abstracts reading ai_context_hint from MongoDB.
type ContextProvider interface {
	GetAIContextHint(ctx context.Context, topicID string) (string, error)
}

// MentionPayload is sent to the AI agents endpoint.
type MentionPayload struct {
	MentionID string `json:"mention_id"`
	Text      string `json:"text"`
}

// MentionResult is returned from the AI agents endpoint.
type MentionResult struct {
	MentionID      string   `json:"mention_id"`
	SentimentLabel string   `json:"sentiment_label"`
	SentimentScore float64  `json:"sentiment_score"`
	AITags         []string `json:"ai_tags"`
}

// cachedContext holds a topic's AI context with expiry.
type cachedContext struct {
	value     string
	expiresAt time.Time
}

// bufferedMention holds a mention and when it was buffered.
type bufferedMention struct {
	mention    kafkamodels.ListeningMention
	bufferedAt time.Time
}

// EnrichmentService buffers parsed mentions by topic and batch-enriches them
// via the AI agents endpoint, then writes enriched rows to ClickHouse.
type EnrichmentService struct {
	analyzer AIAnalyzer
	writer   MentionWriter
	ctx      ContextProvider
	log      *logger.Logger

	mu      sync.Mutex
	buffers map[string][]bufferedMention // keyed by topic_id

	contextCache map[string]cachedContext
	cacheMu      sync.RWMutex
}

// NewEnrichmentService creates a new EnrichmentService.
func NewEnrichmentService(
	analyzer AIAnalyzer,
	writer MentionWriter,
	ctx ContextProvider,
	log *logger.Logger,
) *EnrichmentService {
	return &EnrichmentService{
		analyzer:     analyzer,
		writer:       writer,
		ctx:          ctx,
		log:          log,
		buffers:      make(map[string][]bufferedMention),
		contextCache: make(map[string]cachedContext),
	}
}

// HandleParsedMention is a kafka.MessageHandler that buffers mentions for batch enrichment.
func (s *EnrichmentService) HandleParsedMention(ctx context.Context, _ string, _ []byte, value []byte) error {
	var mention kafkamodels.ListeningMention
	if err := json.Unmarshal(value, &mention); err != nil {
		return fmt.Errorf("EnrichmentService.HandleParsedMention: unmarshal: %w", err)
	}

	if mention.PostText == "" {
		return nil
	}

	s.mu.Lock()
	s.buffers[mention.TopicID] = append(s.buffers[mention.TopicID], bufferedMention{
		mention:    mention,
		bufferedAt: time.Now(),
	})
	bufLen := len(s.buffers[mention.TopicID])
	s.mu.Unlock()

	if bufLen >= defaultBatchSize {
		s.flushTopic(ctx, mention.TopicID)
	}

	return nil
}

// StartFlushLoop runs a periodic flush of all topic buffers.
// Call this in a goroutine. It stops when ctx is cancelled.
func (s *EnrichmentService) StartFlushLoop(ctx context.Context) {
	ticker := time.NewTicker(defaultFlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.flushAll(context.Background())
			return
		case <-ticker.C:
			s.flushAll(ctx)
		}
	}
}

// flushAll flushes all topic buffers, sending aged-out mentions to DLQ.
func (s *EnrichmentService) flushAll(ctx context.Context) {
	s.mu.Lock()
	topicIDs := make([]string, 0, len(s.buffers))
	for topicID := range s.buffers {
		topicIDs = append(topicIDs, topicID)
	}
	s.mu.Unlock()

	for _, topicID := range topicIDs {
		s.flushTopic(ctx, topicID)
	}
}

// flushTopic enriches and persists all buffered mentions for a topic.
func (s *EnrichmentService) flushTopic(ctx context.Context, topicID string) {
	s.mu.Lock()
	items := s.buffers[topicID]
	if len(items) == 0 {
		s.mu.Unlock()
		return
	}
	delete(s.buffers, topicID)
	s.mu.Unlock()

	log := s.log.With().Str("topic_id", topicID).Int("batch_size", len(items)).Logger()

	// Check max buffer age — if any item is too old, log a warning
	for _, item := range items {
		if time.Since(item.bufferedAt) > defaultMaxBufferAge {
			log.Warn().Msg("Mentions exceeded max buffer age, enriching now")
			break
		}
	}

	// Get AI context for this topic (cached)
	aiContext := s.getAIContext(ctx, topicID)

	// Build payloads for AI endpoint
	payloads := make([]MentionPayload, len(items))
	for i, item := range items {
		payloads[i] = MentionPayload{
			MentionID: item.mention.MentionID,
			Text:      item.mention.PostText,
		}
	}

	// Call AI agents
	results, err := s.analyzer.AnalyzeBatch(ctx, payloads, aiContext)
	if err != nil {
		log.Error().Err(err).Msg("AI batch analysis failed, re-buffering eligible mentions")
		// Re-buffer only items that haven't exceeded max buffer age; drop the rest
		s.mu.Lock()
		for _, item := range items {
			if time.Since(item.bufferedAt) < defaultMaxBufferAge {
				s.buffers[topicID] = append(s.buffers[topicID], item)
			} else {
				log.Warn().Str("mention_id", item.mention.MentionID).Msg("Dropping aged mention from enrichment buffer (already stored without enrichment)")
			}
		}
		s.mu.Unlock()
		return
	}

	// Build result lookup
	resultMap := make(map[string]MentionResult, len(results))
	for _, r := range results {
		resultMap[r.MentionID] = r
	}

	// Merge results into full ClickHouse rows
	rows := make([]chmodels.ListeningMentionRow, 0, len(items))
	for _, item := range items {
		m := item.mention
		if r, ok := resultMap[m.MentionID]; ok {
			m.SentimentLabel = r.SentimentLabel
			m.SentimentScore = r.SentimentScore
			m.AITags = r.AITags
		}
		m.UpdatedAt = time.Now().UTC()
		rows = append(rows, mentionToRow(m))
	}

	// Write to ClickHouse
	if err := s.writer.InsertMentions(ctx, rows); err != nil {
		log.Error().Err(err).Msg("ClickHouse insert failed for enriched mentions")
		return
	}

	log.Info().
		Int("enriched", len(rows)).
		Msg("Flushed enriched mentions to ClickHouse")
}

// getAIContext retrieves the ai_context_hint for a topic, with in-memory caching.
func (s *EnrichmentService) getAIContext(ctx context.Context, topicID string) string {
	s.cacheMu.RLock()
	if cached, ok := s.contextCache[topicID]; ok && time.Now().Before(cached.expiresAt) {
		s.cacheMu.RUnlock()
		return cached.value
	}
	s.cacheMu.RUnlock()

	hint, err := s.ctx.GetAIContextHint(ctx, topicID)
	if err != nil {
		s.log.Warn().Err(err).Str("topic_id", topicID).Msg("Failed to get AI context hint")
		return ""
	}

	s.cacheMu.Lock()
	s.contextCache[topicID] = cachedContext{
		value:     hint,
		expiresAt: time.Now().Add(defaultCacheTTL),
	}
	s.cacheMu.Unlock()

	return hint
}

// mentionToRow converts a Kafka mention to a ClickHouse row.
func mentionToRow(m kafkamodels.ListeningMention) chmodels.ListeningMentionRow {
	return chmodels.ListeningMentionRow{
		MentionID:         m.MentionID,
		TopicID:           m.TopicID,
		Platform:          m.Platform,
		NativeID:          m.NativeID,
		ContentHash:       m.ContentHash,
		AuthorID:          m.AuthorID,
		AuthorName:        m.AuthorName,
		AuthorHandle:      m.AuthorHandle,
		AuthorImageURL:    m.AuthorImageURL,
		AuthorURL:         m.AuthorURL,
		AuthorFollowers:   m.AuthorFollowers,
		PostText:          m.PostText,
		Language:          m.Language,
		PostedAt:          m.PostedAt,
		MatchedKeywords:   m.MatchedKeywords,
		TotalEngagement:   m.TotalEngagement,
		LikesCount:        m.LikesCount,
		CommentsCount:     m.CommentsCount,
		SharesCount:       m.SharesCount,
		ContentType:       m.ContentType,
		MediaType:         m.MediaType,
		URL:               m.URL,
		MediaURLs:         m.MediaURLs,
		AITags:            m.AITags,
		SentimentLabel:    m.SentimentLabel,
		SentimentScore:    m.SentimentScore,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
		PostRead:          m.PostRead,
		PostIrrelevant:    m.PostIrrelevant,
		Bookmark:          m.Bookmark,
		SentimentOverride: m.SentimentOverride,
	}
}
```

- [ ] **Step 2: Create the AI analyzer HTTP client**

Create `src/services/listening/enrichment/analyzer.go`:

```go
package enrichment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// AgentAnalyzer calls the AI agents batch analysis endpoint over HTTP.
type AgentAnalyzer struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     zerolog.Logger
}

// NewAgentAnalyzer creates a new AgentAnalyzer.
func NewAgentAnalyzer(baseURL string, apiKey string, timeout int, logger zerolog.Logger) *AgentAnalyzer {
	if timeout <= 0 {
		timeout = 300
	}
	return &AgentAnalyzer{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		logger: logger.With().Str("component", "ai-batch-analyzer").Logger(),
	}
}

// batchRequest mirrors the AI agents AnalyzeBatchRequest.
type batchRequest struct {
	Mentions []MentionPayload       `json:"mentions"`
	Context  map[string]interface{} `json:"context"`
}

// batchResponse mirrors the AI agents AnalyzeBatchResponse.
type batchResponse struct {
	Success bool            `json:"success"`
	Results []MentionResult `json:"results"`
}

// AnalyzeBatch sends a batch of mentions to the AI agents endpoint for enrichment.
// aiContextHint is the raw JSON string from MongoDB's ai_context_hint field.
func (a *AgentAnalyzer) AnalyzeBatch(ctx context.Context, mentions []MentionPayload, aiContextHint string) ([]MentionResult, error) {
	var aiContext map[string]interface{}
	if aiContextHint != "" {
		if err := json.Unmarshal([]byte(aiContextHint), &aiContext); err != nil {
			a.logger.Warn().Err(err).Msg("Failed to parse ai_context_hint, proceeding without context")
			aiContext = map[string]interface{}{}
		}
	}
	if aiContext == nil {
		aiContext = map[string]interface{}{
			"brand_name":     "",
			"brand_keywords": []string{},
			"competitors":    []interface{}{},
			"industry":       "",
		}
	}

	reqBody := batchRequest{
		Mentions: mentions,
		Context:  aiContext,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("AgentAnalyzer.AnalyzeBatch: marshal: %w", err)
	}

	url := a.baseURL + "/api/v1/listening/analyze-batch"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("AgentAnalyzer.AnalyzeBatch: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if a.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.apiKey)
	}

	a.logger.Debug().Int("batch_size", len(mentions)).Msg("Sending batch to AI agents")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AgentAnalyzer.AnalyzeBatch: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("AgentAnalyzer.AnalyzeBatch: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("AgentAnalyzer.AnalyzeBatch: AI agents returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result batchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("AgentAnalyzer.AnalyzeBatch: unmarshal response: %w", err)
	}

	return result.Results, nil
}
```

- [ ] **Step 3: Verify it compiles**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-social-analytics-go/src
go build ./services/listening/enrichment/...
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-social-analytics-go
git add src/services/listening/enrichment/
git commit -m "feat(listening): add enrichment service with buffered batch AI analysis"
```

---

## Task 9: Rewire the Pipeline

**Project:** `contentstudio-social-analytics-go`

**Files:**
- Modify: `src/services/listening/listening-consumer/main.go`

This task removes the sentiment stage, rewires sink to consume from `listening-parsed`, and adds the enrichment stage.

- [ ] **Step 1: Remove sentiment import and add enrichment import**

In `src/services/listening/listening-consumer/main.go`, replace the sentiment import:

```go
// Remove this line:
"github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/sentiment"
// Add this line:
"github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/enrichment"
```

- [ ] **Step 2: Remove the sentimentGroup constant**

Replace the consumer group constants block:

```go
const (
	fetcherGroup     = "listening-fetcher-group"
	parserGroup      = "listening-parser-group"
	sinkGroup        = "listening-sink-group"
	enrichmentGroup  = "listening-enrichment-group"
)
```

(Removed `sentimentGroup`, added `enrichmentGroup`.)

- [ ] **Step 3: Replace sentiment service setup with enrichment service setup**

In `main()`, find these lines:

```go
	var sentimentAgent sentiment.SentimentAnalyzer
	if cfg.AIAgents.BaseURL != "" {
		sentimentAgent = ai.NewAgentClient(&cfg.AIAgents, log.Logger)
	} else {
		log.Warn().Msg("AI_AGENTS_BASE_URL not set; sentiment stage will pass through without enrichment")
	}
```

Replace with:

```go
	var batchAnalyzer enrichment.AIAnalyzer
	if cfg.AIAgents.BaseURL != "" {
		batchAnalyzer = enrichment.NewAgentAnalyzer(cfg.AIAgents.BaseURL, cfg.AIAgents.APIKey, cfg.AIAgents.Timeout, log.Logger)
	} else {
		log.Warn().Msg("AI_AGENTS_BASE_URL not set; enrichment stage will be disabled")
	}
```

- [ ] **Step 4: Replace sentiment service creation with enrichment service creation**

Find this line:

```go
	sentimentService := sentiment.NewSentimentService(sentimentAgent, producer, log)
```

Replace with:

```go
	enrichmentService := enrichment.NewEnrichmentService(batchAnalyzer, listeningWriter, listeningRepo, log)
```

- [ ] **Step 5: Replace sentiment consumer with enrichment consumer**

Find these lines:

```go
	sentimentConsumer, err := kafka.NewConsumer(cfg.Kafka, sentimentGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create sentiment consumer")
	}
	defer sentimentConsumer.Close()
```

Replace with:

```go
	enrichmentConsumer, err := kafka.NewConsumer(cfg.Kafka, enrichmentGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create enrichment consumer")
	}
	defer enrichmentConsumer.Close()
```

- [ ] **Step 6: Rewire the stages**

Find these four `startStage` calls:

```go
	startStage(&wg, runCtx, log, "fetcher", fetchConsumer, kafkamodels.TopicListeningWork, fetcherService.HandleWorkOrder)
	startStage(&wg, runCtx, log, "parser", parserConsumer, kafkamodels.TopicListeningRaw, parserService.HandleRawPayload)
	startStage(&wg, runCtx, log, "sentiment", sentimentConsumer, kafkamodels.TopicListeningParsed, sentimentService.HandleParsedMention)
	startStage(&wg, runCtx, log, "sink", sinkConsumer, kafkamodels.TopicListeningEnriched, sinkService.HandleEnrichedMention)
```

Replace with:

```go
	startStage(&wg, runCtx, log, "fetcher", fetchConsumer, kafkamodels.TopicListeningWork, fetcherService.HandleWorkOrder)
	startStage(&wg, runCtx, log, "parser", parserConsumer, kafkamodels.TopicListeningRaw, parserService.HandleRawPayload)
	startStage(&wg, runCtx, log, "sink", sinkConsumer, kafkamodels.TopicListeningParsed, sinkService.HandleEnrichedMention)
	startStage(&wg, runCtx, log, "enrichment", enrichmentConsumer, kafkamodels.TopicListeningParsed, enrichmentService.HandleParsedMention)

	// Start the enrichment flush loop in the background
	wg.Add(1)
	go func() {
		defer wg.Done()
		enrichmentService.StartFlushLoop(runCtx)
	}()
```

Key changes:
- Sentiment stage removed entirely.
- Sink now consumes from `TopicListeningParsed` instead of `TopicListeningEnriched`.
- Enrichment stage added on `TopicListeningParsed` (separate consumer group).
- Enrichment flush loop started as a background goroutine.

- [ ] **Step 7: Remove the `ai` service import if no longer used**

Check if the `ai` package import (`"github.com/d4interactive/contentstudio-social-analytics-go/src/services/ai"`) is still used elsewhere in the file. If not, remove it.

- [ ] **Step 8: Verify it compiles**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-social-analytics-go/src
go build ./services/listening/listening-consumer/...
```

Expected: no errors.

- [ ] **Step 9: Commit**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-social-analytics-go
git add src/services/listening/listening-consumer/main.go
git commit -m "refactor(listening): replace sentiment stage with async batch enrichment

Remove one-by-one sentiment stage. Sink now consumes listening-parsed
directly. New enrichment service buffers mentions and batch-enriches
via AI agents endpoint."
```

---

## Task 10: Integration Verification

**Project:** Both

- [ ] **Step 1: Verify Go project compiles fully**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-social-analytics-go/src
go build ./...
```

Expected: no errors.

- [ ] **Step 2: Verify AI agents project imports resolve**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-ai-agents
python -c "
from src.models.listening import BrandDiscoveryRequest, AnalyzeBatchRequest, LISTENING_TAGS
from src.agents.listening.brand_discovery import create_brand_discovery_agent
from src.agents.listening.mention_analyzer import create_mention_analyzer_agent
from src.api.routers.listening.listening_router import router
print(f'Models OK, Tags: {len(LISTENING_TAGS)}')
print(f'Routes: {[r.path for r in router.routes]}')
print('All imports OK')
"
```

Expected:
```
Models OK, Tags: 10
Routes: ['/discover', '/analyze-batch']
All imports OK
```

- [ ] **Step 3: Verify no unused imports in Go**

```bash
cd /home/zaid-bin-tariq/Projects/contentstudio-social-analytics-go/src
go vet ./services/listening/...
```

Expected: no errors.
