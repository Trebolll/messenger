package ru.lambdahub.media.validation;

import ru.lambdahub.validation.valid.ValidCode;

/** Scenario-bound ValidInfo for media-service. */
public enum LhmediaValidInfo implements ValidCode {

    LHMEDIA_0001("Object not found: %s", true),
    LHMEDIA_0002("Upload file is required", true);

    private final String code;
    private final String message;
    private final boolean critical;

    LhmediaValidInfo(String message, boolean critical) {
        this.code = name();
        this.message = message;
        this.critical = critical;
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
        return critical;
    }
}
