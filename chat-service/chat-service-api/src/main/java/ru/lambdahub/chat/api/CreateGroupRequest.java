package ru.lambdahub.chat.api;

import jakarta.annotation.Nullable;
import jakarta.validation.constraints.NotBlank;
import java.util.List;
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
public class CreateGroupRequest extends InboundMessage {

    @NotBlank
    private String name;

    @Nullable
    private List<String> usernames;
}
