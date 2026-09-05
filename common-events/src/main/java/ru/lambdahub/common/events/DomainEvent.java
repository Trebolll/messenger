package ru.lambdahub.common.events;

import com.fasterxml.jackson.annotation.JsonInclude;

import java.time.Instant;
import java.util.Map;
import java.util.UUID;

@JsonInclude(JsonInclude.Include.NON_NULL)
public record DomainEvent(
        UUID eventId,
        Instant occurredAt,
        String producer,
        int version,
        String type,
        Map<String, Object> payload
) {
    public static DomainEvent of(String producer, String type, Map<String, Object> payload) {
        return new DomainEvent(UUID.randomUUID(), Instant.now(), producer, 1, type, payload);
    }
}
