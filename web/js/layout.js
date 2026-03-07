/* ══════════════════════════════════════════════════════════════
   ══ layout.js — навигация и состояние раскладки
   ══════════════════════════════════════════════════════════════ */
(function () {

  var state = {
    chatsVisible:   false,
    feedVisible:    false,
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
    elem.classList.remove('hidden');
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
      leftPanel.style.marginTop = '12px';
      leftPanel.style.height = 'calc(100% - 12px)';

      // Ширина левой панели зависит от режима: чаты уже, лента шире
      var panelWidth = state.chatsVisible ? '533px' : '800px';
      leftPanel.style.flex = '0 0 ' + panelWidth;
      leftPanel.style.width = panelWidth;
      leftPanel.style.minWidth = panelWidth;

      leftPanel.style.transition = 'transform 0.4s cubic-bezier(0.2, 0.8, 0.2, 1), opacity 0.4s ease, width 0.4s cubic-bezier(0.2, 0.8, 0.2, 1), flex-basis 0.4s cubic-bezier(0.2, 0.8, 0.2, 1), min-width 0.4s cubic-bezier(0.2, 0.8, 0.2, 1)';
      if (!state.chatsVisible && !state.feedVisible) {
        leftPanel.style.transform = 'translateY(30px)';
        leftPanel.style.opacity   = '0';
      }
    }

    if (viewChat) {
      viewChat.style.marginTop = '12px';
      viewChat.style.marginLeft = '12px';
      viewChat.style.marginRight = '12px';
      viewChat.style.height = 'calc(100% - 12px)';

      // Чат занимает фиксированные 59%
      viewChat.style.flex = '0 0 59%';
      viewChat.style.width = '59%';
      viewChat.style.maxWidth = '59%';

      viewChat.style.transition = 'transform 0.4s cubic-bezier(0.2, 0.8, 0.2, 1), opacity 0.4s ease';

      if (!state.chatOpen) {
        viewChat.style.transform = 'translateX(-40px)';
        viewChat.style.opacity   = '0';
      }
    }

    if (infoPanel) {
      infoPanel.style.marginTop = '12px';
      infoPanel.style.marginRight = '12px';
      infoPanel.style.height = 'calc(100% - 12px)';

      // Инфо-панель занимает ровно то, что осталось от чата (примерно 32% с учетом отступов)
      infoPanel.style.flex = '0 0 32%';
      infoPanel.style.width = '32%';
      infoPanel.style.marginLeft = 'auto';

      infoPanel.style.transition = 'transform 0.5s cubic-bezier(0.2, 0.8, 0.2, 1), opacity 0.5s ease';
    }

    // 1. Управление окном чата
    if (state.chatOpen) {
      document.body.classList.add('chat-open');
      if (viewChat) {
        viewChat.style.transform = 'translateX(-40px)';
        viewChat.style.opacity   = '0';
        show(viewChat, 'flex');
        requestAnimationFrame(() => requestAnimationFrame(() => {
          viewChat.style.transform = 'translateX(0)';
          viewChat.style.opacity   = '1';
        }));
      }
      if (inputArea) show(inputArea, 'flex');
      if (noChat)    hide(noChat);
    } else {
      // Нет активного чата: панель чата скрыта, ничего не открыто заранее
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
        if (viewHome) {
          viewHome.style.transition = 'opacity 0.25s ease, transform 0.25s ease';
          viewHome.style.opacity = '0';
          viewHome.style.transform = 'translateX(-16px)';
          setTimeout(() => { hide(viewHome); viewHome.style.transform = ''; }, 250);
        }
        if (chatsSidebar) {
          chatsSidebar.style.opacity = '0';
          chatsSidebar.style.transform = 'translateX(16px)';
          show(chatsSidebar, 'flex');
          requestAnimationFrame(() => {
            chatsSidebar.style.transition = 'opacity 0.3s ease, transform 0.3s ease';
            chatsSidebar.style.opacity = '1';
            chatsSidebar.style.transform = 'translateX(0)';
          });
        }
      } else {
        if (chatsSidebar) {
          chatsSidebar.style.transition = 'opacity 0.25s ease, transform 0.25s ease';
          chatsSidebar.style.opacity = '0';
          chatsSidebar.style.transform = 'translateX(16px)';
          setTimeout(() => { hide(chatsSidebar); chatsSidebar.style.transform = ''; }, 250);
        }
        if (viewHome) {
          viewHome.style.opacity = '0';
          viewHome.style.transform = 'translateX(-16px)';
          show(viewHome, 'flex');
          requestAnimationFrame(() => {
            viewHome.style.transition = 'opacity 0.3s ease, transform 0.3s ease';
            viewHome.style.opacity = '1';
            viewHome.style.transform = 'translateX(0)';
          });
        }
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
    var wallOverlay = document.getElementById('wall-overlay');
    var wallIsOpen  = wallOverlay && !wallOverlay.classList.contains('hidden');

    function doSwitch() {
      if (tab === 'chats') {
        state.chatsVisible = !state.chatsVisible;
        if (state.chatsVisible) {
          state.feedVisible = false;
          if (typeof renderChats === 'function' && window.app && window.app.chats) {
            renderChats();
          }
        }
      } else if (tab === 'home') {
        if (state.chatOpen) {
          closeChat(true);
          state.feedVisible = true;
        } else {
          state.feedVisible = !state.feedVisible;
          if (state.feedVisible) {
            state.chatsVisible = false;
            if (typeof loadActivityFeed === 'function') {
              loadActivityFeed();
            }
          }
        }
      }
      applyLayout();
    }

    if (wallIsOpen) {
      if (typeof closeWall === 'function') closeWall();
      setTimeout(doSwitch, 220);
    } else {
      doSwitch();
    }
  };

  // switchView — совместимость с api.js
  window.switchView = function (view) {
    if (view === 'chat') {
      // Закрываем стену плавно, не мешая открытию чата
      var wallOverlay = document.getElementById('wall-overlay');
      if (wallOverlay && !wallOverlay.classList.contains('hidden')) {
        if (typeof closeWall === 'function') closeWall();
      }
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
    if (state.feedVisible && typeof loadActivityFeed === 'function') {
      loadActivityFeed();
    }
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
      // После успешной авторизации на главном экране сразу показываем ленту
      state.chatsVisible = false;
      state.feedVisible  = true;
      state.chatOpen     = false;
      if (typeof loadActivityFeed === 'function') {
        loadActivityFeed();
      }
      applyLayout();
    };
  }

  // ── Init ────────────────────────────────────────────────────
  function init() {
    initDock();
    updateDockActive();

    // Применить layout сразу если уже в чат-режиме
    if (isMainChatVisible()) {
      // По умолчанию сразу показываем ленту
      state.chatsVisible = false;
      state.feedVisible  = true;
      state.chatOpen     = false;
      if (typeof loadActivityFeed === 'function') {
        loadActivityFeed();
      }
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