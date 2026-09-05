# common-kafka-starter

Spring Cloud Stream starter with `OutputBridge` beans from `kafka-bindings.yml`.

## Usage

1. Depend on `common-kafka-starter`
2. Add `src/main/resources/kafka-stream/kafka-bindings.yml` with bindings
3. Import it:

```yaml
spring:
  config:
    import:
      - optional:classpath:kafka-stream/kafka-bindings.yml
```

4. Inject output bean (name = binding without `-out-N`):

```java
private final OutputBridge userCreatedOutput;

userCreatedOutput.send(event, Map.of(KafkaHeaders.KEY, userId.toString()));
```

5. For inputs, declare a `Consumer` / `Function` bean with the same base name as `*-in-0`.

Build: `mvn -DskipTests package`
