#!/bin/bash

set -e

echo "Waiting for Kafka to be ready..."
cub kafka-ready -b broker:9092 1 60

echo "Creating topics..."
declare -a topics=(
  "predictions:1:1"
  "predictions_delete:1:1"
)

for topic in "${topics[@]}"; do
  IFS=':' read -r -a config <<< "$topic"
  topic_name="${config[0]}"
  partitions="${config[1]}"
  replication="${config[2]}"

  echo "Creating topic: $topic_name"
  kafka-topics --bootstrap-server broker:9092 \
    --create \
    --if-not-exists \
    --topic "$topic_name" \
    --partitions "$partitions" \
    --replication-factor "$replication"
done

echo "Topics created successfully!"