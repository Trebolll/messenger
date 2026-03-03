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

  // Показать/скрыть элемент с нужным display (flex или block)
  function show(elem, displayType) {
    if (!elem) return;
    elem.style.display = displayType || 'block';
  }
  function hide(elem) {
    if (!elem) return;
    elem.style.display = 'none';
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
      // Левая панель: видна только если чаты открыты
      if (leftPanel) { state.chatsVisible ? show(leftPanel, 'flex') : hide(leftPanel); }
      if (chatsSidebar) show(chatsSidebar, 'flex');
      if (viewHome)     hide(viewHome);
      if (viewChat)     show(viewChat, 'flex');
      if (noChat)       hide(noChat);
      if (inputArea)    show(inputArea, 'flex');
      hideFeedRight();
    } else {
      document.body.classList.remove('chat-open');
      if (viewChat)  hide(viewChat);
      if (inputArea) hide(inputArea);
      if (noChat)    show(noChat);

      if (state.chatsVisible && state.feedVisible) {
        // Чаты слева + лента справа
        if (leftPanel)    show(leftPanel, 'flex');
        if (chatsSidebar) show(chatsSidebar, 'flex');
        if (viewHome)     hide(viewHome);
        showFeedRight();
      } else if (state.chatsVisible) {
        // Только чаты
        if (leftPanel)    show(leftPanel, 'flex');
        if (chatsSidebar) show(chatsSidebar, 'flex');
        if (viewHome)     hide(viewHome);
        hideFeedRight();
      } else {
        // Только лента (дефолт)
        state.feedVisible = true;
        if (leftPanel)    show(leftPanel, 'flex');
        if (chatsSidebar) hide(chatsSidebar);
        if (viewHome)     show(viewHome, 'flex');
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
    panel.style.cssText = 'display:none;flex:1;height:100%;overflow-y:auto;background:var(--bg-main);';
    mainChat.insertBefore(panel, viewChat);
  }

  function showFeedRight() {
    ensureFeedRightPanel();
    var panel    = el('feed-right-panel');
    var viewHome = el('view-home');
    if (!panel) return;
    panel.style.display = 'flex';
    panel.style.flexDirection = 'column';
    if (viewHome && panel.children.length === 0) {
      Array.from(viewHome.children).forEach(function (child) {
        panel.appendChild(child.cloneNode(true));
      });
    }
  }

  function hideFeedRight() {
    var panel = el('feed-right-panel');
    if (panel) panel.style.display = 'none';
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
  window.switchView = function (view) {
    if (view === 'chat') {
      state.feedWasVisible = state.feedVisible;
      state.chatOpen = true;
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
      if (window.innerHeight - e.clientY <= TRIGGER_PX) showDock();
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

    // Применить layout сразу если уже в чат-режиме
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