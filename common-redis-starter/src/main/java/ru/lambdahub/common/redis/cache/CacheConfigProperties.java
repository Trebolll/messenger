package ru.lambdahub.common.redis.cache;

import jakarta.validation.Valid;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Positive;
import java.util.LinkedHashMap;
import java.util.Map;
import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.validation.annotation.Validated;

@Validated
@ConfigurationProperties(prefix = "cache-config")
@Data
public class CacheConfigProperties {

    private boolean enabled;
    private boolean inspectionEnabled;
    private String cacheNamePrefix = "";
    @Positive
    private int clearScanBatchSize = 1000;
    private RedisCacheSettings defaults = new RedisCacheSettings();

    @Valid
    private Map<@NotBlank String, RedisCacheSettings> caches = new LinkedHashMap<>();

    public String getCacheName(String alias) {
        if (!caches.containsKey(alias)) {
            throw new IllegalArgumentException("Cache is not configured: " + alias);
        }
        return cacheNamePrefix + alias;
    }
}
