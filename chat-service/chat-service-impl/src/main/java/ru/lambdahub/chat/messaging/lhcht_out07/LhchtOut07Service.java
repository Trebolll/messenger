package ru.lambdahub.chat.messaging.lhcht_out07;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Service;
import ru.lambdahub.common.events.kafka.DtoKafkaMessagesRead;
import ru.lambdahub.common.kafka.OutputBridge;

/** LHCHT-OUT07 — messages.read */
@Slf4j
@Service
@RequiredArgsConstructor
public class LhchtOut07Service {

    @Qualifier("messagesReadOutput")
    private final ObjectProvider<OutputBridge> messagesReadOutput;

    public void send(DtoKafkaMessagesRead dto) {
        log.info("[LHCHT-OUT07] Starting. requestId={} chatId={}", dto.getRequestId(), dto.getChatId());
        messagesReadOutput.ifAvailable(ob -> ob.send(dto));
        log.info("[LHCHT-OUT07] Completed. requestId={} chatId={}", dto.getRequestId(), dto.getChatId());
    }
}
