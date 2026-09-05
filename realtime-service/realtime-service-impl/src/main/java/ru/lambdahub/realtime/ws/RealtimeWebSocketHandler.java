package ru.lambdahub.realtime.ws;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import java.util.Map;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import org.springframework.cache.Cache;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Component;
import org.springframework.web.socket.CloseStatus;
import org.springframework.web.socket.TextMessage;
import org.springframework.web.socket.WebSocketSession;
import org.springframework.web.socket.handler.TextWebSocketHandler;
import ru.lambdahub.common.events.RedisChannels;
import ru.lambdahub.common.redis.cache.RedisCacheProviderService;
import ru.lambdahub.realtime.feign.InternalChatFeignClient;

@Component
@RequiredArgsConstructor
public class RealtimeWebSocketHandler extends TextWebSocketHandler {

    private final SessionRegistry registry;
    private final StringRedisTemplate redis;
    private final RedisCacheProviderService redisCacheProviderService;
    private final ObjectMapper objectMapper;
    private final InternalChatFeignClient chatClient;

    @Override
    public void afterConnectionEstablished(WebSocketSession session) throws Exception {
        UUID userId = userId(session);
        if (userId == null) {
            session.close(CloseStatus.NOT_ACCEPTABLE.withReason("unauthorized"));
            return;
        }

        registry.register(session, userId);
        redis.opsForSet().add("presence:online", userId.toString());
        presenceCache().put(userId.toString(), "online");
        registry.subscribe(session, RedisChannels.user(userId.toString()));

        redis.convertAndSend(
                RedisChannels.presence(userId.toString()),
                objectMapper.writeValueAsString(Map.of(
                        "type", "user_status",
                        "userId", userId.toString(),
                        "status", "online")));

        session.sendMessage(new TextMessage(objectMapper.writeValueAsString(
                Map.of("type", "connected", "userId", userId.toString()))));
    }

    @Override
    protected void handleTextMessage(WebSocketSession session, TextMessage message) throws Exception {
        UUID userId = registry.userId(session);
        if (userId == null) {
            return;
        }

        JsonNode root = objectMapper.readTree(message.getPayload());
        switch (root.path("type").asText()) {
            case "ping" -> {
                presenceCache().put(userId.toString(), "online");
                session.sendMessage(new TextMessage("{\"type\":\"pong\"}"));
            }
            case "subscribe" -> handleSubscribe(session, userId, root);
            case "unsubscribe" -> handleUnsubscribe(session, root);
            default -> session.sendMessage(new TextMessage("{\"type\":\"error\",\"message\":\"unknown type\"}"));
        }
    }

    @Override
    public void afterConnectionClosed(WebSocketSession session, CloseStatus status) throws Exception {
        UUID userId = registry.userId(session);
        registry.remove(session);
        if (userId == null) {
            return;
        }
        redis.opsForSet().remove("presence:online", userId.toString());
        presenceCache().evict(userId.toString());
        redis.convertAndSend(
                RedisChannels.presence(userId.toString()),
                objectMapper.writeValueAsString(Map.of(
                        "type", "user_status",
                        "userId", userId.toString(),
                        "status", "offline")));
    }

    private void handleSubscribe(WebSocketSession session, UUID userId, JsonNode root) throws Exception {
        JsonNode channels = root.path("channels");
        if (!channels.isArray()) {
            return;
        }
        for (JsonNode ch : channels) {
            String kind = ch.path("kind").asText();
            if ("chat".equals(kind)) {
                String chatId = ch.path("id").asText();
                if (isMember(chatId, userId)) {
                    registry.subscribe(session, RedisChannels.chat(chatId));
                }
            } else if ("presence".equals(kind)) {
                JsonNode ids = ch.path("userIds");
                if (ids.isArray()) {
                    for (JsonNode id : ids) {
                        registry.subscribe(session, RedisChannels.presence(id.asText()));
                    }
                }
            } else if ("self".equals(kind)) {
                registry.subscribe(session, RedisChannels.user(userId.toString()));
            }
        }
        session.sendMessage(new TextMessage("{\"type\":\"subscribed\"}"));
    }

    private void handleUnsubscribe(WebSocketSession session, JsonNode root) {
        JsonNode channels = root.path("channels");
        if (!channels.isArray()) {
            return;
        }
        for (JsonNode ch : channels) {
            String kind = ch.path("kind").asText();
            if ("chat".equals(kind)) {
                registry.unsubscribe(session, RedisChannels.chat(ch.path("id").asText()));
            } else if ("presence".equals(kind)) {
                JsonNode ids = ch.path("userIds");
                if (ids.isArray()) {
                    for (JsonNode id : ids) {
                        registry.unsubscribe(session, RedisChannels.presence(id.asText()));
                    }
                }
            }
        }
    }

    private UUID userId(WebSocketSession session) {
        Object value = session.getAttributes().get(JwtHandshakeInterceptor.ATTR_USER_ID);
        return value instanceof UUID id ? id : null;
    }

    private Cache presenceCache() {
        return redisCacheProviderService.getCache("presence")
                .orElseThrow(() -> new IllegalStateException("Cache not configured: presence"));
    }

    private boolean isMember(String chatId, UUID userId) {
        try {
            Map<String, Boolean> body = chatClient.checkMember(UUID.fromString(chatId), userId, null);
            return body != null && Boolean.TRUE.equals(body.get("member"));
        } catch (Exception e) {
            return false;
        }
    }
}
