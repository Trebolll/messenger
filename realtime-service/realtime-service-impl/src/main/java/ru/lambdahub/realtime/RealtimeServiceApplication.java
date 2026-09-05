package ru.lambdahub.realtime;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.openfeign.EnableFeignClients;

@SpringBootApplication(scanBasePackages = "ru.lambdahub.realtime")
@EnableFeignClients(basePackages = "ru.lambdahub.realtime.feign")
public class RealtimeServiceApplication {
    public static void main(String[] args) {
        SpringApplication.run(RealtimeServiceApplication.class, args);
    }
}
