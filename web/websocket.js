// ─── websocket.js — подключение и обработка WebSocket ─────────────────────

function connectWebSocket() {
    const app = window.app;
    if (app.socket) app.socket.close();

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/ws?token=${app.token}`;
    console.log('Connecting to WebSocket:', wsUrl);

    app.socket = new WebSocket(wsUrl);

    app.socket.onopen = () => {
        console.log('WebSocket connected ✅');
        app.notify('Соединение установлено', 'success');
    };

    app.socket.onmessage = (event) => {
        console.log('WS Message received:', event.data);
        try {
            const wrapper = JSON.parse(event.data);

            if (wrapper.type === 'new_message') {
                _handleNewMessage(wrapper.content);
            } else if (wrapper.type === 'messages_read') {
                _handleMessagesRead(wrapper.content);
            } else if (wrapper.type === 'message_edited') {
                _handleMessageEdited(wrapper.content);
            } else if (wrapper.type === 'message_deleted') {
                _handleMessageDeleted(wrapper.content);
            } else if (wrapper.type === 'user_status') {
                updateUserStatus(wrapper.content);
            } else if (wrapper.type === 'user_profile_updated') {
                updateUserProfile(wrapper.content);
            }
        } catch (err) {
            console.error('Error parsing WS message:', err);
        }
    };

    app.socket.onclose = (e) => {
        console.log('WebSocket closed ❌', e.reason);
        setTimeout(() => {
            if (app.token) connectWebSocket();
        }, 3000);
    };

    app.socket.onerror = (err) => {
        console.error('WebSocket error ⚠️', err);
    };
}

function _handleNewMessage(msg) {
    const app = window.app;
    console.log('New message:', msg, '| activeChatId:', app.activeChatId);

    if (app.activeChatId && String(msg.chat_id) === String(app.activeChatId)) {
        // Защита от дублей: сравниваем id, или контент+отправитель+время
        const already = app.messages.find(m =>
            (m.id && msg.id && String(m.id) === String(msg.id)) ||
            (m.content === msg.content &&
                String(m.sender_id) === String(msg.sender_id) &&
                m.created_at === msg.created_at)
        );
        if (!already) {
            app.messages.push(msg);
            renderMessages();
            scrollToBottom();
        }
        if (String(msg.sender_id) !== String(app.currentUser?.id)) {
            apiMarkChatAsRead(app.activeChatId);
        }
    }
    updateLastMessageInChatList(msg);
}

function _handleMessagesRead(data) {
    const app = window.app;
    if (app.activeChatId && String(data.chat_id) === String(app.activeChatId)) {
        app.messages = app.messages.map(m => {
            if (String(m.sender_id) !== String(data.reader_id)) {
                return { ...m, read_at: new Date().toISOString() };
            }
            return m;
        });
        renderMessages();
    }
}

function _handleMessageEdited(msg) {
    const app = window.app;
    const idx = app.messages.findIndex(m => String(m.id) === String(msg.id));
    if (idx !== -1) {
        app.messages[idx] = msg;
        renderMessages();
    }
}

function _handleMessageDeleted(data) {
    const app   = window.app;
    const msgId = String(data.message_id);
    const el    = document.querySelector(`[data-msg-id="${msgId}"]`);

    if (el) {
        ashDisintegrate(el, () => {
            app.messages = app.messages.filter(m => String(m.id) !== msgId);
            el.remove();
        });
    } else {
        app.messages = app.messages.filter(m => String(m.id) !== msgId);
    }
}

function updateUserProfile(data) {
    const app = window.app;

    // Обновляем все чаты где этот пользователь — собеседник
    let changed = false;
    app.chats.forEach(chat => {
        if (String(chat.interlocutor_id) === String(data.user_id)) {
            chat.avatar_url = data.avatar_url;
            chat.name       = data.username || chat.name;
            changed = true;
        }
    });
    if (changed) {
        renderChats();
        renderChatHeader();
    }
}

function updateUserStatus(status) {
    const app = window.app;
    const chat = app.chats.find(c => String(c.interlocutor_id) === String(status.user_id));
    if (chat) {
        chat.is_online   = status.online;
        chat.user_status = status.status || '';
        renderChats();
    }
    const activeChat = app.chats.find(c => String(c.id) === String(app.activeChatId));
    if (activeChat && String(activeChat.interlocutor_id) === String(status.user_id)) {
        renderStatusElements(status.online, status.status || '');
    }
}