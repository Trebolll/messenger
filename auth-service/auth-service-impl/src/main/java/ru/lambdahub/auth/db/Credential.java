package ru.lambdahub.auth.db;

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
import ru.lambdahub.auth.db.entity.audit.AuditableWithUser;

@Entity
@Table(name = "credentials")
@Data
@SuperBuilder
@AllArgsConstructor
@NoArgsConstructor
@FieldNameConstants
@EqualsAndHashCode(callSuper = true)
public class Credential extends AuditableWithUser {

  @Id
  @Column(name = "user_id", nullable = false)
  private UUID userId;

  @Column(name = "email")
  private String email;

  @Column(name = "phone")
  private String phone;

  @Column(name = "username", nullable = false, unique = true)
  private String username;

  @Column(name = "password_hash", nullable = false)
  private String passwordHash;
}
