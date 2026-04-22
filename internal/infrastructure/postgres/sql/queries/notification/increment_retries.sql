UPDATE notifications
SET retries = retries + 1
WHERE id = $1
