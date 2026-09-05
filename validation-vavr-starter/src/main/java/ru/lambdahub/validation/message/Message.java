package ru.lambdahub.validation.message;

import java.util.UUID;

public interface Message {

    UUID getRequestId();

    void setRequestId(UUID reqId);
}
