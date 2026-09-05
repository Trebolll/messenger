package ru.lambdahub.common.events;

public final class RedisChannels {
    public static String chat(String chatId) {
        return "rt:chat:" + chatId;
    }

    public static String presence(String userId) {
        return "rt:presence:" + userId;
    }

    public static String user(String userId) {
        return "rt:user:" + userId;
    }

    public static final String PATTERN_ALL = "rt:*";

    private RedisChannels() {}
}
