// ─── app.js — главный класс приложения ────────────────────────────────────

class AlphaApp {
  constructor() {
    this.token        = localStorage.getItem('alpha_token');
    this.currentUser  = JSON.parse(localStorage.getItem('alpha_user') || 'null');
    this.activeChatId = null;
    this.socket       = null;
    this.chats        = [];
    this.messages     = [];
    this.userStatusMap = {}; // { userId: { online: bool, status: string } }

    const userId = this.currentUser?.id || 'default';
    this.theme   = localStorage.getItem(`alpha_theme_${userId}`) || 'gray';

    // Присваиваем window.app ДО init() — через setter в index.html
    window.app = this;

    this.init();
    this.applyTheme();

    // После инициализации убираем прелоадер, чтобы не мигали все окна при загрузке
    if (document && document.body) {
      document.body.style.visibility = '';
    }
  }

  // ── Инициализация ──────────────────────────────────────────────────────

  init() {
    if (this.token) {
      if (!this.currentUser || !this.currentUser.id) {
        console.warn('Current user has no ID, clearing local storage');
        this.logout();
        return;
      }
      this.showChat();
      setTimeout(() => document.dispatchEvent(new CustomEvent('app:authenticated')), 300);
    } else {
      this.showLanding();
    }

    // Enter в поле сообщения
    document.addEventListener('keydown', (e) => {
      if (e.target.id === 'message-input' && e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        const form = document.getElementById('message-form');
        if (form && typeof handleSendMessage === 'function') {
          handleSendMessage({ preventDefault: () => {} });
        } else {
          this.sendMessage();
        }
      }
    });

    // Авто-расширение textarea
    document.addEventListener('input', (e) => {
      if (e.target.id !== 'message-input') return;
      const ta = e.target;
      ta.style.height = 'auto';
      ta.style.height = Math.min(ta.scrollHeight, 160) + 'px';
      ta.classList.toggle('has-text', ta.value.trim().length > 0);
    });

    document.addEventListener('_msgSent', () => {
      const ta = document.getElementById('message-input');
      if (!ta) return;
      ta.style.height = '';
      ta.classList.remove('has-text');
    });
  }

  // ── Роутинг ────────────────────────────────────────────────────────────

  showLanding() {
    document.getElementById('landing-page').classList.remove('hidden');
    document.getElementById('main-chat').classList.add('hidden');
  }

  showChat() {
    document.getElementById('landing-page').classList.add('hidden');
    document.getElementById('main-chat').classList.remove('hidden');
    // Восстанавливаем dock (мог быть скрыт в гостевом режиме)
    const dock = document.getElementById('bottom-dock');
    if (dock) {
      dock.classList.remove('guest-hidden');
      dock.style.display = '';
    }
    loadUserData();
    this.loadChats();
    this.connectWebSocket();
  }

  // ── Тема ───────────────────────────────────────────────────────────────

  applyTheme() {
    const themes = ['gray', 'twilight', 'dawn', 'sunset', 'coral', 'ocean', 'mint', 'dark-gold', 'jungle', 'blood', 'cyberpunk', 'glacier', 'cosmos', 'amethyst', 'desert', 'volcano'];
    themes.forEach(t => document.body.classList.toggle(`theme-${t}`, this.theme === t));
  }

  toggleTheme() {
    const order = ['light', 'gray', 'twilight', 'dawn', 'sunset', 'coral', 'ocean', 'mint', 'dark-gold', 'jungle', 'blood', 'cyberpunk', 'glacier', 'cosmos', 'amethyst', 'desert', 'volcano'];
    const idx = order.indexOf(this.theme);
    this.theme = order[(idx + 1) % order.length];
    localStorage.setItem(`alpha_theme_${this.currentUser?.id || 'default'}`, this.theme);
    this.applyTheme();
  }

  // ── Авторизация ────────────────────────────────────────────────────────

  // auth() оставлен для обратной совместимости, основная логика в Auth модуле (app.js низ файла)
  auth(event, type) { event.preventDefault(); }

  // Вызывается Auth модулем после успешной авторизации
  onAuthSuccess(token, user) {
    this.token       = token;
    this.currentUser = user;
    localStorage.setItem('alpha_token', token);
    localStorage.setItem('alpha_user', JSON.stringify(user));
    this.notify('Добро пожаловать!', 'success');
    closeAuthModal();
    this.showChat();
    document.dispatchEvent(new CustomEvent('app:authenticated'));
  }

  logout() {
    localStorage.clear();
    this.token       = null;
    this.currentUser = null;
    if (this.socket) this.socket.close();
    window.location.reload();
  }

  // ── Делегирование к модулям ────────────────────────────────────────────

  loadChats()               { return apiLoadChats(); }
  loadMessages(chatId)      { return apiLoadMessages(chatId); }
  sendMessage()             { return apiSendMessage(); }
  searchUsers(query)        { return apiSearchUsers(query); }
  createPrivateChat(userId) { return apiCreatePrivateChat(userId); }
  apiFetch(url, options)    { return apiFetch(url, options); }
  connectWebSocket()        { return connectWebSocket(); }

  // ── Уведомления ────────────────────────────────────────────────────────

  notify(text, type = 'info') {
    const center = document.getElementById('notification-center');
    const note   = document.createElement('div');
    note.className = `app-toast app-toast-${type}`;
    note.textContent = text;
    center.appendChild(note);
    setTimeout(() => note.remove(), 5000);
  }
}

// ── Глобальный экземпляр ───────────────────────────────────────────────────
new AlphaApp();
// ── Auth модуль ────────────────────────────────────────────────────────────────

const Auth = (() => {
  let _login  = '';
  let _method = '';
  let _regData = {};

  function _el(id)   { return document.getElementById(id); }
  function _show(id) { _el(id)?.classList.remove('hidden'); }
  function _hide(id) { _el(id)?.classList.add('hidden'); }
  function _err(msg) {
    const el = _el('auth-error');
    if (!el) return;
    el.textContent = msg;
    msg ? _show('auth-error') : _hide('auth-error');
  }
  function _step(name) {
    ['login','register','code','reset-send','reset-code','reset-confirm'].forEach(s => _hide(`auth-step-${s}`));
    _show(`auth-step-${name}`);
    _err('');
  }

  // Вызывается из openAuthModal('login') или openAuthModal('register')
  function open(mode) {
    _step(mode === 'register' ? 'register' : 'login');
    if (mode === 'login') _el('auth-login-input')?.focus();
    else                  _el('reg-login')?.focus();
  }

  // ── Вход ────────────────────────────────────────────────────────────────

  async function submitLogin() {
    const login    = _el('auth-login-input')?.value?.trim();
    const password = _el('auth-password-input')?.value;
    if (!login)    { _err('Введите email или телефон'); return; }
    if (!password) { _err('Введите пароль'); return; }
    _err('');
    const btn = _el('auth-btn-login');
    btn.disabled = true; btn.textContent = 'Вход...';
    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ login, password })
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Неверный логин или пароль');
      window.app.onAuthSuccess(data.token, data.user);
    } catch (e) {
      _err(e.message);
    } finally {
      btn.disabled = false; btn.textContent = 'Войти';
    }
  }

  // ── Регистрация: валидация и отправка кода ───────────────────────────────

  async function submitRegister() {
    const login     = _el('reg-login')?.value?.trim();
    const username  = _el('reg-username')?.value?.trim();
    const fullName  = _el('reg-fullname')?.value?.trim();
    const birthDate = _el('reg-birthdate')?.value;
    const location  = _el('reg-location')?.value?.trim();
    const extra     = _el('reg-extra')?.value?.trim();
    const password  = _el('reg-password')?.value;
    const password2 = _el('reg-password2')?.value;

    if (!login)             { _err('Введите email или телефон'); return; }
    if (!username)          { _err('Введите имя пользователя'); return; }
    if (username.length < 3){ _err('Имя пользователя минимум 3 символа'); return; }
    if (!password)          { _err('Введите пароль'); return; }
    if (password.length < 6){ _err('Пароль минимум 6 символов'); return; }
    if (password !== password2){ _err('Пароли не совпадают'); return; }
    _err('');

    _regData = { username, full_name: fullName, birth_date: birthDate, location, extra_contact: extra, password, password2 };

    const btn = _el('auth-btn-send-code');
    btn.disabled = true; btn.textContent = 'Отправка кода...';
    try {
      const res = await fetch('/api/auth/send', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ login })
      });
      const data = await res.json();
      if (!res.ok) {
        if (data.exists) { _err('Аккаунт уже существует — войдите'); return; }
        throw new Error(data.error || 'Ошибка отправки');
      }
      _method = data.method;
      _login  = data.login;
      _el('auth-dest-display').textContent = _login;
      _el('auth-code-input').value = '';
      _setVerifyReady(false);
      _step('code');
      _el('auth-code-input')?.focus();
    } catch (e) {
      _err(e.message);
    } finally {
      btn.disabled = false; btn.textContent = 'Получить код подтверждения';
    }
  }

  // ── Код подтверждения ────────────────────────────────────────────────────

  function onCodeInput(val) {
    _setVerifyReady(val.replace(/\D/g, '').length === 6);
    if (val.replace(/\D/g, '').length === 6) submitCode();
  }

  function _setVerifyReady(ready) {
    const btn = _el('auth-btn-verify');
    if (!btn) return;
    btn.disabled = !ready;
    btn.classList.toggle('opacity-50', !ready);
    btn.classList.toggle('cursor-not-allowed', !ready);
  }

  async function submitCode() {
    const code = _el('auth-code-input')?.value?.replace(/\D/g, '');
    if (!code || code.length !== 6) { _err('Введите 6-значный код'); return; }
    _err('');
    const btn = _el('auth-btn-verify');
    btn.disabled = true; btn.textContent = 'Проверка...';
    try {
      const verifyRes = await fetch('/api/auth/verify-code', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ login: _login, code })
      });
      const verifyData = await verifyRes.json();
      if (!verifyRes.ok) throw new Error(verifyData.error || 'Неверный код');

      const regRes = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ confirm_token: verifyData.confirm_token, login: _login, ..._regData })
      });
      const regData = await regRes.json();
      if (!regRes.ok) throw new Error(regData.error || 'Ошибка регистрации');
      window.app.onAuthSuccess(regData.token, regData.user);
    } catch (e) {
      _err(e.message);
      _setVerifyReady(true);
    } finally {
      btn.textContent = 'Подтвердить и создать аккаунт';
    }
  }

  async function resendCode() {
    const btn = _el('auth-btn-resend');
    btn.disabled = true; btn.textContent = 'Отправка...';
    _el('auth-code-input').value = '';
    _setVerifyReady(false);
    try {
      const res = await fetch('/api/auth/send', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ login: _login })
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error);
      btn.textContent = 'Отправлено ✓';
      setTimeout(() => { btn.textContent = 'Отправить повторно'; btn.disabled = false; }, 60000);
    } catch(e) {
      _err(e.message);
      btn.disabled = false; btn.textContent = 'Отправить повторно';
    }
  }

  // ── Сброс пароля ──────────────────────────────────────────────────────────
  let _resetLogin = '';
  let _resetToken = '';

  async function resetSend() {
    const login = _el('reset-login-input')?.value?.trim();
    if (!login) { _err('Введите email или телефон'); return; }
    _err('');
    const btn = _el('auth-btn-reset-send');
    btn.disabled = true; btn.textContent = 'Отправка...';
    try {
      const res = await fetch('/api/auth/reset/send', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ login })
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Ошибка');
      _resetLogin = login;
      _step('reset-code');
    } catch (e) {
      _err(e.message);
    } finally {
      btn.disabled = false; btn.textContent = 'Отправить код';
    }
  }

  async function resetVerify() {
    const code = _el('reset-code-input')?.value?.trim();
    if (!code) { _err('Введите код'); return; }
    _err('');
    const btn = _el('auth-btn-reset-verify');
    btn.disabled = true; btn.textContent = 'Проверка...';
    try {
      const res = await fetch('/api/auth/reset/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ login: _resetLogin, code })
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Неверный код');
      _resetToken = data.reset_token;
      _step('reset-confirm');
    } catch (e) {
      _err(e.message);
    } finally {
      btn.disabled = false; btn.textContent = 'Подтвердить';
    }
  }

  async function resetConfirm() {
    const p1 = _el('reset-password-input')?.value;
    const p2 = _el('reset-password2-input')?.value;
    if (!p1 || !p2) { _err('Заполните оба поля'); return; }
    if (p1 !== p2)   { _err('Пароли не совпадают'); return; }
    if (p1.length < 6) { _err('Пароль должен быть не менее 6 символов'); return; }
    _err('');
    const btn = _el('auth-btn-reset-confirm');
    btn.disabled = true; btn.textContent = 'Сохранение...';
    try {
      const res = await fetch('/api/auth/reset/confirm', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ reset_token: _resetToken, password: p1, password2: p2 })
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Ошибка');
      _step('login');
      _err('');
      // Показываем уведомление об успехе
      if (_el('auth-error')) {
        const el = _el('auth-error');
        el.textContent = 'Пароль успешно изменён — войдите с новым паролем';
        el.style.color = '#4ade80';
        _show('auth-error');
        setTimeout(() => { _hide('auth-error'); el.style.color = ''; }, 4000);
      }
    } catch (e) {
      _err(e.message);
    } finally {
      btn.disabled = false; btn.textContent = 'Сохранить пароль';
    }
  }

  function _bind() {
    _el('auth-btn-login')          ?.addEventListener('click',   submitLogin);
    _el('auth-password-input')     ?.addEventListener('keydown', e => e.key === 'Enter' && submitLogin());
    _el('auth-login-input')        ?.addEventListener('keydown', e => e.key === 'Enter' && _el('auth-password-input')?.focus());
    _el('auth-btn-send-code')      ?.addEventListener('click',   submitRegister);
    _el('auth-switch-to-login')    ?.addEventListener('click',   () => _step('login'));
    _el('auth-switch-to-register') ?.addEventListener('click',   () => _step('register'));
    _el('auth-code-input')         ?.addEventListener('input',   e => onCodeInput(e.target.value));
    _el('auth-btn-verify')         ?.addEventListener('click',   submitCode);
    _el('auth-btn-resend')         ?.addEventListener('click',   resendCode);
    _el('auth-back-from-code')     ?.addEventListener('click',   () => _step('register'));
    // Сброс пароля
    _el('auth-forgot-link')           ?.addEventListener('click', () => { _step('reset-send'); if (_el('reset-login-input')) _el('reset-login-input').value = ''; });
    _el('reset-back-to-login')        ?.addEventListener('click', () => _step('login'));
    _el('auth-btn-reset-send')        ?.addEventListener('click', resetSend);
    _el('reset-login-input')          ?.addEventListener('keydown', e => e.key === 'Enter' && resetSend());
    _el('auth-btn-reset-verify')      ?.addEventListener('click', resetVerify);
    _el('reset-code-input')           ?.addEventListener('keydown', e => e.key === 'Enter' && resetVerify());
    _el('auth-btn-reset-confirm')     ?.addEventListener('click', resetConfirm);
    _el('reset-password2-input')      ?.addEventListener('keydown', e => e.key === 'Enter' && resetConfirm());
    _el('reset-resend-code')          ?.addEventListener('click', async () => {
      if (!_resetLogin) return;
      try {
        const res = await fetch('/api/auth/reset/send', {
          method: 'POST', headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ login: _resetLogin })
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error);
        _err('Код отправлен повторно');
      } catch(e) { _err(e.message); }
    });
  }

  document.addEventListener('DOMContentLoaded', _bind);

  return { open };
})();

// openAuthModal вызывается из HTML кнопок
function openAuthModal(mode = 'login') {
  const overlay = document.getElementById('auth-overlay');
  const modal   = document.getElementById('auth-modal');
  overlay.classList.remove('hidden');
  setTimeout(() => { overlay.classList.add('opacity-100'); modal.classList.remove('scale-95'); }, 10);
  Auth.open(mode);
}