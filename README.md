# LambdaHub

Социальная сеть с мессенджером, лентой активности и встроенным AI-ассистентом.

🌐 **Продакшн:** [lambdahub.ru](https://www.lambdahub.ru)

---

## Функциональность

- **Мессенджер в реальном времени** — личные и групповые чаты на WebSocket
- **Стена пользователя** — публикации с медиафайлами, лайки, комментарии
- **Лента активности** — агрегация событий по подпискам
- **AI-ассистент** — встроенный чат с языковой моделью (DeepSeek)
- **Умная авторизация** — вход по email или телефону с OTP-кодом (Twilio)
- **Превью ссылок** — автоматическое получение Open Graph метаданных
- **Медиафайлы** — загрузка и хранение фото/видео через MinIO (S3-совместимое)
- **Рейтинговая система** — оценки сообщений в реальном времени

---

## Стек технологий

| Слой | Технологии |
|---|---|
| Backend | Go 1.24, Gin |
| База данных | PostgreSQL, golang-migrate |
| Real-time | WebSocket (gorilla/websocket) |
| Хранилище файлов | MinIO (S3) |
| Авторизация | JWT, OTP через Twilio / SMTP |
| AI | DeepSeek API |
| Инфраструктура | Docker, Docker Compose |
| Тесты | testify, testcontainers-go |

---

## Архитектура

Чистая слоистая архитектура с ручным dependency injection:

```
cmd/server/
└── main.go              # точка входа, инициализация зависимостей

internal/
├── handler/             # HTTP и WebSocket хендлеры
├── service/             # бизнес-логика
├── repository/          # работа с БД
├── model/               # доменные модели
├── middleware/           # JWT-аутентификация
├── config/              # конфигурация
└── db/                  # инициализация подключения
```

---

## Запуск локально

### Требования
- Go 1.24+
- Docker и Docker Compose

### Настройка окружения

Создайте `.env` файл на основе примера:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=messenger

MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=your_key
MINIO_SECRET_KEY=your_secret
MINIO_BUCKET=messenger
MINIO_PUBLIC_ENDPOINT=http://localhost:9000

JWT_SECRET=your_jwt_secret

DEEPSEEK_API_KEY=your_key

TWILIO_ACCOUNT_SID=your_sid
TWILIO_AUTH_TOKEN=your_token
TWILIO_FROM_PHONE=+1234567890
```

### Запуск

```bash
# Запуск всей инфраструктуры
docker compose up -d

# Или только локальная разработка
docker compose -f docker-compose.local.yml up -d --build
```

---

## Тесты

```bash
go test ./...
```

Интеграционные тесты используют testcontainers — PostgreSQL поднимается автоматически в Docker.


cat ~/.ssh/id_ed25519Значит в репо всё ещё старый docker-compose.yml с DB_PASSWORD: postgres. Проверь локально в репо — есть ли там DB_PASSWORD в секции бэкенда?
docker exec -it messenger-db psql -U postgres -d messenger


.\ngrok.exe start --all --config=ngrok.yml
docker-compose up --build -d backend
docker-compose logs -f backend
docker compose -f docker-compose.local.yml up -d --build

ssh root@193.233.103.158



авторизация гит на сервере
ssh -p 9871 root@193.233.103.158
cd /opt/messenger
git remote set-url origin git@github.com:Trebolll/messenger.git
git fetch origin
git reset --hard origin/master

обновить код на сервере
cd /opt/messenger
git pull origin master
docker compose build backend
docker compose up -d --force-recreate backend

//чистка
docker builder prune -af        # только кеш сборки (самый тяжёлый обычно)
docker image prune -af          # неиспользуемые образы
docker container prune -f       # остановленные контейнеры
docker volume prune -f          # неиспользуемые тома
docker network prune -f         # неиспользуемые сети


//логи
docker compose logs backend --tail=50

docker compose -f docker-compose.local.yml up -d --build

//пушнуть строку в env
echo "DEEPSEEK_API_KEY=..." >> .env