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

    const isGroup    = !!chat.is_group;
    const isCreator  = isGroup && chat.creator_id && String(chat.creator_id) === String(app.currentUser?.id);
    const displayName = chatDisplayName(chat);

    // ── Имя в хедере чата ──────────────────────────────────────────────
    document.getElementById('active-chat-name').textContent = displayName;

    // info-name: для создателя — кликабелен для inline-редактирования
    const infoNameEl = document.getElementById('info-name');
    infoNameEl.textContent = displayName;
    if (isCreator) {
        infoNameEl.style.cursor = 'pointer';
        infoNameEl.title = 'Нажмите для редактирования';
        infoNameEl.classList.add('group-edit-name');
        infoNameEl.onclick = () => startInlineGroupNameEdit(chat.id, chat.name || '');
    } else {
        infoNameEl.style.cursor = '';
        infoNameEl.title = '';
        infoNameEl.classList.remove('group-edit-name');
        infoNameEl.onclick = null;
    }

    const headerStatus = document.getElementById('active-chat-status');
    if (isGroup) {
        if (headerStatus) {
            headerStatus.textContent = '';   // не показываем "офлайн/онлайн" для группы
            headerStatus.className   = 'text-xs text-custom-muted';
        }
        const badge = document.getElementById('info-status-badge');
        if (badge) badge.style.display = 'none';
        const statusEl = document.getElementById('info-user-status');
        if (statusEl) statusEl.textContent = '';
    } else {
        renderStatusElements(chat.is_online, chat.user_status || '');
        const badge = document.getElementById('info-status-badge');
        if (badge) badge.style.display = '';
    }

    // ── Аватар в хедере и info-panel ──────────────────────────────────
    const _hdr = document.getElementById('active-chat-avatar');
    const _inf = document.getElementById('info-avatar');

    if (isGroup && chat.members && chat.members.length > 1) {
        if (chat.avatar_url) {
            const groupImgHtml = `<img src="${chat.avatar_url}" style="width:100%;height:100%;object-fit:cover;">`;
            _hdr.innerHTML = groupImgHtml; _hdr.style.overflow = 'hidden';
            _inf.innerHTML = groupImgHtml; _inf.style.overflow = 'hidden';
        } else {
            const stackHtml = groupAvatarStackInline(chat.members, 28);
            _hdr.innerHTML = stackHtml; _hdr.style.overflow = 'visible';
            _inf.innerHTML = groupAvatarStackInline(chat.members, 36);
            _inf.style.overflow = 'visible';
        }
    } else {
        const _ava = chatAvatarHtml(chat);
        _hdr.innerHTML = _ava; _hdr.style.overflow = 'hidden';
        _inf.innerHTML = _ava; _inf.style.overflow = 'hidden';
    }

    if (isGroup && isCreator) {
        _inf.style.cursor = 'pointer';
        _inf.title = 'Сменить аватар группы';
        _inf.onclick = () => triggerGroupAvatarUpload();
    } else {
        _inf.style.cursor = '';
        _inf.title = '';
        _inf.onclick = null;
    }

    // ── Блок участников в info-panel ──────────────────────────────────
    let membersBlock = document.getElementById('info-members-block');
    if (!membersBlock) {
        membersBlock = document.createElement('div');
        membersBlock.id = 'info-members-block';
        membersBlock.className = 'w-full text-left px-4 pb-4';
        const statusP = document.getElementById('info-user-status');
        if (statusP && statusP.parentNode) {
            statusP.parentNode.insertBefore(membersBlock, statusP.nextSibling);
        }
    }
    membersBlock.innerHTML = '';

    if (!isGroup) return;

    apiGetGroupMembers(chat.id).then(members => {
        if (!members || !members.length) return;

        const localChat = app.chats.find(c => String(c.id) === String(chat.id));
        if (localChat) localChat.members = members;

        const membersList = members.map(m => {
            const isMe = String(m.id) === String(app.currentUser?.id);
            const avatarInner = m.avatar_url
                ? `<img src="${m.avatar_url}" style="width:100%;height:100%;object-fit:cover;">`
                : (m.username || '?')[0].toUpperCase();
            const onlineDot = m.is_online
                ? `<span class="member-online-dot"></span>`
                : '';
            const removeBtn = (isCreator && !isMe)
                ? `<button onclick="removeGroupMember('${chat.id}','${m.id}')"
                        title="Удалить из группы"
                        class="ml-auto text-red-400 hover:text-red-600 p-1 rounded-lg hover:bg-red-50 dark:hover:bg-red-950/20 transition flex-shrink-0 member-remove-btn">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                        </svg>
                    </button>`
                : '';
            return `
                <div class="member-row flex items-center gap-3 py-1.5">
                    <div class="relative flex-shrink-0">
                        <div onclick="openMemberAvatarViewer(${JSON.stringify(m).replace(/"/g,'&quot;')})"
                            class="w-9 h-9 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold overflow-hidden cursor-pointer hover:ring-2 hover:ring-blue-400 hover:ring-offset-1 transition text-sm">
                            ${avatarInner}
                        </div>
                        ${onlineDot}
                    </div>
                    <div class="overflow-hidden flex-1 min-w-0">
                        <div class="text-sm font-semibold text-custom-main truncate">
                            ${m.username}${isMe ? ' <span class="text-xs text-custom-muted font-normal">(вы)</span>' : ''}
                        </div>
                        ${m.full_name ? `<div class="text-xs text-custom-muted truncate">${m.full_name}</div>` : ''}
                        ${m.status ? `<div class="text-xs text-custom-muted truncate" style="font-style:italic;opacity:0.8;">${m.status}</div>` : ''}
                    </div>
                    ${removeBtn}
                </div>`;
        }).join('');

        // Кнопка + рядом с заголовком — только для создателя
        const addBtnHtml = isCreator
            ? `<button id="add-member-plus-btn" onclick="handleAddMemberBtn()" title="Добавить участника"
                    class="add-member-plus-btn w-7 h-7 flex items-center justify-center rounded-full relative">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                    </svg>
                </button>`
            : `<div class="w-7 h-7"></div>`;  // заглушка для выравнивания

        membersBlock.innerHTML = `
            <div class="flex items-center justify-between mb-2 mt-1">
                <div class="text-xs font-semibold text-custom-muted uppercase tracking-wider">
                    Участники (${members.length})
                </div>
                ${addBtnHtml}
            </div>
            <div id="group-members-list" class="space-y-0.5">
                ${membersList}
            </div>`;
    }).catch(() => {});
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

    // В режиме добавления — исключаем тех, кто уже в группе
    const excludeIds = window._addMemberMode ? (window._ncExcludeIds || new Set()) : new Set();
    const filtered = users.filter(u => !excludeIds.has(String(u.id)));

    if (!filtered.length) {
        const msg = (window._addMemberMode && users.length)
            ? 'Все найденные пользователи уже в группе'
            : 'Пользователи не найдены';
        container.innerHTML = `<p class="text-xs text-custom-muted text-center py-3 opacity-60">${msg}</p>`;
        return;
    }
    container.innerHTML = filtered.map(user => {
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