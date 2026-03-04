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

// Единая функция анимации плюса — используется везде
function animatePlusBtn(btn) {
  if (!btn) return;
  btn.classList.remove('clicked', 'pulsing');
  void btn.offsetWidth;
  btn.classList.add('clicked');
  setTimeout(() => btn.classList.add('pulsing'), 80);
  setTimeout(() => btn.classList.remove('clicked', 'pulsing'), 800);
}

function handleNewChatBtn() {
  const overlay = document.getElementById('new-chat-overlay');
  const isOpen  = !overlay.classList.contains('hidden') && !overlay.classList.contains('opacity-0');

  animatePlusBtn(document.getElementById('dock-new-chat'));

  if (isOpen) {
    closeNewChatModal();
  } else {
    openNewChatMenu();
  }
}

function openNewChatMenu() {
  // Сбрасываем состояние выбора
  window._ncSelected = [];
  ncRenderChips();
  ncUpdateFooter();

  const overlay = document.getElementById('new-chat-overlay');
  const modal   = document.getElementById('new-chat-modal');
  const input   = document.getElementById('user-search-input');
  const results = document.getElementById('search-results');
  if (input) { input.value = ''; }
  if (results) { results.innerHTML = ''; }

  overlay.classList.remove('hidden');
  setTimeout(() => {
    overlay.classList.remove('opacity-0');
    modal.classList.remove('scale-95');
    modal.style.transform = '';
    if (input) input.focus();
  }, 10);
}

function closeNewChatModal() {
  const overlay = document.getElementById('new-chat-overlay');
  const modal   = document.getElementById('new-chat-modal');
  overlay.classList.add('opacity-0');
  modal.style.transform = 'scale(0.88) translateY(12px)';
  modal.style.opacity = '0';
  setTimeout(() => {
    overlay.classList.add('hidden');
    modal.style.transform = '';
    modal.style.opacity = '';
    window._ncSelected = [];
    window._ncExcludeIds = new Set();
    if (window._addMemberMode) {
      window._addMemberMode = false;
      const modalTitle = document.querySelector('#new-chat-modal h2');
      if (modalTitle) modalTitle.textContent = 'Новый чат';
      const subtitle = document.getElementById('new-chat-subtitle');
      if (subtitle) subtitle.textContent = 'Найди пользователя для начала общения';
    }
  }, 360);
}

// Переключить пользователя в выборке (принимает id, берёт из Map)
function ncToggleUser(userId) {
  if (!window._ncSelected) window._ncSelected = [];
  const uid = String(userId);
  // Берём полный объект из Map, заполненного в renderSearchResults
  const user = (window._ncUserMap || {})[uid];
  if (!user) return;

  const idx = window._ncSelected.findIndex(u => String(u.id) === uid);
  if (idx === -1) {
    window._ncSelected.push(user);
  } else {
    window._ncSelected.splice(idx, 1);
  }
  ncRenderChips();
  ncUpdateFooter();

  // Обновляем галочки в списке без перерендера
  document.querySelectorAll('#search-results .nc-result-item').forEach(el => {
    const elUid = el.dataset.uid;
    const sel = window._ncSelected.some(u => String(u.id) === elUid);
    el.classList.toggle('selected', sel);
    const check = el.querySelector('.nc-check');
    if (check) {
      check.innerHTML = sel
          ? '<svg class="w-3 h-3" fill="none" stroke="white" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"/></svg>'
          : '';
    }
  });
}

// Рендерим плашки выбранных
function ncRenderChips() {
  const bar    = document.getElementById('selected-users-bar');
  const chips  = document.getElementById('selected-chips');
  const sel    = window._ncSelected || [];
  if (!bar || !chips) return;

  if (sel.length === 0) {
    bar.classList.add('hidden');
    chips.innerHTML = '';
    return;
  }
  bar.classList.remove('hidden');
  chips.innerHTML = sel.map(u => `
    <span class="nc-chip">
      <span class="nc-chip-avatar">${(u.username||'U')[0].toUpperCase()}</span>
      <span>${u.username}</span>
      <svg onclick="ncToggleUser(${JSON.stringify(u).replace(/"/g,'&quot;')})" class="nc-chip-remove w-3.5 h-3.5 cursor-pointer" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12"/>
      </svg>
    </span>
  `).join('');
}

// Обновляем кнопку снизу
function ncUpdateFooter() {
  const footer    = document.getElementById('new-chat-footer');
  const btn       = document.getElementById('create-chat-btn');
  const label     = document.getElementById('create-chat-label');
  const subtitle  = document.getElementById('new-chat-subtitle');
  const groupWrap = document.getElementById('group-name-wrap');
  const sel       = window._ncSelected || [];

  if (!footer || !btn || !label) return;

  // Режим добавления участника в группу
  if (window._addMemberMode) {
    if (sel.length === 0) {
      footer.classList.add('hidden');
      if (groupWrap) groupWrap.classList.add('hidden');
    } else {
      footer.classList.remove('hidden');
      if (groupWrap) groupWrap.classList.add('hidden');
      btn.classList.add('group-mode');
      label.textContent = sel.length === 1
          ? `Добавить ${sel[0].username}`
          : `Добавить ${sel.length} участника`;
      if (subtitle) subtitle.textContent = `Выбрано: ${sel.length}`;
    }
    return;
  }

  if (sel.length === 0) {
    footer.classList.add('hidden');
    if (groupWrap) groupWrap.classList.add('hidden');
    if (subtitle) subtitle.textContent = 'Найди пользователя для начала общения';
  } else if (sel.length === 1) {
    footer.classList.remove('hidden');
    if (groupWrap) groupWrap.classList.add('hidden');
    btn.classList.remove('group-mode');
    label.textContent = `Открыть чат с ${sel[0].username}`;
    if (subtitle) subtitle.textContent = 'Выбран 1 участник · приватный чат';
  } else {
    footer.classList.remove('hidden');
    if (groupWrap) groupWrap.classList.remove('hidden');
    btn.classList.add('group-mode');
    label.textContent = `Создать группу · ${sel.length} участника`;
    if (subtitle) subtitle.textContent = `Выбрано: ${sel.length}`;
  }
}

// Действие при нажатии на кнопку создания
async function handleCreateChat() {
  const sel = window._ncSelected || [];
  if (sel.length === 0) return;

  // Режим добавления участника в группу
  if (window._addMemberMode) {
    const chatId = window.app?.activeChatId;
    if (!chatId) return;
    try {
      for (const u of sel) {
        await apiAddChatMember(chatId, u.username);
      }
      const names = sel.map(u => u.username).join(', ');
      window.app.notify(`${names} добавлен(ы) в группу ✓`, 'success');
      closeAddMemberModal();
      await apiLoadChats();
      renderChatHeader();
    } catch (err) {
      window.app.notify('Ошибка: ' + err.message, 'error');
    }
    return;
  }

  if (sel.length === 1) {
    await app.createPrivateChat(sel[0].id);
  } else {
    const groupName = document.getElementById('group-name-input')?.value.trim() || '';
    await apiCreateGroupChat(sel.map(u => u.id), groupName);
  }
}

// ── Info Panel ─────────────────────────────────────────────────────────────

function toggleInfoPanel() {
  document.getElementById('info-panel').classList.toggle('hidden');
}

// ── Profile Modal ──────────────────────────────────────────────────────────

let _profileModalOpen = false;

function toggleProfileModal() {
  if (_profileModalOpen) {
    closeProfileModal();
  } else {
    openProfileModal();
  }
}

function openProfileModal() {
  const overlay = document.getElementById('profile-overlay');
  const modal   = document.getElementById('profile-modal');
  const user    = window.app.currentUser;

  document.body.classList.add('profile-open');

  if (user) {
    document.getElementById('profile-username').value    = user.username  || '';
    document.getElementById('profile-email').value       = user.email     || '';
    document.getElementById('profile-fullname').value    = user.full_name || '';
    document.getElementById('profile-phone').value       = user.phone     || '';
    document.getElementById('profile-status-text').value = user.status    || '';
    setAvatarEl(document.getElementById('profile-avatar'), user);
  }

  overlay.classList.remove('hidden');
  setTimeout(() => { overlay.classList.remove('opacity-0'); modal.classList.remove('scale-95'); }, 10);
  _profileModalOpen = true;
  const btn = document.getElementById('profile-dock-btn');
  if (btn) btn.classList.add('active');
}

function closeProfileModal() {
  const overlay = document.getElementById('profile-overlay');
  const modal   = document.getElementById('profile-modal');
  overlay.classList.add('opacity-0');
  modal.classList.add('scale-95');
  document.body.classList.remove('profile-open');
  _profileModalOpen = false;
  setTimeout(() => overlay.classList.add('hidden'), 300);
  const btn = document.getElementById('profile-dock-btn');
  if (btn) btn.classList.remove('active');
}

function triggerAvatarUpload() {
  const input = document.getElementById('avatar-upload');
  if (!input) return;
  // Сбрасываем value чтобы повторный выбор того же файла тоже срабатывал
  input.value = '';
  input.onchange = handleAvatarSelected;
  input.click();
}

async function handleAvatarSelected(e) {
  const file = e.target.files[0];
  if (!file) return;

  // Показываем превью мгновенно
  const avatarEl = document.getElementById('profile-avatar');
  const previewUrl = URL.createObjectURL(file);
  avatarEl.innerHTML = `<img src="${previewUrl}" style="width:100%;height:100%;object-fit:cover;border-radius:50%;">`;

  // Показываем спиннер поверх
  avatarEl.style.position = 'relative';
  const spinner = document.createElement('div');
  spinner.id = 'avatar-spinner';
  spinner.innerHTML = `<div style="position:absolute;inset:0;background:rgba(0,0,0,0.4);border-radius:50%;
        display:flex;align-items:center;justify-content:center;">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2"
            style="animation:spin 0.8s linear infinite">
            <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
        </svg>
    </div>`;
  avatarEl.appendChild(spinner);

  try {
    const result = await apiUploadAvatar(file);
    // Сохраняем в currentUser и localStorage
    window.app.currentUser.avatar_url = result.avatar_url;
    const stored = JSON.parse(localStorage.getItem('alpha_user') || '{}');
    stored.avatar_url = result.avatar_url;
    localStorage.setItem('alpha_user', JSON.stringify(stored));

    // Обновляем все аватары на странице
    loadUserData();
    setAvatarEl(avatarEl, window.app.currentUser);
    window.app.notify('Фото обновлено ✓', 'success');
  } catch (err) {
    window.app.notify('Ошибка: ' + err.message, 'error');
    // Откатываем превью
    setAvatarEl(avatarEl, window.app.currentUser);
  } finally {
    const sp = document.getElementById('avatar-spinner');
    if (sp) sp.remove();
    URL.revokeObjectURL(previewUrl);
  }
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
  const circleEl = document.getElementById('avatar-viewer-circle');
  circleEl.style.overflow = 'hidden';
  circleEl.innerHTML = userAvatarHtml(user);
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
  const circleEl = document.getElementById('avatar-viewer-circle');
  circleEl.style.overflow = 'hidden';
  if (chat.avatar_url) {
    circleEl.innerHTML = `<img src="${chat.avatar_url}" style="width:100%;height:100%;object-fit:cover;border-radius:50%;">`;
  } else {
    circleEl.textContent = name[0].toUpperCase();
  }
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
  const btnLanding = document.getElementById('theme-btn-landing');
  const btnNav = document.getElementById('theme-btn');

  if (btnLanding) {
    btnLanding.classList.add('animating');
    setTimeout(() => btnLanding.classList.remove('animating'), 600);
  }
  if (btnNav) {
    btnNav.classList.add('animating');
    setTimeout(() => btnNav.classList.remove('animating'), 600);
  }

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

  closeMessageMenu();

  const msg = window.app?.messages?.find(m => String(m.id) === String(messageId));
  const msgText = msg?.content || '';

  const menu = document.createElement('div');
  menu.id = 'msg-context-menu';
  menu.className = 'fixed z-[300] bg-custom-main border border-custom rounded-2xl shadow-xl py-1 min-w-[160px]';
  menu.style.left = event.clientX + 'px';
  menu.style.top  = event.clientY + 'px';

  let menuHtml = '';

  // Копировать — для всех сообщений с текстом
  if (msgText) {
    menuHtml += `
        <button onclick="copyMessageText('${messageId}')"
            class="w-full text-left px-4 py-2.5 text-sm text-custom-main hover:bg-custom-sidebar transition flex items-center gap-2">
            <svg class="w-4 h-4 opacity-60" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
            </svg>
            Копировать
        </button>`;
  }

  // Редактировать и удалить — только для своих
  if (isMe) {
    if (menuHtml) menuHtml += '<div class="mx-3 my-1 border-t border-custom opacity-50"></div>';
    menuHtml += `
        <button onclick="startEditMessage('${messageId}')"
            class="w-full text-left px-4 py-2.5 text-sm text-custom-main hover:bg-custom-sidebar transition flex items-center gap-2">
            <svg class="w-4 h-4 opacity-60" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
            </svg>
            Редактировать
        </button>
        <div class="mx-3 my-1 border-t border-custom opacity-50"></div>
        <button onclick="executeDeleteMessage('${messageId}')"
            class="w-full text-left px-4 py-2.5 text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-950/20 transition flex items-center gap-2">
            <svg class="w-4 h-4 opacity-70" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
            </svg>
            Удалить
        </button>`;
  }

  if (!menuHtml) return; // нечего показывать
  menu.innerHTML = menuHtml;
  document.body.appendChild(menu);

  // Коррекция позиции если меню выходит за экран
  const rect = menu.getBoundingClientRect();
  if (rect.right > window.innerWidth)  menu.style.left = (event.clientX - rect.width) + 'px';
  if (rect.bottom > window.innerHeight) menu.style.top = (event.clientY - rect.height) + 'px';

  // Закрытие по клику вне меню
  setTimeout(() => document.addEventListener('click', closeMessageMenu, { once: true }), 0);
}

function copyMessageText(messageId) {
  closeMessageMenu();
  const msg = window.app?.messages?.find(m => String(m.id) === String(messageId));
  if (!msg || !msg.content) return;
  navigator.clipboard.writeText(msg.content).then(() => {
    window.app?.notify?.('Текст скопирован', 'success');
  }).catch(() => {
    // Fallback
    const ta = document.createElement('textarea');
    ta.value = msg.content;
    document.body.appendChild(ta);
    ta.select();
    document.execCommand('copy');
    ta.remove();
    window.app?.notify?.('Текст скопирован', 'success');
  });
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

  // Сохраняем оригинальный HTML для восстановления при отмене
  contentEl.dataset.originalHtml = contentEl.innerHTML;

  contentEl.innerHTML = `
        <div class="edit-inline-wrap">
            <textarea id="edit-input-${messageId}"
                class="edit-inline-textarea"
                rows="1"
            >${escapeHtmlAttr(original)}</textarea>
            <div class="edit-inline-actions">
                <button onclick="cancelEditMessage('${messageId}')"
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
    if (e.key === 'Escape') cancelEditMessage(messageId);
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
    cancelEditMessage(messageId);
  }
}

function cancelEditMessage(messageId) {
  const contentEl = document.getElementById(`msg-content-${messageId}`);
  if (!contentEl) return;
  const original = contentEl.dataset.originalHtml;
  if (original !== undefined) contentEl.innerHTML = original;
}

function escapeHtmlAttr(text) {
  return text.replace(/&/g,'&amp;').replace(/"/g,'&quot;').replace(/'/g,'&#39;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

// ── Pin (булавка) animation ──────────────────────────────────────────────────

(function () {
  document.addEventListener('DOMContentLoaded', function () {
    const btn  = document.getElementById('attach-btn');
    if (!btn) return;

    // При клике — лёгкая тактильная вибрация (мобильные)
    btn.addEventListener('click', () => {
      if (navigator.vibrate) navigator.vibrate(8);
    });

    // После active — плавно возвращаемся в нормальное положение
    // (браузер сам убирает :active, но добавим микро-задержку для эффекта втыкания)
    btn.addEventListener('mousedown', () => {
      const wrap = document.getElementById('pin-wrap');
      if (!wrap) return;
      // Короткая задержка — булавка "застряла" на 80мс
      wrap.style.transitionDuration = '0.08s';
    });

    btn.addEventListener('mouseup', () => {
      const wrap = document.getElementById('pin-wrap');
      if (!wrap) return;
      setTimeout(() => {
        wrap.style.transitionDuration = '';
      }, 80);
    });
  });
})();

// ── Delete Message ─────────────────────────────────────────────────────────

function confirmDeleteMessage(messageId) {
  closeMessageMenu();

  // Закрываем другие открытые inline-диалоги
  document.querySelectorAll('.inline-delete-wrap').forEach(el => {
    const mid = el.closest('[data-msg-id]')?.dataset.msgId;
    if (mid) cancelDeleteMessage(mid);
  });

  const contentEl = document.getElementById(`msg-content-${messageId}`);
  if (!contentEl) return;

  // Сохраняем оригинальный HTML чтобы восстановить при отмене
  const originalHTML = contentEl.innerHTML;
  contentEl.dataset.originalHtml = originalHTML;

  contentEl.innerHTML = `
        <div class="inline-delete-wrap">
            <p class="inline-delete-text">Удалить сообщение?</p>
            <div class="edit-inline-actions">
                <button onclick="cancelDeleteMessage('${messageId}')"
                    class="edit-btn-cancel">
                    <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M18 6L6 18M6 6l12 12"/></svg>
                    Отмена
                </button>
                <button onclick="executeDeleteMessage('${messageId}')"
                    class="edit-btn-delete">
                    <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
                    Удалить
                </button>
            </div>
        </div>
    `;

  // Escape отменяет
  const onKey = (e) => {
    if (e.key === 'Escape') {
      cancelDeleteMessage(messageId);
      document.removeEventListener('keydown', onKey);
    }
  };
  document.addEventListener('keydown', onKey);
}

function cancelDeleteMessage(messageId) {
  const contentEl = document.getElementById(`msg-content-${messageId}`);
  if (!contentEl) return;
  const original = contentEl.dataset.originalHtml;
  if (original) contentEl.innerHTML = original;
}

async function executeDeleteMessage(messageId) {
  try {
    await apiDeleteMessage(messageId);
    const el = document.querySelector(`[data-msg-id="${messageId}"]`);
    if (el) {
      ashDisintegrate(el, () => {
        window.app.messages = window.app.messages.filter(m => String(m.id) !== String(messageId));
        el.remove();
      });
    }
  } catch (err) {
    cancelDeleteMessage(messageId);
    window.app.notify('Не удалось удалить: ' + err.message, 'error');
  }
}

// ── Ash Disintegrate Animation ─────────────────────────────────────────────

function ashDisintegrate(el, onDone) {
  const bubble = el.querySelector('.message-bubble') || el;
  const r      = bubble.getBoundingClientRect();
  if (r.width === 0) { onDone?.(); return; }

  const isMe = el.classList.contains('justify-end');

  // ── Считываем реальный цвет пузырька ──
  const computed  = getComputedStyle(bubble);
  const bgRaw     = computed.backgroundColor;
  const match     = bgRaw.match(/[\d.]+/g) || [];
  const [rr=80, gg=120, bb=200] = match.map(Number);

  // ── Canvas поверх всего ──
  const cv = document.createElement('canvas');
  cv.style.cssText = 'position:fixed;inset:0;pointer-events:none;z-index:9999;';
  cv.width  = window.innerWidth;
  cv.height = window.innerHeight;
  document.body.appendChild(cv);
  const ctx = cv.getContext('2d');

  // ── Параметры ──
  const PARTICLE_COUNT = 380;
  const cx = r.left + r.width  / 2;
  const cy = r.top  + r.height / 2;

  // ── Генерируем частицы ──
  const particles = [];
  for (let i = 0; i < PARTICLE_COUNT; i++) {
    // Стартовая позиция — случайная внутри пузырька
    const startX = r.left + Math.random() * r.width;
    const startY = r.top  + Math.random() * r.height;

    // Угол вылета — от центра пузырька наружу + случайное отклонение
    const baseAngle  = Math.atan2(startY - cy, startX - cx);
    const scatter    = (Math.random() - 0.5) * Math.PI * 1.4;
    const angle      = baseAngle + scatter;

    // Скорость — быстрее у краёв
    const distRatio  = Math.hypot(startX - cx, startY - cy) / Math.hypot(r.width, r.height) * 2;
    const speed      = (1.5 + Math.random() * 5.5) * (0.6 + distRatio * 0.8);

    // Размер частицы — mix мелких и крупных
    const size = Math.random() < 0.7
        ? 1.5 + Math.random() * 2.5   // мелкие
        : 3.5 + Math.random() * 4.5;  // крупные

    // Цвет — вариации базового цвета пузырька + белые блики
    const bright = 0.7 + Math.random() * 0.6;
    const isHighlight = Math.random() < 0.15;
    const pr = isHighlight ? 255 : Math.min(255, Math.round(rr * bright));
    const pg = isHighlight ? 255 : Math.min(255, Math.round(gg * bright));
    const pb = isHighlight ? 255 : Math.min(255, Math.round(bb * bright));

    // Форма: круг или прямоугольник
    const shape = Math.random() < 0.55 ? 'circle' : 'rect';

    particles.push({
      x: startX,  y: startY,
      vx: Math.cos(angle) * speed,
      vy: Math.sin(angle) * speed - (1 + Math.random() * 2),
      size,
      shape,
      rot: Math.random() * Math.PI * 2,
      rotV: (Math.random() - 0.5) * 0.3,
      life: 1,
      decay: 0.022 + Math.random() * 0.025,
      delay: Math.random() * 12,  // небольшой стаггер
      color: `${pr},${pg},${pb}`,
    });
  }

  // ── Скрываем оригинальный элемент ──
  bubble.style.transition = 'opacity 0.06s ease';
  bubble.style.opacity = '0';

  let frame = 0;
  let done  = false;

  function tick() {
    ctx.clearRect(0, 0, cv.width, cv.height);
    frame++;
    let alive = 0;

    for (const p of particles) {
      if (frame < p.delay) { alive++; continue; }
      if (p.life <= 0) continue;
      alive++;

      // Физика
      p.x   += p.vx;
      p.y   += p.vy;
      p.vy  += 0.14;      // гравитация
      p.vx  *= 0.968;     // трение воздуха
      p.vy  *= 0.985;
      p.life -= p.decay;
      p.rot  += p.rotV;

      const alpha = Math.max(0, p.life);
      const sz    = p.size * (0.3 + alpha * 0.7);

      ctx.save();
      ctx.globalAlpha = alpha * alpha; // квадратичное затухание — резче
      ctx.translate(p.x, p.y);
      ctx.rotate(p.rot);
      ctx.fillStyle = `rgba(${p.color},1)`;

      if (p.shape === 'circle') {
        ctx.beginPath();
        ctx.arc(0, 0, sz / 2, 0, Math.PI * 2);
        ctx.fill();
      } else {
        const hw = sz * (0.5 + Math.random() * 0.3);
        const hh = sz * (0.3 + Math.random() * 0.2);
        ctx.fillRect(-hw, -hh, hw * 2, hh * 2);
      }

      ctx.restore();
    }

    if (alive > 0) {
      requestAnimationFrame(tick);
    } else if (!done) {
      done = true;
      cv.remove();
      onDone?.();
    }
  }

  requestAnimationFrame(tick);
}
// ─── Аватарка группового чата ─────────────────────────────────────────────

function triggerGroupAvatarUpload() {
  let input = document.getElementById('group-avatar-upload-input');
  if (!input) {
    input = document.createElement('input');
    input.type = 'file';
    input.id = 'group-avatar-upload-input';
    input.accept = 'image/*';
    input.style.display = 'none';
    document.body.appendChild(input);
  }
  input.value = '';
  input.onchange = handleGroupAvatarSelected;
  input.click();
}

async function handleGroupAvatarSelected(e) {
  const file = e.target.files[0];
  if (!file) return;
  const chatId = window.app?.activeChatId;
  if (!chatId) return;

  // Мгновенный превью в info-panel
  const infAva = document.getElementById('info-avatar');
  const previewUrl = URL.createObjectURL(file);
  if (infAva) {
    infAva.innerHTML = `<img src="${previewUrl}" style="width:100%;height:100%;object-fit:cover;">`;
    infAva.style.overflow = 'hidden';
  }

  try {
    const result = await apiUploadGroupAvatar(chatId, file);
    // Обновляем данные чата в памяти
    const chat = window.app.chats.find(c => String(c.id) === String(chatId));
    if (chat) {
      chat.avatar_url = result.avatar_url;
    }
    renderChats();
    renderChatHeader();
    window.app.notify('Аватарка группы обновлена ✓', 'success');
  } catch (err) {
    window.app.notify('Ошибка: ' + err.message, 'error');
    // Откатываем превью
    renderChatHeader();
  } finally {
    URL.revokeObjectURL(previewUrl);
  }
}
// ── Group Chat Management ─────────────────────────────────────────────────

async function saveGroupName(chatId) {
  // Legacy — kept for compatibility, inline edit uses startInlineGroupNameEdit
  const input = document.getElementById('group-name-edit-input');
  if (!input) return;
  const name = input.value.trim();
  if (!name) { window.app.notify('Название не может быть пустым', 'error'); return; }
  try {
    const updated = await apiUpdateGroupInfo(chatId, name);
    const chat = window.app.chats.find(c => String(c.id) === String(chatId));
    if (chat) { chat.name = updated.name || name; }
    renderChats();
    renderChatHeader();
    window.app.notify('Название обновлено ✓', 'success');
  } catch (err) {
    window.app.notify('Ошибка: ' + err.message, 'error');
  }
}

// ── Inline-редактирование названия группы ────────────────────────────────

function startInlineGroupNameEdit(chatId, currentName) {
  const infoNameEl = document.getElementById('info-name');
  if (!infoNameEl || infoNameEl.querySelector('input')) return; // уже редактируется

  const originalText = currentName || infoNameEl.textContent.trim();

  // Заменяем h4 на input
  infoNameEl.innerHTML = '';
  const input = document.createElement('input');
  input.id = 'inline-group-name-input';
  input.type = 'text';
  input.value = originalText;
  input.className = 'inline-group-name-input';
  infoNameEl.appendChild(input);
  input.focus();
  input.select();

  let saved = false;
  const save = async () => {
    if (saved) return;
    saved = true;
    const name = input.value.trim();

    // Восстанавливаем вид
    infoNameEl.textContent = name || originalText;
    infoNameEl.style.cursor = 'pointer';
    infoNameEl.title = 'Нажмите для редактирования';
    infoNameEl.classList.add('group-edit-name');
    infoNameEl.onclick = () => startInlineGroupNameEdit(chatId, window.app.chats.find(c => String(c.id) === String(chatId))?.name || '');

    if (!name || name === originalText) return;
    try {
      const updated = await apiUpdateGroupInfo(chatId, name);
      const chat = window.app.chats.find(c => String(c.id) === String(chatId));
      if (chat) { chat.name = updated.name || name; }
      document.getElementById('active-chat-name').textContent = name;
      infoNameEl.textContent = name;
      renderChats();
      window.app.notify('Название обновлено ✓', 'success');
    } catch (err) {
      infoNameEl.textContent = originalText;
      window.app.notify('Ошибка: ' + err.message, 'error');
    }
  };

  const cancel = () => {
    if (saved) return;
    saved = true;
    infoNameEl.textContent = originalText;
    infoNameEl.style.cursor = 'pointer';
    infoNameEl.title = 'Нажмите для редактирования';
    infoNameEl.classList.add('group-edit-name');
    infoNameEl.onclick = () => startInlineGroupNameEdit(chatId, originalText);
  };

  input.addEventListener('keydown', (e) => {
    if (e.key === 'Enter')  { e.preventDefault(); save(); }
    if (e.key === 'Escape') { cancel(); }
  });
  input.addEventListener('blur', save);
}

// ── Добавление участника через модал поиска ──────────────────────────────

function handleAddMemberBtn() {
  animatePlusBtn(document.getElementById('add-member-plus-btn'));
  openAddMemberModal();
}

function openAddMemberModal() {
  // Открываем модал поиска в режиме "добавить участника"
  window._addMemberMode = true;
  window._ncSelected = [];
  ncRenderChips();

  // Сохраняем ID текущих участников чтобы исключить их из поиска
  const activeChat = window.app?.chats?.find(c => String(c.id) === String(window.app?.activeChatId));
  window._ncExcludeIds = new Set(
      (activeChat?.members || []).map(m => String(m.id))
  );

  const overlay   = document.getElementById('new-chat-overlay');
  const modal     = document.getElementById('new-chat-modal');
  const input     = document.getElementById('user-search-input');
  const results   = document.getElementById('search-results');
  const subtitle  = document.getElementById('new-chat-subtitle');
  const footer    = document.getElementById('new-chat-footer');
  const groupWrap = document.getElementById('group-name-wrap');

  if (input)    { input.value = ''; }
  if (results)  { results.innerHTML = ''; }
  if (subtitle) { subtitle.textContent = 'Найди пользователя для добавления'; }
  if (footer)   { footer.classList.add('hidden'); }
  if (groupWrap){ groupWrap.classList.add('hidden'); }

  // Меняем заголовок модала
  const modalTitle = document.querySelector('#new-chat-modal h2');
  if (modalTitle) modalTitle.textContent = 'Добавить участника';

  overlay.classList.remove('hidden');
  setTimeout(() => {
    overlay.classList.remove('opacity-0');
    modal.classList.remove('scale-95');
    if (input) input.focus();
  }, 10);
}

function closeAddMemberModal() {
  window._addMemberMode = false;
  // Восстанавливаем заголовок
  const modalTitle = document.querySelector('#new-chat-modal h2');
  if (modalTitle) modalTitle.textContent = 'Новый чат';
  const subtitle = document.getElementById('new-chat-subtitle');
  if (subtitle) subtitle.textContent = 'Найди пользователя для начала общения';
  closeNewChatModal();
}

async function removeGroupMember(chatId, userId) {
  try {
    await apiRemoveChatMember(chatId, userId);
    const chat = window.app.chats.find(c => String(c.id) === String(chatId));
    if (chat && chat.members) {
      chat.members = chat.members.filter(m => String(m.id) !== String(userId));
    }
    window.app.notify('Участник удалён', 'success');
    renderChatHeader();
    renderChats();
  } catch (err) {
    window.app.notify('Ошибка: ' + err.message, 'error');
  }
}

async function addGroupMember(chatId) {
  const input = document.getElementById('add-member-username-input');
  if (!input) return;
  const username = input.value.trim();
  if (!username) return;
  try {
    await apiAddChatMember(chatId, username);
    input.value = '';
    window.app.notify(`${username} добавлен в группу ✓`, 'success');
    await apiLoadChats();
    renderChatHeader();
  } catch (err) {
    window.app.notify('Ошибка: ' + err.message, 'error');
  }
}

// Просмотр аватара участника группы
function openMemberAvatarViewer(member) {
  const circleEl = document.getElementById('avatar-viewer-circle');
  if (!circleEl) return;
  circleEl.style.overflow = 'hidden';
  if (member.avatar_url) {
    circleEl.innerHTML = `<img src="${member.avatar_url}" style="width:100%;height:100%;object-fit:cover;border-radius:50%;">`;
  } else {
    circleEl.textContent = (member.username || '?')[0].toUpperCase();
  }
  document.getElementById('avatar-viewer-name').textContent   = member.username || '';
  document.getElementById('avatar-viewer-status').textContent = member.full_name || member.status || '';

  // Определяем онлайн-статус
  const app = window.app;
  let isOnline = typeof member.is_online !== 'undefined' ? member.is_online : false;
  if (!isOnline && member.id) {
    const chat = (app.chats || []).find(c => String(c.interlocutor_id) === String(member.id));
    if (chat) isOnline = !!chat.is_online;
    if (!isOnline) {
      for (const c of (app.chats || [])) {
        const m = (c.members || []).find(m => String(m.id) === String(member.id));
        if (m) { isOnline = !!m.is_online; break; }
      }
    }
  }

  const dot   = document.getElementById('avatar-viewer-online-dot');
  const label = document.getElementById('avatar-viewer-online-label');
  if (dot && label) {
    dot.style.display    = 'block';
    label.style.display  = 'inline-block';
    dot.style.background = isOnline ? '#22c55e' : '#9ca3af';
    label.textContent    = isOnline ? 'онлайн' : 'офлайн';
    label.style.background  = isOnline ? 'rgba(34,197,94,0.15)' : 'rgba(156,163,175,0.15)';
    label.style.color        = isOnline ? '#16a34a' : '#6b7280';
  }

  document.getElementById('avatar-viewer-overlay').classList.remove('hidden');
}