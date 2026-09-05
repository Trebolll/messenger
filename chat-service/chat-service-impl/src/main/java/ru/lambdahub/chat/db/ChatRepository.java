package ru.lambdahub.chat.db;

import io.vavr.control.Option;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;

public interface ChatRepository extends JpaRepository<Chat, UUID> {

    default Option<Chat> findOne(UUID id) {
        return Option.ofOptional(findById(id));
    }
}
