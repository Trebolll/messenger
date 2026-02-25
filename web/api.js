// ─── api.js — все запросы к серверу ───────────────────────────────────────

async function apiFetch(url, options = {}) {
    const token = window.app?.token;
    const headers = {
        'Content-Type': 'application/json',
        ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
        ...options.headers,
    };
    const response = await fetch(url, { ...options, headers });
    if (response.status === 401) window.app?.logout();
    const result = await response.json();
    if (!response.ok) throw new Error(result.error || 'Ошибка запроса');
    return result;
}

async function apiLoadChats() {
    try {
        const res = await apiFetch('/api/chats');
        window.app.chats = res || [];
        renderChats();
    } catch (err) {
        window.app.notify('Ошибка загрузки чатов', 'error');
    }
}

async function apiLoadMessages(chatId) {
    window.app.activeChatId = chatId;
    document.getElementById('no-chat-selected').classList.add('hidden');
    renderChatHeader();
    try {
        const res = await apiFetch(`/api/chats/${chatId}/messages`);
        window.app.messages = res || [];
        renderMessages();
        scrollToBottom();
        await apiMarkChatAsRead(chatId);
    } catch (err) {
        window.app.notify('Ошибка загрузки сообщений', 'error');
    }
}

async function apiSearchUsers(query) {
    if (!query || query.length < 2) {
        document.getElementById('search-results').innerHTML = '';
        return;
    }
    try {
        const res = await apiFetch(`/api/users/search?q=${encodeURIComponent(query)}`);
        renderSearchResults(res || []);
    } catch (err) { /* silent */ }
}

async function apiCreatePrivateChat(userId) {
    try {
        const res = await apiFetch('/api/chats/private', {
            method: 'POST',
            body: JSON.stringify({ user_id: userId }),
        });
        closeNewChatModal();
        await apiLoadChats();
        await apiLoadMessages(res.id);
    } catch (err) {
        window.app.notify(err.message, 'error');
    }
}

async function apiSendMessage() {
    const input = document.getElementById('message-input');
    const text = input.value.trim();
    if (!text || !window.app.activeChatId) return;
    input.value = '';
    try {
        const msg = await apiFetch('/api/messages', {
            method: 'POST',
            body: JSON.stringify({ chat_id: window.app.activeChatId, content: text }),
        });
        // Добавляем только если ещё нет (WS может прийти раньше REST-ответа)
        const already = window.app.messages.find(m =>
            (m.id && msg.id && String(m.id) === String(msg.id))
        );
        if (!already) {
            window.app.messages.push(msg);
            renderMessages();
            scrollToBottom();
        }
    } catch (err) {
        window.app.notify('Не удалось отправить сообщение', 'error');
    }
}

async function apiMarkChatAsRead(chatId) {
    try {
        await apiFetch(`/api/chats/${chatId}/read`, { method: 'POST' });
    } catch (err) {
        console.error('Error marking chat as read:', err);
    }
}

async function apiSaveProfile({ fullname, phone, username, statusText }) {
    // Сохраняем статус
    await apiFetch('/api/users/status', {
        method: 'PUT',
        body: JSON.stringify({ status: statusText }),
    });
    if (window.app.currentUser) {
        window.app.currentUser.status = statusText;
    }
    // Сохраняем профиль
    const profileBody = {};
    if (fullname)  profileBody.full_name = fullname;
    if (phone)     profileBody.phone     = phone;
    if (username)  profileBody.username  = username;
    if (Object.keys(profileBody).length > 0) {
        const updated = await apiFetch('/api/users/profile', {
            method: 'PUT',
            body: JSON.stringify(profileBody),
        });
        window.app.currentUser = { ...window.app.currentUser, ...updated };
    }
    window.app.currentUser.status = statusText;
    localStorage.setItem('alpha_user', JSON.stringify(window.app.currentUser));
    // Обновляем сайдбар сразу
    loadUserData();
}

async function apiEditMessage(messageId, content) {
    return await apiFetch(`/api/messages/${messageId}`, {
        method: 'PUT',
        body: JSON.stringify({ content }),
    });
}

async function apiDeleteMessage(messageId) {
    return await apiFetch(`/api/messages/${messageId}`, {
        method: 'DELETE',
    });
}