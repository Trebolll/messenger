package ru.lambdahub.realtime.config;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;
import ru.lambdahub.common.security.JwtService;

@Configuration
public class JwtConfig {

    @Bean
    @Primary
    public JwtService jwtService(@Value("${app.jwt-secret}") String secret) {
        return new JwtService(secret, 86400);
    }
}
