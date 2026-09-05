package ru.lambdahub.user.mapstruct;

import java.util.UUID;
import org.mapstruct.Mapper;
import org.mapstruct.Mapping;
import org.mapstruct.ReportingPolicy;
import ru.lambdahub.common.events.kafka.DtoKafkaUserCreated;
import ru.lambdahub.common.events.kafka.DtoKafkaUserProfile;
import ru.lambdahub.common.events.kafka.DtoKafkaUserStatus;
import ru.lambdahub.user.api.ProfileResponse;
import ru.lambdahub.user.db.Profile;
import ru.lambdahub.user.db.User;
import ru.lambdahub.user.dto.ProfileDto;
import ru.lambdahub.user.dto.UserDto;

@Mapper(componentModel = "spring", unmappedTargetPolicy = ReportingPolicy.IGNORE)
public interface UserMapper {

    UserDto toDto(User entity);

    User toEntity(UserDto dto);

    @Mapping(target = "id", source = "userId")
    @Mapping(target = "status", constant = "ACTIVE")
    UserDto toUserDto(DtoKafkaUserCreated dto);

    ProfileDto toDto(Profile entity);

    Profile toEntity(ProfileDto dto);

    @Mapping(
            target = "displayName",
            expression = "java(dto.getFullName() != null && !dto.getFullName().isBlank() ? dto.getFullName() : dto.getUsername())")
    @Mapping(target = "avatarUrl", ignore = true)
    @Mapping(target = "statusText", constant = "")
    @Mapping(target = "profession", constant = "")
    @Mapping(target = "location", constant = "")
    ProfileDto toDto(DtoKafkaUserCreated dto);

    default Profile toEntity(DtoKafkaUserCreated dto) {
        return toEntity(toDto(dto));
    }

    DtoKafkaUserProfile toKafka(ProfileDto profile, UUID requestId);

    DtoKafkaUserStatus toStatusKafka(ProfileDto profile, UUID requestId);

    default DtoKafkaUserProfile toKafka(Profile profile, UUID requestId) {
        return toKafka(toDto(profile), requestId);
    }

    default DtoKafkaUserStatus toStatusKafka(Profile profile, UUID requestId) {
        return toStatusKafka(toDto(profile), requestId);
    }

    default DtoKafkaUserProfile toDto(Profile profile, UUID requestId) {
        return toKafka(profile, requestId);
    }

    default DtoKafkaUserStatus toStatusDto(Profile profile, UUID requestId) {
        return toStatusKafka(profile, requestId);
    }

    @Mapping(target = "id", source = "userId")
    @Mapping(target = "requestId", ignore = true)
    @Mapping(target = "outcome", ignore = true)
    ProfileResponse toApi(ProfileDto profile);

    default ProfileResponse toApi(Profile profile) {
        return toApi(toDto(profile));
    }
}
