package ru.lambdahub.auth.api;

import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.ResponseStatus;

public interface AuthApi {

    String BASE_API = "/api/auth";

    @PostMapping(BASE_API + "/send")
    SendCodeResponse send(@Valid @RequestBody SendCodeRequest req);

    @PostMapping(BASE_API + "/verify-code")
    VerifyCodeResponse verify(@Valid @RequestBody VerifyCodeRequest req);

    @PostMapping(BASE_API + "/register")
    @ResponseStatus(HttpStatus.CREATED)
    AuthResponse register(@Valid @RequestBody RegisterRequest req);

    @PostMapping(BASE_API + "/login")
    AuthResponse login(@Valid @RequestBody LoginRequest req);

    @PostMapping(BASE_API + "/reset/send")
    SendCodeResponse resetSend(@Valid @RequestBody ResetSendRequest req);

    @PostMapping(BASE_API + "/reset/verify")
    ResetVerifyResponse resetVerify(@Valid @RequestBody ResetVerifyRequest req);

    @PostMapping(BASE_API + "/reset/confirm")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    void resetConfirm(@Valid @RequestBody ResetConfirmRequest req);
}
