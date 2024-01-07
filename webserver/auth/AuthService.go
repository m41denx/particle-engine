package auth

type AuthService interface {
	GetUIDByToken(token string) (uid int64, err error)
	Login(username, password string) (token string, err error)
}
