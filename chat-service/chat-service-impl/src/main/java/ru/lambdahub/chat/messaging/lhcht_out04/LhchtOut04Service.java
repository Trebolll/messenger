package ru.lambdahub.chat.messaging.lhcht_out04;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Service;
import ru.lambdahub.common.events.kafka.DtoKafkaMessage;
import ru.lambdahub.common.kafka.OutputBridge;

/** LHCHT-OUT04 — message.created */
@Slf4j
@Service
@RequiredArgsConstructor
public class LhchtOut04Service {

    @Qualifier("messageCreatedOutput")
    private final ObjectProvider<OutputBridge> messageCreatedOutput;

    public void send(DtoKafkaMessage dto) {
        log.info("[LHCHT-OUT04] Starting. requestId={} chatId={}", dto.getRequestId(), dto.getChatId());
        messageCreatedOutput.ifAvailable(ob -> ob.send(dto));
        log.info("[LHCHT-OUT04] Completed. requestId={} chatId={}", dto.getRequestId(), dto.getChatId());
    }
}
