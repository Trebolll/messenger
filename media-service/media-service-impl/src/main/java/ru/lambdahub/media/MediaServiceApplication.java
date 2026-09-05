package ru.lambdahub.media;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.persistence.autoconfigure.EntityScan;
import org.springframework.data.jpa.repository.config.EnableJpaRepositories;

@SpringBootApplication(scanBasePackages = {"ru.lambdahub.media", "ru.lambdahub.common.web"})
@EntityScan(basePackages = "ru.lambdahub.media.db")
@EnableJpaRepositories(basePackages = "ru.lambdahub.media.db")
public class MediaServiceApplication {
    public static void main(String[] args) {
        SpringApplication.run(MediaServiceApplication.class, args);
    }
}
