package ru.lambdahub.auth;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.persistence.autoconfigure.EntityScan;
import org.springframework.cloud.openfeign.EnableFeignClients;
import org.springframework.data.jpa.repository.config.EnableJpaRepositories;

@SpringBootApplication(scanBasePackages = {"ru.lambdahub.auth", "ru.lambdahub.common.web"})
@EntityScan(basePackages = "ru.lambdahub.auth.db")
@EnableJpaRepositories(basePackages = "ru.lambdahub.auth.db")
@EnableFeignClients(basePackages = "ru.lambdahub.auth.feign")
public class AuthServiceApplication {

  public static void main(String[] args) {
    SpringApplication.run(AuthServiceApplication.class, args);
  }
}
