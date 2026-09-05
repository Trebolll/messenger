package ru.lambdahub.user.messaging.lhuser_out01;

import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Service;
import ru.lambdahub.common.events.kafka.DtoKafkaUserProfile;
import ru.lambdahub.common.kafka.OutputBridge;
import ru.lambdahub.user.dto.ProfileDto;
import ru.lambdahub.user.mapstruct.UserMapper;

/** LHUSER-OUT01 — user.profile.updated */
@Slf4j
@Service
@RequiredArgsConstructor
public class LhuserOut01Service {

    @Qualifier("userProfileUpdatedOutput")
    private final ObjectProvider<OutputBridge> userProfileUpdatedOutput;
    private final UserMapper userMapper;

    public void send(ProfileDto profile, UUID requestId) {
        log.info("[LHUSER-OUT01] Starting. requestId={} userId={}", requestId, profile.getUserId());
        DtoKafkaUserProfile dto = userMapper.toKafka(profile, requestId);
        userProfileUpdatedOutput.ifAvailable(ob -> ob.send(dto));
        log.info("[LHUSER-OUT01] Completed. requestId={} userId={}", requestId, profile.getUserId());
    }
}
