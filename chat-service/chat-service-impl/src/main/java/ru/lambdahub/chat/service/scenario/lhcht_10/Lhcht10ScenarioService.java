package ru.lambdahub.chat.service.scenario.lhcht_10;

import java.time.Instant;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.chat.api.EditMessageRequest;
import ru.lambdahub.chat.api.MessageView;
import ru.lambdahub.chat.db.MessageRepository;
import ru.lambdahub.chat.dto.MessageDto;
import ru.lambdahub.chat.mapstruct.ChatMapper;
import ru.lambdahub.chat.messaging.lhcht_out05.LhchtOut05Service;
import ru.lambdahub.chat.service.support.ChatSupportService;
import ru.lambdahub.chat.service.support.RequestIds;
import ru.lambdahub.chat.validation.LhchtValidInfo;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;

@Slf4j
@Service
@RequiredArgsConstructor
public class Lhcht10ScenarioService {

    private final MessageRepository messages;
    private final ChatSupportService support;
    private final LhchtOut05Service lhchtOut05;
    private final ChatMapper chatMapper;

    @Transactional
    public MessageView process(UUID userId, UUID messageId, EditMessageRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHCHT-10] Starting scenario. requestId={} messageId={} request={}", requestId, messageId, req);

        var found = messages.findOne(messageId).map(chatMapper::toDto);
        Outcome outcome = new Outcome();
        if (found.isEmpty()) {
            outcome = outcome.addValidCode(LhchtValidInfo.LHCHT_10_0001, messageId);
        }
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }
        MessageDto msg = found.get();
        if (!msg.getSenderId().equals(userId)) {
            throw new SecurityException("Not your message");
        }
        msg.setContent(req.getContent());
        msg.setEditedAt(Instant.now());
        messages.save(chatMapper.toEntity(msg));
        lhchtOut05.send(chatMapper.toKafka(msg, requestId));
        MessageView result = support.toMessage(msg, requestId);
        log.info("[LHCHT-10] Scenario completed. requestId={}", requestId);
        return result;
    }
}
