package ru.lambdahub.chat.service.scenario.lhcht_06;

import java.time.Instant;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.chat.api.AddMemberRequest;
import ru.lambdahub.chat.db.ChatMemberRepository;
import ru.lambdahub.chat.db.ChatRepository;
import ru.lambdahub.chat.dto.ChatDto;
import ru.lambdahub.chat.dto.ChatMemberDto;
import ru.lambdahub.chat.mapstruct.ChatMapper;
import ru.lambdahub.chat.messaging.lhcht_out02.LhchtOut02Service;
import ru.lambdahub.chat.service.support.ChatSupportService;
import ru.lambdahub.chat.service.support.RequestIds;
import ru.lambdahub.chat.validation.LhchtValidInfo;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;

@Slf4j
@Service
@RequiredArgsConstructor
public class Lhcht06ScenarioService {

    private final ChatRepository chats;
    private final ChatMemberRepository members;
    private final ChatSupportService support;
    private final LhchtOut02Service lhchtOut02;
    private final ChatMapper chatMapper;

    @Transactional
    public void process(UUID actor, UUID chatId, AddMemberRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHCHT-06] Starting scenario. requestId={} chatId={} request={}", requestId, chatId, req);

        support.requireMember(chatId, actor);
        var found = chats.findOne(chatId).map(chatMapper::toDto);
        Outcome outcome = new Outcome();
        if (found.isEmpty()) {
            outcome = outcome.addValidCode(LhchtValidInfo.LHCHT_06_0001, chatId);
        }
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }
        ChatDto chat = found.get();
        if (!"group".equals(chat.getType())) {
            outcome = outcome.addValidCode(LhchtValidInfo.LHCHT_06_0002);
            if (outcome.hasCriticalErrors()) {
                throw ValidationExceptionUtils.createValidationException(outcome);
            }
        }
        UUID userId = support.resolveUsername(req.getUsername());
        if (members.existsByChatIdAndUserId(chatId, userId)) {
            log.info("[LHCHT-06] Scenario completed. requestId={}", requestId);
            return;
        }
        ChatMemberDto m = ChatMemberDto.builder()
                .chatId(chatId)
                .userId(userId)
                .role("member")
                .joinedAt(Instant.now())
                .build();
        members.save(chatMapper.toEntity(m));
        support.cacheMembers(chatId);
        lhchtOut02.send(chatMapper.toKafka(chatId, userId, actor, requestId));

        log.info("[LHCHT-06] Scenario completed. requestId={}", requestId);
    }
}
