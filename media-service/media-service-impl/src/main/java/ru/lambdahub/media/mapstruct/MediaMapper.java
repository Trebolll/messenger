package ru.lambdahub.media.mapstruct;

import java.util.UUID;
import org.mapstruct.Mapper;
import org.mapstruct.Mapping;
import org.mapstruct.ReportingPolicy;
import ru.lambdahub.common.events.kafka.DtoKafkaAttachment;
import ru.lambdahub.media.api.CreateAttachmentCommand;
import ru.lambdahub.media.api.CreateMediaObjectCommand;
import ru.lambdahub.media.api.UploadResponse;
import ru.lambdahub.media.db.Attachment;
import ru.lambdahub.media.db.MediaObject;
import ru.lambdahub.media.dto.AttachmentDto;
import ru.lambdahub.media.dto.MediaObjectDto;

@Mapper(componentModel = "spring", unmappedTargetPolicy = ReportingPolicy.IGNORE)
public interface MediaMapper {

    MediaObjectDto toDto(MediaObject entity);

    MediaObject toEntity(MediaObjectDto dto);

    AttachmentDto toDto(Attachment entity);

    Attachment toEntity(AttachmentDto dto);

    MediaObjectDto toDto(CreateMediaObjectCommand cmd);

    AttachmentDto toDto(CreateAttachmentCommand cmd);

    @Mapping(target = "attachmentId", source = "attachment.id")
    @Mapping(target = "objectId", source = "attachment.objectId")
    @Mapping(target = "chatId", source = "attachment.chatId")
    @Mapping(target = "messageId", source = "attachment.messageId")
    DtoKafkaAttachment toKafka(AttachmentDto attachment, String url, String mime, UUID requestId);

    @Mapping(target = "id", expression = "java(attachment.getId() == null ? null : attachment.getId().toString())")
    @Mapping(target = "objectName", source = "object.objectName")
    @Mapping(target = "url", source = "url")
    @Mapping(target = "mime", source = "object.mime")
    @Mapping(target = "sizeBytes", source = "object.sizeBytes")
    @Mapping(target = "requestId", ignore = true)
    @Mapping(target = "outcome", ignore = true)
    UploadResponse toApi(AttachmentDto attachment, MediaObjectDto object, String url);
}
