// ─── feed.js — Лента активности + медиа-навигация ───────────────────────

const Feed = (() => {

  let _mediaList        = [];
  let _mediaIndex       = 0;
  let _mediaWheelLocked = false;
  let _renderObserver   = null;
  let _dwellTimers      = new Map();

  // ── Трекинг ──────────────────────────────────────────────────────────
  function _track(postId, eventType, watchSeconds = 0, mime = '') {
    fetch(`/api/feed/track?mime=${encodeURIComponent(mime)}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('alpha_token')}`
      },
      body: JSON.stringify({ post_id: postId, event_type: eventType, watch_seconds: watchSeconds })
    }).catch(() => {});
  }

  // ── Загрузка ─────────────────────────────────────────────────────────
  async function load() {
    const container = document.getElementById('activity-feed-container');
    if (!container) return;

    try {
      const response = await fetch('/api/feed', {
        headers: { 'Authorization': `Bearer ${localStorage.getItem('alpha_token')}` }
      });
      if (!response.ok) throw new Error('Failed');
      const posts = await response.json();

      if (!posts || posts.length === 0) {
        container.innerHTML = `<div class="col-span-3 flex items-center justify-center py-20 opacity-40">
                    <p class="text-sm font-semibold text-custom-main">В ленте пока пусто...</p></div>`;
        return;
      }
      _render(container, posts);
    } catch (err) {
      console.error('Feed error:', err);
      container.innerHTML = `<p class="col-span-3 text-xs text-red-400 p-4">Ошибка загрузки ленты</p>`;
    }
  }

  // ── Рендер грида ─────────────────────────────────────────────────────
  function _render(container, posts) {
    if (_renderObserver) { _renderObserver.disconnect(); _renderObserver = null; }
    _dwellTimers.forEach(t => clearTimeout(t));
    _dwellTimers.clear();
    container.innerHTML = '';

    posts.forEach((p, i) => {
      const card = document.createElement('div');
      card.dataset.postId   = p.id;
      card.dataset.postMime = p.attachments?.[0]?.mime_type || '';

      if (i < 9) {
        card.innerHTML = _renderPost(p);
      } else {
        card.className = 'activity-post-card rounded-[24px] bg-custom-sidebar border border-white/5 aspect-square';
        card.innerHTML = `<div class="w-full h-full rounded-[24px] bg-white/5 animate-pulse"></div>`;
        card.dataset.postData = JSON.stringify(p);
      }
      container.appendChild(card);
    });

    // Lazy render + трекинг просмотра (1с dwell)
    _renderObserver = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (!entry.isIntersecting) return;
        const card = entry.target;
        if (card.dataset.postData) {
          const p = JSON.parse(card.dataset.postData);
          card.innerHTML = _renderPost(p);
          delete card.dataset.postData;
          // Запускаем превью для новых видео-карточек
          card.querySelectorAll('.feed-video-thumb').forEach(el => {
            _thumbQueue.push({ src: el.dataset.videoSrc, cid: el.dataset.canvasId });
          });
          _processThumbQueue();
        }
        const postId = card.dataset.postId;
        const mime   = card.dataset.postMime;
        if (postId && !_dwellTimers.has(postId)) {
          const t = setTimeout(() => {
            _track(postId, 'view', 0, mime);
            _dwellTimers.delete(postId);
          }, 1000);
          _dwellTimers.set(postId, t);
        }
        _renderObserver.unobserve(card);
      });
    }, { rootMargin: '300px' });

    Array.from(container.children).forEach((child, i) => {
      if (i >= 9) _renderObserver.observe(child);
    });

    // Запускаем превью для видео карточек — по одному через очередь
    _initThumbObserver(container);
  }

  // ── HTML карточки поста ───────────────────────────────────────────────
  function _renderPost(p) {
    const hasAtt = p.attachments && p.attachments.length > 0;
    const openAction = hasAtt
        ? `Feed.openMedia('${p.id}', '${p.attachments[0].url}', ${(p.attachments[0].mime_type || '').startsWith('video/')})`
        : `openPostChat('${p.id}', '${p.chat_id || ''}')`;

    return `
        <div class="activity-post-card rounded-[24px] group relative overflow-hidden bg-custom-sidebar border border-white/5 hover:border-custom-accent/30 transition-all shadow-sm aspect-square" id="activity-post-${p.id}"
             data-post-id="${p.id}" data-media-url="${hasAtt ? p.attachments[0].url : ''}" data-media-video="${hasAtt && (p.attachments[0].mime_type||'').startsWith('video/')}">
           ${hasAtt
        ? `<div class="w-full h-full overflow-hidden">
                    ${_renderThumb(p.id, p.attachments[0])}
                    ${p.attachments.length > 1 ? `<div class="absolute top-2 right-2 bg-black/60 text-white text-[9px] px-1.5 py-0.5 rounded-full z-10">+${p.attachments.length - 1}</div>` : ''}
                  </div>`
        : `<div class="w-full h-full p-3 flex items-center justify-center text-center bg-custom-sidebar/50">
                     <div class="text-[10px] text-custom-main leading-relaxed line-clamp-6 italic">${p.content || ''}</div>
                  </div>`}
           <div class="absolute inset-0 bg-gradient-to-t from-black/75 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity">
              <div class="absolute bottom-0 left-0 right-0 p-2.5">
                 <div class="flex items-center gap-1.5 mb-1">
                    <div class="w-4 h-4 rounded-full bg-white/20 overflow-hidden border border-white/20 flex-shrink-0 flex items-center justify-center text-[7px] font-bold text-white">
                       ${p.author_avatar ? `<img src="${p.author_avatar}" class="w-full h-full object-cover">` : (p.author_name?.[0] ?? '?')}
                    </div>
                    <span class="text-[9px] font-semibold text-white truncate">${p.author_name}</span>
                 </div>
                 <div class="flex gap-2 text-white/70 text-[8px]">
                    <span class="flex items-center gap-0.5">
                       <svg class="w-2.5 h-2.5" fill="currentColor" viewBox="0 0 24 24"><path d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"/></svg>
                       ${p.likes_count || 0}
                    </span>
                    <span class="flex items-center gap-0.5">
                       <svg class="w-2.5 h-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/></svg>
                       ${p.comments_count || 0}
                    </span>
                 </div>
              </div>
           </div>
           <button onclick="${openAction}" class="absolute inset-0 z-0" style="background:transparent;border:none;"></button>
        </div>`;
  }

  // Превью: для видео — canvas рисуется только когда карточка видна (через data-src)
  function _renderThumb(postId, a) {
    const isVideo = (a.mime_type || '').startsWith('video/');
    if (isVideo) {
      const id = `vt-${postId.slice(0,8)}`;
      return `<div class="relative w-full h-full bg-black feed-video-thumb" data-video-src="${a.url}" data-canvas-id="${id}" onclick="Feed.openMedia('${postId}', '${a.url}', true)" style="cursor:pointer;">
                <canvas id="${id}" class="w-full h-full" style="object-fit:cover;display:block;"></canvas>
                <div class="absolute inset-0 flex items-center justify-center pointer-events-none">
                    <div class="w-10 h-10 rounded-full bg-black/40 backdrop-blur-sm flex items-center justify-center">
                        <svg class="w-5 h-5 text-white ml-0.5" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
                    </div>
                </div>
            </div>`;
    }
    return `<div class="w-full h-full" onclick="Feed.openMedia('${postId}', '${a.url}', false)" style="cursor:pointer;">
            <img src="${a.url}" class="w-full h-full object-cover" loading="lazy" decoding="async">
        </div>`;
  }

  // Превью видео — параллельно до 3 штук одновременно
  const _thumbQueue   = [];
  const THUMB_PARALLEL = 3;
  let   _thumbActive  = 0;

  function _initThumbObserver(container) {
    const obs = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (!entry.isIntersecting) return;
        const el  = entry.target;
        const src = el.dataset.videoSrc;
        const cid = el.dataset.canvasId;
        if (src && cid) { _thumbQueue.push({ src, cid }); _processThumbQueue(); }
        obs.unobserve(el);
      });
    }, { rootMargin: '150px' });
    container.querySelectorAll('.feed-video-thumb').forEach(el => obs.observe(el));
  }

  function _processThumbQueue() {
    while (_thumbActive < THUMB_PARALLEL && _thumbQueue.length > 0) {
      _thumbActive++;
      const { src, cid } = _thumbQueue.shift();
      _loadThumb(src, cid).finally(() => {
        _thumbActive--;
        _processThumbQueue();
      });
    }
  }

  function _loadThumb(src, cid) {
    return new Promise(resolve => {
      const v = document.createElement('video');
      v.muted = true; v.preload = 'metadata'; v.style.display = 'none';
      document.body.appendChild(v);
      v.onloadedmetadata = () => { v.currentTime = 1; };
      v.onseeked = () => {
        const c = document.getElementById(cid);
        if (c) {
          c.width  = v.videoWidth  || 320;
          c.height = v.videoHeight || 180;
          c.getContext('2d').drawImage(v, 0, 0, c.width, c.height);
        }
        v.pause(); v.src = ''; v.load(); v.remove();
        resolve();
      };
      v.onerror = () => { v.remove(); resolve(); };
      v.src = src;
    });
  }

  // ── Открытие медиа ────────────────────────────────────────────────────
  async function openMedia(postId, url, isVideo) {
    // Собираем список медиа только из ленты (не все посты подряд)
    _mediaList = [];
    const wallGrid    = document.getElementById('wall-media-grid-main');
    const wallOverlay = document.getElementById('wall-overlay');
    const wallIsOpen  = wallGrid && wallOverlay && !wallOverlay.classList.contains('hidden');

    if (wallIsOpen) {
      wallGrid.querySelectorAll('.wall-media-item').forEach(item => {
        const m = (item.getAttribute('onclick') || '').match(/Feed\.openMedia\('([^']+)',\s*'([^']+)',\s*(true|false)\)/)
            || (item.getAttribute('onclick') || '').match(/openMediaDetail\('([^']+)',\s*'([^']+)',\s*(true|false)\)/);
        if (m) _mediaList.push({ postId: m[1], url: m[2], isVideo: m[3] === 'true' });
      });
    } else {
      // Из ленты — читаем data-атрибуты карточек, без парсинга onclick
      document.querySelectorAll('#activity-feed-container .activity-post-card[data-media-url]').forEach(card => {
        const url = card.dataset.mediaUrl;
        if (!url) return;
        _mediaList.push({
          postId:  card.dataset.postId,
          url:     url,
          isVideo: card.dataset.mediaVideo === 'true'
        });
      });
    }

    _mediaIndex = _mediaList.findIndex(m => m.postId === String(postId) && m.url === url);
    if (_mediaIndex === -1) _mediaIndex = 0;

    _track(postId, isVideo ? 'video_complete' : 'view', 0, isVideo ? 'video/mp4' : 'image/jpeg');
    await _showMedia(postId, url, isVideo, null);
  }

  async function _showMedia(postId, url, isVideo, direction) {
    const mediaContainer = document.getElementById('wall-comments-media-container');
    const mediaContent   = document.getElementById('wall-comments-media-content');
    const likeBtn        = document.getElementById('wall-comments-media-like');
    const indicator      = document.getElementById('media-nav-indicator');

    // Сбрасываем гостевые стили (если были)
    const panel = document.querySelector('.wall-comments-panel');
    if (panel) {
      const rightCol = panel.querySelector('.flex-col.flex-grow');
      if (rightCol && rightCol.style.display === 'none' && localStorage.getItem('alpha_token')) {
        rightCol.style.display = '';
      }
    }
    if (mediaContainer) {
      if (localStorage.getItem('alpha_token')) {
        mediaContainer.style.width  = '';
        mediaContainer.style.border = '';
      }
    }

    // Убираем старый индикатор при каждом открытии
    if (indicator) indicator.remove();

    mediaContainer.classList.remove('hidden');

    if (direction !== null && mediaContent) {
      mediaContent.style.transition = 'opacity 0.15s ease, transform 0.15s ease';
      mediaContent.style.opacity = '0';
      mediaContent.style.transform = `translateY(${direction > 0 ? '-20px' : '20px'})`;
      await new Promise(r => setTimeout(r, 150));
    }

    if (mediaContent) {
      // Видео на весь контейнер
      mediaContent.innerHTML = isVideo
          ? `<div style="position:relative;width:100%;height:100%;">
                    <video src="${url}" class="w-full h-full object-contain" controls playsinline id="modal-video-player"></video>
                    <div id="video-player-controls" style="position:absolute;bottom:48px;left:12px;display:flex;gap:8px;z-index:10;">
                        <button id="btn-loop" onclick="Feed.toggleLoop()" title="Повтор" style="background:rgba(0,0,0,0.55);border:none;border-radius:8px;padding:6px 12px;color:#fff;cursor:pointer;font-size:12px;display:flex;align-items:center;gap:5px;transition:background 0.2s;">
                            <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>
                            Повтор
                        </button>
                        <button id="btn-autonext" onclick="Feed.toggleAutoNext()" title="Следующее" style="background:rgba(0,0,0,0.55);border:none;border-radius:8px;padding:6px 12px;color:#fff;cursor:pointer;font-size:12px;display:flex;align-items:center;gap:5px;transition:background 0.2s;">
                            <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M13 5l7 7-7 7M5 5l7 7-7 7"/></svg>
                            Авто
                        </button>
                        <button id="btn-share" onclick="Feed.sharePost('${postId}')" title="Поделиться" style="background:rgba(0,0,0,0.55);border:none;border-radius:8px;padding:6px 12px;color:#fff;cursor:pointer;font-size:12px;display:flex;align-items:center;gap:5px;transition:background 0.2s;">
                            <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"/></svg>
                            Поделиться
                        </button>
                    </div>
                  </div>`
          : `<div style="position:relative;width:100%;height:100%;">
                    <img src="${url}" class="w-full h-full object-contain">
                    <div style="position:absolute;bottom:12px;left:12px;z-index:10;">
                        <button onclick="Feed.sharePost('${postId}')" title="Поделиться" style="background:rgba(0,0,0,0.55);border:none;border-radius:8px;padding:6px 12px;color:#fff;cursor:pointer;font-size:12px;display:flex;align-items:center;gap:5px;transition:background 0.2s;">
                            <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"/></svg>
                            Поделиться
                        </button>
                    </div>
                  </div>`;

      mediaContent.style.opacity = '0';
      mediaContent.style.transform = direction === null ? 'none' : `translateY(${direction > 0 ? '20px' : '-20px'})`;
      requestAnimationFrame(() => {
        mediaContent.style.transition = 'opacity 0.22s ease, transform 0.22s ease';
        mediaContent.style.opacity = '1';
        mediaContent.style.transform = 'translateY(0)';
      });
      // Запускаем видео через JS после рендера — не через autoplay атрибут
      if (isVideo) {
        const vEl = document.getElementById('modal-video-player');
        if (vEl) {
          vEl.loop = _loopMode;
          if (_autoNextMode) {
            vEl.onended = () => { if (_mediaList.length > 1) openMedia(_mediaList[(_mediaIndex + 1) % _mediaList.length].postId, _mediaList[(_mediaIndex + 1) % _mediaList.length].url, true, 1); };
          }
          vEl.load(); vEl.play().catch(() => {});
        }
        _updateVideoButtons();
      }
    }

    // Лайк
    if (likeBtn) {
      const postEl = document.getElementById(`activity-post-${postId}`) || document.getElementById(`post-${postId}`);
      let isLiked = false;
      if (postEl) {
        const lb = postEl.querySelector('button[onclick^="togglePostLike"]');
        if (lb) isLiked = lb.dataset.liked === 'true';
      }
      likeBtn.dataset.postId = postId;
      likeBtn.dataset.liked = isLiked;
      const svg = likeBtn.querySelector('svg');
      if (isLiked) { likeBtn.classList.add('text-red-400'); svg?.setAttribute('fill', 'currentColor'); }
      else { likeBtn.classList.remove('text-red-400'); svg?.setAttribute('fill', 'none'); }
    }

    // Комментарии (только для авторизованных)
    const _token = localStorage.getItem('alpha_token');
    if (_token) {
      try {
        const res = await fetch(`/api/wall/posts/${postId}/chat`, {
          headers: { 'Authorization': `Bearer ${_token}` }
        }).then(r => r.json());
        if (res.chat_id) openWallComments(postId, res.chat_id, true);
      } catch (e) {}
    } else {
      // Гость — открываем overlay для медиа (стандартный вид, как у авторизованных)
      const overlay = document.getElementById('wall-comments-overlay');
      if (overlay) {
        overlay.classList.remove('hidden');
        requestAnimationFrame(() => overlay.classList.add('comments-open'));
        // Скрываем правую панель комментариев
        const panel = overlay.querySelector('.wall-comments-panel');
        if (panel) {
          const rightCol = panel.querySelector('.flex-col.flex-grow');
          if (rightCol) rightCol.style.display = 'none';
          const mediaCol = document.getElementById('wall-comments-media-container');
          if (mediaCol) { mediaCol.style.width = '100%'; mediaCol.style.border = 'none'; }
        }
        // Кнопка «← На главную»
        if (!document.getElementById('guest-close-btn')) {
          const btn = document.createElement('button');
          btn.id = 'guest-close-btn';
          btn.innerHTML = '<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M10 19l-7-7m0 0l7-7m-7 7h18"/></svg> На главную';
          btn.onclick = function() { closeWallComments(); };
          btn.style.cssText = 'position:absolute;top:14px;left:14px;z-index:300;background:rgba(0,0,0,0.55);border:none;border-radius:10px;padding:7px 14px;color:#fff;cursor:pointer;font-size:13px;font-weight:600;display:flex;align-items:center;gap:6px;backdrop-filter:blur(8px);transition:background 0.2s;';
          btn.onmouseenter = () => btn.style.background = 'rgba(99,102,241,0.75)';
          btn.onmouseleave = () => btn.style.background = 'rgba(0,0,0,0.55)';
          overlay.appendChild(btn);
        }
      }
    }

    // Wheel навигация (вешаем один раз)
    const overlay = document.getElementById('wall-comments-overlay');
    if (overlay && !overlay._mediaWheelAdded) {
      overlay._mediaWheelAdded = true;
      overlay.addEventListener('wheel', (e) => {
        if (_mediaList.length < 2) return;
        e.preventDefault();
        if (_mediaWheelLocked) return;
        _mediaWheelLocked = true;
        const dir = e.deltaY > 0 ? 1 : -1;
        _mediaIndex = Math.max(0, Math.min(_mediaList.length - 1, _mediaIndex + dir));
        const next = _mediaList[_mediaIndex];
        if (next) _showMedia(next.postId, next.url, next.isVideo, dir);
        setTimeout(() => { _mediaWheelLocked = false; }, 500);
      }, { passive: false });
    }

    // Индикатор — только если медиа больше 1
    if (_mediaList.length > 1) _renderIndicator(mediaContainer);
  }

  function _renderIndicator(mediaContainer) {
    const el = document.createElement('div');
    el.id = 'media-nav-indicator';
    el.style.cssText = 'position:absolute;right:12px;top:50%;transform:translateY(-50%);display:flex;flex-direction:column;gap:4px;z-index:20;pointer-events:none;';
    // Показываем максимум 10 точек вокруг текущего
    const total = _mediaList.length;
    const start = Math.max(0, Math.min(_mediaIndex - 4, total - 10));
    const end   = Math.min(total, start + 10);
    el.innerHTML = Array.from({ length: end - start }, (_, i) => {
      const idx = start + i;
      const active = idx === _mediaIndex;
      return `<div style="width:3px;height:${active ? '18px' : '5px'};border-radius:2px;
                background:${active ? 'rgba(255,255,255,0.9)' : 'rgba(255,255,255,0.3)'};
                transition:all 0.2s ease;"></div>`;
    }).join('');
    mediaContainer.appendChild(el);
  }

  let _loopMode = false;
  let _autoNextMode = false;

  function toggleLoop() {
    _loopMode = !_loopMode;
    _autoNextMode = false;
    const vEl = document.getElementById('modal-video-player');
    if (vEl) vEl.loop = _loopMode;
    _updateVideoButtons();
  }

  function toggleAutoNext() {
    _autoNextMode = !_autoNextMode;
    _loopMode = false;
    const vEl = document.getElementById('modal-video-player');
    if (vEl) {
      vEl.loop = false;
      if (_autoNextMode) {
        vEl.onended = () => { if (_mediaList.length > 1) openMedia(_mediaList[(_mediaIndex + 1) % _mediaList.length].postId, _mediaList[(_mediaIndex + 1) % _mediaList.length].url, true, 1); };
      } else {
        vEl.onended = null;
      }
    }
    _updateVideoButtons();
  }

  function _updateVideoButtons() {
    const btnLoop = document.getElementById('btn-loop');
    const btnNext = document.getElementById('btn-autonext');
    if (btnLoop) btnLoop.style.background = _loopMode ? 'rgba(99,102,241,0.85)' : 'rgba(0,0,0,0.55)';
    if (btnNext) btnNext.style.background = _autoNextMode ? 'rgba(99,102,241,0.85)' : 'rgba(0,0,0,0.55)';
  }

  // ── Поделиться постом ─────────────────────────────────────────────────
  function sharePost(postId) {
    const url = `${location.origin}${location.pathname}?post=${postId}`;
    navigator.clipboard.writeText(url).then(() => {
      window.app?.notify?.('Ссылка скопирована ✓', 'success');
    }).catch(() => {
      const ta = document.createElement('textarea');
      ta.value = url;
      document.body.appendChild(ta);
      ta.select();
      document.execCommand('copy');
      document.body.removeChild(ta);
      window.app?.notify?.('Ссылка скопирована ✓', 'success');
    });
  }

  return { load, openMedia, track: _track, toggleLoop, toggleAutoNext, sharePost };

})();

window.loadActivityFeed = () => Feed.load();

// ── Обработка ссылки ?post=ID (поделиться) ──────────────────────────────────
(function handleShareLink() {
  const params = new URLSearchParams(location.search);
  const postId = params.get('post');
  if (!postId) return;

  history.replaceState(null, '', location.pathname);

  // Показываем main-chat вместо лендинга (гостевой режим)
  function _showGuestFeed() {
    const landing  = document.getElementById('landing-page');
    const mainChat = document.getElementById('main-chat');
    const dock     = document.getElementById('bottom-dock');
    if (!mainChat) return;

    if (landing) landing.classList.add('hidden');
    if (dock)    dock.classList.add('guest-hidden');
    mainChat.classList.remove('hidden');

    // Открываем ленту через layout
    if (typeof window.openFeedPanel === 'function') {
      window.openFeedPanel();
    }
  }

  // Всегда вызывается при закрытии просмотрщика (и гость, и авторизованный)
  function _onClose() {
    const isAuth = !!(window.app && window.app.currentUser);

    // Убираем кнопку «На главную» и восстанавливаем стили панели
    const closeBtn = document.getElementById('guest-close-btn');
    if (closeBtn) closeBtn.remove();
    const panel = document.querySelector('.wall-comments-panel');
    if (panel) {
      const rightCol = panel.querySelector('.flex-col.flex-grow');
      if (rightCol) rightCol.style.display = '';
    }
    const mediaCol = document.getElementById('wall-comments-media-container');
    if (mediaCol) { mediaCol.style.width = ''; mediaCol.style.border = ''; }

    if (isAuth) {
      // Авторизован — просто показываем dock
      const dock = document.getElementById('bottom-dock');
      if (dock) {
        dock.classList.remove('guest-hidden');
        dock.style.display = '';
        // Показываем dock через layout если доступно
        if (typeof window.openFeedPanel !== 'undefined') {
          dock.classList.add('visible');
          setTimeout(() => dock.classList.remove('visible'), 5000);
        }
      }
    } else {
      // Гость — возвращаем лендинг
      const landing  = document.getElementById('landing-page');
      const mainChat = document.getElementById('main-chat');
      const dock     = document.getElementById('bottom-dock');
      if (landing)  landing.classList.remove('hidden');
      if (mainChat) mainChat.classList.add('hidden');
      if (dock)     dock.classList.remove('guest-hidden');
    }
  }

  // Патчим closeWallComments один раз — навсегда
  function _patchClose() {
    const origClose = window.closeWallComments;
    if (!origClose || origClose._guestPatched) return;
    window.closeWallComments = function() {
      origClose();
      setTimeout(_onClose, 350);
    };
    window.closeWallComments._guestPatched = true;
  }

  function _tryOpenPost() {
    const card = document.getElementById('activity-post-' + postId);
    if (!card) return false;
    card.scrollIntoView({ behavior: 'smooth', block: 'center' });
    setTimeout(() => {
      const mediaUrl = card.dataset.mediaUrl;
      const isVideo  = card.dataset.mediaVideo === 'true';
      if (mediaUrl) {
        _patchClose();
        Feed.openMedia(postId, mediaUrl, isVideo);
      }
    }, 400);
    return true;
  }

  function _poll(attempts) {
    if (attempts <= 0) return;
    if (_tryOpenPost()) return;
    setTimeout(() => _poll(attempts - 1), 300);
  }

  function _start() {
    _showGuestFeed();
    setTimeout(() => _poll(30), 700);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => setTimeout(_start, 300));
  } else {
    setTimeout(_start, 300);
  }
})();