package ru.lambdahub.common.events.kafka;

import com.fasterxml.jackson.annotation.JsonIgnore;

/**
 * Kafka payload marker — {@link ru.lambdahub.common.kafka.OutputBridge} uses {@link #kafkaKey()} as partition key.
 */
public interface KafkaDto {

    @JsonIgnore
    Object kafkaKey();
}
