package ru.lambdahub.auth.db;

import io.vavr.control.Option;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;

public interface CredentialRepository extends JpaRepository<Credential, UUID> {

    boolean existsByUsernameIgnoreCase(String username);

    Option<Credential> findByEmailIgnoreCase(String email);

    Option<Credential> findByPhone(String phone);

    Option<Credential> findByUsernameIgnoreCase(String username);

    default Option<Credential> findOne(UUID id) {
        return Option.ofOptional(findById(id));
    }
}
