# validation-vavr-starter

Starter в стиле `vavr-demo`: `OutcomeValidator` + `Outcome.of(...)`.

```java
@Service
@RequiredArgsConstructor
public class SomeScenarioService {

  private final OutcomeValidator vavrValidator;

  public OutboundMessage process(InboundMessage input) {
    Outcome outcome = Outcome.of(vavrValidator.validate(input))
        .addNoteIf(condition, ValidError.RECORD_NOT_FOUND, id)
        .saveAction(dbEither, ValidError.DB_SAVE_ERROR, id);

    return OutboundMessage.builder()
        .requestId(input.getRequestId())
        .outcome(outcome)
        .build()
        .executeOrThrow();
  }
}
```

Для простых REST DTO:

```java
Outcome outcome = Outcome.of(vavrValidator.validate(request));
if (outcome.hasCriticalErrors()) {
  throw ValidationExceptionUtils.createValidationException(outcome);
}
```
