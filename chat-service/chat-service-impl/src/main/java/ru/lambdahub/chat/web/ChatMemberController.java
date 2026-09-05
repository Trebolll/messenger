package ru.lambdahub.chat.web;

import java.util.List;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.RestController;
import ru.lambdahub.chat.api.AddMemberRequest;
import ru.lambdahub.chat.api.ChatMemberApi;
import ru.lambdahub.chat.api.MemberView;
import ru.lambdahub.chat.service.scenario.lhcht_05.Lhcht05ScenarioService;
import ru.lambdahub.chat.service.scenario.lhcht_06.Lhcht06ScenarioService;
import ru.lambdahub.chat.service.scenario.lhcht_07.Lhcht07ScenarioService;
import ru.lambdahub.common.security.UserPrincipal;

@RestController
@RequiredArgsConstructor
public class ChatMemberController implements ChatMemberApi {

    private final Lhcht05ScenarioService lhcht05;
    private final Lhcht06ScenarioService lhcht06;
    private final Lhcht07ScenarioService lhcht07;

    @Override
    public List<MemberView> members(UserPrincipal principal, UUID chatId) {
        return lhcht05.process(principal.userId(), chatId);
    }

    @Override
    public void addMember(UserPrincipal principal, UUID chatId, AddMemberRequest req) {
        lhcht06.process(principal.userId(), chatId, req);
    }

    @Override
    public void removeMember(UserPrincipal principal, UUID chatId, UUID memberId) {
        lhcht07.process(principal.userId(), chatId, memberId);
    }
}
