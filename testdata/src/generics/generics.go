package generics

type Option[T any] interface { // want Option: "generics.Just, generics.None"
	IsOption(T)
}

type Just[T any] struct {
	Val T
}

func (Just[T]) IsOption(v T) {}

type None[T any] struct {
}

func (None[T]) IsOption(v T) {}

type Result[TOk any, TErr any] interface { // want Result: "generics.Ok, generics.Err"
	IsResult(TOk, TErr)
}

type Ok[TOk any, TErr any] struct {
	Ok TOk
}

func (Ok[TOk, TErr]) IsResult(TOk, TErr) {}

type Err[TOk any, TErr any] struct {
	Err TErr
}

func (Err[TOk, TErr]) IsResult(TOk, TErr) {}

func WithDefaultString(s Option[string], defaultVal string) string {
	switch v := s.(type) {
	case Just[string]:
		return v.Val
	case None[string]:
		return defaultVal
	}
	panic("invalid")
}

func WithDefaultStringMissingSwitch(s Option[string], defaultVal string) string {
	switch v := s.(type) {
	case Just[string]:
		return v.Val
	} // want "^missing cases for discriminated union types: generics.None$"
	panic("invalid")
}

func MapResult[TOk any, TErr any, TOkNew any](r Result[TOk, TErr], fn func(TOk) TOkNew) Result[TOkNew, TErr] {
	switch v := r.(type) {
	case Ok[TOk, TErr]:
		return Ok[TOkNew, TErr]{fn(v.Ok)}
	case Err[TOk, TErr]:
		return Err[TOkNew, TErr]{v.Err}
	}
	panic("invalid")
}
