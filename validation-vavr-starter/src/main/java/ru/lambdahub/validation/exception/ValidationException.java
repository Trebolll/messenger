package ru.lambdahub.validation.exception;

import java.util.Optional;
import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.ToString;
import org.springframework.http.HttpStatus;
import ru.lambdahub.validation.message.Message;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.outcome.OutcomeContainer;

@Getter
@EqualsAndHashCode(callSuper = false)
@ToString
public class ValidationException extends RuntimeException {

    private final transient HttpStatus status;
    private final transient Outcome outcome;
    private final transient Message response;

    public ValidationException(HttpStatus status, Outcome outcome) {
        super(bodyToMessage(outcome));
        this.status = status;
        this.outcome = outcome == null ? new Outcome() : outcome;
        this.response = null;
    }

    public ValidationException(HttpStatus status, Message response) {
        super(bodyToMessage(response));
        this.status = status;
        this.response = response;
        this.outcome = response instanceof OutcomeContainer oc ? oc.getOutcome() : new Outcome();
    }

    private static String bodyToMessage(Object body) {
        return Optional.ofNullable(body).map(Object::toString).orElse(null);
    }
}
