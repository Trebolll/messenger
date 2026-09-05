package ru.lambdahub.common.security;

import java.util.ArrayList;
import java.util.List;
import org.springframework.boot.context.properties.ConfigurationProperties;

@ConfigurationProperties(prefix = "lambdahub.security")
public class LambdahubSecurityProperties {

    private boolean enabled = true;
    private String jwtSecret = "changeme-jwt-secret-key-32chars-min";
    private long jwtTtlSeconds = 86400;
    private List<String> publicPathPrefixes = new ArrayList<>(List.of(
            "/api/auth/",
            "/actuator/health",
            "/api/internal/",
            "/api/ws"
    ));

    public boolean isEnabled() {
        return enabled;
    }

    public void setEnabled(boolean enabled) {
        this.enabled = enabled;
    }

    public String getJwtSecret() {
        return jwtSecret;
    }

    public void setJwtSecret(String jwtSecret) {
        this.jwtSecret = jwtSecret;
    }

    public long getJwtTtlSeconds() {
        return jwtTtlSeconds;
    }

    public void setJwtTtlSeconds(long jwtTtlSeconds) {
        this.jwtTtlSeconds = jwtTtlSeconds;
    }

    public List<String> getPublicPathPrefixes() {
        return publicPathPrefixes;
    }

    public void setPublicPathPrefixes(List<String> publicPathPrefixes) {
        this.publicPathPrefixes = publicPathPrefixes;
    }
}
