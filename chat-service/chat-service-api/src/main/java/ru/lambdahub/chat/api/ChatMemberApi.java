package ru.lambdahub.chat.api;

import java.util.List;
import java.util.UUID;
import org.springframework.http.HttpStatus;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.ResponseStatus;
import ru.lambdahub.common.security.UserPrincipal;

public interface ChatMemberApi {

    @GetMapping(ChatPaths.BASE_API + "/{chatId}/members")
    List<MemberView> members(@AuthenticationPrincipal UserPrincipal principal,
                             @PathVariable UUID chatId);

    @PostMapping(ChatPaths.BASE_API + "/{chatId}/members")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    void addMember(@AuthenticationPrincipal UserPrincipal principal,
                   @PathVariable UUID chatId,
                   @RequestBody AddMemberRequest req);

    @DeleteMapping(ChatPaths.BASE_API + "/{chatId}/members/{memberId}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    void removeMember(@AuthenticationPrincipal UserPrincipal principal,
                      @PathVariable UUID chatId,
                      @PathVariable UUID memberId);
}
