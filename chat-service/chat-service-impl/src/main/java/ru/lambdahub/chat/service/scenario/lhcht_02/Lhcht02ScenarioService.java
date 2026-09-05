package ru.lambdahub.chat.service.scenario.lhcht_02;

import java.util.ArrayList;
import java.util.List;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.lambdahub.chat.api.ChatListItem;
import ru.lambdahub.chat.api.CreateGroupRequest;
import ru.lambdahub.chat.service.support.ChatSupportService;
import ru.lambdahub.chat.service.support.RequestIds;
import ru.lambdahub.chat.validation.LhchtValidInfo;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;

@Slf4j
@Service
@RequiredArgsConstructor
public class Lhcht02ScenarioService {

    private final ChatSupportService support;

    @Transactional
    public ChatListItem process(UUID me, CreateGroupRequest req) {
        UUID requestId = RequestIds.requestId(req);
        log.info("[LHCHT-02] Starting scenario. requestId={} request={}", requestId, req);

        String name = req.getName();
        if (name == null || name.isBlank()) {
            Outcome outcome = new Outcome().addValidCode(LhchtValidInfo.LHCHT_02_0001);
            throw ValidationExceptionUtils.createValidationException(outcome);
        }
        List<UUID> memberIds = new ArrayList<>();
        memberIds.add(me);
        List<String> usernames = req.getUsernames();
        if (usernames != null) {
            for (String username : usernames) {
                UUID id = support.resolveUsername(username);
                if (!memberIds.contains(id)) {
                    memberIds.add(id);
                }
            }
        }
        ChatListItem result = support.createChat("group", name.trim(), me, memberIds, requestId);
        log.info("[LHCHT-02] Scenario completed. requestId={}", requestId);
        return result;
    }
}
