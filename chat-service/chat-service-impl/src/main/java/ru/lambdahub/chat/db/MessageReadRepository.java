package ru.lambdahub.chat.db;

import io.vavr.control.Option;
import org.springframework.data.jpa.repository.JpaRepository;

public interface MessageReadRepository extends JpaRepository<MessageRead, MessageRead.Pk> {

    default Option<MessageRead> findOne(MessageRead.Pk id) {
        return Option.ofOptional(findById(id));
    }
}
