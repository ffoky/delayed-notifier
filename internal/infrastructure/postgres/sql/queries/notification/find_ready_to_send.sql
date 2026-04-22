UPDATE notifications
SET status = 'sending'
WHERE id IN (
    SELECT id FROM notifications
    WHERE status = 'planned' AND planned_at <= NOW()
    ORDER BY planned_at
    LIMIT 100
    FOR UPDATE SKIP LOCKED
)
RETURNING id, user_id, channel, text, status, planned_at, created_at, sent_at, retries
