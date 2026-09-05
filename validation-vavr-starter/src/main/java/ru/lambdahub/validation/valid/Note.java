package ru.lambdahub.validation.valid;

import lombok.Builder;
import lombok.experimental.FieldNameConstants;

@Builder
@FieldNameConstants
public record Note(ValidCode code, Object[] args) {}
