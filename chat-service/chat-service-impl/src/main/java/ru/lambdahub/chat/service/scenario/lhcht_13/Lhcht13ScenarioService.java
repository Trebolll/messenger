package ru.lambdahub.chat.service.scenario.lhcht_13;

import java.util.Map;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.chat.service.support.ChatSupportService;

@Slf4j
@Service
@RequiredArgsConstructor
public class Lhcht13ScenarioService {

    private final ChatSupportService support;

    @Transactional(readOnly = true)
    public Map<String, Boolean> process(UUID chatId, UUID userId) {
        UUID requestId = UUID.randomUUID();
        log.info("[LHCHT-13] Starting scenario. requestId={} chatId={} userId={}", requestId, chatId, userId);

        Map<String, Boolean> result = Map.of("member", support.isMember(chatId, userId));

        log.info("[LHCHT-13] Scenario completed. requestId={}", requestId);
        return result;
    }
}
