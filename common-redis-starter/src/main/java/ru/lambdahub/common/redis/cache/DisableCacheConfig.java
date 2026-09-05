package ru.lambdahub.common.redis.cache;

import org.springframework.boot.autoconfigure.AutoConfiguration;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.cache.CacheManager;
import org.springframework.cache.support.NoOpCacheManager;
import org.springframework.context.annotation.Bean;

@AutoConfiguration
@EnableConfigurationProperties(CacheConfigProperties.class)
@ConditionalOnProperty(name = "cache-config.enabled", havingValue = "false", matchIfMissing = true)
public class DisableCacheConfig {

    @Bean("redisCacheManager")
    @ConditionalOnMissingBean(name = "redisCacheManager")
    public CacheManager redisCacheManager() {
        return new NoOpCacheManager();
    }

    @Bean
    @ConditionalOnMissingBean
    public RedisCacheProviderService redisCacheProviderService(
            CacheManager redisCacheManager,
            CacheConfigProperties cacheConfigProperties) {
        return new RedisCacheProviderService(redisCacheManager, cacheConfigProperties);
    }
}
