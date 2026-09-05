export type User = {
  id: string;
  username: string;
  displayName?: string;
  email?: string;
  phone?: string;
  avatarUrl?: string;
  statusText?: string;
};

export type ChatListItem = {
  id: string;
  type: string;
  name?: string;
  avatarUrl?: string;
  lastMessage?: string;
  lastMessageAt?: string;
  memberIds: string[];
};

export type Message = {
  id: string;
  chatId: string;
  senderId: string;
  content: string;
  parentId?: string;
  createdAt: string;
  editedAt?: string;
};
