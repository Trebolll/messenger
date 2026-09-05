package ru.lambdahub.user.web;

import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.RestController;
import ru.lambdahub.common.events.kafka.DtoKafkaUserCreated;
import ru.lambdahub.user.api.InternalUsersApi;
import ru.lambdahub.user.service.scenario.lhuser_06.Lhuser06ScenarioService;

@RestController
@RequiredArgsConstructor
public class InternalUserController implements InternalUsersApi {

    private final Lhuser06ScenarioService lhuser06;
    @Value("${app.internal-key}")
    private final String internalKey;

    @Override
    public void create(String key, DtoKafkaUserCreated payload) {
        if (internalKey != null && !internalKey.equals(key)) {
            throw new SecurityException("Invalid internal key");
        }
        lhuser06.process(payload);
    }
}
