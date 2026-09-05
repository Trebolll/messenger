package ru.lambdahub.chat.web;

import java.util.Map;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.RestController;
import ru.lambdahub.chat.api.InternalChatApi;
import ru.lambdahub.chat.service.scenario.lhcht_13.Lhcht13ScenarioService;

@RestController
@RequiredArgsConstructor
public class InternalChatController implements InternalChatApi {

    private final Lhcht13ScenarioService lhcht13;

    @Override
    public Map<String, Boolean> checkMember(UUID chatId, UUID userId, String key) {
        return lhcht13.process(chatId, userId);
    }
}
