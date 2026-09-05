package ru.lambdahub.common.web;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import org.springframework.web.context.request.ServletWebRequest;
import org.springframework.web.context.request.WebRequest;
import ru.lambdahub.validation.exception.RequestFailedException;
import ru.lambdahub.validation.exception.ValidationException;

@RestControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(IllegalArgumentException.class)
    public ResponseEntity<ApiError> badRequest(IllegalArgumentException ex, WebRequest request) {
        return build(HttpStatus.BAD_REQUEST, ex.getMessage(), request);
    }

    @ExceptionHandler(IllegalStateException.class)
    public ResponseEntity<ApiError> conflict(IllegalStateException ex, WebRequest request) {
        return build(HttpStatus.CONFLICT, ex.getMessage(), request);
    }

    @ExceptionHandler(SecurityException.class)
    public ResponseEntity<ApiError> forbidden(SecurityException ex, WebRequest request) {
        return build(HttpStatus.FORBIDDEN, ex.getMessage(), request);
    }

    @ExceptionHandler(ValidationException.class)
    public ResponseEntity<?> validationManual(ValidationException ex) {
        if (ex.getResponse() != null) {
            return ResponseEntity.status(ex.getStatus()).body(ex.getResponse());
        }
        return ResponseEntity.status(ex.getStatus()).body(ex.getOutcome());
    }

    @ExceptionHandler(RequestFailedException.class)
    public ResponseEntity<?> requestFailed(RequestFailedException ex) {
        int status = ex.getStatus() == 0 ? HttpStatus.BAD_REQUEST.value() : ex.getStatus();
        return ResponseEntity.status(status).body(ex.getNotes());
    }

    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ResponseEntity<ApiError> validation(MethodArgumentNotValidException ex, WebRequest request) {
        String msg = ex.getBindingResult().getFieldErrors().stream()
                .findFirst()
                .map(err -> err.getField() + ": " + err.getDefaultMessage())
                .orElse("Validation failed");
        return build(HttpStatus.BAD_REQUEST, msg, request);
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<ApiError> generic(Exception ex, WebRequest request) {
        return build(HttpStatus.INTERNAL_SERVER_ERROR, ex.getMessage(), request);
    }

    private ResponseEntity<ApiError> build(HttpStatus status, String message, WebRequest request) {
        String path = "";
        if (request instanceof ServletWebRequest swr) {
            path = swr.getRequest().getRequestURI();
        }
        return ResponseEntity.status(status)
                .body(ApiError.of(status.value(), status.getReasonPhrase(), message, path));
    }
}
