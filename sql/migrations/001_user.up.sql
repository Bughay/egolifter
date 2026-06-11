
Create table IF NOT EXISTS  "users" (
    id uuid primary key default gen_random_uuid(),
    email text not null unique,
    password text not null,
    role text not null default 'user',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

Create table IF NOT EXISTS  "user_profiles" (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null unique references "users"(id) on delete cascade,
    first_name text,
    last_name text,
    date_of_birth date,
    height_cm float,
    weight_kg float,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

Create table IF NOT EXISTS  "food" (
    id uuid primary key default gen_random_uuid(),
    name text not null,
    calories_100 float not null,
    protein_100 float not null,
    carbohydrates_100 float not null,
    fat_100 float not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

Create table IF NOT EXISTS  "meal" (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references "users"(id) on delete cascade,
    name text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

Create table IF NOT EXISTS  "food_consumed" (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references "users"(id) on delete cascade,
    meal_id uuid not null references "meal"(id) on delete cascade,
    food_id uuid not null references "food"(id) on delete cascade,
    weight_g float not null,
    total_calories float not null,
    total_protein float not null,
    total_carbohydrates float not null,
    total_fat float not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

Create table IF NOT EXISTS  "workout" (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references "users"(id) on delete cascade,
    name text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

Create Table IF NOT EXISTS  "exercise" (
    id uuid primary key default gen_random_uuid(),
    workout_id uuid not null references "workout"(id) on delete cascade,
    name text not null,
    weight_kg float not null,
    reps integer not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()

);

Create table IF NOT EXISTS  "recipe" (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references "users"(id) on delete cascade,
    name text not null,
    notes text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);


Create table IF NOT EXISTS  "recipe_ingredients" ( 
    id uuid primary key default gen_random_uuid(),
    recipe_id uuid not null references "recipe"(id) on delete cascade,
    food_id uuid not null references "food"(id) on delete cascade,
    weight_g float not null,
    notes text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);