package ru.lambdahub.chat.feign;

import org.springframework.cloud.openfeign.FeignClient;
import ru.lambdahub.user.api.UsersLookupApi;

@FeignClient(
        name = "user-service",
        contextId = "usersLookup",
        url = "${spring.cloud.openfeign.client.config.user-service.url}"
)
public interface UsersLookupFeignClient extends UsersLookupApi {
}
