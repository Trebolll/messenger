package ru.lambdahub.user.service.scenario.lhuser_05;

import io.vavr.control.Option;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.user.api.ProfileResponse;
import ru.lambdahub.user.api.UpdateAvatarRequest;
import ru.lambdahub.user.db.ProfileRepository;
import ru.lambdahub.user.dto.ProfileDto;
import ru.lambdahub.user.mapstruct.UserMapper;
import ru.lambdahub.user.messaging.lhuser_out01.LhuserOut01Service;
import ru.lambdahub.user.service.support.RequestIds;
import ru.lambdahub.user.validation.LhuserValidInfo;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;
import ru.lambdahub.validation.validator.OutcomeValidator;

/** LHUSER-05 — update avatar */
@Slf4j
@Service
@RequiredArgsConstructor
public class Lhuser05ScenarioService {

    private final ProfileRepository profiles;
    private final UserMapper userMapper;
    private final LhuserOut01Service lhuserOut01;
    private final OutcomeValidator vavrValidator;

    @Transactional
    public ProfileResponse process(UUID userId, UpdateAvatarRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHUSER-05] Starting scenario. requestId={} userId={} request={}", requestId, userId, req);

        Outcome outcome = Outcome.of(vavrValidator.validate(req));
        Option<ProfileDto> found = profiles.findOne(userId).map(userMapper::toDto);
        if (found.isEmpty()) {
            outcome = outcome.addValidCode(LhuserValidInfo.LHUSER_05_0001, userId);
        }
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }

        ProfileDto p = found.get();
        p.setAvatarUrl(req.getAvatarUrl());
        profiles.save(userMapper.toEntity(p));
        lhuserOut01.send(p, requestId);

        ProfileResponse result = userMapper.toApi(p);
        log.info("[LHUSER-05] Scenario completed. requestId={}", requestId);
        return result;
    }
}
