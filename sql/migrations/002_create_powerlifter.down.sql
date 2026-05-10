-- Drop indexes
DROP INDEX IF EXISTS idx_food_entries_recipe_id;
DROP INDEX IF EXISTS idx_food_entries_food_id;
DROP INDEX IF EXISTS idx_food_entries_user_created;
DROP INDEX IF EXISTS idx_recipe_ingredients_recipe_id;
DROP INDEX IF EXISTS idx_recipes_user_id;
DROP INDEX IF EXISTS idx_food_user_id;

-- Drop tables (in reverse dependency order)
DROP TABLE IF EXISTS food_entries;
DROP TABLE IF EXISTS recipe_ingredients;
DROP TABLE IF EXISTS recipes;
DROP TABLE IF EXISTS food_Cache;
DROP TABLE IF EXISTS food;
DROP TABLE IF EXISTS users_profile;