package ru.lambdahub.media.db;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import lombok.experimental.FieldNameConstants;
import lombok.experimental.SuperBuilder;
import ru.lambdahub.media.db.entity.audit.AuditableWithUser;

@Entity
@Table(name = "attachments")
@Data
@SuperBuilder
@AllArgsConstructor
@NoArgsConstructor
@FieldNameConstants
@EqualsAndHashCode(callSuper = true)
public class Attachment extends AuditableWithUser {

    @Id
    @Column(name = "id", nullable = false)
    private UUID id;

    @Column(name = "object_id", nullable = false)
    private UUID objectId;

    @Column(name = "chat_id")
    private UUID chatId;

    @Column(name = "message_id")
    private UUID messageId;

    @Builder.Default
    @Column(name = "kind")
    private String kind = "file";
}
