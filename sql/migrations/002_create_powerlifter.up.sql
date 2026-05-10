CREATE TABLE IF NOT EXISTS users_profile (
    user_id INTEGER PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    date_of_birth DATE,
    height DECIMAL,
    weight DECIMAL,
    is_trainer   BOOLEAN NOT NULL DEFAULT FALSE,
    last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS food (
    food_id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(user_id),
    food_name VARCHAR(255) NOT NULL,
    calories_100 DECIMAL NOT NULL,
    protein_100 DECIMAL NOT NULL,
    carbs_100 DECIMAL NOT NULL,
    fats_100 DECIMAL NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS recipes (
    recipe_id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(user_id),
    recipe_name VARCHAR(255) NOT NULL,
    instructions TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS recipe_ingredients (
    ingredient_id SERIAL PRIMARY KEY,
    recipe_id INTEGER NOT NULL REFERENCES recipes(recipe_id),
    food_id INTEGER NOT NULL REFERENCES food(food_id),
    total_grams DECIMAL NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS food_entries (
    food_entries_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ,
    food_id INTEGER REFERENCES food(food_id),
    calories DECIMAL NOT NULL,
    total_grams DECIMAL NOT NULL,
    protein DECIMAL NOT NULL,
    carbs DECIMAL NOT NULL,
    fats DECIMAL NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);



-- Food lookups by user (your custom foods list)
CREATE INDEX idx_food_user_id ON food(user_id);

-- Recipe lookups by user
CREATE INDEX idx_recipes_user_id ON recipes(user_id);

-- Recipe ingredients
CREATE INDEX idx_recipe_ingredients_recipe_id ON recipe_ingredients(recipe_id);

-- Daily food log entries 
CREATE INDEX idx_food_entries_user_created ON food_entries(user_id, created_at);

-- Joins from entries to food/recipes
CREATE INDEX idx_food_entries_food_id ON food_entries(food_id);
