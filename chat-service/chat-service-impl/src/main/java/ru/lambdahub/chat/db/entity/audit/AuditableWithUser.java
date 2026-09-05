package ru.lambdahub.chat.db.entity.audit;

import jakarta.persistence.Column;
import jakarta.persistence.EntityListeners;
import jakarta.persistence.MappedSuperclass;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import lombok.experimental.SuperBuilder;
import org.springframework.data.annotation.CreatedBy;
import org.springframework.data.annotation.LastModifiedBy;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;

@Data
@SuperBuilder
@NoArgsConstructor
@EqualsAndHashCode(callSuper = true)
@EntityListeners(AuditingEntityListener.class)
@MappedSuperclass
public abstract class AuditableWithUser extends Auditable {

    @CreatedBy
    @Column(name = "create_user", updatable = false, nullable = false)
    private String createUser;

    @LastModifiedBy
    @Column(name = "update_user", nullable = false)
    private String updateUser;
}
