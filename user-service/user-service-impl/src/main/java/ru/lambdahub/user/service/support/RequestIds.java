package ru.lambdahub.user.service.support;

import java.util.UUID;
import ru.lambdahub.validation.message.InboundMessage;

public final class RequestIds {

    private RequestIds() {
    }

    public static UUID requestId(InboundMessage msg) {
        if (msg == null) {
            return UUID.randomUUID();
        }
        if (msg.getRequestId() == null) {
            UUID id = UUID.randomUUID();
            msg.setRequestId(id);
            return id;
        }
        return msg.getRequestId();
    }
}
