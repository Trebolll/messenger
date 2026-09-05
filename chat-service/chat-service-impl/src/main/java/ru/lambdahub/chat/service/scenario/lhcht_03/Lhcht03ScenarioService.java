package ru.lambdahub.chat.service.scenario.lhcht_03;

import java.time.Instant;
import java.util.List;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.chat.api.ChatListItem;
import ru.lambdahub.chat.db.ChatMemberRepository;
import ru.lambdahub.chat.db.ChatRepository;
import ru.lambdahub.chat.mapstruct.ChatMapper;
import ru.lambdahub.chat.service.support.ChatSupportService;

@Slf4j
@Service
@RequiredArgsConstructor
public class Lhcht03ScenarioService {

    private final ChatRepository chats;
    private final ChatMemberRepository members;
    private final ChatSupportService support;
    private final ChatMapper chatMapper;

    @Transactional(readOnly = true)
    public List<ChatListItem> process(UUID userId) {
        UUID requestId = UUID.randomUUID();
        log.info("[LHCHT-03] Starting scenario. requestId={} userId={}", requestId, userId);

        List<ChatListItem> result = members.findChatIdsByUserId(userId)
                .flatMap(id -> chats.findOne(id))
                .map(chatMapper::toDto)
                .map(chat -> support.toListItem(chat, requestId))
                .sortBy(c -> c.getLastMessageAt() == null ? Instant.EPOCH : c.getLastMessageAt())
                .reverse()
                .toJavaList();

        log.info("[LHCHT-03] Scenario completed. requestId={}", requestId);
        return result;
    }
}
