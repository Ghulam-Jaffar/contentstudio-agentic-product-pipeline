# ContentStudio Social Analytics Pipeline

## Overview

This project is a robust data pipeline designed to fetch, parse, and store social media data from various platforms into a ClickHouse database. It leverages Kafka for message queuing and MongoDB for auxiliary data storage or scheduling. The pipeline is built using Go and is structured into microservices for each stage of data processing for different social media platforms.

## Key Features

*   **Multi-Platform Support**: Designed to integrate with Facebook, Instagram, LinkedIn, Pinterest, TikTok, Twitter, and YouTube.
*   **Scalable Architecture**: Utilizes Kafka for decoupling services and handling high throughput.
*   **Efficient Data Storage**: Leverages ClickHouse for fast analytics queries on large datasets.
*   **Modular Design**: Each platform and processing step (fetch, parse, sink) is a separate microservice.
*   **Configuration Management**: Uses Viper for easy configuration via environment variables.
*   **Observability**: Includes support for Prometheus metrics and structured logging with Zerolog.

## Tech Stack

*   **Programming Language**: Go
*   **Messaging Queue**: Apache Kafka (via `franz-go`)
*   **Primary Data Store (Analytics)**: ClickHouse (via `clickhouse-go`)
*   **Auxiliary Data Store/Scheduler**: MongoDB (via `mongo-driver`)
*   **Configuration**: Viper
*   **Logging**: Zerolog
*   **Metrics**: Prometheus (`client_golang`)
*   **API/Web Framework (if needed for internal tools/services)**: Gin
*   **Containerization (for dev stack)**: Docker, Docker Compose

## Documentation

### Interactive Documentation

The project includes a modern, interactive documentation site built with [Fumadocs](https://fumadocs.vercel.app) and Next.js. 

#### Running the Documentation Locally

1. **Navigate to the docs directory**:
   ```bash
   cd docs/
   ```

2. **Install dependencies** (first time only):
   ```bash
   npm install
   ```

3. **Start the development server**:
   ```bash
   npm run dev
   ```

4. **Open your browser** and navigate to [http://localhost:3000](http://localhost:3000)

The documentation site provides:
- Interactive navigation and search
- Platform-specific implementation guides
- Architecture diagrams and data flow visualizations
- API reference documentation
- Code examples and best practices

#### Building for Production

To build the documentation for production deployment:
```bash
npm run build
npm start
```

### Key Documentation Resources

- Onboarding guide: `docs/content/docs/guides/engineer-onboarding.mdx`
- Migration plan: `docs/content/docs/planning/migration-plan.mdx`
- Architecture overview: `docs/content/docs/architecture/project-overview-and-code-review.mdx`
- Kafka topic reference: `docs/content/docs/reference/kafka-topics.mdx`
- Platform implementations: `docs/content/docs/platforms/`

## Project Structure

The Go source code and module files are located in the `src/` directory.

*   `src/cmd/`: Contains the main applications for each microservice, scheduler, and job runner.
    *   `jobs/url-refresher/`: Platform-specific job for refreshing stale media URLs and thumbnails in ClickHouse.
    *   `services/`: Individual microservices for fetching, parsing, processing, and sinking data for each platform.
    *   `src/services/listening/`: Core logic for the social listening pipeline.
        *   `listening-scheduler/`: Periodically dispatches sync work orders to Kafka.
        *   `listening-consumer/`: Processes sync work, polls platform APIs, and sinks data to ClickHouse.
*   `src/deployments/`: Docker Compose configurations for local development stack (Kafka, ClickHouse, Mongo, etc.).
*   `src/internal/`: Houses internal application logic.
    *   `clients/`: Clients for interacting with ClickHouse, MongoDB, and social media APIs.
    *   `config/`: Environment variable loading and validation.
    *   `kafka/`: Standardized Kafka consumer and producer wrappers.
    *   `parsing/`: Data parsing logic for each social media platform.
*   `src/pkg/`: Shared libraries and data models.
    *   `models/`: Go structs for ClickHouse tables, Kafka messages, and MongoDB collections.
*   `src/scripts/`: Utility scripts (e.g., creating Kafka topics).
*   `src/sql/`: SQL migration files for ClickHouse.
*   `src/go.mod`, `src/go.sum`: Go module files defining dependencies.
*   `.gitignore`: Specifies intentionally untracked files that Git should ignore.
*   `README.md`: This file.

## Getting Started

### Prerequisites

*   Go (version 1.21 or higher recommended - *check your `go.mod` for the exact version*)
*   Docker and Docker Compose (for running the local development stack)
*   Access keys/tokens for the respective social media APIs

### Setup

1.  **Clone the repository (if applicable)**:
    ```bash
    git clone <your-repository-url>
    cd contentstudio-social-analytics-go
    ```

2.  **Configure Environment Variables**:
    *   Navigate to the `src/` directory.
    *   Copy the `src/.env.example` file to `src/.env`.
    *   Fill in the necessary API keys, database credentials, Kafka broker addresses, and other configurations in `src/.env`.

3.  **Start Local Development Stack**:
    *   Navigate to `src/deployments/`.
    *   Run `docker-compose up -d` to start Kafka, ClickHouse, MongoDB, etc.
    *   See `docs/guides/engineer-onboarding.md` for details.

4.  **Create Kafka Topics**:
    *   Navigate to `src/scripts/`.
    *   Make `create-topics.sh` executable: `chmod +x create-topics.sh`.
    *   Run the script: `./create-topics.sh`.
    *   See topic list in `docs/reference/kafka-topics.md`.

5.  **Run Database Migrations**:
    *   Apply the ClickHouse schema migrations located in `src/sql/clickhouse_schemas/` (per-platform SQL files). Execute them manually against your ClickHouse instance or via your migration tool.
    *   Example (manual, assuming `clickhouse-client` and database `contentstudiobackend`):
        ```bash
        clickhouse-client --query="CREATE DATABASE IF NOT EXISTS contentstudiobackend"
        cat src/sql/clickhouse/facebook_schema.sql | clickhouse-client -d contentstudiobackend --multiquery
        # Repeat for other files in src/sql/schema/
        ```

### Building and Running Services

Each service in `src/cmd/services/` and each job in `src/cmd/jobs/` is a separate Go application.

1.  **Navigate to a service or job directory**:
    ```bash
    cd src/cmd/services/facebook-fetcher
    
    # For the listening consumer (fetcher + sink)
    cd src/services/listening/listening-consumer
    
    # For the listening scheduler
    cd src/services/listening/listening-scheduler
    ```
2.  **Build the service**:
    ```bash
    go build -o ../../../bin/facebook-fetcher main.go
    go build -o ../../../bin/listening-consumer main.go
    ```
    (Consider creating a `bin/` directory in the project root: `/Users/azhar/Dev/ContentStudio/contentstudio-social-analytics-go/bin`)
3.  **Run the service**:
    ```bash
    ../../../bin/facebook-fetcher
    ```
    Or, run directly:
    ```bash
    go run main.go
    ```

Repeat for other services as needed.

To run the URL refresher job, build the `url_refersher` binary from `src/` and pass a platform flag:

```bash
cd src
make url_refersher
../bin/url_refersher -platform all
```

## Development Checklist & Next Steps

This is a starting point for development tasks:

*   **Core Infrastructure:**
    *   [ ] Finalize `src/deployments/docker-compose.yml` to run Kafka, ClickHouse, and MongoDB.
    *   [ ] Implement robust configuration loading in `src/internal/config/config.go`.
    *   [ ] Implement standardized Kafka producer/consumer wrappers in `src/internal/kafka/`.
    *   [ ] Implement ClickHouse client logic (batch inserts, connections) in `src/internal/clients/clickhouse.go`.
    *   [ ] Implement MongoDB client logic in `src/internal/clients/mongodb.go`.
*   **Data Models:**
    *   [ ] Define Go structs for all ClickHouse tables in `src/pkg/models/clickhouse.go`.
    *   [ ] Define Go structs for all Kafka message types (e.g., `WorkOrder`, `RawData`) in `src/pkg/models/kafka.go`.
    *   [ ] Define Go structs for MongoDB collections (e.g., scheduling tasks) in `src/pkg/models/mongo.go`.
*   **Scheduler (`src/cmd/scheduler/`):**
    *   [ ] Implement logic to schedule data fetching tasks (e.g., store schedules in MongoDB, produce work orders to Kafka).
*   **For Each Social Media Platform (Facebook, Instagram, etc.):**
    *   **API Client (`src/internal/clients/social/`):**
        *   [ ] Implement Go client for the platform's API (e.g., `facebook.go`).
    *   **Fetcher Service (`src/cmd/services/<platform>-fetcher/`):**
        *   [ ] Consume work orders from Kafka.
        *   [ ] Use the API client to fetch data.
        *   [ ] Produce raw data to another Kafka topic.
    *   **Parser Service (`src/cmd/services/<platform>-<type>-parser/`):**
        *   [ ] Consume raw data from Kafka.
        *   [ ] Implement parsing logic in `src/internal/parsing/<platform>.go` to transform raw data into ClickHouse-compatible structs.
        *   [ ] Produce parsed data to another Kafka topic.
    *   **Sink Service (`src/cmd/services/<platform>-<type>-sink/`):**
        *   [ ] Consume parsed data from Kafka.
        *   [ ] Batch insert data into the appropriate ClickHouse tables.
*   **Observability:**
    *   [ ] Integrate Zerolog for structured logging in all services.
    *   [ ] Add Prometheus metrics for key operations (e.g., messages processed, API call latencies, errors).
*   **Testing:**
    *   [ ] Write unit tests for parsing logic, client interactions, etc.
    *   [ ] Consider integration tests for pipeline segments.
*   **Documentation:**
    *   [ ] Add detailed documentation for each service and its configuration.
    *   [ ] Document API client usage and rate limiting considerations.

---

This README provides a solid foundation. You can expand on each section as the project develops.
