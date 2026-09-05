package ru.lambdahub.chat.service.scenario.lhcht_08;

import java.time.Instant;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.chat.api.MessageView;
import ru.lambdahub.chat.api.SendMessageRequest;
import ru.lambdahub.chat.db.MessageRepository;
import ru.lambdahub.chat.dto.MessageDto;
import ru.lambdahub.chat.mapstruct.ChatMapper;
import ru.lambdahub.chat.messaging.lhcht_out04.LhchtOut04Service;
import ru.lambdahub.chat.service.support.ChatSupportService;
import ru.lambdahub.chat.service.support.RequestIds;

@Slf4j
@Service
@RequiredArgsConstructor
public class Lhcht08ScenarioService {

    private final MessageRepository messages;
    private final ChatSupportService support;
    private final LhchtOut04Service lhchtOut04;
    private final ChatMapper chatMapper;

    @Transactional
    public MessageView process(UUID senderId, SendMessageRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHCHT-08] Starting scenario. requestId={} request={}", requestId, req);

        support.requireMember(req.getChatId(), senderId);
        MessageDto msg = chatMapper.toDto(req);
        msg.setId(UUID.randomUUID());
        msg.setSenderId(senderId);
        msg.setCreatedAt(Instant.now());
        messages.save(chatMapper.toEntity(msg));
        lhchtOut04.send(chatMapper.toKafka(msg, requestId));
        MessageView result = support.toMessage(msg, requestId);
        log.info("[LHCHT-08] Scenario completed. requestId={}", requestId);
        return result;
    }
}