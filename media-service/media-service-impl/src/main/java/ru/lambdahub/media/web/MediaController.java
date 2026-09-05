package ru.lambdahub.media.web;

import java.util.UUID;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.multipart.MultipartFile;
import ru.lambdahub.common.security.UserPrincipal;
import ru.lambdahub.media.api.MediaApi;
import ru.lambdahub.media.api.UploadResponse;
import ru.lambdahub.media.service.MediaService;

@RestController
@RequiredArgsConstructor
public class MediaController implements MediaApi {

    private final MediaService mediaService;

    @Override
    public UploadResponse upload(UserPrincipal principal, MultipartFile file, UUID chatId, UUID messageId)
            throws Exception {
        return mediaService.upload(principal.userId(), file, chatId, messageId);
    }

    @Override
    public ResponseEntity<byte[]> download(String objectName) throws Exception {
        byte[] data = mediaService.download(objectName);
        return ResponseEntity.ok()
                .header(HttpHeaders.CONTENT_DISPOSITION, "inline; filename=\"" + objectName + "\"")
                .contentType(MediaType.APPLICATION_OCTET_STREAM)
                .body(data);
    }

    @Override
    public void delete(String objectName) throws Exception {
        mediaService.delete(objectName);
    }
}
