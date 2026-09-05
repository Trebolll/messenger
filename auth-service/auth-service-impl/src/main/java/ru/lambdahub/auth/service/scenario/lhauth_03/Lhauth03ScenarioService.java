package ru.lambdahub.auth.service.scenario.lhauth_03;

import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.auth.api.AuthResponse;
import ru.lambdahub.auth.api.RegisterRequest;
import ru.lambdahub.auth.db.CredentialRepository;
import ru.lambdahub.auth.dto.CredentialDto;
import ru.lambdahub.auth.feign.InternalUsersFeignClient;
import ru.lambdahub.auth.mapstruct.AuthMapper;
import ru.lambdahub.auth.service.support.RequestIds;
import ru.lambdahub.auth.validation.LhauthValidInfo;
import ru.lambdahub.common.events.kafka.DtoKafkaUserCreated;
import ru.lambdahub.common.kafka.OutputBridge;
import ru.lambdahub.common.security.JwtService;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;
import ru.lambdahub.validation.validator.OutcomeValidator;

/** LHAUTH-03 — register */
@Slf4j
@Service
@RequiredArgsConstructor
public class Lhauth03ScenarioService {

    private final CredentialRepository credentials;
    private final PasswordEncoder passwordEncoder;
    @Qualifier("sessionJwt")
    private final JwtService sessionJwt;
    @Qualifier("confirmJwt")
    private final JwtService confirmJwt;
    private final ObjectProvider<OutputBridge> userCreatedOutput;
    private final InternalUsersFeignClient internalUsersClient;
    private final OutcomeValidator vavrValidator;
    private final AuthMapper authMapper;

    @Transactional
    public AuthResponse process(RegisterRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHAUTH-03] Starting scenario. requestId={} request={}", requestId, req);

        Outcome outcome = Outcome.of(vavrValidator.validate(req));
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }

        String login = confirmJwt.requireConfirmLogin(req.getConfirmToken());
        if (login == null || login.isBlank()) {
            outcome = outcome.addValidCode(LhauthValidInfo.LHAUTH_03_0002);
        }
        if (credentials.existsByUsernameIgnoreCase(req.getUsername())) {
            outcome = outcome.addValidCode(LhauthValidInfo.LHAUTH_03_0001, req.getUsername());
        }
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }

        UUID userId = UUID.randomUUID();
        CredentialDto cred = authMapper.fromRequest(req);
        cred.setUserId(userId);
        if (login.contains("@")) {
            cred.setEmail(login);
        } else {
            cred.setPhone(login);
        }
        cred.setPasswordHash(passwordEncoder.encode(req.getPassword()));

        String fullName = req.getFullName() == null ? cred.getUsername() : req.getFullName();
        DtoKafkaUserCreated created = authMapper.toKafka(cred, requestId, fullName);
        credentials.save(authMapper.toEntity(cred));
        userCreatedOutput.ifAvailable(ob -> ob.send(created));
        try {
            internalUsersClient.create(null, created);
        } catch (Exception ignored) {
            // profile/user created asynchronously via Kafka if sync call fails
        }

        String token = sessionJwt.issueSessionToken(userId, cred.getUsername());
        AuthResponse result = AuthResponse.builder()
                .requestId(requestId)
                .outcome(outcome)
                .token(token)
                .user(authMapper.toApi(cred, fullName))
                .build();
        log.info("[LHAUTH-03] Scenario completed. requestId={} userId={}", requestId, userId);
        return result;
    }
}
