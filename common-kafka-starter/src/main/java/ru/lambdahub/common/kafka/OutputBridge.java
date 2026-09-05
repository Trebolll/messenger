package ru.lambdahub.common.kafka;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.Executor;
import java.util.concurrent.Future;
import java.util.function.Function;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.cloud.stream.function.StreamBridge;
import org.springframework.kafka.support.KafkaHeaders;
import org.springframework.messaging.MessageHeaders;
import org.springframework.messaging.support.MessageBuilder;
import ru.lambdahub.common.events.DomainEvent;
import ru.lambdahub.common.events.kafka.KafkaDto;

/**
 * Sends messages through Spring Cloud Stream bindings via {@link StreamBridge}.
 * Bean per output binding group is registered from {@code kafka-bindings.yml}
 * (e.g. binding {@code userCreatedOutput-out-0} → bean {@code userCreatedOutput}).
 * Prefer typed {@link KafkaDto} payloads: {@code output.ifAvailable(ob -> ob.send(dto))}.
 */
public abstract class OutputBridge implements Function<org.springframework.messaging.Message<?>, Future<Void>> {

    private static final Logger log = LoggerFactory.getLogger(OutputBridge.class);

    private final StreamBridge streamBridge;
    private final Executor executor;

    protected OutputBridge(StreamBridge streamBridge, Executor executor) {
        this.streamBridge = streamBridge;
        this.executor = executor;
    }

    public <T> Future<Void> send(T data) {
        return send(data, Map.of());
    }

    /** Send with explicit Kafka partition key. */
    public <T> Future<Void> send(T data, Object kafkaKey) {
        if (kafkaKey == null) {
            return send(data);
        }
        return send(data, Map.of(KafkaHeaders.KEY, kafkaKey.toString()));
    }

    public <T> Future<Void> send(T data, Map<String, Object> additionalHeaders) {
        Map<String, Object> sourceHeaders = new HashMap<>(additionalHeaders);
        sourceHeaders.putIfAbsent(MessageHeaders.CONTENT_TYPE, "application/json");

        if (data instanceof KafkaDto kafkaDto && kafkaDto.kafkaKey() != null) {
            sourceHeaders.putIfAbsent(KafkaHeaders.KEY, kafkaDto.kafkaKey().toString());
        } else if (data instanceof DomainEvent event) {
            sourceHeaders.putIfAbsent(KafkaHeaders.KEY, event.eventId().toString());
        }

        return apply(MessageBuilder.createMessage(data, new MessageHeaders(sourceHeaders)));
    }

    @Override
    public Future<Void> apply(org.springframework.messaging.Message<?> message) {
        List<CompletableFuture<Void>> futures = new ArrayList<>();
        for (String channelId : channels()) {
            CompletableFuture<Void> future = CompletableFuture.runAsync(() -> {
                try {
                    log.debug("Sending data to channel {}", channelId);
                    boolean sent = streamBridge.send(channelId, message);
                    if (!sent) {
                        throw new IllegalStateException("Failed to send message to channel " + channelId);
                    }
                    log.debug("Sent data to channel {}", channelId);
                } catch (Exception exception) {
                    log.warn("Failed to send message to channel {}", channelId, exception);
                    throw new IllegalStateException(exception);
                }
            }, executor);
            futures.add(future);
        }
        return CompletableFuture.allOf(futures.toArray(new CompletableFuture[0])).thenApply(value -> null);
    }

    public abstract List<String> channels();
}
