package ru.lambdahub.user.db;

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
import ru.lambdahub.user.db.entity.audit.AuditableWithUser;

@Entity
@Table(name = "users")
@Data
@SuperBuilder
@AllArgsConstructor
@NoArgsConstructor
@FieldNameConstants
@EqualsAndHashCode(callSuper = true)
public class User extends AuditableWithUser {

    @Id
    @Column(name = "id", nullable = false)
    private UUID id;

    @Builder.Default
    @Column(name = "status", nullable = false)
    private String status = "ACTIVE";
}
