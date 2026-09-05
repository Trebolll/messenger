package ru.lambdahub.chat.messaging.lhcht_out06;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Service;
import ru.lambdahub.common.events.kafka.DtoKafkaMessage;
import ru.lambdahub.common.kafka.OutputBridge;

/** LHCHT-OUT06 — message.deleted */
@Slf4j
@Service
@RequiredArgsConstructor
public class LhchtOut06Service {

    @Qualifier("messageDeletedOutput")
    private final ObjectProvider<OutputBridge> messageDeletedOutput;

    public void send(DtoKafkaMessage dto) {
        log.info("[LHCHT-OUT06] Starting. requestId={} chatId={}", dto.getRequestId(), dto.getChatId());
        messageDeletedOutput.ifAvailable(ob -> ob.send(dto));
        log.info("[LHCHT-OUT06] Completed. requestId={} chatId={}", dto.getRequestId(), dto.getChatId());
    }
}
