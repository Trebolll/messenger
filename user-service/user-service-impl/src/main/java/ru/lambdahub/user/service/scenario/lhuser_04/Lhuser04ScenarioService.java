package ru.lambdahub.user.service.scenario.lhuser_04;

import io.vavr.control.Option;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.user.api.ProfileResponse;
import ru.lambdahub.user.api.UpdateStatusRequest;
import ru.lambdahub.user.db.ProfileRepository;
import ru.lambdahub.user.dto.ProfileDto;
import ru.lambdahub.user.mapstruct.UserMapper;
import ru.lambdahub.user.messaging.lhuser_out02.LhuserOut02Service;
import ru.lambdahub.user.service.support.RequestIds;
import ru.lambdahub.user.validation.LhuserValidInfo;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;
import ru.lambdahub.validation.validator.OutcomeValidator;

/** LHUSER-04 — update status */
@Slf4j
@Service
@RequiredArgsConstructor
public class Lhuser04ScenarioService {

    private final ProfileRepository profiles;
    private final UserMapper userMapper;
    private final LhuserOut02Service lhuserOut02;
    private final OutcomeValidator vavrValidator;

    @Transactional
    public ProfileResponse process(UUID userId, UpdateStatusRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHUSER-04] Starting scenario. requestId={} userId={} request={}", requestId, userId, req);

        Outcome outcome = Outcome.of(vavrValidator.validate(req));
        Option<ProfileDto> found = profiles.findOne(userId).map(userMapper::toDto);
        if (found.isEmpty()) {
            outcome = outcome.addValidCode(LhuserValidInfo.LHUSER_04_0001, userId);
        }
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }

        ProfileDto p = found.get();
        p.setStatusText(req.getStatus() == null ? "" : req.getStatus());
        profiles.save(userMapper.toEntity(p));
        lhuserOut02.send(p, requestId);

        ProfileResponse result = userMapper.toApi(p);
        log.info("[LHUSER-04] Scenario completed. requestId={}", requestId);
        return result;
    }
}
