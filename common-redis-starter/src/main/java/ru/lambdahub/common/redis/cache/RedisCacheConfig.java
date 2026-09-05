package ru.lambdahub.common.redis.cache;

import tools.jackson.databind.ObjectMapper;
import tools.jackson.databind.json.JsonMapper;
import java.time.Duration;
import java.util.LinkedHashMap;
import java.util.Map;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.boot.autoconfigure.AutoConfiguration;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.core.env.Environment;
import org.springframework.data.redis.cache.BatchStrategies;
import org.springframework.data.redis.cache.RedisCacheConfiguration;
import org.springframework.data.redis.cache.RedisCacheManager;
import org.springframework.data.redis.cache.RedisCacheWriter;
import org.springframework.data.redis.connection.RedisConnectionFactory;
import org.springframework.data.redis.serializer.GenericJacksonJsonRedisSerializer;
import org.springframework.data.redis.serializer.RedisSerializationContext;
import org.springframework.data.redis.serializer.RedisSerializer;

@AutoConfiguration
@EnableConfigurationProperties(CacheConfigProperties.class)
@ConditionalOnProperty(name = "cache-config.enabled", havingValue = "true")
public class RedisCacheConfig {

    @Bean("redisCacheValueSerializer")
    @ConditionalOnMissingBean(name = "redisCacheValueSerializer")
    public RedisSerializer<Object> redisCacheValueSerializer() {
        return GenericJacksonJsonRedisSerializer.builder()
                .enableUnsafeDefaultTyping()
                .enableSpringCacheNullValueSupport()
                .build();
    }

    @Bean("redisCacheKeyConverter")
    @ConditionalOnMissingBean(name = "redisCacheKeyConverter")
    public CanonicalJsonCacheKeyConverter redisCacheKeyConverter(ObjectProvider<ObjectMapper> objectMapperProvider) {
        ObjectMapper objectMapper = objectMapperProvider.getIfAvailable();
        if (objectMapper == null) {
            objectMapper = JsonMapper.builder().findAndAddModules().build();
        }
        return new CanonicalJsonCacheKeyConverter(objectMapper);
    }

    @Bean("redisCacheConfigurationMap")
    @ConditionalOnMissingBean(name = "redisCacheConfigurationMap")
    public Map<String, RedisCacheConfiguration> redisCacheConfigurationMap(
            CacheConfigProperties properties,
            Environment environment,
            @Qualifier("redisCacheKeyConverter") CanonicalJsonCacheKeyConverter keyConverter,
            @Qualifier("redisCacheValueSerializer") RedisSerializer<Object> serializer) {
        RedisCacheConfiguration baseConfiguration = RedisCacheConfiguration.defaultCacheConfig()
                .serializeValuesWith(RedisSerializationContext.SerializationPair.fromSerializer(serializer));
        baseConfiguration.addCacheKeyConverter(keyConverter);
        Map<String, RedisCacheConfiguration> configurations = new LinkedHashMap<>();

        properties.getCaches().forEach((alias, cacheProperties) -> configurations.put(
                properties.getCacheName(alias),
                createCacheConfiguration(baseConfiguration, properties.getDefaults(), cacheProperties, alias, environment)));

        return configurations;
    }

    @Bean("redisCacheManager")
    @ConditionalOnMissingBean(name = "redisCacheManager")
    public RedisCacheManager redisCacheManager(
            RedisConnectionFactory connectionFactory,
            CacheConfigProperties properties,
            @Qualifier("redisCacheKeyConverter") CanonicalJsonCacheKeyConverter keyConverter,
            @Qualifier("redisCacheConfigurationMap") Map<String, RedisCacheConfiguration> cacheConfigurations,
            @Qualifier("redisCacheValueSerializer") RedisSerializer<Object> serializer) {
        RedisCacheConfiguration defaultConfiguration = createDefaultCacheConfiguration(
                RedisCacheConfiguration.defaultCacheConfig()
                        .serializeValuesWith(RedisSerializationContext.SerializationPair.fromSerializer(serializer)),
                properties.getDefaults());
        defaultConfiguration.addCacheKeyConverter(keyConverter);

        RedisCacheWriter cacheWriter = RedisCacheWriter.nonLockingRedisCacheWriter(
                connectionFactory, BatchStrategies.scan(properties.getClearScanBatchSize()));
        var builder = RedisCacheManager.builder(cacheWriter)
                .cacheDefaults(defaultConfiguration)
                .withInitialCacheConfigurations(cacheConfigurations);

        if (properties.getDefaults().isEnableStatistics()
                || properties.getCaches().values().stream().anyMatch(RedisCacheSettings::isEnableStatistics)) {
            builder.enableStatistics();
        }

        return builder.build();
    }

    @Bean
    @ConditionalOnMissingBean
    public RedisCacheProviderService redisCacheProviderService(
            @Qualifier("redisCacheManager") org.springframework.cache.CacheManager redisCacheManager,
            CacheConfigProperties cacheConfigProperties) {
        return new RedisCacheProviderService(redisCacheManager, cacheConfigProperties);
    }

    private RedisCacheConfiguration createCacheConfiguration(
            RedisCacheConfiguration baseConfiguration,
            RedisCacheSettings defaults,
            RedisCacheSettings cacheProperties,
            String alias,
            Environment environment) {
        Duration timeToLive = cacheProperties.getTimeToLive() != null
                ? cacheProperties.getTimeToLive()
                : defaults.getTimeToLive();
        boolean cacheNullValues = isConfigured(environment, alias, "cache-null-values")
                ? cacheProperties.isCacheNullValues()
                : defaults.isCacheNullValues();
        boolean useKeyPrefix = isConfigured(environment, alias, "use-key-prefix")
                ? cacheProperties.isUseKeyPrefix()
                : defaults.isUseKeyPrefix();
        String keyPrefix = cacheProperties.getKeyPrefix() != null
                ? cacheProperties.getKeyPrefix()
                : defaults.getKeyPrefix();

        return applyProperties(baseConfiguration, timeToLive, cacheNullValues, useKeyPrefix, keyPrefix);
    }

    private RedisCacheConfiguration createDefaultCacheConfiguration(
            RedisCacheConfiguration baseConfiguration,
            RedisCacheSettings defaults) {
        return applyProperties(
                baseConfiguration,
                defaults.getTimeToLive(),
                defaults.isCacheNullValues(),
                defaults.isUseKeyPrefix(),
                defaults.getKeyPrefix());
    }

    private RedisCacheConfiguration applyProperties(
            RedisCacheConfiguration baseConfiguration,
            Duration timeToLive,
            boolean cacheNullValues,
            boolean useKeyPrefix,
            String keyPrefix) {
        RedisCacheConfiguration configuration = baseConfiguration;

        if (timeToLive != null) {
            configuration = configuration.entryTtl(timeToLive);
        }
        if (!cacheNullValues) {
            configuration = configuration.disableCachingNullValues();
        }
        if (keyPrefix != null) {
            configuration = configuration.prefixCacheNameWith(keyPrefix);
        }
        if (!useKeyPrefix) {
            configuration = configuration.disableKeyPrefix();
        }

        return configuration;
    }

    private boolean isConfigured(Environment environment, String alias, String property) {
        return environment.containsProperty("cache-config.caches." + alias + "." + property);
    }
}
