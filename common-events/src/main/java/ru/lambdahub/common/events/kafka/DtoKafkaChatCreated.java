package ru.lambdahub.common.events.kafka;

import java.util.List;
import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class DtoKafkaChatCreated implements KafkaDto {

    private UUID requestId;
    private UUID chatId;
    private String type;
    private String name;
    private List<UUID> memberIds;

    @Override
    public Object kafkaKey() {
        return chatId;
    }
}
