package ru.lambdahub.chat.db;

import io.vavr.collection.List;
import io.vavr.control.Option;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

public interface ChatMemberRepository extends JpaRepository<ChatMember, ChatMember.Pk> {

    List<ChatMember> findByUserId(UUID userId);

    List<ChatMember> findByChatId(UUID chatId);

    boolean existsByChatIdAndUserId(UUID chatId, UUID userId);

    void deleteByChatIdAndUserId(UUID chatId, UUID userId);

    @Query("select m.chatId from ChatMember m where m.userId = :userId")
    List<UUID> findChatIdsByUserId(@Param("userId") UUID userId);

    default Option<ChatMember> findOne(ChatMember.Pk id) {
        return Option.ofOptional(findById(id));
    }
}
