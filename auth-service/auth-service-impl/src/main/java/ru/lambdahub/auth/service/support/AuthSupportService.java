package ru.lambdahub.auth.service.support;

import io.vavr.control.Option;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import ru.lambdahub.auth.db.Credential;
import ru.lambdahub.auth.db.CredentialRepository;
import ru.lambdahub.auth.dto.CredentialDto;
import ru.lambdahub.auth.mapstruct.AuthMapper;

@Service
@RequiredArgsConstructor
public class AuthSupportService {

    private final CredentialRepository credentials;
    private final AuthMapper authMapper;

    public Option<CredentialDto> findByLogin(String login) {
        return findCredential(login).map(authMapper::toDto);
    }

    public Option<Credential> findCredential(String login) {
        if (login == null || login.isBlank()) {
            return Option.none();
        }
        if (login.contains("@")) {
            return credentials.findByEmailIgnoreCase(login);
        }
        if (login.chars().allMatch(c -> Character.isDigit(c) || c == '+')) {
            return credentials.findByPhone(login);
        }
        return credentials.findByUsernameIgnoreCase(login);
    }
}
