SELECT id, telegram_id, email, created_at
FROM users
WHERE telegram_id = $1
LIMIT 1