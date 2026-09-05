package ru.lambdahub.user.api;

import java.util.UUID;
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
public class ProfileResponse extends OutboundMessage {

    private UUID id;
    private String username;
    private String displayName;
    private String email;
    private String phone;
    private String avatarUrl;
    private String statusText;
    private String profession;
    private String location;
}
