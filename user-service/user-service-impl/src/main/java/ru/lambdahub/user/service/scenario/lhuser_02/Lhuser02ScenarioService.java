package ru.lambdahub.user.service.scenario.lhuser_02;

import java.util.List;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import ru.lambdahub.user.api.ProfileResponse;
import ru.lambdahub.user.db.ProfileRepository;
import ru.lambdahub.user.mapstruct.UserMapper;

/** LHUSER-02 — search profiles */
@Slf4j
@Service
@RequiredArgsConstructor
public class Lhuser02ScenarioService {

    private final ProfileRepository profiles;
    private final UserMapper userMapper;

    public List<ProfileResponse> process(String q) {
        UUID requestId = UUID.randomUUID();
        log.info("[LHUSER-02] Starting scenario. requestId={} q={}", requestId, q);
        if (q == null || q.isBlank()) {
            log.info("[LHUSER-02] Scenario completed. requestId={} count=0", requestId);
            return List.of();
        }
        List<ProfileResponse> result = profiles.search(q.trim())
                .map(userMapper::toDto)
                .map(userMapper::toApi)
                .toJavaList();
        log.info("[LHUSER-02] Scenario completed. requestId={} count={}", requestId, result.size());
        return result;
    }
}
