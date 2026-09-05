package ru.lambdahub.common.kafka;

import java.util.AbstractMap;
import java.util.Arrays;
import java.util.Collection;
import java.util.LinkedHashMap;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.Executor;
import java.util.function.Supplier;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.config.BeanDefinition;
import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.beans.factory.support.BeanDefinitionRegistry;
import org.springframework.beans.factory.support.BeanNameGenerator;
import org.springframework.beans.factory.support.GenericBeanDefinition;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.EnvironmentPostProcessor;
import org.springframework.cloud.stream.function.StreamBridge;
import org.springframework.context.EnvironmentAware;
import org.springframework.context.annotation.ImportBeanDefinitionRegistrar;
import org.springframework.core.Ordered;
import org.springframework.core.env.ConfigurableEnvironment;
import org.springframework.core.env.Environment;
import org.springframework.core.env.MapPropertySource;
import org.springframework.core.type.AnnotationMetadata;

/**
 * Reads {@code spring.cloud.stream.bindings.*-out-*} / {@code *-in-*} from kafka-bindings,
 * registers {@code spring.cloud.function.definition} for inputs, and creates {@link OutputBridge}
 * beans for pure outputs.
 */
public class EnvironmentConfig
        implements EnvironmentPostProcessor, EnvironmentAware, ImportBeanDefinitionRegistrar, Ordered {

    private static final Logger log = LoggerFactory.getLogger(EnvironmentConfig.class);

    private static final Pattern INPUT_BINDING_PATTERN = Pattern.compile("(.+)-in-\\d+");
    private static final Pattern OUTPUT_BINDING_PATTERN = Pattern.compile("(.+)-out-\\d+");
    private static final Pattern BINDING_PROPERTY_PATTERN = Pattern.compile(
            "spring\\.cloud\\.stream\\.bindings\\.(.+)\\.destination");

    private ConfigurableEnvironment environment;

    @Override
    public void postProcessEnvironment(ConfigurableEnvironment environment, SpringApplication application) {
        Set<String> inputDefinitions = createFunctionalBindings(environment).stream()
                .map(INPUT_BINDING_PATTERN::matcher)
                .filter(Matcher::matches)
                .map(matcher -> matcher.group(1))
                .collect(Collectors.toCollection(LinkedHashSet::new));

        String definitions = String.join(";", inputDefinitions);
        log.info("Spring cloud input definitions registered: {}", definitions);

        environment.getPropertySources().addFirst(new MapPropertySource(
                "Kafka functional bindings",
                Map.of("spring.cloud.function.definition", definitions)));
    }

    @Override
    public void setEnvironment(Environment environment) {
        if (!(environment instanceof ConfigurableEnvironment configurableEnvironment)) {
            throw new IllegalStateException("ConfigurableEnvironment is required for Kafka output bindings");
        }
        this.environment = configurableEnvironment;
    }

    @Override
    public void registerBeanDefinitions(
            AnnotationMetadata importingClassMetadata,
            BeanDefinitionRegistry registry,
            BeanNameGenerator importBeanNameGenerator) {
        if (!(registry instanceof ConfigurableBeanFactory factory)) {
            return;
        }

        Map<String, List<String>> outputBindings = createOutputBindings(environment);
        log.info("Spring cloud output definitions registered: {}", outputBindings);

        outputBindings.entrySet().stream()
                .filter(entry -> !registry.isBeanNameInUse(entry.getKey()))
                .map(entry -> new AbstractMap.SimpleEntry<>(entry.getKey(), createOutputBridge(entry, factory)))
                .map(this::createBeanDefinition)
                .forEach(entry -> registry.registerBeanDefinition(entry.getKey(), entry.getValue()));
    }

    private Supplier<OutputBridge> createOutputBridge(
            Map.Entry<String, List<String>> binding,
            ConfigurableBeanFactory factory) {
        return () -> {
            StreamBridge streamBridge = factory.getBean(StreamBridge.class);
            Executor executor = factory.getBean(LambdahubKafkaAutoConfiguration.OUTPUT_EXECUTOR_BEAN_NAME, Executor.class);
            return new OutputBridge(streamBridge, executor) {
                @Override
                public List<String> channels() {
                    return binding.getValue();
                }
            };
        };
    }

    private Map.Entry<String, BeanDefinition> createBeanDefinition(
            Map.Entry<String, Supplier<OutputBridge>> binding) {
        GenericBeanDefinition beanDefinition = new GenericBeanDefinition();
        beanDefinition.setBeanClass(OutputBridge.class);
        beanDefinition.setInstanceSupplier(binding.getValue());
        return new AbstractMap.SimpleEntry<>(binding.getKey(), beanDefinition);
    }

    private Map<String, List<String>> createOutputBindings(ConfigurableEnvironment environment) {
        Collection<String> functionalBindings = createFunctionalBindings(environment);
        Set<String> inputBindings = functionalBindings.stream()
                .map(INPUT_BINDING_PATTERN::matcher)
                .filter(Matcher::matches)
                .map(matcher -> matcher.group(1))
                .collect(Collectors.toSet());

        return functionalBindings.stream()
                .map(binding -> new AbstractMap.SimpleEntry<>(OUTPUT_BINDING_PATTERN.matcher(binding), binding))
                .filter(entry -> entry.getKey().matches())
                .filter(entry -> !inputBindings.contains(entry.getKey().group(1)))
                .collect(Collectors.groupingBy(
                        entry -> entry.getKey().group(1),
                        LinkedHashMap::new,
                        Collectors.mapping(AbstractMap.SimpleEntry::getValue, Collectors.toList())));
    }

    private Collection<String> createFunctionalBindings(ConfigurableEnvironment environment) {
        return environment.getPropertySources().stream()
                .filter(MapPropertySource.class::isInstance)
                .map(MapPropertySource.class::cast)
                .map(MapPropertySource::getPropertyNames)
                .flatMap(Arrays::stream)
                .map(BINDING_PROPERTY_PATTERN::matcher)
                .filter(Matcher::matches)
                .map(matcher -> matcher.group(1))
                .distinct()
                .toList();
    }

    @Override
    public int getOrder() {
        return Ordered.LOWEST_PRECEDENCE;
    }
}
