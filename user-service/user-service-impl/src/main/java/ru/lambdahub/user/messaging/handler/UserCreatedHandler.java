package ru.lambdahub.user.messaging.handler;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.messaging.Message;
import org.springframework.stereotype.Component;
import ru.lambdahub.common.events.kafka.DtoKafkaUserCreated;
import ru.lambdahub.user.service.scenario.lhuser_06.Lhuser06ScenarioService;

@Slf4j
@Component
@RequiredArgsConstructor
public class UserCreatedHandler {

    private final Lhuser06ScenarioService lhuser06;

    public void handleUserCreated(Message<DtoKafkaUserCreated> message) {
        DtoKafkaUserCreated dto = message.getPayload();
        switch (dto) {
            case null -> log.warn("userCreatedInput: empty payload, message={}", message);
            case DtoKafkaUserCreated d when d.getUserId() == null ->
                    log.warn("userCreatedInput: missing userId, message={}", message);
            case DtoKafkaUserCreated d -> lhuser06.process(d);
        }
    }
}
