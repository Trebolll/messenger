package ru.lambdahub.chat.dto;

import java.time.Instant;
import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.experimental.FieldNameConstants;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@FieldNameConstants
public class MessageReadDto {
    private UUID chatId;
    private UUID userId;
    private UUID lastReadMessageId;
    private Instant readAt;
}
