package ru.lambdahub.realtime.ws;

import org.springframework.stereotype.Component;
import org.springframework.web.socket.TextMessage;
import org.springframework.web.socket.WebSocketSession;

import java.io.IOException;
import java.util.Map;
import java.util.Set;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;

@Component
public class SessionRegistry {

    private final Map<String, WebSocketSession> sessions = new ConcurrentHashMap<>();
    private final Map<String, UUID> sessionUser = new ConcurrentHashMap<>();
    private final Map<String, Set<String>> sessionSubs = new ConcurrentHashMap<>();
    private final Map<String, Set<String>> channelSessions = new ConcurrentHashMap<>();

    public void register(WebSocketSession session, UUID userId) {
        sessions.put(session.getId(), session);
        sessionUser.put(session.getId(), userId);
        sessionSubs.put(session.getId(), ConcurrentHashMap.newKeySet());
    }

    public void remove(WebSocketSession session) {
        Set<String> subs = sessionSubs.remove(session.getId());
        if (subs != null) {
            for (String ch : subs) {
                Set<String> set = channelSessions.get(ch);
                if (set != null) {
                    set.remove(session.getId());
                }
            }
        }
        sessions.remove(session.getId());
        sessionUser.remove(session.getId());
    }

    public UUID userId(WebSocketSession session) {
        return sessionUser.get(session.getId());
    }

    public void subscribe(WebSocketSession session, String channel) {
        sessionSubs.computeIfAbsent(session.getId(), k -> ConcurrentHashMap.newKeySet()).add(channel);
        channelSessions.computeIfAbsent(channel, k -> ConcurrentHashMap.newKeySet()).add(session.getId());
    }

    public void unsubscribe(WebSocketSession session, String channel) {
        Set<String> subs = sessionSubs.get(session.getId());
        if (subs != null) subs.remove(channel);
        Set<String> set = channelSessions.get(channel);
        if (set != null) set.remove(session.getId());
    }

    public void deliver(String channel, String payload) {
        Set<String> sessionIds = channelSessions.get(channel);
        if (sessionIds == null) return;
        for (String sid : sessionIds) {
            WebSocketSession s = sessions.get(sid);
            if (s != null && s.isOpen()) {
                try {
                    synchronized (s) {
                        s.sendMessage(new TextMessage(payload));
                    }
                } catch (IOException ignored) {
                }
            }
        }
    }
}
