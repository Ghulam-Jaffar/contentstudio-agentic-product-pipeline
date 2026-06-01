#!/bin/sh

# Set default host/port values (can be overridden via env vars)
KAFKA_HOST=${KAFKA_HOST:-kafka}
KAFKA_PORT=${KAFKA_PORT:-9092}

CLICKHOUSE_HOST=${CLICKHOUSE_HOST:-clickhouse}
CLICKHOUSE_PORT=${CLICKHOUSE_PORT:-9000}
CLICKHOUSE_USER=${CLICKHOUSE_USER:-default}
CLICKHOUSE_PASSWORD=${CLICKHOUSE_PASSWORD:-mypassword123}

# Wait for Kafka
echo "⏳ Waiting for Kafka at $KAFKA_HOST:$KAFKA_PORT..."
until nc -z "$KAFKA_HOST" "$KAFKA_PORT"; do
  echo "Kafka not available yet..."
  sleep 2
done
echo "Kafka is up!"

## Wait for ClickHouse
#echo "⏳ Waiting for ClickHouse at $CLICKHOUSE_HOST:$CLICKHOUSE_PORT..."
#until schema-client --host schema --port 9000 --user default --password mypassword123 --query "SELECT 1" &> /dev/null; do
#  echo "ClickHouse not ready yet..."
#  sleep 3
#done
#echo "ClickHouse is ready!"

# Run actual service passed as CMD args
echo "Starting service: $@"
exec "$@"
