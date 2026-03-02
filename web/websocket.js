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
            } else if (wrapper.type === 'vote_updated') {
                _handleVoteUpdated(wrapper.content);
            } else if (wrapper.type === 'user_profile_updated') {
                updateUserProfile(wrapper.content);
            } else if (wrapper.type === 'member_added') {
                _handleMemberAdded(wrapper.content);
            } else if (wrapper.type === 'member_removed') {
                _handleMemberRemoved(wrapper.content);
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
    } else {
        // Подсвечиваем чат в списке только если сообщение не от меня
        if (String(msg.sender_id) !== String(app.currentUser?.id)) {
            if (!app._unreadHighlight) app._unreadHighlight = new Set();
            app._unreadHighlight.add(String(msg.chat_id));
            _applyUnreadHighlight(String(msg.chat_id));
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
    const el    = document.querySelector('[data-msg-id="' + msgId + '"]');

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
    // Обновляем онлайн-статус в членах групп
    app.chats.forEach(c => {
        if (c.members) {
            c.members.forEach(m => {
                if (String(m.id) === String(status.user_id)) m.is_online = status.online;
            });
        }
    });
    const activeChat = app.chats.find(c => String(c.id) === String(app.activeChatId));
    if (activeChat && String(activeChat.interlocutor_id) === String(status.user_id)) {
        renderStatusElements(status.online, status.status || '');
    }
    // Обновляем индикаторы онлайн в сообщениях без перерисовки всего списка
    const dots = document.querySelectorAll(`[data-online-uid="${status.user_id}"]`);
    dots.forEach(dot => {
        dot.style.background = status.online ? '#22c55e' : '#9ca3af';
    });
}

// ─── Подсветка чата при новом входящем сообщении ────────────────────────────

function _applyUnreadHighlight(chatId) {
    const el = document.querySelector(`.chat-list-item[data-chat-id="${chatId}"]`);
    if (el) {
        el.classList.add('chat-unread-flash');
    }
}

function _clearUnreadHighlight(chatId) {
    const app = window.app;
    if (app._unreadHighlight) app._unreadHighlight.delete(String(chatId));
    const el = document.querySelector(`.chat-list-item[data-chat-id="${chatId}"]`);
    if (el) el.classList.remove('chat-unread-flash');
}



function _handleMemberAdded(data) {
    const app  = window.app;
    const myId = app.currentUser && app.currentUser.id;

    console.log('[member_added] data:', JSON.stringify(data), '| myId:', myId);

    const chat = app.chats.find(c => String(c.id) === String(data.chat_id));

    if (!chat) {
        // Чат не найден локально — нас только что добавили в эту группу
        apiLoadChats().then(function() {
            renderChats();
            app.notify('Вас добавили в группу', 'success');
        });
        return;
    }

    // Добавляем нового участника в локальный список если его ещё нет
    if (chat.members && data.user) {
        const already = chat.members.find(m => String(m.id) === String(data.user.id));
        if (!already) {
            chat.members.push(Object.assign({}, data.user, { is_online: false }));
        }
    }

    // Обновляем UI если этот чат сейчас открыт
    if (String(app.activeChatId) === String(data.chat_id)) {
        renderChatHeader();
        if (data.user && String(data.user.id) !== String(myId)) {
            app.notify((data.user.username || 'Участник') + ' добавлен(а) в группу', 'success');
        }
    }
}

function _handleMemberRemoved(data) {
    const app       = window.app;
    const myId      = app.currentUser && app.currentUser.id;
    const removedId = data.user_id;

    // DEBUG — убрать после проверки
    console.log('[member_removed] data:', JSON.stringify(data));
    console.log('[member_removed] myId:', myId, '| removedId:', removedId, '| match:', String(removedId) === String(myId));
    console.log('[member_removed] all chat ids:', app.chats.map(c => c.id));

    // Если удалили нас самих — убираем чат из списка и закрываем окно
    if (myId && String(removedId) === String(myId)) {
        app.chats = app.chats.filter(c => String(c.id) !== String(data.chat_id));

        if (String(app.activeChatId) === String(data.chat_id)) {
            app.activeChatId = null;
            app.messages = [];

            var msgContainer = document.getElementById('messages-container');
            if (msgContainer) msgContainer.innerHTML = '';

            var infoPanel = document.getElementById('info-panel');
            if (infoPanel) infoPanel.classList.add('hidden');

            var noChatSelected = document.getElementById('no-chat-selected');
            if (noChatSelected) noChatSelected.classList.remove('hidden');
        }

        renderChats();
        app.notify('Вас удалили из группы', 'error');
        return;
    }

    // Нас не удалили — убираем участника из локального списка
    var chat = app.chats.find(c => String(c.id) === String(data.chat_id));
    if (chat && chat.members) {
        chat.members = chat.members.filter(m => String(m.id) !== String(removedId));
    }

    if (chat && String(app.activeChatId) === String(data.chat_id)) {
        renderChatHeader();
    }
}
// ─── Обновление голосов в реальном времени ───────────────────────────────────

function _handleVoteUpdated(data) {
    const app = window.app;
    const msg = (app.messages || []).find(m => String(m.id) === String(data.message_id));
    if (msg) {
        msg.likes    = data.likes;
        msg.dislikes = data.dislikes;
        msg.my_vote  = data.my_vote;
    }
    // Точечное обновление DOM
    const el = document.querySelector(`[data-msg-id="${data.message_id}"]`);
    if (el) {
        const votesEl = el.querySelector('.msg-votes');
        if (votesEl) {
            votesEl.innerHTML = renderVotesHtml(data.likes, data.dislikes, data.my_vote, data.message_id);
            // has-active — фолбэк для браузеров без :has()
            votesEl.classList.toggle('has-active', data.my_vote !== 0);
        }
    }
    // Обновляем бейджи рейтинга отправителя
    document.querySelectorAll(`[data-sender-id="${data.sender_id}"] .msg-rating-badge`).forEach(b => {
        b.outerHTML = renderRatingBadge(data.sender_rating);
    });
}