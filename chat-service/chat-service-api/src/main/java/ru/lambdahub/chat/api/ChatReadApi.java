package ru.lambdahub.chat.api;

import java.util.UUID;
import org.springframework.http.HttpStatus;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.ResponseStatus;
import ru.lambdahub.common.security.UserPrincipal;

public interface ChatReadApi {

    @PostMapping(ChatPaths.BASE_API + "/{chatId}/read")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    void markRead(@AuthenticationPrincipal UserPrincipal principal,
                  @PathVariable UUID chatId,
                  @RequestBody(required = false) MarkReadRequest req);
}
