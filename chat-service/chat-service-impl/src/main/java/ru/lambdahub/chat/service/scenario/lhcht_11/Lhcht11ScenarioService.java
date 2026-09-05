package ru.lambdahub.chat.service.scenario.lhcht_11;

import java.time.Instant;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.chat.db.MessageRepository;
import ru.lambdahub.chat.dto.MessageDto;
import ru.lambdahub.chat.mapstruct.ChatMapper;
import ru.lambdahub.chat.messaging.lhcht_out06.LhchtOut06Service;
import ru.lambdahub.chat.validation.LhchtValidInfo;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;

@Slf4j
@Service
@RequiredArgsConstructor
public class Lhcht11ScenarioService {

    private final MessageRepository messages;
    private final LhchtOut06Service lhchtOut06;
    private final ChatMapper chatMapper;

    @Transactional
    public void process(UUID userId, UUID messageId) {
        UUID requestId = UUID.randomUUID();
        log.info("[LHCHT-11] Starting scenario. requestId={} messageId={} userId={}", requestId, messageId, userId);

        var found = messages.findOne(messageId).map(chatMapper::toDto);
        Outcome outcome = new Outcome();
        if (found.isEmpty()) {
            outcome = outcome.addValidCode(LhchtValidInfo.LHCHT_11_0001, messageId);
        }
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }
        MessageDto msg = found.get();
        if (!msg.getSenderId().equals(userId)) {
            throw new SecurityException("Not your message");
        }
        msg.setDeletedAt(Instant.now());
        messages.save(chatMapper.toEntity(msg));
        lhchtOut06.send(chatMapper.toDeleteKafka(msg, requestId));

        log.info("[LHCHT-11] Scenario completed. requestId={}", requestId);
    }
}
