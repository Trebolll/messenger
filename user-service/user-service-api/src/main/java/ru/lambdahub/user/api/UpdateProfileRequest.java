package ru.lambdahub.user.api;

import jakarta.annotation.Nullable;
import lombok.AllArgsConstructor;
import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.ToString;
import lombok.experimental.FieldNameConstants;
import lombok.experimental.SuperBuilder;
import ru.lambdahub.validation.message.InboundMessage;

@Getter
@ToString(callSuper = true)
@SuperBuilder(toBuilder = true)
@AllArgsConstructor
@NoArgsConstructor
@FieldNameConstants
@EqualsAndHashCode(callSuper = false)
public class UpdateProfileRequest extends InboundMessage {

    @Nullable
    private String username;
    @Nullable
    private String displayName;
    @Nullable
    private String phone;
    @Nullable
    private String profession;
    @Nullable
    private String location;
    @Nullable
    private String status;
}
