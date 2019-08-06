package error

type StatusError struct {
	Status         int
	Error          error
	Text           string
	ValidationErrs ValidationErrors
}
