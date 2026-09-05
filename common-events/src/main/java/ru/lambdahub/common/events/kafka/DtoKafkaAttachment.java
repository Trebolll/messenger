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
public class DtoKafkaAttachment implements KafkaDto {

    private UUID requestId;
    private UUID attachmentId;
    private UUID objectId;
    private UUID chatId;
    private UUID messageId;
    private String url;
    private String mime;

    @Override
    public Object kafkaKey() {
        return chatId != null ? chatId : objectId;
    }
}
