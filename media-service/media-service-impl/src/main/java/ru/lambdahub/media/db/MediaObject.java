package ru.lambdahub.media.db;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import lombok.experimental.FieldNameConstants;
import lombok.experimental.SuperBuilder;
import ru.lambdahub.media.db.entity.audit.AuditableWithUser;

@Entity
@Table(name = "objects")
@Data
@SuperBuilder
@AllArgsConstructor
@NoArgsConstructor
@FieldNameConstants
@EqualsAndHashCode(callSuper = true)
public class MediaObject extends AuditableWithUser {

    @Id
    @Column(name = "id", nullable = false)
    private UUID id;

    @Column(name = "bucket", nullable = false)
    private String bucket;

    @Column(name = "object_name", nullable = false)
    private String objectName;

    @Column(name = "mime")
    private String mime;

    @Column(name = "size_bytes")
    private Long sizeBytes;

    @Column(name = "uploaded_by")
    private UUID uploadedBy;
}
