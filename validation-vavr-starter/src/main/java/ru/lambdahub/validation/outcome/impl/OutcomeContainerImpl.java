package ru.lambdahub.validation.outcome.impl;

import lombok.Setter;
import ru.lambdahub.validation.outcome.Outcome;
import ru.lambdahub.validation.outcome.OutcomeContainer;
import ru.lambdahub.validation.valid.ValidCode;

public class OutcomeContainerImpl implements OutcomeContainer {

    @Setter
    private Outcome outcome = new Outcome();

    @Override
    public Outcome getOutcome() {
        return outcome;
    }

    @Override
    public void addValidCode(ValidCode code, Object... args) {
        outcome = outcome.addValidCode(code, args);
    }
}
