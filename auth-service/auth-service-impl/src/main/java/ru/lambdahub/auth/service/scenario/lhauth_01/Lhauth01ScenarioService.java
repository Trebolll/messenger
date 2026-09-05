package ru.lambdahub.auth.service.scenario.lhauth_01;

import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import ru.lambdahub.auth.api.SendCodeRequest;
import ru.lambdahub.auth.api.SendCodeResponse;
import ru.lambdahub.auth.otp.OtpService;
import ru.lambdahub.auth.service.support.RequestIds;
import ru.lambdahub.auth.validation.LhauthValidInfo;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;
import ru.lambdahub.validation.validator.OutcomeValidator;

/** LHAUTH-01 — send registration OTP */
@Slf4j
@Service
@RequiredArgsConstructor
public class Lhauth01ScenarioService {

    private final OtpService otpService;
    private final OutcomeValidator vavrValidator;

    public SendCodeResponse process(SendCodeRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHAUTH-01] Starting scenario. requestId={} request={}", requestId, req);

        Outcome outcome = Outcome.of(vavrValidator.validate(req));
        String login = req.getLogin() == null ? null : req.getLogin().trim();
        if (login == null || login.isBlank()) {
            outcome = outcome.addValidCode(LhauthValidInfo.LHAUTH_01_0001);
        }
        String code = null;
        if (!outcome.hasCriticalErrors()) {
            code = otpService.send("reg", login);
        }
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }

        SendCodeResponse result = SendCodeResponse.builder()
                .requestId(requestId)
                .outcome(outcome)
                .ok(true)
                .message("Code sent")
                .debugCode(code)
                .build();
        log.info("[LHAUTH-01] Scenario completed. requestId={}", requestId);
        return result;
    }
}
