package ru.lambdahub.chat.web;

import java.util.List;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.RestController;
import ru.lambdahub.chat.api.EditMessageRequest;
import ru.lambdahub.chat.api.MessageApi;
import ru.lambdahub.chat.api.MessageView;
import ru.lambdahub.chat.api.SendMessageRequest;
import ru.lambdahub.chat.service.scenario.lhcht_08.Lhcht08ScenarioService;
import ru.lambdahub.chat.service.scenario.lhcht_09.Lhcht09ScenarioService;
import ru.lambdahub.chat.service.scenario.lhcht_10.Lhcht10ScenarioService;
import ru.lambdahub.chat.service.scenario.lhcht_11.Lhcht11ScenarioService;
import ru.lambdahub.common.security.UserPrincipal;

@RestController
@RequiredArgsConstructor
public class MessageController implements MessageApi {

    private final Lhcht08ScenarioService lhcht08;
    private final Lhcht09ScenarioService lhcht09;
    private final Lhcht10ScenarioService lhcht10;
    private final Lhcht11ScenarioService lhcht11;

    @Override
    public MessageView send(UserPrincipal principal, SendMessageRequest req) {
        return lhcht08.process(principal.userId(), req);
    }

    @Override
    public List<MessageView> messages(UserPrincipal principal, UUID chatId) {
        return lhcht09.process(principal.userId(), chatId);
    }

    @Override
    public MessageView edit(UserPrincipal principal, UUID messageId, EditMessageRequest req) {
        return lhcht10.process(principal.userId(), messageId, req);
    }

    @Override
    public void delete(UserPrincipal principal, UUID messageId) {
        lhcht11.process(principal.userId(), messageId);
    }
}
