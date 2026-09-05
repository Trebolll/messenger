package ru.lambdahub.gateway;

import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.UUID;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.cloud.gateway.filter.GlobalFilter;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.Ordered;
import org.springframework.core.io.buffer.DataBuffer;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpMethod;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.web.server.ServerWebExchange;
import reactor.core.publisher.Mono;
import ru.lambdahub.common.security.AuthHeaders;
import ru.lambdahub.common.security.JwtService;

@Configuration
public class GatewayConfig {

  @Bean
  public JwtService jwtService(@Value("${JWT_SECRET:changeme-jwt-secret-key-32chars-min}") String secret) {
    return new JwtService(secret, 86400);
  }

  @Bean
  public GlobalFilter jwtAuthFilter(JwtService jwtService) {
    return new JwtAuthFilter(jwtService);
  }

  /**
   * /api/ws must be a real WebSocket upgrade. Plain HTTP GET is forwarded to Tomcat and fails with
   * {@code Can "Upgrade" only to "WebSocket"}.
   */
  @Bean
  public GlobalFilter webSocketUpgradeFilter() {
    return new WebSocketUpgradeFilter();
  }

  static class WebSocketUpgradeFilter implements GlobalFilter, Ordered {

    @Override
    public Mono<Void> filter(ServerWebExchange exchange,
        org.springframework.cloud.gateway.filter.GatewayFilterChain chain) {
      String path = exchange.getRequest().getURI().getPath();
      if (!path.equals("/api/ws") && !path.startsWith("/api/ws/")) {
        return chain.filter(exchange);
      }
      String upgrade = exchange.getRequest().getHeaders().getUpgrade();
      if (upgrade != null && "websocket".equalsIgnoreCase(upgrade.trim())) {
        return chain.filter(exchange);
      }
      exchange.getResponse().setStatusCode(HttpStatus.UPGRADE_REQUIRED);
      exchange.getResponse().getHeaders().setContentType(MediaType.TEXT_PLAIN);
      byte[] body = ("Use a WebSocket client: ws://localhost:8080/api/ws?token=<jwt> "
          + "(not HTTP GET).").getBytes(StandardCharsets.UTF_8);
      DataBuffer buffer = exchange.getResponse().bufferFactory().wrap(body);
      return exchange.getResponse().writeWith(Mono.just(buffer));
    }

    @Override
    public int getOrder() {
      return -200;
    }
  }

  static class JwtAuthFilter implements GlobalFilter, Ordered {

    private static final List<String> PUBLIC_PREFIXES = List.of(
        "/api/auth/",
        "/api/ws",
        "/actuator/health"
    );

    private final JwtService jwtService;

    JwtAuthFilter(JwtService jwtService) {
      this.jwtService = jwtService;
    }

    @Override
    public Mono<Void> filter(ServerWebExchange exchange,
        org.springframework.cloud.gateway.filter.GatewayFilterChain chain) {
      String path = exchange.getRequest().getURI().getPath();
      if (HttpMethod.OPTIONS.equals(exchange.getRequest().getMethod())) {
        return chain.filter(exchange);
      }
      for (String prefix : PUBLIC_PREFIXES) {
        if (path.startsWith(prefix)) {
          return chain.filter(exchange);
        }
      }
      if (!path.startsWith("/api/")) {
        return chain.filter(exchange);
      }

      String token = extractToken(exchange);
      if (token == null || token.isBlank()) {
        exchange.getResponse().setStatusCode(HttpStatus.UNAUTHORIZED);
        return exchange.getResponse().setComplete();
      }
      try {
        UUID userId = jwtService.requireUserId(token);
        var claims = jwtService.parse(token);
        Object username = claims.get("username");
        ServerWebExchange mutated = exchange.mutate()
            .request(builder -> builder
                .header(AuthHeaders.USER_ID, userId.toString())
                .header(AuthHeaders.USERNAME, username == null ? "" : username.toString()))
            .build();
        return chain.filter(mutated);
      } catch (Exception e) {
        exchange.getResponse().setStatusCode(HttpStatus.UNAUTHORIZED);
        return exchange.getResponse().setComplete();
      }
    }

    private String extractToken(ServerWebExchange exchange) {
      String auth = exchange.getRequest().getHeaders().getFirst(HttpHeaders.AUTHORIZATION);
      if (auth != null && auth.startsWith("Bearer ")) {
        return auth.substring(7);
      }
      return exchange.getRequest().getQueryParams().getFirst("token");
    }

    @Override
    public int getOrder() {
      return -100;
    }
  }
}
