package ru.lambdahub.chat.service.scenario.lhcht_01;

import java.util.List;
import java.util.Set;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.chat.api.ChatListItem;
import ru.lambdahub.chat.api.CreatePrivateRequest;
import ru.lambdahub.chat.db.ChatMemberRepository;
import ru.lambdahub.chat.db.ChatRepository;
import ru.lambdahub.chat.dto.ChatDto;
import ru.lambdahub.chat.dto.ChatMemberDto;
import ru.lambdahub.chat.mapstruct.ChatMapper;
import ru.lambdahub.chat.service.support.ChatSupportService;
import ru.lambdahub.chat.service.support.RequestIds;
import ru.lambdahub.chat.validation.LhchtValidInfo;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;

@Slf4j
@Service
@RequiredArgsConstructor
public class Lhcht01ScenarioService {

    private final ChatRepository chats;
    private final ChatMemberRepository members;
    private final ChatSupportService support;
    private final ChatMapper chatMapper;

    @Transactional
    public ChatListItem process(UUID me, CreatePrivateRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHCHT-01] Starting scenario. requestId={} request={}", requestId, req);

        UUID otherUserId = req.getUserId();
        if (me.equals(otherUserId)) {
            Outcome outcome = new Outcome().addValidCode(LhchtValidInfo.LHCHT_01_0001);
            throw ValidationExceptionUtils.createValidationException(outcome);
        }
        for (UUID chatId : members.findChatIdsByUserId(me)) {
            ChatDto chat = chats.findOne(chatId).map(chatMapper::toDto).getOrNull();
            if (chat != null && "private".equals(chat.getType())) {
                Set<UUID> ids = members.findByChatId(chatId)
                        .map(chatMapper::toDto)
                        .map(ChatMemberDto::getUserId)
                        .toJavaSet();
                if (ids.contains(otherUserId) && ids.size() == 2) {
                    ChatListItem existing = support.toListItem(chat, requestId);
                    log.info("[LHCHT-01] Scenario completed. requestId={}", requestId);
                    return existing;
                }
            }
        }
        ChatListItem created = support.createChat("private", null, me, List.of(me, otherUserId), requestId);
        log.info("[LHCHT-01] Scenario completed. requestId={}", requestId);
        return created;
    }
}
