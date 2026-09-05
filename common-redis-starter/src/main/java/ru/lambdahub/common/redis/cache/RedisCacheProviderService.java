package ru.lambdahub.common.redis.cache;

import java.util.Optional;
import lombok.RequiredArgsConstructor;
import org.springframework.cache.Cache;
import org.springframework.cache.CacheManager;

@RequiredArgsConstructor
public class RedisCacheProviderService {

    private final CacheManager redisCacheManager;
    private final CacheConfigProperties cacheConfigProperties;

    public Optional<Cache> getCache(String alias) {
        return Optional.ofNullable(redisCacheManager.getCache(cacheConfigProperties.getCacheName(alias)));
    }
}
