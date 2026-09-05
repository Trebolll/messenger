package ru.lambdahub.chat.messaging.lhcht_out05;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Service;
import ru.lambdahub.common.events.kafka.DtoKafkaMessage;
import ru.lambdahub.common.kafka.OutputBridge;

/** LHCHT-OUT05 — message.edited */
@Slf4j
@Service
@RequiredArgsConstructor
public class LhchtOut05Service {

    @Qualifier("messageEditedOutput")
    private final ObjectProvider<OutputBridge> messageEditedOutput;

    public void send(DtoKafkaMessage dto) {
        log.info("[LHCHT-OUT05] Starting. requestId={} chatId={}", dto.getRequestId(), dto.getChatId());
        messageEditedOutput.ifAvailable(ob -> ob.send(dto));
        log.info("[LHCHT-OUT05] Completed. requestId={} chatId={}", dto.getRequestId(), dto.getChatId());
    }
}
