package ru.lambdahub.common.security;

import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.security.Keys;

import javax.crypto.SecretKey;
import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.Date;
import java.util.Map;
import java.util.UUID;

public final class JwtService {

    private final SecretKey secretKey;
    private final long ttlSeconds;

    public JwtService(String secret, long ttlSeconds) {
        if (secret == null || secret.length() < 32) {
            throw new IllegalArgumentException("JWT secret must be at least 32 characters");
        }
        this.secretKey = Keys.hmacShaKeyFor(secret.getBytes(StandardCharsets.UTF_8));
        this.ttlSeconds = ttlSeconds;
    }

    public String issueSessionToken(UUID userId, String username) {
        Instant now = Instant.now();
        return Jwts.builder()
                .subject(userId.toString())
                .claim("user_id", userId.toString())
                .claim("username", username)
                .issuedAt(Date.from(now))
                .expiration(Date.from(now.plusSeconds(ttlSeconds)))
                .signWith(secretKey)
                .compact();
    }

    public String issueConfirmToken(String login, long confirmTtlSeconds) {
        Instant now = Instant.now();
        return Jwts.builder()
                .id(UUID.randomUUID().toString())
                .claim("confirm_login", login)
                .issuedAt(Date.from(now))
                .expiration(Date.from(now.plusSeconds(confirmTtlSeconds)))
                .signWith(secretKey)
                .compact();
    }

    public Claims parse(String token) {
        return Jwts.parser()
                .verifyWith(secretKey)
                .build()
                .parseSignedClaims(token)
                .getPayload();
    }

    public UUID requireUserId(String token) {
        Claims claims = parse(token);
        Object raw = claims.get("user_id");
        if (raw == null) {
            raw = claims.getSubject();
        }
        if (raw == null) {
            throw new IllegalArgumentException("Missing user_id claim");
        }
        return UUID.fromString(raw.toString());
    }

    public String requireConfirmLogin(String token) {
        Claims claims = parse(token);
        Object login = claims.get("confirm_login");
        if (login == null) {
            throw new IllegalArgumentException("Missing confirm_login claim");
        }
        return login.toString();
    }

    public Map<String, Object> asMap(Claims claims) {
        return Map.copyOf(claims);
    }
}
