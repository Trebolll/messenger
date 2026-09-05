package ru.lambdahub.chat.service.scenario.lhcht_04;

import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.chat.api.ChatListItem;
import ru.lambdahub.chat.api.UpdateChatRequest;
import ru.lambdahub.chat.db.ChatRepository;
import ru.lambdahub.chat.dto.ChatDto;
import ru.lambdahub.chat.mapstruct.ChatMapper;
import ru.lambdahub.chat.service.support.ChatSupportService;
import ru.lambdahub.chat.service.support.RequestIds;
import ru.lambdahub.chat.validation.LhchtValidInfo;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;

@Slf4j
@Service
@RequiredArgsConstructor
public class Lhcht04ScenarioService {

    private final ChatRepository chats;
    private final ChatSupportService support;
    private final ChatMapper chatMapper;

    @Transactional
    public ChatListItem process(UUID userId, UUID chatId, UpdateChatRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHCHT-04] Starting scenario. requestId={} chatId={} request={}", requestId, chatId, req);

        support.requireMember(chatId, userId);
        var found = chats.findOne(chatId).map(chatMapper::toDto);
        Outcome outcome = new Outcome();
        if (found.isEmpty()) {
            outcome = outcome.addValidCode(LhchtValidInfo.LHCHT_04_0001, chatId);
        }
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }
        ChatDto chat = found.get();
        if (!"group".equals(chat.getType())) {
            outcome = outcome.addValidCode(LhchtValidInfo.LHCHT_04_0002);
            if (outcome.hasCriticalErrors()) {
                throw ValidationExceptionUtils.createValidationException(outcome);
            }
        }
        chat.setName(req.getName());
        chats.save(chatMapper.toEntity(chat));
        ChatListItem result = support.toListItem(chat, requestId);
        log.info("[LHCHT-04] Scenario completed. requestId={}", requestId);
        return result;
    }
}
