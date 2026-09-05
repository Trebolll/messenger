package ru.lambdahub.validation.message;

import ru.lambdahub.validation.valid.ValidCode;

public record OutcomeMessage(String code, String message, boolean critical) {

    public static OutcomeMessage of(ValidCode code, Object... args) {
        return new OutcomeMessage(code.getCode(), String.format(code.getMessage(), args), code.isCritical());
    }
}
