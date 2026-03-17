/* ══════════════════════════════════════════════════════════════
   ══ layout.js — навигация и состояние раскладки
   ══════════════════════════════════════════════════════════════ */
(function () {

  var state = {
    chatsVisible:   false,
    feedVisible:    false,
    chatOpen:       false,
    infoOpen:       false,
    feedWasVisible: false
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
    var isMobile = window.innerWidth <= 768;
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

    if (isMobile) {
      // На мобилках сбрасываем инлайновые стили и полагаемся на CSS
      [leftPanel, viewChat, infoPanel].forEach(function(e) {
        if (e) {
          e.style.width = '';
          e.style.minWidth = '';
          e.style.maxWidth = '';
          e.style.flex = '';
          e.style.margin = '';
          e.style.height = '';
          e.style.transform = '';
          e.style.opacity = '';
        }
      });
    }

    // ── Десктоп: ширина панелей ──────────────────────────────
    if (!isMobile) {
      var TRANSITION = 'transform 0.4s cubic-bezier(0.2, 0.8, 0.2, 1), opacity 0.4s ease, width 0.4s cubic-bezier(0.2, 0.8, 0.2, 1), flex-basis 0.4s cubic-bezier(0.2, 0.8, 0.2, 1), min-width 0.4s cubic-bezier(0.2, 0.8, 0.2, 1)';

      if (leftPanel) {
        leftPanel.style.height = '100%';
        leftPanel.style.transition = TRANSITION;

        if (state.chatOpen) {
          // Чат открыт — список сжимается до боковой колонки
          leftPanel.style.flex = '0 0 280px';
          leftPanel.style.width = '280px';
          leftPanel.style.minWidth = '280px';
          leftPanel.style.maxWidth = '280px';
          leftPanel.style.transform = 'translateX(0)';
          leftPanel.style.opacity = '1';
        } else {
          // Чата нет — левая панель занимает весь #main-chat
          leftPanel.style.flex = '1 1 0';
          leftPanel.style.width = '100%';
          leftPanel.style.minWidth = '0';
          leftPanel.style.maxWidth = '100%';

          if (!state.chatsVisible && !state.feedVisible) {
            leftPanel.style.transform = 'translateY(30px)';
            leftPanel.style.opacity   = '0';
          }
        }
      }

      if (viewChat) {
        viewChat.style.height = '100%';
        viewChat.style.marginLeft = state.chatOpen ? '12px' : '0';
        viewChat.style.transition = 'transform 0.4s cubic-bezier(0.2, 0.8, 0.2, 1), opacity 0.4s ease';
        viewChat.style.flex = '1 1 0';
        viewChat.style.minWidth = '0';
        viewChat.style.maxWidth = '';
      }

      if (infoPanel) {
        infoPanel.style.height = '100%';
        infoPanel.style.flex = '0 0 280px';
        infoPanel.style.width = '280px';
        infoPanel.style.marginLeft = '12px';
        infoPanel.style.transition = 'transform 0.5s cubic-bezier(0.2, 0.8, 0.2, 1), opacity 0.5s ease';
      }
    }

    // ── body классы ─────────────────────────────────────────
    var anyPanelOpen = !!(state.chatOpen || state.chatsVisible || state.feedVisible);
    document.body.classList.toggle('chat-open', !!state.chatOpen);
    document.body.classList.toggle('feed-open', !!(state.feedVisible && !state.chatOpen));
    document.body.classList.toggle('panel-open', anyPanelOpen);

    // 1. Управление окном чата
    if (state.chatOpen) {
      if (isMobile) {
        // Мобильный: слайд управляется через CSS (body.chat-open) — сбрасываем JS-стили
        if (leftPanel) { leftPanel.style.display = ''; leftPanel.style.transform = ''; leftPanel.style.opacity = ''; }
        if (viewChat)  { viewChat.style.display  = 'flex'; viewChat.style.transform = ''; viewChat.style.opacity = ''; }
        if (infoPanel && !state.infoOpen) hide(infoPanel);
      } else {
        // Десктоп: левая панель сжимается (не скрывается) — список чатов уезжает влево
        if (leftPanel) {
          if (!state.chatsVisible && !state.feedVisible) {
            // Если ни лента ни чаты не открыты — скрыть левую панель
            leftPanel.style.transform = 'translateX(-40px)';
            leftPanel.style.opacity = '0';
            setTimeout(function() { if (state.chatOpen && !state.chatsVisible && !state.feedVisible) hide(leftPanel); }, 400);
          }
        }
      }
      if (viewChat) {
        if (!isMobile) {
          viewChat.style.transform = 'translateY(20px)';
          viewChat.style.opacity   = '0';
        }
        show(viewChat, 'flex');
        if (!isMobile) {
          requestAnimationFrame(function() { requestAnimationFrame(function() {
            viewChat.style.transform = 'translateY(0)';
            viewChat.style.opacity   = '1';
          }); });
        }
      }
      if (inputArea) show(inputArea, 'flex');
      if (noChat)    hide(noChat);

      // Dok всегда виден — показываем его
      showDock();
      scheduleHide();
    } else {
      // Нет активного чата: панель чата скрыта, ничего не открыто заранее
      document.body.classList.remove('chat-open');
      if (isMobile) {
        // Мобильный: сброс JS-стилей, CSS вернёт панели на место
        if (leftPanel) { leftPanel.style.display = ''; leftPanel.style.transform = ''; leftPanel.style.opacity = ''; }
        if (viewChat)  { viewChat.style.display  = ''; viewChat.style.transform  = ''; viewChat.style.opacity  = ''; }
      }
      if (viewChat) {
        if (!isMobile) {
          viewChat.style.transform = 'translateY(40px)';
          viewChat.style.opacity   = '0';
          // Убираем из потока после анимации
          if (viewChat._hideTimer) clearTimeout(viewChat._hideTimer);
          viewChat._hideTimer = setTimeout(function() {
            if (!state.chatOpen) viewChat.style.display = 'none';
          }, 500);
        } else {
          // На мобильном — CSS slide анимация через body.chat-open, не прячем через JS
          viewChat.style.display = '';
          viewChat.style.transform = '';
          viewChat.style.opacity = '';
        }
      }
      if (inputArea) hide(inputArea);
      if (noChat && !isMobile) show(noChat);
      else if (noChat) hide(noChat);
    }

    // 2. Управление инфо-панелью (независимо от чата)
    if (infoPanel) {
      if (state.infoOpen) {
        show(infoPanel, 'flex');
        requestAnimationFrame(function() {
          infoPanel.style.transform = 'translateX(0)';
          infoPanel.style.opacity   = '1';
        });
      } else {
        infoPanel.style.transform = 'translateX(40px)';
        infoPanel.style.opacity   = '0';
        setTimeout(function() { if (!state.infoOpen) hide(infoPanel); }, 500);
      }
    }

    // 3. Управление левой панелью (список чатов / лента)
    if (state.chatsVisible || state.feedVisible) {
      if (!(isMobile && state.chatOpen)) {
        show(leftPanel, 'flex');
        requestAnimationFrame(function() {
          leftPanel.style.transform = 'translateY(0)';
          leftPanel.style.opacity   = '1';
        });
      }

      if (state.chatsVisible) {
        if (viewHome) {
          viewHome.style.transition = 'opacity 0.25s ease, transform 0.25s ease';
          viewHome.style.opacity = '0';
          viewHome.style.transform = 'translateX(-16px)';
          setTimeout(function() { hide(viewHome); viewHome.style.transform = ''; }, 250);
        }
        if (chatsSidebar) {
          chatsSidebar.style.opacity = '0';
          chatsSidebar.style.transform = 'translateX(16px)';
          show(chatsSidebar, 'flex');
          requestAnimationFrame(function() {
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
          setTimeout(function() { hide(chatsSidebar); chatsSidebar.style.transform = ''; }, 250);
        }
        if (viewHome) {
          viewHome.style.opacity = '0';
          viewHome.style.transform = 'translateX(-16px)';
          show(viewHome, 'flex');
          requestAnimationFrame(function() {
            viewHome.style.transition = 'opacity 0.3s ease, transform 0.3s ease';
            viewHome.style.opacity = '1';
            viewHome.style.transform = 'translateX(0)';
          });
        }
      }
    } else {
      // Фиксируем текущую ширину чтобы не было расширения при скрытии
      var currentWidth = leftPanel.offsetWidth;
      leftPanel.style.flex = '0 0 ' + currentWidth + 'px';
      leftPanel.style.width = currentWidth + 'px';
      leftPanel.style.minWidth = currentWidth + 'px';
      leftPanel.style.maxWidth = currentWidth + 'px';
      leftPanel.style.transform = 'translateY(30px)';
      leftPanel.style.opacity   = '0';
      setTimeout(function() {
        if (!state.chatsVisible && !state.feedVisible) {
          hide(leftPanel);
          // Сбрасываем инлайн-стили после скрытия
          leftPanel.style.flex = '';
          leftPanel.style.width = '';
          leftPanel.style.minWidth = '';
          leftPanel.style.maxWidth = '';
        }
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
    var isMobile    = window.innerWidth <= 768;

    function doSwitch() {
      if (tab === 'chats') {
        if (isMobile && state.chatOpen) {
          closeChat(false, true); // true = skip applyLayout
          state.chatsVisible = true;
          state.feedVisible  = false;
        } else {
          state.chatsVisible = !state.chatsVisible;
          if (state.chatsVisible) state.feedVisible = false;
        }
        if (state.chatsVisible && typeof renderChats === 'function' && window.app && window.app.chats) {
          renderChats();
        }
      } else if (tab === 'home') {
        if (state.chatOpen) {
          closeChat(true, true); // true = skip applyLayout
        } else {
          state.feedVisible = !state.feedVisible;
          if (state.feedVisible) state.chatsVisible = false;
        }
        if (state.feedVisible && typeof loadActivityFeed === 'function') {
          loadActivityFeed();
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

  function closeChat(showFeed, skipApply) {
    state.chatOpen = false;
    document.querySelectorAll('.chat-list-item').forEach(function (i) {
      i.classList.remove('active');
    });
    if (window.app) window.app.activeChatId = null;
    state.feedVisible = showFeed ? true : state.feedWasVisible;
    if (state.feedVisible && typeof loadActivityFeed === 'function') {
      loadActivityFeed();
    }
    if (!skipApply) applyLayout();
  }

  // ── Dock ────────────────────────────────────────────────────
  function updateDockActive() {
    var dockHome  = el('dock-home');
    var dockChats = el('dock-chats');
    if (dockHome)  dockHome.classList.toggle('dock-active',  state.feedVisible && !state.chatOpen);
    if (dockChats) dockChats.classList.toggle('dock-active', state.chatsVisible && !state.chatOpen);
  }

  var TRIGGER_PX  = 80;
  var HIDE_DELAY  = 5000;
  var HIDE_DELAY_MOBILE = 3000;
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
    var delay = window.innerWidth <= 768 ? HIDE_DELAY_MOBILE : HIDE_DELAY;
    hideTimer = setTimeout(function () {
      var dock = el('bottom-dock');
      if (dock) dock.classList.remove('visible');
    }, delay);
  }

  function initDock() {
    var dock = el('bottom-dock');
    if (!dock) return;

    // Десктоп: триггер по краю экрана мышкой
    document.addEventListener('mousemove', function (e) {
      if (!isMainChatVisible()) return;
      if (window.innerHeight - e.clientY <= TRIGGER_PX) showDock();
    });

    // Мобильный: триггер по касанию нижней зоны
    document.addEventListener('touchstart', function (e) {
      if (!isMainChatVisible()) return;
      var touch = e.touches[0];
      if (touch && window.innerHeight - touch.clientY <= TRIGGER_PX) {
        showDock();
        scheduleHide();
      }
    }, { passive: true });

    // Мобильный: свайп вверх снизу показывает dok
    var touchStartY = 0;
    document.addEventListener('touchstart', function (e) {
      touchStartY = e.touches[0] ? e.touches[0].clientY : 0;
    }, { passive: true });

    document.addEventListener('touchend', function (e) {
      if (!isMainChatVisible()) return;
      var touch = e.changedTouches[0];
      if (!touch) return;
      var deltaY = touchStartY - touch.clientY;
      // Свайп вверх от низа экрана — показываем dok
      if (deltaY < -30 && touch.clientY > window.innerHeight - 150) {
        showDock();
        scheduleHide();
      }
    }, { passive: true });

    dock.addEventListener('mouseenter', function () {
      dockHovered = true;
      showDock();
    });
    dock.addEventListener('mouseleave', function () {
      dockHovered = false;
      scheduleHide();
    });

    // Тап по доку — не скрываем сразу
    dock.addEventListener('touchstart', function () {
      dockHovered = true;
      showDock();
    }, { passive: true });
    dock.addEventListener('touchend', function () {
      dockHovered = false;
      scheduleHide();
    }, { passive: true });
  }

  // ── Патч AlphaApp ───────────────────────────────────────────
  function patchApp() {
    var app = window.app;
    if (!app || app._layoutPatched) return;
    app._layoutPatched = true;

    var origShowChat = app.showChat.bind(app);
    app.showChat = function () {
      origShowChat();
      // Убираем принудительное открытие чатов при логине
      applyLayout();
      // Показываем dock сразу при входе
      showDock();
      scheduleHide();
    };
  }

  // ── Init ────────────────────────────────────────────────────
  function init() {
    initDock();
    updateDockActive();

    // Применить layout сразу если уже в чат-режиме
    if (isMainChatVisible()) {
      applyLayout();
    }

    var _lastMobile = null;
    window.addEventListener('resize', function() {
      var nowMobile = window.innerWidth <= 768;
      // При переходе мобильный↔десктоп сбрасываем все инлайн-стили панелей
      if (_lastMobile !== null && _lastMobile !== nowMobile) {
        var panels = ['left-panel', 'view-chat', 'info-panel'];
        panels.forEach(function(id) {
          var el = document.getElementById(id);
          if (el) el.removeAttribute('style');
        });
      }
      _lastMobile = nowMobile;
      applyLayout();
    });

    // Показываем dock сразу если уже авторизован
    if (window.app && window.app.currentUser) {
      var dock = el('bottom-dock');
      if (dock) { dock.classList.remove('guest-hidden'); dock.style.display = ''; }
      showDock();
      scheduleHide();
    } else {
      // Ждём авторизации
      document.addEventListener('app:authenticated', function() {
        var dock = el('bottom-dock');
        if (dock) { dock.classList.remove('guest-hidden'); dock.style.display = ''; }
        showDock();
        scheduleHide();
      }, { once: true });
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

  // Публичная функция — открыть ленту принудительно (для share-ссылок)
  window.openFeedPanel = function () {
    state.feedVisible  = true;
    state.chatsVisible = false;
    applyLayout();
    if (typeof loadActivityFeed === 'function') loadActivityFeed();
  };

})();