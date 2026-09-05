package ru.lambdahub.chat.api;

import java.util.List;
import java.util.UUID;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import ru.lambdahub.common.security.UserPrincipal;

public interface ChatQueryApi {

    @GetMapping(ChatPaths.BASE_API)
    List<ChatListItem> list(@AuthenticationPrincipal UserPrincipal principal);

    @PutMapping(ChatPaths.BASE_API + "/{chatId}")
    ChatListItem update(@AuthenticationPrincipal UserPrincipal principal,
                        @PathVariable UUID chatId,
                        @RequestBody UpdateChatRequest req);
}
