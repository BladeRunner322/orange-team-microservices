CREATE SCHEMA IF NOT EXISTS profiles;

CREATE TABLE IF NOT EXISTS profiles.users (
    user_id UUID PRIMARY KEY,
    sex VARCHAR(16),
    weight_grams INTEGER,
    birth_date DATE,
    height_cm SMALLINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

-- sex
ALTER TABLE profiles.users ADD CONSTRAINT chk_profiles_users_sex
    CHECK (sex IN ('male', 'female'));

-- weight_grams
ALTER TABLE profiles.users ADD CONSTRAINT chk_profiles_users_weight_grams
    CHECK (weight_grams BETWEEN 40_000 AND 150_000);
ALTER TABLE profiles.users ADD CONSTRAINT chk_profiles_users_weight_grams_precision
    CHECK (weight_grams % 100 = 0);

-- birth_date
ALTER TABLE profiles.users ADD CONSTRAINT chk_profiles_users_birth_date
    CHECK (birth_date BETWEEN DATE '1900-01-01' AND CURRENT_DATE);

-- height
ALTER TABLE profiles.users ADD CONSTRAINT chk_profiles_users_height_cm
    CHECK (height_cm BETWEEN 140 AND 210);
