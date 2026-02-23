// ─── ui.js — обработчики модалей, профиля, аватара, темы ──────────────────

// ── Auth Modal ─────────────────────────────────────────────────────────────

function openAuthModal(type) {
    const overlay = document.getElementById('auth-overlay');
    const modal   = document.getElementById('auth-modal');
    overlay.classList.remove('hidden');
    setTimeout(() => { overlay.classList.remove('opacity-0'); modal.classList.remove('scale-95'); }, 10);
    switchForm(type);
}

function closeAuthModal() {
    const overlay = document.getElementById('auth-overlay');
    const modal   = document.getElementById('auth-modal');
    overlay.classList.add('opacity-0');
    modal.classList.add('scale-95');
    setTimeout(() => overlay.classList.add('hidden'), 300);
}

function switchForm(type) {
    document.getElementById('login-form-container').classList.toggle('hidden', type !== 'login');
    document.getElementById('register-form-container').classList.toggle('hidden', type !== 'register');
}

// ── New Chat Modal ─────────────────────────────────────────────────────────

function openNewChatMenu() {
    const overlay = document.getElementById('new-chat-overlay');
    const modal   = document.getElementById('new-chat-modal');
    overlay.classList.remove('hidden');
    setTimeout(() => { overlay.classList.remove('opacity-0'); modal.classList.remove('scale-95'); }, 10);
}

function closeNewChatModal() {
    const overlay = document.getElementById('new-chat-overlay');
    const modal   = document.getElementById('new-chat-modal');
    overlay.classList.add('opacity-0');
    modal.classList.add('scale-95');
    setTimeout(() => overlay.classList.add('hidden'), 300);
}

// ── Info Panel ─────────────────────────────────────────────────────────────

function toggleInfoPanel() {
    document.getElementById('info-panel').classList.toggle('hidden');
}

// ── Profile Modal ──────────────────────────────────────────────────────────

function openProfileModal() {
    const overlay = document.getElementById('profile-overlay');
    const modal   = document.getElementById('profile-modal');
    const user    = window.app.currentUser;

    if (user) {
        document.getElementById('profile-username').value    = user.username  || '';
        document.getElementById('profile-email').value       = user.email     || '';
        document.getElementById('profile-fullname').value    = user.full_name || '';
        document.getElementById('profile-phone').value       = user.phone     || '';
        document.getElementById('profile-status-text').value = user.status    || '';
        document.getElementById('profile-avatar').textContent = (user.username || 'U')[0].toUpperCase();
    }

    overlay.classList.remove('hidden');
    setTimeout(() => { overlay.classList.remove('opacity-0'); modal.classList.remove('scale-95'); }, 10);
}

function closeProfileModal() {
    const overlay = document.getElementById('profile-overlay');
    const modal   = document.getElementById('profile-modal');
    overlay.classList.add('opacity-0');
    modal.classList.add('scale-95');
    setTimeout(() => overlay.classList.add('hidden'), 300);
}

function triggerAvatarUpload() {
    document.getElementById('avatar-upload').click();
}

async function saveProfile() {
    const fullname   = document.getElementById('profile-fullname').value.trim();
    const phone      = document.getElementById('profile-phone').value.trim();
    const username   = document.getElementById('profile-username').value.trim();
    const statusText = document.getElementById('profile-status-text').value.trim();
    try {
        await apiSaveProfile({ fullname, phone, username, statusText });
        window.app.notify('Профиль обновлён ✓', 'success');
        closeProfileModal();
    } catch (err) {
        window.app.notify('Ошибка сохранения: ' + err.message, 'error');
    }
}

// ── Avatar Viewer ──────────────────────────────────────────────────────────

function openMyAvatarViewer() {
    const user = window.app.currentUser;
    if (!user) return;
    const name = user.username || 'U';
    document.getElementById('avatar-viewer-circle').textContent = name[0].toUpperCase();
    document.getElementById('avatar-viewer-name').textContent   = name;
    document.getElementById('avatar-viewer-status').textContent = 'вы · онлайн';
    document.getElementById('avatar-viewer-overlay').classList.remove('hidden');
}

function openAvatarViewer() {
    const app = window.app;
    if (!app.activeChatId) return;
    const chat = app.chats.find(c => String(c.id) === String(app.activeChatId));
    if (!chat) return;
    const name = chat.name || '?';
    document.getElementById('avatar-viewer-circle').textContent = name[0].toUpperCase();
    document.getElementById('avatar-viewer-name').textContent   = name;
    document.getElementById('avatar-viewer-status').textContent = chat.is_online ? 'онлайн' : 'офлайн';
    document.getElementById('avatar-viewer-overlay').classList.remove('hidden');
}

function closeAvatarViewer() {
    document.getElementById('avatar-viewer-overlay').classList.add('hidden');
}

// ── Send Button animation ──────────────────────────────────────────────────

function handleSendMessage(e) {
    e.preventDefault();
    animateSendButton(); // определена в neural-ai-btn.js
    window.app.sendMessage();
}

// ── Theme ──────────────────────────────────────────────────────────────────

function handleToggleTheme() {
    const btn = document.getElementById('theme-btn');
    btn.classList.add('animating');
    setTimeout(() => btn.classList.remove('animating'), 600);
    window.app.toggleTheme();
}

// ── Logout ─────────────────────────────────────────────────────────────────

function handleLogout() {
    window.app.logout();
}

// ── Global keyboard & click listeners ─────────────────────────────────────

document.addEventListener('keydown', function (e) {
    if (e.key === 'Escape') closeAvatarViewer();
});

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('profile-overlay').addEventListener('click', function (e) {
        if (e.target === this) closeProfileModal();
    });
});

// ── Message Context Menu ───────────────────────────────────────────────────

let _menuTimeout = null;

function showMessageMenu(event, messageId, isMe) {
    event.preventDefault();
    if (!isMe) return; // редактировать можно только свои

    closeMessageMenu();

    const menu = document.createElement('div');
    menu.id = 'msg-context-menu';
    menu.className = 'fixed z-[300] bg-custom-main border border-custom rounded-2xl shadow-xl py-1 min-w-[140px]';
    menu.style.left = event.clientX + 'px';
    menu.style.top  = event.clientY + 'px';
    menu.innerHTML = `
        <button onclick="startEditMessage('${messageId}')"
            class="w-full text-left px-4 py-2.5 text-sm text-custom-main hover:bg-custom-sidebar transition flex items-center gap-2">
            <svg class="w-4 h-4 opacity-60" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
            </svg>
            Редактировать
        </button>
    `;
    document.body.appendChild(menu);

    // Коррекция позиции если меню выходит за экран
    const rect = menu.getBoundingClientRect();
    if (rect.right > window.innerWidth)  menu.style.left = (event.clientX - rect.width) + 'px';
    if (rect.bottom > window.innerHeight) menu.style.top = (event.clientY - rect.height) + 'px';

    // Закрытие по клику вне меню
    setTimeout(() => document.addEventListener('click', closeMessageMenu, { once: true }), 0);
}

function closeMessageMenu() {
    const menu = document.getElementById('msg-context-menu');
    if (menu) menu.remove();
}

function startEditMessage(messageId) {
    closeMessageMenu();

    const msg = window.app.messages.find(m => String(m.id) === String(messageId));
    if (!msg) return;

    const contentEl = document.getElementById(`msg-content-${messageId}`);
    if (!contentEl) return;

    const original = msg.content;

    contentEl.innerHTML = `
        <div class="edit-inline-wrap">
            <textarea id="edit-input-${messageId}"
                class="edit-inline-textarea"
                rows="1"
            >${escapeHtmlAttr(original)}</textarea>
            <div class="edit-inline-actions">
                <button onclick="cancelEditMessage('${messageId}', ${JSON.stringify(escapeHtmlAttr(original))})"
                    class="edit-btn-cancel">
                    <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M18 6L6 18M6 6l12 12"/></svg>
                    Отмена
                </button>
                <button onclick="confirmEditMessage('${messageId}')"
                    class="edit-btn-confirm">
                    <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M20 6L9 17l-5-5"/></svg>
                    Сохранить
                </button>
            </div>
        </div>
    `;

    const input = document.getElementById(`edit-input-${messageId}`);

    // Автовысота textarea
    input.style.height = 'auto';
    input.style.height = input.scrollHeight + 'px';
    input.addEventListener('input', () => {
        input.style.height = 'auto';
        input.style.height = input.scrollHeight + 'px';
    });

    input.focus();
    input.setSelectionRange(input.value.length, input.value.length);

    input.addEventListener('keydown', function(e) {
        if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); confirmEditMessage(messageId); }
        if (e.key === 'Escape') cancelEditMessage(messageId, original);
    });
}

async function confirmEditMessage(messageId) {
    const input = document.getElementById(`edit-input-${messageId}`);
    if (!input) return;
    const newContent = input.value.trim();
    if (!newContent) return;

    try {
        await apiEditMessage(messageId, newContent);
        // WS придёт и обновит — но обновим и локально сразу для быстрого отклика
        const msg = window.app.messages.find(m => String(m.id) === String(messageId));
        if (msg) {
            msg.content   = newContent;
            msg.edited_at = new Date().toISOString();
            renderMessages();
        }
    } catch (err) {
        window.app.notify('Ошибка редактирования: ' + err.message, 'error');
        cancelEditMessage(messageId, window.app.messages.find(m => String(m.id) === String(messageId))?.content || '');
    }
}

function cancelEditMessage(messageId, original) {
    const msg = window.app.messages.find(m => String(m.id) === String(messageId));
    if (msg) {
        msg.content = original;
        renderMessages();
    }
}

function escapeHtmlAttr(text) {
    return text.replace(/&/g,'&amp;').replace(/"/g,'&quot;').replace(/'/g,'&#39;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

// ── Paperclip unbend animation ─────────────────────────────────────────────

(function () {
    document.addEventListener('DOMContentLoaded', function () {
        const btn  = document.getElementById('attach-btn');
        const hook = document.getElementById('clip-hook');
        const body = document.getElementById('clip-body');
        if (!btn || !hook || !body) return;

        // hook transform-origin — точка где крюк крепится к телу скрепки
        hook.style.transformOrigin = '17.2px 6.8px';
        body.style.transformOrigin = '12px 12px';

        btn.addEventListener('mouseenter', () => {
            // Крюк разгибается: поворот + смещение вправо-вверх
            hook.style.transition = 'transform 0.38s cubic-bezier(0.34,1.6,0.64,1)';
            hook.style.transform  = 'rotate(-52deg) translate(3.5px, -2px)';

            // Тело чуть смещается вниз — как будто держат второй рукой
            body.style.transition = 'transform 0.3s cubic-bezier(0.34,1.4,0.64,1) 0.05s';
            body.style.transform  = 'translateY(1.5px) rotate(3deg)';
        });

        btn.addEventListener('mouseleave', () => {
            // Всё возвращается с пружиной
            hook.style.transition = 'transform 0.35s cubic-bezier(0.34,1.8,0.64,1)';
            hook.style.transform  = 'rotate(0deg) translate(0,0)';

            body.style.transition = 'transform 0.28s cubic-bezier(0.34,1.5,0.64,1)';
            body.style.transform  = 'translateY(0) rotate(0deg)';
        });
    });
})();