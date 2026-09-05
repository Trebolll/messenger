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
public class MediaObjectDto {
    private UUID id;
    private String bucket;
    private String objectName;
    private String mime;
    private Long sizeBytes;
    private UUID uploadedBy;
}
