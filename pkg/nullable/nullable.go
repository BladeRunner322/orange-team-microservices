package nullable

import "encoding/json"

// Nullable — обёртка для patch-полей: позволяет отличить
// "не трогать" (Set: false) от "установить в NULL" (Set: true, Value: nil).
type Nullable[T any] struct {
	Value *T
	Set   bool
}

// UnmarshalJSON различает три состояния JSON-поля: отсутствует, null, значение.
func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Set = true
		n.Value = nil

		return nil
	}

	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	n.Value = &v
	n.Set = true

	return nil
}
