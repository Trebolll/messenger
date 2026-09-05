# common-security

Shared JWT helpers + Spring Security auto-config for servlet services.

## Behavior

- `JwtOrHeaderAuthenticationFilter` builds `UserPrincipal` from:
  1. `Authorization: Bearer <jwt>`, or
  2. gateway headers `X-User-Id` / `X-Username`
- `SecurityFilterChain` (stateless): public `/api/auth/**`, `/actuator/health`, `/api/internal/**`; rest authenticated
- Disable with `lambdahub.security.enabled=false`

## Usage

```java
@GetMapping("/me")
UserDtos.ProfileResponse me(@AuthenticationPrincipal UserPrincipal principal) {
    return userService.get(principal.userId());
}
```

```java
SecurityUtils.currentAuditor().orElse("system");
```

WebFlux gateway should exclude `spring-boot-starter-security` from this dependency.

Build: `mvn -DskipTests package`
