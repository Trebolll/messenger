# MVP cutover checklist

1. Infra healthy: Postgres schemas `auth/users/chat/media`, Redis, Kafka topics, MinIO bucket.
2. Auth: register with OTP (dev code in response), login, JWT via gateway.
3. Users: `/api/users/me` works after register; search by username.
4. Chats: create private + group; list; members add/remove.
5. Messages: send/edit/delete/read over REST; history loads.
6. Media: upload via `/api/storage/upload`, download URL works, attachment event published.
7. Realtime: second browser receives `new_message` only for subscribed `chat_id`.
8. Presence: online/offline only for subscribed userIds.
9. Dual realtime: scale `realtime-service` to 2+ replicas; both clients should receive the same chat events via Redis fan-out.
10. Flip proxy to `api-gateway:8080` + React static (`web`).
11. Set strong `JWT_SECRET` / `CONFIRM_SECRET`; set `AUTH_DEV_OTP=false` in production.
