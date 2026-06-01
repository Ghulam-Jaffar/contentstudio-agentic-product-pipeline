# Social Media Data Pipeline (Kafka + ClickHouse + MongoDB)

This project defines a complete social media data processing pipeline using **Docker Compose**.  
It integrates **Kafka** as the event streaming platform, **ClickHouse** as the analytical database, and **MongoDB** as the document store.  
Several custom Go microservices (fetchers, parsers, processors, sinks) are included to handle ingestion, parsing, transformation, and storage of social media data (Facebook, Instagram, LinkedIn, TikTok).

---

## 📋 Prerequisites

Before running this project, ensure you have the following installed:

- [Docker](https://docs.docker.com/get-docker/) >= 20.x
- [Docker Compose](https://docs.docker.com/compose/) >= v2.x
- Adequate system resources (recommended: **4+ CPUs**, **8GB RAM**)
- Access to your Doppler secret (`DOPPLER_SECRET`) if required for your Go binaries

---

---

## 🚀 Services Overview

### Core Infrastructure
- **Kafka** (`apache/kafka:latest`)
    - Event streaming platform.
    - Runs in **KRaft mode** (no ZooKeeper).
    - Exposes port `9092`.
    - Topics are created by the `init-kafka` service.

- **ClickHouse** (`clickhouse/clickhouse-server:latest`)
    - Columnar OLAP database optimized for analytics.
    - Ports:
        - `8123`: HTTP interface
        - `9000`: Native TCP protocol

- **MongoDB** (`mongo:latest`)
    - NoSQL document database.
    - Port: `27017`.
    - Seeded with data via `seed-mongo`.

### Initialization Helpers
- **init-kafka**
    - Waits until Kafka is ready.
    - Runs `.create-topics.sh` to create required Kafka topics.

- **seed-mongo**
    - Seeds MongoDB using `./mongodb/dumps` and `.seed-mongo.sh`.

### Go Microservices (custom binaries)
Each follows a **fetch → parse → process → sink** pattern for multiple platforms:

- **Facebook Services**
    - `facebook_fetcher`
    - `facebook_posts_videos_parser`
    - `facebook_immediate_processor`
    - `facebook_clickhouse_sink`

- **Instagram Services**
    - `instagram_fetcher`
    - `instagram_parser`
    - `instagram_immediate_processor`
    - `instagram_clickhouse_sink`

- **LinkedIn Services**
    - `linkedin_fetcher`
    - `linkedin_parser`
    - `linkedin_immediate_processor`
    - `linkedin_clickhouse_sink`

- **TikTok Services**
    - `tiktok_fetcher`
    - `tiktok_parser`
    - `tiktok_immediate_processor`
    - `tiktok_clickhouse_sink`

- **Immediate Work API**
    - Exposes REST API at `http://localhost:8080`.

---

## ⚙️ Environment Variables

The following variables are used (defined in your environment or `.env` file):

| Variable        | Description                          |
|-----------------|--------------------------------------|
| `DOPPLER_SECRET` | Secret token for Doppler integration |

---

## ▶️ Steps to Run

1. **Cd into deployments directory**

2. **Commands to Run**
   ```bash
   echo "DOPPLER_SECRET=your_secret_here" > .env
   chmod +x .create-topics.sh
   chmod +x .init_clickhouse.sh
   chmod +x mongodb/.seed-mongo.sh
   docker compose up --build



