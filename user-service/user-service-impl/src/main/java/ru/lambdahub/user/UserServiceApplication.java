package ru.lambdahub.user;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.persistence.autoconfigure.EntityScan;
import org.springframework.data.jpa.repository.config.EnableJpaRepositories;

@SpringBootApplication(scanBasePackages = {"ru.lambdahub.user", "ru.lambdahub.common.web"})
@EntityScan(basePackages = "ru.lambdahub.user.db")
@EnableJpaRepositories(basePackages = "ru.lambdahub.user.db")
public class UserServiceApplication {
    public static void main(String[] args) {
        SpringApplication.run(UserServiceApplication.class, args);
    }
}
