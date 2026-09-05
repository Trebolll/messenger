package ru.lambdahub.common.kafka;

import java.io.IOException;
import java.util.List;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.EnvironmentPostProcessor;
import org.springframework.boot.env.YamlPropertySourceLoader;
import org.springframework.core.Ordered;
import org.springframework.core.env.ConfigurableEnvironment;
import org.springframework.core.env.PropertySource;
import org.springframework.core.io.ClassPathResource;

/**
 * Loads starter defaults and optional {@code kafka-stream/*.yml} from the service classpath
 * so {@link EnvironmentConfig} can see bindings during environment post-processing.
 */
public class KafkaDefaultsEnvironmentPostProcessor implements EnvironmentPostProcessor, Ordered {

    private static final String DEFAULTS = "kafka-defaults.yml";
    private static final String BINDINGS = "kafka-stream/kafka-bindings.yml";
    private static final String KAFKA = "kafka-stream/kafka.yml";

    @Override
    public void postProcessEnvironment(ConfigurableEnvironment environment, SpringApplication application) {
        loadLast(environment, DEFAULTS, "lambdahubKafkaDefaults");
        loadFirst(environment, KAFKA, "lambdahubKafkaStream");
        loadFirst(environment, BINDINGS, "lambdahubKafkaBindings");
    }

    private void loadLast(ConfigurableEnvironment environment, String resourcePath, String sourceName) {
        if (environment.getPropertySources().contains(sourceName)) {
            return;
        }
        ClassPathResource resource = new ClassPathResource(resourcePath);
        if (!resource.exists()) {
            return;
        }
        try {
            List<PropertySource<?>> sources = new YamlPropertySourceLoader().load(sourceName, resource);
            for (int i = sources.size() - 1; i >= 0; i--) {
                environment.getPropertySources().addLast(sources.get(i));
            }
        } catch (IOException e) {
            throw new IllegalStateException("Failed to load " + resourcePath, e);
        }
    }

    private void loadFirst(ConfigurableEnvironment environment, String resourcePath, String sourceName) {
        if (environment.getPropertySources().contains(sourceName)) {
            return;
        }
        ClassPathResource resource = new ClassPathResource(resourcePath);
        if (!resource.exists()) {
            return;
        }
        try {
            List<PropertySource<?>> sources = new YamlPropertySourceLoader().load(sourceName, resource);
            for (PropertySource<?> source : sources) {
                environment.getPropertySources().addFirst(source);
            }
        } catch (IOException e) {
            throw new IllegalStateException("Failed to load " + resourcePath, e);
        }
    }

    @Override
    public int getOrder() {
        return Ordered.LOWEST_PRECEDENCE - 10;
    }
}
