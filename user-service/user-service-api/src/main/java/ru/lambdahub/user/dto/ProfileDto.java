package ru.lambdahub.user.dto;

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
public class ProfileDto {
    private UUID userId;
    private String username;
    private String displayName;
    private String email;
    private String phone;
    private String avatarUrl;
    private String statusText;
    private String profession;
    private String location;
}
