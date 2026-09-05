package ru.lambdahub.auth.dto;

import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.experimental.FieldNameConstants;

/** Auth-facing user projection (from credentials). Canonical row is {@code users.users}. */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@FieldNameConstants
public class UserDto {
    private UUID userId;
    private String username;
    private String email;
    private String phone;
    private String fullName;
}
