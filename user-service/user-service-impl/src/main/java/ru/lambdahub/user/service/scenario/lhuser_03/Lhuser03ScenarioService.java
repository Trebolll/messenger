package ru.lambdahub.user.service.scenario.lhuser_03;

import io.vavr.control.Option;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.user.api.ProfileResponse;
import ru.lambdahub.user.api.UpdateProfileRequest;
import ru.lambdahub.user.db.ProfileRepository;
import ru.lambdahub.user.dto.ProfileDto;
import ru.lambdahub.user.mapstruct.UserMapper;
import ru.lambdahub.user.messaging.lhuser_out01.LhuserOut01Service;
import ru.lambdahub.user.service.support.RequestIds;
import ru.lambdahub.user.validation.LhuserValidInfo;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;
import ru.lambdahub.validation.validator.OutcomeValidator;

/** LHUSER-03 — update profile */
@Slf4j
@Service
@RequiredArgsConstructor
public class Lhuser03ScenarioService {

    private final ProfileRepository profiles;
    private final UserMapper userMapper;
    private final LhuserOut01Service lhuserOut01;
    private final OutcomeValidator vavrValidator;

    @Transactional
    public ProfileResponse process(UUID userId, UpdateProfileRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHUSER-03] Starting scenario. requestId={} userId={} request={}", requestId, userId, req);

        Outcome outcome = Outcome.of(vavrValidator.validate(req));
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }
        if (!hasAnyField(req)) {
            outcome = outcome.addValidCode(LhuserValidInfo.LHUSER_03_0002);
            if (outcome.hasCriticalErrors()) {
                throw ValidationExceptionUtils.createValidationException(outcome);
            }
        }

        Option<ProfileDto> found = profiles.findOne(userId).map(userMapper::toDto);
        if (found.isEmpty()) {
            outcome = outcome.addValidCode(LhuserValidInfo.LHUSER_03_0001, userId);
        }
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }
        ProfileDto p = found.get();

        if (req.getUsername() != null && !req.getUsername().isBlank()) {
            String username = req.getUsername().trim();
            Option<ProfileDto> existingOpt = profiles.findByUsernameIgnoreCase(username)
                    .map(userMapper::toDto)
                    .filter(existing -> !existing.getUserId().equals(userId));
            if (existingOpt.isDefined()) {
                outcome = outcome.addValidCode(LhuserValidInfo.LHUSER_03_0003, username);
                if (outcome.hasCriticalErrors()) {
                    throw ValidationExceptionUtils.createValidationException(outcome);
                }
            }
            p.setUsername(username);
        }
        if (req.getDisplayName() != null) {
            p.setDisplayName(req.getDisplayName());
        }
        if (req.getPhone() != null) {
            p.setPhone(req.getPhone());
        }
        if (req.getProfession() != null) {
            p.setProfession(req.getProfession());
        }
        if (req.getLocation() != null) {
            p.setLocation(req.getLocation());
        }
        if (req.getStatus() != null) {
            p.setStatusText(req.getStatus());
        }
        profiles.save(userMapper.toEntity(p));
        lhuserOut01.send(p, requestId);

        ProfileResponse result = userMapper.toApi(p);
        log.info("[LHUSER-03] Scenario completed. requestId={}", requestId);
        return result;
    }

    private static boolean hasAnyField(UpdateProfileRequest req) {
        return (req.getUsername() != null && !req.getUsername().isBlank())
                || req.getDisplayName() != null
                || req.getPhone() != null
                || req.getProfession() != null
                || req.getLocation() != null
                || req.getStatus() != null;
    }
}
