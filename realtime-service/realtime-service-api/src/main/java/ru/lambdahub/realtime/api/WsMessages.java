package ru.lambdahub.realtime.api;

import jakarta.validation.Valid;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotEmpty;
import jakarta.validation.constraints.NotNull;

import java.util.List;

/** WebSocket client/server message contracts. */
public final class WsMessages {
    private WsMessages() {}

    public record Ping() {}
    public record Pong() {}
    public record Connected(String userId) {}

    public record SubscribeRequest(@NotEmpty List<@Valid @NotNull ChannelRef> channels) {}

    public record UnsubscribeRequest(@NotEmpty List<@Valid @NotNull ChannelRef> channels) {}

    public record ChannelRef(
            @NotBlank String kind,
            String id,
            List<String> userIds
    ) {}

    public record Subscribed() {}
    public record ErrorMessage(String message) {}
}
