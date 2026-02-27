// Inline стек аватарок для групп (в хедере и info-panel)
function groupAvatarStackInline(members, size) {
    const shown = members.slice(0, 4);
    const offset = Math.round(size * 0.55);
    const total = shown.length * size - (shown.length - 1) * (size - offset);
    let html = `<div style="position:relative;width:${total}px;height:${size}px;flex-shrink:0;">`;
    shown.forEach((m, i) => {
        const letter = (m.username || '?')[0].toUpperCase();
        const img = m.avatar_url
            ? `<img src="${m.avatar_url}" style="width:100%;height:100%;object-fit:cover;border-radius:50%;">`
            : letter;
        html += `<div style="position:absolute;left:${i * offset}px;top:0;width:${size}px;height:${size}px;
            border-radius:50%;border:2px solid var(--bg-main);background:#dbeafe;
            display:flex;align-items:center;justify-content:center;font-weight:700;
            font-size:${Math.round(size * 0.38)}px;color:#2563eb;overflow:hidden;z-index:${shown.length - i};">${img}</div>`;
    });
    html += '</div>';
    return html;
}

// Helpers для отображения имени чата
function chatDisplayName(chat) {
    if (chat.name) return chat.name;
    if (chat.is_group && chat.members && chat.members.length) {
        return chat.members.map(m => m.username).join(', ');
    }
    return chat.other_user_name || chat.partner_name || 'Чат';
}

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
    list.innerHTML = app.chats.map(chat => {
        const isActive = String(app.activeChatId) === String(chat.id);
        const displayName = chatDisplayName(chat);
        const lastMsg = chat.last_message || 'Нет сообщений';

        // Групповой чат — аватарка + название (как приватный чат)
        if (chat.is_group) {
            const groupLetter = (chat.name || 'G')[0].toUpperCase();
            const groupAvatarInner = chat.avatar_url
                ? `<img src="${chat.avatar_url}" style="width:100%;height:100%;object-fit:cover;">`
                : groupLetter;

            return `<div onclick="app.loadMessages('${chat.id}')" class="chat-list-item p-4 flex items-center gap-3 transition ${isActive ? 'active' : ''}">
                <div class="relative flex-shrink-0">
                    <div class="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold overflow-hidden">
                        ${groupAvatarInner}
                    </div>
                </div>
                <div class="flex-grow overflow-hidden">
                    <div class="flex justify-between items-baseline">
                        <h4 class="font-bold text-custom-main truncate">${displayName}</h4>
                        <span class="text-[10px] text-custom-muted">группа</span>
                    </div>
                    <p class="text-xs text-custom-muted truncate">${lastMsg}</p>
                </div>
            </div>`;
        }

        // Приватный чат
        const avatarHtml = `<div class="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold overflow-hidden">
            ${chatAvatarHtml(chat)}
        </div>
        ${chat.is_online ? '<div class="absolute bottom-0 right-0 w-3 h-3 bg-green-500 border-2 border-white rounded-full"></div>' : ''}`;

        return `<div onclick="app.loadMessages('${chat.id}')" class="chat-list-item p-4 flex items-center gap-3 transition ${isActive ? 'active' : ''}">
            <div class="relative flex-shrink-0">${avatarHtml}</div>
            <div class="flex-grow overflow-hidden">
                <div class="flex justify-between items-baseline">
                    <h4 class="font-bold text-custom-main truncate">${displayName}</h4>
                    <span class="text-[10px] text-custom-muted">12:45</span>
                </div>
                <p class="text-xs text-custom-muted truncate">${lastMsg}</p>
            </div>
        </div>`;
    }).join('');
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

    const displayName = chatDisplayName(chat);
    document.getElementById('active-chat-name').textContent = displayName;
    document.getElementById('info-name').textContent = displayName;

    // Групповой чат — стек аватарок в хедере
    const _hdr = document.getElementById('active-chat-avatar');
    const _inf = document.getElementById('info-avatar');
    if (chat.is_group && chat.members && chat.members.length > 1) {
        if (chat.avatar_url) {
            // Есть своя аватарка группы
            const groupImgHtml = `<img src="${chat.avatar_url}" style="width:100%;height:100%;object-fit:cover;">`;
            _hdr.innerHTML = groupImgHtml;
            _hdr.style.overflow = 'hidden';
            _inf.innerHTML = groupImgHtml;
            _inf.style.overflow = 'hidden';
        } else {
            // Стек аватарок участников
            const stackHtml = groupAvatarStackInline(chat.members, 28);
            _hdr.innerHTML = stackHtml;
            _hdr.style.overflow = 'visible';
            _inf.innerHTML = groupAvatarStackInline(chat.members, 36);
            _inf.style.overflow = 'visible';
        }
        // Добавляем кнопку загрузки аватарки группы в info-panel
        let uploadBtn = document.getElementById('group-avatar-upload-btn');
        if (!uploadBtn) {
            uploadBtn = document.createElement('button');
            uploadBtn.id = 'group-avatar-upload-btn';
            uploadBtn.className = 'mt-3 text-xs text-blue-500 hover:text-blue-700 underline cursor-pointer';
            uploadBtn.textContent = 'Загрузить аватарку группы';
            uploadBtn.onclick = triggerGroupAvatarUpload;
            _inf.parentNode.insertBefore(uploadBtn, _inf.nextSibling);
        }
        uploadBtn.style.display = '';
    } else {
        const _ava = chatAvatarHtml(chat);
        _hdr.innerHTML = _ava; _hdr.style.overflow = 'hidden';
        _inf.innerHTML = _ava; _inf.style.overflow = 'hidden';
        // Скрываем кнопку загрузки если не группа
        const uploadBtn = document.getElementById('group-avatar-upload-btn');
        if (uploadBtn) uploadBtn.style.display = 'none';
    }
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
    const selected = window._ncSelected || [];

    // Храним пользователей в глобальном Map по id — безопасная передача в onclick
    if (!window._ncUserMap) window._ncUserMap = {};
    users.forEach(u => { window._ncUserMap[String(u.id)] = u; });

    if (!users.length) {
        container.innerHTML = '<p class="text-xs text-custom-muted text-center py-3 opacity-60">Пользователи не найдены</p>';
        return;
    }
    container.innerHTML = users.map(user => {
        const isSelected = selected.some(u => String(u.id) === String(user.id));
        const initial = (user.username || 'U')[0].toUpperCase();
        const checkSvg = isSelected
            ? `<svg class="w-3 h-3" fill="none" stroke="white" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"/></svg>`
            : '';
        return `<div onclick="ncToggleUser('${user.id}')"
             class="nc-result-item ${isSelected ? 'selected' : ''}"
             data-uid="${user.id}">
            <div class="w-8 h-8 rounded-full bg-blue-500/20 flex items-center justify-center text-blue-500 font-bold text-sm flex-shrink-0">${initial}</div>
            <div class="flex-1 min-w-0">
                <div class="font-semibold text-custom-main text-sm truncate">${user.username}</div>
                <div class="text-xs text-custom-muted truncate">${user.email}</div>
            </div>
            <div class="nc-check">${checkSvg}</div>
        </div>`;
    }).join('');
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