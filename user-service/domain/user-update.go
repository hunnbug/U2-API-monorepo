package domain

const (
	FieldLogin    = "login"
	FieldEmail    = "email"
	FieldPhone    = "phoneNumber"
	FieldPassword = "password"
)

type UserUpdate struct {
	FieldsToUpdate map[string]string
}

func NewUserUpdate() *UserUpdate {
	return &UserUpdate{make(map[string]string, 0)}
}
