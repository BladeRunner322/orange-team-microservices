// Package domain содержит доменные сущности Profiles-сервиса.
//
// Профиль пользователя: sex, weight_grams, birth_date, height_cm.
// Поля nullable — профиль может быть пустым (lazy-create, см. ADR-001).
//
// Профиль считается "completed", если все 4 поля заполнены.
// Но это вычисляемое свойство, не хранится в БД.
package domain

// TODO: type Profile struct + NewProfile() + RestoreProfile() + геттеры
// TODO: метод Profile.Completed() bool
