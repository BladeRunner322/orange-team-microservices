package domain

// Name — value object названия привычки.
// 1–100 символов после trim.
type Name string

// NewName создаёт Name или возвращает ErrInvalidName.
func NewName(raw string) (Name, error) {
	// TODO: trim, проверить длину через utf8.RuneCountInString
	return "", nil
}

func (n Name) String() string { return string(n) }

// Description — value object описания привычки.
// 0–500 символов после trim. Может быть пустым.
type Description string

// NewDescription создаёт Description или возвращает ErrInvalidDescription.
func NewDescription(raw string) (Description, error) {
	// TODO: trim, проверить длину
	return "", nil
}

func (d Description) String() string { return string(d) }
