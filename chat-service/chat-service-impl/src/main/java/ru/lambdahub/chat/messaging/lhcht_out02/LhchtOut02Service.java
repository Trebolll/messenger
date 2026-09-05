package ru.lambdahub.chat.messaging.lhcht_out02;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Service;
import ru.lambdahub.common.events.kafka.DtoKafkaChatMember;
import ru.lambdahub.common.kafka.OutputBridge;

/** LHCHT-OUT02 — chat.member_added */
@Slf4j
@Service
@RequiredArgsConstructor
public class LhchtOut02Service {

    @Qualifier("chatMemberAddedOutput")
    private final ObjectProvider<OutputBridge> chatMemberAddedOutput;

    public void send(DtoKafkaChatMember dto) {
        log.info("[LHCHT-OUT02] Starting. requestId={} chatId={}", dto.getRequestId(), dto.getChatId());
        chatMemberAddedOutput.ifAvailable(ob -> ob.send(dto));
        log.info("[LHCHT-OUT02] Completed. requestId={} chatId={}", dto.getRequestId(), dto.getChatId());
    }
}
