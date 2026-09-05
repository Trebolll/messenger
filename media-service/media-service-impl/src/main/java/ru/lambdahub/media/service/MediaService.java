package ru.lambdahub.media.service;

import io.minio.GetObjectArgs;
import io.minio.MinioClient;
import io.minio.PutObjectArgs;
import io.minio.RemoveObjectArgs;
import io.vavr.control.Option;
import java.io.InputStream;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.multipart.MultipartFile;
import ru.lambdahub.common.kafka.OutputBridge;
import ru.lambdahub.media.api.CreateAttachmentCommand;
import ru.lambdahub.media.api.CreateMediaObjectCommand;
import ru.lambdahub.media.api.MediaApi;
import ru.lambdahub.media.api.UploadResponse;
import ru.lambdahub.media.db.AttachmentRepository;
import ru.lambdahub.media.db.MediaObject;
import ru.lambdahub.media.db.MediaObjectRepository;
import ru.lambdahub.media.dto.AttachmentDto;
import ru.lambdahub.media.dto.MediaObjectDto;
import ru.lambdahub.media.mapstruct.MediaMapper;
import ru.lambdahub.media.validation.LhmediaValidInfo;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;

@Slf4j
@Service
@RequiredArgsConstructor
public class MediaService {

    private final MinioClient minio;
    private final MediaObjectRepository objects;
    private final AttachmentRepository attachments;
    private final ObjectProvider<OutputBridge> attachmentCreatedOutput;
    private final MediaMapper mediaMapper;
    @Value("${app.minio.bucket}")
    private final String bucket;

    @Transactional
    public UploadResponse upload(UUID userId, MultipartFile file, UUID chatId, UUID messageId)
            throws Exception {
        if (file == null || file.isEmpty()) {
            Outcome outcome = new Outcome().addValidCode(LhmediaValidInfo.LHMEDIA_0002);
            throw ValidationExceptionUtils.createValidationException(outcome);
        }

        String objectName = UUID.randomUUID() + "_" + sanitize(file.getOriginalFilename());
        try (InputStream in = file.getInputStream()) {
            minio.putObject(PutObjectArgs.builder()
                    .bucket(bucket)
                    .object(objectName)
                    .stream(in, file.getSize(), -1)
                    .contentType(file.getContentType())
                    .build());
        }

        CreateMediaObjectCommand objectCmd = CreateMediaObjectCommand.builder()
                .id(UUID.randomUUID())
                .bucket(bucket)
                .objectName(objectName)
                .mime(file.getContentType())
                .sizeBytes(file.getSize())
                .uploadedBy(userId)
                .build();
        MediaObjectDto obj = mediaMapper.toDto(objectCmd);
        objects.save(mediaMapper.toEntity(obj));

        CreateAttachmentCommand attachmentCmd = CreateAttachmentCommand.builder()
                .id(UUID.randomUUID())
                .objectId(obj.getId())
                .chatId(chatId)
                .messageId(messageId)
                .kind("file")
                .build();
        AttachmentDto att = mediaMapper.toDto(attachmentCmd);
        attachments.save(mediaMapper.toEntity(att));

        String url = MediaApi.BASE_API + "/download/" + objectName;
        attachmentCreatedOutput.ifAvailable(ob ->
                ob.send(mediaMapper.toKafka(att, url, obj.getMime(), null)));

        return mediaMapper.toApi(att, obj, url);
    }

    public byte[] download(String objectName) throws Exception {
        try (InputStream in = minio.getObject(GetObjectArgs.builder().bucket(bucket).object(objectName).build())) {
            return in.readAllBytes();
        }
    }

    @Transactional
    public void delete(String objectName) throws Exception {
        Option<MediaObject> found = objects.findByObjectName(objectName);
        Outcome outcome = new Outcome();
        if (found.isEmpty()) {
            outcome = outcome.addValidCode(LhmediaValidInfo.LHMEDIA_0001, objectName);
        }
        if (outcome.hasCriticalErrors()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }
        MediaObject obj = found.get();
        minio.removeObject(RemoveObjectArgs.builder().bucket(bucket).object(objectName).build());
        objects.delete(obj);
    }

    private String sanitize(String name) {
        if (name == null || name.isBlank()) {
            return "file";
        }
        return name.replaceAll("[^a-zA-Z0-9._-]", "_");
    }
}
