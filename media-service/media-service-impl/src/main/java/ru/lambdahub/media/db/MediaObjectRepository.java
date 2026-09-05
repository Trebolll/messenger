package ru.lambdahub.media.db;

import io.vavr.control.Option;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;

public interface MediaObjectRepository extends JpaRepository<MediaObject, UUID> {

    Option<MediaObject> findByObjectName(String objectName);

    default Option<MediaObject> findOne(UUID id) {
        return Option.ofOptional(findById(id));
    }
}
