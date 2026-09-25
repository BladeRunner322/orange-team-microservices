-- Таблица профилей пользователей.
--
-- user_id — идентификатор из auth.users (cross-service reference,
-- без FOREIGN KEY, потому что Auth в другой БД).
--
-- Все поля nullable — профиль создаётся лениво (lazy-create, см. ADR-001).

-- TODO: CREATE SCHEMA users;
-- TODO: CREATE TABLE users.profiles (...);
-- TODO: CHECK-констрейнты на диапазоны (sex, weight, height, birth_date)
