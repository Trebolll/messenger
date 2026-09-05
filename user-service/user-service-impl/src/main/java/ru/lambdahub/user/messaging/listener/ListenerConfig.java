package ru.lambdahub.user.messaging.listener;

import java.util.function.Consumer;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.messaging.Message;
import ru.lambdahub.common.events.kafka.DtoKafkaUserCreated;
import ru.lambdahub.user.messaging.handler.UserCreatedHandler;

@Configuration
public class ListenerConfig {

    @Bean
    public Consumer<Message<DtoKafkaUserCreated>> userCreatedInput(UserCreatedHandler userCreatedHandler) {
        return userCreatedHandler::handleUserCreated;
    }
}
