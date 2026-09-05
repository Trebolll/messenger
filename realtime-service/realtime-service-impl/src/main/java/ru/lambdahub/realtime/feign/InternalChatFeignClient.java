package ru.lambdahub.realtime.feign;

import org.springframework.cloud.openfeign.FeignClient;
import ru.lambdahub.chat.api.InternalChatApi;

@FeignClient(
        name = "chat-service",
        contextId = "chatInternal",
        url = "${spring.cloud.openfeign.client.config.chat-service.url}"
)
public interface InternalChatFeignClient extends InternalChatApi {
}
