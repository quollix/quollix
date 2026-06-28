package users

const UsernameMaxLength = 20

type Credentials struct {
	Username string `json:"username" validate:"username"`
	Password string `json:"password" validate:"password"`
}

type PasswordResetRequest struct {
	UserId      string `json:"user_id" validate:"number"`
	NewPassword string `json:"new_password" validate:"password"`
}
