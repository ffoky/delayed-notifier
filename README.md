# DelayedNotifier

Сервис отложенных уведомлений через RabbitMQ. Принимает запросы на создание уведомлений, хранит их в PostgreSQL, в нужное время отправляет через Telegram или Email. При ошибке повторяет с экспоненциальной задержкой через DLX-очереди.

## Архитектура

```
webserver  — HTTP API + фронтенд (Gin)
notifier   — обработчик очереди + планировщик (RabbitMQ consumer + gocron)

PostgreSQL — хранилище уведомлений и пользователей
RabbitMQ   — очередь с DLX retry: 5s → 30s → 2m
Redis      — кэш статусов (cache-aside, TTL 5 мин)
```

## API

| Метод | Путь | Описание |
|---|---|---|
| `POST` | `/notify` | Создать уведомление |
| `GET` | `/notify/{id}` | Статус уведомления |
| `DELETE` | `/notify/{id}` | Отменить уведомление |
| `POST` | `/users` | Создать пользователя |
| `POST` | `/auth` | Войти или создать пользователя |

OpenAPI-спецификация: [`api/web-server.yaml`](api/web-server.yaml)

## Настройка

Скопируйте `.env.example` в `.env` и заполните переменные:

```bash
cp .env.example .env
```

### Telegram-бот

1. Напишите [@BotFather](https://t.me/BotFather) → `/newbot`
2. Скопируйте токен в `.env`:
   ```
   TELEGRAM_BOT_API_TOKEN=1234567890:ABC-...
   ```
3. Запустите сервис — пользователи регистрируются командой `/start` в чате с ботом.
   Бот возвращает `user_id`, который нужен для создания уведомлений.

### Email (Gmail)

Gmail требует пароль приложения (не основной пароль аккаунта):

1. Включите двухфакторную аутентификацию в Google-аккаунте
2. Перейдите: [myaccount.google.com/apppasswords](https://myaccount.google.com/apppasswords)
3. Создайте пароль для приложения «Почта»
4. Заполните `.env`:
   ```
   EMAIL_SMTP_HOST=smtp.gmail.com
   EMAIL_SMTP_PORT=587
   EMAIL_FROM=your@gmail.com
   EMAIL_USERNAME=your@gmail.com
   EMAIL_PASSWORD=xxxx xxxx xxxx xxxx
   ```

## Запуск

### Docker Compose

```bash
make up
```

Фронтенд откроется на [http://localhost:8080](http://localhost:8080).  
RabbitMQ Management: [http://localhost:15672](http://localhost:15672) (guest/guest по умолчанию).

Остановка:

```bash
make down
```
