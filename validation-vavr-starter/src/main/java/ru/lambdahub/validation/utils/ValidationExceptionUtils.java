package ru.lambdahub.validation.utils;

import lombok.experimental.UtilityClass;
import org.springframework.http.HttpStatus;
import ru.lambdahub.validation.exception.ValidationException;
import ru.lambdahub.validation.message.Message;
import ru.lambdahub.validation.message.OutboundMessage;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.valid.Note;
import ru.lambdahub.validation.valid.error.ValidError;

@UtilityClass
public class ValidationExceptionUtils {

    public ValidationException createValidationException(Message output) {
        return new ValidationException(HttpStatus.PRECONDITION_FAILED, output);
    }

    public ValidationException createValidationException(Outcome outcome) {
        return new ValidationException(HttpStatus.PRECONDITION_FAILED, outcome);
    }

    public ValidationException createValidationExceptionByException(Throwable ex) {
        OutboundMessage outboundMessage = new OutboundMessage();
        outboundMessage.addValidCode(ValidError.EXCEPTION, ex.getMessage());
        return new ValidationException(HttpStatus.PRECONDITION_FAILED, outboundMessage);
    }

    public Note noteByException(Throwable ex) {
        return ValidError.DB_ERROR.withArgs(ex.getMessage());
    }
}
