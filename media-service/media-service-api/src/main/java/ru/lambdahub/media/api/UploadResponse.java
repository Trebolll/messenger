package ru.lambdahub.media.api;

import lombok.AllArgsConstructor;
import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.Setter;
import lombok.NoArgsConstructor;
import lombok.ToString;
import lombok.experimental.FieldNameConstants;
import lombok.experimental.SuperBuilder;
import ru.lambdahub.validation.message.OutboundMessage;

@Getter
@Setter
@NoArgsConstructor
@SuperBuilder(toBuilder = true)
@AllArgsConstructor
@FieldNameConstants
@ToString(callSuper = true)
@EqualsAndHashCode(callSuper = false)
public class UploadResponse extends OutboundMessage {

    private String id;
    private String objectName;
    private String url;
    private String mime;
    private long sizeBytes;
}
