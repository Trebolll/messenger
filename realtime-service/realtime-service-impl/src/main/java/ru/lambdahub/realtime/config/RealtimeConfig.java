package ru.lambdahub.realtime.config;

import feign.RequestInterceptor;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.data.redis.connection.RedisConnectionFactory;
import org.springframework.data.redis.listener.PatternTopic;
import org.springframework.data.redis.listener.RedisMessageListenerContainer;
import org.springframework.web.socket.config.annotation.EnableWebSocket;
import org.springframework.web.socket.config.annotation.WebSocketConfigurer;
import org.springframework.web.socket.config.annotation.WebSocketHandlerRegistry;
import ru.lambdahub.common.events.RedisChannels;
import ru.lambdahub.common.security.AuthHeaders;
import ru.lambdahub.realtime.ws.JwtHandshakeInterceptor;
import ru.lambdahub.realtime.ws.RealtimeWebSocketHandler;
import ru.lambdahub.realtime.ws.SessionRegistry;

@Configuration
@EnableWebSocket
@RequiredArgsConstructor
public class RealtimeConfig implements WebSocketConfigurer {

    private final RealtimeWebSocketHandler handler;
    private final JwtHandshakeInterceptor jwtHandshakeInterceptor;

    /** Default Feign interceptor: X-Internal-Key from app.internal-key (YAML). */
    @Bean
    public RequestInterceptor feignInternalKeyInterceptor(
            @Value("${app.internal-key}") String internalKey) {
        return template -> template.header(AuthHeaders.INTERNAL_KEY, internalKey);
    }

    @Bean
    public RedisMessageListenerContainer redisContainer(RedisConnectionFactory factory,
                                                        SessionRegistry registry) {
        RedisMessageListenerContainer container = new RedisMessageListenerContainer();
        container.setConnectionFactory(factory);
        container.addMessageListener((message, pattern) -> {
            String channel = new String(message.getChannel());
            String body = new String(message.getBody());
            registry.deliver(channel, body);
        }, new PatternTopic(RedisChannels.PATTERN_ALL));
        return container;
    }

    @Override
    public void registerWebSocketHandlers(WebSocketHandlerRegistry registry) {
        registry.addHandler(handler, "/api/ws")
                .addInterceptors(jwtHandshakeInterceptor)
                .setAllowedOrigins("*");
    }
}
