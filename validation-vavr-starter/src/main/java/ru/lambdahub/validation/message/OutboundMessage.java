package ru.lambdahub.validation.message;

import com.fasterxml.jackson.annotation.JsonIgnore;
import jakarta.annotation.Nullable;
import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;
import lombok.experimental.SuperBuilder;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.outcome.OutcomeContainer;
import ru.lambdahub.validation.utils.ValidationExceptionUtils;
import ru.lambdahub.validation.valid.ValidCode;

@Getter
@SuperBuilder(toBuilder = true)
@NoArgsConstructor
@AllArgsConstructor
public class OutboundMessage implements Message, OutcomeContainer {

    @Nullable
    private UUID requestId;

    @Setter
    @Builder.Default
    private Outcome outcome = new Outcome();

    @Override
    public void setRequestId(UUID reqId) {
        this.requestId = reqId;
    }

    @JsonIgnore
    public boolean isExecutable() {
        return outcome == null || !outcome.hasCriticalErrors();
    }

    @JsonIgnore
    public boolean isNoneExecutable() {
        return outcome != null && outcome.hasCriticalErrors();
    }

    @JsonIgnore
    public boolean hasIssues() {
        return outcome != null && !outcome.isEmpty();
    }

    @SuppressWarnings("unchecked")
    public <T extends OutboundMessage> T executeOrThrow() {
        if (!isExecutable()) {
            throw ValidationExceptionUtils.createValidationException(outcome);
        }
        return (T) this;
    }

    @Override
    public void addValidCode(ValidCode code, Object... args) {
        outcome = outcome.addValidCode(code, args);
    }
}
