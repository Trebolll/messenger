package ru.lambdahub.auth.service.scenario.lhauth_05;

import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import ru.lambdahub.auth.api.ResetSendRequest;
import ru.lambdahub.auth.api.SendCodeResponse;
import ru.lambdahub.auth.otp.OtpService;
import ru.lambdahub.auth.service.support.AuthSupportService;
import ru.lambdahub.auth.service.support.RequestIds;
import ru.lambdahub.auth.validation.LhauthValidInfo;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;
import ru.lambdahub.validation.validator.OutcomeValidator;

/** LHAUTH-05 — send password-reset OTP */
@Slf4j
@Service
@RequiredArgsConstructor
public class Lhauth05ScenarioService {

    private final AuthSupportService support;
    private final OtpService otpService;
    private final OutcomeValidator vavrValidator;

    public SendCodeResponse process(ResetSendRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHAUTH-05] Starting scenario. requestId={} request={}", requestId, req);

        Outcome outcome = Outcome.of(vavrValidator.validate(req));
        String code = null;
        if (!outcome.hasCriticalErrors()) {
            String login = req.getLogin().trim();
            if (support.findByLogin(login).isEmpty()) {
                outcome = outcome.addValidCode(LhauthValidInfo.LHAUTH_05_0001, login);
            } else {
                code = otpService.send("reset", login);
            }
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
        log.info("[LHAUTH-05] Scenario completed. requestId={}", requestId);
        return result;
    }
}
