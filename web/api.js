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
    // Снимаем подсветку непрочитанных при входе в чат
    if (typeof _clearUnreadHighlight === 'function') _clearUnreadHighlight(chatId);
    document.getElementById('no-chat-selected').classList.add('hidden');
    const inputArea = document.getElementById('input-area');
    inputArea.classList.remove('hidden');
    requestAnimationFrame(() => inputArea.classList.add('visible'));
    if (typeof switchView === 'function') switchView('chat');
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
    if (!query || query.length < 3) {
        document.getElementById('search-results').innerHTML = '';
        return;
    }
    try {
        const res = await apiFetch(`/api/users/search?q=${encodeURIComponent(query)}`);
        renderSearchResults(res || []);
    } catch (err) { /* silent */ }
}

async function apiCreateGroupChat(userIds, groupName) {
    try {
        // Бэкенд ожидает usernames[] — собираем из Map выбранных пользователей
        const userMap = window._ncUserMap || {};
        const usernames = userIds.map(id => {
            const u = userMap[String(id)];
            return u ? u.username : null;
        }).filter(Boolean);

        const res = await apiFetch('/api/chats/group', {
            method: 'POST',
            body: JSON.stringify({
                name: groupName || usernames.join(', '),
                usernames,
            }),
        });
        closeNewChatModal();
        await apiLoadChats();
        await apiLoadMessages(res.id);
    } catch (err) {
        window.app.notify(err.message || 'Не удалось создать группу', 'error');
    }
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
    document.dispatchEvent(new Event('_msgSent'));
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
async function apiUploadAvatar(file) {
    const formData = new FormData();
    formData.append('avatar', file);
    // Для FormData НЕ ставим Content-Type — браузер сам выставит multipart/form-data
    // Токен берём из window.app.token как все остальные запросы
    const token = window.app?.token;
    const res = await fetch('/api/users/avatar', {
        method: 'PUT',
        headers: { ...(token ? { 'Authorization': `Bearer ${token}` } : {}) },
        body: formData,
    });
    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || 'Ошибка загрузки');
    }
    return res.json();
}

async function apiUploadGroupAvatar(chatId, file) {
    const formData = new FormData();
    formData.append('avatar', file);
    const token = window.app?.token;
    const res = await fetch(`/api/chats/${chatId}/avatar`, {
        method: 'PUT',
        headers: { ...(token ? { 'Authorization': `Bearer ${token}` } : {}) },
        body: formData,
    });
    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || 'Ошибка загрузки');
    }
    return res.json();
}
// ── Group Management API ───────────────────────────────────────────────────

async function apiUpdateGroupInfo(chatId, name) {
    return await apiFetch(`/api/chats/${chatId}`, {
        method: 'PUT',
        body: JSON.stringify({ name }),
    });
}

async function apiRemoveChatMember(chatId, userId) {
    return await apiFetch(`/api/chats/${chatId}/members/${userId}`, { method: 'DELETE' });
}

async function apiAddChatMember(chatId, username) {
    return await apiFetch(`/api/chats/${chatId}/members`, {
        method: 'POST',
        body: JSON.stringify({ username }),
    });
}

async function apiGetGroupMembers(chatId) {
    return await apiFetch(`/api/chats/${chatId}/members`);
}
async function apiVoteMessage(messageId, vote) {
    return await apiFetch(`/api/messages/${messageId}/vote`, {
        method: 'POST',
        body: JSON.stringify({ vote }),
    });
}
