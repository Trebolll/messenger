package ru.lambdahub.chat.db;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import lombok.experimental.FieldNameConstants;
import lombok.experimental.SuperBuilder;
import ru.lambdahub.chat.db.entity.audit.AuditableWithUser;

@Entity
@Table(name = "chats")
@Data
@SuperBuilder
@AllArgsConstructor
@NoArgsConstructor
@FieldNameConstants
@EqualsAndHashCode(callSuper = true)
public class Chat extends AuditableWithUser {

    @Id
    @Column(name = "id", nullable = false)
    private UUID id;

    @Column(name = "type", nullable = false)
    private String type;

    @Column(name = "name")
    private String name;

    @Column(name = "avatar_url")
    private String avatarUrl;

    @Column(name = "created_by", nullable = false)
    private UUID createdBy;
}
