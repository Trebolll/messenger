// ─── file-upload.js — обработка файлов (скрепка, drag-drop, прогресс) ─────

// Список выбранных файлов перед отправкой
window._pendingFiles = [];

// ── Открыть диалог выбора файла ─────────────────────────────────────────────
function triggerFileInput() {
    document.getElementById('file-input').click();
}

// ── Обработка выбора файла через input ──────────────────────────────────────
document.addEventListener('DOMContentLoaded', () => {
    const fileInput = document.getElementById('file-input');
    if (fileInput) {
        fileInput.addEventListener('change', (e) => {
            const files = Array.from(e.target.files);
            addPendingFiles(files);
            fileInput.value = '';
        });
    }

    // Drag-drop на область чата
    const chatView = document.getElementById('view-chat');
    if (chatView) {
        chatView.addEventListener('dragover', handleDragOver);
        chatView.addEventListener('dragleave', handleDragLeave);
        chatView.addEventListener('drop', handleDrop);
    }
});

// ── Drag & Drop ─────────────────────────────────────────────────────────────
let _dragCounter = 0;

function handleDragOver(e) {
    e.preventDefault();
    e.stopPropagation();
    _dragCounter++;
    if (!window.app?.activeChatId) return;
    const overlay = document.getElementById('drag-drop-overlay');
    if (overlay) overlay.classList.add('visible');
}

function handleDragLeave(e) {
    e.preventDefault();
    e.stopPropagation();
    _dragCounter--;
    if (_dragCounter <= 0) {
        _dragCounter = 0;
        const overlay = document.getElementById('drag-drop-overlay');
        if (overlay) overlay.classList.remove('visible');
    }
}

function handleDrop(e) {
    e.preventDefault();
    e.stopPropagation();
    _dragCounter = 0;
    const overlay = document.getElementById('drag-drop-overlay');
    if (overlay) overlay.classList.remove('visible');

    if (!window.app?.activeChatId) return;

    const files = Array.from(e.dataTransfer.files);
    if (files.length > 0) {
        addPendingFiles(files);
    }
}

// ── Добавить файлы в очередь и показать превью ──────────────────────────────
function addPendingFiles(files) {
    files.forEach(f => {
        window._pendingFiles.push(f);
    });
    renderFilePreviewPanel();
    // Показываем input area если скрыта
    const inputArea = document.getElementById('input-area');
    if (inputArea && inputArea.classList.contains('hidden')) {
        inputArea.classList.remove('hidden');
        setTimeout(() => inputArea.classList.add('visible'), 10);
    }
    // Фокус на textarea
    const textarea = document.getElementById('message-input');
    if (textarea) setTimeout(() => textarea.focus(), 50);
}

// ── Рендер панели превью ─────────────────────────────────────────────────────
function renderFilePreviewPanel() {
    const panel = document.getElementById('file-preview-panel');
    if (!panel) return;

    if (window._pendingFiles.length === 0) {
        panel.classList.remove('visible');
        panel.innerHTML = '';
        return;
    }

    panel.classList.add('visible');
    panel.innerHTML = window._pendingFiles.map((file, idx) => {
        const isImage = file.type.startsWith('image/');
        const sizeFmt = formatFileSize(file.size);

        if (isImage) {
            const url = URL.createObjectURL(file);
            return `
                <div class="file-preview-item">
                    <img src="${url}" class="file-preview-thumb" alt="${escapeHtml(file.name)}" onload="URL.revokeObjectURL(this.src)">
                    <span class="file-preview-name">${escapeHtml(file.name)}</span>
                    <button class="file-preview-remove" onclick="removePendingFile(${idx})" title="Удалить">×</button>
                </div>`;
        } else {
            const ext = file.name.split('.').pop().toUpperCase().slice(0, 4);
            return `
                <div class="file-preview-item">
                    <div class="file-preview-icon">
                        <svg width="28" height="28" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                        </svg>
                        <span style="font-size:9px;font-weight:700;color:var(--text-muted)">${ext}</span>
                    </div>
                    <span class="file-preview-name">${escapeHtml(file.name)}</span>
                    <span class="file-preview-name" style="opacity:0.6">${sizeFmt}</span>
                    <button class="file-preview-remove" onclick="removePendingFile(${idx})" title="Удалить">×</button>
                </div>`;
        }
    }).join('');
}

function removePendingFile(idx) {
    window._pendingFiles.splice(idx, 1);
    renderFilePreviewPanel();
}

// ── Форматирование размера файла ─────────────────────────────────────────────
function formatFileSize(bytes) {
    if (bytes < 1024) return bytes + ' Б';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' КБ';
    return (bytes / (1024 * 1024)).toFixed(1) + ' МБ';
}

// Расчёт оставшегося времени
function formatTimeLeft(bytesLeft, bytesPerSec) {
    if (bytesPerSec <= 0) return '∞';
    const sec = Math.round(bytesLeft / bytesPerSec);
    if (sec < 60) return sec + ' с';
    return Math.ceil(sec / 60) + ' мин';
}

// ── Оптимизация файлов перед загрузкой (сжатие изображений и ограничение размера) ──

async function optimizeFileForUpload(file) {
    // Ограничиваем максимальный размер файла (например, 25 МБ)
    const MAX_FILE_SIZE = 25 * 1024 * 1024;

    // Для изображений пытаемся ужать и уменьшить разрешение
    if (file.type && file.type.startsWith('image/')) {
        try {
            const optimizedImage = await compressImageFile(file, {
                maxWidth: 1920,
                maxHeight: 1080,
                quality: 0.8,
            });
            // Если после сжатия файл всё ещё слишком большой — всё равно вернём его
            return optimizedImage.size <= MAX_FILE_SIZE ? optimizedImage : optimizedImage;
        } catch (e) {
            console.warn('Image optimize failed, sending original file', e);
            return file;
        }
    }

    // Для видео и остальных типов пока только ограничиваем по размеру
    if (file.size > MAX_FILE_SIZE) {
        window.app?.notify('Файл слишком большой для отправки (макс. 25 МБ)', 'error');
        throw new Error('File too large');
    }

    return file;
}

async function compressImageFile(file, { maxWidth, maxHeight, quality }) {
    // Если исходное изображение и так маленькое, нет смысла его перекодировать
    if (!file.type.startsWith('image/')) return file;

    const imageBitmap = await createImageBitmap(file).catch(() => null);
    if (!imageBitmap) return file;

    let { width, height } = imageBitmap;

    const scale = Math.min(
        maxWidth ? maxWidth / width : 1,
        maxHeight ? maxHeight / height : 1,
        1
    );

    if (scale >= 1 && file.size < 2 * 1024 * 1024) {
        // Уже подходящего размера и не слишком большой
        return file;
    }

    const targetWidth = Math.round(width * scale);
    const targetHeight = Math.round(height * scale);

    const canvas = document.createElement('canvas');
    canvas.width = targetWidth;
    canvas.height = targetHeight;

    const ctx = canvas.getContext('2d');
    if (!ctx) return file;

    ctx.drawImage(imageBitmap, 0, 0, targetWidth, targetHeight);

    const mime = file.type === 'image/png' ? 'image/jpeg' : file.type;

    const blob = await new Promise((resolve, reject) => {
        canvas.toBlob(
            b => (b ? resolve(b) : reject(new Error('toBlob failed'))),
            mime,
            quality
        );
    }).catch(() => null);

    if (!blob) return file;

    // Создаём новый File, чтобы backend по-прежнему получал имя и тип
    const optimizedFile = new File([blob], file.name, { type: blob.type || mime });
    return optimizedFile;
}

// ── Отправка сообщения с файлами ─────────────────────────────────────────────
async function handleSendMessage(event) {
    event.preventDefault();

    const input = document.getElementById('message-input');
    const text = input.value.trim();
    const chatId = window.app?.activeChatId;
    if (!chatId) return;

    // Если нет ни текста, ни файлов — ничего не делаем
    if (!text && window._pendingFiles.length === 0) return;

    // Если только текст — стандартная отправка
    if (window._pendingFiles.length === 0) {
        input.value = '';
        animateSendBtn();
        try {
            await apiFetch('/api/messages', {
                method: 'POST',
                body: JSON.stringify({ chat_id: chatId, content: text }),
            });
            // Не добавляем вручную — WebSocket сам доставит сообщение
        } catch (err) {
            window.app.notify('Не удалось отправить сообщение', 'error');
        }
        return;
    }

    // Отправка с файлами
    const filesToSend = [...window._pendingFiles];
    const caption = text;

    // Очищаем UI сразу
    input.value = '';
    window._pendingFiles = [];
    renderFilePreviewPanel();
    animateSendBtn();

    // Отправляем каждый файл
    for (let i = 0; i < filesToSend.length; i++) {
        const file = filesToSend[i];
        const isLast = i === filesToSend.length - 1;
        const msgCaption = isLast ? caption : ''; // текст только к последнему файлу

        await uploadAndSendFile(chatId, file, msgCaption);
    }
}

async function uploadAndSendFile(chatId, file, caption) {
    const tempId = 'upload_' + Date.now() + '_' + Math.random();
    let lastLoaded = 0, lastTime = Date.now();

    // Добавляем временное сообщение с прогрессом
    const tempMsg = {
        id: tempId,
        chat_id: chatId,
        sender_id: window.app.currentUser?.id,
        sender_name: window.app.currentUser?.username,
        content: caption || '',
        created_at: new Date().toISOString(),
        _uploading: true,
        _fileName: file.name,
        _fileSize: file.size,
        _progress: 0,
        _timeLeft: '...',
    };
    window.app.messages.push(tempMsg);
    renderMessages();
    scrollToBottom();

    try {
        // Перед загрузкой оптимизируем файл (сжимаем большие изображения и т.п.)
        const processedFile = await optimizeFileForUpload(file);
        // Шаг 1: отправляем сообщение — получаем message_id
        const msgContent = caption || '';
        const sentMsg = await apiFetch('/api/messages', {
            method: 'POST',
            body: JSON.stringify({ chat_id: chatId, content: msgContent }),
        });
        const messageId = sentMsg.id;

        // Обновляем tempMsg — сохраняем tempId, только обновляем прогресс-данные
        // НЕ заменяем id на реальный — иначе WS не сможет найти дубль и добавит второй раз
        const idxTemp = window.app.messages.findIndex(m => m.id === tempId);
        if (idxTemp !== -1) {
            window.app.messages[idxTemp]._realMsgId = messageId; // запоминаем реальный id
            window.app.messages[idxTemp]._progress = 0;
            window.app.messages[idxTemp]._timeLeft = '...';
        }

        // Шаг 2: загружаем файл с message_id
        const attachment = await apiUploadFile(chatId, processedFile, messageId, (loaded, total) => {
            const now = Date.now();
            const elapsed = (now - lastTime) / 1000;
            const speed = elapsed > 0.2 ? (loaded - lastLoaded) / elapsed : 0;
            if (elapsed > 0.2) { lastLoaded = loaded; lastTime = now; }

            const progress = total > 0 ? Math.round(loaded / total * 100) : 0;
            const timeLeft = speed > 0 ? formatTimeLeft(total - loaded, speed) : '...';

            const idx = window.app.messages.findIndex(m => m.id === tempId);
            if (idx !== -1) {
                window.app.messages[idx]._progress = progress;
                window.app.messages[idx]._timeLeft = timeLeft;
                window.app.messages[idx]._loadedBytes = loaded;
            }
            updateUploadProgress(tempId, progress, loaded, total, timeLeft);
        });

        // Шаг 3: загрузка завершена — убираем tempMsg и показываем реальное сообщение
        const idxNow = window.app.messages.findIndex(m => m.id === tempId);
        const wsMsg = idxNow !== -1 ? window.app.messages[idxNow]._wsMsg : null;

        // Ищем реальное сообщение уже доставленное WS
        const realIdx = window.app.messages.findIndex(m => String(m.id) === String(messageId));
        if (realIdx !== -1) {
            // WS уже доставил — добавляем вложение и убираем tempMsg
            window.app.messages[realIdx]._attachment = attachment;
            window.app.messages[realIdx].content = caption || '';
            window.app.messages = window.app.messages.filter(m => m.id !== tempId);
        } else if (wsMsg) {
            // WS пришёл пока грузили файл — используем его данные
            wsMsg._attachment = attachment;
            wsMsg.content = caption || '';
            if (idxNow !== -1) {
                window.app.messages[idxNow] = wsMsg;
            }
        } else {
            // WS ещё не пришёл — заменяем tempMsg реальным сообщением с вложением
            sentMsg._attachment = attachment;
            sentMsg.content = caption || '';
            if (idxNow !== -1) {
                window.app.messages[idxNow] = sentMsg;
            }
            // Помечаем id как уже добавленный чтобы WS пропустил когда придёт
            if (!window._sentMsgIds) window._sentMsgIds = new Set();
            window._sentMsgIds.add(String(messageId));
        }
        renderMessages();
        scrollToBottom();

    } catch (err) {
        // Убираем временное/реальное сообщение с прогрессом при ошибке
        window.app.messages = window.app.messages.filter(m => m.id !== tempId);
        renderMessages();
        window.app.notify('Не удалось загрузить файл: ' + (err.message || 'Ошибка'), 'error');
    }
}

function updateUploadProgress(tempId, progress, loaded, total, timeLeft) {
    const bar = document.querySelector(`[data-msg-id="${tempId}"] .upload-progress-bar`);
    const info = document.querySelector(`[data-msg-id="${tempId}"] .upload-progress-info`);
    if (bar) bar.style.width = progress + '%';
    if (info) {
        info.innerHTML = `<span>${formatFileSize(loaded)} / ${formatFileSize(total)}</span><span>~${timeLeft}</span>`;
    }
}

function animateSendBtn() {
    const btn = document.getElementById('send-btn');
    if (!btn) return;
    btn.classList.add('sending');
    setTimeout(() => btn.classList.remove('sending'), 500);
}

// ── Лайтбокс для изображений ──────────────────────────────────────────────
function openImgLightbox(src) {
    const lb = document.getElementById('img-lightbox');
    const img = document.getElementById('img-lightbox-img');
    if (!lb || !img) return;
    img.src = src;
    lb.classList.add('visible');
}

function closeImgLightbox() {
    const lb = document.getElementById('img-lightbox');
    if (lb) lb.classList.remove('visible');
}

document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') closeImgLightbox();
});

// ── Парсинг вложений из контента сообщений ──────────────────────────────────

function parseAttachmentFromContent(content) {
    if (!content) return null;
    const match = content.match(/^\[\[ATTACH:(\{.*?\})\]\]/s);
    if (!match) return null;
    try { return JSON.parse(match[1]); }
    catch { return null; }
}

function extractCaptionFromContent(content) {
    if (!content) return '';
    return content.replace(/^\[\[ATTACH:\{.*?\}\]\]\n?/s, '');
}

// Патчим загрузку сообщений — парсим вложения при получении из API
const _origLoadMessages = typeof apiLoadMessages === 'function' ? apiLoadMessages : null;

// Хук: обогащаем сообщения данными вложений после получения
function enrichMessagesWithAttachments(messages) {
    return messages.map(msg => {
        if (msg._attachment) return msg; // уже обогащено
        const att = parseAttachmentFromContent(msg.content);
        if (att) {
            msg._attachment = att;
            msg.content = extractCaptionFromContent(msg.content);
        }
        return msg;
    });
}

// Перехватываем установку app.messages
(function patchAppMessages() {
    const interval = setInterval(() => {
        if (window.app) {
            clearInterval(interval);
            const origLoadMessages = window.app.loadMessages.bind(window.app);
            window.app.loadMessages = async function(chatId) {
                await origLoadMessages(chatId);
                if (Array.isArray(window.app.messages)) {
                    window.app.messages = enrichMessagesWithAttachments(window.app.messages);
                    renderMessages();
                }
            };
        }
    }, 100);
})();