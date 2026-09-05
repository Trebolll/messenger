package ru.lambdahub.chat.mapstruct;

import java.time.Instant;
import java.util.List;
import java.util.UUID;
import org.mapstruct.Mapper;
import org.mapstruct.Mapping;
import org.mapstruct.ReportingPolicy;
import ru.lambdahub.chat.api.ChatListItem;
import ru.lambdahub.chat.api.CreateGroupRequest;
import ru.lambdahub.chat.api.MemberView;
import ru.lambdahub.chat.api.MessageView;
import ru.lambdahub.chat.api.SendMessageRequest;
import ru.lambdahub.chat.db.Chat;
import ru.lambdahub.chat.db.ChatMember;
import ru.lambdahub.chat.db.Message;
import ru.lambdahub.chat.db.MessageRead;
import ru.lambdahub.chat.dto.ChatDto;
import ru.lambdahub.chat.dto.ChatMemberDto;
import ru.lambdahub.chat.dto.MessageDto;
import ru.lambdahub.chat.dto.MessageReadDto;
import ru.lambdahub.common.events.kafka.DtoKafkaChatCreated;
import ru.lambdahub.common.events.kafka.DtoKafkaChatMember;
import ru.lambdahub.common.events.kafka.DtoKafkaMessage;
import ru.lambdahub.common.events.kafka.DtoKafkaMessagesRead;

@Mapper(componentModel = "spring", unmappedTargetPolicy = ReportingPolicy.IGNORE)
public interface ChatMapper {

    ChatDto toDto(Chat entity);

    Chat toEntity(ChatDto dto);

    ChatMemberDto toDto(ChatMember entity);

    ChatMember toEntity(ChatMemberDto dto);

    MessageDto toDto(Message entity);

    Message toEntity(MessageDto dto);

    MessageReadDto toDto(MessageRead entity);

    MessageRead toEntity(MessageReadDto dto);

    MessageView toApi(MessageDto message, UUID requestId);

    MemberView toApi(ChatMemberDto member, UUID requestId);

    @Mapping(target = "id", source = "chat.id")
    @Mapping(target = "type", source = "chat.type")
    @Mapping(target = "name", source = "chat.name")
    @Mapping(target = "avatarUrl", source = "chat.avatarUrl")
    ChatListItem toApi(ChatDto chat, String lastMessage, Instant lastMessageAt, List<UUID> memberIds, UUID requestId);

    @Mapping(target = "messageId", source = "message.id")
    @Mapping(target = "chatId", source = "message.chatId")
    @Mapping(target = "senderId", source = "message.senderId")
    @Mapping(target = "content", source = "message.content")
    @Mapping(target = "parentId", source = "message.parentId")
    @Mapping(target = "createdAt", source = "message.createdAt")
    @Mapping(target = "editedAt", source = "message.editedAt")
    DtoKafkaMessage toKafka(MessageDto message, UUID requestId);

    @Mapping(target = "chatId", source = "chat.id")
    @Mapping(target = "type", source = "chat.type")
    @Mapping(target = "name", source = "chat.name")
    DtoKafkaChatCreated toKafka(ChatDto chat, List<UUID> memberIds, UUID requestId);

    @Mapping(target = "chatId", source = "chatId")
    @Mapping(target = "userId", source = "userId")
    @Mapping(target = "addedBy", source = "addedBy")
    DtoKafkaChatMember toKafka(UUID chatId, UUID userId, UUID addedBy, UUID requestId);

    @Mapping(target = "readerId", source = "read.userId")
    @Mapping(target = "upToMessageId", source = "read.lastReadMessageId")
    @Mapping(target = "chatId", source = "read.chatId")
    @Mapping(target = "readAt", source = "read.readAt")
    DtoKafkaMessagesRead toKafka(MessageReadDto read, UUID requestId);

    @Mapping(target = "messageId", source = "message.id")
    @Mapping(target = "chatId", source = "message.chatId")
    @Mapping(target = "senderId", ignore = true)
    @Mapping(target = "content", ignore = true)
    @Mapping(target = "parentId", ignore = true)
    @Mapping(target = "createdAt", ignore = true)
    @Mapping(target = "editedAt", ignore = true)
    DtoKafkaMessage toDeleteKafka(MessageDto message, UUID requestId);

    @Mapping(target = "id", ignore = true)
    @Mapping(target = "senderId", ignore = true)
    @Mapping(target = "editedAt", ignore = true)
    @Mapping(target = "deletedAt", ignore = true)
    @Mapping(target = "createdAt", ignore = true)
    @Mapping(target = "content", expression = "java(req.getContent() == null ? \"\" : req.getContent())")
    MessageDto toDto(SendMessageRequest req);

    @Mapping(target = "id", ignore = true)
    @Mapping(target = "type", constant = "group")
    @Mapping(target = "avatarUrl", ignore = true)
    @Mapping(target = "createdBy", ignore = true)
    ChatDto toDto(CreateGroupRequest req);

    default String map(Instant instant) {
        return instant == null ? null : instant.toString();
    }
}
