package ru.lambdahub.chat.service.scenario.lhcht_12;

import java.time.Instant;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.chat.api.MarkReadRequest;
import ru.lambdahub.chat.db.MessageRead;
import ru.lambdahub.chat.db.MessageReadRepository;
import ru.lambdahub.chat.dto.MessageReadDto;
import ru.lambdahub.chat.mapstruct.ChatMapper;
import ru.lambdahub.chat.messaging.lhcht_out07.LhchtOut07Service;
import ru.lambdahub.chat.service.support.ChatSupportService;
import ru.lambdahub.chat.service.support.RequestIds;

@Slf4j
@Service
@RequiredArgsConstructor
public class Lhcht12ScenarioService {

    private final MessageReadRepository reads;
    private final ChatSupportService support;
    private final LhchtOut07Service lhchtOut07;
    private final ChatMapper chatMapper;

    @Transactional
    public void process(UUID userId, UUID chatId, MarkReadRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHCHT-12] Starting scenario. requestId={} chatId={} request={}", requestId, chatId, req);

        UUID upToMessageId = req == null ? null : req.getUpToMessageId();
        support.requireMember(chatId, userId);
        MessageRead.Pk pk = new MessageRead.Pk();
        pk.setChatId(chatId);
        pk.setUserId(userId);
        MessageReadDto read = reads.findOne(pk)
                .map(chatMapper::toDto)
                .getOrElse(() -> MessageReadDto.builder()
                        .chatId(chatId)
                        .userId(userId)
                        .build());
        read.setLastReadMessageId(upToMessageId);
        read.setReadAt(Instant.now());
        reads.save(chatMapper.toEntity(read));
        lhchtOut07.send(chatMapper.toKafka(read, requestId));

        log.info("[LHCHT-12] Scenario completed. requestId={}", requestId);
    }
}
