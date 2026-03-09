// Inline стек аватарок для групп (в хедере и info-panel)
function groupAvatarStackInline(members, size) {
  const shown = members.slice(0, 4);
  const offset = Math.round(size * 0.55);
  const total = shown.length * size - (shown.length - 1) * (size - offset);
  let html = `<div style="position:relative;width:${total}px;height:${size}px;flex-shrink:0;">`;
  shown.forEach((m, i) => {
    const letter = (m.username || '?')[0].toUpperCase();
    const img = m.avatar_url
        ? `<img src="${m.avatar_url}" style="width:100%;height:100%;object-fit:cover;border-radius:50%;">`
        : letter;
    html += `<div style="position:absolute;left:${i * offset}px;top:0;width:${size}px;height:${size}px;
            border-radius:50%;border:2px solid var(--bg-main);background:#dbeafe;
            display:flex;align-items:center;justify-content:center;font-weight:700;
            font-size:${Math.round(size * 0.38)}px;color:#2563eb;overflow:hidden;z-index:${shown.length - i};">${img}</div>`;
  });
  html += '</div>';
  return html;
}

// Helpers для отображения имени чата
function chatDisplayName(chat) {
  if (chat.name) return chat.name;
  if (chat.is_group && chat.members && chat.members.length) {
    return chat.members.map(m => m.username).join(', ');
  }
  return chat.other_user_name || chat.partner_name || 'Чат';
}

// Хелперы аватаров — избегаем вложенных backtick в template literals
function chatAvatarHtml(chat) {
  if (chat.avatar_url) {
    return '<img src="' + chat.avatar_url + '" style="width:100%;height:100%;object-fit:cover;">';
  }
  return (chat.name || '?')[0].toUpperCase();
}

function userAvatarHtml(user) {
  if (user && user.avatar_url) {
    return '<img src="' + user.avatar_url + '" style="width:100%;height:100%;object-fit:cover;border-radius:50%;">';
  }
  return (user && user.username ? user.username[0].toUpperCase() : 'U');
}

// ─── render.js — отрисовка интерфейса ─────────────────────────────────────

function formatMessageContent(content) {
  if (!content) return '';
  let escaped = escapeHtml(content);

  // Regex для YouTube (full & short links)
  const ytRegex = /(?:https?:\/\/)?(?:www\.)?(?:youtube\.com\/watch\?v=|youtu\.be\/)([a-zA-Z0-9_-]{11})/g;
  
  // Если есть ссылка на YouTube — добавляем плеер с автоплеем
  return escaped.replace(ytRegex, (match, videoId) => {
    return `<div class="yt-embed-container" style="position:relative;padding-bottom:56.25%;height:0;overflow:hidden;border-radius:12px;margin:10px 0;background:#000;box-shadow:0 4px 12px rgba(0,0,0,0.15);">
              <iframe 
                src="https://www.youtube.com/embed/${videoId}?autoplay=1&mute=1&playsinline=1&rel=0" 
                style="position:absolute;top:0;left:0;width:100%;height:100%;" 
                frameborder="0" 
                allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" 
                allowfullscreen>
              </iframe>
            </div>` + match; // Оставляем саму ссылку под видео
  });
}

function renderChats() {
  const app  = window.app;
  const list = document.getElementById('chats-list');
  list.innerHTML = app.chats.map(chat => {
    const isActive = String(app.activeChatId) === String(chat.id);
    const displayName = chatDisplayName(chat);
    const lastMsg = chat.last_message || 'Нет сообщений';

    // Групповой чат
    if (chat.is_group) {
      const groupLetter = (chat.name || 'G')[0].toUpperCase();
      const groupAvatarInner = chat.avatar_url
          ? `<img src="${chat.avatar_url}" style="width:100%;height:100%;object-fit:cover;">`
          : groupLetter;

      return `<div onclick="app.loadMessages('${chat.id}')" class="chat-list-item p-4 flex items-center gap-3 transition ${isActive ? 'active' : ''}" data-chat-id="${chat.id}">
                <div class="relative flex-shrink-0">
                    <div class="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold overflow-hidden">
                        ${groupAvatarInner}
                    </div>
                </div>
                <div class="flex-grow overflow-hidden">
                    <div class="flex justify-between items-baseline">
                        <h4 class="font-bold text-custom-main truncate">${displayName}</h4>
                        <span class="text-[10px] text-custom-muted">группа</span>
                    </div>
                    <p class="text-xs text-custom-muted truncate">${lastMsg}</p>
                </div>
            </div>`;
    }

    // Приватный чат
    let isOnline = !!chat.is_online;
    if (chat.interlocutor_id && app.userStatusMap && app.userStatusMap[String(chat.interlocutor_id)]) {
        isOnline = app.userStatusMap[String(chat.interlocutor_id)].online;
    }

    const avatarHtml = `<div class="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold overflow-hidden">
            ${chatAvatarHtml(chat)}
        </div>
        ${isOnline ? '<div class="online-dot-glass"></div>' : ''}`;

    return `<div onclick="app.loadMessages('${chat.id}')" class="chat-list-item p-4 flex items-center gap-3 transition ${isActive ? 'active' : ''}" data-chat-id="${chat.id}">
            <div class="relative flex-shrink-0">${avatarHtml}</div>
            <div class="flex-grow overflow-hidden">
                <div class="flex justify-between items-baseline">
                    <h4 class="font-bold text-custom-main truncate">${displayName}</h4>
                    <span class="text-[10px] text-custom-muted">12:45</span>
                </div>
                <p class="text-xs text-custom-muted truncate">${lastMsg}</p>
            </div>
        </div>`;
  }).join('');

  // Восстанавливаем подсветку непрочитанных после перерисовки
  const unread = window.app._unreadHighlight;
  if (unread && unread.size) {
    unread.forEach(chatId => _applyUnreadHighlight(chatId));
  }
}

function renderMessages() {
  const app       = window.app;
  const container = document.getElementById('messages-container');

  // Запоминаем какие id уже отрисованы чтобы анимировать только новые
  const existing = new Set(
      [...container.querySelectorAll('[data-msg-id]')].map(el => el.dataset.msgId)
  );

// Строим карту онлайн-статусов из чатов
  const onlineMap = {};
  (app.chats || []).forEach(chat => {
    if (chat.interlocutor_id) {
      onlineMap[String(chat.interlocutor_id)] = !!chat.is_online;
    }
    if (chat.members) {
      chat.members.forEach(m => { onlineMap[String(m.id)] = !!m.is_online; });
    }
  });

  container.innerHTML = app.messages.map((msg, idx) => {
    const isMe     = String(msg.sender_id) === String(app.currentUser?.id);
    const isRead   = msg.read_at   != null;
    const isEdited = msg.edited_at != null;
    const isNew    = !existing.has(String(msg.id));

    const delay = (existing.size === 0 && isNew)
        ? `animation-delay:${Math.min(idx * 35, 400)}ms`
        : '';

    const animClass = isNew ? (isMe ? 'msg-anim-sent' : 'msg-anim-received') : '';

    // Аватар отправителя (только для входящих)
    let senderAvatarHtml = '';
    if (!isMe) {
      const isOnline = !!onlineMap[String(msg.sender_id)];
      const avatarInner = msg.sender_avatar_url
          ? `<img src="${msg.sender_avatar_url}" style="width:100%;height:100%;object-fit:cover;border-radius:50%;">`
          : `<span>${(msg.sender_name || '?')[0].toUpperCase()}</span>`;

      const senderData = JSON.stringify({ id: msg.sender_id, username: msg.sender_name, avatar_url: msg.sender_avatar_url || '' }).replace(/"/g, '&quot;');
      senderAvatarHtml = `
                <div style="position:relative;flex-shrink:0;align-self:flex-end;cursor:pointer;"
                     onclick="openMemberAvatarViewer(JSON.parse(this.dataset.member))"
                     data-member="${senderData}"
                     title="${escapeHtml(msg.sender_name || '')}">
                    <div style="width:30px;height:30px;border-radius:50%;background:#dbeafe;display:flex;align-items:center;justify-content:center;font-weight:700;font-size:12px;color:#2563eb;overflow:hidden;">
                        ${avatarInner}
                    </div>
                </div>`;
    }

    // Никнейм над сообщением (только для входящих в группе)
    const activeChat = app.chats.find(c => String(c.id) === String(app.activeChatId));
    const isGroup = activeChat && activeChat.is_group;
    const nicknameHtml = (!isMe && isGroup && msg.sender_name)
        ? `<div style="font-size:11px;font-weight:600;color:var(--text-muted,#6b7280);margin-bottom:2px;padding-left:2px;cursor:pointer;" onclick="openWall('${msg.sender_id}')">${escapeHtml(msg.sender_name)}</div>`
        : '';

    // Если сообщение загружается — показываем прогресс
    if (msg._uploading) {
      const fileName = escapeHtml(msg._fileName || 'Файл');
      const fileSize = (typeof formatFileSize === 'function') ? formatFileSize(msg._fileSize || 0) : '';
      return `
            <div class="flex items-end gap-2 ${animClass}"
                 style="justify-content:flex-end;${delay}"
                 data-msg-id="${msg.id}"
                 data-sender-id="${msg.sender_id}">
                <div style="display:flex;flex-direction:column;align-items:flex-end;max-width:75%;">
                    <div class="message-bubble p-3.5 message-sent" style="width:fit-content;max-width:100%;">
                        <div class="upload-progress-msg">
                            <div style="display:flex;align-items:center;gap:8px;margin-bottom:4px;">
                                <svg width="18" height="18" fill="none" stroke="rgba(255,255,255,0.85)" stroke-width="1.8" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M21.44 11.05l-9.19 9.19a6 6 0 01-8.49-8.49l9.19-9.19a4 4 0 015.66 5.66l-9.2 9.19a2 2 0 01-2.83-2.83l8.49-8.48"/></svg>
                                <span style="font-size:12px;font-weight:600;opacity:0.9;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:180px;">${fileName}</span>
                            </div>
                            <div class="upload-progress-bar-wrap">
                                <div class="upload-progress-bar" style="width:${msg._progress || 0}%"></div>
                            </div>
                            <div class="upload-progress-info">
                                <span>${(typeof formatFileSize === 'function') ? formatFileSize(msg._loadedBytes || 0) + ' / ' + fileSize : ''}</span>
                                <span>~${msg._timeLeft || '...'}</span>
                            </div>
                            ${msg.content ? `<p style="font-size:13px;margin-top:4px;">${escapeHtml(msg.content)}</p>` : ''}
                        </div>
                    </div>
                </div>
            </div>`;
    }

    // Рендер вложения (если есть)
    let attachmentHtml = '';
    let isMediaAttachment = false;
    if (msg._attachment || msg.attachment) {
      const att = msg._attachment || msg.attachment;
      const mime = att.mime_type || '';
      const attUrl = att.url || '';
      const attName = escapeHtml(att.filename || 'Файл');
      const attSize = (typeof formatFileSize === 'function') ? formatFileSize(att.size_bytes || 0) : '';

      const isGif   = mime === 'image/gif';
      const isImage = !isGif && mime.startsWith('image/');
      const isVideo = mime.startsWith('video/');
      const isAudio = mime.startsWith('audio/');

      if (isGif) {
        isMediaAttachment = true;
        attachmentHtml = `<img src="${attUrl}" class="msg-attachment-gif" alt="${attName}" loading="lazy">`;
      } else if (isImage) {
        isMediaAttachment = true;
        attachmentHtml = `<img src="${attUrl}" class="msg-attachment-image" alt="${attName}" onclick="openImgLightbox('${attUrl}')" loading="lazy">`;
      } else if (isVideo) {
        isMediaAttachment = true;
        attachmentHtml = `<video class="msg-attachment-video" controls preload="metadata">
                    <source src="${attUrl}" type="${mime}">
                </video>`;
      } else if (isAudio) {
        attachmentHtml = `<audio controls style="width:100%;min-width:220px;margin-bottom:4px;border-radius:8px;outline:none;">
                    <source src="${attUrl}" type="${mime}">
                </audio>`;
      } else {
        attachmentHtml = `<a href="${attUrl}" class="msg-attachment-file" target="_blank" download>
                    <svg width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>
                    <div class="msg-attachment-file-info">
                        <span class="msg-attachment-file-name">${attName}</span>
                        <span class="msg-attachment-file-size">${attSize}</span>
                    </div>
                    <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" style="opacity:0.6;flex-shrink:0;"><path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/></svg>
                </a>`;
      }
    }

    // Для медиа — пузырь без горизонтальных паддингов чтобы изображение заполняло его целиком
    const bubblePadding = isMediaAttachment ? 'p-1.5' : 'p-3.5';
    // Контент (подпись или текст) под медиа — с паддингом
    let captionWrap;
    if (isMediaAttachment) {
      if ((msg.content || '').trim()) {
        captionWrap = '<p class="text-sm leading-relaxed" id="msg-content-' + msg.id + '" style="padding:2px 8px 4px;">' + formatMessageContent(msg.content) + '</p>';
      } else {
        captionWrap = '<span id="msg-content-' + msg.id + '" style="display:none;"></span>';
      }
    } else if (msg.content || !attachmentHtml) {
      captionWrap = '<p class="text-sm leading-relaxed" id="msg-content-' + msg.id + '" style="white-space:pre-wrap;word-break:break-word;overflow-wrap:anywhere;">' + formatMessageContent(msg.content) + '</p>';
    } else {
      captionWrap = '<span id="msg-content-' + msg.id + '" style="display:none;"></span>';
    }

    return `
            <div class="flex items-end gap-2 ${animClass}"
                 style="justify-content:${isMe ? 'flex-end' : 'flex-start'};${delay}"
                 data-msg-id="${msg.id}"
                 data-sender-id="${msg.sender_id}"
                 oncontextmenu="showMessageMenu(event, '${msg.id}', ${isMe})">
                ${!isMe ? senderAvatarHtml : ''}
                <div style="display:flex;flex-direction:column;align-items:${isMe ? 'flex-end' : 'flex-start'};max-width:75%;">
                    <div class="msg-header-line" style="display:flex;align-items:center;gap:6px;">
                        ${nicknameHtml}
                        ${renderRatingBadge(msg.sender_rating)}
                    </div>
                    <!-- Обертка для пузыря и кнопок -->
                    <div style="position:relative; width:fit-content; max-width:100%;">
                        <div class="message-bubble ${bubblePadding} ${isMe ? 'message-sent' : 'message-received'}" 
                             style="position:relative; z-index:10; width:fit-content; max-width:100%; ${isMediaAttachment ? 'overflow:hidden;' : ''}">
                            ${attachmentHtml}
                            ${captionWrap}
                            <div style="display:flex;align-items:center;justify-content:flex-end;gap:4px;margin-top:2px;flex-wrap:nowrap;${isMediaAttachment ? 'padding:0 6px 4px;' : ''}">
                                ${isEdited ? `<span class="msg-edited-label">изменено</span>` : ''}
                                <span style="font-size:10px;white-space:nowrap;flex-shrink:0;opacity:${isMe ? '0.7' : '1'};" class="${isMe ? '' : 'text-custom-muted'}">
                                    ${new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                </span>
                                ${isMe ? `
                                    <span style="font-size:11px;letter-spacing:-0.5em;display:inline-block;flex-shrink:0;opacity:${isRead ? '0.9' : '0.4'};">
                                        ${isRead ? '✓✓' : '✓'}
                                    </span>
                                ` : ''}
                            </div>
                        </div>

                        ${!isMe ? `
                        <!-- Контейнер для Gooey-эффекта (только для чужих сообщений) -->
                        <div class="gooey-vote-container">
                            <div class="message-bubble message-received" 
                                 style="position:absolute; inset:0; z-index:-1; margin:0; opacity: 1;"></div>
                            
                            <div class="msg-votes ${msg.my_vote !== 0 ? 'has-active' : ''}">
                                ${renderVotesHtml(msg.likes || 0, msg.dislikes || 0, msg.my_vote || 0, msg.id)}
                            </div>
                        </div>
                        ` : ''}
                    </div>
                </div>
            </div>
        `;
  }).join('');
}

// ─── Рейтинговый бейдж ────────────────────────────────────────────────────────
function renderRatingBadge(rating) {
  if (rating === undefined || rating === null) return '';
  const rVal = Number(rating);
  if (isNaN(rVal) || rVal < 0) return '';

  const ranks = [
    { min: 1000000000, name: 'SINGULARITY',   color: '#ffd700',    border: '2px solid #ffd700', bg: 'rgba(255, 215, 0, 0.1)' },
    { min: 100000000,  name: 'GALACTIC',      color: '#ffb347',    border: '2px solid #ffb347', bg: 'rgba(255, 179, 71, 0.1)' },
    { min: 10000000,   name: 'STELLAR',       color: '#ff8c00',    border: '2px solid #ff8c00', bg: 'rgba(255, 140, 0, 0.1)' },
    { min: 1000000,    name: 'MYTHIC',        color: '#ff6600',    border: '2px solid #ff6600', bg: 'rgba(255, 102, 0, 0.1)' },
    { min: 100000,     name: 'GODLIKE',       color: '#ff4444',    border: '2px solid #ff4444', bg: 'rgba(255, 68, 68, 0.1)' },
    { min: 50000,      name: 'IMMORTAL',      color: '#ff6b6b',    border: '2px solid #ff6b6b', bg: 'rgba(255, 107, 107, 0.1)' },
    { min: 1000,       name: 'LEGEND',        color: '#f59e0b',    border: '2px solid #f59e0b', bg: 'rgba(245, 158, 11, 0.1)' },
    { min: 500,        name: 'ELITE',         color: '#8b5cf6',    border: '2px solid #8b5cf6', bg: 'rgba(139, 92, 246, 0.1)' },
    { min: 200,        name: 'EXPERT',        color: '#22c55e',    border: '2px solid #22c55e', bg: 'rgba(34, 197, 94, 0.1)' },
    { min: 50,         name: 'SKILLED',       color: '#3b82f6',    border: '2px solid #3b82f6', bg: 'rgba(59, 130, 246, 0.1)' },
    { min: 0,          name: 'BEGINNER',      color: '#9ca3af',    border: '1px solid #9ca3af', bg: 'transparent' }
  ];
  const rank = ranks.find(r => rVal >= r.min) || ranks[ranks.length - 1];
  return `<span class="msg-rating-badge" title="${rank.name}"
        style="display:inline-flex;align-items:center;gap:2px;font-size:10px;font-weight:600;
               color:${rank.color};background:${rank.color}18;
               padding:1px 5px;border-radius:3px;line-height:1.4;">
        ★ ${rank.name}
    </span>`;
}

// ─── HTML кнопок голосования ──────────────────────────────────────────────────
function renderVotesHtml(likes, dislikes, myVote, messageId) {
  const likeActive    = myVote === 1;
  const dislikeActive = myVote === -1;
  return `
        <button class="vote-btn vote-like ${likeActive ? 'vote-active-like' : ''}"
                data-vote-msg="${messageId}" data-vote="1"
                onclick="voteMessage(event, '${messageId}', 1)"
                title="Лайк">
            <span class="vote-icon">+</span>
            <span class="vote-ring"></span>
        </button>
        <button class="vote-btn vote-dislike ${dislikeActive ? 'vote-active-dislike' : ''}"
                data-vote-msg="${messageId}" data-vote="-1"
                onclick="voteMessage(event, '${messageId}', -1)"
                title="Дизлайк">
            <span class="vote-icon">−</span>
            <span class="vote-ring"></span>
        </button>
    `;
}

// ─── Обработчик клика на голосование ─────────────────────────────────────────
async function voteMessage(e, messageId, vote) {
  e.stopPropagation();
  const btn = e.currentTarget;
  const isLike = vote === 1;

  // Находим пузырь сообщения
  const msgWrap = btn.closest('[data-msg-id]');
  const bubble  = msgWrap ? msgWrap.querySelector('.message-bubble') : null;
  const votesPanel = msgWrap ? msgWrap.querySelector('.msg-votes') : null;

  // Проверка на своё сообщение
  if (bubble && bubble.classList.contains('message-sent')) {
    window.app && window.app.notify('Нельзя голосовать за своё сообщение', 'error');
    return;
  }

  // Анимация кнопки — вспышка + пульс кольца
  btn.classList.remove('vote-flash-like', 'vote-flash-dislike');
  void btn.offsetWidth; // reflow
  btn.classList.add(isLike ? 'vote-flash-like' : 'vote-flash-dislike');

  // Через 180ms — свечение пузыря
  setTimeout(() => {
    if (bubble) {
      bubble.classList.remove('bubble-glow-like', 'bubble-glow-dislike');
      void bubble.offsetWidth;
      bubble.classList.add(isLike ? 'bubble-glow-like' : 'bubble-glow-dislike');
      setTimeout(() => bubble.classList.remove('bubble-glow-like', 'bubble-glow-dislike'), 1000);
    }
  }, 180);

  // Скрываем панель через 1.5 секунды после голоса
  if (votesPanel) {
    setTimeout(() => {
      votesPanel.classList.add('msg-votes-hidden');
      // Возвращаем возможность показа при следующем ховере через некоторое время
      setTimeout(() => votesPanel.classList.remove('msg-votes-hidden'), 3000);
    }, 1500);
  }

  try {
    await apiVoteMessage(messageId, vote);
  } catch(err) {
    // убираем анимацию если ошибка
    btn.classList.remove('vote-flash-like', 'vote-flash-dislike');
    if (votesPanel) votesPanel.classList.remove('msg-votes-hidden');
  }
}

function renderChatHeader() {
  const app  = window.app;
  const chat = app.chats.find(c => String(c.id) === String(app.activeChatId));
  if (!chat) return;

  const isGroup    = !!chat.is_group;
  const isCreator  = isGroup && chat.creator_id && String(chat.creator_id) === String(app.currentUser?.id);
  const displayName = chatDisplayName(chat);

  // ── Плавная смена имени и аватара ──────────────────────────────────
  const nameEl   = document.getElementById('active-chat-name');
  const avatarEl = document.getElementById('active-chat-avatar');
  if (nameEl)   { nameEl.style.opacity   = '0'; }
  if (avatarEl) { avatarEl.style.opacity = '0'; avatarEl.style.transform = 'scale(0.85)'; }
  setTimeout(() => {
    if (nameEl)   { nameEl.style.opacity   = '1'; }
    if (avatarEl) { avatarEl.style.opacity = '1'; avatarEl.style.transform = 'scale(1)'; }
  }, 160);

  // ── Имя в хедере чата ──────────────────────────────────────────────
  document.getElementById('active-chat-name').textContent = displayName;

  // info-name: для создателя — кликабелен для inline-редактирования
  const infoNameEl = document.getElementById('info-name');
  infoNameEl.textContent = displayName;
  if (isCreator) {
    infoNameEl.style.cursor = 'pointer';
    infoNameEl.title = 'Нажмите для редактирования';
    infoNameEl.classList.add('group-edit-name');
    infoNameEl.onclick = () => startInlineGroupNameEdit(chat.id, chat.name || '');
  } else {
    infoNameEl.style.cursor = '';
    infoNameEl.title = '';
    infoNameEl.classList.remove('group-edit-name');
    infoNameEl.onclick = null;
  }

  const headerStatus = document.getElementById('active-chat-status');
  if (isGroup) {
    if (headerStatus) {
      const total = (chat.members || []).length;
      const online = (chat.members || []).filter(m => {
          if (String(m.id) === String(app.currentUser?.id)) return true;
          // Используем статус из глобального мапа, если он там есть
          const globalStatus = app.userStatusMap && app.userStatusMap[String(m.id)];
          if (globalStatus) return globalStatus.online;
          return m.is_online;
      }).length;
      headerStatus.textContent = `${online} из ${total} online`;
      headerStatus.className   = 'text-xs text-custom-muted';
    }
    const badge = document.getElementById('info-status-badge');
    if (badge) badge.style.display = 'none';
    const statusEl = document.getElementById('info-user-status');
    if (statusEl) statusEl.textContent = '';
  } else {
    let isOnline = !!chat.is_online;
    let userStatus = chat.user_status || '';
    if (chat.interlocutor_id && app.userStatusMap && app.userStatusMap[String(chat.interlocutor_id)]) {
        isOnline = app.userStatusMap[String(chat.interlocutor_id)].online;
        userStatus = app.userStatusMap[String(chat.interlocutor_id)].status || '';
    }
    renderStatusElements(isOnline, userStatus);
    const badge = document.getElementById('info-status-badge');
    if (badge) badge.style.display = '';
  }

  // ── Аватар в хедере и info-panel ──────────────────────────────────
  const _hdr = document.getElementById('active-chat-avatar');
  const _inf = document.getElementById('info-avatar');

  if (isGroup && chat.members && chat.members.length > 1) {
    if (chat.avatar_url) {
      const groupImgHtml = `<img src="${chat.avatar_url}" style="width:100%;height:100%;object-fit:cover;">`;
      _hdr.innerHTML = groupImgHtml; _hdr.style.overflow = 'hidden';
      _inf.innerHTML = groupImgHtml; _inf.style.overflow = 'hidden';
    } else {
      const stackHtml = groupAvatarStackInline(chat.members, 28);
      _hdr.innerHTML = stackHtml; _hdr.style.overflow = 'visible';
      _inf.innerHTML = groupAvatarStackInline(chat.members, 36);
      _inf.style.overflow = 'visible';
    }
  } else {
    const _ava = chatAvatarHtml(chat);
    _hdr.innerHTML = _ava; _hdr.style.overflow = 'hidden';
    _inf.innerHTML = _ava; _inf.style.overflow = 'hidden';
  }

  if (isGroup && isCreator) {
    _inf.style.cursor = 'pointer';
    _inf.title = 'Сменить аватар группы';
    _inf.onclick = () => triggerGroupAvatarUpload();
  } else {
    _inf.style.cursor = '';
    _inf.title = '';
    _inf.onclick = null;
  }

  // ── Блок участников в info-panel ──────────────────────────────────
  let membersBlock = document.getElementById('info-members-block');
  if (!membersBlock) {
    membersBlock = document.createElement('div');
    membersBlock.id = 'info-members-block';
    membersBlock.className = 'w-full text-left px-4 pb-4';
    const statusP = document.getElementById('info-user-status');
    if (statusP && statusP.parentNode) {
      statusP.parentNode.insertBefore(membersBlock, statusP.nextSibling);
    }
  }
  membersBlock.innerHTML = '';

  if (!isGroup) return;

  apiGetGroupMembers(chat.id).then(members => {
    if (!members || !members.length) return;

    const localChat = app.chats.find(c => String(c.id) === String(chat.id));
    if (localChat) localChat.members = members;

    // После загрузки участников — обновляем счетчик онлайн в хедере
    const headerStatus = document.getElementById('active-chat-status');
    if (headerStatus) {
        const total = members.length;
        const online = members.filter(m => {
            if (String(m.id) === String(app.currentUser?.id)) return true;
            const globalStatus = app.userStatusMap && app.userStatusMap[String(m.id)];
            if (globalStatus) return globalStatus.online;
            return m.is_online;
        }).length;
        headerStatus.textContent = `${online} из ${total} online`;
    }

    const membersList = members.map(m => {
      const isMe = String(m.id) === String(app.currentUser?.id);
      
      // Статус из глобального мапа (если есть) имеет приоритет над тем что пришло от API
      let isOnline = m.is_online;
      const globalStatus = app.userStatusMap && app.userStatusMap[String(m.id)];
      if (globalStatus) isOnline = globalStatus.online;
      if (isMe) isOnline = true;

      const avatarInner = m.avatar_url
          ? `<img src="${m.avatar_url}" style="width:100%;height:100%;object-fit:cover;">`
          : (m.username || '?')[0].toUpperCase();
      const onlineDot = isOnline ? `<span class="member-online-dot"></span>` : '';
      const removeBtn = (isCreator && !isMe)
          ? `<button onclick="removeGroupMember('${chat.id}','${m.id}')"
                        title="Удалить из группы"
                        class="ml-auto text-red-400 hover:text-red-600 p-1 rounded-lg hover:bg-red-50 dark:hover:bg-red-950/20 transition flex-shrink-0 member-remove-btn">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                        </svg>
                    </button>`
          : '';
      return `
                <div class="member-row flex items-center gap-3 py-1.5" data-uid="${m.id}">
                    <div class="relative flex-shrink-0">
                        <div onclick="openMemberAvatarViewer(JSON.parse(this.parentNode.parentNode.dataset.member))"
                            class="w-9 h-9 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold overflow-hidden cursor-pointer hover:ring-2 hover:ring-blue-400 hover:ring-offset-1 transition text-sm">
                            ${avatarInner}
                        </div>
                        ${onlineDot}
                    </div>
                    <div class="overflow-hidden flex-1 min-w-0">
                        <div class="text-sm font-semibold text-custom-main truncate">
                            ${m.username}${isMe ? ' <span class="text-xs text-custom-muted font-normal">(вы)</span>' : ''}
                        </div>
                        ${m.full_name ? `<div class="text-xs text-custom-muted truncate">${m.full_name}</div>` : ''}
                        ${m.status ? `<div class="text-xs text-custom-muted truncate" style="font-style:italic;opacity:0.8;">${m.status}</div>` : ''}
                    </div>
                    ${removeBtn}
                </div>`.replace('this.parentNode.parentNode.dataset.member', `'${JSON.stringify(m).replace(/'/g, "\\'").replace(/"/g, '&quot;')}'`);
    }).join('');

    // Кнопка + рядом с заголовком — только для создателя
    const addBtnHtml = isCreator
        ? `<button id="add-member-plus-btn" onclick="handleAddMemberBtn()" title="Добавить участника"
                    class="add-member-plus-btn w-7 h-7 flex items-center justify-center rounded-full relative">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                    </svg>
                </button>`
        : `<div class="w-7 h-7"></div>`;  // заглушка для выравнивания

    membersBlock.innerHTML = `
            <div class="flex items-center justify-between mb-2 mt-1">
                <div class="text-xs font-semibold text-custom-muted uppercase tracking-wider">
                    Участники (${members.length})
                </div>
                ${addBtnHtml}
            </div>
            <div id="group-members-list" class="space-y-0.5">
                ${membersList}
            </div>`;
  }).catch(() => {});
}


function renderStatusElements(isOnline, userStatus) {
  // Бейдж онлайн в правой панели
  const badge = document.getElementById('info-status-badge');
  const text  = document.getElementById('info-status-text');
  if (badge) {
    if (isOnline) {
      badge.style.display = 'inline-flex';
      badge.className = 'info-status-badge online mb-2';
    } else {
      badge.style.display = 'none';
    }
  }
  if (text) text.textContent = isOnline ? 'online' : '';

  // Текстовый статус — всегда показываем, пустой если нет
  const statusEl = document.getElementById('info-user-status');
  if (statusEl) statusEl.textContent = userStatus ? `«${userStatus}»` : '';

  // Статус под именем в хедере чата
  const headerStatus = document.getElementById('active-chat-status');
  if (headerStatus) {
    headerStatus.textContent = isOnline ? 'online' : '';
    headerStatus.className   = 'text-xs text-green-500';
  }
}

function renderSearchResults(users) {
  const container = document.getElementById('search-results');
  const selected = window._ncSelected || [];

  // Храним пользователей в глобальном Map по id — безопасная передача в onclick
  if (!window._ncUserMap) window._ncUserMap = {};
  users.forEach(u => { window._ncUserMap[String(u.id)] = u; });

  // В режиме добавления — исключаем тех, кто уже в группе
  const excludeIds = window._addMemberMode ? (window._ncExcludeIds || new Set()) : new Set();
  const filtered = users.filter(u => !excludeIds.has(String(u.id)));

  if (!filtered.length) {
    const msg = (window._addMemberMode && users.length)
        ? 'Все найденные пользователи уже в группе'
        : 'Пользователи не найдены';
    container.innerHTML = `<p class="text-xs text-custom-muted text-center py-3 opacity-60">${msg}</p>`;
    return;
  }
  container.innerHTML = filtered.map(user => {
    const isSelected = selected.some(u => String(u.id) === String(user.id));
    const initial = (user.username || 'U')[0].toUpperCase();
    const checkSvg = isSelected
        ? `<svg class="w-3 h-3" fill="none" stroke="white" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"/></svg>`
        : '';
    return `<div onclick="ncToggleUser('${user.id}')"
             class="nc-result-item ${isSelected ? 'selected' : ''}"
             data-uid="${user.id}">
            <div class="w-8 h-8 rounded-full bg-blue-500/20 flex items-center justify-center text-blue-500 font-bold text-sm flex-shrink-0">${initial}</div>
            <div class="flex-1 min-w-0">
                <div class="font-semibold text-custom-main text-sm truncate">${user.username}</div>
                <div class="text-xs text-custom-muted truncate">${user.email}</div>
            </div>
            <div class="flex items-center gap-2">
                <div class="nc-check">${checkSvg}</div>
            </div>
        </div>`;
  }).join('');
}

function loadUserData() {
  const user = window.app?.currentUser || JSON.parse(localStorage.getItem('alpha_user') || 'null');
  if (!user) return;
  document.getElementById('current-user-name').textContent = user.username;
  setAvatarEl(document.getElementById('current-user-avatar'), user);
}

// Универсальная функция: ставит фото или букву в элемент-аватар
function setAvatarEl(el, user) {
  if (!el) return;
  if (user.avatar_url) {
    el.innerHTML = `<img src="${user.avatar_url}" alt="${user.username}"
            style="width:100%;height:100%;object-fit:cover;border-radius:50%;">`;
  } else {
    el.textContent = (user.username || 'U')[0].toUpperCase();
  }
}

function updateLastMessageInChatList(msg) {
  const app  = window.app;
  const chat = app.chats.find(c => String(c.id) === String(msg.chat_id));
  if (chat) {
    chat.last_message = msg.content;
    renderChats();
  } else {
    apiLoadChats();
  }
}

// ─── Helpers ───────────────────────────────────────────────────────────────

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

function scrollToBottom(instant) {
  const container = document.getElementById('messages-container');
  if (!container) return;
  if (instant) {
    // Скрываем контейнер, скроллим в самый низ, потом показываем — без вспышки верха
    container.style.visibility = 'hidden';
    container.scrollTop = container.scrollHeight;
    requestAnimationFrame(function() {
      container.scrollTop = container.scrollHeight;
      container.style.visibility = '';
    });
  } else {
    setTimeout(function() { container.scrollTop = container.scrollHeight; }, 50);
  }
}