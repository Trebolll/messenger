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

    // Если пришло сообщение — отправитель точно онлайн.
    // Обновляем его статус в глобальном мапе, чтобы интерфейс сразу отреагировал.
    if (!app.userStatusMap) app.userStatusMap = {};
    app.userStatusMap[String(msg.sender_id)] = { online: true, status: '' };
    
    // Если это приватный чат с этим отправителем — помечаем его онлайн и там
    const directChat = app.chats.find(c => String(c.interlocutor_id) === String(msg.sender_id));
    if (directChat) directChat.is_online = true;

    if (app.activeChatId && String(msg.chat_id) === String(app.activeChatId)) {
        // Проверяем по id — и по множеству уже добавленных нами сообщений
        const already = app.messages.find(m =>
            m.id && msg.id && String(m.id) === String(msg.id)
        );
        const alreadySent = window._sentMsgIds && window._sentMsgIds.has(String(msg.id));
        if (alreadySent) window._sentMsgIds.delete(String(msg.id)); // чистим

        // Если tempMsg ещё ждёт — WS пришёл раньше загрузки файла
        // Запоминаем данные сообщения чтобы file-upload мог применить вложение
        const pendingTemp = app.messages.find(m =>
            m._realMsgId && String(m._realMsgId) === String(msg.id)
        );
        if (pendingTemp) {
            // Не добавляем — file-upload сам заменит tempMsg когда закончит загрузку
            // Сохраняем данные sender/time из WS-сообщения в tempMsg
            pendingTemp._wsMsg = msg;
        } else if (!already && !alreadySent) {
            // Парсим вложение если есть
            if (typeof parseAttachmentFromContent === 'function') {
                const att = parseAttachmentFromContent(msg.content);
                if (att) { msg._attachment = att; msg.content = extractCaptionFromContent(msg.content); }
            }
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
    const uid = String(status.user_id);
    const isOnline = !!status.online;
    
    console.log(`[WS] Status Change: User=${uid} Online=${isOnline}`);

    // Сохраняем в глобальный мап статусов
    if (!app.userStatusMap) app.userStatusMap = {};
    app.userStatusMap[uid] = {
        online: isOnline,
        status: status.status || ''
    };

    // 1. Точечное обновление в списке чатов (левая панель)
    app.chats.forEach(chat => {
        if (chat.interlocutor_id && String(chat.interlocutor_id) === uid) {
            chat.is_online = isOnline;
            chat.user_status = status.status || '';
            const chatEl = document.querySelector(`.chat-list-item[data-chat-id="${chat.id}"]`);
            if (chatEl) {
                const avatarContainer = chatEl.querySelector('.relative.flex-shrink-0');
                if (avatarContainer) {
                    let dot = avatarContainer.querySelector('.online-dot-glass');
                    if (isOnline && !dot) {
                        avatarContainer.insertAdjacentHTML('beforeend', '<div class="online-dot-glass"></div>');
                    } else if (!isOnline && dot) {
                        dot.remove();
                    }
                }
            }
        }
        if (chat.members) {
            chat.members.forEach(m => {
                if (String(m.id) === uid) m.is_online = isOnline;
            });
        }
    });

    // 2. Обновление если этот пользователь сейчас в активном чате
    const activeChat = app.chats.find(c => String(c.id) === String(app.activeChatId));
    if (activeChat) {
        // Обновляем заголовок (если это группа — пересчитываем онлайн, если приват — меняем статус)
        if (activeChat.is_group) {
            const headerStatus = document.getElementById('active-chat-status');
            if (headerStatus) {
                const total = (activeChat.members || []).length;
                const online = (activeChat.members || []).filter(m => {
                    if (String(m.id) === String(app.currentUser?.id)) return true;
                    const globalStatus = app.userStatusMap && app.userStatusMap[String(m.id)];
                    if (globalStatus) return globalStatus.online;
                    return m.is_online;
                }).length;
                headerStatus.textContent = `${online} из ${total} online`;
            }
            
            // Обновляем точку в списке участников (инфо-панель)
            const memberRow = document.querySelector(`.member-row[data-uid="${uid}"]`);
            if (memberRow) {
                const avatarContainer = memberRow.querySelector('.relative.flex-shrink-0');
                if (avatarContainer) {
                    let dot = avatarContainer.querySelector('.member-online-dot');
                    if (isOnline && !dot) {
                        avatarContainer.insertAdjacentHTML('beforeend', '<span class="member-online-dot"></span>');
                    } else if (!isOnline && dot) {
                        dot.remove();
                    }
                }
            }
        } else if (String(activeChat.interlocutor_id) === uid) {
            renderStatusElements(isOnline, status.status || '');
        }
    }
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
    
    // 1. Обновляем кэш сообщений и рейтинг отправителя во всех его сообщениях
    if (app.messages) {
        app.messages.forEach(m => {
            // Обновляем счетчики голосов для конкретного сообщения
            if (String(m.id) === String(data.message_id)) {
                m.likes    = data.likes;
                m.dislikes = data.dislikes;
                if (data.voter_id && String(data.voter_id) === String(app.currentUser?.id)) {
                    m.my_vote = data.my_vote;
                }
            }
            // Обновляем рейтинг отправителя во ВСЕХ его сообщениях в текущем чате
            if (String(m.sender_id) === String(data.sender_id)) {
                m.sender_rating = data.sender_rating;
            }
        });
    }

    // 2. Обновляем рейтинг пользователя в списках участников всех чатов
    if (app.chats) {
        app.chats.forEach(chat => {
            if (chat.members) {
                const member = chat.members.find(m => String(m.id) === String(data.sender_id));
                if (member) {
                    member.rating = data.sender_rating;
                }
            }
        });
    }

    // 3. Точечное обновление DOM для кнопок голосования
    const msg = (app.messages || []).find(m => String(m.id) === String(data.message_id));
    const el = document.querySelector(`[data-msg-id="${data.message_id}"]`);
    if (el) {
        const votesEl = el.querySelector('.msg-votes');
        if (votesEl) {
            const currentMyVote = msg ? msg.my_vote : 0;
            votesEl.innerHTML = renderVotesHtml(data.likes, data.dislikes, currentMyVote, data.message_id);
            votesEl.classList.toggle('has-active', currentMyVote !== 0);
        }

        // Анимация вспышки
        const bubble = el.querySelector('.message-bubble');
        if (bubble && data.just_voted) {
            const isLike = data.just_voted === 1;
            const glowClass = isLike ? 'bubble-glow-like' : 'bubble-glow-dislike';
            const btn = el.querySelector(isLike ? '.vote-like' : '.vote-dislike');
            if (btn) {
                const flashClass = isLike ? 'vote-flash-like' : 'vote-flash-dislike';
                btn.classList.remove('vote-flash-like', 'vote-flash-dislike');
                void btn.offsetWidth;
                btn.classList.add(flashClass);
            }
            bubble.classList.remove('bubble-glow-like', 'bubble-glow-dislike');
            void bubble.offsetWidth;
            bubble.classList.add(glowClass);
            setTimeout(() => bubble.classList.remove('bubble-glow-like', 'bubble-glow-dislike'), 1000);
        }
    }

    // 4. Обновляем бейджи рейтинга отправителя во всех видимых сообщениях
    document.querySelectorAll(`[data-sender-id="${data.sender_id}"]`).forEach(msgEl => {
        const headerLine = msgEl.querySelector('.msg-header-line');
        if (headerLine) {
            const oldBadge = headerLine.querySelector('.msg-rating-badge');
            if (oldBadge) oldBadge.remove();
            
            const badgeHtml = renderRatingBadge(data.sender_rating);
            if (badgeHtml) {
                headerLine.insertAdjacentHTML('beforeend', badgeHtml);
            }
        }
    });

    // 5. Если открыта панель информации о группе и там есть этот пользователь — можно было бы обновить,
    // но обычно это не критично в реальном времени. Если нужно — вызываем renderGroupInfo()
    // if (document.getElementById('info-panel') && !document.getElementById('info-panel').classList.contains('hidden')) {
    //    renderGroupInfo();
    // }
}