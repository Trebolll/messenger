package ru.lambdahub.common.redis;

import org.springframework.boot.autoconfigure.AutoConfiguration;
import org.springframework.boot.autoconfigure.condition.ConditionalOnClass;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.data.redis.core.StringRedisTemplate;

/**
 * Connection comes from Boot Redis auto-config + redis-defaults.yml.
 * Award redis starter builds Spring Cache on top of the same spring.data.redis settings;
 * here we only share connection defaults (OTP / presence / pub-sub use StringRedisTemplate).
 */
@AutoConfiguration
@ConditionalOnClass(StringRedisTemplate.class)
@ConditionalOnProperty(prefix = "lambdahub.redis", name = "enabled", havingValue = "true", matchIfMissing = true)
@EnableConfigurationProperties(LambdahubRedisProperties.class)
public class LambdahubRedisAutoConfiguration {
}
