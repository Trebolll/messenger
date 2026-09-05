package ru.lambdahub.chat.api;

import java.util.List;
import java.util.UUID;
import org.springframework.http.HttpStatus;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.ResponseStatus;
import ru.lambdahub.common.security.UserPrincipal;

public interface MessageApi {

    @PostMapping(ChatPaths.MESSAGES_BASE_API)
    @ResponseStatus(HttpStatus.CREATED)
    MessageView send(@AuthenticationPrincipal UserPrincipal principal,
                     @RequestBody SendMessageRequest req);

    @GetMapping(ChatPaths.BASE_API + "/{chatId}/messages")
    List<MessageView> messages(@AuthenticationPrincipal UserPrincipal principal,
                               @PathVariable UUID chatId);

    @PutMapping(ChatPaths.MESSAGES_BASE_API + "/{messageId}")
    MessageView edit(@AuthenticationPrincipal UserPrincipal principal,
                     @PathVariable UUID messageId,
                     @RequestBody EditMessageRequest req);

    @DeleteMapping(ChatPaths.MESSAGES_BASE_API + "/{messageId}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    void delete(@AuthenticationPrincipal UserPrincipal principal,
                @PathVariable UUID messageId);
}
