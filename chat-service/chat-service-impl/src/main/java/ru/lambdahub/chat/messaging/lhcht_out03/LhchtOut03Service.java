package ru.lambdahub.chat.messaging.lhcht_out03;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Service;
import ru.lambdahub.common.events.kafka.DtoKafkaChatMember;
import ru.lambdahub.common.kafka.OutputBridge;

/** LHCHT-OUT03 — chat.member_removed */
@Slf4j
@Service
@RequiredArgsConstructor
public class LhchtOut03Service {

    @Qualifier("chatMemberRemovedOutput")
    private final ObjectProvider<OutputBridge> chatMemberRemovedOutput;

    public void send(DtoKafkaChatMember dto) {
        log.info("[LHCHT-OUT03] Starting. requestId={} chatId={}", dto.getRequestId(), dto.getChatId());
        chatMemberRemovedOutput.ifAvailable(ob -> ob.send(dto));
        log.info("[LHCHT-OUT03] Completed. requestId={} chatId={}", dto.getRequestId(), dto.getChatId());
    }
}
