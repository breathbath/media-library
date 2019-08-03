package authentication

type AuthError struct{}

func (ae AuthError) Error() string {
	return "Invalid login or password"
}
