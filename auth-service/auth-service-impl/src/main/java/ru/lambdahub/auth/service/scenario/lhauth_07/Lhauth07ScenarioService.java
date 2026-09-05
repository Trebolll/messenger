package ru.lambdahub.auth.service.scenario.lhauth_07;

import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.auth.api.ResetConfirmRequest;
import ru.lambdahub.auth.db.CredentialRepository;
import ru.lambdahub.auth.dto.CredentialDto;
import ru.lambdahub.auth.mapstruct.AuthMapper;
import ru.lambdahub.auth.otp.OtpService;
import ru.lambdahub.auth.service.support.AuthSupportService;
import ru.lambdahub.auth.service.support.RequestIds;
import ru.lambdahub.auth.validation.LhauthValidInfo;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;
import ru.lambdahub.validation.validator.OutcomeValidator;

/** LHAUTH-07 — confirm password reset */
@Slf4j
@Service
@RequiredArgsConstructor
public class Lhauth07ScenarioService {

    private final AuthSupportService support;
    private final CredentialRepository credentials;
    private final OtpService otpService;
    private final PasswordEncoder passwordEncoder;
    private final OutcomeValidator vavrValidator;
    private final AuthMapper authMapper;

    @Transactional
    public void process(ResetConfirmRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHAUTH-07] Starting scenario. requestId={} request={}", requestId, req);

        Outcome outcome = Outcome.of(vavrValidator.validate(req));
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }

        String login = otpService.consumeResetToken(req.getResetToken());
        CredentialDto cred = support.findByLogin(login).getOrNull();
        if (cred == null) {
            outcome = outcome.addValidCode(LhauthValidInfo.LHAUTH_07_0001, login);
            if (outcome.hasCriticalErrors()) {
                throw ValidationExceptionUtils.createValidationException(outcome);
            }
            return;
        }

        cred.setPasswordHash(passwordEncoder.encode(req.getPassword()));
        credentials.save(authMapper.toEntity(cred));
        log.info("[LHAUTH-07] Scenario completed. requestId={} userId={}", requestId, cred.getUserId());
    }
}
