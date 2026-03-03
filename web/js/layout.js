/* ══════════════════════════════════════════════════════════════
   ══ layout.js — навигация и состояние раскладки
   ══════════════════════════════════════════════════════════════ */
(function () {

  var state = {
    chatsVisible:   false,
    feedVisible:    true,
    chatOpen:       false,
    infoOpen:       false,
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
    var mainChat     = el('main-chat');
    var leftPanel    = el('left-panel');
    var chatsSidebar = el('chats-sidebar');
    var viewHome     = el('view-home');
    var viewChat     = el('view-chat');
    var noChat       = el('no-chat-selected');
    var inputArea    = el('input-area');
    var chatHeader   = el('chat-header');
    var msgContainer = el('messages-container');
    var infoPanel    = el('info-panel');

    if (leftPanel) {
      // ... существующие стили левой панели ...
      leftPanel.style.background = 'transparent'; 
      leftPanel.style.backdropFilter = 'blur(10px)';
      leftPanel.style.backgroundColor = 'rgba(var(--bg-sidebar-rgb), 0.3)';
      if (chatsSidebar) chatsSidebar.style.background = 'transparent';
      if (viewHome)     viewHome.style.background = 'transparent';
      leftPanel.style.borderRadius = '24px 24px 0 0';
      leftPanel.style.marginTop = '12px';
      leftPanel.style.height = 'calc(100% - 12px)';
      leftPanel.style.transition = 'transform 0.4s cubic-bezier(0.2, 0.8, 0.2, 1), opacity 0.4s ease';
      if (!state.chatsVisible && !state.feedVisible) {
        leftPanel.style.transform = 'translateY(30px)';
        leftPanel.style.opacity   = '0';
      }
    }

    if (viewChat) {
      // Стилизуем окно чата под стекло
      viewChat.style.background = 'transparent';
      viewChat.style.backdropFilter = 'blur(15px)';
      viewChat.style.backgroundColor = 'rgba(var(--bg-sidebar-rgb), 0.2)';
      viewChat.style.borderRadius = '24px 24px 0 0';
      viewChat.style.marginTop = '12px';
      viewChat.style.marginLeft = '12px';
      viewChat.style.marginRight = '12px';
      viewChat.style.height = 'calc(100% - 12px)';
      
      // Чат занимает фиксированные 59%
      viewChat.style.flex = '0 0 59%'; 
      viewChat.style.width = '59%';
      viewChat.style.maxWidth = '59%';
      
      viewChat.style.transition = 'transform 0.5s cubic-bezier(0.2, 0.8, 0.2, 1), opacity 0.5s ease';
      
      if (!state.chatOpen) {
        viewChat.style.transform = 'translateX(-40px)'; 
        viewChat.style.opacity   = '0';
      }
      
      if (chatHeader) {
        chatHeader.style.background = 'rgba(var(--bg-main-rgb), 0.4)';
        chatHeader.style.borderRadius = '24px 24px 0 0';
      }
      if (msgContainer) msgContainer.style.background = 'transparent';
      if (inputArea) {
        inputArea.style.background = 'rgba(var(--bg-main-rgb), 0.4)';
        inputArea.style.backdropFilter = 'blur(8px)';
      }
    }

    if (infoPanel) {
      infoPanel.style.background = 'transparent';
      infoPanel.style.backdropFilter = 'blur(15px)';
      infoPanel.style.backgroundColor = 'rgba(var(--bg-sidebar-rgb), 0.2)';
      infoPanel.style.borderRadius = '24px 24px 0 0';
      infoPanel.style.marginTop = '12px';
      infoPanel.style.marginRight = '12px';
      infoPanel.style.height = 'calc(100% - 12px)';
      
      // Инфо-панель занимает ровно то, что осталось от чата (примерно 32% с учетом отступов)
      // Она всегда прижата к правому краю благодаря margin-left: auto
      infoPanel.style.flex = '0 0 32%'; 
      infoPanel.style.width = '32%';
      infoPanel.style.marginLeft = 'auto'; 
      
      infoPanel.style.transition = 'transform 0.5s cubic-bezier(0.2, 0.8, 0.2, 1), opacity 0.5s ease';
      
      var infoHeader = infoPanel.querySelector('div:first-child');
      if (infoHeader) {
        infoHeader.style.background = 'rgba(var(--bg-main-rgb), 0.4)';
        infoHeader.style.borderRadius = '24px 24px 0 0';
      }
    }

    // 1. Управление окном чата
    if (state.chatOpen) {
      document.body.classList.add('chat-open');
      if (viewChat) {
        show(viewChat, 'flex');
        requestAnimationFrame(() => {
          viewChat.style.transform = 'translateX(0)';
          viewChat.style.opacity   = '1';
        });
      }
      if (inputArea) show(inputArea, 'flex');
      if (noChat)    hide(noChat);
    } else {
      document.body.classList.remove('chat-open');
      if (viewChat) {
        viewChat.style.transform = 'translateX(-40px)'; 
        viewChat.style.opacity   = '0';
        // Убираем из потока после анимации
        if (viewChat._hideTimer) clearTimeout(viewChat._hideTimer);
        viewChat._hideTimer = setTimeout(() => { 
          if (!state.chatOpen) viewChat.style.display = 'none'; 
        }, 500);
      }
      if (inputArea) hide(inputArea);
      if (noChat)    show(noChat);
    }

    // 2. Управление инфо-панелью (независимо от чата)
    if (infoPanel) {
      if (state.infoOpen) {
        show(infoPanel, 'flex');
        requestAnimationFrame(() => {
          infoPanel.style.transform = 'translateX(0)';
          infoPanel.style.opacity   = '1';
        });
      } else {
        infoPanel.style.transform = 'translateX(40px)';
        infoPanel.style.opacity   = '0';
        setTimeout(() => { if (!state.infoOpen) hide(infoPanel); }, 500);
      }
    }

    // 3. Управление левой панелью (список чатов / лента)
    if (state.chatsVisible || state.feedVisible) {
      show(leftPanel, 'flex');
      requestAnimationFrame(() => {
        leftPanel.style.transform = 'translateY(0)';
        leftPanel.style.opacity   = '1';
      });
      
      if (state.chatsVisible) {
        if (chatsSidebar) show(chatsSidebar, 'flex');
        if (viewHome)     hide(viewHome);
      } else {
        if (chatsSidebar) hide(chatsSidebar);
        if (viewHome)     show(viewHome, 'flex');
      }
    } else {
      leftPanel.style.transform = 'translateY(30px)';
      leftPanel.style.opacity   = '0';
      setTimeout(() => {
        if (!state.chatsVisible && !state.feedVisible) hide(leftPanel);
      }, 400);
    }
    // Удаляем или скрываем правую панель ленты (она мешает анимации)
    hideFeedRight();
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
    panel.style.cssText = 'display:none;flex:1;height:100%;overflow-y:auto;background:transparent;position:relative;z-index:5;';
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
  window.toggleInfoPanel = function () {
    state.infoOpen = !state.infoOpen;
    applyLayout();
  };

  window.switchLeftTab = function (tab) {
    if (tab === 'chats') {
      state.chatsVisible = !state.chatsVisible;
      if (state.chatsVisible) state.feedVisible = false; 
    } else if (tab === 'home') {
      if (state.chatOpen) {
        closeChat(true);
        state.feedVisible = true;
      } else {
        state.feedVisible = !state.feedVisible;
        if (state.feedVisible) state.chatsVisible = false;
      }
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

    // Стилизуем под стекло
    dock.style.background = 'transparent';
    dock.style.backdropFilter = 'blur(10px)';
    dock.style.backgroundColor = 'rgba(var(--bg-sidebar-rgb), 0.3)';
    dock.style.border = '1px solid rgba(255,255,255,0.05)';
    dock.style.borderRadius = '24px 24px 0 0'; // Закругляем верхние углы

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
      state.feedVisible  = false; // По умолчанию скрываем ленту, чтобы видеть анимацию
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