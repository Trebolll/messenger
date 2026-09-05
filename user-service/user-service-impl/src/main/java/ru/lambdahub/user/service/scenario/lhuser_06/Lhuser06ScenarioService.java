package ru.lambdahub.user.service.scenario.lhuser_06;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.common.events.kafka.DtoKafkaUserCreated;
import ru.lambdahub.user.db.ProfileRepository;
import ru.lambdahub.user.db.UserRepository;
import ru.lambdahub.user.dto.ProfileDto;
import ru.lambdahub.user.dto.UserDto;
import ru.lambdahub.user.mapstruct.UserMapper;
import ru.lambdahub.user.validation.LhuserValidInfo;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;

/** LHUSER-06 — create user + profile from user.created event / internal API */
@Slf4j
@Service
@RequiredArgsConstructor
public class Lhuser06ScenarioService {

    private final UserRepository users;
    private final ProfileRepository profiles;
    private final UserMapper userMapper;

    @Transactional
    public void process(DtoKafkaUserCreated dto) {
        log.info("[LHUSER-06] Starting scenario. requestId={} userId={}",
                dto == null ? null : dto.getRequestId(),
                dto == null ? null : dto.getUserId());

        if (dto == null || dto.getUserId() == null
                || dto.getUsername() == null || dto.getUsername().isBlank()) {
            Outcome outcome = new Outcome().addValidCode(LhuserValidInfo.LHUSER_06_0001);
            throw ValidationExceptionUtils.createValidationException(outcome);
        }

        if (!users.existsById(dto.getUserId())) {
            UserDto user = userMapper.toUserDto(dto);
            users.save(userMapper.toEntity(user));
        }

        if (profiles.existsById(dto.getUserId())) {
            log.info("[LHUSER-06] Profile already exists. userId={}", dto.getUserId());
            return;
        }

        ProfileDto profile = userMapper.toDto(dto);
        profiles.save(userMapper.toEntity(profile));
        log.info("[LHUSER-06] Scenario completed. requestId={} userId={}", dto.getRequestId(), dto.getUserId());
    }
}
