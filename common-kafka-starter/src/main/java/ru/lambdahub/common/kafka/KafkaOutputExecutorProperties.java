package ru.lambdahub.common.kafka;

import org.springframework.boot.context.properties.ConfigurationProperties;

@ConfigurationProperties("spring.cloud.stream.output-bridge.executor")
public class KafkaOutputExecutorProperties {

    private int corePoolSize = 10;
    private int maxPoolSize = 10;
    private int queueCapacity = 1000;
    private int awaitTerminationSeconds = 30;

    public int getCorePoolSize() {
        return corePoolSize;
    }

    public void setCorePoolSize(int corePoolSize) {
        this.corePoolSize = corePoolSize;
    }

    public int getMaxPoolSize() {
        return maxPoolSize;
    }

    public void setMaxPoolSize(int maxPoolSize) {
        this.maxPoolSize = maxPoolSize;
    }

    public int getQueueCapacity() {
        return queueCapacity;
    }

    public void setQueueCapacity(int queueCapacity) {
        this.queueCapacity = queueCapacity;
    }

    public int getAwaitTerminationSeconds() {
        return awaitTerminationSeconds;
    }

    public void setAwaitTerminationSeconds(int awaitTerminationSeconds) {
        this.awaitTerminationSeconds = awaitTerminationSeconds;
    }
}
