package ru.lambdahub.chat.service.support;

import java.time.Instant;
import java.util.List;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import org.springframework.cache.Cache;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.chat.api.ChatListItem;
import ru.lambdahub.chat.api.MemberView;
import ru.lambdahub.chat.api.MessageView;
import ru.lambdahub.chat.db.ChatMemberRepository;
import ru.lambdahub.chat.db.ChatRepository;
import ru.lambdahub.chat.db.MessageRepository;
import ru.lambdahub.chat.dto.ChatDto;
import ru.lambdahub.chat.dto.ChatMemberDto;
import ru.lambdahub.chat.dto.MessageDto;
import ru.lambdahub.chat.feign.UsersLookupFeignClient;
import ru.lambdahub.chat.mapstruct.ChatMapper;
import ru.lambdahub.chat.messaging.lhcht_out01.LhchtOut01Service;
import ru.lambdahub.chat.validation.LhchtValidInfo;
import ru.lambdahub.common.redis.cache.RedisCacheProviderService;
import ru.lambdahub.user.api.ProfileResponse;
import ru.lambdahub.validation.exception.ValidationException;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;

@Service
@RequiredArgsConstructor
public class ChatSupportService {

    private final ChatRepository chats;
    private final ChatMemberRepository members;
    private final MessageRepository messages;
    private final LhchtOut01Service lhchtOut01;
    private final RedisCacheProviderService redisCacheProviderService;
    private final UsersLookupFeignClient usersClient;
    private final ChatMapper chatMapper;

    public boolean isMember(UUID chatId, UUID userId) {
        return members.existsByChatIdAndUserId(chatId, userId);
    }

    public void requireMember(UUID chatId, UUID userId) {
        if (!members.existsByChatIdAndUserId(chatId, userId)) {
            throw new SecurityException("Not a chat member");
        }
    }

    public void cacheMembers(UUID chatId) {
        List<String> memberIds = members.findByChatId(chatId)
                .map(chatMapper::toDto)
                .map(m -> m.getUserId().toString())
                .toJavaList();
        chatMembersCache().put(chatId.toString(), memberIds);
    }

    public UUID resolveUsername(String username) {
        UUID cached = userByUsernameCache().get(username.toLowerCase(), UUID.class);
        if (cached != null) {
            return cached;
        }
        try {
            List<ProfileResponse> found = usersClient.search(username);
            if (found == null || found.isEmpty()) {
                Outcome outcome = new Outcome().addValidCode(LhchtValidInfo.LHCHT_USER_0001, username);
                throw ValidationExceptionUtils.createValidationException(outcome);
            }
            UUID resolved = null;
            for (ProfileResponse u : found) {
                if (username.equalsIgnoreCase(u.getUsername())) {
                    resolved = u.getId();
                    break;
                }
            }
            if (resolved == null) {
                resolved = found.get(0).getId();
            }
            userByUsernameCache().put(username.toLowerCase(), resolved);
            return resolved;
        } catch (ValidationException e) {
            throw e;
        } catch (Exception e) {
            Outcome outcome = new Outcome().addValidCode(LhchtValidInfo.LHCHT_USER_0001, username);
            throw ValidationExceptionUtils.createValidationException(outcome);
        }
    }

    @Transactional
    public ChatListItem createChat(String type, String name, UUID createdBy, List<UUID> memberIds, UUID requestId) {
        UUID chatId = UUID.randomUUID();
        ChatDto chat = ChatDto.builder()
                .id(chatId)
                .type(type)
                .name(name)
                .createdBy(createdBy)
                .build();
        chats.save(chatMapper.toEntity(chat));
        for (UUID uid : memberIds) {
            ChatMemberDto m = ChatMemberDto.builder()
                    .chatId(chatId)
                    .userId(uid)
                    .role(uid.equals(createdBy) ? "owner" : "member")
                    .joinedAt(Instant.now())
                    .build();
            members.save(chatMapper.toEntity(m));
        }
        cacheMembers(chatId);
        lhchtOut01.send(chatMapper.toKafka(chat, memberIds, requestId));
        return toListItem(chat, requestId);
    }

    public MemberView toMember(ChatMemberDto m, UUID requestId) {
        return chatMapper.toApi(m, requestId);
    }

    public ChatListItem toListItem(ChatDto chat, UUID requestId) {
        List<UUID> memberIds = members.findByChatId(chat.getId())
                .map(chatMapper::toDto)
                .map(ChatMemberDto::getUserId)
                .toJavaList();
        var last = messages.findFirstByChatIdAndDeletedAtIsNullOrderByCreatedAtDesc(chat.getId())
                .map(chatMapper::toDto);
        return chatMapper.toApi(
                chat,
                last.map(MessageDto::getContent).getOrNull(),
                last.map(MessageDto::getCreatedAt).getOrNull(),
                memberIds,
                requestId);
    }

    public MessageView toMessage(MessageDto m, UUID requestId) {
        return chatMapper.toApi(m, requestId);
    }

    private Cache chatMembersCache() {
        return redisCacheProviderService.getCache("chat-members")
                .orElseThrow(() -> new IllegalStateException("Cache not configured: chat-members"));
    }

    private Cache userByUsernameCache() {
        return redisCacheProviderService.getCache("user-by-username")
                .orElseThrow(() -> new IllegalStateException("Cache not configured: user-by-username"));
    }
}
