package ru.lambdahub.common.kafka;

import org.springframework.boot.context.properties.ConfigurationProperties;

@ConfigurationProperties(prefix = "lambdahub.kafka")
public class LambdahubKafkaProperties {

    /** Feature toggle for this starter's auto-config. */
    private boolean enabled = true;

    public boolean isEnabled() {
        return enabled;
    }

    public void setEnabled(boolean enabled) {
        this.enabled = enabled;
    }
}
