package ru.lambdahub.validation.message;

import jakarta.annotation.Nullable;
import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.experimental.SuperBuilder;

@Getter
@SuperBuilder(toBuilder = true)
@NoArgsConstructor
@AllArgsConstructor
public class InboundMessage implements Message {

    @Nullable
    private UUID requestId;

    @Override
    public void setRequestId(UUID reqId) {
        this.requestId = reqId;
    }
}
