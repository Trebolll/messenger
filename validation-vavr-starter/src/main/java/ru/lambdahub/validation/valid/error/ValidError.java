package ru.lambdahub.validation.valid.error;

import ru.lambdahub.validation.valid.ValidCode;

public enum ValidError implements ValidCode {

    REQUIRED_FIELD_MISSING(ErrNotice.REQUIRED_FIELD_MISMATCH),
    VALID_ERROR_1(ErrNotice.NOTICE_1),
    VALID_ERROR_2(ErrNotice.NOTICE_2),
    INVALID_NAME_FORMAT(ErrNotice.INVALID_NAME),
    INVALID_EMAIL_FORMAT(ErrNotice.INVALID_EMAIL),
    FIELD_LENGTH_INVALID(ErrNotice.INVALID_LENGTH),
    RECORD_NOT_FOUND(ErrNotice.RECORD_NOT_FOUND),
    DB_UPDATE_ERROR(ErrNotice.DB_UPDATE_ERROR),
    DB_SAVE_ERROR(ErrNotice.DB_SAVE_ERROR),
    DB_ERROR(ErrNotice.DB_ERROR),
    EXCEPTION(ErrNotice.EXCEPTION);

    private final String code;
    private final String message;

    ValidError(String message) {
        this.code = name();
        this.message = message;
    }

    @Override
    public String getCode() {
        return code;
    }

    @Override
    public String getMessage() {
        return message;
    }

    @Override
    public boolean isCritical() {
        return true;
    }
}
