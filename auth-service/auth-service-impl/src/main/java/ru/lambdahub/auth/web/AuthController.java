package ru.lambdahub.auth.web;

import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.RestController;
import ru.lambdahub.auth.api.AuthApi;
import ru.lambdahub.auth.api.AuthResponse;
import ru.lambdahub.auth.api.LoginRequest;
import ru.lambdahub.auth.api.RegisterRequest;
import ru.lambdahub.auth.api.ResetConfirmRequest;
import ru.lambdahub.auth.api.ResetSendRequest;
import ru.lambdahub.auth.api.ResetVerifyRequest;
import ru.lambdahub.auth.api.ResetVerifyResponse;
import ru.lambdahub.auth.api.SendCodeRequest;
import ru.lambdahub.auth.api.SendCodeResponse;
import ru.lambdahub.auth.api.VerifyCodeRequest;
import ru.lambdahub.auth.api.VerifyCodeResponse;
import ru.lambdahub.auth.service.scenario.lhauth_01.Lhauth01ScenarioService;
import ru.lambdahub.auth.service.scenario.lhauth_02.Lhauth02ScenarioService;
import ru.lambdahub.auth.service.scenario.lhauth_03.Lhauth03ScenarioService;
import ru.lambdahub.auth.service.scenario.lhauth_04.Lhauth04ScenarioService;
import ru.lambdahub.auth.service.scenario.lhauth_05.Lhauth05ScenarioService;
import ru.lambdahub.auth.service.scenario.lhauth_06.Lhauth06ScenarioService;
import ru.lambdahub.auth.service.scenario.lhauth_07.Lhauth07ScenarioService;

@RestController
@RequiredArgsConstructor
public class AuthController implements AuthApi {

    private final Lhauth01ScenarioService lhauth01;
    private final Lhauth02ScenarioService lhauth02;
    private final Lhauth03ScenarioService lhauth03;
    private final Lhauth04ScenarioService lhauth04;
    private final Lhauth05ScenarioService lhauth05;
    private final Lhauth06ScenarioService lhauth06;
    private final Lhauth07ScenarioService lhauth07;

    @Override
    public SendCodeResponse send(SendCodeRequest req) {
        return lhauth01.process(req);
    }

    @Override
    public VerifyCodeResponse verify(VerifyCodeRequest req) {
        return lhauth02.process(req);
    }

    @Override
    public AuthResponse register(RegisterRequest req) {
        return lhauth03.process(req);
    }

    @Override
    public AuthResponse login(LoginRequest req) {
        return lhauth04.process(req);
    }

    @Override
    public SendCodeResponse resetSend(ResetSendRequest req) {
        return lhauth05.process(req);
    }

    @Override
    public ResetVerifyResponse resetVerify(ResetVerifyRequest req) {
        return lhauth06.process(req);
    }

    @Override
    public void resetConfirm(ResetConfirmRequest req) {
        lhauth07.process(req);
    }
}
