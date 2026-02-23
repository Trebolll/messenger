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