package async

type Result[T any] struct {
	Value T
	Err   error
}

func (r *Result[T]) IsErr() bool {
	return r.Err != nil
}

func SuccessResult[T any](v T) Result[T] {
	return Result[T]{
		Value: v,
		Err:   nil,
	}
}

func ErrorResult[T any](err error) Result[T] {
	return Result[T]{
		Err: err,
	}
}

func ResultFrom[T any](v T, err error) Result[T] {
	return Result[T]{v, err}
}
