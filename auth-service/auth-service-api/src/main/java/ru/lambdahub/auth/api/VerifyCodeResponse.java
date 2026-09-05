package ru.lambdahub.auth.api;

import lombok.AllArgsConstructor;
import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.ToString;
import lombok.experimental.FieldNameConstants;
import lombok.experimental.SuperBuilder;
import ru.lambdahub.validation.message.OutboundMessage;
@Getter
@NoArgsConstructor
@SuperBuilder(toBuilder = true)
@AllArgsConstructor
@FieldNameConstants
@ToString(callSuper = true)
@EqualsAndHashCode(callSuper = false)
public class VerifyCodeResponse extends OutboundMessage {

    private String confirmToken;
}
