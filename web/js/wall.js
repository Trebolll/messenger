// ─── wall.js — логика Стены (Профиля) ──────────────────

let _wallSelectedFiles = [];   // Глобальное хранилище для выбранных файлов перед публикацией
let _wallMode = 'posts';        // Текущий режим: 'posts' | 'media'
let _wallUserId = null;         // userId текущей открытой стены
let _wallPostsWS = null;        // WebSocket для обновлений стены (новые посты)

// ── Проверка авторизации для интерактивных действий ────────────────────────
function requireAuth() {
    if (!window.app?.currentUser) {
        window.app?.notify?.('Вы не авторизованы', 'warning');
        return false;
    }
    return true;
}

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
    _wallSelectedFiles = [];
    updateWallMediaPreview();

    // Если id не передан или совпадает с текущим, открываем свою стену
    const isMe = !userId || String(userId) === String(window.app.currentUser?.id);
    const targetUserId = isMe ? window.app.currentUser?.id : userId;

    if (!targetUserId) return;
    _wallUserId = targetUserId;

    // Скрываем/показываем кнопки редактирования
    const editProfileBtn = document.getElementById('wall-settings-btn');
    const editBioBtn     = document.getElementById('edit-bio-btn');
    const postInputContainer = document.getElementById('wall-post-creator');
    const wallStatusEl       = document.getElementById('wall-status');

    if (editProfileBtn) editProfileBtn.style.display = isMe ? 'flex' : 'none';
    if (editBioBtn)     editBioBtn.style.display     = isMe ? 'block' : 'none';
    if (postInputContainer) {
        postInputContainer.style.display = isMe ? 'block' : 'none';
        postInputContainer.dataset.visible = isMe ? 'true' : 'false';
    }
    // Статус кликабельный только на своей стене
    if (wallStatusEl) {
        if (isMe) {
            wallStatusEl.onclick = startWallStatusEdit;
            wallStatusEl.style.cursor = 'pointer';
        } else {
            wallStatusEl.onclick = null;
            wallStatusEl.style.cursor = 'default';
        }
    }

    // Очищаем начальные данные
    document.getElementById('wall-username').textContent = isMe ? (window.app.currentUser?.username || 'User') : 'Загрузка...';
    document.getElementById('wall-status').textContent   = isMe ? (window.app.currentUser?.status || '...') : '';

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

    // Переключаем в нужный режим без reload, потом грузим данные
    _setWallModeUI('posts');
    loadWallPosts(targetUserId);
    connectWallPostsWS(targetUserId);
}

function closeWall() {
    const overlay = document.getElementById('wall-overlay');
    const modal   = document.getElementById('wall-modal');

    if (_wallPostsWS) {
        _wallPostsWS.close();
        _wallPostsWS = null;
    }

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

        // Обновляем данные владельца стены (Username, Avatar)
        if (result.wall) {
            const uname = result.wall.username || (isMe ? window.app.currentUser?.username : 'User');
            const status = result.wall.status || (isMe ? window.app.currentUser?.status : '');

            document.getElementById('wall-username').textContent = uname;
            document.getElementById('wall-status').textContent   = status || '...';

            // Био
            renderWallBio(result.wall.bio || '', isMe);

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
            const createdRow = document.getElementById('wall-info-created-row');
            if (createdEl && result.wall.user_created_at) {
                const date = new Date(result.wall.user_created_at);
                const y = date.getFullYear();
                const m = String(date.getMonth() + 1).padStart(2, '0');
                const d = String(date.getDate()).padStart(2, '0');
                createdEl.textContent = `${d}.${m}.${y}`;
                if (createdRow) createdRow.classList.remove('hidden');
            } else {
                if (createdRow) createdRow.classList.add('hidden');
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

            // Лайки постов стены
            const likesEl = document.getElementById('wall-info-likes');
            const likesRow = document.getElementById('wall-info-likes-row');
            const wallLikes = result.wall.total_wall_likes || 0;
            if (wallLikes > 0) {
                likesEl.textContent = `${wallLikes} лайк${wallLikes === 1 ? '' : wallLikes < 5 ? 'а' : 'ов'}`;
                if (likesRow) likesRow.classList.remove('hidden');
            } else {
                if (likesRow) likesRow.classList.add('hidden');
            }

            // Votes (рейтинг сообщений)
            const votesEl = document.getElementById('wall-info-votes');
            const votesRow = document.getElementById('wall-info-votes-row');
            const userRating = result.wall.user_rating || 0;
            if (userRating !== 0) {
                votesEl.textContent = `${userRating > 0 ? '+' : ''}${userRating} vote`;
                if (votesRow) votesRow.classList.remove('hidden');
            } else {
                if (votesRow) votesRow.classList.add('hidden');
            }

            renderMediaGrid(result.media || [], isMe);

            // Счётчик медиа на кнопке в сайдбаре
            const mediaCount = (result.media || []).length;
            const mediaCountEl = document.getElementById('wall-media-count');
            if (mediaCountEl) mediaCountEl.textContent = mediaCount > 0 ? `${mediaCount} файлов` : 'фото и видео';

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
        renderWallPostsFeed(posts, isMe);

    } catch (err) {
        console.error('Ошибка загрузки стены:', err);
    }
}

function renderWallPostsFeed(posts, isMe) {
    const feed = document.getElementById('wall-feed');
    // Показываем в ленте только посты с текстом или смешанным контентом
    const feedPosts = posts.filter(p => {
        const hasText = (p.content || '').trim().length > 0;
        const hasNonMedia = (p.attachments || []).some(a => !(a.mime_type || '').match(/^(image|video)\//));
        return hasText || hasNonMedia;
    });

    if (!feedPosts || feedPosts.length === 0) {
        feed.innerHTML = `
            <div class="text-center py-10 opacity-50">
               <p class="text-xs">На стене пока нет записей...</p>
            </div>
        `;
    } else {
        feed.innerHTML = feedPosts.map(p => renderSinglePostHtml(p, isMe)).join('');
    }
}

function renderSinglePostHtml(p, isMe) {
    return `
        <div class="wall-post-card flex flex-col gap-3" id="post-${p.id}">
           <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-xl bg-custom-sidebar flex items-center justify-center font-bold text-xs overflow-hidden">
                 ${p.author_avatar ? `<img src="${p.author_avatar}" class="w-full h-full object-cover">` : p.author_name[0]}
              </div>
              <div class="flex-grow">
                 <p class="text-xs font-bold text-custom-main">${p.author_name}</p>
                 <p class="text-[10px] text-custom-muted">${new Date(p.created_at).toLocaleString()}</p>
              </div>
              ${isMe ? `
              <button onclick="deleteWallPost('${p.id}')" class="w-7 h-7 rounded-xl flex items-center justify-center text-custom-muted/40 hover:text-red-400 hover:bg-red-400/10 transition-all opacity-0 group-hover:opacity-100 wall-post-delete-btn" title="Удалить запись">
                 <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
              </button>` : ''}
           </div>
           <div class="text-sm text-custom-main leading-relaxed">
              ${formatMessageContent(p.content)}
           </div>
           ${p.attachments && p.attachments.length > 0 ? `
              <div class="grid ${p.attachments.length === 1 ? 'grid-cols-1' : (p.attachments.length === 2 ? 'grid-cols-2' : 'grid-cols-3')} gap-2 mt-2 post-attachments-grid">
                 ${p.attachments.map(a => renderAttachmentHtml(p.id, a)).join('')}
              </div>
           ` : `<div class="post-attachments-grid grid grid-cols-2 gap-2 mt-2 hidden"></div>`}
           <div class="mt-2 flex gap-4 pt-3 border-t border-white/5">
              <button onclick="togglePostLike('${p.id}', this)" data-liked="${p.is_liked}" class="flex items-center gap-1.5 text-xs transition ${p.is_liked ? 'text-red-400' : 'text-custom-muted hover:text-red-400'}">
                 <svg class="w-4 h-4 transition-transform" fill="${p.is_liked ? 'currentColor' : 'none'}" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"/></svg>
                 <span class="like-count">${p.likes_count || ''}</span>
              </button>
              <button onclick="openPostChat('${p.id}', '${p.chat_id || ''}')" class="flex items-center gap-1.5 text-xs text-custom-muted hover:text-custom-accent transition">
                 <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/></svg>
                 <span class="comment-count">${p.comments_count || 0} комментариев</span>
              </button>
              <!-- Кнопки шаринга справа -->
              <div class="ml-auto flex gap-2">
                 <!-- Копировать ссылку -->
                 <button onclick="copyPostLink('${p.id}', this)" title="Копировать ссылку"
                    class="flex items-center gap-1 text-xs text-custom-muted hover:text-custom-accent transition px-2 py-1 rounded-lg hover:bg-white/5">
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/></svg>
                 </button>
                 <!-- Поделиться в Telegram -->
                 <button onclick="shareToTelegram('${p.id}')" title="Поделиться в Telegram"
                    class="flex items-center gap-1 text-xs text-custom-muted hover:text-[#2AABEE] transition px-2 py-1 rounded-lg hover:bg-[#2AABEE]/10">
                    <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm5.562 8.248l-1.97 9.289c-.145.658-.537.818-1.084.508l-3-2.21-1.447 1.394c-.16.16-.295.295-.605.295l.213-3.053 5.56-5.023c.242-.213-.054-.333-.373-.12l-6.871 4.326-2.962-.924c-.643-.204-.657-.643.136-.953l11.57-4.461c.537-.194 1.006.131.833.932z"/></svg>
                 </button>
              </div>
           </div>
        </div>
    `;
}

function renderAttachmentHtml(postId, a) {
    const isVideo = (a.mime_type || '').startsWith('video/');
    if (isVideo) {
        return `<div class="rounded-lg overflow-hidden border border-white/5 bg-black/20 relative aspect-video cursor-pointer" onclick="Feed.openMedia('${postId}', '${a.url}', true)">
            <video src="${a.url}" class="w-full h-full object-cover" preload="metadata" onloadedmetadata="setVideoPoster(this)" onclick="event.stopPropagation()"></video>
            <div class="absolute inset-0 flex items-center justify-center bg-black/20 pointer-events-none">
                <svg class="w-10 h-10 text-white opacity-80" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
            </div>
        </div>`;
    }
    return `<div class="rounded-lg overflow-hidden border border-white/5 bg-black/20 aspect-video cursor-pointer" onclick="Feed.openMedia('${postId}', '${a.url}', false)">
       <img src="${a.url}" class="w-full h-full object-cover">
    </div>`;
}

function connectWallPostsWS(userId) {
    if (_wallPostsWS) _wallPostsWS.close();

    const token = localStorage.getItem('alpha_token') || '';
    const proto = location.protocol === 'https:' ? 'wss' : 'ws';
    _wallPostsWS = new WebSocket(`${proto}://${location.host}/ws/wall-posts/${userId}?token=${token}`);

    _wallPostsWS.onmessage = (e) => {
        const msg = JSON.parse(e.data);
        const feed = document.getElementById('wall-feed');
        const isMe = String(userId) === String(window.app.currentUser?.id);

        if (msg.type === 'new_post') {
            const placeholder = feed.querySelector('.text-center.py-10');
            if (placeholder) placeholder.remove();

            const html = renderSinglePostHtml(msg.post, isMe);
            feed.insertAdjacentHTML('afterbegin', html);
        } else if (msg.type === 'delete_post') {
            const card = document.getElementById(`post-${msg.post_id}`) || document.getElementById(`activity-post-${msg.post_id}`);
            if (card) {
                card.style.opacity = '0';
                card.style.transform = 'scale(0.95)';
                setTimeout(() => card.remove(), 300);
            }
        } else if (msg.type === 'update_post_attachment') {
            if (String(_wallUserId) === String(userId)) {
                loadWallPosts(_wallUserId);
            }
            if (document.getElementById(`activity-post-${msg.post_id}`)) {
                loadActivityFeed();
            }
        } else if (msg.type === 'update_post_like') {
            const syncIds = [`post-${msg.post_id}`, `activity-post-${msg.post_id}`];
            syncIds.forEach(id => {
                const card = document.getElementById(id);
                if (card) {
                    const span = card.querySelector('.like-count, span');
                    if (span) span.textContent = msg.likes_count || '';
                }
            });
            // Обновляем общий счетчик лайков в сайдбаре если мы на этой стене
            if (msg.total_wall_likes !== undefined && String(_wallUserId) === String(userId)) {
                const likesEl = document.getElementById('wall-info-likes');
                const likesRow = document.getElementById('wall-info-likes-row');
                const wallLikes = msg.total_wall_likes;
                if (wallLikes > 0) {
                    if (likesEl) likesEl.textContent = `${wallLikes} лайк${wallLikes === 1 ? '' : wallLikes < 5 ? 'а' : 'ов'}`;
                    if (likesRow) likesRow.classList.remove('hidden');
                } else if (likesRow) {
                    likesRow.classList.add('hidden');
                }
            }
        } else if (msg.type === 'update_post_comment_count') {
            const btns = document.querySelectorAll(`button[onclick*="${msg.chat_id}"]`);
            btns.forEach(btn => {
                const card = btn.closest('.wall-post-card, .activity-post-card');
                if (card) {
                    const span = card.querySelector('.comment-count, span:last-child');
                    if (span) {
                        if (card.classList.contains('wall-post-card')) {
                            span.textContent = `${msg.comments_count} комментариев`;
                        } else {
                            span.textContent = msg.comments_count;
                        }
                    }
                }
            });
        } else if (msg.type === 'update_wall_info') {
            if (String(_wallUserId) === String(userId)) {
                renderWallBio(msg.bio, String(userId) === String(window.app.currentUser?.id));
            }
        } else if (msg.type === 'update_wall_status') {
            if (String(_wallUserId) === String(userId)) {
                const statusEl = document.getElementById('wall-status');
                if (statusEl) statusEl.textContent = msg.status || '';
            }
        } else if (msg.type === 'update_wall_info_full') {
            if (String(_wallUserId) === String(userId)) {
                if (msg.username) document.getElementById('wall-username').textContent = msg.username;
                if (msg.status !== undefined) document.getElementById('wall-status').textContent = msg.status || '';
                if (msg.location !== undefined) {
                    const locEl = document.getElementById('wall-info-location');
                    const locRow = document.getElementById('wall-info-location-row');
                    if (msg.location) {
                        if (locEl) locEl.textContent = msg.location;
                        if (locRow) locRow.classList.remove('hidden');
                    } else if (locRow) {
                        locRow.classList.add('hidden');
                    }
                }
                if (msg.profession !== undefined) {
                    const profEl = document.getElementById('wall-info-profession');
                    const profRow = document.getElementById('wall-info-profession-row');
                    if (msg.profession) {
                        if (profEl) profEl.textContent = msg.profession;
                        if (profRow) profRow.classList.remove('hidden');
                    } else if (profRow) {
                        profRow.classList.add('hidden');
                    }
                }
                if (msg.birth_date !== undefined) {
                    const ageEl = document.getElementById('wall-info-age');
                    const ageRow = document.getElementById('wall-info-age-row');
                    if (msg.birth_date) {
                        const age = calculateAge(new Date(msg.birth_date));
                        if (ageEl) ageEl.textContent = `${age} лет`;
                        if (ageRow) ageRow.classList.remove('hidden');
                    } else if (ageRow) {
                        ageRow.classList.add('hidden');
                    }
                }
                if (msg.rating !== undefined) {
                    const votesEl = document.getElementById('wall-info-votes');
                    const votesRow = document.getElementById('wall-info-votes-row');
                    if (msg.rating !== 0) {
                        if (votesEl) votesEl.textContent = `${msg.rating > 0 ? '+' : ''}${msg.rating} vote`;
                        if (votesRow) votesRow.classList.remove('hidden');
                    } else if (votesRow) {
                        votesRow.classList.add('hidden');
                    }
                }
                if (msg.total_wall_likes !== undefined) {
                    const likesEl = document.getElementById('wall-info-likes');
                    const likesRow = document.getElementById('wall-info-likes-row');
                    const wallLikes = msg.total_wall_likes;
                    if (wallLikes > 0) {
                        if (likesEl) likesEl.textContent = `${wallLikes} лайк${wallLikes === 1 ? '' : wallLikes < 5 ? 'а' : 'ов'}`;
                        if (likesRow) likesRow.classList.remove('hidden');
                    } else if (likesRow) {
                        likesRow.classList.add('hidden');
                    }
                }
                if (msg.avatar_url) {
                    const avatarEl = document.getElementById('wall-avatar');
                    if (avatarEl) {
                        avatarEl.innerHTML = `<img src="${msg.avatar_url}" class="w-full h-full object-cover">`;
                    }
                }
            }
        }
    };

    _wallPostsWS.onclose = () => {
        _wallPostsWS = null;
    };
}

async function publishWallPost() {
    const input = document.getElementById('wall-post-input');
    const content = input.value.trim();

    if (!content && _wallSelectedFiles.length === 0) return;

    try {
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
            await Promise.all(uploadPromises);
        }

        window.app.notify('Опубликовано ✓', 'success');
        input.value = '';
        _wallSelectedFiles = [];
        updateWallMediaPreview();
        // UI self-updates via WS or manual reload if WS fails
        if (!_wallPostsWS) loadWallPosts(_wallUserId);
    } catch (err) {
        console.error('Ошибка публикации:', err);
        window.app.notify('Ошибка при публикации', 'error');
    }
}

async function publishWallMediaDirect(e) {
    const files = Array.from(e.target.files);
    if (!files.length) return;

    try {
        const response = await fetch('/api/wall/posts', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('alpha_token')}`
            },
            body: JSON.stringify({ content: "" })
        });

        if (!response.ok) throw new Error('Ошибка создания поста');
        const post = await response.json();

        for (const file of files) {
            const formData = new FormData();
            formData.append('file', file);
            await fetch(`/api/wall/posts/${post.id}/attachments`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('alpha_token')}`
                },
                body: formData
            });
        }

        window.app.notify('Медиа опубликовано ✓', 'success');
        e.target.value = '';
        if (!_wallPostsWS) loadWallPosts(_wallUserId);
        switchWallTab('media');
    } catch (err) {
        console.error('Ошибка публикации медиа:', err);
        window.app.notify('Ошибка при загрузке', 'error');
    }
}

function handleWallMediaSelect(e) {
    const files = Array.from(e.target.files);
    if (!files.length) return;
    _wallSelectedFiles = [..._wallSelectedFiles, ...files].slice(0, 10);
    updateWallMediaPreview();
    e.target.value = '';
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

function _setWallModeUI(mode) {
    _wallMode = mode;
    const postsBtn       = document.getElementById('wall-tab-posts');
    const mediaBtn       = document.getElementById('wall-tab-media');
    const feed           = document.getElementById('wall-feed');
    const mediaContainer = document.getElementById('wall-media-container');
    const creator        = document.getElementById('wall-post-creator');
    const leftSidebar    = document.getElementById('wall-left-sidebar');

    if (postsBtn) {
        if (mode === 'posts') {
            postsBtn.classList.add('text-custom-main', 'border-custom-accent');
            postsBtn.classList.remove('text-custom-muted', 'border-transparent');
        } else {
            postsBtn.classList.remove('text-custom-main', 'border-custom-accent');
            postsBtn.classList.add('text-custom-muted', 'border-transparent');
        }
    }
    if (mediaBtn) {
        if (mode === 'media') {
            mediaBtn.classList.add('text-custom-main', 'border-custom-accent');
            mediaBtn.classList.remove('text-custom-muted', 'border-transparent');
        } else {
            mediaBtn.classList.remove('text-custom-main', 'border-custom-accent');
            mediaBtn.classList.add('text-custom-muted', 'border-transparent');
        }
    }

    if (leftSidebar) {
        leftSidebar.classList.toggle('wall-sidebar-hidden', mode === 'media');
    }

    const isMe = creator && creator.dataset.visible !== 'false';
    if (creator) creator.classList.toggle('hidden', !(isMe && mode === 'posts'));

    if (feed) feed.classList.toggle('hidden', mode !== 'posts');
    if (mediaContainer) mediaContainer.classList.toggle('hidden', mode !== 'media');
}

function switchWallTab(tab) {
    if (_wallMode === tab) return;
    _setWallModeUI(tab);
}

function renderMediaGrid(media, isMe = false) {
    const grid = document.getElementById('wall-media-grid-main');
    if (!grid) return;

    let addBtnHtml = '';
    if (isMe) {
        addBtnHtml = `
            <input type="file" id="wall-media-direct-upload" class="hidden" accept="image/*,video/*" multiple onchange="publishWallMediaDirect(event)">
            <div class="aspect-square rounded-xl border-2 border-dashed border-white/10 hover:border-custom-accent/40 flex items-center justify-center cursor-pointer transition-all group" onclick="document.getElementById('wall-media-direct-upload').click()">
                <div class="w-12 h-12 rounded-full bg-white/5 flex items-center justify-center group-hover:bg-custom-accent/10 transition-colors">
                    <svg class="w-6 h-6 text-custom-muted group-hover:text-custom-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
                </div>
            </div>
        `;
    }

    if (!media || media.length === 0) {
        grid.innerHTML = addBtnHtml + (isMe ? '' : `<div class="col-span-full py-10 text-center opacity-40 text-xs italic">Медиа-файлов пока нет</div>`);
        return;
    }

    grid.innerHTML = addBtnHtml + media.map(m => {
        const isVideo = (m.mime_type || '').startsWith('video/');
        const deleteBtn = isMe ? `
            <button onclick="event.stopPropagation(); deleteWallMedia('${m.id}')" class="wall-media-delete-btn absolute top-1.5 right-1.5 w-6 h-6 bg-black/60 backdrop-blur-sm text-white rounded-full flex items-center justify-center hover:bg-red-500/80 transition-all z-10" title="Удалить">
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12"/></svg>
            </button>` : '';
        if (isVideo) {
            return `
                <div id="media-item-${m.id}" class="aspect-square rounded-xl overflow-hidden bg-black/20 border border-white/5 relative group cursor-pointer wall-media-item" onclick="Feed.openMedia('${m.post_id}', '${m.url}', true)">
                    <video src="${m.url}" class="w-full h-full object-cover" muted preload="metadata" onloadedmetadata="setVideoPoster(this)"></video>
                    <div class="absolute inset-0 flex items-center justify-center bg-black/30 opacity-0 group-hover:opacity-100 transition-opacity">
                        <svg class="w-8 h-8 text-white drop-shadow" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
                    </div>
                    ${deleteBtn}
                </div>
            `;
        }
        return `
            <div id="media-item-${m.id}" class="aspect-square rounded-xl overflow-hidden bg-black/20 border border-white/5 cursor-pointer hover:scale-[1.02] transition-transform wall-media-item relative group" onclick="Feed.openMedia('${m.post_id}', '${m.url}', false)">
                <img src="${m.url}" class="w-full h-full object-cover" loading="lazy">
                ${deleteBtn}
            </div>
        `;
    }).join('');
}

// Медиа-навигация перенесена в feed.js — используй Feed.openMedia()

async function toggleMediaPostLike(btn) {
    if (!requireAuth()) return;
    const postId = btn.dataset.postId;
    if (!postId) return;
    await togglePostLike(postId, btn);
}

function calculateAge(birthDate) {
    const today = new Date();
    let age = today.getFullYear() - birthDate.getFullYear();
    const m = today.getMonth() - birthDate.getMonth();
    if (m < 0 || (m === 0 && today.getDate() < birthDate.getDate())) age--;
    return age;
}

async function togglePostLike(postId, btn) {
    if (!requireAuth()) return;
    try {
        const res = await fetch(`/api/wall/posts/${postId}/like`, {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${localStorage.getItem('alpha_token')}` }
        });
        if (!res.ok) throw new Error();
        const data = await res.json();

        const liked = data.liked;
        const count = data.likes_count;
        const svg = btn.querySelector('svg');
        const span = btn.querySelector('.like-count');

        btn.dataset.liked = liked;
        if (liked) {
            btn.classList.add('text-red-400');
            btn.classList.remove('text-custom-muted');
            svg.setAttribute('fill', 'currentColor');
            svg.style.transform = 'scale(1.3)';
            setTimeout(() => svg.style.transform = '', 300);
        } else {
            btn.classList.remove('text-red-400');
            btn.classList.add('text-custom-muted');
            svg.setAttribute('fill', 'none');
        }
        if (span) span.textContent = count || '';

        const modalLikeBtn = document.getElementById('wall-comments-media-like');
        if (modalLikeBtn && modalLikeBtn.dataset.postId === postId) {
            modalLikeBtn.dataset.liked = liked;
            const modalSvg = modalLikeBtn.querySelector('svg');
            if (liked) {
                modalLikeBtn.classList.add('text-red-400');
                modalSvg.setAttribute('fill', 'currentColor');
            } else {
                modalLikeBtn.classList.remove('text-red-400');
                modalSvg.setAttribute('fill', 'none');
            }
        }

        const syncIds = [`post-${postId}`, `activity-post-${postId}`];
        syncIds.forEach(id => {
            const card = document.getElementById(id);
            if (card) {
                const feedBtn = card.querySelector('button[onclick^="togglePostLike"]');
                if (feedBtn && feedBtn !== btn) {
                    feedBtn.dataset.liked = liked;
                    const feedSvg = feedBtn.querySelector('svg');
                    const feedSpan = feedBtn.querySelector('.like-count, span'); // В ленте активности span без класса
                    if (liked) {
                        feedBtn.classList.add('text-red-400');
                        feedBtn.classList.remove('text-custom-muted');
                        feedSvg.setAttribute('fill', 'currentColor');
                    } else {
                        feedBtn.classList.remove('text-red-400');
                        feedBtn.classList.add('text-custom-muted');
                        feedSvg.setAttribute('fill', 'none');
                    }
                    if (feedSpan) feedSpan.textContent = count || '';
                }
            }
        });
    } catch {
        window.app.notify('Ошибка', 'error');
    }
}

async function openPostChat(postId, chatId) {
    if (!requireAuth()) return;
    if (!chatId) {
        try {
            const res = await fetch(`/api/wall/posts/${postId}/chat`, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('alpha_token')}` }
            });
            if (!res.ok) throw new Error();
            const data = await res.json();
            chatId = data.chat_id;
        } catch {
            window.app.notify('Не удалось открыть чат', 'error');
            return;
        }
    }
    openWallComments(postId, chatId);
}

async function deleteWallPost(postId) {
    if (!confirm('Удалить эту запись?')) return;
    try {
        const res = await fetch(`/api/wall/posts/${postId}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${localStorage.getItem('alpha_token')}` }
        });
        if (!res.ok) throw new Error('Ошибка удаления');
        const card = document.getElementById(`post-${postId}`) || document.getElementById(`activity-post-${postId}`);
        if (card) {
            card.style.opacity = '0';
            card.style.transform = 'scale(0.97)';
            setTimeout(() => card.remove(), 320);
        }
        window.app.notify('Запись удалена', 'success');
    } catch (err) {
        console.error(err);
        window.app.notify('Ошибка при удалении', 'error');
    }
}

async function deleteWallMedia(attachmentId) {
    if (!confirm('Удалить этот файл?')) return;
    try {
        const res = await fetch(`/api/wall/media/${attachmentId}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${localStorage.getItem('alpha_token')}` }
        });
        if (!res.ok) throw new Error('Ошибка удаления');
        const cell = document.getElementById(`media-item-${attachmentId}`);
        if (cell) {
            cell.style.opacity = '0';
            cell.style.transform = 'scale(0.9)';
            setTimeout(() => cell.remove(), 260);
        }
        await loadWallPosts(_wallUserId);
        window.app.notify('Медиа удалено', 'success');
    } catch (err) {
        console.error(err);
        window.app.notify('Ошибка при удалении', 'error');
    }
}

function renderWallBio(bio, isMe) {
    const bioText    = document.getElementById('wall-bio-text');
    const bioEditBtn = document.getElementById('wall-bio-edit-btn');
    const bioDisplay = document.getElementById('wall-bio-display');

    if (!bioDisplay) return;

    if (bio || isMe) {
        bioDisplay.classList.remove('hidden');
        bioDisplay.classList.add('pb-3', 'border-b', 'border-white/5');
    } else {
        bioDisplay.classList.add('hidden');
        return;
    }

    if (bioText) {
        if (bio) {
            bioText.textContent = bio;
            bioText.classList.remove('hidden');
            if (isMe) {
                bioText.classList.add('cursor-pointer', 'hover:text-custom-main');
                bioText.onclick = startEditBio;
            } else {
                bioText.classList.remove('cursor-pointer', 'hover:text-custom-main');
                bioText.onclick = null;
            }
        } else {
            bioText.classList.add('hidden');
        }
    }

    if (bioEditBtn) bioEditBtn.classList.toggle('hidden', !isMe || !!bio);
}

function startEditBio() {
    const bio = document.getElementById('wall-bio-text')?.textContent || '';
    const display = document.getElementById('wall-bio-display');
    const editor  = document.getElementById('wall-bio-editor');
    const input   = document.getElementById('wall-bio-input');
    if (!editor || !display) return;
    input.value = bio;
    display.classList.add('hidden');
    editor.classList.remove('hidden');
    input.focus();
}

function cancelEditBio() {
    document.getElementById('wall-bio-display').classList.remove('hidden');
    document.getElementById('wall-bio-editor').classList.add('hidden');
}

async function saveWallBio() {
    const input = document.getElementById('wall-bio-input');
    const bio = input.value.trim();
    try {
        const res = await fetch('/api/wall/settings', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('alpha_token')}`
            },
            body: JSON.stringify({ bio })
        });
        if (!res.ok) throw new Error('Ошибка сохранения');
        cancelEditBio();
        renderWallBio(bio, true);
        window.app.notify('Сохранено ✓', 'success');
    } catch (err) {
        console.error(err);
        window.app.notify('Ошибка сохранения', 'error');
    }
}

// ═══════════════════════════════════════════════════════════
// WALL COMMENTS — модальное окно с деревом комментариев + WS
// ═══════════════════════════════════════════════════════════

let _commentsWS = null;
let _commentsChatId = null;

function openWallComments(postId, chatId, keepMediaVisible = false) {
    _commentsChatId = chatId;
    const overlay = document.getElementById('wall-comments-overlay');
    if (!keepMediaVisible) {
        document.getElementById('wall-comments-media-container').classList.add('hidden');
    }
    overlay.classList.remove('hidden');
    requestAnimationFrame(() => overlay.classList.add('comments-open'));
    document.getElementById('wall-comments-input').value = '';
    document.getElementById('wall-comments-input').dataset.parentId = '';
    document.getElementById('wall-comments-reply-to').classList.add('hidden');
    document.getElementById('wall-comments-feed').innerHTML = `
        <div class="flex justify-center py-8 opacity-40">
            <svg class="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z"/>
            </svg>
        </div>`;
    loadComments(chatId);
    connectCommentsWS(chatId);
}

function closeWallComments() {
    const overlay = document.getElementById('wall-comments-overlay');
    overlay.classList.remove('comments-open');
    setTimeout(() => {
        overlay.classList.add('hidden');
        document.getElementById('wall-comments-media-container').classList.add('hidden');
        document.getElementById('wall-comments-media-content').innerHTML = '';
        // Скрываем сайдбар и сбрасываем панель комментариев
        const sidebar = document.getElementById('media-sidebar');
        if (sidebar) { sidebar.classList.add('hidden'); sidebar.style.display = ''; }
        if (typeof Feed !== 'undefined' && Feed.resetComments) Feed.resetComments();
    }, 320);
    if (_commentsWS) { _commentsWS.close(); _commentsWS = null; }
    _commentsChatId = null;
}

async function loadComments(chatId) {
    try {
        const res = await fetch(`/api/wall/chat/${chatId}/comments`, {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('alpha_token')}` }
        });
        const comments = await res.json();
        renderCommentTree(comments);
        // Обновляем счётчик в сайдбаре
        const countEl = document.getElementById('sidebar-comment-count');
        if (countEl && comments?.length) countEl.textContent = comments.length;
    } catch (e) {
        document.getElementById('wall-comments-feed').innerHTML =
            '<p class="text-center text-xs opacity-40 py-8">Ошибка загрузки</p>';
    }
}

function connectCommentsWS(chatId) {
    if (_commentsWS) _commentsWS.close();
    const token = localStorage.getItem('alpha_token') || '';
    const proto = location.protocol === 'https:' ? 'wss' : 'ws';
    _commentsWS = new WebSocket(`${proto}://${location.host}/ws/wall/${chatId}?token=${token}`);
    _commentsWS.onmessage = (e) => {
        const msg = JSON.parse(e.data);
        if (msg.type === 'new_comment') {
            appendComment(msg.comment);
        }
    };
    _commentsWS.onclose = () => { _commentsWS = null; };
}

function renderCommentTree(comments) {
    const feed = document.getElementById('wall-comments-feed');
    if (!comments || comments.length === 0) {
        feed.innerHTML = '<p class="text-center text-xs opacity-30 italic py-8">Первым напишите комментарий</p>';
        return;
    }
    feed.innerHTML = comments.map(c => renderCommentNode(c, 0)).join('');
}

function renderCommentNode(c, depth) {
    const indent = depth > 0 ? `style="margin-left:${Math.min(depth * 16, 64)}px"` : '';
    const hasReplies = c.replies && c.replies.length > 0;
    const repliesId = `replies-${c.id}`;
    const time = new Date(c.created_at).toLocaleTimeString('ru', { hour: '2-digit', minute: '2-digit' });

    return `
    <div class="wall-comment-node" ${indent} data-id="${c.id}">
        <div class="flex gap-2.5 group/comment">
            <div class="w-7 h-7 rounded-full bg-custom-sidebar flex-shrink-0 overflow-hidden border border-white/10 flex items-center justify-center text-xs font-bold">
                ${c.sender_avatar_url ? `<img src="${c.sender_avatar_url}" class="w-full h-full object-cover">` : c.sender_name?.[0]?.toUpperCase() || '?'}
            </div>
            <div class="flex-grow min-w-0">
                <div class="flex items-baseline gap-2 mb-0.5">
                    <span class="text-xs font-semibold text-custom-main">${c.sender_name}</span>
                    <span class="text-[10px] text-custom-muted/50">${time}</span>
                </div>
                <p class="text-xs text-custom-main/80 leading-relaxed break-words">${escapeHtml(c.content)}</p>
                <div class="flex items-center gap-3 mt-1.5">
                    <button onclick="setCommentReply('${c.id}', '${escapeHtml(c.sender_name)}')"
                        class="text-[10px] text-custom-muted/50 hover:text-custom-accent transition">
                        ответить
                    </button>
                    ${hasReplies ? `
                    <button onclick="toggleReplies('${repliesId}', this)"
                        class="text-[10px] text-custom-muted/50 hover:text-custom-accent transition flex items-center gap-1">
                        <svg class="w-3 h-3 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
                        </svg>
                        ${c.replies.length} ${c.replies.length === 1 ? 'ответ' : c.replies.length < 5 ? 'ответа' : 'ответов'}
                    </button>` : ''}
                </div>
            </div>
        </div>
        ${hasReplies ? `
        <div id="${repliesId}" class="hidden mt-2 space-y-2 border-l border-white/5 pl-3">
            ${c.replies.map(r => renderCommentNode(r, depth + 1)).join('')}
        </div>` : ''}
    </div>`;
}

function toggleReplies(id, btn) {
    const el = document.getElementById(id);
    if (!el) return;
    const hidden = el.classList.toggle('hidden');
    const icon = btn.querySelector('svg');
    if (icon) icon.style.transform = hidden ? '' : 'rotate(90deg)';
}

function setCommentReply(parentId, authorName) {
    const input = document.getElementById('wall-comments-input');
    const replyBadge = document.getElementById('wall-comments-reply-to');
    input.dataset.parentId = parentId;
    replyBadge.classList.remove('hidden');
    document.getElementById('wall-comments-reply-name').textContent = authorName;
    input.focus();
}

function clearCommentReply() {
    const input = document.getElementById('wall-comments-input');
    input.dataset.parentId = '';
    document.getElementById('wall-comments-reply-to').classList.add('hidden');
}

async function sendWallComment() {
    if (!requireAuth()) return;
    const input = document.getElementById('wall-comments-input');
    const content = input.value.trim();
    if (!content || !_commentsChatId) return;
    const parentId = input.dataset.parentId || null;
    const body = { content };
    if (parentId) body.parent_id = parentId;
    try {
        const res = await fetch(`/api/wall/chat/${_commentsChatId}/comments`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('alpha_token')}`
            },
            body: JSON.stringify(body)
        });
        if (!res.ok) throw new Error();
        input.value = '';
        clearCommentReply();
    } catch {
        window.app.notify('Ошибка отправки', 'error');
    }
}

function appendComment(comment) {
    const feed = document.getElementById('wall-comments-feed');
    const placeholder = feed.querySelector('p');
    if (placeholder) placeholder.remove();
    if (comment.parent_id) {
        const parentEl = feed.querySelector(`[data-id="${comment.parent_id}"]`);
        if (parentEl) {
            let repliesContainer = parentEl.querySelector('[id^="replies-"]');
            if (!repliesContainer) {
                repliesContainer = document.createElement('div');
                repliesContainer.id = `replies-${comment.parent_id}`;
                repliesContainer.className = 'mt-2 space-y-2 border-l border-white/5 pl-3';
                parentEl.appendChild(repliesContainer);
            }
            repliesContainer.classList.remove('hidden');
            repliesContainer.insertAdjacentHTML('beforeend', renderCommentNode(comment, 1));
            return;
        }
    }
    feed.insertAdjacentHTML('beforeend', renderCommentNode(comment, 0));
    feed.scrollTop = feed.scrollHeight;
}

function escapeHtml(str) {
    return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;');
}


// ── Inline статус на стене ────────────────────────────────────────────────
function startWallStatusEdit() {
    const statusEl = document.getElementById('wall-status');
    const editor   = document.getElementById('wall-status-editor');
    const input    = document.getElementById('wall-status-input');
    if (!statusEl || !editor || !input) return;
    input.value = statusEl.textContent.trim();
    statusEl.classList.add('hidden');
    editor.classList.remove('hidden');
    input.focus();
    input.select();
    input.onkeydown = function(e) {
        if (e.key === 'Enter') saveWallStatus();
        if (e.key === 'Escape') cancelWallStatus();
    };
}

function cancelWallStatus() {
    document.getElementById('wall-status').classList.remove('hidden');
    document.getElementById('wall-status-editor').classList.add('hidden');
}

async function saveWallStatus() {
    const input    = document.getElementById('wall-status-input');
    const statusEl = document.getElementById('wall-status');
    const newStatus = input ? input.value.trim() : '';
    try {
        await apiSaveProfile({ statusText: newStatus });
        if (statusEl) statusEl.textContent = newStatus || '...';
        if (window.app?.currentUser) window.app.currentUser.status = newStatus;
        const stored = JSON.parse(localStorage.getItem('alpha_user') || '{}');
        stored.status = newStatus;
        localStorage.setItem('alpha_user', JSON.stringify(stored));
        window.app?.notify?.('Статус обновлён ✓', 'success');
    } catch (err) {
        window.app?.notify?.('Ошибка: ' + err.message, 'error');
    }
    cancelWallStatus();
}
// ── Post sharing ─────────────────────────────────────────────────────────────

function copyPostLink(postId, btn) {
    const url = `${location.origin}/?post=${postId}`;
    navigator.clipboard.writeText(url).then(() => {
        // Временно меняем иконку на галочку
        const origHTML = btn.innerHTML;
        btn.innerHTML = `<svg class="w-3.5 h-3.5 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"/></svg>`;
        btn.classList.add('text-green-400');
        setTimeout(() => {
            btn.innerHTML = origHTML;
            btn.classList.remove('text-green-400');
        }, 1500);
    }).catch(() => {
        window.app?.notify?.('Не удалось скопировать ссылку', 'error');
    });
}

function shareToTelegram(postId) {
    const url = encodeURIComponent(`${location.origin}/?post=${postId}`);
    window.open(`https://t.me/share/url?url=${url}`, '_blank', 'width=600,height=500,noopener');
}