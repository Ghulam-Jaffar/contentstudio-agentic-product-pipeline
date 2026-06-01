---
trigger: manual
---

The Big Picture: Our Architecture

Our analytics pipeline is a highly specialized assembly line for social media data, built in Go for performance and reliability. Data flows in one direction only, with Apache Kafka as the central conveyor belt that connects each stage. This ensures our services are decoupled, independently scalable, and resilient.

    The Flow: Mongo → Scheduler → Kafka → Fetcher → Kafka → Parser → Kafka → Sink → ClickHouse.
    The Rule: Services never communicate directly with each other. They only consume from and produce to Kafka. This is the golden rule that keeps the system simple and scalable.

Building Production-Grade Services

Every microservice in our cmd/ directory must be a well-behaved, production-ready citizen. This means it must be observable, configurable, and resilient.

    Rule: Measure Everything (Metrics)
    We use the prometheus/client_golang library to track the health of our services. Every service must expose a /metrics endpoint. Don't just assume a service is working; prove it with data. Key metrics to track include Kafka consumer lag, message throughput, and error rates.

    Rule: Trace the Journey (Tracing)
    We use OpenTelemetry (otel) to trace requests as they move through the pipeline. When a fetcher is slow, we need to know if the problem is the external API, Kafka, or our own code. Tracing connects the dots and makes debugging complex flows possible.

    Rule: Know Your Health (Health Checks)
    Every service must respond to HTTP requests on a /healthz endpoint. A 200 OK response means the service is alive and ready to work. This allows our orchestration system (like Kubernetes) to automatically manage the health of the pipeline.

    Rule: Exit Gracefully (Graceful Shutdown)
    A service must never die unexpectedly and lose data. It must listen for system signals (like SIGINT when a user presses Ctrl+C) and perform a graceful shutdown. This means it should finish processing any in-flight messages, close database connections, and shut down its Kafka clients cleanly before exiting.

    Rule: Manage Your Concurrency
    Go is great at doing many things at once, but we are not allowed to overwhelm our APIs or databases. When processing from Kafka, use a worker pool pattern to control the number of concurrent goroutines. This ensures we process messages as fast as possible without causing cascading failures.

    Rule: Unify Your Configuration
    We use Viper to handle configuration. Hardcoding is forbidden. All settings (database URLs, API keys, topic names) must be loaded from environment variables or configuration files. This allows us to promote the exact same code from development to staging to production without any changes.

Our Development Pact

These are the core agreements that ensure our code remains high-quality, consistent, and easy to maintain.

    Rule: The Data Pact is Law
    This is our most important rule. If you need to change a data structure (e.g., add a views_count field), you must update it in three places simultaneously:
        The database schema in sql/clickhouse/.
        The Go struct in pkg/models/clickhouse.go.
        The transformation logic in the relevant internal/parsing/ file. Failure to keep these in sync breaks the entire pipeline.

    Rule: Quality is Automatic (Linting & Testing)
    We use golangci-lint to automatically enforce code style and catch common bugs before they are even committed. Furthermore, all new logic must be accompanied by meaningful tests using the testify toolkit. Your code isn't done until the tests pass and the linter is happy.

    Rule: Structured Logging is Non-Negotiable
    We use zerolog for structured JSON logging. Always add relevant context to your log messages (.Str("account_id", id), .Str("topic", topicName)). This transforms logs from a messy stream of text into a searchable, filterable database for debugging.
