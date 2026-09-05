package ru.lambdahub.chat.web;

import java.util.UUID;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.RestController;
import ru.lambdahub.chat.api.ChatReadApi;
import ru.lambdahub.chat.api.MarkReadRequest;
import ru.lambdahub.chat.service.scenario.lhcht_12.Lhcht12ScenarioService;
import ru.lambdahub.common.security.UserPrincipal;

@RestController
@RequiredArgsConstructor
public class ChatReadController implements ChatReadApi {

    private final Lhcht12ScenarioService lhcht12;

    @Override
    public void markRead(UserPrincipal principal, UUID chatId, MarkReadRequest req) {
        lhcht12.process(principal.userId(), chatId, req);
    }
}
