-- name: CreateFood :one
INSERT INTO food (name, calories_100, protein_100, carbohydrates_100, fat_100)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetFoodByID :one
SELECT * FROM food
WHERE id = $1;

-- name: ListFoods :many
SELECT * FROM food
ORDER BY created_at DESC;

-- name: UpdateFood :one
UPDATE food
SET name              = $2,
    calories_100      = $3,
    protein_100       = $4,
    carbohydrates_100 = $5,
    fat_100           = $6,
    updated_at        = now()
WHERE id = $1
RETURNING *;

-- name: DeleteFood :exec
DELETE FROM food
WHERE id = $1;
