#!/bin/bash

/opt/kafka/bin/kafka-server-start.sh /opt/kafka/config/server.properties &
kafka_pid=$!

echo "Waiting for Kafka to start on port 9092..."
while ! nc -z localhost 9092; do
  sleep 1
done

echo "Kafka started, creating topics..."

bash /usr/bin/create-topics.sh

wait $kafka_pid
