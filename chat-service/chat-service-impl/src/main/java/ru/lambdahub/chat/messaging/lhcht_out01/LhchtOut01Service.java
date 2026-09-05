package ru.lambdahub.chat.messaging.lhcht_out01;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Service;
import ru.lambdahub.common.events.kafka.DtoKafkaChatCreated;
import ru.lambdahub.common.kafka.OutputBridge;

/** LHCHT-OUT01 — chat.created */
@Slf4j
@Service
@RequiredArgsConstructor
public class LhchtOut01Service {

    @Qualifier("chatCreatedOutput")
    private final ObjectProvider<OutputBridge> chatCreatedOutput;

    public void send(DtoKafkaChatCreated dto) {
        log.info("[LHCHT-OUT01] Starting. requestId={} chatId={}", dto.getRequestId(), dto.getChatId());
        chatCreatedOutput.ifAvailable(ob -> ob.send(dto));
        log.info("[LHCHT-OUT01] Completed. requestId={} chatId={}", dto.getRequestId(), dto.getChatId());
    }
}
