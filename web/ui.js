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
    setAvatarEl(document.getElementById('profile-avatar'), user);
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
        <div class="mx-3 my-1 border-t border-custom opacity-50"></div>
        <button onclick="executeDeleteMessage('${messageId}')"
            class="w-full text-left px-4 py-2.5 text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-950/20 transition flex items-center gap-2">
            <svg class="w-4 h-4 opacity-70" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
            </svg>
            Удалить
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
  const bubble  = el.querySelector('.message-bubble') || el;
  const r       = bubble.getBoundingClientRect();
  const isMe    = el.classList.contains('justify-end');
  const isDark  = document.body.classList.contains('theme-gray');

  // ── Цвета в зависимости от темы и стороны ──
  const baseColor = isMe
      ? (isDark ? [29, 78, 216] : [59, 130, 246])
      : (isDark ? [55, 65,  81] : [226, 232, 240]);

  // ── Canvas поверх всего ──
  const cv  = document.createElement('canvas');
  cv.style.cssText = `position:fixed;inset:0;pointer-events:none;z-index:9999;`;
  cv.width  = window.innerWidth;
  cv.height = window.innerHeight;
  document.body.appendChild(cv);
  const ctx = cv.getContext('2d');

  // ── Параметры тайлов ──
  const COLS  = 18;
  const ROWS  = Math.max(4, Math.round(COLS * r.height / r.width));
  const tw    = r.width  / COLS;
  const th    = r.height / ROWS;
  const cx    = r.left + r.width  / 2;
  const cy    = r.top  + r.height / 2;

  // ── Рисуем пузырёк как набор тайлов ──
  // Каждый тайл = закрашенный прямоугольник цвета пузырька
  // (без html2canvas — эмулируем внешний вид)
  const tiles = [];
  for (let row = 0; row < ROWS; row++) {
    for (let col = 0; col < COLS; col++) {
      const tx = r.left + col * tw;
      const ty = r.top  + row * th;

      // Расстояние от центра — определяет задержку волны
      const dx     = (col + 0.5) / COLS - 0.5;
      const dy     = (row + 0.5) / ROWS - 0.5;
      const dist   = Math.sqrt(dx*dx + dy*dy) * 2; // 0..1

      // Угол вылета — от центра наружу + закрутка
      const angle  = Math.atan2(dy, dx) + (Math.random() - 0.5) * 0.9;
      const speed  = 2.5 + Math.random() * 4.5;

      // Лёгкая вариация цвета — имитирует текстуру
      const bright = 0.85 + Math.random() * 0.3;
      const [r0,g0,b0] = baseColor;
      const color  = `rgb(${Math.round(r0*bright)},${Math.round(g0*bright)},${Math.round(b0*bright)})`;

      tiles.push({
        // Позиция
        x: tx, y: ty,
        // Размер тайла
        w: tw + 0.5, h: th + 0.5,
        // Скорость
        vx: Math.cos(angle) * speed,
        vy: Math.sin(angle) * speed - 1.8,
        // Вращение
        rot: 0,
        rotV: (Math.random() - 0.5) * 0.25,
        // Задержка старта: волна от центра
        delay: dist * 18 + Math.random() * 8,
        // Жизнь
        life: 1,
        decay: 0.016 + Math.random() * 0.014,
        color,
        started: false,
      });
    }
  }

  // ── Скрываем оригинальный элемент плавно ──
  bubble.style.transition = 'opacity 0.08s ease';
  bubble.style.opacity    = '0';

  let frame  = 0;
  let done   = false;

  function tick() {
    ctx.clearRect(0, 0, cv.width, cv.height);
    frame++;
    let alive = 0;

    for (const t of tiles) {
      if (frame < t.delay) {
        // Ещё не стартовал — рисуем на месте (исходник виден)
        alive++;
        continue;
      }

      if (!t.started) {
        t.started = true;
      }

      if (t.life <= 0) continue;
      alive++;

      // Физика
      t.x   += t.vx;
      t.y   += t.vy;
      t.vy  += 0.18;   // гравитация
      t.vx  *= 0.975;  // сопротивление
      t.life -= t.decay;
      t.rot  += t.rotV;

      const alpha = Math.max(0, t.life);
      const scale = 0.4 + alpha * 0.6;
      const hw    = t.w * scale / 2;
      const hh    = t.h * scale / 2;

      ctx.save();
      ctx.globalAlpha = alpha;
      ctx.translate(t.x + t.w / 2, t.y + t.h / 2);
      ctx.rotate(t.rot);

      // Основной тайл
      ctx.fillStyle = t.color;
      ctx.fillRect(-hw, -hh, hw*2, hh*2);

      // Блик — имитирует объём
      if (alpha > 0.3) {
        ctx.fillStyle = `rgba(255,255,255,${alpha * 0.18})`;
        ctx.fillRect(-hw, -hh, hw * 2, hh * 0.45);
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