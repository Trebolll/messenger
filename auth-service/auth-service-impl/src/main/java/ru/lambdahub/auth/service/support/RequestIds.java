package ru.lambdahub.auth.service.support;

import java.util.UUID;
import ru.lambdahub.validation.message.InboundMessage;

public final class RequestIds {

    private RequestIds() {}

    public static UUID requestId(InboundMessage message) {
        if (message != null && message.getRequestId() != null) {
            return message.getRequestId();
        }
        return UUID.randomUUID();
    }
}
