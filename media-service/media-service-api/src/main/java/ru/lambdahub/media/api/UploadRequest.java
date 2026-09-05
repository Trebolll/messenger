package ru.lambdahub.media.api;

import jakarta.annotation.Nullable;
import jakarta.validation.constraints.NotNull;
import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.ToString;
import lombok.experimental.FieldNameConstants;
import lombok.experimental.SuperBuilder;
import org.springframework.web.multipart.MultipartFile;
import ru.lambdahub.validation.message.InboundMessage;

@Getter
@ToString(callSuper = true)
@SuperBuilder(toBuilder = true)
@AllArgsConstructor
@NoArgsConstructor
@FieldNameConstants
@EqualsAndHashCode(callSuper = false)
public class UploadRequest extends InboundMessage {

    @NotNull
    private UUID userId;

    @NotNull
    private MultipartFile file;

    @Nullable
    private UUID chatId;

    @Nullable
    private UUID messageId;
}
