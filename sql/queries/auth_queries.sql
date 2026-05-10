

-- name: Create :one
-- Inserts a new user
INSERT INTO users (email, password)
VALUES ($1, $2)
RETURNING user_id, email, role, created_at;


-- name: FindByEmail :one
-- Finds a user by email
SELECT user_id, email, password, role, created_at 
FROM users 
WHERE email = $1;



-- name: FindByID :one
-- Finds a user by ID
SELECT user_id, email, role, created_at 
FROM users 
WHERE user_id = $1;

