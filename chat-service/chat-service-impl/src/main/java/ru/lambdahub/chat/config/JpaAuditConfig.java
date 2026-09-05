package ru.lambdahub.chat.config;

import java.util.Optional;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.data.domain.AuditorAware;
import org.springframework.data.jpa.repository.config.EnableJpaAuditing;
import ru.lambdahub.common.security.SecurityUtils;

@Configuration
@EnableJpaAuditing
public class JpaAuditConfig {

    @Bean
    public AuditorAware<String> auditorAware() {
        return () -> SecurityUtils.currentAuditor().or(() -> Optional.of("system"));
    }
}
