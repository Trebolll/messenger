package ru.lambdahub.common.redis.cache;

import org.springframework.beans.BeanUtils;
import org.springframework.cache.interceptor.SimpleKey;
import org.springframework.core.convert.converter.Converter;
import org.springframework.data.redis.serializer.SerializationException;
import tools.jackson.core.JacksonException;
import tools.jackson.databind.JsonNode;
import tools.jackson.databind.ObjectMapper;
import tools.jackson.databind.node.ArrayNode;
import tools.jackson.databind.node.ObjectNode;

public class CanonicalJsonCacheKeyConverter implements Converter<Object, String> {

    private final ObjectMapper objectMapper;

    public CanonicalJsonCacheKeyConverter(ObjectMapper objectMapper) {
        this.objectMapper = objectMapper.rebuild().deactivateDefaultTyping().build();
    }

    @Override
    public String convert(Object source) {
        if (source == null) {
            throw new SerializationException("Redis cache key must not be null");
        }
        if (BeanUtils.isSimpleValueType(source.getClass()) || source instanceof SimpleKey) {
            return source.toString();
        }

        JsonNode json = objectMapper.valueToTree(source);
        if (json == null || json.isNull()) {
            throw new SerializationException("Redis cache key must not be null");
        }

        try {
            return objectMapper.writeValueAsString(canonicalize(json));
        } catch (JacksonException exception) {
            throw new SerializationException("Failed to serialize Redis cache key", exception);
        }
    }

    private JsonNode canonicalize(JsonNode json) {
        if (json.isObject()) {
            ObjectNode result = objectMapper.createObjectNode();
            json.propertyNames().stream()
                    .sorted()
                    .filter(field -> !json.get(field).isNull())
                    .forEach(field -> result.set(field, canonicalize(json.get(field))));
            return result;
        }
        if (json.isArray()) {
            ArrayNode result = objectMapper.createArrayNode();
            json.forEach(item -> result.add(canonicalize(item)));
            return result;
        }
        return json;
    }
}
