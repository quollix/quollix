package users

const UsernameMaxLength = 20

type PasswordResetRequest struct {
	UserId      string `json:"user_id" validate:"number"`
	NewPassword string `json:"new_password" validate:"password"`
}
