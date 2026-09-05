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
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import lombok.experimental.FieldNameConstants;
import lombok.experimental.SuperBuilder;
import ru.lambdahub.chat.db.entity.audit.AuditableWithUser;

@Entity
@Table(name = "message_reads")
@IdClass(MessageRead.Pk.class)
@Data
@SuperBuilder
@AllArgsConstructor
@NoArgsConstructor
@FieldNameConstants
@EqualsAndHashCode(callSuper = true)
public class MessageRead extends AuditableWithUser {

    @Id
    @Column(name = "chat_id", nullable = false)
    private UUID chatId;

    @Id
    @Column(name = "user_id", nullable = false)
    private UUID userId;

    @Column(name = "last_read_message_id")
    private UUID lastReadMessageId;

    @Column(name = "read_at", nullable = false)
    private Instant readAt;

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Pk implements Serializable {
        private UUID chatId;
        private UUID userId;
    }
}
