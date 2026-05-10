
-- name: CreateFood :one
-- Creates a food item
INSERT INTO food(user_id,food_name,calories_100,protein_100,carbs_100,fats_100)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING food_id,food_name,calories_100,protein_100,carbs_100,fats_100;

-- name: LogFood :one
-- Logs a particular food eaten
INSERT INTO food_entries(user_id,food_id,calories,total_grams,protein,carbs,fats)
VALUES ($1,$2,$3,$4,$5,$6,$7)
RETURNING food_entries_id,calories,total_grams,protein,carbs,fats;

-- name: ViewFoodsByUser :many
-- Views all the foods associated with a user
SELECT food_id, food_name, calories_100, protein_100, carbs_100, fats_100, created_at
FROM food
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteFoodById :exec
-- Removes a particular food item (by ID)
DELETE FROM food
WHERE food_id = $1 AND user_id = $2;
