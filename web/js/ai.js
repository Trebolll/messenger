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
    const prevChatId = window.app.activeChatId;
    window.app.activeChatId = chatId;
    if (typeof _clearUnreadHighlight === 'function') _clearUnreadHighlight(chatId);

    const container = document.getElementById('messages-container');
    const header    = document.getElementById('chat-header');
    const noChat    = document.getElementById('no-chat-selected');
    const inputArea = document.getElementById('input-area');

    // ── 1. Fade-out текущего контента ──────────────────────────
    // Закрываем AI панель при переходе в другой чат
    if (typeof window.toggleAiPanel === 'function') {
        const panel = document.getElementById('ai-panel');
        if (panel && !panel.classList.contains('hidden')) {
            window.toggleAiPanel();
        }
    }

    const isSwitch = !!prevChatId && prevChatId !== chatId;
    if (isSwitch && container) {
        container.classList.add('chat-switching');
        if (header) header.classList.add('chat-switching');
        await new Promise(r => setTimeout(r, 180));
    }

    // ── 2. Показываем UI чата (если первый раз) ────────────────
    if (noChat)    noChat.classList.add('hidden');
    if (inputArea) { inputArea.classList.remove('hidden'); requestAnimationFrame(() => inputArea.classList.add('visible')); }
    if (typeof switchView === 'function') switchView('chat');

    // ── 3. Skeleton пока грузим ────────────────────────────────
    if (container) {
        container.classList.remove('chat-switching', 'chat-loaded');
        container.innerHTML = _buildSkeleton();
    }
    if (header) {
        header.classList.remove('chat-switching');
        renderChatHeader();
    }

    // ── 4. Запрос ──────────────────────────────────────────────
    try {
        const res = await apiFetch(`/api/chats/${chatId}/messages`);
        // Проверяем что пользователь не переключился пока грузили
        if (window.app.activeChatId !== chatId) return;
        window.app.messages = res || [];

        // ── 5. Fade-in сообщений ───────────────────────────────
        renderMessages();
        scrollToBottom(true); // instant — прячем → скроллим → показываем
        if (container) {
            container.classList.add('chat-loaded');
            setTimeout(() => container.classList.remove('chat-loaded'), 600);
        }
        await apiMarkChatAsRead(chatId);
    } catch (err) {
        if (container) container.innerHTML = '';
        window.app.notify('Ошибка загрузки сообщений', 'error');
    }
}

// Skeleton — 6 рандомных пузырей разной ширины
function _buildSkeleton() {
    const rows = [
        { me: false, w: 180 }, { me: true,  w: 220 },
        { me: false, w: 140 }, { me: true,  w: 160 },
        { me: false, w: 200 }, { me: true,  w: 110 },
    ];
    return rows.map(function(r, i) {
        const avatarHtml = r.me ? '' : '<div class="skeleton-avatar"></div>';
        const msgClass   = 'skeleton-msg' + (r.me ? ' me' : '');
        const rowDelay   = 'animation-delay:' + (i * 40) + 'ms';
        const bubStyle   = 'width:' + r.w + 'px; animation-delay:' + (i * 60) + 'ms';
        return '<div class="' + msgClass + '" style="' + rowDelay + '">'
            + avatarHtml
            + '<div class="skeleton-bubble" style="' + bubStyle + '"></div>'
            + '</div>';
    }).join('');
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
    // Убираем лишние переносы: несколько \n подряд → один, trim по краям
    const text = input.value.replace(/\n{2,}/g, '\n').trim();
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

// ─── File Upload API ────────────────────────────────────────────────────────

async function apiUploadFile(chatId, file, messageId, onProgress) {
    const token = localStorage.getItem('alpha_token');
    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();
        const formData = new FormData();
        formData.append('file', file);
        if (messageId) formData.append('message_id', messageId);

        xhr.upload.addEventListener('progress', (e) => {
            if (e.lengthComputable && onProgress) {
                onProgress(e.loaded, e.total);
            }
        });

        xhr.addEventListener('load', () => {
            if (xhr.status >= 200 && xhr.status < 300) {
                try { resolve(JSON.parse(xhr.responseText)); }
                catch { reject(new Error('Invalid response')); }
            } else {
                try {
                    const err = JSON.parse(xhr.responseText);
                    reject(new Error(err.error || 'Upload failed'));
                } catch { reject(new Error('Upload failed')); }
            }
        });
        xhr.addEventListener('error', () => reject(new Error('Network error')));

        xhr.open('POST', `/api/chats/${chatId}/attachments`);
        if (token) xhr.setRequestHeader('Authorization', 'Bearer ' + token);
        xhr.send(formData);
    });
}
async function apiVoteMessage(messageId, vote) {
    return await apiFetch(`/api/messages/${messageId}/vote`, {
        method: 'POST',
        body: JSON.stringify({ vote }),
    });
}