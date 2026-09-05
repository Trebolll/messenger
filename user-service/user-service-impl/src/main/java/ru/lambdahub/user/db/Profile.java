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
@Table(name = "profiles")
@Data
@SuperBuilder
@AllArgsConstructor
@NoArgsConstructor
@FieldNameConstants
@EqualsAndHashCode(callSuper = true)
public class Profile extends AuditableWithUser {

    @Id
    @Column(name = "user_id", nullable = false)
    private UUID userId;

    @Column(name = "username", nullable = false, unique = true)
    private String username;

    @Column(name = "display_name", nullable = false)
    private String displayName;

    @Column(name = "email")
    private String email;

    @Column(name = "phone")
    private String phone;

    @Column(name = "avatar_url")
    private String avatarUrl;

    @Builder.Default
    @Column(name = "status_text")
    private String statusText = "";

    @Builder.Default
    @Column(name = "profession")
    private String profession = "";

    @Builder.Default
    @Column(name = "location")
    private String location = "";
}
