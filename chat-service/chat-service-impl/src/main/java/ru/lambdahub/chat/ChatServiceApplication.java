package ru.lambdahub.chat;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.persistence.autoconfigure.EntityScan;
import org.springframework.cloud.openfeign.EnableFeignClients;
import org.springframework.data.jpa.repository.config.EnableJpaRepositories;

@SpringBootApplication(scanBasePackages = {"ru.lambdahub.chat", "ru.lambdahub.common.web"})
@EntityScan(basePackages = "ru.lambdahub.chat.db")
@EnableJpaRepositories(basePackages = "ru.lambdahub.chat.db")
@EnableFeignClients(basePackages = "ru.lambdahub.chat.feign")
public class ChatServiceApplication {
    public static void main(String[] args) {
        SpringApplication.run(ChatServiceApplication.class, args);
    }
}
