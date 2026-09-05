package ru.lambdahub.auth.validation;

import ru.lambdahub.validation.valid.ValidCode;

/** Scenario-bound ValidInfo for auth-service. */
public enum LhauthValidInfo implements ValidCode {

    LHAUTH_01_0001("Login is required", true),
    LHAUTH_03_0001("Username already taken: %s", true),
    LHAUTH_03_0002("Confirm token login is required", true),
    LHAUTH_04_0001("Invalid credentials", true),
    LHAUTH_05_0001("User not found: %s", true),
    LHAUTH_06_0001("User not found: %s", true),
    LHAUTH_07_0001("User not found: %s", true);

    private final String code;
    private final String message;
    private final boolean critical;

    LhauthValidInfo(String message, boolean critical) {
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
