import { useEffect, useMemo, useState, type FormEvent } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '../../shared/api/http';
import type { ChatListItem, Message } from '../../shared/types';
import { useAuth } from '../auth/AuthContext';
import { useRealtime } from '../realtime/RealtimeContext';

export function ChatPage() {
  const { user, logout, token } = useAuth();
  const { subscribe, on } = useRealtime();
  const qc = useQueryClient();
  const [activeChatId, setActiveChatId] = useState<string | null>(null);
  const [text, setText] = useState('');
  const [peerUserId, setPeerUserId] = useState('');
  const [groupName, setGroupName] = useState('');
  const [groupUsers, setGroupUsers] = useState('');
  const [online, setOnline] = useState<Record<string, string>>({});

  const chatsQuery = useQuery({
    queryKey: ['chats'],
    queryFn: () => api<ChatListItem[]>('/api/chats'),
    enabled: !!token,
  });

  const messagesQuery = useQuery({
    queryKey: ['messages', activeChatId],
    queryFn: () => api<Message[]>(`/api/chats/${activeChatId}/messages`),
    enabled: !!activeChatId,
  });

  const presenceIds = useMemo(() => {
    const ids = new Set<string>();
    (chatsQuery.data ?? []).forEach((c) => c.memberIds?.forEach((id) => {
      if (id !== user?.id) ids.add(id);
    }));
    return [...ids];
  }, [chatsQuery.data, user?.id]);

  useEffect(() => {
    if (!chatsQuery.data) return;
    subscribe([
      ...chatsQuery.data.map((c) => ({ kind: 'chat', id: c.id })),
      { kind: 'presence', userIds: presenceIds },
      { kind: 'self' },
    ]);
  }, [chatsQuery.data, presenceIds, subscribe]);

  useEffect(() => {
    const offs = [
      on('new_message', (ev) => {
        const chatId = String(ev.chatId);
        qc.invalidateQueries({ queryKey: ['messages', chatId] });
        qc.invalidateQueries({ queryKey: ['chats'] });
      }),
      on('message_edited', (ev) => qc.invalidateQueries({ queryKey: ['messages', String(ev.chatId)] })),
      on('message_deleted', (ev) => qc.invalidateQueries({ queryKey: ['messages', String(ev.chatId)] })),
      on('messages_read', (ev) => qc.invalidateQueries({ queryKey: ['messages', String(ev.chatId)] })),
      on('member_added', () => qc.invalidateQueries({ queryKey: ['chats'] })),
      on('member_removed', () => qc.invalidateQueries({ queryKey: ['chats'] })),
      on('chat_created', () => qc.invalidateQueries({ queryKey: ['chats'] })),
      on('attachment_created', (ev) => {
        if (ev.chatId) qc.invalidateQueries({ queryKey: ['messages', String(ev.chatId)] });
      }),
      on('user_status', (ev) => {
        setOnline((prev) => ({ ...prev, [String(ev.userId)]: String(ev.status) }));
      }),
    ];
    return () => offs.forEach((off) => off());
  }, [on, qc]);

  useEffect(() => {
    if (!activeChatId || !messagesQuery.data?.length) return;
    const last = messagesQuery.data[messagesQuery.data.length - 1];
    api(`/api/chats/${activeChatId}/read`, {
      method: 'POST',
      body: JSON.stringify({ upToMessageId: last.id }),
    }).catch(() => undefined);
  }, [activeChatId, messagesQuery.data]);

  async function sendMessage(e: FormEvent) {
    e.preventDefault();
    if (!activeChatId || !text.trim()) return;
    await api('/api/messages', {
      method: 'POST',
      body: JSON.stringify({ chatId: activeChatId, content: text.trim() }),
    });
    setText('');
    qc.invalidateQueries({ queryKey: ['messages', activeChatId] });
    qc.invalidateQueries({ queryKey: ['chats'] });
  }

  async function createPrivate(e: FormEvent) {
    e.preventDefault();
    const chat = await api<ChatListItem>('/api/chats/private', {
      method: 'POST',
      body: JSON.stringify({ userId: peerUserId }),
    });
    setPeerUserId('');
    await qc.invalidateQueries({ queryKey: ['chats'] });
    setActiveChatId(chat.id);
  }

  async function createGroup(e: FormEvent) {
    e.preventDefault();
    const usernames = groupUsers.split(',').map((s) => s.trim()).filter(Boolean);
    const chat = await api<ChatListItem>('/api/chats/group', {
      method: 'POST',
      body: JSON.stringify({ name: groupName, usernames }),
    });
    setGroupName('');
    setGroupUsers('');
    await qc.invalidateQueries({ queryKey: ['chats'] });
    setActiveChatId(chat.id);
  }

  async function onUpload(file: File) {
    if (!activeChatId) return;
    const fd = new FormData();
    fd.append('file', file);
    fd.append('chatId', activeChatId);
    const uploaded = await api<{ url: string }>('/api/storage/upload', { method: 'POST', body: fd });
    await api('/api/messages', {
      method: 'POST',
      body: JSON.stringify({ chatId: activeChatId, content: `[file] ${uploaded.url}` }),
    });
    qc.invalidateQueries({ queryKey: ['messages', activeChatId] });
  }

  const activeChat = (chatsQuery.data ?? []).find((c) => c.id === activeChatId);

  return (
    <div className="shell">
      <aside className="sidebar">
        <header className="sidebar-head">
          <div>
            <strong>{user?.displayName || user?.username}</strong>
            <div className="muted">@{user?.username}</div>
          </div>
          <button className="ghost" onClick={logout}>Logout</button>
        </header>

        <form className="stack" onSubmit={createPrivate}>
          <input placeholder="User UUID for DM" value={peerUserId} onChange={(e) => setPeerUserId(e.target.value)} />
          <button type="submit">New DM</button>
        </form>
        <form className="stack" onSubmit={createGroup}>
          <input placeholder="Group name" value={groupName} onChange={(e) => setGroupName(e.target.value)} />
          <input placeholder="usernames, comma-separated" value={groupUsers} onChange={(e) => setGroupUsers(e.target.value)} />
          <button type="submit">New group</button>
        </form>

        <div className="chat-list">
          {(chatsQuery.data ?? []).map((c) => (
            <button
              key={c.id}
              className={`chat-item ${c.id === activeChatId ? 'active' : ''}`}
              onClick={() => setActiveChatId(c.id)}
            >
              <div className="chat-title">{c.name || c.type}</div>
              <div className="muted ellipsis">{c.lastMessage || 'No messages'}</div>
              <div className="presence">
                {(c.memberIds || []).filter((id) => id !== user?.id).map((id) => (
                  <span key={id} className={online[id] === 'online' ? 'dot on' : 'dot'} title={id} />
                ))}
              </div>
            </button>
          ))}
        </div>
      </aside>

      <main className="chat-main">
        {!activeChat ? (
          <div className="empty">Select a chat</div>
        ) : (
          <>
            <header className="chat-head">
              <h2>{activeChat.name || activeChat.type}</h2>
            </header>
            <div className="messages">
              {(messagesQuery.data ?? []).map((m) => (
                <div key={m.id} className={`bubble ${m.senderId === user?.id ? 'mine' : ''}`}>
                  <div>{m.content}</div>
                  <time>{new Date(m.createdAt).toLocaleString()}</time>
                </div>
              ))}
            </div>
            <form className="composer" onSubmit={sendMessage}>
              <input type="file" onChange={(e) => e.target.files?.[0] && onUpload(e.target.files[0])} />
              <input
                placeholder="Message…"
                value={text}
                onChange={(e) => setText(e.target.value)}
              />
              <button type="submit">Send</button>
            </form>
          </>
        )}
      </main>
    </div>
  );
}
