package ru.lambdahub.chat.db;

import io.vavr.collection.List;
import io.vavr.control.Option;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

public interface MessageRepository extends JpaRepository<Message, UUID> {

    @Query("select m from Message m where m.chatId = :chatId and m.deletedAt is null order by m.createdAt desc")
    List<Message> findRecent(@Param("chatId") UUID chatId);

    Option<Message> findFirstByChatIdAndDeletedAtIsNullOrderByCreatedAtDesc(UUID chatId);

    default Option<Message> findOne(UUID id) {
        return Option.ofOptional(findById(id));
    }
}
