package ru.lambdahub.common.kafka;

import java.util.concurrent.Executor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.AutoConfiguration;
import org.springframework.boot.autoconfigure.condition.ConditionalOnClass;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.cloud.stream.function.StreamBridge;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Import;
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor;

@AutoConfiguration
@ConditionalOnClass(StreamBridge.class)
@ConditionalOnProperty(prefix = "lambdahub.kafka", name = "enabled", havingValue = "true", matchIfMissing = true)
@EnableConfigurationProperties({LambdahubKafkaProperties.class, KafkaOutputExecutorProperties.class})
@Import(EnvironmentConfig.class)
public class LambdahubKafkaAutoConfiguration {

    public static final String OUTPUT_EXECUTOR_BEAN_NAME = "lambdahubKafkaOutputExecutor";

    @Bean(name = OUTPUT_EXECUTOR_BEAN_NAME)
    @ConditionalOnMissingBean(name = OUTPUT_EXECUTOR_BEAN_NAME)
    public ThreadPoolTaskExecutor lambdahubKafkaOutputExecutor(
            KafkaOutputExecutorProperties properties,
            @Value("${spring.application.name:app}") String applicationName) {
        ThreadPoolTaskExecutor executor = new ThreadPoolTaskExecutor();
        executor.setCorePoolSize(properties.getCorePoolSize());
        executor.setMaxPoolSize(Math.max(properties.getCorePoolSize(), properties.getMaxPoolSize()));
        executor.setQueueCapacity(properties.getQueueCapacity());
        executor.setThreadNamePrefix(applicationName + "-kafka-output-");
        executor.setWaitForTasksToCompleteOnShutdown(true);
        executor.setAwaitTerminationSeconds(properties.getAwaitTerminationSeconds());
        return executor;
    }
}
