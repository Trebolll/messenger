package ru.lambdahub.validation.validator;

import io.vavr.collection.List;
import io.vavr.collection.Seq;
import io.vavr.control.Validation;
import jakarta.validation.Validator;
import lombok.RequiredArgsConstructor;
import ru.lambdahub.validation.message.OutcomeMessage;
import ru.lambdahub.validation.valid.ValidCode;
import ru.lambdahub.validation.valid.error.ValidError;

@RequiredArgsConstructor
public class OutcomeValidator {

    private final Validator validator;

    public <T> Validation<Seq<OutcomeMessage>, T> validate(T dto) {
        Seq<OutcomeMessage> notes = List.ofAll(validator.validate(dto))
                .map(v -> {
                    ValidCode code = ValidError.REQUIRED_FIELD_MISSING;
                    String formattedMessage = String.format(code.getMessage(), v.getPropertyPath().toString());
                    return new OutcomeMessage(code.getCode(), formattedMessage, code.isCritical());
                });

        if (notes.isEmpty()) {
            return Validation.valid(dto);
        }
        return Validation.invalid(notes);
    }
}
