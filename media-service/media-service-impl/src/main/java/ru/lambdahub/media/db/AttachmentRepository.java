package ru.lambdahub.media.db;

import io.vavr.control.Option;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;

public interface AttachmentRepository extends JpaRepository<Attachment, UUID> {

    default Option<Attachment> findOne(UUID id) {
        return Option.ofOptional(findById(id));
    }
}
