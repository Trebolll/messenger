package ru.lambdahub.chat.api;

import java.util.Map;
import java.util.UUID;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestHeader;
import ru.lambdahub.common.security.AuthHeaders;

/**
 * Feign-friendly internal chat endpoints (no {@code @AuthenticationPrincipal}).
 */
public interface InternalChatApi {

    String MEMBER_CHECK = "/api/internal/chats/{chatId}/members/{userId}/check";

    @GetMapping(MEMBER_CHECK)
    Map<String, Boolean> checkMember(@PathVariable UUID chatId,
                                     @PathVariable UUID userId,
                                     @RequestHeader(value = AuthHeaders.INTERNAL_KEY, required = false) String key);
}
