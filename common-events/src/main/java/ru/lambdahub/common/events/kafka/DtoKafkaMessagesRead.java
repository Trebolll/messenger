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
public class DtoKafkaMessagesRead implements KafkaDto {

    private UUID requestId;
    private UUID chatId;
    private UUID readerId;
    private UUID upToMessageId;
    private String readAt;

    @Override
    public Object kafkaKey() {
        return chatId;
    }
}
