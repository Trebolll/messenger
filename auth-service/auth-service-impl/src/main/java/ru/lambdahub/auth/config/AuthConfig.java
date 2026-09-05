package ru.lambdahub.auth.config;

import feign.RequestInterceptor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;
import ru.lambdahub.common.security.AuthHeaders;
import ru.lambdahub.common.security.JwtService;

@Configuration
public class AuthConfig {

  @Bean
  @Primary
  public JwtService sessionJwt(@Value("${app.jwt-secret}") String secret,
      @Value("${app.jwt-ttl-seconds}") long ttl) {
    return new JwtService(secret, ttl);
  }

  @Bean
  public JwtService confirmJwt(@Value("${app.confirm-secret}") String secret,
      @Value("${app.confirm-ttl-seconds}") long ttl) {
    return new JwtService(secret, ttl);
  }

  /** Default Feign interceptor: X-Internal-Key from app.internal-key (YAML). */
  @Bean
  public RequestInterceptor feignInternalKeyInterceptor(
      @Value("${app.internal-key}") String internalKey) {
    return template -> template.header(AuthHeaders.INTERNAL_KEY, internalKey);
  }
}
