package ru.lambdahub.validation.exception;

import io.vavr.collection.List;
import io.vavr.collection.Seq;
import lombok.Data;
import lombok.EqualsAndHashCode;
import ru.lambdahub.validation.message.OutcomeMessage;

@Data
@EqualsAndHashCode(callSuper = true)
public class RequestFailedException extends RuntimeException {

    private int status;
    private Seq<OutcomeMessage> notes = List.empty();

    public RequestFailedException() {
        super("Request failed");
    }

    public RequestFailedException(String message) {
        super(message);
    }
}
