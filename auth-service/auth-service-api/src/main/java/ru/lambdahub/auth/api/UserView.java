package ru.lambdahub.auth.api;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;
import lombok.ToString;
import lombok.experimental.FieldNameConstants;

@Getter
@Setter
@Builder
@NoArgsConstructor
@AllArgsConstructor
@FieldNameConstants
@ToString
@EqualsAndHashCode
public class UserView {

    private String id;
    private String username;
    private String email;
    private String phone;
    private String fullName;
}
