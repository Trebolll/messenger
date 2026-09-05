package ru.lambdahub.chat.validation;

import ru.lambdahub.validation.valid.ValidCode;

/**
 * Scenario-bound ValidInfo for LHCHT (LambdaHub Chat). Code prefix = scenario id.
 */
public enum LhchtValidInfo implements ValidCode {

    /** LHCHT-01 create private — cannot chat with yourself */
    LHCHT_01_0001("Cannot create private chat with yourself", true),

    /** LHCHT-02 create group — name required */
    LHCHT_02_0001("Group name required", true),
    /** LHCHT-02 — user not found by username */
    LHCHT_02_0002("User not found: %s", true),

    /** LHCHT-04 update chat — chat not found */
    LHCHT_04_0001("Chat not found by id: %s", true),
    /** LHCHT-04 — only groups can be renamed */
    LHCHT_04_0002("Only groups can be renamed", true),

    /** LHCHT-06 add member — chat not found */
    LHCHT_06_0001("Chat not found by id: %s", true),
    /** LHCHT-06 — only groups support members */
    LHCHT_06_0002("Only groups support members", true),
    /** LHCHT-06 — user not found by username */
    LHCHT_06_0003("User not found: %s", true),

    /** LHCHT-10 edit message — message not found */
    LHCHT_10_0001("Message not found by id: %s", true),

    /** LHCHT-11 delete message — message not found */
    LHCHT_11_0001("Message not found by id: %s", true),

    /** Shared username resolve (used from scenarios that resolve usernames) */
    LHCHT_USER_0001("User not found: %s", true);

    private final String code;
    private final String message;
    private final boolean critical;

    LhchtValidInfo(String message, boolean critical) {
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
