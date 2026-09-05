package ru.lambdahub.chat.web;

import java.util.List;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.RestController;
import ru.lambdahub.chat.api.ChatListItem;
import ru.lambdahub.chat.api.ChatQueryApi;
import ru.lambdahub.chat.api.UpdateChatRequest;
import ru.lambdahub.chat.service.scenario.lhcht_03.Lhcht03ScenarioService;
import ru.lambdahub.chat.service.scenario.lhcht_04.Lhcht04ScenarioService;
import ru.lambdahub.common.security.UserPrincipal;

@RestController
@RequiredArgsConstructor
public class ChatQueryController implements ChatQueryApi {

    private final Lhcht03ScenarioService lhcht03;
    private final Lhcht04ScenarioService lhcht04;

    @Override
    public List<ChatListItem> list(UserPrincipal principal) {
        return lhcht03.process(principal.userId());
    }

    @Override
    public ChatListItem update(UserPrincipal principal, UUID chatId, UpdateChatRequest req) {
        return lhcht04.process(principal.userId(), chatId, req);
    }
}
