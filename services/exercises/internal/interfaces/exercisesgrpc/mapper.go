package exercisesgrpc

// TODO: мапперы proto ↔ domain (по образцу services/profiles/internal/interfaces/profilesgrpc/mapper.go):
//   - ToProtoExercise(domain.Exercise) *exercises.Exercise
//   - patchFromProto(*exercises.PatchExerciseRequest) (domain.ExercisePatch, error)
//   - обёртки nullable: NullableString → nullable.Nullable[domain.Name/Description],
//     NullableInt32 → nullable.Nullable[domain.Difficulty]
