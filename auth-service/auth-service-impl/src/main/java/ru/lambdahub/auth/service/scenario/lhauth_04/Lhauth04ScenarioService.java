package ru.lambdahub.auth.service.scenario.lhauth_04;

import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import ru.lambdahub.auth.api.AuthResponse;
import ru.lambdahub.auth.api.LoginRequest;
import ru.lambdahub.auth.dto.CredentialDto;
import ru.lambdahub.auth.mapstruct.AuthMapper;
import ru.lambdahub.auth.service.support.AuthSupportService;
import ru.lambdahub.auth.service.support.RequestIds;
import ru.lambdahub.auth.validation.LhauthValidInfo;
import ru.lambdahub.common.security.JwtService;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;
import ru.lambdahub.validation.validator.OutcomeValidator;

/** LHAUTH-04 — login */
@Slf4j
@Service
@RequiredArgsConstructor
public class Lhauth04ScenarioService {

    private final AuthSupportService support;
    private final PasswordEncoder passwordEncoder;
    @Qualifier("sessionJwt")
    private final JwtService sessionJwt;
    private final OutcomeValidator vavrValidator;
    private final AuthMapper authMapper;

    public AuthResponse process(LoginRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHAUTH-04] Starting scenario. requestId={} request={}", requestId, req);

        Outcome outcome = Outcome.of(vavrValidator.validate(req));
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }

        String login = req.getLogin().trim();
        CredentialDto cred = support.findByLogin(login).getOrNull();
        if (cred == null || !passwordEncoder.matches(req.getPassword(), cred.getPasswordHash())) {
            outcome = outcome.addValidCode(LhauthValidInfo.LHAUTH_04_0001);
            if (outcome.hasCriticalErrors()) {
                throw ValidationExceptionUtils.createValidationException(outcome);
            }
        }

        String token = sessionJwt.issueSessionToken(cred.getUserId(), cred.getUsername());
        AuthResponse result = AuthResponse.builder()
                .requestId(requestId)
                .outcome(outcome)
                .token(token)
                .user(authMapper.toApi(cred, null))
                .build();
        log.info("[LHAUTH-04] Scenario completed. requestId={} userId={}", requestId, cred.getUserId());
        return result;
    }
}
