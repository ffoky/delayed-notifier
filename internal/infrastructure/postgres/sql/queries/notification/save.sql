INSERT INTO notifications (id, user_id, channel, text, status, planned_at, created_at, sent_at, retries)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
