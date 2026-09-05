package ru.lambdahub.chat.api;

import java.time.Instant;
import java.util.List;
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
public class ChatListItem extends OutboundMessage {

    private UUID id;
    private String type;
    private String name;
    private String avatarUrl;
    private String lastMessage;
    private Instant lastMessageAt;
    private List<UUID> memberIds;
}
