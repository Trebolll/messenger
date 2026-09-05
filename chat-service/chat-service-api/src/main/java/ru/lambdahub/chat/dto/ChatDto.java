package ru.lambdahub.chat.dto;

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
public class ChatDto {
    private UUID id;
    private String type;
    private String name;
    private String avatarUrl;
    private UUID createdBy;
}
