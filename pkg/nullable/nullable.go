package nullable

// Nullable — обёртка для patch-полей: позволяет отличить
// "не трогать" (Set: false) от "установить в NULL" (Set: true, Value: nil).
type Nullable[T any] struct {
	Value *T
	Set   bool
}
