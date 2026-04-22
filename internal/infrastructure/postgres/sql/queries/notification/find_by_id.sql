SELECT id, user_id, channel, text, status, planned_at, created_at, sent_at, retries
FROM notifications
WHERE id = $1
