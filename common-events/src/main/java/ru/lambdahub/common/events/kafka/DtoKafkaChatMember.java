package ru.lambdahub.common.events.kafka;

import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class DtoKafkaChatMember implements KafkaDto {

    private UUID requestId;
    private UUID chatId;
    private UUID userId;
    private UUID addedBy;

    @Override
    public Object kafkaKey() {
        return chatId;
    }
}
