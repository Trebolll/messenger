package ru.lambdahub.common.security;

import java.util.Optional;
import java.util.UUID;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContextHolder;

public final class SecurityUtils {

    private SecurityUtils() {}

    public static Optional<UserPrincipal> currentUser() {
        Authentication authentication = SecurityContextHolder.getContext().getAuthentication();
        if (authentication == null || !authentication.isAuthenticated()) {
            return Optional.empty();
        }
        Object principal = authentication.getPrincipal();
        if (principal instanceof UserPrincipal userPrincipal) {
            return Optional.of(userPrincipal);
        }
        return Optional.empty();
    }

    public static Optional<UUID> currentUserId() {
        return currentUser().map(UserPrincipal::userId);
    }

    public static Optional<String> currentAuditor() {
        return currentUserId().map(UUID::toString);
    }
}
