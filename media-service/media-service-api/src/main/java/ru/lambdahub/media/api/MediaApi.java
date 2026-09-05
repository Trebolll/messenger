package ru.lambdahub.media.api;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.multipart.MultipartFile;
import ru.lambdahub.common.security.UserPrincipal;

import java.util.UUID;

public interface MediaApi {

    String BASE_API = "/api/storage";

    @PostMapping(BASE_API + "/upload")
    UploadResponse upload(@AuthenticationPrincipal UserPrincipal principal,
                                    @RequestParam("file") MultipartFile file,
                                    @RequestParam(value = "chatId", required = false) UUID chatId,
                                    @RequestParam(value = "messageId", required = false) UUID messageId) throws Exception;

    @GetMapping(BASE_API + "/download/{objectName}")
    ResponseEntity<byte[]> download(@PathVariable String objectName) throws Exception;

    @DeleteMapping(BASE_API + "/delete/{objectName}")
    void delete(@PathVariable String objectName) throws Exception;
}
