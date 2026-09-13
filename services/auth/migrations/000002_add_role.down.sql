ALTER TABLE auth.users DROP CONSTRAINT IF EXISTS chk_auth_users_role;
ALTER TABLE auth.users DROP COLUMN IF EXISTS role;