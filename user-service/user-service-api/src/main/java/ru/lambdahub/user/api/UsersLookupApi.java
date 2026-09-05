package ru.lambdahub.user.api;

import java.util.List;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;

/**
 * Feign-friendly user lookup (no {@code @AuthenticationPrincipal}).
 */
public interface UsersLookupApi {

    String BASE_API = "/api/users";

    @GetMapping(BASE_API + "/search")
    List<ProfileResponse> search(@RequestParam("q") String q);
}
