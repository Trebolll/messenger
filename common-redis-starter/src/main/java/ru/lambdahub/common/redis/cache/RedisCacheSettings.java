package ru.lambdahub.common.redis.cache;

import java.time.Duration;
import lombok.Data;

/**
 * Per-cache Redis settings (same shape as spring.cache.redis / award CacheProperties.Redis).
 */
@Data
public class RedisCacheSettings {

    private Duration timeToLive;
    private boolean cacheNullValues = true;
    private boolean useKeyPrefix = true;
    private String keyPrefix;
    private boolean enableStatistics;
}
