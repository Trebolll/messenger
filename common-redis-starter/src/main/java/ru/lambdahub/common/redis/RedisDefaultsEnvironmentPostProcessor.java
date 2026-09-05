package ru.lambdahub.common.redis;

import java.io.IOException;
import java.util.List;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.EnvironmentPostProcessor;
import org.springframework.boot.env.YamlPropertySourceLoader;
import org.springframework.core.env.ConfigurableEnvironment;
import org.springframework.core.env.PropertySource;
import org.springframework.core.io.ClassPathResource;

/**
 * Loads classpath:redis-defaults.yml with lowest precedence.
 * Connection shape mirrors award: spring.data.redis (+ optional cluster.nodes).
 */
public class RedisDefaultsEnvironmentPostProcessor implements EnvironmentPostProcessor {

    private static final String RESOURCE = "redis-defaults.yml";
    private static final String SOURCE_NAME = "lambdahubRedisDefaults";

    @Override
    public void postProcessEnvironment(ConfigurableEnvironment environment, SpringApplication application) {
        if (environment.getPropertySources().contains(SOURCE_NAME)) {
            return;
        }
        ClassPathResource resource = new ClassPathResource(RESOURCE);
        if (!resource.exists()) {
            return;
        }
        try {
            List<PropertySource<?>> sources = new YamlPropertySourceLoader().load(SOURCE_NAME, resource);
            for (int i = sources.size() - 1; i >= 0; i--) {
                environment.getPropertySources().addLast(sources.get(i));
            }
        } catch (IOException e) {
            throw new IllegalStateException("Failed to load " + RESOURCE, e);
        }
    }
}
