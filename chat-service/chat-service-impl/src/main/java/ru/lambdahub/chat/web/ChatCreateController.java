package ru.lambdahub.chat.web;

import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.RestController;
import ru.lambdahub.chat.api.ChatCreateApi;
import ru.lambdahub.chat.api.ChatListItem;
import ru.lambdahub.chat.api.CreateGroupRequest;
import ru.lambdahub.chat.api.CreatePrivateRequest;
import ru.lambdahub.chat.service.scenario.lhcht_01.Lhcht01ScenarioService;
import ru.lambdahub.chat.service.scenario.lhcht_02.Lhcht02ScenarioService;
import ru.lambdahub.common.security.UserPrincipal;

@RestController
@RequiredArgsConstructor
public class ChatCreateController implements ChatCreateApi {

    private final Lhcht01ScenarioService lhcht01;
    private final Lhcht02ScenarioService lhcht02;

    @Override
    public ChatListItem createPrivate(UserPrincipal principal, CreatePrivateRequest req) {
        return lhcht01.process(principal.userId(), req);
    }

    @Override
    public ChatListItem createGroup(UserPrincipal principal, CreateGroupRequest req) {
        return lhcht02.process(principal.userId(), req);
    }
}
