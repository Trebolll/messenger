package ru.lambdahub.user.db;

import io.vavr.control.Option;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;

public interface UserRepository extends JpaRepository<User, UUID> {

    default Option<User> findOne(UUID id) {
        return Option.ofOptional(findById(id));
    }
}
