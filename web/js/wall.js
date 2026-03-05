// ─── wall.js — логика Стены (Профиля) ──────────────────

let _wallSelectedFiles = []; // Глобальное хранилище для выбранных файлов перед публикацией

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
    
    // Сбрасываем вкладку при открытии
    switchWallTab('posts', false);
    _wallSelectedFiles = [];
    updateWallMediaPreview();

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
    if (postInputContainer) {
        postInputContainer.style.display = isMe ? 'block' : 'none';
        postInputContainer.dataset.visible = isMe ? 'true' : 'false';
    }

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
        
        // Обновляем данные владельца стены (Username, Avatar)
        if (result.wall) {
            // Всегда обновляем заголовок, если данные есть
            const uname = result.wall.username || (isMe ? window.app.currentUser?.username : 'User');
            const status = result.wall.status || (isMe ? window.app.currentUser?.status : '');
            
            document.getElementById('wall-username').textContent = uname;
            document.getElementById('wall-status').textContent   = status;

            // Обновляем статистику пользователя (место, дата, лайки)
            const locEl = document.getElementById('wall-info-location');
            const locRow = document.getElementById('wall-info-location-row');
            if (result.wall.user_location) {
                locEl.textContent = result.wall.user_location;
                if (locRow) locRow.classList.remove('hidden');
            } else {
                if (locRow) locRow.classList.add('hidden');
            }

            const createdEl = document.getElementById('wall-info-created');
            if (createdEl && result.wall.user_created_at) {
                const date = new Date(result.wall.user_created_at);
                const y = String(date.getFullYear()).slice(-2);
                const m = String(date.getMonth() + 1).padStart(2, '0');
                const d = String(date.getDate()).padStart(2, '0');
                createdEl.textContent = `${y}.${m}.${d}`;
            }

            // Профессия
            const profEl = document.getElementById('wall-info-profession');
            const profRow = document.getElementById('wall-info-profession-row');
            if (result.wall.user_profession) {
                profEl.textContent = result.wall.user_profession;
                if (profRow) profRow.classList.remove('hidden');
            } else {
                if (profRow) profRow.classList.add('hidden');
            }

            // Возраст (из birth_date)
            const ageEl = document.getElementById('wall-info-age');
            const ageRow = document.getElementById('wall-info-age-row');
            if (result.wall.user_birth_date) {
                const age = calculateAge(new Date(result.wall.user_birth_date));
                ageEl.textContent = `${age} лет`;
                if (ageRow) ageRow.classList.remove('hidden');
            } else {
                if (ageRow) ageRow.classList.add('hidden');
            }

            const likesEl = document.getElementById('wall-info-likes');
            if (likesEl) {
                likesEl.textContent = result.wall.user_rating || 0;
            }

            renderMediaGrid(result.media || [], isMe);
            
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
        } else {
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
                      <div class="grid ${p.attachments.length === 1 ? 'grid-cols-1' : 'grid-cols-2'} gap-2 mt-2">
                         ${p.attachments.map(a => {
                            const isVideo = (a.mime_type || '').startsWith('video/');
                            if (isVideo) {
                                return `<div class="rounded-lg overflow-hidden border border-white/5 bg-black/20 relative aspect-video">
                                    <video src="${a.url}" class="w-full h-full object-cover" controls preload="metadata"></video>
                                </div>`;
                            }
                            return `<div class="rounded-lg overflow-hidden border border-white/5 bg-black/20 aspect-video cursor-pointer" onclick="openImgLightbox('${a.url}')">
                               <img src="${a.url}" class="w-full h-full object-cover">
                            </div>`;
                         }).join('')}
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
        }
    } catch (err) {
        console.error('Ошибка загрузки стены:', err);
    }
}

async function publishWallPost() {
    const input = document.getElementById('wall-post-input');
    const content = input.value.trim();
    
    // Можно публиковать если есть текст ИЛИ если есть файлы
    if (!content && _wallSelectedFiles.length === 0) return;

    try {
        // 1. Создаем пост
        const response = await fetch('/api/wall/posts', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('alpha_token')}`
            },
            body: JSON.stringify({ content: content || "" })
        });

        if (!response.ok) throw new Error('Ошибка создания поста');
        const post = await response.json();

        // 2. Если есть файлы — загружаем их по одному и привязываем к посту
        if (_wallSelectedFiles.length > 0) {
            const uploadPromises = _wallSelectedFiles.map(file => {
                const formData = new FormData();
                formData.append('file', file);
                
                return fetch(`/api/wall/posts/${post.id}/attachments`, {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('alpha_token')}`
                    },
                    body: formData
                });
            });
            
            // Ждем завершения всех загрузок перед обновлением интерфейса
            await Promise.all(uploadPromises);
        }

        window.app.notify('Опубликовано ✓', 'success');
        input.value = '';
        _wallSelectedFiles = [];
        updateWallMediaPreview();
        loadWallPosts();
    } catch (err) {
        console.error('Ошибка публикации:', err);
        window.app.notify('Ошибка при публикации', 'error');
    }
}

async function publishWallMediaDirect(e) {
    const files = Array.from(e.target.files);
    if (!files.length) return;

    try {
        // 1. Создаем пустой пост для медиа
        const response = await fetch('/api/wall/posts', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('alpha_token')}`
            },
            body: JSON.stringify({ content: "" })
        });

        if (!response.ok) throw new Error('Ошибка создания поста: ' + response.status);
        const post = await response.json();
        if (!post || !post.id) throw new Error('Сервер не вернул ID поста');

        // 2. Загружаем файлы ПО ОЧЕРЕДИ (последовательно), чтобы избежать TypeError / Network Error
        for (const file of files) {
            const formData = new FormData();
            formData.append('file', file);
            const uploadRes = await fetch(`/api/wall/posts/${post.id}/attachments`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('alpha_token')}`
                },
                body: formData
            });
            if (!uploadRes.ok) console.warn('Файл не загружен:', file.name);
        }

        window.app.notify('Медиа опубликовано ✓', 'success');
        e.target.value = '';
        
        // Перезагружаем данные (текущий userId берем из URL или глобального состояния)
        // Для простоты вызовем loadWallPosts без ID (он сам подставит текущего юзера)
        loadWallPosts();
    } catch (err) {
        console.error('Ошибка быстрой публикации медиа:', err);
        window.app.notify('Ошибка при загрузке', 'error');
    }
}

function handleWallMediaSelect(e) {
    const files = Array.from(e.target.files);
    if (!files.length) return;

    // Добавляем к текущим (лимит например 10)
    _wallSelectedFiles = [..._wallSelectedFiles, ...files].slice(0, 10);
    updateWallMediaPreview();
    e.target.value = ''; // сброс для возможности выбора того же файла
}

function updateWallMediaPreview() {
    const container = document.getElementById('wall-media-preview');
    if (!container) return;

    if (_wallSelectedFiles.length === 0) {
        container.classList.add('hidden');
        container.innerHTML = '';
        return;
    }

    container.classList.remove('hidden');
    container.innerHTML = _wallSelectedFiles.map((file, idx) => {
        const isVideo = file.type.startsWith('video/');
        const url = URL.createObjectURL(file);
        return `
            <div class="relative w-20 h-20 flex-shrink-0 rounded-lg overflow-hidden border border-white/10 bg-black/40 group">
                ${isVideo ? `<video src="${url}" class="w-full h-full object-cover"></video>` : `<img src="${url}" class="w-full h-full object-cover">`}
                <button onclick="removeWallSelectedFile(${idx})" class="absolute top-1 right-1 w-5 h-5 bg-black/60 text-white rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                    <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
                </button>
            </div>
        `;
    }).join('');
}

function removeWallSelectedFile(index) {
    _wallSelectedFiles.splice(index, 1);
    updateWallMediaPreview();
}

function switchWallTab(tab, reload = true) {
    const postsBtn = document.getElementById('wall-tab-posts');
    const mediaBtn = document.getElementById('wall-tab-media');
    const feed = document.getElementById('wall-feed');
    const mediaContainer = document.getElementById('wall-media-container');
    const creator = document.getElementById('wall-post-creator');

    const sidebarInfo = document.getElementById('wall-info-sidebar-box');
    const sidebarMedia = document.getElementById('wall-media-preview-box');

    if (tab === 'posts') {
        postsBtn.classList.add('text-custom-main', 'border-custom-accent');
        postsBtn.classList.remove('text-custom-muted', 'border-transparent');
        mediaBtn.classList.remove('text-custom-main', 'border-custom-accent');
        mediaBtn.classList.add('text-custom-muted', 'border-transparent');
        
        feed.classList.remove('hidden');
        mediaContainer.classList.add('hidden');
        if (sidebarInfo) sidebarInfo.classList.remove('hidden');
        if (sidebarMedia) sidebarMedia.classList.remove('hidden');
    } else {
        mediaBtn.classList.add('text-custom-main', 'border-custom-accent');
        mediaBtn.classList.remove('text-custom-muted', 'border-transparent');
        postsBtn.classList.remove('text-custom-main', 'border-custom-accent');
        postsBtn.classList.add('text-custom-muted', 'border-transparent');
        
        feed.classList.add('hidden');
        mediaContainer.classList.remove('hidden');
        if (sidebarInfo) sidebarInfo.classList.add('hidden');
        if (sidebarMedia) sidebarMedia.classList.add('hidden');
    }

    // Блок создания поста показываем только если это СВОЯ стена (isMe) И только во вкладке "Записи"
    const isMe = creator && creator.dataset.visible !== 'false';
    if (isMe && tab === 'posts') {
        creator.classList.remove('hidden');
    } else if (creator) {
        creator.classList.add('hidden');
    }
}

function renderMediaGrid(media, isMe = false) {
    const grid = document.getElementById('wall-media-grid-main');
    if (!grid) return;

    // Сначала формируем HTML для кнопки добавления (только для владельца)
    let gridHtml = '';
    if (isMe) {
        gridHtml = `
            <div id="media-add-btn-wrapper" class="aspect-square rounded-xl border-2 border-dashed border-white/10 hover:border-custom-accent/40 flex items-center justify-center cursor-pointer transition-all group" onclick="document.getElementById('wall-media-direct-upload').click()">
                 <input type="file" id="wall-media-direct-upload" class="hidden" accept="image/*,video/*" multiple onchange="publishWallMediaDirect(event)">
                 <div class="w-12 h-12 rounded-full bg-white/5 flex items-center justify-center group-hover:bg-custom-accent/10 transition-colors">
                    <svg class="w-6 h-6 text-custom-muted group-hover:text-custom-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
                 </div>
            </div>
        `;
    }

    if (!media || media.length === 0) {
        grid.innerHTML = gridHtml + (isMe ? '' : `<div class="col-span-full py-10 text-center opacity-40 text-xs italic">Медиа-файлов пока нет</div>`);
        return;
    }

    grid.innerHTML = gridHtml + media.map(m => {
        const isVideo = (m.mime_type || '').startsWith('video/');
        if (isVideo) {
            return `
                <div class="aspect-square rounded-xl overflow-hidden bg-black/20 border border-white/5 relative group cursor-pointer" onclick="openVideoPlayer('${m.url}')">
                    <video src="${m.url}" class="w-full h-full object-cover"></video>
                    <div class="absolute inset-0 flex items-center justify-center bg-black/20 opacity-0 group-hover:opacity-100 transition-opacity">
                        <svg class="w-8 h-8 text-white" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
                    </div>
                </div>
            `;
        }
        return `
            <div class="aspect-square rounded-xl overflow-hidden bg-black/20 border border-white/5 cursor-pointer hover:scale-[1.02] transition-transform" onclick="openImgLightbox('${m.url}')">
                <img src="${m.url}" class="w-full h-full object-cover" loading="lazy">
            </div>
        `;
    }).join('');
}

function openVideoPlayer(url) {
    // Используем встроенный лайтбокс или просто создаем оверлей
    // Для простоты можно использовать тот же lightbox если он поддерживает видео, или отдельный.
    // Пока просто откроем в новом окне или используем существующий render.js функционал если он есть.
    window.open(url, '_blank');
}

function calculateAge(birthDate) {
    const today = new Date();
    let age = today.getFullYear() - birthDate.getFullYear();
    const m = today.getMonth() - birthDate.getMonth();
    if (m < 0 || (m === 0 && today.getDate() < birthDate.getDate())) {
        age--;
    }
    return age;
}
