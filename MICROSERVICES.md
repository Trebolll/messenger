# LambdaHub Microservices (MVP)

React frontend + Spring Boot services with Kafka/Redis realtime.

## Multi-repo layout

Each top-level folder is its own GitLab repository. The root `pom.xml` is a **local workspace aggregator only** (not published).

| Folder | Role |
|--------|------|
| `api-gateway` | Edge gateway (JWT, routing) |
| `auth-service` | Auth / OTP / JWT |
| `user-service` | Profiles & contacts |
| `chat-service` | Chats & messages |
| `media-service` | Media (MinIO) |
| `realtime-service` | WebSocket fan-out |
| `common-security` | Shared JWT helpers |
| `common-events` | Shared events / topics |
| `common-web` | Shared web error helpers |
| `common-kafka-starter` | Kafka connection defaults |
| `common-redis-starter` | Redis connection defaults |
| `web` | React 19 + Vite SPA |
| `docs` | Cutover / ops notes |

Each Java service is a standalone multi-module Maven project (`*-api`, `*-db`, `*-impl`) with parent `spring-boot-starter-parent:4.1.0`.

## Local JVM run

1. Postgres / Redis / Kafka / MinIO — снаружи (свои инстансы или k8s).
2. Build (Java **26**, Spring Boot **4.1**):

```bash
mvn -s .mvn/local-settings.xml -DskipTests package
```

3. Frontend:

```bash
cd web
npm install
npm run dev
```

## Realtime model

- Mutations go through REST
- WebSocket is push-only with explicit `subscribe` to `chat`, `presence`, `self`
- Domain events → Kafka → realtime-service → Redis pub/sub → WS clients

## Cutover checklist

- [ ] Point nginx/reverse proxy `/api` to gateway `:8080`
- [ ] Point SPA to `web` build
- [ ] Set strong `JWT_SECRET` / `CONFIRM_SECRET`
- [ ] Disable `AUTH_DEV_OTP` in production
- [ ] Scale `realtime-service` to 2+ replicas and verify fan-out
