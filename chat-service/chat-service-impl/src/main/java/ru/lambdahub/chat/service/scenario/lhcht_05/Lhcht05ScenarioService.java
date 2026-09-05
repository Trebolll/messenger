package ru.lambdahub.chat.service.scenario.lhcht_05;

import java.util.List;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.chat.api.MemberView;
import ru.lambdahub.chat.db.ChatMemberRepository;
import ru.lambdahub.chat.mapstruct.ChatMapper;
import ru.lambdahub.chat.service.support.ChatSupportService;

@Slf4j
@Service
@RequiredArgsConstructor
public class Lhcht05ScenarioService {

    private final ChatMemberRepository members;
    private final ChatSupportService support;
    private final ChatMapper chatMapper;

    @Transactional(readOnly = true)
    public List<MemberView> process(UUID userId, UUID chatId) {
        UUID requestId = UUID.randomUUID();
        log.info("[LHCHT-05] Starting scenario. requestId={} chatId={} userId={}", requestId, chatId, userId);

        support.requireMember(chatId, userId);
        List<MemberView> result = members.findByChatId(chatId)
                .map(chatMapper::toDto)
                .map(m -> support.toMember(m, requestId))
                .toJavaList();

        log.info("[LHCHT-05] Scenario completed. requestId={}", requestId);
        return result;
    }
}
