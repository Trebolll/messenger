package ru.lambdahub.user.api;

import jakarta.validation.Valid;
import java.util.UUID;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import ru.lambdahub.common.security.UserPrincipal;

public interface UsersApi extends UsersLookupApi {

    String BASE_API = UsersLookupApi.BASE_API;

    @GetMapping(BASE_API + "/me")
    ProfileResponse me(@AuthenticationPrincipal UserPrincipal principal);

    @GetMapping(BASE_API + "/{id}")
    ProfileResponse byId(@PathVariable UUID id);

    @PutMapping(BASE_API + "/profile")
    ProfileResponse updateProfile(@AuthenticationPrincipal UserPrincipal principal,
                                  @Valid @RequestBody UpdateProfileRequest req);

    @PutMapping(BASE_API + "/status")
    ProfileResponse updateStatus(@AuthenticationPrincipal UserPrincipal principal,
                                 @Valid @RequestBody UpdateStatusRequest req);

    @PutMapping(BASE_API + "/avatar")
    ProfileResponse updateAvatar(@AuthenticationPrincipal UserPrincipal principal,
                                 @Valid @RequestBody UpdateAvatarRequest req);
}
