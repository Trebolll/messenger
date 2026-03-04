// ─── wall.js — логика Стены (Профиля) ──────────────────

function toggleWall() {
    const overlay = document.getElementById('wall-overlay');
    if (overlay.classList.contains('hidden')) {
        openWall();
    } else {
        closeWall();
    }
}

async function openWall() {
    const overlay = document.getElementById('wall-overlay');
    const modal   = document.getElementById('wall-modal');
    
    // Заполняем данные пользователя
    const user = window.app.currentUser;
    if (user) {
        document.getElementById('wall-username').textContent = user.username || 'User';
        document.getElementById('wall-status').textContent   = user.status || 'Статус не установлен';
        
        const avatarEl = document.getElementById('wall-avatar');
        setAvatarEl(avatarEl, user);
        
        const creatorAvatar = document.getElementById('creator-avatar');
        setAvatarEl(creatorAvatar, user);
    }

    overlay.classList.remove('hidden');
    document.body.classList.add('wall-open');
    const dockBtn = document.getElementById('profile-dock-btn');
    if (dockBtn) dockBtn.classList.add('active');

    setTimeout(() => {
        overlay.classList.remove('opacity-0');
        modal.classList.remove('scale-95');
    }, 10);
    
    loadWallPosts();
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

async function loadWallPosts() {
    const feed = document.getElementById('wall-feed');
    const user = window.app.currentUser;
    if (!user) return;

    try {
        const response = await fetch(`/api/wall/${user.id}`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('alpha_token')}`
            }
        });
        const result = await response.json();
        const posts = result.posts || [];
        
        // Обновляем БИО на стене из данных сервера
        if (result.wall && result.wall.bio) {
            document.getElementById('wall-info-bio').textContent = result.wall.bio;
        }

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
