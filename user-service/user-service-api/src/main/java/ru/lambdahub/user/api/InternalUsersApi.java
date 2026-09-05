package ru.lambdahub.user.api;

import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestHeader;
import ru.lambdahub.common.events.kafka.DtoKafkaUserCreated;
import ru.lambdahub.common.security.AuthHeaders;

public interface InternalUsersApi {

    String BASE_API = "/api/internal/users";

    @PostMapping(BASE_API)
    void create(@RequestHeader(value = AuthHeaders.INTERNAL_KEY, required = false) String key,
                @RequestBody DtoKafkaUserCreated payload);
}
