package ru.lambdahub.realtime.messaging.listener;

import java.util.function.Consumer;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.messaging.Message;
import ru.lambdahub.common.events.kafka.DtoKafkaAttachment;
import ru.lambdahub.common.events.kafka.DtoKafkaChatCreated;
import ru.lambdahub.common.events.kafka.DtoKafkaChatMember;
import ru.lambdahub.common.events.kafka.DtoKafkaMessage;
import ru.lambdahub.common.events.kafka.DtoKafkaMessagesRead;
import ru.lambdahub.common.events.kafka.DtoKafkaUserProfile;
import ru.lambdahub.common.events.kafka.DtoKafkaUserStatus;
import ru.lambdahub.realtime.messaging.handler.DomainEventFanoutHandler;

@Configuration
public class ListenerConfig {

    @Bean
    public Consumer<Message<DtoKafkaMessage>> messageCreatedInput(DomainEventFanoutHandler handler) {
        return msg -> handler.onMessageCreated(msg.getPayload());
    }

    @Bean
    public Consumer<Message<DtoKafkaMessage>> messageEditedInput(DomainEventFanoutHandler handler) {
        return msg -> handler.onMessageEdited(msg.getPayload());
    }

    @Bean
    public Consumer<Message<DtoKafkaMessage>> messageDeletedInput(DomainEventFanoutHandler handler) {
        return msg -> handler.onMessageDeleted(msg.getPayload());
    }

    @Bean
    public Consumer<Message<DtoKafkaMessagesRead>> messagesReadInput(DomainEventFanoutHandler handler) {
        return msg -> handler.onMessagesRead(msg.getPayload());
    }

    @Bean
    public Consumer<Message<DtoKafkaChatMember>> chatMemberAddedInput(DomainEventFanoutHandler handler) {
        return msg -> handler.onChatMemberAdded(msg.getPayload());
    }

    @Bean
    public Consumer<Message<DtoKafkaChatMember>> chatMemberRemovedInput(DomainEventFanoutHandler handler) {
        return msg -> handler.onChatMemberRemoved(msg.getPayload());
    }

    @Bean
    public Consumer<Message<DtoKafkaChatCreated>> chatCreatedInput(DomainEventFanoutHandler handler) {
        return msg -> handler.onChatCreated(msg.getPayload());
    }

    @Bean
    public Consumer<Message<DtoKafkaUserProfile>> userProfileUpdatedInput(DomainEventFanoutHandler handler) {
        return msg -> handler.onUserProfileUpdated(msg.getPayload());
    }

    @Bean
    public Consumer<Message<DtoKafkaUserStatus>> userStatusChangedInput(DomainEventFanoutHandler handler) {
        return msg -> handler.onUserStatusChanged(msg.getPayload());
    }

    @Bean
    public Consumer<Message<DtoKafkaAttachment>> attachmentCreatedInput(DomainEventFanoutHandler handler) {
        return msg -> handler.onAttachmentCreated(msg.getPayload());
    }
}
