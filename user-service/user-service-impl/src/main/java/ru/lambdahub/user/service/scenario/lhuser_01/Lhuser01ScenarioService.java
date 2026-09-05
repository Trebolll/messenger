package ru.lambdahub.user.service.scenario.lhuser_01;

import io.vavr.control.Option;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import ru.lambdahub.user.api.ProfileResponse;
import ru.lambdahub.user.db.ProfileRepository;
import ru.lambdahub.user.dto.ProfileDto;
import ru.lambdahub.user.mapstruct.UserMapper;
import ru.lambdahub.user.validation.LhuserValidInfo;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;

@Slf4j
@Service
@RequiredArgsConstructor
public class Lhuser01ScenarioService {

    private final ProfileRepository profiles;
    private final UserMapper userMapper;

    public ProfileResponse process(UUID userId) {
        UUID requestId = UUID.randomUUID();
        log.info("[LHUSER-01] Starting scenario. requestId={} userId={}", requestId, userId);
        Option<ProfileDto> found = profiles.findOne(userId).map(userMapper::toDto);
        Outcome outcome = new Outcome();
        if (found.isEmpty()) {
            outcome = outcome.addValidCode(LhuserValidInfo.LHUSER_01_0001, userId);
        }
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }
        ProfileResponse result = userMapper.toApi(found.get());
        log.info("[LHUSER-01] Scenario completed. requestId={}", requestId);
        return result;
    }
}
