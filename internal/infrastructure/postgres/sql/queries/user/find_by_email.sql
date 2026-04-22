SELECT id, telegram_id, email, created_at
FROM users
WHERE email = $1
LIMIT 1