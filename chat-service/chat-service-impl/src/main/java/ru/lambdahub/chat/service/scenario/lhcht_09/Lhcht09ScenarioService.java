package ru.lambdahub.chat.service.scenario.lhcht_09;

import java.util.List;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.chat.api.MessageView;
import ru.lambdahub.chat.db.MessageRepository;
import ru.lambdahub.chat.mapstruct.ChatMapper;
import ru.lambdahub.chat.service.support.ChatSupportService;

@Slf4j
@Service
@RequiredArgsConstructor
public class Lhcht09ScenarioService {

    private final MessageRepository messages;
    private final ChatSupportService support;
    private final ChatMapper chatMapper;

    @Transactional(readOnly = true)
    public List<MessageView> process(UUID userId, UUID chatId) {
        UUID requestId = UUID.randomUUID();
        log.info("[LHCHT-09] Starting scenario. requestId={} chatId={} userId={}", requestId, chatId, userId);

        support.requireMember(chatId, userId);
        List<MessageView> result = messages.findRecent(chatId)
                .reverse()
                .take(100)
                .map(chatMapper::toDto)
                .map(m -> support.toMessage(m, requestId))
                .toJavaList();

        log.info("[LHCHT-09] Scenario completed. requestId={}", requestId);
        return result;
    }
}
