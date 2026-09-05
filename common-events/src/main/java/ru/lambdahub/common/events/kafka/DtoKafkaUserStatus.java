package ru.lambdahub.common.events.kafka;

import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

/** Kafka: user status changed. */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class DtoKafkaUserStatus implements KafkaDto {

    private UUID requestId;
    private UUID userId;
    private String statusText;

    @Override
    public Object kafkaKey() {
        return userId;
    }
}
