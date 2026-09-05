package ru.lambdahub.validation.valid.error;

public final class ErrNotice {

    public static final String REQUIRED_FIELD_MISMATCH = "Required field mismatch: %s";
    public static final String NOTICE_1 = "Validation scenario required field error: %s";
    public static final String NOTICE_2 = "Something went wrong. Check the code. RequestId: %s";
    public static final String INVALID_NAME = "Invalid name format: %s";
    public static final String INVALID_EMAIL = "Email must contain '@' symbol, but was: %s";
    public static final String INVALID_LENGTH = "Field '%s' length must be between %d and %d";
    public static final String RECORD_NOT_FOUND = "Record not found by id: %s";
    public static final String DB_UPDATE_ERROR = "Database update error for id: %s";
    public static final String DB_SAVE_ERROR = "Database save error for id: %s";
    public static final String DB_ERROR = "Database error: %s";
    public static final String EXCEPTION = "Exception: %s";

    private ErrNotice() {}
}
