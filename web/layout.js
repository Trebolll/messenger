/* ══════════════════════════════════════════════════════════════
   ══ layout.js — навигация и состояние раскладки
   ══════════════════════════════════════════════════════════════ */
(function () {

  var state = {
    chatsVisible:   false,
    feedVisible:    true,
    chatOpen:       false,
    feedWasVisible: true
  };

  function el(id) { return document.getElementById(id); }

  function isMainChatVisible() {
    var mc = el('main-chat');
    return mc && !mc.classList.contains('hidden');
  }

  // ── Layout ──────────────────────────────────────────────────
  function applyLayout() {
    var leftPanel    = el('left-panel');
    var chatsSidebar = el('chats-sidebar');
    var viewHome     = el('view-home');
    var viewChat     = el('view-chat');
    var noChat       = el('no-chat-selected');
    var inputArea    = el('input-area');

    if (state.chatOpen) {
      document.body.classList.add('chat-open');
      if (leftPanel)    leftPanel.classList.toggle('hidden', !state.chatsVisible);
      if (chatsSidebar) chatsSidebar.classList.remove('hidden');
      if (viewHome)     viewHome.classList.add('hidden');
      if (viewChat)     viewChat.classList.remove('hidden');
      if (noChat)       noChat.classList.add('hidden');
      if (inputArea)    inputArea.classList.remove('hidden');
      hideFeedRight();
    } else {
      document.body.classList.remove('chat-open');
      if (viewChat)  viewChat.classList.add('hidden');
      if (inputArea) inputArea.classList.add('hidden');
      if (noChat)    noChat.classList.remove('hidden');

      if (state.chatsVisible && state.feedVisible) {
        if (leftPanel)    leftPanel.classList.remove('hidden');
        if (chatsSidebar) chatsSidebar.classList.remove('hidden');
        if (viewHome)     viewHome.classList.add('hidden');
        showFeedRight();
      } else if (state.chatsVisible) {
        if (leftPanel)    leftPanel.classList.remove('hidden');
        if (chatsSidebar) chatsSidebar.classList.remove('hidden');
        if (viewHome)     viewHome.classList.add('hidden');
        hideFeedRight();
      } else {
        state.feedVisible = true;
        if (leftPanel)    leftPanel.classList.remove('hidden');
        if (chatsSidebar) chatsSidebar.classList.add('hidden');
        if (viewHome)     viewHome.classList.remove('hidden');
        hideFeedRight();
      }
    }

    updateDockActive();
  }

  // ── Правая панель ленты ─────────────────────────────────────
  function ensureFeedRightPanel() {
    if (el('feed-right-panel')) return;
    var mainChat = el('main-chat');
    var viewChat = el('view-chat');
    if (!mainChat || !viewChat) return;
    var panel = document.createElement('div');
    panel.id = 'feed-right-panel';
    panel.className = 'flex-grow h-full bg-custom-main overflow-y-auto hidden';
    mainChat.insertBefore(panel, viewChat);
  }

  function showFeedRight() {
    ensureFeedRightPanel();
    var panel    = el('feed-right-panel');
    var viewHome = el('view-home');
    if (!panel) return;
    panel.classList.remove('hidden');
    if (viewHome && panel.children.length === 0) {
      Array.from(viewHome.children).forEach(function (child) {
        panel.appendChild(child.cloneNode(true));
      });
    }
  }

  function hideFeedRight() {
    var panel = el('feed-right-panel');
    if (panel) panel.classList.add('hidden');
  }

  // ── Публичное API ───────────────────────────────────────────
  window.switchLeftTab = function (tab) {
    if (tab === 'chats') {
      state.chatsVisible = !state.chatsVisible;
      if (!state.chatOpen) state.feedVisible = true;
    } else {
      if (state.chatOpen) { closeChat(true); return; }
      state.feedVisible = true;
    }
    applyLayout();
  };

  // switchView — совместимость с api.js
  // 'chat' — открыть чат (вызывается из apiLoadMessages)
  // 'home' — закрыть чат, показать ленту
  window.switchView = function (view) {
    if (view === 'chat') {
      state.feedWasVisible = state.feedVisible;
      state.chatOpen = true;
      // Если чаты не открыты — скрываем левую панель в режиме чата
      applyLayout();
    } else if (view === 'home') {
      closeChat(true);
    }
  };

  window.handleBackBtn      = function () { closeChat(false); };
  window.toggleChatsSidebar = function () { window.switchLeftTab('chats'); };

  window.onChatOpened = function () {
    state.feedWasVisible = state.feedVisible;
    state.chatOpen = true;
    applyLayout();
  };

  function closeChat(showFeed) {
    state.chatOpen = false;
    document.querySelectorAll('.chat-list-item').forEach(function (i) {
      i.classList.remove('active');
    });
    if (window.app) window.app.activeChatId = null;
    state.feedVisible = showFeed ? true : state.feedWasVisible;
    applyLayout();
  }

  // ── Dock ────────────────────────────────────────────────────
  function updateDockActive() {
    var dockHome  = el('dock-home');
    var dockChats = el('dock-chats');
    if (dockHome)  dockHome.classList.toggle('dock-active',  state.feedVisible && !state.chatOpen);
    if (dockChats) dockChats.classList.toggle('dock-active', state.chatsVisible);
  }

  var TRIGGER_PX  = 80;
  var HIDE_DELAY  = 5000;
  var hideTimer   = null;
  var dockHovered = false;

  function showDock() {
    clearTimeout(hideTimer);
    var dock = el('bottom-dock');
    if (dock) dock.classList.add('visible');
  }

  function scheduleHide() {
    if (dockHovered) return;
    clearTimeout(hideTimer);
    hideTimer = setTimeout(function () {
      var dock = el('bottom-dock');
      if (dock) dock.classList.remove('visible');
    }, HIDE_DELAY);
  }

  function initDock() {
    var dock = el('bottom-dock');
    if (!dock) return;

    document.addEventListener('mousemove', function (e) {
      if (!isMainChatVisible()) return;
      if (state.chatOpen) return;
      if (window.innerHeight - e.clientY <= TRIGGER_PX) {
        showDock();
      }
    });

    dock.addEventListener('mouseenter', function () {
      dockHovered = true;
      showDock();
    });
    dock.addEventListener('mouseleave', function () {
      dockHovered = false;
      scheduleHide();
    });
  }

  // ── Патч AlphaApp ───────────────────────────────────────────
  function patchApp() {
    var app = window.app;
    if (!app || app._layoutPatched) return;
    app._layoutPatched = true;

    var origShowChat = app.showChat.bind(app);
    app.showChat = function () {
      origShowChat();
      state.chatsVisible = false;
      state.feedVisible  = true;
      state.chatOpen     = false;
      applyLayout();
    };
  }

  // ── Init ────────────────────────────────────────────────────
  function init() {
    initDock();
    updateDockActive();

    if (isMainChatVisible()) {
      state.chatsVisible = false;
      state.feedVisible  = true;
      state.chatOpen     = false;
      applyLayout();
    }

    var attempts = 0;
    var iv = setInterval(function () {
      attempts++;
      patchApp();
      if (attempts > 50) clearInterval(iv);
    }, 100);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }

})();