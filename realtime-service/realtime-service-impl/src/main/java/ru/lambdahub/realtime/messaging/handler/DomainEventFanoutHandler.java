package ru.lambdahub.realtime.messaging.handler;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Component;
import ru.lambdahub.common.events.RedisChannels;
import ru.lambdahub.common.events.kafka.DtoKafkaAttachment;
import ru.lambdahub.common.events.kafka.DtoKafkaChatCreated;
import ru.lambdahub.common.events.kafka.DtoKafkaChatMember;
import ru.lambdahub.common.events.kafka.DtoKafkaMessage;
import ru.lambdahub.common.events.kafka.DtoKafkaMessagesRead;
import ru.lambdahub.common.events.kafka.DtoKafkaUserProfile;
import ru.lambdahub.common.events.kafka.DtoKafkaUserStatus;
import ru.lambdahub.common.events.kafka.KafkaDto;

@Slf4j
@Component
@RequiredArgsConstructor
public class DomainEventFanoutHandler {

    private final StringRedisTemplate redis;
    private final ObjectMapper mapper = new ObjectMapper();

    public void onMessageCreated(DtoKafkaMessage dto) {
        fanoutMessage(dto, "new_message");
    }

    public void onMessageEdited(DtoKafkaMessage dto) {
        fanoutMessage(dto, "message_edited");
    }

    public void onMessageDeleted(DtoKafkaMessage dto) {
        fanoutMessage(dto, "message_deleted");
    }

    public void onMessagesRead(DtoKafkaMessagesRead dto) {
        switch (dto) {
            case null -> log.warn("Empty DtoKafkaMessagesRead");
            case DtoKafkaMessagesRead r when r.getChatId() != null ->
                    publish(RedisChannels.chat(r.getChatId().toString()), "messages_read", r);
            case DtoKafkaMessagesRead r ->
                    log.warn("DtoKafkaMessagesRead without chatId requestId={}", r.getRequestId());
        }
    }

    public void onChatMemberAdded(DtoKafkaChatMember dto) {
        fanoutChatMember(dto, "member_added");
    }

    public void onChatMemberRemoved(DtoKafkaChatMember dto) {
        fanoutChatMember(dto, "member_removed");
    }

    public void onChatCreated(DtoKafkaChatCreated dto) {
        switch (dto) {
            case null -> log.warn("Empty DtoKafkaChatCreated");
            case DtoKafkaChatCreated c when c.getMemberIds() != null && !c.getMemberIds().isEmpty() -> {
                String json = toWsJson("chat_created", c);
                for (UUID memberId : c.getMemberIds()) {
                    if (memberId != null) {
                        redis.convertAndSend(RedisChannels.user(memberId.toString()), json);
                    }
                }
            }
            case DtoKafkaChatCreated c ->
                    log.warn("DtoKafkaChatCreated without members chatId={}", c.getChatId());
        }
    }

    public void onUserProfileUpdated(DtoKafkaUserProfile dto) {
        switch (dto) {
            case null -> log.warn("Empty DtoKafkaUserProfile");
            case DtoKafkaUserProfile p when p.getUserId() != null ->
                    publish(RedisChannels.user(p.getUserId().toString()), "user_profile_updated", p);
            case DtoKafkaUserProfile ignored -> log.warn("DtoKafkaUserProfile without userId");
        }
    }

    public void onUserStatusChanged(DtoKafkaUserStatus dto) {
        switch (dto) {
            case null -> log.warn("Empty DtoKafkaUserStatus");
            case DtoKafkaUserStatus s when s.getUserId() != null ->
                    publish(RedisChannels.presence(s.getUserId().toString()), "user_status", s);
            case DtoKafkaUserStatus ignored -> log.warn("DtoKafkaUserStatus without userId");
        }
    }

    public void onAttachmentCreated(DtoKafkaAttachment dto) {
        switch (dto) {
            case null -> log.warn("Empty DtoKafkaAttachment");
            case DtoKafkaAttachment a when a.getChatId() != null ->
                    publish(RedisChannels.chat(a.getChatId().toString()), "attachment_created", a);
            case DtoKafkaAttachment a ->
                    log.warn("DtoKafkaAttachment without chatId attachmentId={}", a.getAttachmentId());
        }
    }

    private void fanoutMessage(DtoKafkaMessage dto, String wsType) {
        switch (dto) {
            case null -> log.warn("Empty DtoKafkaMessage type={}", wsType);
            case DtoKafkaMessage m when m.getChatId() != null ->
                    publish(RedisChannels.chat(m.getChatId().toString()), wsType, m);
            case DtoKafkaMessage m ->
                    log.warn("DtoKafkaMessage without chatId type={} messageId={}", wsType, m.getMessageId());
        }
    }

    private void fanoutChatMember(DtoKafkaChatMember dto, String wsType) {
        switch (dto) {
            case null -> log.warn("Empty DtoKafkaChatMember type={}", wsType);
            case DtoKafkaChatMember m when m.getChatId() != null ->
                    publish(RedisChannels.chat(m.getChatId().toString()), wsType, m);
            case DtoKafkaChatMember m ->
                    log.warn("DtoKafkaChatMember without chatId type={} userId={}", wsType, m.getUserId());
        }
    }

    private void publish(String channel, String wsType, KafkaDto dto) {
        try {
            redis.convertAndSend(channel, toWsJson(wsType, dto));
        } catch (Exception e) {
            log.warn("Failed to publish wsType={} channel={}", wsType, channel, e);
        }
    }

    private String toWsJson(String wsType, Object dto) {
        try {
            ObjectNode node = mapper.valueToTree(dto);
            node.put("type", wsType);
            return mapper.writeValueAsString(node);
        } catch (Exception e) {
            throw new IllegalStateException("Failed to serialize ws payload type=" + wsType, e);
        }
    }
}
