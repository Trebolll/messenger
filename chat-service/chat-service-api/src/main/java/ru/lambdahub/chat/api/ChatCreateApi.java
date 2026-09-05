package ru.lambdahub.chat.api;

import org.springframework.http.HttpStatus;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.ResponseStatus;
import ru.lambdahub.common.security.UserPrincipal;

public interface ChatCreateApi {

    @PostMapping(ChatPaths.BASE_API + "/private")
    @ResponseStatus(HttpStatus.CREATED)
    ChatListItem createPrivate(@AuthenticationPrincipal UserPrincipal principal,
                               @RequestBody CreatePrivateRequest req);

    @PostMapping(ChatPaths.BASE_API + "/group")
    @ResponseStatus(HttpStatus.CREATED)
    ChatListItem createGroup(@AuthenticationPrincipal UserPrincipal principal,
                             @RequestBody CreateGroupRequest req);
}
