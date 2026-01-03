class AlphaApp {
    constructor() {
        this.token = localStorage.getItem('alpha_token');
        this.currentUser = JSON.parse(localStorage.getItem('alpha_user') || 'null');
        this.activeChatId = null;
        this.socket = null;
        this.chats = [];
        this.messages = [];
        
        this.init();
    }

    init() {
        if (this.token) {
            // Если пользователь загружен без ID, попробуем восстановить его из токена или перелогиниться
            if (!this.currentUser || !this.currentUser.id) {
                console.warn('Current user has no ID, clearing local storage');
                this.logout();
                return;
            }
            this.showChat();
        } else {
            this.showLanding();
        }
        
        // Обработка Enter в textarea
        document.addEventListener('keydown', (e) => {
            if (e.target.id === 'message-input' && e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });
    }

    // --- Routing ---
    showLanding() {
        document.getElementById('landing-page').classList.remove('hidden');
        document.getElementById('main-chat').classList.add('hidden');
    }

    showChat() {
        document.getElementById('landing-page').classList.add('hidden');
        document.getElementById('main-chat').classList.remove('hidden');
        this.loadUserData();
        this.loadChats();
        this.connectWebSocket();
    }

    // --- Auth ---
    async auth(event, type) {
        event.preventDefault();
        const form = event.target;
        const formData = new FormData(form);
        const data = Object.fromEntries(formData.entries());
        const errorEl = document.getElementById(`${type}-error`);
        
        try {
            const response = await fetch(`/api/${type}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });

            const result = await response.json();
            if (!response.ok) throw new Error(result.error || 'Ошибка авторизации');

            this.token = result.token;
            this.currentUser = result.user || { email: data.email, username: data.username || data.email.split('@')[0] };
            localStorage.setItem('alpha_token', this.token);
            localStorage.setItem('alpha_user', JSON.stringify(this.currentUser));
            
            this.notify('Успешно!', 'success');
            closeAuthModal();
            this.showChat();
        } catch (err) {
            errorEl.textContent = err.message;
            errorEl.classList.remove('hidden');
        }
    }

    logout() {
        localStorage.clear();
        this.token = null;
        this.currentUser = null;
        if (this.socket) this.socket.close();
        window.location.reload();
    }

    // --- API Calls ---
    async loadChats() {
        try {
            const res = await this.apiFetch('/api/chats');
            this.chats = res || [];
            this.renderChats();
        } catch (err) {
            this.notify('Ошибка загрузки чатов', 'error');
        }
    }

    async loadMessages(chatId) {
        this.activeChatId = chatId;
        document.getElementById('no-chat-selected').classList.add('hidden');
        this.renderChatHeader();
        
        try {
            const res = await this.apiFetch(`/api/chats/${chatId}/messages`);
            this.messages = res || [];
            this.renderMessages();
            this.scrollToBottom();
        } catch (err) {
            this.notify('Ошибка загрузки сообщений', 'error');
        }
    }

    async searchUsers(query) {
        if (!query || query.length < 2) {
            document.getElementById('search-results').innerHTML = '';
            return;
        }
        try {
            const res = await this.apiFetch(`/api/users/search?q=${encodeURIComponent(query)}`);
            this.renderSearchResults(res || []);
        } catch (err) {}
    }

    async createPrivateChat(userId) {
        try {
            const res = await this.apiFetch('/api/chats/private', {
                method: 'POST',
                body: JSON.stringify({ user_id: userId })
            });
            closeNewChatModal();
            this.loadChats();
            this.loadMessages(res.id);
        } catch (err) {
            this.notify(err.message, 'error');
        }
    }

    async sendMessage() {
        const input = document.getElementById('message-input');
        const text = input.value.trim();
        if (!text || !this.activeChatId) return;

        input.value = '';
        try {
            await this.apiFetch('/api/messages', {
                method: 'POST',
                body: JSON.stringify({ chat_id: this.activeChatId, content: text })
            });
            // Сообщение придет через WebSocket
        } catch (err) {
            this.notify('Не удалось отправить сообщение', 'error');
        }
    }

    // --- WebSocket ---
    connectWebSocket() {
        if (this.socket) {
            this.socket.close();
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/api/ws?token=${this.token}`;
        console.log('Connecting to WebSocket:', wsUrl);
        
        this.socket = new WebSocket(wsUrl);
        
        this.socket.onopen = () => {
            console.log('WebSocket connected ✅');
            this.notify('Соединение установлено', 'success');
        };

        this.socket.onmessage = (event) => {
            console.log('WS Message received:', event.data);
            try {
                const wrapper = JSON.parse(event.data);
                console.log('Parsed wrapper:', wrapper);
                
                if (wrapper.type === 'new_message') {
                    const msg = wrapper.content;
                    console.log('New message content:', msg);
                    console.log('Current activeChatId:', this.activeChatId);
                    
                    // Сравниваем ID как строки
                    if (this.activeChatId && String(msg.chat_id) === String(this.activeChatId)) {
                        console.log('Match found, adding message to UI');
                        this.messages.push(msg);
                        this.renderMessages();
                        this.scrollToBottom();
                    } else {
                        console.log('No match or no active chat');
                    }
                    this.updateLastMessageInChatList(msg);
                } else if (wrapper.type === 'user_status') {
                    this.updateUserStatus(wrapper.content);
                }
            } catch (err) {
                console.error('Error parsing WS message:', err);
            }
        };

        this.socket.onclose = (e) => {
            console.log('WebSocket closed ❌', e.reason);
            // Пытаемся переподключиться через 3 секунды
            setTimeout(() => {
                if (this.token) this.connectWebSocket();
            }, 3000);
        };

        this.socket.onerror = (err) => {
            console.error('WebSocket error ⚠️', err);
        };
    }

    updateUserStatus(status) {
        // Обновляем статус в локальном списке чатов
        const chat = this.chats.find(c => String(c.interlocutor_id) === String(status.user_id));
        if (chat) {
            chat.is_online = status.online;
            this.renderChats();
        }

        // Если это активный чат, обновляем заголовок
        const activeChat = this.chats.find(c => String(c.id) === String(this.activeChatId));
        if (activeChat && String(activeChat.interlocutor_id) === String(status.user_id)) {
            const statusEl = document.getElementById('active-chat-status');
            const infoStatusEl = document.getElementById('info-status');
            
            const statusText = status.online ? 'онлайн' : 'был(а) недавно';
            const statusClass = status.online ? 'text-xs text-green-500' : 'text-xs text-gray-400';
            
            if (statusEl) {
                statusEl.textContent = statusText;
                statusEl.className = statusClass;
            }
            if (infoStatusEl) {
                infoStatusEl.textContent = statusText;
            }
        }
    }

    // --- UI Rendering ---
    renderChats() {
        const list = document.getElementById('chats-list');
        list.innerHTML = this.chats.map(chat => `
            <div onclick="app.loadMessages('${chat.id}')" class="chat-list-item p-4 flex items-center gap-3 transition ${String(this.activeChatId) === String(chat.id) ? 'active' : ''}">
                <div class="relative">
                    <div class="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold">
                        ${(chat.name || 'Chat')[0].toUpperCase()}
                    </div>
                    ${chat.is_online ? '<div class="absolute bottom-0 right-0 w-3 h-3 bg-green-500 border-2 border-white rounded-full"></div>' : ''}
                </div>
                <div class="flex-grow overflow-hidden">
                    <div class="flex justify-between items-baseline">
                        <h4 class="font-bold text-gray-900 truncate">${chat.name || 'Chat'}</h4>
                        <span class="text-[10px] text-gray-400">12:45</span>
                    </div>
                    <p class="text-xs text-gray-500 truncate">${chat.last_message || 'Нет сообщений'}</p>
                </div>
            </div>
        `).join('');
    }

    renderMessages() {
        // ... (без изменений, но для точности замены)
        const container = document.getElementById('messages-container');
        container.innerHTML = this.messages.map(msg => {
            const isMe = String(msg.sender_id) === String(this.currentUser?.id);
            return `
                <div class="flex ${isMe ? 'justify-end' : 'justify-start'}">
                    <div class="message-bubble p-4 ${isMe ? 'message-sent' : 'message-received'} shadow-sm">
                        <p class="text-sm">${this.escapeHtml(msg.content)}</p>
                        <div class="text-[10px] ${isMe ? 'text-blue-100' : 'text-gray-400'} mt-1 text-right">
                            ${new Date(msg.created_at).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}
                        </div>
                    </div>
                </div>
            `;
        }).join('');
    }

    renderChatHeader() {
        const chat = this.chats.find(c => String(c.id) === String(this.activeChatId));
        if (!chat) return;
        
        const statusEl = document.getElementById('active-chat-status');
        const infoStatusEl = document.getElementById('info-status');
        
        const isOnline = chat.is_online;
        const statusText = isOnline ? 'онлайн' : 'был(а) недавно';
        const statusClass = isOnline ? 'text-xs text-green-500' : 'text-xs text-gray-400';
        
        if (statusEl) {
            statusEl.textContent = statusText;
            statusEl.className = statusClass;
        }
        
        if (infoStatusEl) {
            infoStatusEl.textContent = statusText;
        }

        document.getElementById('active-chat-name').textContent = chat.name;
        document.getElementById('active-chat-avatar').textContent = chat.name[0].toUpperCase();
        
        // Update Info Panel
        document.getElementById('info-name').textContent = chat.name;
        document.getElementById('info-avatar').textContent = chat.name[0].toUpperCase();
    }

    renderSearchResults(users) {
        const container = document.getElementById('search-results');
        container.innerHTML = users.map(user => `
            <div onclick="app.createPrivateChat('${user.id}')" class="p-3 hover:bg-gray-50 rounded-xl cursor-pointer flex items-center gap-3 transition border border-transparent hover:border-gray-100">
                <div class="w-10 h-10 rounded-full bg-blue-50 flex items-center justify-center text-blue-600 font-bold">${user.username[0].toUpperCase()}</div>
                <div>
                    <div class="font-bold text-gray-900 text-sm">${user.username}</div>
                    <div class="text-xs text-gray-500">${user.email}</div>
                </div>
            </div>
        `).join('');
    }

    loadUserData() {
        document.getElementById('current-user-name').textContent = this.currentUser.username;
        document.getElementById('current-user-avatar').textContent = this.currentUser.username[0].toUpperCase();
    }

    updateLastMessageInChatList(msg) {
        const chat = this.chats.find(c => String(c.id) === String(msg.chat_id));
        if (chat) {
            chat.last_message = msg.content;
            this.renderChats();
        }
    }

    // --- Helpers ---
    async apiFetch(url, options = {}) {
        const headers = {
            'Content-Type': 'application/json',
            ...(this.token ? { 'Authorization': `Bearer ${this.token}` } : {}),
            ...options.headers
        };
        const response = await fetch(url, { ...options, headers });
        if (response.status === 401) this.logout();
        const result = await response.json();
        if (!response.ok) throw new Error(result.error || 'Ошибка запроса');
        return result;
    }

    notify(text, type = 'info') {
        const center = document.getElementById('notification-center');
        const note = document.createElement('div');
        note.className = `p-4 rounded-2xl shadow-xl border text-sm font-medium fade-in ${
            type === 'success' ? 'bg-green-50 border-green-100 text-green-600' : 
            type === 'error' ? 'bg-red-50 border-red-100 text-red-600' : 'bg-white border-gray-100 text-gray-600'
        }`;
        note.textContent = text;
        center.appendChild(note);
        setTimeout(() => note.remove(), 5000);
    }

    scrollToBottom() {
        const container = document.getElementById('messages-container');
        setTimeout(() => {
            container.scrollTop = container.scrollHeight;
        }, 50);
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Global App Instance
window.app = new AlphaApp();
