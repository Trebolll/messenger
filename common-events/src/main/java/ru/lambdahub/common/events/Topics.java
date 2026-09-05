package ru.lambdahub.common.events;

public final class Topics {
    public static final String USER_CREATED = "user.created";
    public static final String USER_PROFILE_UPDATED = "user.profile_updated";
    public static final String USER_STATUS_CHANGED = "user.status_changed";
    public static final String CHAT_CREATED = "chat.created";
    public static final String CHAT_MEMBER_ADDED = "chat.member_added";
    public static final String CHAT_MEMBER_REMOVED = "chat.member_removed";
    public static final String MESSAGE_CREATED = "message.created";
    public static final String MESSAGE_EDITED = "message.edited";
    public static final String MESSAGE_DELETED = "message.deleted";
    public static final String MESSAGES_READ = "messages.read";
    public static final String ATTACHMENT_CREATED = "attachment.created";

    private Topics() {}
}
