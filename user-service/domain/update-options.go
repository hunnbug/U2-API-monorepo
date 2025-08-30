package domain

import "user-service/valueObjects"

type UpdateOption func(update *UserUpdate)

func WithEmail(email valueObjects.Email) UpdateOption {
	return func(update *UserUpdate) {
		update.FieldsToUpdate[FieldEmail] = email.String()
	}
}

func WithLogin(login valueObjects.Login) UpdateOption {
	return func(update *UserUpdate) {
		update.FieldsToUpdate[FieldLogin] = login.String()
	}
}

func WithPhone(phone valueObjects.Phone) UpdateOption {
	return func(update *UserUpdate) {
		update.FieldsToUpdate[FieldPhone] = phone.String()
	}
}

func WithPassword(password valueObjects.Password) UpdateOption {
	return func(update *UserUpdate) {
		update.FieldsToUpdate[FieldEmail] = password.String()
	}
}
