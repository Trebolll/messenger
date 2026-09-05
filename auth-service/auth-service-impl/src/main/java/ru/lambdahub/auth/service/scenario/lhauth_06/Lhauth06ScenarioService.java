package ru.lambdahub.auth.service.scenario.lhauth_06;

import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import ru.lambdahub.auth.api.ResetVerifyRequest;
import ru.lambdahub.auth.api.ResetVerifyResponse;
import ru.lambdahub.auth.otp.OtpService;
import ru.lambdahub.auth.service.support.AuthSupportService;
import ru.lambdahub.auth.service.support.RequestIds;
import ru.lambdahub.auth.validation.LhauthValidInfo;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;
import ru.lambdahub.validation.validator.OutcomeValidator;

/** LHAUTH-06 — verify password-reset OTP */
@Slf4j
@Service
@RequiredArgsConstructor
public class Lhauth06ScenarioService {

    private final AuthSupportService support;
    private final OtpService otpService;
    private final OutcomeValidator vavrValidator;

    public ResetVerifyResponse process(ResetVerifyRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHAUTH-06] Starting scenario. requestId={} request={}", requestId, req);

        Outcome outcome = Outcome.of(vavrValidator.validate(req));
        String resetToken = null;
        if (!outcome.hasCriticalErrors()) {
            String login = req.getLogin().trim();
            otpService.verify("reset", login, req.getCode().trim());
            if (support.findByLogin(login).isEmpty()) {
                outcome = outcome.addValidCode(LhauthValidInfo.LHAUTH_06_0001, login);
            } else {
                resetToken = otpService.createResetToken(login.toLowerCase());
            }
        }
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }

        ResetVerifyResponse result = ResetVerifyResponse.builder()
                .requestId(requestId)
                .outcome(outcome)
                .resetToken(resetToken)
                .build();
        log.info("[LHAUTH-06] Scenario completed. requestId={}", requestId);
        return result;
    }
}
