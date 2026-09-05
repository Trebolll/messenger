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
public class MessageDto {
    private UUID id;
    private UUID chatId;
    private UUID senderId;
    private String content;
    private UUID parentId;
    private Instant createdAt;
    private Instant editedAt;
    private Instant deletedAt;
}
