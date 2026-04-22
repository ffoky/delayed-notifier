#!/bin/bash

BASE_URL="http://localhost:8080"
EMAIL="test@mail.ru"
DELAY_MINUTES=5

echo "Создаём пользователя с email: $EMAIL"
USER_RESPONSE=$(curl -s -X POST "$BASE_URL/users" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL\"}")

echo "Ответ: $USER_RESPONSE"
USER_ID=$(echo "$USER_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

if [ -z "$USER_ID" ]; then
  echo "Ошибка: не удалось получить user_id"
  exit 1
fi

echo "User ID: $USER_ID"

PLANNED_AT=$(date -u -d "+${DELAY_MINUTES} minutes" "+%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || \
             date -u -v+${DELAY_MINUTES}M "+%Y-%m-%dT%H:%M:%SZ")

echo "Создаём уведомление на $PLANNED_AT"
NOTIFY_RESPONSE=$(curl -s -X POST "$BASE_URL/notify" \
  -H "Content-Type: application/json" \
  -d "{\"user_id\": \"$USER_ID\", \"channel\": \"email\", \"text\": \"Тест отложенного уведомления!\", \"planned_at\": \"$PLANNED_AT\"}")

echo "Ответ: $NOTIFY_RESPONSE"
NOTIFY_ID=$(echo "$NOTIFY_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

if [ -z "$NOTIFY_ID" ]; then
  echo "Ошибка: не удалось создать уведомление"
  exit 1
fi

echo ""
echo "Уведомление создано!"
echo "Notification ID: $NOTIFY_ID"
echo "Проверить статус: curl $BASE_URL/notify/$NOTIFY_ID"
echo "Письмо придёт через $DELAY_MINUTES мин на $EMAIL"