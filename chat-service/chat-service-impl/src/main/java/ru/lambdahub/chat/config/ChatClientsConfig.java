package ru.lambdahub.chat.config;

import feign.RequestInterceptor;
import java.util.UUID;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import ru.lambdahub.common.security.AuthHeaders;

@Configuration
public class ChatClientsConfig {

    /** Default Feign interceptor: headers from YAML + service-call principal. */
    @Bean
    public RequestInterceptor feignRequestInterceptor(
            @Value("${app.internal-key}") String internalKey) {
        return template -> {
            template.header(AuthHeaders.INTERNAL_KEY, internalKey);
            template.header(AuthHeaders.USER_ID, UUID.randomUUID().toString());
        };
    }
}
