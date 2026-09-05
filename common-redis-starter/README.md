# common-redis-starter

Redis connection defaults + optional Spring Cache (`cache-config`).

## Connection

```yaml
spring:
  data:
    redis:
      host: localhost
      port: 6379
```

## Named caches (per-cache TTL in YAML)

```yaml
cache-config:
  enabled: true
  inspection-enabled: ${CACHE_INSPECTION_ENABLED:true}
  cache-name-prefix: lambdahub:auth.
  defaults:
    cache-null-values: false
  caches:
    credential-by-login:
      time-to-live: ${CREDENTIAL_CACHE_TTL:5m}
```

```java
@Service
@RequiredArgsConstructor
public class SomeService {
  private final RedisCacheProviderService redisCacheProviderService;

  public Optional<Cache> credentialCache() {
    return redisCacheProviderService.getCache("credential-by-login");
  }
}
```

Usage:

```java
var cached = redisCacheProviderService.getCache("credential-by-login")
    .map(c -> c.get(login, UserView.class))
    .orElse(null);
```
