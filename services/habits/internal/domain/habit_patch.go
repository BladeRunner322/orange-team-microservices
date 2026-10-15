package domain

// HabitPatch — патч для обновления привычки.
// nil-поле означает «не трогать», не-nil — «установить значение».
type HabitPatch struct {
	Name        *Name
	Description *Description
}

// NewHabitPatch создаёт патч из опциональных полей.
func NewHabitPatch(name *Name, description *Description) HabitPatch {
	return HabitPatch{
		Name:        name,
		Description: description,
	}
}
