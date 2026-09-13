ALTER TABLE auth.users
    ADD COLUMN role VARCHAR(16) NOT NULL DEFAULT 'user';

ALTER TABLE auth.users
    ADD CONSTRAINT chk_auth_users_role
    CHECK (role IN ('user', 'admin'));