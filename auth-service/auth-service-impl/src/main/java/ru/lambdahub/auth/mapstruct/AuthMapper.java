package ru.lambdahub.auth.mapstruct;

import java.util.UUID;
import org.mapstruct.Mapper;
import org.mapstruct.Mapping;
import org.mapstruct.ReportingPolicy;
import ru.lambdahub.auth.api.RegisterRequest;
import ru.lambdahub.auth.api.UserView;
import ru.lambdahub.auth.db.Credential;
import ru.lambdahub.auth.dto.CredentialDto;
import ru.lambdahub.auth.dto.UserDto;
import ru.lambdahub.common.events.kafka.DtoKafkaUserCreated;

@Mapper(componentModel = "spring", unmappedTargetPolicy = ReportingPolicy.IGNORE)
public interface AuthMapper {

    CredentialDto toDto(Credential entity);

    Credential toEntity(CredentialDto dto);

    @Mapping(target = "fullName", ignore = true)
    UserDto toUserDto(CredentialDto cred);

    @Mapping(target = "fullName", source = "fullName")
    UserDto toUserDto(CredentialDto cred, String fullName);

    @Mapping(target = "id", expression = "java(user.getUserId() == null ? null : user.getUserId().toString())")
    UserView toApi(UserDto user);

    default UserView toApi(CredentialDto cred, String fullName) {
        return toApi(toUserDto(cred, fullName));
    }

    default UserView toApi(Credential cred, String fullName) {
        return toApi(toDto(cred), fullName);
    }

    @Mapping(target = "userId", source = "cred.userId")
    @Mapping(target = "username", source = "cred.username")
    @Mapping(target = "email", source = "cred.email")
    @Mapping(target = "phone", source = "cred.phone")
    DtoKafkaUserCreated toKafka(CredentialDto cred, UUID requestId, String fullName);

    default DtoKafkaUserCreated toKafka(Credential cred, UUID requestId, String fullName) {
        return toKafka(toDto(cred), requestId, fullName);
    }

    /** @deprecated prefer {@link #toKafka(CredentialDto, UUID, String)} */
    default DtoKafkaUserCreated toDto(Credential cred, UUID requestId, String fullName) {
        return toKafka(cred, requestId, fullName);
    }

    @Mapping(target = "username", expression = "java(req.getUsername() == null ? null : req.getUsername().trim())")
    @Mapping(target = "passwordHash", ignore = true)
    @Mapping(target = "userId", ignore = true)
    @Mapping(target = "email", ignore = true)
    @Mapping(target = "phone", ignore = true)
    CredentialDto fromRequest(RegisterRequest req);

    default Credential toEntity(RegisterRequest req) {
        return toEntity(fromRequest(req));
    }
}
