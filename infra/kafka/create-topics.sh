#!/bin/bash
set -e
kafka-topics.sh --bootstrap-server kafka:9092 --create --if-not-exists --topic user.created --partitions 6 --replication-factor 1
kafka-topics.sh --bootstrap-server kafka:9092 --create --if-not-exists --topic user.profile_updated --partitions 6 --replication-factor 1
kafka-topics.sh --bootstrap-server kafka:9092 --create --if-not-exists --topic user.status_changed --partitions 6 --replication-factor 1
kafka-topics.sh --bootstrap-server kafka:9092 --create --if-not-exists --topic chat.created --partitions 6 --replication-factor 1
kafka-topics.sh --bootstrap-server kafka:9092 --create --if-not-exists --topic chat.member_added --partitions 6 --replication-factor 1
kafka-topics.sh --bootstrap-server kafka:9092 --create --if-not-exists --topic chat.member_removed --partitions 6 --replication-factor 1
kafka-topics.sh --bootstrap-server kafka:9092 --create --if-not-exists --topic message.created --partitions 6 --replication-factor 1
kafka-topics.sh --bootstrap-server kafka:9092 --create --if-not-exists --topic message.edited --partitions 6 --replication-factor 1
kafka-topics.sh --bootstrap-server kafka:9092 --create --if-not-exists --topic message.deleted --partitions 6 --replication-factor 1
kafka-topics.sh --bootstrap-server kafka:9092 --create --if-not-exists --topic messages.read --partitions 6 --replication-factor 1
kafka-topics.sh --bootstrap-server kafka:9092 --create --if-not-exists --topic attachment.created --partitions 6 --replication-factor 1
echo "Kafka topics ready"
