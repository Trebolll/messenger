package ru.lambdahub.validation.outcome;

import com.fasterxml.jackson.annotation.JsonIgnore;
import io.vavr.collection.List;
import io.vavr.collection.Seq;
import io.vavr.control.Either;
import io.vavr.control.Try;
import io.vavr.control.Validation;
import java.util.function.Supplier;
import lombok.AccessLevel;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.With;
import lombok.extern.slf4j.Slf4j;
import ru.lambdahub.validation.exception.RequestFailedException;
import ru.lambdahub.validation.message.OutcomeMessage;
import ru.lambdahub.validation.valid.Note;
import ru.lambdahub.validation.valid.ValidCode;
import ru.lambdahub.validation.valid.error.ValidError;

/**
 * Immutable container for operational results and notes (errors, info).
 */
@Getter
@With
@Slf4j
@NoArgsConstructor
@AllArgsConstructor(access = AccessLevel.PRIVATE)
public class Outcome {

    private Seq<OutcomeMessage> validCodes = List.empty();

    public static Outcome of(Validation<?, ?> validation) {
        return validation.toEither().fold(
                err -> new Outcome().addAllNotes(flattenNotes(err)),
                ignored -> new Outcome()
        );
    }

    public static Outcome ofCombined(Validation<?, ?> validation) {
        return of(validation);
    }

    private static Seq<OutcomeMessage> flattenNotes(Object error) {
        return switch (error) {
            case null -> List.empty();
            case OutcomeMessage om -> List.of(om);
            case OutcomeContainer oc -> oc.getOutcome().getValidCodes();
            case Note vn -> {
                String msg = String.format(vn.code().getMessage(), vn.args());
                yield List.of(new OutcomeMessage(vn.code().getCode(), msg, vn.code().isCritical()));
            }
            case ValidCode vc -> List.of(new OutcomeMessage(vc.getCode(), vc.getMessage(), vc.isCritical()));
            case Iterable<?> it -> List.ofAll(it).flatMap(Outcome::flattenNotes);
            default -> throw new IllegalArgumentException("Not accept this object");
        };
    }

    public static Outcome ofErrors(Validation<Seq<String>, ?> validation) {
        return validation.toEither().fold(
                errors -> errors.foldLeft(new Outcome(), (acc, err) -> acc.addValidCode(ValidError.VALID_ERROR_1, err)),
                ignored -> new Outcome()
        );
    }

    public Outcome addValidCode(ValidCode code, Object... args) {
        String formattedMessage = String.format(code.getMessage(), args);
        return this.withValidCodes(validCodes.append(
                new OutcomeMessage(code.getCode(), formattedMessage, code.isCritical())));
    }

    public Outcome addCustomNote(ValidCode code, String text, boolean isCritical) {
        return this.withValidCodes(validCodes.append(new OutcomeMessage(code.getCode(), text, isCritical)));
    }

    public Outcome addNoteIf(boolean condition, ValidCode code, Object... args) {
        return condition ? addValidCode(code, args) : this;
    }

    public Outcome addNoteIf(Either<?, ?> condition, ValidCode code, Object... args) {
        return condition.isRight() ? addValidCode(code, args) : this;
    }

    public Outcome mergeOutcome(Throwable throwable) {
        if (throwable instanceof RequestFailedException ex) {
            return this.addAllNotes(ex.getNotes());
        }
        return this.addValidCode(ValidError.VALID_ERROR_2, throwable.getMessage());
    }

    public Outcome asCriticalOutcome(Throwable throwable) {
        return this.mergeOutcome(throwable);
    }

    public Outcome updateAction(Try<?> result, ValidCode code, Object... args) {
        return result.isSuccess() ? this : addValidCode(code, args);
    }

    public Outcome updateAction(Either<?, ?> result, ValidCode code, Object... args) {
        if (result.isRight()) {
            return this;
        }
        return addValidCode(code, args);
    }

    public Outcome saveAction(Try<?> result, ValidCode code, Object... args) {
        return hasCriticalErrors() ? this : updateAction(result, code, args);
    }

    public Outcome saveAction(Either<Note, ?> result, Object... args) {
        if (hasCriticalErrors()) {
            return this;
        }
        if (result.isLeft()) {
            Note note = result.getLeft();
            Object[] formatArgs = args.length > 0 ? args : note.args();
            return addValidCode(note.code(), formatArgs);
        }
        return this;
    }

    public Outcome saveAction(Either<?, ?> result, ValidCode code, Object... args) {
        return hasCriticalErrors() ? this : updateAction(result, code, args);
    }

    public Outcome performAction(Supplier<?> action) {
        if (hasCriticalErrors()) {
            return this;
        }
        Object result = action.get();
        if (result instanceof OutcomeContainer container) {
            return this.addAllNotes(container.getOutcome().getValidCodes());
        }
        return this;
    }

    public Outcome addAllNotes(Seq<OutcomeMessage> newMessages) {
        return this.withValidCodes(validCodes.appendAll(newMessages));
    }

    @JsonIgnore
    public boolean hasCriticalErrors() {
        return validCodes.exists(OutcomeMessage::critical);
    }

    @JsonIgnore
    public boolean hasNoneCriticalErrors() {
        return !hasCriticalErrors();
    }

    @JsonIgnore
    public boolean isEmpty() {
        return validCodes.isEmpty();
    }

    @JsonIgnore
    public boolean isNotEmpty() {
        return !validCodes.isEmpty();
    }
}
