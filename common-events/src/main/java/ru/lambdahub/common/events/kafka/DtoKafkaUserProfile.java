package ru.lambdahub.common.events.kafka;

import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

/** Kafka: user profile updated. */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class DtoKafkaUserProfile implements KafkaDto {

    private UUID requestId;
    private UUID userId;
    private String username;
    private String displayName;
    private String avatarUrl;
    private String statusText;

    @Override
    public Object kafkaKey() {
        return userId;
    }
}
