package ru.lambdahub.media.dto;

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
public class AttachmentDto {
    private UUID id;
    private UUID objectId;
    private UUID chatId;
    private UUID messageId;
    private String kind;
}
