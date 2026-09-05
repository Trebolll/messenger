package ru.lambdahub.common.security;

import java.util.UUID;

/**
 * Authenticated user in {@link org.springframework.security.core.context.SecurityContext}
 * (analogue of Impulse {@code UserLoginData}).
 */
public record UserPrincipal(UUID userId, String username) {
}
