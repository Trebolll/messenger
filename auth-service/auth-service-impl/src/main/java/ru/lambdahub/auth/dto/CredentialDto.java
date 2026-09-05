package ru.lambdahub.auth.dto;

import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.experimental.FieldNameConstants;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@FieldNameConstants
public class CredentialDto {
    private UUID userId;
    private String email;
    private String phone;
    private String username;
    private String passwordHash;
}
