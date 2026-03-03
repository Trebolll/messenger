// ─── app.js — главный класс приложения ────────────────────────────────────

class AlphaApp {
  constructor() {
    this.token        = localStorage.getItem('alpha_token');
    this.currentUser  = JSON.parse(localStorage.getItem('alpha_user') || 'null');
    this.activeChatId = null;
    this.socket       = null;
    this.chats        = [];
    this.messages     = [];

    const userId = this.currentUser?.id || 'default';
    this.theme   = localStorage.getItem(`alpha_theme_${userId}`) || 'light';

    // Присваиваем window.app ДО init() — через setter в index.html
    window.app = this;

    this.init();
    this.applyTheme();
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
    loadUserData();
    this.loadChats();
    this.connectWebSocket();
  }

  // ── Тема ───────────────────────────────────────────────────────────────

  applyTheme() {
    const themes = ['gray', 'twilight', 'dawn', 'sunset', 'coral', 'ocean', 'mint'];
    themes.forEach(t => document.body.classList.toggle(`theme-${t}`, this.theme === t));
  }

  toggleTheme() {
    const order = ['light', 'gray', 'twilight', 'dawn', 'sunset', 'coral', 'ocean', 'mint'];
    const idx = order.indexOf(this.theme);
    this.theme = order[(idx + 1) % order.length];
    localStorage.setItem(`alpha_theme_${this.currentUser?.id || 'default'}`, this.theme);
    this.applyTheme();
  }

  // ── Авторизация ────────────────────────────────────────────────────────

  async auth(event, type) {
    event.preventDefault();
    const form    = event.target;
    const data    = Object.fromEntries(new FormData(form).entries());
    const errorEl = document.getElementById(`${type}-error`);
    try {
      const response = await fetch(`/api/${type}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });
      const result = await response.json();
      if (!response.ok) throw new Error(result.error || 'Ошибка авторизации');

      this.token       = result.token;
      this.currentUser = result.user || { email: data.email, username: data.username || data.email.split('@')[0] };
      localStorage.setItem('alpha_token', this.token);
      localStorage.setItem('alpha_user', JSON.stringify(this.currentUser));

      this.notify('Успешно!', 'success');
      closeAuthModal();
      this.showChat();
    } catch (err) {
      errorEl.textContent = err.message;
      errorEl.classList.remove('hidden');
    }
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
    note.className = `p-4 rounded-2xl shadow-xl border text-sm font-medium fade-in ${
        type === 'success' ? 'bg-green-500/10 border-green-500/20 text-green-500'  :
            type === 'error'   ? 'bg-red-500/10 border-red-500/20 text-red-500'        :
                'bg-custom-main border-custom text-custom-main'
    }`;
    note.textContent = text;
    center.appendChild(note);
    setTimeout(() => note.remove(), 5000);
  }
}

// ── Глобальный экземпляр ───────────────────────────────────────────────────
new AlphaApp();