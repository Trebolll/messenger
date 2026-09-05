package ru.lambdahub.auth.api;

import jakarta.validation.constraints.NotBlank;
import lombok.AllArgsConstructor;
import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.ToString;
import lombok.experimental.FieldNameConstants;
import lombok.experimental.SuperBuilder;
import ru.lambdahub.validation.message.InboundMessage;
import jakarta.annotation.Nullable;
@Getter
@ToString(callSuper = true)
@SuperBuilder(toBuilder = true)
@AllArgsConstructor
@NoArgsConstructor
@FieldNameConstants
@EqualsAndHashCode(callSuper = false)
public class RegisterRequest extends InboundMessage {

    @NotBlank
    private String confirmToken;

    @NotBlank
    private String username;

    @NotBlank
    private String password;

    @Nullable
    private String fullName;
}
