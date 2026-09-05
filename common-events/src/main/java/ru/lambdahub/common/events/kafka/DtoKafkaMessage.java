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
public class DtoKafkaMessage implements KafkaDto {

    private UUID requestId;
    private UUID messageId;
    private UUID chatId;
    private UUID senderId;
    private String content;
    private UUID parentId;
    private String createdAt;
    private String editedAt;

    @Override
    public Object kafkaKey() {
        return chatId;
    }
}
