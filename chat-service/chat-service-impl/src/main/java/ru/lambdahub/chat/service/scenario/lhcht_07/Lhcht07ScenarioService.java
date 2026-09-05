package ru.lambdahub.chat.service.scenario.lhcht_07;

import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.chat.db.ChatMemberRepository;
import ru.lambdahub.chat.mapstruct.ChatMapper;
import ru.lambdahub.chat.messaging.lhcht_out03.LhchtOut03Service;
import ru.lambdahub.chat.service.support.ChatSupportService;

@Slf4j
@Service
@RequiredArgsConstructor
public class Lhcht07ScenarioService {

    private final ChatMemberRepository members;
    private final ChatSupportService support;
    private final LhchtOut03Service lhchtOut03;
    private final ChatMapper chatMapper;

    @Transactional
    public void process(UUID actor, UUID chatId, UUID userId) {
        UUID requestId = UUID.randomUUID();
        log.info("[LHCHT-07] Starting scenario. requestId={} chatId={} userId={}", requestId, chatId, userId);

        support.requireMember(chatId, actor);
        members.deleteByChatIdAndUserId(chatId, userId);
        support.cacheMembers(chatId);
        lhchtOut03.send(chatMapper.toKafka(chatId, userId, null, requestId));

        log.info("[LHCHT-07] Scenario completed. requestId={}", requestId);
    }
}
