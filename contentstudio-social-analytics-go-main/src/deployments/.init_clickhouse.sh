#!/bin/bash

set -euo pipefail

CLICKHOUSE_HOST=${CLICKHOUSE_HOST:-clickhouse}
CLICKHOUSE_USER=${CLICKHOUSE_USER:-default}
CLICKHOUSE_PASSWORD=${CLICKHOUSE_PASSWORD:-contentstudio123}
CLICKHOUSE_DATABASE=${CLICKHOUSE_DATABASE:-contentstudiobackend}

echo "⏳ Waiting for ClickHouse to be ready..."

# Wait for ClickHouse to be accessible
until clickhouse-client --host "${CLICKHOUSE_HOST}" --user "${CLICKHOUSE_USER}" --password "${CLICKHOUSE_PASSWORD}" --query "SELECT 1" &> /dev/null; do
  echo "Waiting..."
  sleep 2
done

echo "✅ ClickHouse is up!"

clickhouse-client --host "${CLICKHOUSE_HOST}" --user "${CLICKHOUSE_USER}" --password "${CLICKHOUSE_PASSWORD}" --query "CREATE DATABASE IF NOT EXISTS ${CLICKHOUSE_DATABASE}"

# Run all .sql files
for f in /scripts/sql/*.sql; do
  echo "▶️ Running $f"
  clickhouse-client --host "${CLICKHOUSE_HOST}" --user "${CLICKHOUSE_USER}" --password "${CLICKHOUSE_PASSWORD}" --database "${CLICKHOUSE_DATABASE}" --multiquery < "$f"
done
