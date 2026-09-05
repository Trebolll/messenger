package ru.lambdahub.user.web;

import java.util.List;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.RestController;
import ru.lambdahub.common.security.UserPrincipal;
import ru.lambdahub.user.api.ProfileResponse;
import ru.lambdahub.user.api.UpdateAvatarRequest;
import ru.lambdahub.user.api.UpdateProfileRequest;
import ru.lambdahub.user.api.UpdateStatusRequest;
import ru.lambdahub.user.api.UsersApi;
import ru.lambdahub.user.service.scenario.lhuser_01.Lhuser01ScenarioService;
import ru.lambdahub.user.service.scenario.lhuser_02.Lhuser02ScenarioService;
import ru.lambdahub.user.service.scenario.lhuser_03.Lhuser03ScenarioService;
import ru.lambdahub.user.service.scenario.lhuser_04.Lhuser04ScenarioService;
import ru.lambdahub.user.service.scenario.lhuser_05.Lhuser05ScenarioService;

@RestController
@RequiredArgsConstructor
public class UserController implements UsersApi {

    private final Lhuser01ScenarioService lhuser01;
    private final Lhuser02ScenarioService lhuser02;
    private final Lhuser03ScenarioService lhuser03;
    private final Lhuser04ScenarioService lhuser04;
    private final Lhuser05ScenarioService lhuser05;

    @Override
    public ProfileResponse me(UserPrincipal principal) {
        return lhuser01.process(principal.userId());
    }

    @Override
    public ProfileResponse byId(UUID id) {
        return lhuser01.process(id);
    }

    @Override
    public List<ProfileResponse> search(String q) {
        return lhuser02.process(q);
    }

    @Override
    public ProfileResponse updateProfile(UserPrincipal principal, UpdateProfileRequest req) {
        return lhuser03.process(principal.userId(), req);
    }

    @Override
    public ProfileResponse updateStatus(UserPrincipal principal, UpdateStatusRequest req) {
        return lhuser04.process(principal.userId(), req);
    }

    @Override
    public ProfileResponse updateAvatar(UserPrincipal principal, UpdateAvatarRequest req) {
        return lhuser05.process(principal.userId(), req);
    }
}
