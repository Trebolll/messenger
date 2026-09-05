package ru.lambdahub.user.messaging.lhuser_out02;

import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Service;
import ru.lambdahub.common.events.kafka.DtoKafkaUserStatus;
import ru.lambdahub.common.kafka.OutputBridge;
import ru.lambdahub.user.dto.ProfileDto;
import ru.lambdahub.user.mapstruct.UserMapper;

/** LHUSER-OUT02 — user.status.changed */
@Slf4j
@Service
@RequiredArgsConstructor
public class LhuserOut02Service {

    @Qualifier("userStatusChangedOutput")
    private final ObjectProvider<OutputBridge> userStatusChangedOutput;
    private final UserMapper userMapper;

    public void send(ProfileDto profile, UUID requestId) {
        log.info("[LHUSER-OUT02] Starting. requestId={} userId={}", requestId, profile.getUserId());
        DtoKafkaUserStatus dto = userMapper.toStatusKafka(profile, requestId);
        userStatusChangedOutput.ifAvailable(ob -> ob.send(dto));
        log.info("[LHUSER-OUT02] Completed. requestId={} userId={}", requestId, profile.getUserId());
    }
}
