package ru.lambdahub.chat.db;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.IdClass;
import jakarta.persistence.Table;
import java.io.Serializable;
import java.time.Instant;
import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import lombok.experimental.FieldNameConstants;
import lombok.experimental.SuperBuilder;
import ru.lambdahub.chat.db.entity.audit.AuditableWithUser;

@Entity
@Table(name = "chat_members")
@IdClass(ChatMember.Pk.class)
@Data
@SuperBuilder
@AllArgsConstructor
@NoArgsConstructor
@FieldNameConstants
@EqualsAndHashCode(callSuper = true)
public class ChatMember extends AuditableWithUser {

    @Id
    @Column(name = "chat_id", nullable = false)
    private UUID chatId;

    @Id
    @Column(name = "user_id", nullable = false)
    private UUID userId;

    @Builder.Default
    @Column(name = "role", nullable = false)
    private String role = "member";

    @Column(name = "joined_at", nullable = false)
    private Instant joinedAt;

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Pk implements Serializable {
        private UUID chatId;
        private UUID userId;
    }
}
