package ru.lambdahub.auth.service.scenario.lhauth_02;

import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Service;
import ru.lambdahub.auth.api.VerifyCodeRequest;
import ru.lambdahub.auth.api.VerifyCodeResponse;
import ru.lambdahub.auth.otp.OtpService;
import ru.lambdahub.auth.service.support.RequestIds;
import ru.lambdahub.common.security.JwtService;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;
import ru.lambdahub.validation.validator.OutcomeValidator;

/** LHAUTH-02 — verify registration OTP */
@Slf4j
@Service
@RequiredArgsConstructor
public class Lhauth02ScenarioService {

    private final OtpService otpService;
    @Qualifier("confirmJwt")
    private final JwtService confirmJwt;
    private final OutcomeValidator vavrValidator;

    public VerifyCodeResponse process(VerifyCodeRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHAUTH-02] Starting scenario. requestId={} request={}", requestId, req);

        Outcome outcome = Outcome.of(vavrValidator.validate(req));
        String token = null;
        if (!outcome.hasCriticalErrors()) {
            String login = req.getLogin().trim();
            otpService.verify("reg", login, req.getCode().trim());
            token = confirmJwt.issueConfirmToken(login.toLowerCase(), 900);
        }
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }

        VerifyCodeResponse result = VerifyCodeResponse.builder()
                .requestId(requestId)
                .outcome(outcome)
                .confirmToken(token)
                .build();
        log.info("[LHAUTH-02] Scenario completed. requestId={}", requestId);
        return result;
    }
}
