package ru.lambdahub.auth.feign;

import org.springframework.cloud.openfeign.FeignClient;
import ru.lambdahub.user.api.InternalUsersApi;

@FeignClient(
        name = "user-service",
        contextId = "internalUsers",
        url = "${spring.cloud.openfeign.client.config.user-service.url}"
)
public interface InternalUsersFeignClient extends InternalUsersApi {
}
