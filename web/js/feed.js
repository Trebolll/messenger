// ─── feed.js — Лента активности + медиа-навигация ───────────────────────

const Feed = (() => {

  // ── Состояние ────────────────────────────────────────────────────────
  let _mediaList        = [];   // [{postId, url, isVideo}]
  let _mediaIndex       = 0;
  let _mediaWheelLocked = false;
  let _observer         = null;

  // ── Загрузка и рендер ленты ──────────────────────────────────────────
  async function load() {
    const container = document.getElementById('activity-feed-container');
    if (!container) return;

    _initSnapScroll(container);

    try {
      const response = await fetch('/api/wall/feed', {
        headers: { 'Authorization': `Bearer ${localStorage.getItem('alpha_token')}` }
      });
      if (!response.ok) throw new Error('Failed to load feed');
      const posts = await response.json();

      if (!posts || posts.length === 0) {
        container.innerHTML = `
                    <div class="flex flex-col items-center justify-center h-full py-20 opacity-40 select-none">
                        <p class="text-sm font-semibold text-custom-main">В ленте пока пусто...</p>
                    </div>`;
        return;
      }

      _render(container, posts);
    } catch (err) {
      console.error('Feed error:', err);
      container.innerHTML = `<p class="text-xs text-red-400 p-4">Ошибка загрузки ленты</p>`;
    }
  }

  function _render(container, posts) {
    // Отключаем старый observer если был
    if (_observer) { _observer.disconnect(); _observer = null; }

    container.innerHTML = '';

    posts.forEach((p, i) => {
      const wrapper = document.createElement('div');
      wrapper.style.cssText = 'flex-shrink:0;width:100%;height:100%;scroll-snap-align:start;';

      if (i === 0) {
        wrapper.innerHTML = _renderPost(p);
      } else {
        wrapper.innerHTML = `<div class="w-full h-full bg-custom-sidebar/40 animate-pulse flex items-center justify-center">
                    <div class="w-12 h-12 rounded-full bg-white/5"></div></div>`;
        wrapper.dataset.postData = JSON.stringify(p);
      }
      container.appendChild(wrapper);
    });

    // Lazy render через IntersectionObserver
    _observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (!entry.isIntersecting) return;
        const wrapper = entry.target;
        if (wrapper.dataset.postData) {
          const p = JSON.parse(wrapper.dataset.postData);
          wrapper.innerHTML = _renderPost(p);
          delete wrapper.dataset.postData;
        }
        _observer.unobserve(wrapper);
      });
    }, { root: container, threshold: 0.05 });

    Array.from(container.children).forEach((child, i) => {
      if (i > 0) _observer.observe(child);
    });
  }

  function _renderPost(p) {
    const hasAttachments = p.attachments && p.attachments.length > 0;
    const openAction = hasAttachments
        ? `Feed.openMedia('${p.id}', '${p.attachments[0].url}', ${(p.attachments[0].mime_type || '').startsWith('video/')})`
        : `openPostChat('${p.id}', '${p.chat_id || ''}')`;

    return `
        <div class="activity-post-card group relative overflow-hidden bg-custom-sidebar" id="activity-post-${p.id}"
             style="width:100%;height:100%;">
           ${hasAttachments
        ? `<div class="w-full h-full overflow-hidden">${_renderAttachment(p.id, p.attachments[0])}${p.attachments.length > 1
            ? `<div class="absolute top-3 right-3 bg-black/50 backdrop-blur-md text-white text-[9px] px-2 py-1 rounded-full z-10">+${p.attachments.length - 1}</div>` : ''}</div>`
        : `<div class="w-full h-full p-4 flex items-center justify-center text-center bg-custom-sidebar/50">
                     <div class="text-[11px] text-custom-main leading-relaxed line-clamp-6 font-medium italic">${p.content || 'Запись без текста'}</div>
                  </div>`}

           <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-transparent opacity-0 group-hover:opacity-100 transition-opacity flex flex-col justify-end p-3">
              <div class="flex items-center gap-2 mb-1">
                 <div class="w-5 h-5 rounded-full bg-white/20 flex items-center justify-center font-bold text-[8px] overflow-hidden border border-white/20">
                    ${p.author_avatar ? `<img src="${p.author_avatar}" class="w-full h-full object-cover">` : (p.author_name?.[0] ?? '?')}
                 </div>
                 <span class="text-[10px] font-bold text-white truncate">${p.author_name}</span>
              </div>
              <div class="flex gap-3 text-white/80">
                 <div class="flex items-center gap-1 text-[9px]">
                    <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 24 24"><path d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"/></svg>
                    <span>${p.likes_count || 0}</span>
                 </div>
                 <div class="flex items-center gap-1 text-[9px]">
                    <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/></svg>
                    <span>${p.comments_count || 0}</span>
                 </div>
              </div>
              <button onclick="${openAction}" class="absolute inset-0 z-0"></button>
              <div class="absolute top-3 left-3 opacity-60 pointer-events-none">
                  <p class="text-[8px] text-white/50">${new Date(p.created_at).toLocaleDateString()}</p>
              </div>
           </div>
        </div>`;
  }

  function _renderAttachment(postId, a) {
    const isVideo = (a.mime_type || '').startsWith('video/');
    if (isVideo) {
      const thumbId = `vthumb-${postId}-${Math.random().toString(36).slice(2, 7)}`;
      return `<div class="relative w-full h-full cursor-pointer bg-black/40 overflow-hidden"
                         onclick="Feed.openMedia('${postId}', '${a.url}', true)">
                <canvas id="${thumbId}" class="w-full h-full" style="display:block;object-fit:cover;"></canvas>
                <video class="hidden" src="${a.url}" preload="metadata" muted playsinline
                    onloadeddata="(function(v){var c=document.getElementById('${thumbId}');if(!c)return;v.currentTime=0.5;v.onseeked=function(){c.width=v.videoWidth||320;c.height=v.videoHeight||180;c.getContext('2d').drawImage(v,0,0,c.width,c.height);v.remove();};})( this)">
                </video>
                <div class="absolute inset-0 flex items-center justify-center bg-black/20 pointer-events-none">
                    <svg class="w-7 h-7 text-white opacity-80 drop-shadow" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
                </div>
            </div>`;
    }
    return `<div class="w-full h-full cursor-pointer bg-black/40 overflow-hidden"
                     onclick="Feed.openMedia('${postId}', '${a.url}', false)">
            <img src="${a.url}" class="w-full h-full object-cover" loading="lazy" decoding="async">
        </div>`;
  }

  // ── Snap-скролл ──────────────────────────────────────────────────────
  function _initSnapScroll(container) {
    if (container._wheelListenerAdded) return;
    container._wheelListenerAdded = true;
    let locked = false;
    container.addEventListener('wheel', (e) => {
      e.preventDefault();
      if (locked) return;
      locked = true;
      const h = container.clientHeight;
      const cur = Math.round(container.scrollTop / h);
      const next = e.deltaY > 0
          ? Math.min(cur + 1, container.children.length - 1)
          : Math.max(cur - 1, 0);
      container.scrollTo({ top: next * h, behavior: 'smooth' });
      setTimeout(() => { locked = false; }, 500);
    }, { passive: false });
  }

  // ── Медиа-навигация ──────────────────────────────────────────────────
  async function openMedia(postId, url, isVideo) {
    _mediaList = [];

    const wallGrid   = document.getElementById('wall-media-grid-main');
    const wallOverlay = document.getElementById('wall-overlay');
    const feedContainer = document.getElementById('activity-feed-container');
    const wallIsOpen = wallGrid && wallOverlay && !wallOverlay.classList.contains('hidden');

    if (wallIsOpen) {
      wallGrid.querySelectorAll('.wall-media-item').forEach(item => {
        const onclick = item.getAttribute('onclick') || '';
        const m = onclick.match(/Feed\.openMedia\('([^']+)',\s*'([^']+)',\s*(true|false)\)/);
        // поддержка старого вызова тоже
        const m2 = !m && onclick.match(/openMediaDetail\('([^']+)',\s*'([^']+)',\s*(true|false)\)/);
        const match = m || m2;
        if (match) _mediaList.push({ postId: match[1], url: match[2], isVideo: match[3] === 'true' });
      });
    } else if (feedContainer) {
      feedContainer.querySelectorAll('[onclick*="openMedia"]').forEach(item => {
        const onclick = item.getAttribute('onclick') || '';
        const m = onclick.match(/Feed\.openMedia\('([^']+)',\s*'([^']+)',\s*(true|false)\)/);
        if (m) _mediaList.push({ postId: m[1], url: m[2], isVideo: m[3] === 'true' });
      });
    }

    _mediaIndex = _mediaList.findIndex(m => m.postId === String(postId) && m.url === url);
    if (_mediaIndex === -1) _mediaIndex = 0;

    await _showMedia(postId, url, isVideo, null);
  }

  async function _showMedia(postId, url, isVideo, direction) {
    const mediaContainer = document.getElementById('wall-comments-media-container');
    const mediaContent   = document.getElementById('wall-comments-media-content');
    const likeBtn        = document.getElementById('wall-comments-media-like');

    mediaContainer.classList.remove('hidden');

    // Анимация
    if (direction !== null && mediaContent) {
      mediaContent.style.transition = 'opacity 0.15s ease, transform 0.15s ease';
      mediaContent.style.opacity = '0';
      mediaContent.style.transform = `translateY(${direction > 0 ? '-20px' : '20px'})`;
      await new Promise(r => setTimeout(r, 150));
    }

    if (mediaContent) {
      mediaContent.innerHTML = isVideo
          ? `<video src="${url}" class="max-w-full max-h-full rounded-lg shadow-2xl" controls autoplay></video>`
          : `<img src="${url}" class="max-w-full max-h-full rounded-lg shadow-2xl object-contain">`;

      const fromY = direction === null ? '0' : direction > 0 ? '20px' : '-20px';
      mediaContent.style.opacity = '0';
      mediaContent.style.transform = `translateY(${fromY})`;
      requestAnimationFrame(() => {
        mediaContent.style.transition = 'opacity 0.22s ease, transform 0.22s ease';
        mediaContent.style.opacity = '1';
        mediaContent.style.transform = 'translateY(0)';
      });
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

    // Комментарии
    try {
      const res = await fetch(`/api/wall/posts/${postId}/chat`, {
        headers: { 'Authorization': `Bearer ${localStorage.getItem('alpha_token')}` }
      }).then(r => r.json());
      if (res.chat_id) openWallComments(postId, res.chat_id, true);
    } catch (e) {}

    // Wheel listener на overlay (один раз)
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

    _renderIndicator(mediaContainer);
  }

  function _renderIndicator(mediaContainer) {
    if (_mediaList.length < 2) return;
    let el = document.getElementById('media-nav-indicator');
    if (!el) {
      el = document.createElement('div');
      el.id = 'media-nav-indicator';
      el.style.cssText = 'position:absolute;right:14px;top:50%;transform:translateY(-50%);display:flex;flex-direction:column;gap:5px;z-index:20;pointer-events:none;';
      mediaContainer.appendChild(el);
    }
    el.innerHTML = _mediaList.map((_, i) => `
            <div style="width:4px;height:${i === _mediaIndex ? '20px' : '6px'};border-radius:2px;
                 background:${i === _mediaIndex ? 'rgba(255,255,255,0.9)' : 'rgba(255,255,255,0.25)'};
                 transition:all 0.25s ease;"></div>
        `).join('');
  }

  // ── Публичное API ────────────────────────────────────────────────────
  return { load, openMedia };

})();

// Глобальные алиасы для обратной совместимости с вызовами из wall.js
window.loadActivityFeed = () => Feed.load();