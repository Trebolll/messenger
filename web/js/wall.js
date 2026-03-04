// ─── wall.js — логика Стены (Профиля) ──────────────────

function toggleWall() {
    const overlay = document.getElementById('wall-overlay');
    if (overlay.classList.contains('hidden')) {
        openWall();
    } else {
        closeWall();
    }
}

async function openWall(userId = null) {
    const overlay = document.getElementById('wall-overlay');
    const modal   = document.getElementById('wall-modal');
    
    // Если id не передан или совпадает с текущим, открываем свою стену
    const isMe = !userId || String(userId) === String(window.app.currentUser?.id);
    const targetUserId = isMe ? window.app.currentUser?.id : userId;

    if (!targetUserId) return;

    // Скрываем/показываем кнопки редактирования
    const editProfileBtn = document.getElementById('wall-settings-btn');
    const editBioBtn     = document.getElementById('edit-bio-btn');
    // Находим контейнер создания поста
    const postInputContainer = document.getElementById('wall-post-creator');
    
    if (editProfileBtn) editProfileBtn.style.display = isMe ? 'flex' : 'none';
    if (editBioBtn)     editBioBtn.style.display     = isMe ? 'block' : 'none';
    if (postInputContainer) postInputContainer.style.display = isMe ? 'block' : 'none';

    // Очищаем/Заполняем начальные данные
    document.getElementById('wall-username').textContent = isMe ? (window.app.currentUser?.username || 'User') : 'Загрузка...';
    document.getElementById('wall-status').textContent   = isMe ? (window.app.currentUser?.status || '') : '';
    
    const avatarEl = document.getElementById('wall-avatar');
    if (isMe) {
        setAvatarEl(avatarEl, window.app.currentUser);
        const creatorAvatar = document.getElementById('creator-avatar');
        setAvatarEl(creatorAvatar, window.app.currentUser);
    }

    overlay.classList.remove('hidden');
    document.body.classList.add('wall-open');
    const dockBtn = document.getElementById('profile-dock-btn');
    if (dockBtn && isMe) dockBtn.classList.add('active');

    setTimeout(() => {
        overlay.classList.remove('opacity-0');
        modal.classList.remove('scale-95');
    }, 10);
    
    loadWallPosts(targetUserId);
}

function closeWall() {
    const overlay = document.getElementById('wall-overlay');
    const modal   = document.getElementById('wall-modal');
    
    overlay.classList.add('opacity-0');
    modal.classList.add('scale-95');
    document.body.classList.remove('wall-open');
    const dockBtn = document.getElementById('profile-dock-btn');
    if (dockBtn) dockBtn.classList.remove('active');

    setTimeout(() => {
        overlay.classList.add('hidden');
    }, 400);
}

async function loadWallPosts(userId) {
    const feed = document.getElementById('wall-feed');
    const targetId = userId || window.app.currentUser?.id;
    const isMe = String(targetId) === String(window.app.currentUser?.id);

    try {
        const response = await fetch(`/api/wall/${targetId}`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('alpha_token')}`
            }
        });
        const result = await response.json();
        console.log('Wall Data Loaded:', result);
        
        // Обновляем данные владельца стены (Bio, Username, Avatar)
        if (result.wall) {
            const bioText = result.wall.bio || 'Привет! Это мой уголок в λ. Здесь я делюсь мыслями и медиа.';
            document.getElementById('wall-info-bio').textContent = bioText;
            document.getElementById('wall-bio-textarea').value = bioText;
            
            // Всегда обновляем заголовок, если данные есть
            const uname = result.wall.username || (isMe ? window.app.currentUser?.username : 'User');
            const status = result.wall.status || (isMe ? window.app.currentUser?.status : '');
            
            document.getElementById('wall-username').textContent = uname;
            document.getElementById('wall-status').textContent   = status;
            
            const avatarEl = document.getElementById('wall-avatar');
            const avatarUrl = result.wall.avatar_url || (isMe ? window.app.currentUser?.avatar_url : '');
            
            if (avatarUrl) {
                avatarEl.innerHTML = `<img src="${avatarUrl}" class="w-full h-full object-cover">`;
            } else {
                avatarEl.innerHTML = '';
                avatarEl.textContent = (uname || 'U')[0].toUpperCase();
            }

            // Если это я — обновляем аватар в поле создания поста
            if (isMe) {
                const creatorAvatar = document.getElementById('creator-avatar');
                if (creatorAvatar) {
                    if (avatarUrl) {
                        creatorAvatar.innerHTML = `<img src="${avatarUrl}" class="w-full h-full object-cover">`;
                    } else {
                        creatorAvatar.textContent = (uname || 'U')[0].toUpperCase();
                    }
                }
            }
        }

        const posts = result.posts || [];
        
        if (!posts || posts.length === 0) {
            feed.innerHTML = `
                <div class="text-center py-10 opacity-50">
                   <p class="text-xs">На стене пока нет записей...</p>
                </div>
            `;
            return;
        }

        feed.innerHTML = posts.map(p => `
            <div class="wall-post-card flex flex-col gap-3">
               <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-xl bg-custom-sidebar flex items-center justify-center font-bold text-xs overflow-hidden">
                     ${p.author_avatar ? `<img src="${p.author_avatar}" class="w-full h-full object-cover">` : p.author_name[0]}
                  </div>
                  <div>
                     <p class="text-xs font-bold text-custom-main">${p.author_name}</p>
                     <p class="text-[10px] text-custom-muted">${new Date(p.created_at).toLocaleString()}</p>
                  </div>
               </div>
               <div class="text-sm text-custom-main leading-relaxed">
                  ${p.content}
               </div>
               ${p.attachments && p.attachments.length > 0 ? `
                  <div class="grid grid-cols-2 gap-2 mt-2">
                     ${p.attachments.map(a => `
                        <div class="rounded-lg overflow-hidden border border-white/5 bg-black/20 aspect-video">
                           <img src="${a.url}" class="w-full h-full object-cover">
                        </div>
                     `).join('')}
                  </div>
               ` : ''}
               <div class="mt-2 flex gap-4 pt-3 border-t border-white/5">
                  <button class="flex items-center gap-1.5 text-xs text-custom-muted hover:text-custom-accent transition">
                     <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"/></svg>
                     <span>Лайк</span>
                  </button>
                  <button class="flex items-center gap-1.5 text-xs text-custom-muted hover:text-custom-accent transition">
                     <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/></svg>
                     <span>Коммент</span>
                  </button>
               </div>
            </div>
        `).join('');
    } catch (err) {
        console.error('Ошибка загрузки стены:', err);
    }
}

async function publishWallPost() {
    const input = document.getElementById('wall-post-input');
    const content = input.value.trim();
    if (!content) return;

    try {
        const response = await fetch('/api/wall/posts', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('alpha_token')}`
            },
            body: JSON.stringify({ content })
        });

        if (response.ok) {
            window.app.notify('Пост опубликован ✓', 'success');
            input.value = '';
            loadWallPosts();
        } else {
            window.app.notify('Ошибка при публикации', 'error');
        }
    } catch (err) {
        console.error('Ошибка публикации:', err);
        window.app.notify('Ошибка сети', 'error');
    }
}

function toggleBioEdit(show = true) {
    const display = document.getElementById('wall-info-bio');
    const edit    = document.getElementById('wall-info-bio-edit');
    const text    = document.getElementById('wall-bio-textarea');
    const btn     = document.getElementById('edit-bio-btn');

    if (show) {
        display.classList.add('hidden');
        edit.classList.remove('hidden');
        btn.classList.add('hidden');
        text.value = display.textContent.trim();
        text.focus();
    } else {
        display.classList.remove('hidden');
        edit.classList.add('hidden');
        btn.classList.remove('hidden');
    }
}

async function saveBio() {
    const text = document.getElementById('wall-bio-textarea').value.trim();
    if (!text) return;

    try {
        const response = await fetch('/api/wall/settings', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('alpha_token')}`
            },
            body: JSON.stringify({ bio: text })
        });

        if (response.ok) {
            document.getElementById('wall-info-bio').textContent = text;
            toggleBioEdit(false);
            window.app.notify('Информация обновлена ✓', 'success');
        }
    } catch (err) {
        window.app.notify('Ошибка сохранения', 'error');
    }
}
