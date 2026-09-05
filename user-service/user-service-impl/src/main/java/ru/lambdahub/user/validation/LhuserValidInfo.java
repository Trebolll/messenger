package ru.lambdahub.user.validation;

import ru.lambdahub.validation.valid.ValidCode;

/** Scenario-bound ValidInfo for LHUSER (LambdaHub User). Code prefix = scenario id. */
public enum LhuserValidInfo implements ValidCode {

    /** LHUSER-01 get profile — user not found */
    LHUSER_01_0001("User not found by id: %s", true),

    /** LHUSER-02 search — reserved (blank query returns empty list) */
    LHUSER_02_0001("Search query is required", false),

    /** LHUSER-03 update profile — user not found */
    LHUSER_03_0001("User not found by id: %s", true),
    /** LHUSER-03 — no fields to update */
    LHUSER_03_0002("At least one profile field is required", true),
    /** LHUSER-03 — username already taken */
    LHUSER_03_0003("Username already taken: %s", true),

    /** LHUSER-04 update status — user not found */
    LHUSER_04_0001("User not found by id: %s", true),

    /** LHUSER-05 update avatar — user not found */
    LHUSER_05_0001("User not found by id: %s", true),

    /** LHUSER-06 create from event — payload invalid */
    LHUSER_06_0001("User created event requires userId and username", true);

    private final String code;
    private final String message;
    private final boolean critical;

    LhuserValidInfo(String message, boolean critical) {
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
