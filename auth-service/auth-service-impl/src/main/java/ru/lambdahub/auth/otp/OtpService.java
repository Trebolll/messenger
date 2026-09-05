package ru.lambdahub.auth.otp;

import java.security.SecureRandom;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.cache.Cache;
import org.springframework.stereotype.Service;
import ru.lambdahub.common.redis.cache.RedisCacheProviderService;

@Service
@RequiredArgsConstructor
public class OtpService {

    private static final String CACHE_OTP = "otp";
    private static final String CACHE_COOLDOWN = "otp-cooldown";
    private static final String CACHE_RESET = "reset-token";

    private final RedisCacheProviderService redisCacheProviderService;
    @Value("${app.max-otp-attempts}")
    private final int maxAttempts;
    @Value("${app.dev-otp}")
    private final boolean devOtp;
    private final SecureRandom random = new SecureRandom();

    public String send(String purpose, String login) {
        String id = purpose + ":" + login;
        if (cooldownCache().get(id) != null) {
            throw new IllegalStateException("Please wait before requesting another code");
        }
        String code = String.format("%06d", random.nextInt(1_000_000));
        otpCache().put(id, new OtpEntry(code, 0));
        cooldownCache().put(id, Boolean.TRUE);
        if (devOtp) {
            return code;
        }
        return null;
    }

    public void verify(String purpose, String login, String code) {
        String id = purpose + ":" + login;
        OtpEntry entry = otpCache().get(id, OtpEntry.class);
        if (entry == null) {
            throw new IllegalArgumentException("Code expired or not found");
        }
        if (entry.attempts() >= maxAttempts) {
            throw new IllegalStateException("Too many attempts");
        }
        if (!entry.code().equals(code)) {
            otpCache().put(id, new OtpEntry(entry.code(), entry.attempts() + 1));
            throw new IllegalArgumentException("Invalid code");
        }
        otpCache().evict(id);
    }

    public String createResetToken(String login) {
        String token = UUID.randomUUID().toString();
        resetTokenCache().put(token, login);
        return token;
    }

    public String consumeResetToken(String resetToken) {
        String login = resetTokenCache().get(resetToken, String.class);
        if (login == null) {
            throw new IllegalArgumentException("Invalid or expired reset token");
        }
        resetTokenCache().evict(resetToken);
        return login;
    }

    private Cache otpCache() {
        return redisCacheProviderService.getCache(CACHE_OTP)
                .orElseThrow(() -> new IllegalStateException("Cache not configured: " + CACHE_OTP));
    }

    private Cache cooldownCache() {
        return redisCacheProviderService.getCache(CACHE_COOLDOWN)
                .orElseThrow(() -> new IllegalStateException("Cache not configured: " + CACHE_COOLDOWN));
    }

    private Cache resetTokenCache() {
        return redisCacheProviderService.getCache(CACHE_RESET)
                .orElseThrow(() -> new IllegalStateException("Cache not configured: " + CACHE_RESET));
    }

    public record OtpEntry(String code, int attempts) {}
}
