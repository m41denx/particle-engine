package structs

type ErrorResponse struct {
	Message string
}

type UserResponse struct {
	Username       string `json:"username"`
	MaxAllowedSize uint   `json:"max_allowed_size"`
	UsedSize       uint   `json:"used_size"`
}
