package domain

// ExercisePatch — патч для обновления упражнения.
// nil-поле означает «не трогать», не-nil — «установить значение».
//
// type НЕ входит — immutable (см. ADR-008).
type ExercisePatch struct {
	Name        *Name
	Description *Description
	Difficulty  *Difficulty
}

// NewExercisePatch создаёт патч из набора опциональных полей.
func NewExercisePatch(
	name *Name,
	description *Description,
	difficulty *Difficulty,
) ExercisePatch {
	return ExercisePatch{
		Name:        name,
		Description: description,
		Difficulty:  difficulty,
	}
}
