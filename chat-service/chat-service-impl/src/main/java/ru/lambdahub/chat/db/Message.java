package ru.lambdahub.chat.db;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
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
@Table(name = "messages")
@Data
@SuperBuilder
@AllArgsConstructor
@NoArgsConstructor
@FieldNameConstants
@EqualsAndHashCode(callSuper = true)
public class Message extends AuditableWithUser {

    @Id
    @Column(name = "id", nullable = false)
    private UUID id;

    @Column(name = "chat_id", nullable = false)
    private UUID chatId;

    @Column(name = "sender_id", nullable = false)
    private UUID senderId;

    @Builder.Default
    @Column(name = "content", nullable = false)
    private String content = "";

    @Column(name = "parent_id")
    private UUID parentId;

    @Column(name = "edited_at")
    private Instant editedAt;

    @Column(name = "deleted_at")
    private Instant deletedAt;
}
