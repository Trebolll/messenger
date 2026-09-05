package ru.lambdahub.common.security;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.util.List;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpHeaders;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.web.filter.OncePerRequestFilter;

/**
 * Builds {@link UserPrincipal} from Bearer JWT or trusted gateway headers {@code X-User-Id}/{@code X-Username}.
 */
@RequiredArgsConstructor
public class JwtOrHeaderAuthenticationFilter extends OncePerRequestFilter {

    private final JwtService jwtService;

    @Override
    protected void doFilterInternal(HttpServletRequest request,
                                    HttpServletResponse response,
                                    FilterChain filterChain) throws ServletException, IOException {
        if (SecurityContextHolder.getContext().getAuthentication() == null) {
            UserPrincipal principal = resolvePrincipal(request);
            if (principal != null) {
                var authentication = new UsernamePasswordAuthenticationToken(
                        principal, null, List.of());
                SecurityContextHolder.getContext().setAuthentication(authentication);
            }
        }
        filterChain.doFilter(request, response);
    }

    private UserPrincipal resolvePrincipal(HttpServletRequest request) {
        String bearer = extractBearer(request);
        if (bearer != null) {
            try {
                UUID userId = jwtService.requireUserId(bearer);
                var claims = jwtService.parse(bearer);
                Object username = claims.get("username");
                return new UserPrincipal(userId, username == null ? "" : username.toString());
            } catch (Exception ignored) {
                return null;
            }
        }

        String userIdHeader = request.getHeader(AuthHeaders.USER_ID);
        if (userIdHeader != null && !userIdHeader.isBlank()) {
            try {
                UUID userId = UUID.fromString(userIdHeader.trim());
                String username = request.getHeader(AuthHeaders.USERNAME);
                return new UserPrincipal(userId, username == null ? "" : username);
            } catch (IllegalArgumentException ignored) {
                return null;
            }
        }
        return null;
    }

    private String extractBearer(HttpServletRequest request) {
        String auth = request.getHeader(HttpHeaders.AUTHORIZATION);
        if (auth != null && auth.startsWith("Bearer ")) {
            return auth.substring(7).trim();
        }
        return null;
    }
}
