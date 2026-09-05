package ru.lambdahub.validation.outcome;

import ru.lambdahub.validation.valid.ValidCode;

public interface OutcomeContainer {

    Outcome getOutcome();

    void addValidCode(ValidCode code, Object... args);
}
