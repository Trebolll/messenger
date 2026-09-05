package ru.lambdahub.common.events.kafka;

import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

/** Kafka / internal: user created (auth → user). */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class DtoKafkaUserCreated implements KafkaDto {

    private UUID requestId;
    private UUID userId;
    private String username;
    private String email;
    private String phone;
    private String fullName;

    @Override
    public Object kafkaKey() {
        return userId;
    }
}
