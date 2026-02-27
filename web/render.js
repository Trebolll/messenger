// Хелперы аватаров — избегаем вложенных backtick в template literals
function chatAvatarHtml(chat) {
    if (chat.avatar_url) {
        return '<img src="' + chat.avatar_url + '" style="width:100%;height:100%;object-fit:cover;">';
    }
    return (chat.name || '?')[0].toUpperCase();
}

function userAvatarHtml(user) {
    if (user && user.avatar_url) {
        return '<img src="' + user.avatar_url + '" style="width:100%;height:100%;object-fit:cover;border-radius:50%;">';
    }
    return (user && user.username ? user.username[0].toUpperCase() : 'U');
}

// ─── render.js — отрисовка интерфейса ─────────────────────────────────────

function renderChats() {
    const app  = window.app;
    const list = document.getElementById('chats-list');
    list.innerHTML = app.chats.map(chat => `
        <div onclick="app.loadMessages('${chat.id}')" class="chat-list-item p-4 flex items-center gap-3 transition ${String(app.activeChatId) === String(chat.id) ? 'active' : ''}">
            <div class="relative">
                <div class="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold overflow-hidden">
                    ${chatAvatarHtml(chat)}
                </div>
                ${chat.is_online ? '<div class="absolute bottom-0 right-0 w-3 h-3 bg-green-500 border-2 border-white rounded-full"></div>' : ''}
            </div>
            <div class="flex-grow overflow-hidden">
                <div class="flex justify-between items-baseline">
                    <h4 class="font-bold text-custom-main truncate">${chat.name || 'Chat'}</h4>
                    <span class="text-[10px] text-custom-muted">12:45</span>
                </div>
                <p class="text-xs text-custom-muted truncate">${chat.last_message || 'Нет сообщений'}</p>
            </div>
        </div>
    `).join('');
}

function renderMessages() {
    const app       = window.app;
    const container = document.getElementById('messages-container');

    // Запоминаем какие id уже отрисованы чтобы анимировать только новые
    const existing = new Set(
        [...container.querySelectorAll('[data-msg-id]')].map(el => el.dataset.msgId)
    );

    container.innerHTML = app.messages.map((msg, idx) => {
        const isMe     = String(msg.sender_id) === String(app.currentUser?.id);
        const isRead   = msg.read_at   != null;
        const isEdited = msg.edited_at != null;
        const isNew    = !existing.has(String(msg.id));

        // Задержка каскадом только при первичной загрузке (все новые)
        const delay = (existing.size === 0 && isNew)
            ? `animation-delay:${Math.min(idx * 35, 400)}ms`
            : '';

        const animClass = isNew ? (isMe ? 'msg-anim-sent' : 'msg-anim-received') : '';

        return `
            <div class="flex ${isMe ? 'justify-end' : 'justify-start'} ${animClass}"
                 style="${delay}"
                 data-msg-id="${msg.id}"
                 oncontextmenu="showMessageMenu(event, '${msg.id}', ${isMe})">
                <div class="message-bubble p-3.5 ${isMe ? 'message-sent' : 'message-received'}">
                    <p class="text-sm leading-relaxed" id="msg-content-${msg.id}">${escapeHtml(msg.content)}</p>
                    <div class="flex items-center justify-end gap-1.5 mt-1.5">
                        ${isEdited ? `<span class="msg-edited-label">изменено</span>` : ''}
                        <span class="text-[10px] ${isMe ? 'opacity-60' : 'text-custom-muted'}">
                            ${new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                        </span>
                        ${isMe ? `
                            <span class="text-[11px] ${isRead ? 'opacity-90' : 'opacity-40'}" style="letter-spacing:-0.5em;display:inline-block;">
                                ${isRead ? '✓✓' : '✓'}
                            </span>
                        ` : ''}
                    </div>
                </div>
            </div>
        `;
    }).join('');
}

function renderChatHeader() {
    const app  = window.app;
    const chat = app.chats.find(c => String(c.id) === String(app.activeChatId));
    if (!chat) return;

    renderStatusElements(chat.is_online, chat.user_status || '');

    document.getElementById('active-chat-name').textContent   = chat.name;
    const _ava = chatAvatarHtml(chat);
    const _hdr = document.getElementById('active-chat-avatar');
    _hdr.innerHTML = _ava; _hdr.style.overflow = 'hidden';
    document.getElementById('info-name').textContent = chat.name;
    const _inf = document.getElementById('info-avatar');
    _inf.innerHTML = _ava; _inf.style.overflow = 'hidden';
}

function renderStatusElements(isOnline, userStatus) {
    // Бейдж онлайн/офлайн в правой панели
    const badge = document.getElementById('info-status-badge');
    const text  = document.getElementById('info-status-text');
    if (badge) badge.className   = 'info-status-badge ' + (isOnline ? 'online' : 'offline');
    if (text)  text.textContent  = isOnline ? 'онлайн' : 'офлайн';

    // Текстовый статус — всегда показываем, пустой если нет
    const statusEl = document.getElementById('info-user-status');
    if (statusEl) statusEl.textContent = userStatus ? `«${userStatus}»` : '';

    // Статус под именем в хедере чата
    const headerStatus = document.getElementById('active-chat-status');
    if (headerStatus) {
        headerStatus.textContent = isOnline ? 'онлайн' : 'офлайн';
        headerStatus.className   = 'text-xs ' + (isOnline ? 'text-green-500' : 'text-custom-muted');
    }
}

function renderSearchResults(users) {
    const container = document.getElementById('search-results');
    container.innerHTML = users.map(user => `
        <div onclick="app.createPrivateChat('${user.id}')" class="p-3 hover:bg-custom-sidebar rounded-xl cursor-pointer flex items-center gap-3 transition border border-transparent hover:border-custom">
            <div class="w-10 h-10 rounded-full bg-blue-50 flex items-center justify-center text-blue-600 font-bold">${user.username[0].toUpperCase()}</div>
            <div>
                <div class="font-bold text-custom-main text-sm">${user.username}</div>
                <div class="text-xs text-custom-muted">${user.email}</div>
            </div>
        </div>
    `).join('');
}

function loadUserData() {
    const user = window.app?.currentUser || JSON.parse(localStorage.getItem('alpha_user') || 'null');
    if (!user) return;
    document.getElementById('current-user-name').textContent = user.username;
    setAvatarEl(document.getElementById('current-user-avatar'), user);
}

// Универсальная функция: ставит фото или букву в элемент-аватар
function setAvatarEl(el, user) {
    if (!el) return;
    if (user.avatar_url) {
        el.innerHTML = `<img src="${user.avatar_url}" alt="${user.username}"
            style="width:100%;height:100%;object-fit:cover;border-radius:50%;">`;
    } else {
        el.textContent = (user.username || 'U')[0].toUpperCase();
    }
}

function updateLastMessageInChatList(msg) {
    const app  = window.app;
    const chat = app.chats.find(c => String(c.id) === String(msg.chat_id));
    if (chat) {
        chat.last_message = msg.content;
        renderChats();
    } else {
        apiLoadChats();
    }
}

// ─── Helpers ───────────────────────────────────────────────────────────────

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function scrollToBottom() {
    const container = document.getElementById('messages-container');
    setTimeout(() => { container.scrollTop = container.scrollHeight; }, 50);
}