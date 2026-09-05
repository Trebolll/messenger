package ru.lambdahub.validation.valid;

import ru.lambdahub.validation.valid.error.ValidError;

public interface ValidCode {

    String getCode();

    String getMessage();

    boolean isCritical();

    default Note withArgs(Object... args) {
        return new Note(this, args);
    }

    default ValidCode withNote(Object... args) {
        return ValidError.EXCEPTION;
    }
}
