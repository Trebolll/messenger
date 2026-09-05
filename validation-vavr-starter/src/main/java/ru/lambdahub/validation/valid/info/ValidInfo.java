package ru.lambdahub.validation.valid.info;

import ru.lambdahub.validation.valid.ValidCode;

/**
 * Generic system ValidInfo only. Business / not-found codes must live in scenario enums
 * (e.g. {@code LhchtValidInfo.LHCHT_04_0001}), not here.
 */
public enum ValidInfo implements ValidCode {

    VALID_INFO_1(InfoNotice.NOTICE_1),
    VALID_INFO_2(InfoNotice.NOTICE_2);

    private final String code;
    private final String message;

    ValidInfo(String message) {
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
        return false;
    }
}
