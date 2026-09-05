package ru.lambdahub.validation;

import jakarta.validation.Validator;
import org.springframework.boot.autoconfigure.AutoConfiguration;
import org.springframework.boot.autoconfigure.condition.ConditionalOnClass;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.context.annotation.Bean;
import ru.lambdahub.validation.validator.OutcomeValidator;

@AutoConfiguration
@ConditionalOnClass(Validator.class)
public class OutcomeAutoConfiguration {

    @Bean
    @ConditionalOnMissingBean
    public OutcomeValidator outcomeValidator(Validator validator) {
        return new OutcomeValidator(validator);
    }
}
